package apiserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// uninstallV2 is the rewrite of `uninstall`. See
// cancel_v2.go for the shared design rationale; this handler adds
// two uninstall-specific wrinkles:
//
//   - Sys apps are rejected outright.
//   - Three annotations are stamped alongside OpType so the
//     downstream Runner handlers can read the operator's choices
//     (uninstallAll = also tear down the server side, deleteData
//     = wipe persistent volumes, token = short-lived K8s SA token
//     the uninstall pipeline uses to talk to per-user resources).
//     Legacy stamped them on Spec; we preserve that so the state
//     machine can pick them up without any schema change.
//
// Not wired into the router yet — same incremental-cutover story
// as cancelV2.
func (h *Handler) uninstallV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	request := &api.UninstallRequest{}
	if req.Request.ContentLength > 0 {
		if err := req.ReadEntity(request); err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}

	if userspace.IsSysApp(app) {
		api.HandleBadRequest(resp, req, errors.New("sys app can not be uninstall"))
		return
	}

	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	name, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}

	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	uninstallAll := isAdmin || request.All

	h.deliverUninstallV2(req.Request.Context(), req, resp, name, amPtr, app, uninstallV2Options{
		Token:        token,
		UninstallAll: uninstallAll,
		DeleteData:   request.DeleteData,
	})
}

// uninstallV2Options is the pre-decoded set of caller choices
// deliverUninstallV2 acts on. Split from the outer shell so unit
// tests do not have to fake kubeConfig / KubeSphere / body decoding
// to exercise the translator core.
type uninstallV2Options struct {
	// Token is the caller-scoped K8s SA token; stamped as an
	// annotation so the state-machine handler can use it to call
	// per-user resources (e.g. helm uninstall against the user's
	// service account).
	Token string
	// UninstallAll asks the uninstall pipeline to also tear down
	// server-side artifacts (admin-only in legacy: the outer
	// shell OR's request.All with isAdmin).
	UninstallAll bool
	// DeleteData asks the uninstall pipeline to wipe persistent
	// volumes on top of the normal helm uninstall.
	DeleteData bool
}

// deliverUninstallV2 is the pure translator step of uninstallV2 —
// mirror of deliverCancelV2 with uninstall-specific annotations.
//
// Contract (see deliverCancelV2 for the rest):
//
//   - Writes Annotations {AppTokenKey, AppUninstallAllKey,
//     AppDeleteDataKey} through the same MergeFrom patch as the
//     Spec.OpType update, so a stale in-memory Status snapshot
//     cannot clobber the Runner-written Status.
func (h *Handler) deliverUninstallV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *v1alpha1.ApplicationManager,
	app string,
	opts uninstallV2Options,
) {
	am := *amPtr

	if blockIfRequestPending(resp, req, &am, v1alpha1.UninstallOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, v1alpha1.UninstallOp) {
		api.HandleBadRequest(resp, req, explainOperationNotAllowed(am.Status.State, v1alpha1.UninstallOp))
		return
	}

	patched := am.DeepCopy()
	patched.Spec.OpType = v1alpha1.UninstallOp
	if patched.Annotations == nil {
		patched.Annotations = make(map[string]string)
	}
	patched.Annotations[api.AppTokenKey] = opts.Token
	patched.Annotations[api.AppUninstallAllKey] = fmt.Sprintf("%t", opts.UninstallAll)
	patched.Annotations[api.AppDeleteDataKey] = fmt.Sprintf("%t", opts.DeleteData)
	var opID string
	patched.Annotations, opID = stampDurableOperationRequest(patched.Annotations)

	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

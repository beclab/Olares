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
// three uninstall-specific wrinkles:
//
//   - Sys apps are rejected outright.
//   - `force=true` in the body bypasses the AllowsOp
//     precheck. Legacy behaviour: an operator can force an
//     uninstall out of any state (e.g. to unstick a wedged
//     Installing → Uninstalling); the state machine's Uninstalling
//     handler is expected to cope with an unusual entry state.
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

	if h.opDispatcher == nil {
		api.HandleError(resp, req, fmt.Errorf("uninstallV2: opDispatcher is not wired"))
		return
	}

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
		Force:        request.Force,
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
	// Force skips the AllowsOp precheck when true.
	Force bool
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
// mirror of deliverCancelV2 with uninstall-specific annotations and
// the `force` override for AllowsOp.
//
// Contract (see deliverCancelV2 for the rest):
//
//   - Force=true skips the AllowsOp check but does NOT
//     bypass the sys-app rejection (the outer shell already gated
//     that; do not re-check here so the split remains a pure
//     "post-auth translator").
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
	if !opts.Force && !domain.DefaultTable.AllowsOp(am.Status.State, v1alpha1.UninstallOp) {
		api.HandleBadRequest(resp, req, explainOperationNotAllowed(am.Status.State, v1alpha1.UninstallOp))
		return
	}
	if opts.Force {
		klog.Warningf("force uninstalling app %s from %s state", app, am.Status.State)
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

	h.opDispatcher.Trigger(name, domain.TriggerOp{Op: v1alpha1.UninstallOp})

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

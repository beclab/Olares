package apiserver

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// appApplyEnvV2 is the rewrite of `appApplyEnv`.
//
// applyenv is the simplest op in the /apps/{name}/... family:
//
//   - No body payload — the caller conveys "apply the currently
//     persisted app envs" and nothing else.
//   - AppTokenKey (the caller-scoped SA token the downstream
//     helm-upgrade uses to talk to per-user resources).
//   - AppPreUpgradeStateKey, snapshotted from the current state the
//     same way the AppEnv controller does, so ApplyingEnv can land a
//     Stopped app back in Stopped and a Running apply replaces a
//     leftover stopped value.
//   - One TriggerOp: {ApplyEnvOp}.
//
// Legacy carried a defensive rejection "if am.Annotations == nil"
// on the assumption that an AM with no annotation map at all is
// a mis-installed / manually-fabricated object that should not be
// operated on. We preserve that check verbatim so the v2 rejection
// surface stays byte-identical to the legacy client contract.
//
// Not wired into the router yet — same incremental-cutover story
// as cancelV2.
func (h *Handler) appApplyEnvV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	name, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}

	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	h.deliverApplyEnvV2(req.Request.Context(), req, resp, name, amPtr, token)
}

// deliverApplyEnvV2 is the pure translator step. Mirror of
// deliverCancelV2 with the ApplyEnv-specific rejection
// (nil-annotations) and token-annotation stamp.
//
// Contract:
//
//   - amPtr MUST NOT be nil.
//   - AM.Annotations == nil is rejected as a 5xx (legacy shape:
//     api.HandleError, not api.HandleBadRequest — preserved so the
//     v2 responses match legacy byte-for-byte on that error path).
//   - Writes Spec.OpType, the token, the operation-request id, and
//     the pre-op state snapshot in the same MergeFrom patch.
//   - Delivers exactly one TriggerOp{ApplyEnvOp}.
//   - Does NOT touch AM.Status directly.
//   - Responds with the plain api.Response{Code: 200} (no UID
//     payload — legacy shape).
func (h *Handler) deliverApplyEnvV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *appv1alpha1.ApplicationManager,
	token string,
) {
	am := *amPtr
	if blockIfRequestPending(resp, req, &am, appv1alpha1.ApplyEnvOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, appv1alpha1.ApplyEnvOp) {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("%s operation is not allowed for %s state",
				appv1alpha1.ApplyEnvOp, am.Status.State))
		return
	}

	// Preserved from legacy — an AM with no annotation map at all
	// is treated as an unsupported target for ApplyEnv. If future
	// requirements relax this, remove both the check here and the
	// legacy check together so the two handlers keep the same
	// rejection surface for as long as they coexist.
	if am.Annotations == nil {
		klog.Errorf("not support operation %s,name:%s", appv1alpha1.ApplyEnvOp, am.Spec.AppName)
		api.HandleError(resp, req, fmt.Errorf("not support operation %s", appv1alpha1.ApplyEnvOp))
		return
	}

	patched := am.DeepCopy()
	patched.Spec.OpType = appv1alpha1.ApplyEnvOp
	patched.Annotations[api.AppTokenKey] = token
	// OnApplyingEnvWorkloadReady reads this to put a Stopped app back
	// in Stopped. Running and Stopped overwrite, so a later apply from
	// Running cannot reuse a leftover stopped value and shut the app
	// down. Failed retries keep the prior steady intent.
	api.SnapshotPreUpgradeState(patched.Annotations, am.Status.State)
	patched.Annotations, _ = stampDurableOperationRequest(patched.Annotations)
	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

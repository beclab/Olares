package apiserver

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// cancelV2 is the rewrite of `cancel`, the first HTTP
// handler collapsed onto the applifecycle Runner. It intentionally
// only does three things:
//
//  1. Authorise + validate the request (unchanged from legacy).
//  2. Persist user intent on the AM.Spec side (OpType = CancelOp)
//     plus a new operation-request-id. The ApplicationManager watch
//     enqueues the Runner, which derives TriggerOp from that id.
//
// Notably absent, compared to `cancel`:
//
//   - No apputils.UpdateAppMgrStatus call. The Runner is the sole
//     writer of ApplicationManager.Status; the target *Canceling
//     state, message, reason, StatusTime, and OpID are all decided
//     inside the state-machine handler (see pkg/applifecycle/
//     handlers/cancel_pipeline.go and cancel_handoff.go).
//   - No per-source-state `cancelState` switch. The state machine
//     computes the correct *Canceling target from (currentState,
//     TriggerOp{CancelOp}) by lookup in the TransitionTable, so
//     drift between the two locations is impossible by construction.
//   - No `?type=timeout|operate` distinction. Legacy carried the
//     cancel origin (user vs timeout) into Status.Reason/Message so
//     dashboards could tell them apart. The state-machine cancel
//     handlers currently write a fixed CancelRequested reason; the
//     origin distinction is deferred until TriggerOp gains an origin
//     field or the *Canceling handlers gain an annotation-driven
//     Reason picker. Tracked as SKELETON in cancel_handoff.go.
//
// Routing note: this handler is NOT wired into the webservice router.
// The legacy `cancel` still services POST /apps/{name}/cancel.
// The router swap will happen in the follow-on PR.
func (h *Handler) cancelV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	name, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}
	h.deliverCancelV2(req.Request.Context(), req, resp, name, amPtr, app)
}

// deliverCancelV2 is the pure translator step of cancelV2. Split
// from the outer HTTP shell so unit tests can exercise the AM
// write without stubbing the auth prelude
// (loadAuthorizedLifecycleAM depends on utils.GetClient which reads
// a real kubeconfig).
//
// Contract:
//
//   - amPtr MUST NOT be nil (guaranteed by loadAuthorizedLifecycleAM
//     on ok=true).
//   - Writes exactly one AM.Spec update via ctrlClient. The
//     ApplicationManager watch enqueues the Runner.
//   - Does NOT touch AM.Status directly — the state machine owns
//     Status.
func (h *Handler) deliverCancelV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *v1alpha1.ApplicationManager,
	app string,
) {
	am := *amPtr
	state := am.Status.State
	if blockIfRequestPending(resp, req, &am, v1alpha1.CancelOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(state, v1alpha1.CancelOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.CancelOp, state))
		return
	}

	// Spec side is user-owned and stays on the plain controller-
	// runtime client. The request id is how the next Reconcile
	// derives TriggerOp{CancelOp}.
	//
	// Prefer Patch(MergeFrom(orig)) over Update(patched) so we send
	// EXACTLY the Spec.OpType diff to the API server: an Update
	// would round-trip the entire object, and if am.Status is stale
	// relative to the on-cluster Status (the state machine writes
	// Status on its own cadence), the Update would overwrite the
	// Runner's newer Status with our snapshot. Legacy handlers get
	// away with this because they follow up with an explicit Status
	// write, which we no longer do — the state machine owns Status.
	patched := am.DeepCopy()
	patched.Spec.OpType = v1alpha1.CancelOp
	if patched.Annotations == nil {
		patched.Annotations = make(map[string]string)
	}
	var opID string
	patched.Annotations, opID = stampDurableOperationRequest(patched.Annotations)
	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// The returned OpID is the durable request ID stamped above; the
	// Runner copies it into Status.OpID when it writes the *Canceling
	// PatchAM, so a caller can poll on it before the first Reconcile.
	resp.WriteAsJson(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

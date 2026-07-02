package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// suspendV2 is the rewrite of `suspend`.
//
// Design notes vs. the trivial op handlers (cancel / uninstall /
// applyenv):
//
//   - There is a mutating compute pre-side-effect: legacy calls
//     compute.DeleteAllocationsForComputeTarget BEFORE writing
//     Status=Stopping so the downstream Stopping state-machine
//     handler doesn't see stale compute allocations. We keep that
//     ordering — compute.DeleteAllocationsForComputeTarget runs
//     between the spec patch and the TriggerOp deliver, so nothing
//     racing on the informer stream can observe an inconsistent
//     "OpType=StopOp but allocations still live" window.
//
//     Compute ops are NOT (yet) a state-machine port — the
//     Stopping handler in the applifecycle module only owns helm
//     scale-to-zero + namespace teardown. When compute becomes a
//     first-class port, this call moves onto a StoppingEffect and
//     this handler thins down to the shape of applyEnvV2. Tracked
//     in ADAPTERS_TODO.md.
//
//   - There is one annotation: AppStopAllKey. Legacy computes
//     stopAll = isAdmin || request.All; we preserve that in the
//     outer shell and pass the resolved boolean into the deliver
//     helper.
//
// Not wired into the router yet — same incremental-cutover story
// as cancelV2.
func (h *Handler) suspendV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	if h.opDispatcher == nil {
		api.HandleError(resp, req, fmt.Errorf("suspendV2: opDispatcher is not wired"))
		return
	}

	request := &api.StopRequest{}
	if req.Request.ContentLength > 0 {
		if err := req.ReadEntity(request); err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}
	if userspace.IsSysApp(app) {
		api.HandleBadRequest(resp, req, errors.New("sys app can not be suspend"))
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
	stopAll := isAdmin || request.All

	h.deliverSuspendV2(req.Request.Context(), req, resp, name, amPtr, app, stopAll)
}

// deliverSuspendV2 is the pure translator + compute-pre-side-effect
// step of suspendV2.
//
// Contract:
//
//   - opDispatcher MUST be non-nil (checked by the outer shell).
//   - amPtr MUST NOT be nil.
//   - AM.Spec.Config MUST unmarshal to an appcfg.ApplicationConfig
//     (invariant: install-time writes it; a malformed value means
//     the AM was hand-edited and we refuse to guess).
//   - Runs compute.DeleteAllocationsForComputeTarget BEFORE
//     Trigger so the state-machine Stopping handler doesn't
//     race on stale allocations.
//   - Writes AM.Spec.OpType + AppStopAllKey first, then persists the
//     durable request ID in a second metadata patch after compute
//     cleanup succeeds; does NOT touch AM.Status directly.
//   - Delivers exactly one TriggerOp{StopOp} on success.
//   - On any pre-Trigger error, returns without dispatching.
func (h *Handler) deliverSuspendV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *v1alpha1.ApplicationManager,
	app string,
	stopAll bool,
) {
	am := *amPtr
	if blockIfRequestPending(resp, req, &am, v1alpha1.StopOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, v1alpha1.StopOp) {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.StopOp, am.Status.State))
		return
	}

	patched := am.DeepCopy()
	patched.Spec.OpType = v1alpha1.StopOp
	if patched.Annotations == nil {
		patched.Annotations = make(map[string]string)
	}
	patched.Annotations[api.AppStopAllKey] = fmt.Sprintf("%t", stopAll)
	delete(patched.Annotations, domain.AnnotationOperationRequestID)
	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// Compute pre-side-effect. Runs AFTER the spec patch (so a
	// spec-write failure short-circuits without touching compute
	// state) but BEFORE the TriggerOp deliver (so the Stopping
	// state-machine handler never sees stale allocations). If this
	// step fails we surface an error and skip the trigger — the AM
	// is left with OpType=StopOp / Status=<prior> for the operator
	// to retry; this matches legacy which behaves the same way.
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(am.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup of app %s failed %v", app, err)
		api.HandleError(resp, req, err)
		return
	}
	if err := compute.DeleteAllocationsForComputeTarget(ctx, h.ctrlClient, &appCfg, stopAll); err != nil {
		klog.Errorf("delete compute allocation for suspended app %s failed %v", app, err)
		api.HandleError(resp, req, err)
		return
	}

	var current v1alpha1.ApplicationManager
	if err := h.ctrlClient.Get(ctx, client.ObjectKey{Name: name}, &current); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	requestPatch := current.DeepCopy()
	if requestPatch.Annotations == nil {
		requestPatch.Annotations = make(map[string]string)
	}
	var opID string
	requestPatch.Annotations, opID = stampDurableOperationRequest(requestPatch.Annotations)
	if err := h.ctrlClient.Patch(ctx, requestPatch, client.MergeFrom(&current)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	h.opDispatcher.Trigger(name, domain.TriggerOp{Op: v1alpha1.StopOp})

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

// resumeV2 is the rewrite of `resume`.
//
// Design notes:
//
//   - Two pre-side-effects, not one:
//     1. validation.Run(ResumePressureValidators) — pure resource
//     gate; failure surfaces as a 400 with the standard
//     RequirementResp envelope.
//     2. compute.ApplyBindingSelection — MUTATING; may allocate
//     compute. If any subsequent step (spec patch, trigger)
//     fails, we MUST undo the allocation via
//     compute.DeleteAllocationsForApp to avoid orphaned
//     resources. Legacy handles this with `defer` +
//     `bindingAccepted bool`; we preserve that shape.
//
//   - Body carries a ComputeBinding selection the caller may
//     supply (empty -> stick with the install-time binding).
//
//   - One annotation: AppResumeAllKey. Legacy sets it to
//     "false" by default and only flips to "true" if
//     compute.ShouldIncludeSharedServerForResume returns true. We
//     mirror that in the outer shell.
//
// Not wired into the router yet.
func (h *Handler) resumeV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	if h.opDispatcher == nil {
		api.HandleError(resp, req, fmt.Errorf("resumeV2: opDispatcher is not wired"))
		return
	}

	request := &ResumeRequest{}
	if err := readOptionalEntity(req, request); err != nil {
		api.HandleBadRequest(resp, req, err)
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
	am := *amPtr
	if blockIfRequestPending(resp, req, &am, v1alpha1.ResumeOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, v1alpha1.ResumeOp) {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.ResumeOp, am.Status.State))
		return
	}

	var appCfg *appcfg.ApplicationConfig
	if err = json.Unmarshal([]byte(am.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		api.HandleError(resp, req, err)
		return
	}

	// Pre-side-effect 1: unified resume-time resource gate. See
	// validation.ResumePressureValidators for the specific rules.
	// Failure returns a 400 with the RequirementResp envelope
	// (legacy shape preserved so existing clients keep parsing).
	decision, err := validation.Run(req.Request.Context(), validation.Input{
		Client:    h.ctrlClient,
		AppConfig: appCfg,
		Op:        v1alpha1.ResumeOp,
		Token:     token,
	}, validation.ResumePressureValidators()...)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if !decision.OK {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: decision.Resource.String(),
			Message:  decision.Message,
			Reason:   decision.Reason.String(),
		})
		return
	}

	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	includeSharedServer, err := compute.ShouldIncludeSharedServerForResume(req.Request.Context(), h.ctrlClient, appCfg, isAdmin)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// Pre-side-effect 2: MUTATING. Legacy handled the cleanup-on-
	// downstream-failure via defer + bindingAccepted bool; we
	// preserve that shape in the outer shell so
	// deliverResumeV2 stays a straightforward translator.
	bindingResult, err := compute.ApplyBindingSelection(req.Request.Context(), h.ctrlClient, appCfg, request.ComputeBinding, includeSharedServer)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	switch bindingResult.Status {
	case compute.BindingApplyStatusRequired:
		api.HandleFailedCheck(resp, api.CheckTypeComputeBindingRequired, ComputeBindingFailedCheck{
			Availability: bindingResult.Availability,
			Validation:   bindingResult.Validation,
		})
		return
	case compute.BindingApplyStatusUnavailable:
		api.HandleFailedCheck(resp, api.CheckTypeComputeBindingUnavailable, ComputeBindingFailedCheck{
			Availability: bindingResult.Availability,
			Validation:   bindingResult.Validation,
		})
		return
	case compute.BindingApplyStatusApplied, compute.BindingApplyStatusNotRequired:
	}

	// bindingAccepted mirrors legacy: only Applied is a mutation
	// that needs undo-on-failure; the other outcomes either
	// returned early above (Required/Unavailable) or are no-ops
	// (NotRequired). See legacy comment for why the flag is
	// inverted with the (bindingResult.Status != Applied) trick;
	// we make it explicit to leave less room for mis-reading.
	bindingAccepted := bindingResult.Status != compute.BindingApplyStatusApplied
	defer func() {
		if bindingAccepted {
			return
		}
		targetApp, targetOwner := bindingResult.TargetApp, bindingResult.TargetOwner
		if targetApp == "" {
			targetApp = appCfg.AppName
		}
		if targetOwner == "" {
			targetOwner = appCfg.OwnerName
		}
		if cleanupErr := compute.DeleteAllocationsForApp(req.Request.Context(), h.ctrlClient, targetApp, targetOwner); cleanupErr != nil {
			klog.Warningf("cleanup compute allocation for failed resume %s failed: %v", appCfg.AppName, cleanupErr)
		}
	}()

	if err := h.deliverResumeV2(req.Request.Context(), req, resp, name, amPtr, app, includeSharedServer); err != nil {
		// Leave bindingAccepted=false so the deferred cleanup
		// undoes ApplyBindingSelection.
		return
	}
	bindingAccepted = true
}

// deliverResumeV2 is the pure translator step. Extracted so unit
// tests can exercise it without stubbing validation.Run /
// compute.ShouldIncludeSharedServerForResume /
// compute.ApplyBindingSelection.
//
// Returns an error (not just writes to resp) so the outer shell
// can decide whether to run the deferred compute-cleanup. Nil on
// success; non-nil on Patch or (defensive) AllowsOp
// failure.
func (h *Handler) deliverResumeV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *v1alpha1.ApplicationManager,
	app string,
	includeSharedServer bool,
) error {
	am := *amPtr

	// Defensive re-check: the outer shell already validated the
	// state, but the load-through-patch window could have raced
	// with an unrelated state change. Duplicating the guard here
	// is cheap and keeps deliverResumeV2 usable as a stand-alone
	// unit-test target.
	if blockIfRequestPending(resp, req, &am, v1alpha1.ResumeOp) {
		return explainRequestPending(am.Spec.OpType, v1alpha1.ResumeOp)
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, v1alpha1.ResumeOp) {
		err := fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.ResumeOp, am.Status.State)
		api.HandleBadRequest(resp, req, err)
		return err
	}

	patched := am.DeepCopy()
	patched.Spec.OpType = v1alpha1.ResumeOp
	if patched.Annotations == nil {
		patched.Annotations = make(map[string]string)
	}
	patched.Annotations[api.AppResumeAllKey] = fmt.Sprintf("%t", includeSharedServer)
	var opID string
	patched.Annotations, opID = stampDurableOperationRequest(patched.Annotations)
	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return err
	}

	h.opDispatcher.Trigger(name, domain.TriggerOp{Op: v1alpha1.ResumeOp})

	resp.WriteAsJson(api.InstallationResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
	return nil
}

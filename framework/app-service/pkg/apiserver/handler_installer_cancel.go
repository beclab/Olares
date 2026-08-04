package apiserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) cancel(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	// type = timeout | operate
	cancelType := req.QueryParameter("type")
	if cancelType == "" {
		cancelType = "operate"
	}

	name, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}
	am := *amPtr
	state := am.Status.State
	if !appstate.IsOperationAllowed(state, v1alpha1.CancelOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.CancelOp, am.Status.State))

		return
	}
	var cancelState v1alpha1.ApplicationManagerState
	switch state {
	case v1alpha1.Pending, v1alpha1.PendingCancelFailed:
		cancelState = v1alpha1.PendingCanceling
	case v1alpha1.Downloading, v1alpha1.DownloadingCancelFailed:
		cancelState = v1alpha1.DownloadingCanceling
	case v1alpha1.Installing, v1alpha1.InstallingCancelFailed:
		cancelState = v1alpha1.InstallingCanceling
	case v1alpha1.Initializing:
		cancelState = v1alpha1.InitializingCanceling
	case v1alpha1.Resuming:
		cancelState = v1alpha1.ResumingCanceling
	case v1alpha1.Upgrading:
		cancelState = v1alpha1.UpgradingCanceling
	case v1alpha1.ApplyingEnv:
		cancelState = v1alpha1.ApplyingEnvCanceling
	}
	opID := strconv.FormatInt(time.Now().Unix(), 10)
	am.Spec.OpType = v1alpha1.CancelOp
	err := h.ctrlClient.Update(req.Request.Context(), &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	now := metav1.Now()
	cancelMsg, cancelReason := cancelStatus(cancelType, cancelState)
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.CancelOp,
		OpID:       opID,
		LastState:  am.Status.LastState,
		State:      cancelState,
		Progress:   "0.00",
		Message:    cancelMsg,
		Reason:     cancelReason,
		StatusTime: &now,
		UpdateTime: &now,
	}
	_, err = apputils.UpdateAppMgrStatus(name, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteAsJson(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

// cancelStatus maps (cancelType, targetCancelingState) to the
// (Message, Reason) tuple written to ApplicationManagerStatus when
// handling POST /apps/{name}/cancel. It mirrors the per-operation
// wording the reconcile-driven *App.Cancel() path writes via
// constants.<X>CanceledByTimeout / <X>CancelBySystem, so downstream
// consumers see consistent vocabulary regardless of how the cancel was
// triggered, and can branch on Reason (a stable camelCase tag) without
// parsing Message (a human-readable sentence).
//
// cancelType comes straight from the ?type= query parameter. The
// handler normalises empty -> "operate" before calling; anything other
// than the documented "timeout" value (including the canonical
// "operate") maps to the *ByUser variants.
//
// Note: the Reason written here is also what propagates into the
// subsequent Stopping state — baseStatefulApp.finishCancelToStopping
// passes reason="" to updateStatus, whose preserve-on-empty semantics
// then carry this same tag forward. That's intentional: the (state,
// reason) tuple on Stopping should still reveal which kind of cancel
// led there, not just the generic "stopping" state name.
//
// Test invariant: controllers/state_flow_lineages_test.go's
// mirrorCancelStatus MUST stay in lockstep with this switch — drift
// will surface as a dump mismatch in TestStateFlow_Lineages_Payloads.
func cancelStatus(cancelType string, cancelState v1alpha1.ApplicationManagerState) (message, reason string) {
	isTimeout := cancelType == "timeout"
	switch cancelState {
	case v1alpha1.PendingCanceling, v1alpha1.InstallingCanceling:
		if isTimeout {
			return constants.InstallCanceledByTimeout, constants.InstallCancelBySystem
		}
		return constants.InstallCanceledByUser, constants.InstallCancelByUser
	case v1alpha1.DownloadingCanceling:
		if isTimeout {
			return constants.DownloadCanceledByTimeout, constants.DownloadCancelBySystem
		}
		return constants.DownloadCanceledByUser, constants.DownloadCancelByUser
	case v1alpha1.InitializingCanceling:
		if isTimeout {
			return constants.InitializeCanceledByTimeout, constants.InitializeCancelBySystem
		}
		return constants.InitializeCanceledByUser, constants.InitializeCancelByUser
	case v1alpha1.ResumingCanceling:
		if isTimeout {
			return constants.ResumeCanceledByTimeout, constants.ResumeCancelBySystem
		}
		return constants.ResumeCanceledByUser, constants.ResumeCancelByUser
	case v1alpha1.UpgradingCanceling:
		if isTimeout {
			return constants.UpgradeCanceledByTimeout, constants.UpgradeCancelBySystem
		}
		return constants.UpgradeCanceledByUser, constants.UpgradeCancelByUser
	case v1alpha1.ApplyingEnvCanceling:
		if isTimeout {
			return constants.ApplyEnvCanceledByTimeout, constants.ApplyEnvCancelBySystem
		}
		return constants.ApplyEnvCanceledByUser, constants.ApplyEnvCancelByUser
	}
	// Defensive: unreachable today because the calling switch on the
	// source state only produces the seven *Canceling values handled
	// above. If a new cancelable state is added, fall back to the raw
	// cancelType / state name so the AM still gets *some* message and
	// reason while the test suite surfaces the gap.
	return cancelType, cancelState.String()
}

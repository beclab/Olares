package apiserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (h *Handler) cancel(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	// type = timeout | operate
	cancelType := req.QueryParameter("type")
	if cancelType == "" {
		cancelType = "operate"
	}

	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var am v1alpha1.ApplicationManager
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
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
	err = h.ctrlClient.Update(req.Request.Context(), &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	now := metav1.Now()
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.CancelOp,
		OpID:       opID,
		LastState:  am.Status.LastState,
		State:      cancelState,
		Progress:   "0.00",
		Message:    cancelType,
		StatusTime: &now,
		UpdateTime: &now,
	}
	if err != nil {
		api.HandleError(resp, req, err)
		return
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

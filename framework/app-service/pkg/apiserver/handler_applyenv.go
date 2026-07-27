package apiserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) appApplyEnv(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	appMgrName, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}
	appMgr := *amPtr

	if !appstate.IsOperationAllowed(appMgr.Status.State, appv1alpha1.ApplyEnvOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", appv1alpha1.ApplyEnvOp, appMgr.Status.State))
		return
	}

	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	appCopy := appMgr.DeepCopy()
	appCopy.Spec.OpType = appv1alpha1.ApplyEnvOp
	if appCopy.Annotations == nil {
		klog.Errorf("not support operation %s,name:%s", appv1alpha1.ApplyEnvOp, appCopy.Spec.AppName)
		api.HandleError(resp, req, fmt.Errorf("not support operation %s", appv1alpha1.ApplyEnvOp))
		return
	}
	appCopy.Annotations[api.AppTokenKey] = token
	// Refresh the pre-op state on every applyEnv so the applyEnv flow
	// (applying_env_app.go) sees the actual state right before this operation.
	// Without this, a stale Stopped value left by a previous applyEnv/upgrade
	// would make a Running app incorrectly land back in Stopped.
	appCopy.Annotations[api.AppPreUpgradeStateKey] = string(appMgr.Status.State)

	err = h.ctrlClient.Patch(req.Request.Context(), appCopy, client.MergeFrom(&appMgr))
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)

	status := appv1alpha1.ApplicationManagerStatus{
		OpType:     appv1alpha1.ApplyEnvOp,
		OpID:       opID,
		State:      appv1alpha1.ApplyingEnv,
		Message:    "waiting for applying env",
		Reason:     appv1alpha1.ApplyingEnv.String(),
		StatusTime: &now,
		UpdateTime: &now,
	}

	_, err = apputils.UpdateAppMgrStatus(appMgrName, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

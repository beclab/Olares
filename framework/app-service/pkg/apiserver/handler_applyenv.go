package apiserver

import (
	"fmt"
	"strconv"
	"time"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/appstate"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) appApplyEnv(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var appMgr appv1alpha1.ApplicationManager
	appMgrName, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: appMgrName}, &appMgr)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if !appstate.IsOperationAllowed(appMgr.Status.State, appv1alpha1.ApplyEnvOp) {
		err = fmt.Errorf("%s operation is not allowed for %s state", appv1alpha1.ApplyEnvOp, appMgr.Status.State)
		api.HandleBadRequest(resp, req, err)
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
	appCopy.Annotations[api.AppTokenKey] = token

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
		StatusTime: &now,
		UpdateTime: &now,
	}

	am, err := apputils.UpdateAppMgrStatus(appMgrName, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishAppEvent(utils.EventParams{
		Owner:      am.Spec.AppOwner,
		Name:       am.Spec.AppName,
		OpType:     string(am.Status.OpType),
		OpID:       opID,
		State:      appv1alpha1.ApplyingEnv.String(),
		RawAppName: am.Spec.RawAppName,
		Type:       "app",
		Title:      apputils.AppTitle(am.Spec.Config),
	})

	resp.WriteEntity(api.Response{Code: 200})
}

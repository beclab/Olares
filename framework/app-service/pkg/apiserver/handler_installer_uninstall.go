package apiserver

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/appstate"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/users/userspace"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (h *Handler) uninstall(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	var err error
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	request := &api.UninstallRequest{}
	if req.Request.ContentLength > 0 {
		err := req.ReadEntity(request)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}

	if userspace.IsSysApp(app) {
		api.HandleBadRequest(resp, req, errors.New("sys app can not be uninstall"))
		return
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

	//var application v1alpha1.Application
	//err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &application)
	//if err != nil {
	//	api.HandleError(resp, req, err)
	//	return
	//}
	//if application.Spec.IsSysApp {
	//	api.HandleBadRequest(resp, req, errors.New("can not uninstall sys app"))
	//	return
	//}
	if !appstate.IsOperationAllowed(am.Status.State, v1alpha1.UninstallOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.UninstallOp, am.Status.State))
		return
	}
	am.Spec.OpType = v1alpha1.UninstallOp
	am.Annotations[api.AppTokenKey] = token
	am.Annotations[api.AppUninstallAllKey] = fmt.Sprintf("%t", request.All)
	err = h.ctrlClient.Update(req.Request.Context(), &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.UninstallOp,
		State:      v1alpha1.Uninstalling,
		OpID:       opID,
		Progress:   "0.00",
		StatusTime: &now,
		UpdateTime: &now,
	}

	a, err := apputils.UpdateAppMgrStatus(name, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishAppEvent(utils.EventParams{
		Owner:      a.Spec.AppOwner,
		Name:       a.Spec.AppName,
		OpType:     string(a.Status.OpType),
		OpID:       opID,
		State:      v1alpha1.Uninstalling.String(),
		RawAppName: a.Spec.RawAppName,
		Type:       "app",
		Title:      apputils.AppTitle(a.Spec.Config),
	})

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

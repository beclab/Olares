package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (h *Handler) suspend(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	// read optional body to support all=true
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
	if !appstate.IsOperationAllowed(am.Status.State, v1alpha1.StopOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.StopOp, am.Status.State))
		return
	}
	am.Spec.OpType = v1alpha1.StopOp
	am.Annotations[api.AppStopAllKey] = fmt.Sprintf("%t", request.All)

	err = h.ctrlClient.Update(req.Request.Context(), &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	opID := strconv.FormatInt(time.Now().Unix(), 10)

	now := metav1.Now()
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.StopOp,
		OpID:       opID,
		State:      v1alpha1.Stopping,
		Reason:     constants.AppStopByUser,
		Message:    fmt.Sprintf("app %s was stop by user %s", am.Spec.AppName, am.Spec.AppOwner),
		StatusTime: &now,
		UpdateTime: &now,
	}
	_, err = apputils.UpdateAppMgrStatus(name, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

func (h *Handler) resume(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
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
	if !appstate.IsOperationAllowed(am.Status.State, v1alpha1.ResumeOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.ResumeOp, am.Status.State))
		return
	}
	var appCfg *appcfg.ApplicationConfig
	err = json.Unmarshal([]byte(am.Spec.Config), &appCfg)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		api.HandleError(resp, req, err)
		return
	}

	resourceType, resourceConditionType, err := apputils.CheckAppRequirement(token, appCfg, v1alpha1.ResumeOp)
	if err != nil {
		klog.Errorf("Failed to check app requirement err=%v", err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: resourceType.String(),
			Message:  err.Error(),
			Reason:   resourceConditionType.String(),
		})
		return
	}

	resourceType, resourceConditionType, err = apputils.CheckUserResRequirement(req.Request.Context(), appCfg, v1alpha1.ResumeOp)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: resourceType.String(),
			Message:  err.Error(),
			Reason:   resourceConditionType.String(),
		})
		return
	}

	am.Spec.OpType = v1alpha1.ResumeOp
	// if current user is admin, also resume server side
	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	am.Annotations[api.AppResumeAllKey] = fmt.Sprintf("%t", false)
	if isAdmin {
		am.Annotations[api.AppResumeAllKey] = fmt.Sprintf("%t", true)
	}
	err = h.ctrlClient.Update(req.Request.Context(), &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.ResumeOp,
		OpID:       opID,
		State:      v1alpha1.Resuming,
		Message:    fmt.Sprintf("app %s was resume by user %s", am.Spec.AppName, am.Spec.AppOwner),
		StatusTime: &now,
		UpdateTime: &now,
	}
	_, err = apputils.UpdateAppMgrStatus(name, status)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

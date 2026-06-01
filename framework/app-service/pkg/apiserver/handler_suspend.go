package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	name, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}
	am := *amPtr
	if !appstate.IsOperationAllowed(am.Status.State, v1alpha1.StopOp) {
		api.HandleBadRequest(resp, req, fmt.Errorf("%s operation is not allowed for %s state", v1alpha1.StopOp, am.Status.State))
		return
	}
	am.Spec.OpType = v1alpha1.StopOp
	if am.Annotations == nil {
		am.Annotations = make(map[string]string)
	}
	am.Annotations[api.AppStopAllKey] = fmt.Sprintf("%t", request.All)

	err := h.ctrlClient.Update(req.Request.Context(), &am)
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

	// Unified resume-time resource gate: cluster pressure + k8s request
	// capacity + user quota. Compute mode and per-node pressure are
	// skipped (resume reuses the binding chosen at install time).
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

	// if current user is admin, also resume server side
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
	case compute.BindingApplyStatusApplied:
	case compute.BindingApplyStatusNotRequired:
	}

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

	am.Spec.OpType = v1alpha1.ResumeOp
	if am.Annotations == nil {
		am.Annotations = map[string]string{}
	}
	am.Annotations[api.AppResumeAllKey] = fmt.Sprintf("%t", false)
	if includeSharedServer {
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

	bindingAccepted = true

	resp.WriteAsJson(api.InstallationResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

type ResumeRequest struct {
	ComputeBinding []compute.BindingSelection `json:"computeBinding,omitempty"`
}

package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComputeResourcesResponse struct {
	api.Response `json:",inline"`
	Data         []compute.Node `json:"data"`
}

type ComputeBindingResponse struct {
	api.Response `json:",inline"`
	Data         []compute.Allocation `json:"data"`
}

// ComputeBindingValidationResponse is the 200 success payload returned by
// the read-only compute binding validation endpoint when the binding is
// acceptable (BindingApplyStatusValid or BindingApplyStatusNotRequired).
// The Required / Unavailable cases instead return a FailedCheckResponse
// carrying ComputeBindingFailedCheck, matching the resume endpoint.
type ComputeBindingValidationResponse struct {
	api.Response `json:",inline"`
	Data         ComputeBindingValidationData `json:"data"`
}

type ComputeBindingValidationData struct {
	Status       string                           `json:"status"`
	Availability *compute.AvailabilityResult      `json:"availability,omitempty"`
	Validation   *compute.BindingValidationResult `json:"validation,omitempty"`
	Allocations  []compute.Allocation             `json:"allocations,omitempty"`
}

// ComputeBindingFailedCheck is the payload returned by the resume endpoint
// under FailedCheckResponse when the caller needs to (re)select a compute
// binding before the resume can proceed. CheckTypeComputeBindingRequired
// signals a binding is required and Validation is nil;
// CheckTypeComputeBindingUnavailable signals the caller-supplied binding
// could not be satisfied and Validation explains why.
type ComputeBindingFailedCheck struct {
	Availability *compute.AvailabilityResult      `json:"availability"`
	Validation   *compute.BindingValidationResult `json:"validation,omitempty"`
}

type UpdateDeviceSupportTypeRequest struct {
	SupportType string `json:"supportType"`
}

type DeviceSupportTypeSwitchResponse struct {
	api.Response `json:",inline"`
	Data         DeviceSupportTypeSwitchResult `json:"data"`
}

type DeviceSupportTypeSwitchResult struct {
	Status      string            `json:"status"`
	Device      compute.Device    `json:"device"`
	StoppedApps []StoppedBoundApp `json:"stoppedApps,omitempty"`
	BlockedApps []BlockedBoundApp `json:"blockedApps,omitempty"`
}

type StoppedBoundApp struct {
	AppName string `json:"appName"`
	Owner   string `json:"owner"`
	State   string `json:"state"`
}

type BlockedBoundApp struct {
	AppName string `json:"appName"`
	Owner   string `json:"owner"`
	State   string `json:"state"`
	Reason  string `json:"reason"`
}

func (h *Handler) updateDeviceSupportType(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()

	owner := req.Attribute(constants.UserContextAttribute).(string)
	isAdmin, err := kubesphere.IsAdmin(ctx, h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if !isAdmin {
		api.HandleBadRequest(resp, req, fmt.Errorf("only admin user can switch compute device mode"))
		return
	}

	nodeName := req.PathParameter(ParamNodeName)
	deviceID := req.PathParameter(ParamDeviceID)
	body := &UpdateDeviceSupportTypeRequest{}
	if err := req.ReadEntity(body); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if body.SupportType == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("supportType is required"))
		return
	}

	_, device, err := compute.FindDevice(ctx, h.ctrlClient, nodeName, deviceID)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if len(device.AvailableSupportTypes) <= 1 {
		api.HandleBadRequest(resp, req, fmt.Errorf("device mode switching is not supported for gpu type %s", device.Mode))
		return
	}
	if !compute.SupportTypeAvailable(device.AvailableSupportTypes, body.SupportType) {
		api.HandleBadRequest(resp, req, fmt.Errorf("support type %s is not available for gpu type %s", body.SupportType, device.Mode))
		return
	}
	if device.SupportType == body.SupportType {
		resp.WriteAsJson(DeviceSupportTypeSwitchResponse{
			Response: api.Response{Code: api.CodeSuccess},
			Data: DeviceSupportTypeSwitchResult{
				Status: "unchanged",
				Device: device,
			},
		})
		return
	}

	// Plan stops for every distinct app currently bound to the device. Bail
	// out without touching anything if any of them can't be stopped right now.
	plan, err := h.planStopForBoundApps(ctx, device.Bindings)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if len(plan.blocked) > 0 {
		api.HandleFailedCheck(resp, api.CheckTypeComputeDeviceSwitchBlocked, DeviceSupportTypeSwitchResult{
			Status:      "bound-apps-stop-blocked",
			Device:      device,
			BlockedApps: plan.blocked,
		})
		return
	}

	// Submit StopOp for each bound app and tear down its allocation/binding
	// immediately so the device is free even before the SuspendingApp finishes
	// its state transition.
	stopped, err := h.commitStopForBoundApps(ctx, plan.stoppable)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if err := compute.SwitchDeviceMode(ctx, h.ctrlClient, nodeName, deviceID, body.SupportType); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	device.SupportType = body.SupportType
	device.Bindings = nil
	resp.WriteAsJson(DeviceSupportTypeSwitchResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data: DeviceSupportTypeSwitchResult{
			Status:      "switched",
			Device:      device,
			StoppedApps: stopped,
		},
	})
}

type stoppableApp struct {
	appName              string
	owner                string
	managerName          string
	manager              *appv1alpha1.ApplicationManager
	managesSharedServer  bool
	previousAppMgrStatus appv1alpha1.ApplicationManagerState
}

type stopPlan struct {
	stoppable []stoppableApp
	blocked   []BlockedBoundApp
}

type boundStopClass int

const (
	boundStopSubmit boundStopClass = iota
	boundStopAlready
	boundStopBlocked
)

// classifyBoundStop decides whether a bound app can take a new StopOp.
// Status lags Spec until the Runner copies the request id, so an
// unconsumed Upgrade, Uninstall, or ApplyEnv blocks the device switch
// instead of being overwritten.
func classifyBoundStop(am *appv1alpha1.ApplicationManager) (boundStopClass, string) {
	if am == nil {
		return boundStopBlocked, "stop-operation-not-allowed"
	}
	if am.Status.State == appv1alpha1.Stopping || am.Status.State == appv1alpha1.Stopped ||
		domain.OperationRequestPending(am, appv1alpha1.StopOp) {
		return boundStopAlready, ""
	}
	if _, pending := domain.UnconsumedOperation(am); pending {
		return boundStopBlocked, "unconsumed-operation"
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, appv1alpha1.StopOp) {
		return boundStopBlocked, "stop-operation-not-allowed"
	}
	return boundStopSubmit, ""
}

// planStopForBoundApps inspects each unique (appName, owner) pair from
// `bindings` and classifies it as either stoppable or blocked. It does not
// mutate any cluster state, so callers can safely abort if anything is
// blocked.
func (h *Handler) planStopForBoundApps(ctx context.Context, bindings []compute.Allocation) (stopPlan, error) {
	plan := stopPlan{}
	seen := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		key := binding.AppName + "/" + binding.Owner
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		// ResolveAppMgrName handles both v1/v2 (FmtAppMgrName style) and v3
		// (cluster-wide '{app}-shared-{app}' AM). A v3 app whose GPU
		// binding is being torn down also goes through here.
		name, _, err := apputils.ResolveAppMgrName(ctx, binding.AppName, binding.Owner)
		if err != nil {
			return plan, err
		}
		var am appv1alpha1.ApplicationManager
		if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: name}, &am); err != nil {
			return plan, err
		}

		switch class, reason := classifyBoundStop(&am); class {
		case boundStopAlready:
			// Stopping / Stopped, or a StopOp accepted onto Spec but not
			// yet copied into Status. The Runner will move it, so do not
			// submit another. Allocations are still cleared below.
			plan.stoppable = append(plan.stoppable, stoppableApp{
				appName:              binding.AppName,
				owner:                binding.Owner,
				managerName:          name,
				manager:              am.DeepCopy(),
				previousAppMgrStatus: am.Status.State,
			})
			continue
		case boundStopBlocked:
			// An unconsumed Upgrade, Uninstall, or ApplyEnv still shows
			// Status as Running. Refuse the switch instead of erasing it.
			plan.blocked = append(plan.blocked, BlockedBoundApp{
				AppName: binding.AppName,
				Owner:   binding.Owner,
				State:   am.Status.State.String(),
				Reason:  reason,
			})
			continue
		}

		var appCfg appcfg.ApplicationConfig
		if err := appcfg.GetAppConfig(&am, &appCfg); err != nil {
			return plan, err
		}
		managesSharedServer, err := compute.ManagesSharedServer(ctx, h.ctrlClient, &appCfg)
		if err != nil {
			return plan, err
		}
		plan.stoppable = append(plan.stoppable, stoppableApp{
			appName:              binding.AppName,
			owner:                binding.Owner,
			managerName:          name,
			manager:              am.DeepCopy(),
			managesSharedServer:  managesSharedServer,
			previousAppMgrStatus: am.Status.State,
		})
	}
	return plan, nil
}

// commitStopForBoundApps submits a StopOp for each app that isn't already
// stopping/stopped and clears the compute allocation/HAMi binding for every
// app in the plan.
func (h *Handler) commitStopForBoundApps(ctx context.Context, apps []stoppableApp) ([]StoppedBoundApp, error) {
	stopped := make([]StoppedBoundApp, 0, len(apps))
	for _, a := range apps {
		class, _ := classifyBoundStop(a.manager)
		if class == boundStopBlocked {
			return stopped, fmt.Errorf("compute device stop: app %s owner %s has an unconsumed operation", a.appName, a.owner)
		}
		if class == boundStopSubmit {
			if h.opDispatcher == nil {
				return stopped, fmt.Errorf("compute device stop: opDispatcher is not wired")
			}
			base := a.manager.DeepCopy()
			patched := base.DeepCopy()
			if patched.Annotations == nil {
				patched.Annotations = make(map[string]string)
			}
			patched.Spec.OpType = appv1alpha1.StopOp
			patched.Annotations[api.AppStopAllKey] = fmt.Sprintf("%t", a.managesSharedServer)
			patched.Annotations[api.AppStopReasonKey] = constants.AppStopByUser
			patched.Annotations[api.AppStopMessageKey] = fmt.Sprintf("app %s was stopped before compute device mode switch", a.appName)
			patched.Annotations, _ = stampDurableOperationRequest(patched.Annotations)
			if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(base)); err != nil {
				return stopped, err
			}
			h.opDispatcher.Trigger(a.managerName, domain.TriggerOp{Op: appv1alpha1.StopOp})
		}

		// SuspendingApp will also call DeleteAllocationsForComputeTarget once
		// it finishes; doing it here is idempotent but lets us flip the
		// device's support type immediately.
		if err := compute.DeleteAllocationsForApp(ctx, h.ctrlClient, a.appName, a.owner); err != nil {
			return stopped, err
		}
		stopped = append(stopped, StoppedBoundApp{
			AppName: a.appName,
			Owner:   a.owner,
			State:   appv1alpha1.Stopping.String(),
		})
	}
	return stopped, nil
}

func (h *Handler) listComputeResources(req *restful.Request, resp *restful.Response) {
	nodes, err := compute.FetchNodeComputeAllocations(req.Request.Context(), h.ctrlClient)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if err := compute.AttachBoundAppSpecs(req.Request.Context(), h.ctrlClient, nodes); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(ComputeResourcesResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data:     nodes,
	})
}

func (h *Handler) listComputeBindings(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	appCfg, err := h.installedComputeAppConfig(req)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	allocations, err := compute.FindAllocationsForApp(req.Request.Context(), h.ctrlClient, appName, appCfg.OwnerName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(ComputeBindingResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data:     allocations,
	})
}

func (h *Handler) validateComputeBinding(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	request := &ResumeRequest{}
	if err := readOptionalEntity(req, request); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	_, amPtr, ok := h.loadAuthorizedLifecycleAM(req.Request.Context(), req, resp, app, owner)
	if !ok {
		return
	}
	am := *amPtr

	var appCfg *appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(am.Spec.Config), &appCfg); err != nil {
		api.HandleError(resp, req, err)
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

	bindingResult, err := compute.ValidateBindingForResume(req.Request.Context(), h.ctrlClient, appCfg, request.ComputeBinding, includeSharedServer)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// Required / Unavailable still report through the same FailedCheckResponse
	// envelope the resume endpoint uses so the frontend can reuse identical
	// handling. Valid / NotRequired mean the binding is acceptable, so return
	// a plain 200 success carrying the resolved status plus the available
	// options for context.
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
	}

	resp.WriteAsJson(ComputeBindingValidationResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data: ComputeBindingValidationData{
			Status:       bindingResult.Status,
			Availability: bindingResult.Availability,
			Validation:   bindingResult.Validation,
			Allocations:  bindingResult.Allocations,
		},
	})
}

func (h *Handler) installedComputeAppConfig(req *restful.Request) (*appcfg.ApplicationConfig, error) {
	appName := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	name, _, err := apputils.ResolveAppMgrName(req.Request.Context(), appName, owner)
	if err != nil {
		return nil, err
	}
	var manager appv1alpha1.ApplicationManager
	if err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &manager); err != nil {
		return nil, err
	}
	var cfg appcfg.ApplicationConfig
	if err := appcfg.GetAppConfig(&manager, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func readOptionalEntity(req *restful.Request, into any) error {
	if err := req.ReadEntity(into); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

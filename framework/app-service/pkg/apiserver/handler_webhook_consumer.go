package apiserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// consumerGPUPatch injects the GPU resources of the consumer named after the
// workload. inject is false when the workload holds no cards, so the caller
// strips any GPU keys left from an earlier render.
func (h *Handler) consumerGPUPatch(ctx context.Context, workload string, tpl *corev1.PodTemplateSpec, replicas *int32, appCfg *appcfg.ApplicationConfig) ([]byte, bool, error) {
	consumer, ok := appCfg.GPUConsumer(workload)
	if !ok {
		return nil, false, nil
	}
	mode, ok := appCfg.SelectedConsumerMode(consumer)
	if !ok || mode.Mode == utils.CPUType {
		return nil, false, nil
	}
	req := compute.ConsumerRequirement(consumer, mode)
	allocations, err := compute.FindAllocationsForApp(ctx, h.ctrlClient, appCfg.AppName, appCfg.OwnerName)
	if err != nil {
		return nil, false, fmt.Errorf("find compute allocation for app %s: %w", appCfg.AppName, err)
	}
	var rows []compute.Allocation
	nodes := map[string]struct{}{}
	for _, allocation := range allocations {
		if allocation.Consumer == consumer.Name {
			rows = append(rows, allocation)
			nodes[allocation.NodeName] = struct{}{}
		}
	}

	capability := compute.CapabilityFor(req.Mode)
	gpu := webhook.GPUResources{
		MemoryResource: capability.MemoryResource,
		RuntimeClass:   capability.RuntimeClass,
	}
	if gpu.Resource, err = h.resolveIntelDriverResource(ctx, req.Mode); err != nil {
		return nil, false, err
	}
	if k := consumer.CardsPerPod(len(rows)); k > 1 {
		gpu.Count = fmt.Sprint(k)
	}
	if capability.HAMI && req.RequiredGPU > 0 {
		perCard := req.RequiredGPU
		if consumer.PerPod != appcfg.PerPodAll && len(rows) > 0 {
			perCard = (req.RequiredGPU + int64(len(rows)) - 1) / int64(len(rows))
		}
		gpu.Memory = ptr.To(resource.NewQuantity(perCard/1024/1024, resource.DecimalSI).String())
	}
	envs := []webhook.EnvKeyValue{{Key: constants.EnvGPUType, Value: req.Mode}}
	patchBytes, err := webhook.CreatePatchForDeployment(tpl, len(consumer.Containers) == 0, consumer.Containers, gpu, envs)
	if err != nil {
		return nil, false, err
	}

	if !compute.UsesGPUExtendedResource(req.Mode) && len(rows) > 0 {
		if existing := tpl.Spec.NodeSelector[corev1.LabelHostname]; existing != "" && existing != rows[0].NodeName {
			return nil, false, fmt.Errorf("attachment node %s conflicts with compute allocation node %s", existing, rows[0].NodeName)
		}
		if patchBytes, err = appendNodeSelectorPatch(patchBytes, tpl, rows[0].NodeName); err != nil {
			return nil, false, err
		}
	}

	// Zero replicas is the install gate or a suspended app; leave it alone.
	if replicas != nil && *replicas > 0 && len(rows) > 0 {
		want := consumer.PodReplicas(len(rows), len(nodes))
		if want > 0 && want != *replicas {
			if patchBytes, err = appendJSONPatch(patchBytes, map[string]any{
				"op": constants.PatchOpReplace, "path": "/spec/replicas", "value": want,
			}); err != nil {
				return nil, false, err
			}
		}
	}
	return patchBytes, true, nil
}

func appendJSONPatch(patchBytes []byte, op map[string]any) ([]byte, error) {
	var patches []map[string]any
	if len(patchBytes) > 0 {
		if err := json.Unmarshal(patchBytes, &patches); err != nil {
			return nil, err
		}
	}
	return json.Marshal(append(patches, op))
}

// consumerPodLabels returns the labels that let HAMI match a consumer's pods
// to its bindings, or nil when the workload holds no cards.
func consumerPodLabels(appCfg *appcfg.ApplicationConfig, workload string) map[string]string {
	consumer, ok := appCfg.GPUConsumer(workload)
	if !ok {
		return nil
	}
	if mode, ok := appCfg.SelectedConsumerMode(consumer); !ok || mode.Mode == utils.CPUType {
		return nil
	}
	policy := "single"
	switch consumer.PerPod {
	case appcfg.PerPodAll:
		policy = "all"
	case appcfg.PerPodSplit:
		policy = "split"
	}
	return map[string]string{
		constants.AppGPUConsumerLabel:    consumer.Name,
		constants.AppPodGPUConsumePolicy: policy,
	}
}

// ConsumerModeSelectCheck asks the caller to choose a mode for the listed
// consumers and resubmit with selectedGpuTypes.
type ConsumerModeSelectCheck struct {
	Consumers []compute.ConsumerPlan `json:"consumers"`
	Pending   []string               `json:"pending"`
}

// selectConsumerModes fills appCfg.SelectedGpuTypes, writing the response
// itself and returning false when the install cannot go on.
func (h *Handler) selectConsumerModes(req *restful.Request, resp *restful.Response, appCfg *appcfg.ApplicationConfig, requested map[string]string) bool {
	ctx := req.Request.Context()
	if err := compute.RequireConsumerAwareHAMI(ctx, h.ctrlClient); err != nil {
		api.HandleBadRequest(resp, req, err)
		return false
	}
	chosen, pending, err := compute.AutoSelectConsumerModes(ctx, h.ctrlClient, appCfg, requested)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return false
	}
	if len(pending) > 0 {
		plans, err := compute.BuildConsumerInstallPlans(ctx, h.ctrlClient, appCfg)
		if err != nil {
			api.HandleError(resp, req, err)
			return false
		}
		api.HandleFailedCheck(resp, api.CheckTypeComputeModeSelect, ConsumerModeSelectCheck{Consumers: plans, Pending: pending})
		return false
	}
	appCfg.SelectedGpuTypes = chosen
	return true
}

// resumeBinding applies (or, with dryRun, only checks) the compute binding a
// resume body carries.
func (h *Handler) resumeBinding(ctx context.Context, appCfg *appcfg.ApplicationConfig, request *ResumeRequest, includeSharedServer, dryRun bool) (*compute.BindingApplyResult, error) {
	if len(request.ComputeBindings) == 0 {
		if dryRun {
			return compute.ValidateBindingForResume(ctx, h.ctrlClient, appCfg, request.ComputeBinding, includeSharedServer)
		}
		return compute.ApplyBindingSelection(ctx, h.ctrlClient, appCfg, request.ComputeBinding, includeSharedServer)
	}
	if !appCfg.HasGPUConsumers() {
		return nil, fmt.Errorf("computeBindings is only accepted for apps declaring workload accelerators")
	}
	if len(request.ComputeBinding) > 0 {
		return nil, fmt.Errorf("send either computeBinding or computeBindings, not both")
	}
	var result *compute.BindingApplyResult
	var err error
	if dryRun {
		result, err = compute.ValidateConsumerBindings(ctx, h.ctrlClient, appCfg, request.ComputeBindings)
	} else {
		result, err = compute.ApplyConsumerBindings(ctx, h.ctrlClient, appCfg, request.ComputeBindings)
	}
	if err != nil || result.Status != compute.BindingApplyStatusUnavailable || result.Consumer == "" {
		return result, err
	}
	if result.Availability, err = compute.ConsumerAvailability(ctx, h.ctrlClient, appCfg, result.Consumer); err != nil {
		return nil, err
	}
	return result, nil
}

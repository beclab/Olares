package compute

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildInstallComputePlan(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) ([]ModePlanResult, error) {
	targetConfig, manage, err := resolveComputeTarget(ctx, c, appConfig, false)
	if err != nil {
		return nil, err
	}
	if !manage {
		return []ModePlanResult{{ComputeType: utils.CPUType, Status: StatusInstallable}}, nil
	}
	nodes, err := FetchNodeComputeAllocations(ctx, c)
	if err != nil {
		return nil, err
	}

	return calculateInstallComputePlan(targetConfig, nodes), nil
}

func calculateInstallComputePlan(appConfig *appcfg.ApplicationConfig, nodes []Node) []ModePlanResult {
	modes := appConfig.ComputeResourceModes()
	items := make([]ModePlanResult, 0, len(modes))
	for _, mode := range modes {
		req := RequirementFromMode(mode)
		result := EvaluateInstallMode(req, nodes)
		items = append(items, result)
	}
	return items
}

func AppInstallable(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (bool, error) {
	plan, err := BuildInstallComputePlan(ctx, c, appConfig)
	if err != nil {
		return false, err
	}
	for _, result := range plan {
		if result.ComputeType == appConfig.SelectedGpuType && result.Status == StatusInstallable {
			return true, nil
		}
	}
	return false, nil
}

func AllocateForInstall(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (*Allocation, error) {
	targetConfig, manage, err := resolveComputeTarget(ctx, c, appConfig, false)
	if err != nil {
		return nil, err
	}
	if !manage {
		return nil, DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
	}
	appConfig = targetConfig
	req, ok := SelectedRequirement(appConfig)
	if !ok {
		return nil, fmt.Errorf("compute type %s not found in application resources", appConfig.SelectedGpuType)
	}
	if req.Mode == utils.CPUType {
		return nil, DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	var pickedAllocations []Allocation
	allocation, err := mutateAllocations(ctx, c, func(nodes []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		attachBindings(nodes, withoutAppAllocations(allocations, appConfig.AppName, appConfig.OwnerName))
		picked, ok := PickAllocations(appConfig, req, nodes, pressure)
		if !ok {
			return nil, nil, fmt.Errorf("no available compute resource for type %s", req.Mode)
		}
		pickedAllocations = picked
		next := replaceAppAllocations(allocations, picked)
		return next, &picked[0], nil
	})
	if err != nil {
		return nil, err
	}
	if err := deleteHAMIBindingsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName); err != nil {
		_ = DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
		return nil, err
	}
	for _, item := range pickedAllocations {
		if err := createHAMIBinding(ctx, c, item); err != nil {
			_ = DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
			return nil, err
		}
	}
	return allocation, nil
}

func RequirementFromMode(mode appcfg.ResourceMode) Requirement {
	source := &mode.ResourceRequirement
	reqCPU := parseQuantityMilli(source.RequiredCPU)
	reqGPU := parseQuantityBytes(source.RequiredGPU)
	limGPU := parseQuantityBytes(source.LimitedGPU)
	if limGPU == 0 {
		limGPU = reqGPU
	}
	reqMem := parseQuantityBytes(source.RequiredMemory)
	limMem := parseQuantityBytes(source.LimitedMemory)
	if limMem == 0 {
		limMem = reqMem
	}
	reqDisk := parseQuantityBytes(source.RequiredDisk)
	supportMultiNodes := mode.Mode == utils.NvidiaCardType && mode.SupportMultiNodes
	supportMultiCards := mode.Mode == utils.NvidiaCardType && (mode.SupportMultiCards || supportMultiNodes)
	return Requirement{
		Mode:              mode.Mode,
		RequiredCPU:       reqCPU,
		RequiredGPU:       reqGPU,
		LimitedGPU:        limGPU,
		RequiredMemory:    reqMem,
		LimitedMemory:     limMem,
		RequiredDisk:      reqDisk,
		SupportMultiCards: supportMultiCards,
		SupportMultiNodes: supportMultiNodes,
	}
}

func SelectedRequirement(appConfig *appcfg.ApplicationConfig) (Requirement, bool) {
	if appConfig == nil {
		return Requirement{}, false
	}
	mode, ok := appConfig.SelectedResourceMode()
	if !ok {
		return Requirement{}, false
	}
	return RequirementFromMode(mode), true
}

func EvaluateInstallMode(req Requirement, nodes []Node) ModePlanResult {
	if req.Mode == utils.CPUType {
		return evaluateCPUInstallMode(req, nodes)
	}
	matching := matchingNodes(req.Mode, nodes)
	if len(matching) == 0 {
		return ModePlanResult{ComputeType: req.Mode, Status: StatusNoMatchingNode, Reason: "no_matching_node"}
	}

	if installCapacityFits(req, matching) {
		return ModePlanResult{ComputeType: req.Mode, Status: StatusInstallable}
	}

	return ModePlanResult{ComputeType: req.Mode, Status: StatusInsufficientResources, Reason: "insufficient_resources"}
}

func PickAllocations(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot) ([]Allocation, bool) {
	if req.Mode == utils.CPUType {
		return nil, false
	}
	matching := matchingNodes(req.Mode, nodes)
	if req.Mode == utils.NvidiaCardType && req.SupportMultiNodes {
		return pickAggregateAllocations(appConfig, req, matching, pressure, true)
	}
	if req.Mode == utils.NvidiaCardType && req.SupportMultiCards {
		return pickAggregateAllocations(appConfig, req, matching, pressure, false)
	}
	return pickSingleAllocation(appConfig, req, matching, pressure)
}

func matchingNodes(mode string, nodes []Node) []Node {
	out := make([]Node, 0)
	for _, node := range nodes {
		if node.GPUType == mode {
			out = append(out, node)
		}
	}
	return out
}

func evaluateCPUInstallMode(req Requirement, nodes []Node) ModePlanResult {
	for _, node := range nodes {
		if node.memoryCapacity*75/100 >= req.RequiredMemory {
			return ModePlanResult{ComputeType: utils.CPUType, Status: StatusInstallable}
		}
	}
	return ModePlanResult{ComputeType: utils.CPUType, Status: StatusInsufficientResources, Reason: "insufficient_resources"}
}

func installCapacityFits(req Requirement, nodes []Node) bool {
	if req.Mode == utils.NvidiaCardType && req.SupportMultiNodes {
		var total int64
		for _, node := range nodes {
			for _, device := range node.Devices {
				total += device.Memory
			}
		}
		return total >= req.RequiredGPU
	}
	if req.Mode == utils.NvidiaCardType && req.SupportMultiCards {
		for _, node := range nodes {
			var total int64
			for _, device := range node.Devices {
				total += device.Memory
			}
			if total >= req.RequiredGPU {
				return true
			}
		}
		return false
	}
	if req.Mode == utils.NvidiaCardType {
		for _, node := range nodes {
			for _, device := range node.Devices {
				if device.Memory >= req.RequiredGPU {
					return true
				}
			}
		}
		return false
	}
	for _, node := range nodes {
		if len(node.Devices) > 0 && node.Devices[0].Memory >= req.RequiredMemory {
			return true
		}
	}
	return false
}

func pickSingleAllocation(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot) ([]Allocation, bool) {
	for _, level := range []string{FitLevelLimit, FitLevelRequired} {
		for _, supportType := range supportTypeOrder(req.Mode) {
			candidates := make([]Allocation, 0)
			for _, node := range nodes {
				for _, device := range node.Devices {
					if supportType != "" && device.SupportType != supportType {
						continue
					}
					fits, amount := deviceFitsLevel(req, node, device, pressure, level, false)
					if fits {
						assigned := requiredTargetForMode(req)
						if assigned == 0 {
							assigned = amount
						}
						candidates = append(candidates, buildAllocation(appConfig, req, node, device, assigned))
					}
				}
			}
			if len(candidates) > 0 {
				return []Allocation{candidates[rand.Intn(len(candidates))]}, true
			}
		}
	}
	return nil, false
}

func pickAggregateAllocations(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot, crossNode bool) ([]Allocation, bool) {
	for _, level := range []string{FitLevelLimit, FitLevelRequired} {
		target := targetGPU(req, level)
		if target <= 0 {
			continue
		}
		if crossNode {
			picked, ok := collectDevicesForTarget(appConfig, req, nodes, pressure, level, target, req.RequiredGPU)
			if ok {
				return picked, ok
			}
			continue
		}
		for _, node := range nodes {
			picked, ok := collectDevicesForTarget(appConfig, req, []Node{node}, pressure, level, target, req.RequiredGPU)
			if ok {
				return picked, ok
			}
		}
	}
	return nil, false
}

func collectDevicesForTarget(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot, level string, target, allocationTarget int64) ([]Allocation, bool) {
	fitRemaining := target
	allocationRemaining := allocationTarget
	var out []Allocation
	for _, supportType := range supportTypeOrder(req.Mode) {
		for _, node := range nodes {
			for _, device := range node.Devices {
				if supportType != "" && device.SupportType != supportType {
					continue
				}
				fits, amount := deviceFitsLevel(req, node, device, pressure, level, true)
				if !fits || amount <= 0 {
					continue
				}
				if allocationRemaining > 0 {
					assigned := minInt64(amount, allocationRemaining)
					out = append(out, buildAllocation(appConfig, req, node, device, assigned))
					allocationRemaining -= assigned
				}
				fitRemaining -= amount
				if fitRemaining <= 0 {
					return out, true
				}
			}
		}
	}
	return nil, false
}

func deviceFitsLevel(req Requirement, node Node, device Device, pressure PressureSnapshot, level string, allowPartial bool) (bool, int64) {
	if device.Health != "" && device.Health != deviceHealthYes {
		return false, 0
	}
	required := targetForMode(req, level)
	if required <= 0 {
		return !pressure.WouldPressure(node, AddedResources{
			CPU:    req.RequiredCPU,
			Memory: levelMemory(req, level),
			Disk:   req.RequiredDisk,
		}), 0
	}
	available := deviceAvailableMemory(device)
	if available <= 0 {
		return false, 0
	}
	if available < required && !(allowPartial && req.Mode == utils.NvidiaCardType) {
		return false, available
	}
	addedGPU := int64(0)
	if req.Mode == utils.NvidiaCardType && device.SupportType == SupportTypeTimeSlice {
		addedGPU = minInt64(available, required)
	}
	if pressure.WouldPressure(node, AddedResources{
		CPU:    req.RequiredCPU,
		Memory: levelMemory(req, level) + addedGPU,
		Disk:   req.RequiredDisk,
	}) {
		return false, available
	}
	return true, available
}

func targetForMode(req Requirement, level string) int64 {
	if req.Mode == utils.NvidiaCardType {
		return targetGPU(req, level)
	}
	return levelMemory(req, level)
}

func requiredTargetForMode(req Requirement) int64 {
	if req.Mode == utils.NvidiaCardType {
		return req.RequiredGPU
	}
	return req.RequiredMemory
}

func targetGPU(req Requirement, level string) int64 {
	if level == FitLevelLimit && req.LimitedGPU > 0 {
		return req.LimitedGPU
	}
	return req.RequiredGPU
}

func levelMemory(req Requirement, level string) int64 {
	if level == FitLevelLimit && req.LimitedMemory > 0 {
		return req.LimitedMemory
	}
	return req.RequiredMemory
}

func buildAllocation(appConfig *appcfg.ApplicationConfig, req Requirement, node Node, device Device, memory int64) Allocation {
	// In Exclusive / TimeSlice modes the pod has access to the entire
	// card (Exclusive: solo binding; TimeSlice: full memory during the
	// pod's time slice). Recording a per-pod memory amount here would
	// cap the pod via the HAMI binding's spec.memory annotation, even
	// though no slicing is happening. Persist 0 so createHAMIBinding
	// omits spec.memory and HAMI treats the pod as unrestricted. The
	// scheduler's accounting (deviceAvailableMemory / remainingMemory)
	// never reads Allocation.Memory for these modes, so this is safe.
	if isWholeCardMode(req.Mode, device.SupportType) {
		memory = 0
	}
	return Allocation{
		AppID:    appConfig.AppID,
		AppName:  appConfig.AppName,
		Owner:    appConfig.OwnerName,
		Mode:     req.Mode,
		NodeName: node.NodeName,
		DeviceID: device.ID,
		Memory:   memory,
	}
}

// isWholeCardMode reports whether binding a pod to a device with this NVIDIA
// support type grants it the entire card: Exclusive (solo binding) or
// TimeSlice (full memory during the pod's slice). buildAllocation records
// Memory=0 for these, and they are always one-binding-per-card, so allocation
// distribution must emit a separate binding for every selected card instead of
// folding several of them into a single shared VRAM budget.
func isWholeCardMode(mode, supportType string) bool {
	return mode == utils.NvidiaCardType &&
		(supportType == SupportTypeExclusive || supportType == SupportTypeTimeSlice)
}

func supportTypeOrder(mode string) []string {
	switch mode {
	case utils.NvidiaCardType:
		return []string{SupportTypeExclusive, SupportTypeMemorySlice, SupportTypeTimeSlice}
	case utils.GB10ChipType:
		// GB10 devices are decoded from the HAMI node-nvidia-register
		// annotation and carry MemorySlice (the default) or Exclusive
		// support types, matching AvailableSupportTypes(GB10ChipType).
		// MemoryShared is a cpu-only support type and never appears on a
		// GB10 device, so the default branch below would filter every
		// candidate out and make AllocateForInstall report
		// "no available compute resource for type nvidia-gb10".
		return []string{SupportTypeExclusive, SupportTypeMemorySlice}
	default:
		return []string{SupportTypeExclusive, SupportTypeMemoryShared}
	}
}

func deviceAvailableMemory(device Device) int64 {
	switch device.SupportType {
	case SupportTypeExclusive:
		if len(device.Bindings) > 0 {
			return 0
		}
		return device.Memory
	case SupportTypeMemorySlice, SupportTypeMemoryShared:
		return remainingMemory(device)
	case SupportTypeTimeSlice:
		return device.Memory
	default:
		return device.Memory
	}
}

func remainingMemory(device Device) int64 {
	remaining := device.Memory
	for _, binding := range device.Bindings {
		remaining -= binding.Memory
	}
	if remaining < 0 {
		return 0
	}
	return remaining
}

func replaceAppAllocations(allocations []Allocation, replacements []Allocation) []Allocation {
	if len(replacements) == 0 {
		return allocations
	}
	appName := replacements[0].AppName
	owner := replacements[0].Owner
	next := make([]Allocation, 0, len(allocations)+len(replacements))
	for _, existing := range allocations {
		if existing.AppName == appName && existing.Owner == owner {
			continue
		}
		next = append(next, existing)
	}
	return append(next, replacements...)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func parseQuantityBytes(value string) int64 {
	// The auto-compute sentinel means the value is resolved from the rendered
	// chart at install time. Until the install handler backfills it, treat it
	// as 0 ("no constraint") so the install-time mode feasibility gate only
	// checks architecture / mode matching for this field, not its capacity.
	if value == "" || appcfg.IsAutoResource(value) {
		return 0
	}
	q, err := resource.ParseQuantity(value)
	if err != nil {
		return 0
	}
	return q.Value()
}

func parseQuantityMilli(value string) int64 {
	if value == "" || appcfg.IsAutoResource(value) {
		return 0
	}
	q, err := resource.ParseQuantity(value)
	if err != nil {
		return 0
	}
	return q.MilliValue()
}

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
		// Same first step as resume: classify every node and device against
		// the app's requirement. Where resume hands that view to the user,
		// install picks from it here and then runs the pick through resume's
		// own validation so an automatic placement can't bypass a rule a
		// manual one has to satisfy.
		availability := listAvailableForLaunch(req, nodes, pressure)
		selections, ok := pickLaunchSelection(req, availability, pressure, allocationOptions{checkPressure: true})
		if !ok {
			return nil, nil, fmt.Errorf("no available compute resource for type %s", req.Mode)
		}
		picked, validation := bindAllocations(appConfig, req, selections, nodes, pressure)
		if !validation.OK {
			return nil, nil, fmt.Errorf("no available compute resource for type %s: %s", req.Mode, validation.Code)
		}
		pickedAllocations = picked
		next := replaceAppAllocations(allocations, picked)
		return next, &picked[0], nil
	})
	if err != nil {
		return nil, err
	}
	if err := syncHAMIBindings(ctx, c, appConfig.AppName, appConfig.OwnerName, pickedAllocations); err != nil {
		return nil, err
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

// matchingNodes returns the nodes that support `mode`, each projected onto that
// single mode (Devices filtered to the mode) so the device-centric helpers only
// ever see the relevant devices of a multi-mode node.
func matchingNodes(mode string, nodes []Node) []Node {
	out := make([]Node, 0)
	for _, node := range nodes {
		if node.SupportsMode(mode) {
			out = append(out, node.viewForMode(mode))
		}
	}
	return out
}

func PickAllocations(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot) ([]Allocation, bool) {
	return pickAllocations(appConfig, req, nodes, pressure, allocationOptions{checkPressure: true})
}

type allocationOptions struct {
	checkPressure bool
	deterministic bool
	pressureAdded *AddedResources
}

// pickAllocations places an app automatically: it builds the availability view
// resume would show the user, picks a binding out of it, and turns that binding
// into allocation rows through the same resolve/build path resume uses for a
// manually submitted one.
func pickAllocations(appConfig *appcfg.ApplicationConfig, req Requirement, nodes []Node, pressure PressureSnapshot, pressureOptions allocationOptions) ([]Allocation, bool) {
	availability := listAvailableForLaunchWithOptions(req, nodes, pressure, pressureOptions)
	selections, ok := pickLaunchSelection(req, availability, pressure, pressureOptions)
	if !ok {
		return nil, false
	}
	resolved, err := resolveSelection(selections, nodes)
	if err != nil {
		return nil, false
	}
	allocations := allocationsFromResolvedSelection(appConfig, req, resolved)
	if len(allocations) == 0 {
		return nil, false
	}
	return allocations, true
}

// pickLaunchSelection auto-picks a compute binding out of the availability view
// that resume hands to the user for a manual pick. Install runs this instead of
// prompting, so both flows choose among exactly the same devices, ranked the
// same way: the best fit level first (an app that fits its limit is placed
// before one that only fits its request), then whole cards before shared ones,
// then at random among the remaining ties to spread load across the cluster.
func pickLaunchSelection(req Requirement, availability *AvailabilityResult, pressure PressureSnapshot, opts allocationOptions) ([]BindingSelection, bool) {
	if availability == nil || req.Mode == utils.CPUType {
		return nil, false
	}
	switch availability.Scope {
	case AvailabilityScopeCrossNode:
		return pickAggregateSelection(req, availability.Nodes, pressure, opts, true)
	case AvailabilityScopeSingleNode:
		return pickAggregateSelection(req, availability.Nodes, pressure, opts, false)
	default:
		return pickSingleSelection(req, availability.Nodes, pressure, opts)
	}
}

func evaluateCPUInstallMode(req Requirement, nodes []Node) ModePlanResult {
	for _, node := range nodes {
		if usableCPUNodeMemory(node.memoryCapacity) >= req.RequiredMemory {
			return ModePlanResult{ComputeType: utils.CPUType, Status: StatusInstallable}
		}
	}
	return ModePlanResult{ComputeType: utils.CPUType, Status: StatusInsufficientResources, Reason: "insufficient_resources"}
}

func usableCPUNodeMemory(capacity int64) int64 {
	if capacity < 0 {
		return -1
	}
	return capacity/4*3 + (capacity%4)*3/4
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
	// Non-nvidia modes are scheduled one device at a time, but a node can still
	// expose several of them (nvidia-gb10 cards, discrete Intel GPUs), so every
	// device gets a look rather than just the first.
	for _, node := range nodes {
		for _, device := range node.Devices {
			if device.Memory >= req.RequiredMemory {
				return true
			}
		}
	}
	return false
}

// pickSingleSelection places the app on one device, which is the unit of
// scheduling for every mode except multi-card nvidia.
func pickSingleSelection(req Requirement, nodes []NodeOption, pressure PressureSnapshot, opts allocationOptions) ([]BindingSelection, bool) {
	for _, level := range []string{FitLevelLimit, FitLevelRequired} {
		for _, supportType := range supportTypeOrder(req.Mode) {
			candidates := make([]BindingSelection, 0)
			for _, node := range nodes {
				for _, option := range node.Devices {
					if option.SupportType != supportType || !option.Operable {
						continue
					}
					if fits, _ := deviceOptionFits(req, option, pressure, level, false, 0, opts); !fits {
						continue
					}
					candidates = append(candidates, selectDevice(option, requiredTargetForMode(req)))
				}
			}
			if len(candidates) > 0 {
				if opts.deterministic {
					return []BindingSelection{candidates[0]}, true
				}
				return []BindingSelection{candidates[rand.Intn(len(candidates))]}, true
			}
		}
	}
	return nil, false
}

// pickAggregateSelection places a multi-card nvidia app across several cards,
// either within a single node or — when the app declares multi-node support —
// across the whole cluster.
func pickAggregateSelection(req Requirement, nodes []NodeOption, pressure PressureSnapshot, opts allocationOptions, crossNode bool) ([]BindingSelection, bool) {
	for _, level := range []string{FitLevelLimit, FitLevelRequired} {
		target := targetGPU(req, level)
		if target <= 0 {
			continue
		}
		if crossNode {
			if picked, ok := collectDevicesForTarget(req, nodes, pressure, opts, level, target, req.RequiredGPU); ok {
				return picked, true
			}
			continue
		}
		for _, node := range nodes {
			if picked, ok := collectDevicesForTarget(req, []NodeOption{node}, pressure, opts, level, target, req.RequiredGPU); ok {
				return picked, true
			}
		}
	}
	return nil, false
}

func collectDevicesForTarget(req Requirement, nodes []NodeOption, pressure PressureSnapshot, opts allocationOptions, level string, target, allocationTarget int64) ([]BindingSelection, bool) {
	fitRemaining := target
	allocationRemaining := allocationTarget
	var out []BindingSelection
	timeSliceMemoryByNode := make(map[string]int64)
	for _, supportType := range supportTypeOrder(req.Mode) {
		for _, node := range nodes {
			for _, option := range node.Devices {
				if option.SupportType != supportType || !option.Operable {
					continue
				}
				fits, amount := deviceOptionFits(req, option, pressure, level, true, timeSliceMemoryByNode[node.NodeName], opts)
				if !fits || amount <= 0 {
					continue
				}
				_, device := option.asNodeDevice(req.Mode)
				nextMemory, ok := checkedAdd(timeSliceMemoryByNode[node.NodeName], timeSliceAddedMemory(device))
				if !ok {
					continue
				}
				timeSliceMemoryByNode[node.NodeName] = nextMemory
				if allocationRemaining > 0 {
					assigned := minInt64(amount, allocationRemaining)
					out = append(out, selectDevice(option, assigned))
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

// selectDevice records a picked device as the same BindingSelection the resume
// frontend submits. `assigned` only reaches the allocation for memory-slice
// cards, where the binding carves out an explicit slice; whole-card placements
// ignore it. A non-positive amount means the app declared no concrete demand
// for this mode, so it takes whatever the card has free.
func selectDevice(option DeviceOption, assigned int64) BindingSelection {
	if assigned <= 0 {
		assigned = option.Available
	}
	return BindingSelection{NodeName: option.NodeName, DeviceID: option.DeviceID, Memory: assigned}
}

func deviceOptionFits(req Requirement, option DeviceOption, pressure PressureSnapshot, level string, allowPartial bool, priorTimeSliceMemory int64, opts allocationOptions) (bool, int64) {
	node, device := option.asNodeDevice(req.Mode)
	return deviceFitsLevelWithPressure(req, node, device, pressure, level, allowPartial, priorTimeSliceMemory, opts)
}

func deviceFitsLevelWithPressure(req Requirement, node Node, device Device, pressure PressureSnapshot, level string, allowPartial bool, priorTimeSliceMemory int64, pressureOptions allocationOptions) (bool, int64) {
	if device.Health != "" && device.Health != deviceHealthYes {
		return false, 0
	}
	required := targetForMode(req, level)
	if required <= 0 {
		if !pressureOptions.checkPressure {
			return true, 0
		}
		return !pressure.WouldPressure(node, levelAddedResources(req, level, pressureOptions)), 0
	}
	available := deviceAvailableMemory(device)
	if available <= 0 {
		return false, 0
	}
	if available < required && !(allowPartial && req.Mode == utils.NvidiaCardType) {
		return false, available
	}
	if allocationWouldPressure(req, node, device, pressure, level, priorTimeSliceMemory, pressureOptions) {
		return false, available
	}
	return true, available
}

func allocationWouldPressure(req Requirement, node Node, device Device, pressure PressureSnapshot, level string, priorTimeSliceMemory int64, options allocationOptions) bool {
	if !options.checkPressure {
		return false
	}
	added := levelAddedResources(req, level, options)
	timeSliceMemory, ok := checkedAdd(priorTimeSliceMemory, timeSliceAddedMemory(device))
	if !ok {
		return true
	}
	added.Memory, ok = checkedAdd(added.Memory, timeSliceMemory)
	if !ok {
		return true
	}
	return pressure.WouldPressure(node, added)
}

func levelAddedResources(req Requirement, level string, options allocationOptions) AddedResources {
	if options.pressureAdded != nil {
		return *options.pressureAdded
	}
	return AddedResources{
		CPU:    req.RequiredCPU,
		Memory: levelMemory(req, level),
		Disk:   req.RequiredDisk,
	}
}

// A time-slice GPU with a positive target needs host-memory headroom equal to
// the card's physical memory. Zero-target fit checks preserve legacy behavior.
func timeSliceAddedMemory(device Device) int64 {
	if device.SupportType != SupportTypeTimeSlice {
		return 0
	}
	return device.Memory
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
	if mode == utils.NvidiaCardType {
		return []string{SupportTypeExclusive, SupportTypeMemorySlice, SupportTypeTimeSlice}
	}
	// Every non-nvidia mode (nvidia-gb10 plus the unified-memory accelerators:
	// cpu / apple-m / intel / amd / moore-soc / discrete GPUs) only ever
	// carries Exclusive or MemorySlice devices — TimeSlice is nvidia-only.
	// Listing Exclusive first keeps the scheduler's preference for whole-device
	// placement before it falls back to a memory slice, and including
	// MemorySlice is what lets GB10 (decoded as MemorySlice by default) and the
	// memory-slicing accelerators be scheduled at all.
	return []string{SupportTypeExclusive, SupportTypeMemorySlice}
}

func deviceAvailableMemory(device Device) int64 {
	switch device.SupportType {
	case SupportTypeExclusive:
		if len(device.Bindings) > 0 {
			return 0
		}
		return device.Memory
	case SupportTypeMemorySlice:
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

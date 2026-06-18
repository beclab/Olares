package compute

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errBindingUnavailable = errors.New("compute binding unavailable")

func listAvailableForLaunch(req Requirement, nodes []Node, pressure PressureSnapshot) *AvailabilityResult {
	result := &AvailabilityResult{
		Requirement: req,
		Scope:       availabilityScope(req),
		Nodes:       make([]NodeOption, 0, len(nodes)),
	}
	classified := classifyLaunchNodes(req, nodes, pressure)
	for _, node := range classified {
		if node.Status == NodeStatusNotMatch {
			continue
		}
		result.Nodes = append(result.Nodes, node)
	}
	if len(result.Nodes) == 0 {
		result.Schedulable = false
		result.Reason = "no-matching-node"
		return result
	}
	markOperable(result)
	result.Schedulable, result.Reason = availabilitySummary(req, result.Nodes)
	return result
}

func classifyLaunchNodes(req Requirement, nodes []Node, pressure PressureSnapshot) []NodeOption {
	out := make([]NodeOption, 0, len(nodes))
	for _, node := range nodes {
		if !node.SupportsMode(req.Mode) {
			out = append(out, NodeOption{
				NodeName: node.NodeName,
				GPUType:  node.primaryGPUType(),
				Status:   NodeStatusNotMatch,
			})
			continue
		}
		view := node.viewForMode(req.Mode)
		var option NodeOption
		if req.Mode == utils.NvidiaCardType {
			option = classifyNvidiaNode(req, view, pressure)
		} else {
			option = classifyNonNvidiaNode(req, view, pressure)
		}
		out = append(out, option)
	}
	return out
}

func classifyNvidiaNode(req Requirement, node Node, pressure PressureSnapshot) NodeOption {
	summary := summarizeNvidiaNode(req, node, pressure)
	option := NodeOption{
		NodeName: node.NodeName,
		GPUType:  req.Mode,
		Devices:  summary.devices,
	}
	if req.SupportMultiCards || req.SupportMultiNodes {
		option.Status = nodeStatusFromCapacity(req.RequiredGPU, summary.totalCapacity, summary.totalAvailable)
		return option
	}
	option.Status = nodeStatusFromCapacity(req.RequiredGPU, summary.maxCapacity, summary.maxAvailable)
	return option
}

func classifyNonNvidiaNode(req Requirement, node Node, pressure PressureSnapshot) NodeOption {
	option := NodeOption{NodeName: node.NodeName, GPUType: req.Mode}
	if len(node.Devices) == 0 {
		option.Status = NodeStatusNotAvailable
		return option
	}
	devOpt := makeDeviceOption(req, node, node.Devices[0], pressure)
	option.Devices = []DeviceOption{devOpt}
	switch {
	case devOpt.Health != deviceHealthYes:
		option.Status = NodeStatusNotAvailable
	case devOpt.Capacity < req.RequiredMemory:
		option.Status = NodeStatusNotEnough
	case devOpt.Available >= req.RequiredMemory && !pressure.WouldPressure(node, AddedResources{
		CPU:    req.RequiredCPU,
		Memory: req.RequiredMemory,
	}):
		option.Status = NodeStatusAvailable
	default:
		option.Status = NodeStatusNotAvailable
	}
	return option
}

type nvidiaNodeSummary struct {
	devices        []DeviceOption
	totalCapacity  int64
	totalAvailable int64
	maxCapacity    int64
	maxAvailable   int64
}

func summarizeNvidiaNode(req Requirement, node Node, pressure PressureSnapshot) nvidiaNodeSummary {
	summary := nvidiaNodeSummary{devices: make([]DeviceOption, 0, len(node.Devices))}
	for _, device := range node.Devices {
		devOpt := makeDeviceOption(req, node, device, pressure)
		summary.devices = append(summary.devices, devOpt)
		if devOpt.Health != deviceHealthYes {
			continue
		}
		summary.totalCapacity += device.Memory
		summary.totalAvailable += devOpt.Available
		if device.Memory > summary.maxCapacity {
			summary.maxCapacity = device.Memory
		}
		if devOpt.Available > summary.maxAvailable {
			summary.maxAvailable = devOpt.Available
		}
	}
	return summary
}

func nodeStatusFromCapacity(required, capacity, available int64) string {
	if capacity < required {
		return NodeStatusNotEnough
	}
	if available >= required {
		return NodeStatusAvailable
	}
	return NodeStatusNotAvailable
}

func makeDeviceOption(req Requirement, node Node, device Device, pressure PressureSnapshot) DeviceOption {
	req.RequiredDisk = 0
	available := deviceAvailableMemory(device)
	option := DeviceOption{
		NodeName:    node.NodeName,
		DeviceID:    device.ID,
		CardModel:   device.CardModel,
		SupportType: device.SupportType,
		Capacity:    device.Memory,
		Available:   available,
		Health:      device.Health,
		Bindings:    device.Bindings,
	}
	if option.Health == "" {
		option.Health = deviceHealthYes
	}
	for _, level := range []string{FitLevelLimit, FitLevelRequired} {
		fits, _ := deviceFitsLevel(req, node, device, pressure, level, req.SupportMultiCards || req.SupportMultiNodes)
		if fits {
			option.FitLevel = level
			break
		}
	}
	return option
}

func availabilityScope(req Requirement) string {
	if req.Mode == utils.NvidiaCardType && req.SupportMultiNodes {
		return AvailabilityScopeCrossNode
	}
	if req.Mode == utils.NvidiaCardType && req.SupportMultiCards {
		return AvailabilityScopeSingleNode
	}
	// Single-nvidia-card and every non-nvidia mode (cpu / amd / intel /
	// apple-m / moore-soc — each modeled as one node-level device) share the
	// per-card scope: the unit of scheduling is a single device.
	return AvailabilityScopeCard
}

func markOperable(result *AvailabilityResult) {
	for ni := range result.Nodes {
		node := &result.Nodes[ni]
		switch result.Scope {
		case AvailabilityScopeCrossNode:
			if node.Status != NodeStatusNotEnough && node.Status != NodeStatusNotAvailable && node.Status != NodeStatusAvailable {
				continue
			}
		default:
			if node.Status != NodeStatusAvailable {
				continue
			}
		}
		for di := range node.Devices {
			node.Devices[di].Operable = node.Devices[di].Health == deviceHealthYes && node.Devices[di].Available > 0
		}
	}
}

// availabilitySummary decides whether the request is schedulable across the
// already-classified nodes. By the time we get here, NotMatch nodes have been
// filtered out by listAvailableForLaunch, and classifyNvidiaNode /
// classifyNonNvidiaNode have already folded pressure + capacity into
// node.Status. So:
//   - CrossNode: sum raw `device.Available` across healthy devices, per PDF.
//   - Everything else: any node already marked NodeStatusAvailable is enough,
//     because the classifier already established it can host the request.
func availabilitySummary(req Requirement, nodes []NodeOption) (bool, string) {
	scope := availabilityScope(req)
	if scope == AvailabilityScopeCrossNode {
		var total int64
		for _, node := range nodes {
			for _, device := range node.Devices {
				if device.Health != deviceHealthYes {
					continue
				}
				total += device.Available
			}
		}
		if total >= req.RequiredGPU {
			return true, ""
		}
		return false, "insufficient-cluster-vram"
	}
	for _, node := range nodes {
		if node.Status == NodeStatusAvailable {
			return true, ""
		}
	}
	return false, scopeNoAvailableReason(scope)
}

func scopeNoAvailableReason(scope string) string {
	switch scope {
	case AvailabilityScopeSingleNode:
		return "no-node-with-enough-cards"
	case AvailabilityScopeCard:
		return "no-card-with-enough-vram"
	default:
		return "no-available-node"
	}
}

func ApplyBindingSelection(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, selections []BindingSelection, includeSharedServer bool) (*BindingApplyResult, error) {
	if appConfig == nil {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	// For v2 cluster-shared apps the only allocation row and the only
	// HAMI bindings live at (appName, sharedServerOwner). When the
	// caller does not intend to touch the shared server (resume the
	// client only — the server is already running with its existing
	// allocation), skip compute binding entirely.
	if appConfig.IsV2() && appConfig.HasClusterSharedCharts() && !includeSharedServer {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	// resolveComputeTarget redirects to the actual server owner's config
	// when resume-all is triggered by someone who is not the original
	// installer of the shared server; in every other reachable case it
	// returns appConfig unchanged.
	targetConfig, _, err := resolveComputeTarget(ctx, c, appConfig, includeSharedServer)
	if err != nil {
		return nil, err
	}
	appConfig = targetConfig
	req, ok := SelectedRequirement(appConfig)
	if !ok || req.Mode == utils.CPUType {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	if len(selections) == 0 {
		nodes, err := FetchNodeComputeAllocationsExcludingApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
		if err != nil {
			return nil, err
		}
		return &BindingApplyResult{
			Status:       BindingApplyStatusRequired,
			Availability: listAvailableForLaunch(req, nodes, pressure),
			TargetApp:    appConfig.AppName,
			TargetOwner:  appConfig.OwnerName,
		}, nil
	}

	var allocations []Allocation
	var unavailable *BindingApplyResult
	if _, err := mutateAllocations(ctx, c, func(nodes []Node, existing []Allocation) ([]Allocation, *Allocation, error) {
		attachBindings(nodes, withoutAppAllocations(existing, appConfig.AppName, appConfig.OwnerName))
		resolved, resolveErr := resolveSelection(selections, nodes)
		if resolveErr != nil {
			unavailable = unavailableBindingApplyResult(req, nodes, pressure, invalidBinding(resolveErr.Error()))
			return nil, nil, errBindingUnavailable
		}
		validation := validateResolvedBindingSelection(req, resolved, pressure)
		if !validation.OK {
			unavailable = unavailableBindingApplyResult(req, nodes, pressure, validation)
			return nil, nil, errBindingUnavailable
		}
		allocations = allocationsFromResolvedSelection(appConfig, req, resolved)
		if len(allocations) == 0 {
			unavailable = unavailableBindingApplyResult(req, nodes, pressure, invalidBinding("empty-compute-binding"))
			return nil, nil, errBindingUnavailable
		}
		next := replaceAppAllocations(existing, allocations)
		return next, &allocations[0], nil
	}); err != nil {
		if errors.Is(err, errBindingUnavailable) {
			return unavailable, nil
		}
		return nil, err
	}
	if err := deleteHAMIBindingsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName); err != nil {
		_ = DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
		return nil, err
	}
	for _, allocation := range allocations {
		if err := createHAMIBinding(ctx, c, allocation); err != nil {
			_ = DeleteAllocationsForApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
			return nil, err
		}
	}
	return &BindingApplyResult{
		Status:      BindingApplyStatusApplied,
		Allocations: allocations,
		TargetApp:   appConfig.AppName,
		TargetOwner: appConfig.OwnerName,
	}, nil
}

// ValidateBindingForResume mirrors ApplyBindingSelection's feasibility
// checks but performs NO writes: it never mutates the allocation config
// map and never creates HAMI bindings. It is the read-only counterpart
// intended for a pre-flight "would this resume succeed?" call the
// frontend can make before issuing the real resume.
//
// The returned BindingApplyResult uses the same status/availability/
// validation shapes as ApplyBindingSelection so the two endpoints stay
// format-compatible, with one difference: where ApplyBindingSelection
// returns BindingApplyStatusApplied after writing, this returns
// BindingApplyStatusValid (the selection is valid but nothing was
// applied). The accompanying Allocations describe what WOULD be written.
func ValidateBindingForResume(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, selections []BindingSelection, includeSharedServer bool) (*BindingApplyResult, error) {
	if appConfig == nil {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	if appConfig.IsV2() && appConfig.HasClusterSharedCharts() && !includeSharedServer {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	targetConfig, _, err := resolveComputeTarget(ctx, c, appConfig, includeSharedServer)
	if err != nil {
		return nil, err
	}
	appConfig = targetConfig
	req, ok := SelectedRequirement(appConfig)
	if !ok || req.Mode == utils.CPUType {
		return &BindingApplyResult{Status: BindingApplyStatusNotRequired}, nil
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	// Exclude the app's own existing allocation rows so a re-binding of the
	// same app does not count its current claim against availability, matching
	// the node view ApplyBindingSelection validates against.
	nodes, err := FetchNodeComputeAllocationsExcludingApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
	if err != nil {
		return nil, err
	}
	if len(selections) == 0 {
		return &BindingApplyResult{
			Status:       BindingApplyStatusRequired,
			Availability: listAvailableForLaunch(req, nodes, pressure),
			TargetApp:    appConfig.AppName,
			TargetOwner:  appConfig.OwnerName,
		}, nil
	}
	resolved, resolveErr := resolveSelection(selections, nodes)
	if resolveErr != nil {
		return unavailableBindingApplyResult(req, nodes, pressure, invalidBinding(resolveErr.Error())), nil
	}
	validation := validateResolvedBindingSelection(req, resolved, pressure)
	if !validation.OK {
		return unavailableBindingApplyResult(req, nodes, pressure, validation), nil
	}
	allocations := allocationsFromResolvedSelection(appConfig, req, resolved)
	if len(allocations) == 0 {
		return unavailableBindingApplyResult(req, nodes, pressure, invalidBinding("empty-compute-binding")), nil
	}
	// Even when the selection is valid we still hand back the full list of
	// available options so the frontend can render the current selection in
	// context and offer alternatives, mirroring the Required / Unavailable
	// payloads exactly (Availability is always populated).
	return &BindingApplyResult{
		Status:       BindingApplyStatusValid,
		Allocations:  allocations,
		Availability: listAvailableForLaunch(req, nodes, pressure),
		Validation:   validation,
		TargetApp:    appConfig.AppName,
		TargetOwner:  appConfig.OwnerName,
	}, nil
}

func unavailableBindingApplyResult(req Requirement, nodes []Node, pressure PressureSnapshot, validation *BindingValidationResult) *BindingApplyResult {
	return &BindingApplyResult{
		Status:       BindingApplyStatusUnavailable,
		Availability: listAvailableForLaunch(req, nodes, pressure),
		Validation:   validation,
	}
}

func ValidateBindingSelection(req Requirement, selections []BindingSelection, nodes []Node, pressure PressureSnapshot) *BindingValidationResult {
	if len(selections) == 0 {
		return invalidBinding("empty-selection")
	}
	resolved, err := resolveSelection(selections, nodes)
	if err != nil {
		return invalidBinding(err.Error())
	}
	return validateResolvedBindingSelection(req, resolved, pressure)
}

func validateResolvedBindingSelection(req Requirement, resolved []resolvedSelection, pressure PressureSnapshot) *BindingValidationResult {
	selectedNodes := map[string]struct{}{}
	for _, item := range resolved {
		selectedNodes[item.node.NodeName] = struct{}{}
	}
	if req.Mode == utils.NvidiaCardType {
		if !req.SupportMultiCards && len(resolved) != 1 {
			return invalidBinding("multi-card-not-supported")
		}
		if !req.SupportMultiNodes && len(selectedNodes) != 1 {
			return invalidBinding("multi-node-not-supported")
		}
	} else if len(resolved) != 1 {
		return invalidBinding("non-nvidia-must-single-selection")
	}
	var totalAssignable int64
	for _, item := range resolved {
		if item.device.Health != "" && item.device.Health != deviceHealthYes {
			return invalidBinding("device-unhealthy:" + item.device.ID)
		}
		if item.device.Mode != req.Mode {
			return invalidBinding("gpu-type-mismatch")
		}
		available := deviceAvailableMemory(item.device)
		if req.Mode == utils.NvidiaCardType {
			switch item.device.SupportType {
			case SupportTypeExclusive:
				if len(item.device.Bindings) > 0 {
					return invalidBinding("exclusive-already-bound:" + item.device.ID)
				}
			case SupportTypeMemorySlice:
				if item.memory <= 0 {
					return invalidBinding("memory-required:" + item.device.ID)
				}
				if item.memory > available {
					return invalidBinding("device-vram-insufficient:" + item.device.ID)
				}
				totalAssignable += item.memory
				continue
			}
		}
		totalAssignable += available
	}
	if req.Mode == utils.NvidiaCardType && req.SupportMultiCards {
		if totalAssignable < req.RequiredGPU {
			return invalidBinding("aggregate-vram-insufficient")
		}
	} else if req.Mode == utils.NvidiaCardType {
		if totalAssignable < req.RequiredGPU {
			return invalidBinding("device-vram-insufficient")
		}
	} else if deviceAvailableMemory(resolved[0].device) < req.RequiredMemory {
		return invalidBinding("device-memory-insufficient")
	}
	for nodeName := range selectedNodes {
		node := findResolvedNode(nodeName, resolved)
		hasTimeSlice := false
		for _, item := range resolved {
			if item.node.NodeName == nodeName && item.device.SupportType == SupportTypeTimeSlice {
				hasTimeSlice = true
				break
			}
		}
		addedGPU := int64(0)
		if req.Mode == utils.NvidiaCardType && hasTimeSlice {
			addedGPU = req.LimitedGPU
		}
		if pressure.WouldPressure(node, AddedResources{
			CPU:    req.RequiredCPU,
			Memory: req.RequiredMemory + addedGPU,
		}) {
			return invalidBinding("node-pressure:" + nodeName)
		}
	}
	return &BindingValidationResult{OK: true, Code: BindingValidationReasonValid}
}

type resolvedSelection struct {
	node   Node
	device Device
	memory int64
}

func resolveSelection(selections []BindingSelection, nodes []Node) ([]resolvedSelection, error) {
	out := make([]resolvedSelection, 0, len(selections))
	seen := make(map[string]struct{}, len(selections))
	for _, item := range selections {
		node, ok := findNode(nodes, item.NodeName)
		if !ok {
			return nil, fmt.Errorf("node-not-found:%s", item.NodeName)
		}
		device, ok := findDevice(node, item.DeviceID)
		if !ok {
			return nil, fmt.Errorf("device-not-found:%s", item.DeviceID)
		}
		key := item.NodeName + "/" + item.DeviceID
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate-selection:%s", key)
		}
		seen[key] = struct{}{}
		out = append(out, resolvedSelection{node: node, device: device, memory: item.Memory})
	}
	return out, nil
}

func findNode(nodes []Node, name string) (Node, bool) {
	for _, node := range nodes {
		if node.NodeName == name {
			return node, true
		}
	}
	return Node{}, false
}

func findDevice(node Node, id string) (Device, bool) {
	for _, device := range node.Devices {
		if device.ID == id {
			return device, true
		}
	}
	return Device{}, false
}

func findResolvedNode(name string, resolved []resolvedSelection) Node {
	for _, item := range resolved {
		if item.node.NodeName == name {
			return item.node
		}
	}
	return Node{}
}

func allocationsFromResolvedSelection(appConfig *appcfg.ApplicationConfig, req Requirement, resolved []resolvedSelection) []Allocation {
	sort.SliceStable(resolved, func(i, j int) bool {
		if resolved[i].node.NodeName == resolved[j].node.NodeName {
			return resolved[i].device.ID < resolved[j].device.ID
		}
		return resolved[i].node.NodeName < resolved[j].node.NodeName
	})
	target := req.RequiredMemory
	if req.Mode == utils.NvidiaCardType {
		target = req.RequiredGPU
	}
	out := make([]Allocation, 0, len(resolved))
	remaining := target
	for _, item := range resolved {
		amount := target
		switch {
		case item.device.SupportType == SupportTypeMemorySlice && item.memory > 0:
			// Memory-slice cards carve out an explicit per-card slice; the
			// frontend always sends a positive Memory for them (enforced by
			// validateResolvedBindingSelection).
			amount = item.memory
		case isWholeCardMode(req.Mode, item.device.SupportType):
			// Exclusive / TimeSlice hand the pod the whole card and
			// buildAllocation records Memory=0, so every selected card must
			// produce its own binding. These must never be gated on the
			// shared `remaining` VRAM budget: once an earlier card covered
			// the RequiredGPU target the budget reaches zero and the rest of
			// a multi-card selection would be silently dropped, leaving only
			// a single HAMI binding for a two-card request.
			amount = deviceAvailableMemory(item.device)
		case len(resolved) > 1:
			amount = minInt64(deviceAvailableMemory(item.device), remaining)
		}
		if amount <= 0 {
			continue
		}
		out = append(out, buildAllocation(appConfig, req, item.node, item.device, amount))
		remaining -= amount
	}
	return out
}

func invalidBinding(code string) *BindingValidationResult {
	return &BindingValidationResult{OK: false, Code: code, Reason: BindingValidationReasonInvalid}
}

package compute

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errBindingUnavailable = errors.New("compute binding unavailable")

// listAvailableForLaunch classifies every node and device in the cluster
// against the app's selected requirement. It is the shared first step of both
// placement flows: resume hands the result to the user and waits for a manual
// pick, while install feeds the very same result to pickLaunchSelection and
// picks automatically. Neither flow can therefore place an app on a device the
// other wouldn't offer.
func listAvailableForLaunch(req Requirement, nodes []Node, pressure PressureSnapshot) *AvailabilityResult {
	return listAvailableForLaunchWithOptions(req, nodes, pressure, allocationOptions{checkPressure: true})
}

// listAvailableForLaunchWithOptions is listAvailableForLaunch with the fit
// checks parameterized, so preflight can build the same view against a
// simulated cluster (its own added-resources budget, and a first pass that
// ignores node pressure entirely).
func listAvailableForLaunchWithOptions(req Requirement, nodes []Node, pressure PressureSnapshot, opts allocationOptions) *AvailabilityResult {
	result := &AvailabilityResult{
		Requirement: req,
		Scope:       availabilityScope(req),
		Nodes:       make([]NodeOption, 0, len(nodes)),
	}
	classified := classifyLaunchNodes(req, nodes, pressure, opts)
	for _, node := range classified {
		if node.Status == NodeStatusNotMatch {
			continue
		}
		result.Nodes = append(result.Nodes, node)
	}
	if len(result.Nodes) == 0 {
		result.Schedulable = false
		result.Reason = "no-matching-node"
	} else {
		markOperable(result)
		result.Schedulable, result.Reason = availabilitySummary(req, result.Nodes)
	}
	// Both flows are built on this view — install picks from it, resume renders
	// it — so one line here explains either one after the fact.
	klog.V(2).Infof("compute: placement options for %s: %s",
		describeRequirement(req), explainPlacement(req, result, pressure))
	return result
}

// classifyLaunchNodes classifies the nodes that can still be placed on. A node
// whose kubelet has stopped heartbeating is dropped outright and is absent from
// the returned slice; every other node gets an entry, including the ones that
// end up NodeStatusNotMatch.
func classifyLaunchNodes(req Requirement, nodes []Node, pressure PressureSnapshot, opts allocationOptions) []NodeOption {
	out := make([]NodeOption, 0, len(nodes))
	for _, node := range nodes {
		// Readiness is a node-level fact and it decides the node on its own:
		// the devices are not consulted at all. They cannot answer the
		// question anyway — the register annotation listing them is written
		// while the node is up and never cleared, so a powered-off node keeps
		// advertising its last known cards indefinitely. Dropping the node
		// here covers both flows at once, since install picks out of this list
		// and resume renders it.
		if !node.isReady() {
			klog.Infof("compute: skipping node %s for %s placement: node is not ready", node.NodeName, req.Mode)
			continue
		}
		if !node.SupportsMode(req.Mode) {
			// NotMatch is filtered out downstream, so this is the only trace of
			// a node that vanished from the picker for advertising the wrong
			// accelerator — the "my GPU node isn't even listed" case.
			klog.V(2).Infof("compute: skipping node %s for %s placement: it advertises %v",
				node.NodeName, req.Mode, node.GPUTypes)
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
			option = classifyNvidiaNode(req, view, pressure, opts)
		} else {
			option = classifyNonNvidiaNode(req, view, pressure, opts)
		}
		out = append(out, option)
	}
	return out
}

func classifyNvidiaNode(req Requirement, node Node, pressure PressureSnapshot, opts allocationOptions) NodeOption {
	summary := summarizeNvidiaNode(req, node, pressure, opts)
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

// classifyNonNvidiaNode classifies every device a node exposes for a non-nvidia
// mode. Most such modes are unified memory and model the whole node as a single
// device, but nvidia-gb10 and discrete Intel GPUs can expose several cards on
// one node, so the node's status is taken from its best card the way
// classifyNvidiaNode does — listing only the first one would hide the rest from
// both the resume picker and install's auto-pick.
func classifyNonNvidiaNode(req Requirement, node Node, pressure PressureSnapshot, opts allocationOptions) NodeOption {
	option := NodeOption{NodeName: node.NodeName, GPUType: req.Mode}
	if len(node.Devices) == 0 {
		option.Status = NodeStatusNotAvailable
		return option
	}
	option.Devices = make([]DeviceOption, 0, len(node.Devices))
	var healthy bool
	var maxCapacity, maxAvailable int64
	for _, device := range node.Devices {
		devOpt := makeDeviceOption(req, node, device, pressure, opts)
		option.Devices = append(option.Devices, devOpt)
		if devOpt.Health != deviceHealthYes {
			continue
		}
		healthy = true
		if devOpt.Capacity > maxCapacity {
			maxCapacity = devOpt.Capacity
		}
		if devOpt.Available > maxAvailable {
			maxAvailable = devOpt.Available
		}
	}
	if !healthy {
		option.Status = NodeStatusNotAvailable
		return option
	}
	option.Status = nodeStatusFromCapacity(req.RequiredMemory, maxCapacity, maxAvailable)
	if option.Status == NodeStatusAvailable && nodeWouldPressure(req, node, pressure, opts) {
		option.Status = NodeStatusNotAvailable
	}
	return option
}

// nodeWouldPressure reports whether hosting the app would push the node past
// its pressure threshold. Unlike the per-device fit checks it deliberately
// leaves disk out: node status has always been a cpu/memory judgement.
func nodeWouldPressure(req Requirement, node Node, pressure PressureSnapshot, opts allocationOptions) bool {
	if !opts.checkPressure {
		return false
	}
	added := AddedResources{CPU: req.RequiredCPU, Memory: req.RequiredMemory}
	if opts.pressureAdded != nil {
		added = *opts.pressureAdded
	}
	return pressure.WouldPressure(node, added)
}

type nvidiaNodeSummary struct {
	devices        []DeviceOption
	totalCapacity  int64
	totalAvailable int64
	maxCapacity    int64
	maxAvailable   int64
}

func summarizeNvidiaNode(req Requirement, node Node, pressure PressureSnapshot, opts allocationOptions) nvidiaNodeSummary {
	summary := nvidiaNodeSummary{devices: make([]DeviceOption, 0, len(node.Devices))}
	for _, device := range node.Devices {
		devOpt := makeDeviceOption(req, node, device, pressure, opts)
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

func makeDeviceOption(req Requirement, node Node, device Device, pressure PressureSnapshot, opts allocationOptions) DeviceOption {
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
		fits, _ := deviceFitsLevelWithPressure(req, node, device, pressure, level, req.SupportMultiCards || req.SupportMultiNodes, opts)
		if fits {
			option.FitLevel = level
			break
		}
	}
	return option
}

// asNodeDevice re-materializes the (node, device) pair a DeviceOption was
// derived from, so the auto-picker can run the fit checks straight against the
// availability view rather than against a second copy of the node snapshot.
// That is what keeps install's automatic pick and resume's manual pick anchored
// to the same candidate set: both are choosing among the very devices this view
// exposes, judged by the very fit level it reports.
//
// The conversion goes this direction, rather than deviceFitsLevelWithPressure
// simply taking a DeviceOption, because that function is also what builds the
// view: makeDeviceOption calls it to compute FitLevel, at which point only the
// raw Node and Device exist and the DeviceOption does not yet. The fit logic
// therefore has to stay Device-shaped, and the view side adapts.
//
// The round trip is exact for everything the fit checks read — Health, plus
// SupportType/Memory/Bindings, which are what deviceAvailableMemory consumes,
// plus the node name that pressure lookups are keyed by. Recomputing
// deviceAvailableMemory on the result therefore reproduces
// DeviceOption.Available. The identity fields (ID, Mode, CardModel) are along
// for the ride so the value is a well-formed Device; no fit check reads them.
func (o DeviceOption) asNodeDevice(mode string) (Node, Device) {
	return Node{NodeName: o.NodeName}, Device{
		ID:          o.DeviceID,
		NodeName:    o.NodeName,
		Mode:        mode,
		CardModel:   o.CardModel,
		Memory:      o.Capacity,
		Health:      o.Health,
		SupportType: o.SupportType,
		Bindings:    o.Bindings,
	}
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
		bound, validation := bindAllocations(appConfig, req, selections, nodes, pressure)
		if !validation.OK {
			// The user picked these cards by hand and got an error back; record
			// what they picked and which rule refused it, so a support report
			// does not depend on the user relaying the code.
			klog.Infof("compute: resume binding for app %s/%s refused (%s), submitted=%s, %s",
				appConfig.OwnerName, appConfig.AppName, validation.Code,
				describeSelections(selections), describeRequirement(req))
			unavailable = unavailableBindingApplyResult(req, nodes, pressure, validation)
			return nil, nil, errBindingUnavailable
		}
		allocations = bound
		next := replaceAppAllocations(existing, allocations)
		return next, &allocations[0], nil
	}); err != nil {
		if errors.Is(err, errBindingUnavailable) {
			return unavailable, nil
		}
		return nil, err
	}
	// Outside the mutation, which is replayed on a write conflict.
	klog.Infof("compute: app %s/%s bound to %s on resume, %s",
		appConfig.OwnerName, appConfig.AppName, describeAllocations(allocations), describeRequirement(req))
	if err := syncHAMIBindings(ctx, c, appConfig.AppName, appConfig.OwnerName, allocations); err != nil {
		return nil, err
	}
	return &BindingApplyResult{
		Status:      BindingApplyStatusApplied,
		Allocations: allocations,
		TargetApp:   appConfig.AppName,
		TargetOwner: appConfig.OwnerName,
	}, nil
}

// bindAllocations turns a compute binding selection — the user's manual pick on
// resume, or install's automatic pick out of the same availability view — into
// the app's allocation rows. Both flows converge here so a placement install
// makes on its own is held to exactly the same rules as one a user submits.
// The returned validation result is always non-nil; the allocations are only
// meaningful when it reports OK.
func bindAllocations(appConfig *appcfg.ApplicationConfig, req Requirement, selections []BindingSelection, nodes []Node, pressure PressureSnapshot) ([]Allocation, *BindingValidationResult) {
	resolved, err := resolveSelection(selections, nodes)
	if err != nil {
		return nil, invalidBinding(err.Error())
	}
	validation := validateResolvedBindingSelection(req, resolved, pressure)
	if !validation.OK {
		return nil, validation
	}
	allocations := allocationsFromResolvedSelection(appConfig, req, resolved)
	if len(allocations) == 0 {
		return nil, invalidBinding("empty-compute-binding")
	}
	return allocations, validation
}

// syncHAMIBindings replaces the app's HAMI GPUBindings with the ones its new
// allocations call for. A failure at this point leaves the allocation rows
// pointing at bindings that were never created, so it releases them.
func syncHAMIBindings(ctx context.Context, c client.Client, appName, owner string, allocations []Allocation) error {
	if err := deleteHAMIBindingsForApp(ctx, c, appName, owner); err != nil {
		_ = DeleteAllocationsForApp(ctx, c, appName, owner)
		return err
	}
	for _, allocation := range allocations {
		if err := createHAMIBinding(ctx, c, allocation); err != nil {
			_ = DeleteAllocationsForApp(ctx, c, appName, owner)
			return err
		}
	}
	return nil
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
	allocations, validation := bindAllocations(appConfig, req, selections, nodes, pressure)
	if !validation.OK {
		// This is the dry run the frontend makes before offering the resume
		// button, so it can be asked many times for one user action and stays
		// quiet by default. It is still the only record of why the button came
		// back disabled, which is a question the real-resume log cannot answer
		// because that resume never happened.
		klog.V(2).Infof("compute: resume pre-check for app %s/%s says the selection is unusable (%s), submitted=%s, %s",
			appConfig.OwnerName, appConfig.AppName, validation.Code,
			describeSelections(selections), describeRequirement(req))
		return unavailableBindingApplyResult(req, nodes, pressure, validation), nil
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
		if dims := pressure.PressuredDimensions(node, AddedResources{
			CPU:    req.RequiredCPU,
			Memory: req.RequiredMemory,
		}); len(dims) > 0 {
			result := invalidBinding("node-pressure:" + nodeName)
			result.NodePressure = &NodePressureDetail{NodeName: nodeName, Dimensions: dims}
			return result
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
			// The app declares no concrete demand for this mode, so the
			// selection's own amount is all we have to go on. Without this the
			// allocation would be dropped and the whole binding rejected as
			// empty.
			amount = item.memory
		}
		if amount <= 0 {
			// Dropping every selection this way is what surfaces later as
			// "empty-compute-binding", a code that says nothing about which
			// card went missing or why.
			klog.Warningf("compute: dropping card %s on node %s from the binding for app %s/%s: nothing left to allocate (card free=%s, requested=%s, remaining budget=%s)",
				item.device.ID, item.node.NodeName, appConfig.OwnerName, appConfig.AppName,
				humanBytes(deviceAvailableMemory(item.device)), humanBytes(item.memory), humanBytes(remaining))
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

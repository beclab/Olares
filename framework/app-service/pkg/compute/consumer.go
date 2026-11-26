package compute

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrConsumerRequired = errors.New("the app has several GPU consumers; name the consumer the binding is for")
	ErrUnknownConsumer  = errors.New("the app declares no such GPU consumer")
	// ErrConsumerUnplaceable is wrapped by every failure to find cards for a
	// consumer; the message names the consumer and why.
	ErrConsumerUnplaceable = errors.New("no available compute resource for GPU consumer")
)

// ConsumerRequirement is the requirement of one consumer in its chosen mode.
func ConsumerRequirement(consumer appcfg.GPUConsumer, mode appcfg.ResourceMode) Requirement {
	req := RequirementFromMode(mode)
	req.Consumer = consumer.Name
	req.PerPod = consumer.PerPod
	req.Replicas = consumer.Replicas
	return req
}

// ConsumerRequirements resolves every consumer of the app in its chosen mode,
// in placement order.
func ConsumerRequirements(appConfig *appcfg.ApplicationConfig) ([]Requirement, error) {
	consumers := appConfig.GPUConsumers()
	out := make([]Requirement, 0, len(consumers))
	for _, consumer := range consumers {
		mode, ok := appConfig.SelectedConsumerMode(consumer)
		if !ok {
			return nil, fmt.Errorf("no compute mode selected for GPU consumer %s", consumer.Name)
		}
		out = append(out, ConsumerRequirement(consumer, mode))
	}
	return out, nil
}

// ConsumerRequirementFor resolves one consumer by name. An empty name is
// accepted when the app has exactly one consumer.
func ConsumerRequirementFor(appConfig *appcfg.ApplicationConfig, name string) (Requirement, error) {
	reqs, err := ConsumerRequirements(appConfig)
	if err != nil {
		return Requirement{}, err
	}
	if name == "" {
		if len(reqs) == 1 {
			return reqs[0], nil
		}
		return Requirement{}, ErrConsumerRequired
	}
	for _, req := range reqs {
		if req.Consumer == name {
			return req, nil
		}
	}
	return Requirement{}, fmt.Errorf("%w: %s", ErrUnknownConsumer, name)
}

// AutoSelectConsumerModes picks a mode for every consumer that declares
// several. Consumers whose choice is ambiguous are returned so the caller can
// ask the user; the error is reserved for consumers no mode of which can run.
func AutoSelectConsumerModes(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, chosen map[string]string) (map[string]string, []string, error) {
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return nil, nil, fmt.Errorf("auto select consumer modes: list nodes: %w", err)
	}
	clusterGPUTypes, err := utils.GetAllGpuTypesFromNodes(&nodes)
	if err != nil {
		return nil, nil, fmt.Errorf("auto select consumer modes: %w", err)
	}
	return autoSelectConsumerModes(appConfig.GPUConsumers(), chosen, clusterGPUTypes)
}

func autoSelectConsumerModes(consumers []appcfg.GPUConsumer, chosen map[string]string, clusterGPUTypes map[string]struct{}) (map[string]string, []string, error) {
	out := make(map[string]string, len(consumers))
	var ambiguous []string
	for _, consumer := range consumers {
		if mode := chosen[consumer.Name]; mode != "" {
			if !consumerDeclaresMode(consumer, mode) {
				return nil, nil, fmt.Errorf("GPU consumer %s does not declare mode %s", consumer.Name, mode)
			}
			out[consumer.Name] = mode
			continue
		}
		modes := make([]string, 0, len(consumer.Modes))
		for _, mode := range consumer.Modes {
			modes = append(modes, mode.Mode)
		}
		mode, err := autoSelectModeFromInputs(modes, clusterGPUTypes)
		if errors.Is(err, ErrAmbiguousComputeMode) {
			ambiguous = append(ambiguous, consumer.Name)
			continue
		}
		if err != nil {
			return nil, nil, fmt.Errorf("GPU consumer %s: %w", consumer.Name, err)
		}
		out[consumer.Name] = mode
	}
	for name := range chosen {
		if !containsConsumer(consumers, name) {
			return nil, nil, fmt.Errorf("%w: %s", ErrUnknownConsumer, name)
		}
	}
	return out, ambiguous, nil
}

func consumerDeclaresMode(consumer appcfg.GPUConsumer, mode string) bool {
	for _, m := range consumer.Modes {
		if m.Mode == mode {
			return true
		}
	}
	return false
}

func containsConsumer(consumers []appcfg.GPUConsumer, name string) bool {
	for _, consumer := range consumers {
		if consumer.Name == name {
			return true
		}
	}
	return false
}

// ConsumerPlan is the install plan of one consumer: every mode it declares,
// judged against the cluster.
type ConsumerPlan struct {
	Consumer string           `json:"consumer"`
	Modes    []ModePlanResult `json:"modes"`
}

func calculateConsumerPlans(appConfig *appcfg.ApplicationConfig, nodes []Node) []ConsumerPlan {
	consumers := appConfig.GPUConsumers()
	out := make([]ConsumerPlan, 0, len(consumers))
	for _, consumer := range consumers {
		plan := ConsumerPlan{Consumer: consumer.Name}
		for _, mode := range consumer.Modes {
			plan.Modes = append(plan.Modes, EvaluateInstallMode(ConsumerRequirement(consumer, mode), nodes))
		}
		out = append(out, plan)
	}
	return out
}

// BuildConsumerInstallPlans is BuildInstallComputePlan for apps declaring
// workload accelerators.
func BuildConsumerInstallPlans(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) ([]ConsumerPlan, error) {
	nodes, err := FetchNodeComputeAllocations(ctx, c)
	if err != nil {
		return nil, err
	}
	return calculateConsumerPlans(appConfig, nodes), nil
}

// consumersInstallable runs the install placement against the current
// cluster without recording it, so the check and the allocation agree.
func consumersInstallable(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) error {
	reqs, err := ConsumerRequirements(appConfig)
	if err != nil {
		return err
	}
	nodes, err := FetchNodeComputeAllocationsExcludingApp(ctx, c, appConfig.AppName, appConfig.OwnerName)
	if err != nil {
		return err
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return err
	}
	_, _, err = placeConsumers(appConfig, reqs, nil, nodes, pressure)
	return err
}

// addedResourcesForConsumers adds every consumer's per-pod demand times its
// declared replicas to the app-level demand.
func addedResourcesForConsumers(appConfig *appcfg.ApplicationConfig, base AddedResources) AddedResources {
	reqs, err := ConsumerRequirements(appConfig)
	if err != nil {
		return base
	}
	for _, req := range reqs {
		replicas := int64(req.Replicas)
		if replicas < 1 {
			replicas = 1
		}
		base.CPU += req.RequiredCPU * replicas
		base.Memory += req.RequiredMemory * replicas
		base.Disk += req.RequiredDisk
	}
	return base
}

func withoutConsumerAllocations(allocations []Allocation, appName, owner, consumer string) []Allocation {
	out := make([]Allocation, 0, len(allocations))
	for _, allocation := range allocations {
		if allocation.AppName == appName && allocation.Owner == owner && allocation.Consumer == consumer {
			continue
		}
		out = append(out, allocation)
	}
	return out
}

func appAllocations(allocations []Allocation, appName, owner string) []Allocation {
	out := make([]Allocation, 0)
	for _, allocation := range allocations {
		if allocation.AppName == appName && allocation.Owner == owner {
			out = append(out, allocation)
		}
	}
	return out
}

// replaceAppAllocationsWith swaps every row of the app for rows, which may be
// empty.
func replaceAppAllocationsWith(allocations []Allocation, appName, owner string, rows []Allocation) []Allocation {
	next := withoutAppAllocations(allocations, appName, owner)
	return append(next, rows...)
}

// validateSplitSelection enforces what lets HAMi carve a split consumer's
// cards into equal groups that each sit on one node.
func validateSplitSelection(req Requirement, resolved []resolvedSelection) *BindingValidationResult {
	replicas := int(req.Replicas)
	cards := len(resolved)
	if replicas <= 0 || cards == 0 || cards%replicas != 0 {
		return invalidBinding("split-cards-not-divisible")
	}
	k := cards / replicas
	perNode := map[string]int{}
	for _, item := range resolved {
		perNode[item.node.NodeName]++
	}
	for _, count := range perNode {
		if count%k != 0 {
			return invalidBinding("split-group-not-on-one-node")
		}
	}
	return nil
}

// pickSplitSelection forms Replicas groups of k cards, each group on one
// node, trying the smallest k first.
func pickSplitSelection(req Requirement, availability *AvailabilityResult, pressure PressureSnapshot, opts allocationOptions) ([]BindingSelection, bool) {
	replicas := int(req.Replicas)
	if availability == nil || replicas <= 0 {
		return nil, false
	}
	type candidate struct {
		option DeviceOption
		amount int64
	}
	nodes := append([]NodeOption(nil), availability.Nodes...)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].NodeName < nodes[j].NodeName })
	perNode := make([][]candidate, len(nodes))
	maxCards := 0
	for i, node := range nodes {
		for _, supportType := range supportTypeOrder(req.Mode) {
			for _, option := range node.Devices {
				if option.SupportType != supportType || !option.Operable {
					continue
				}
				fits, amount := deviceOptionFits(req, option, pressure, FitLevelRequired, true, opts)
				if !fits {
					continue
				}
				perNode[i] = append(perNode[i], candidate{option: option, amount: amount})
			}
		}
		if len(perNode[i]) > maxCards {
			maxCards = len(perNode[i])
		}
	}
	target := requiredTargetForMode(req)
	for k := 1; k <= maxCards; k++ {
		perCard := int64(0)
		if target > 0 {
			perCard = (target + int64(replicas*k) - 1) / int64(replicas*k)
		}
		usable := func(item candidate) bool {
			if item.option.SupportType == SupportTypeMemorySlice {
				return perCard == 0 || item.amount >= perCard
			}
			return true
		}
		tryNodes := [][]int{}
		if req.SupportMultiNodes {
			all := make([]int, len(nodes))
			for i := range nodes {
				all[i] = i
			}
			tryNodes = append(tryNodes, all)
		} else {
			for i := range nodes {
				tryNodes = append(tryNodes, []int{i})
			}
		}
		for _, indexes := range tryNodes {
			var picked []BindingSelection
			var total int64
			groups := 0
			for _, i := range indexes {
				var cards []candidate
				for _, item := range perNode[i] {
					if usable(item) {
						cards = append(cards, item)
					}
				}
				for len(cards) >= k && groups < replicas {
					for _, item := range cards[:k] {
						assigned := item.amount
						if item.option.SupportType == SupportTypeMemorySlice && perCard > 0 {
							assigned = perCard
						}
						picked = append(picked, selectDevice(item.option, assigned))
						total += assigned
					}
					cards = cards[k:]
					groups++
				}
			}
			if groups == replicas && (target <= 0 || total >= target) {
				return picked, true
			}
		}
	}
	return nil, false
}

func pickConsumerSelection(req Requirement, availability *AvailabilityResult, pressure PressureSnapshot) ([]BindingSelection, bool) {
	opts := allocationOptions{checkPressure: true}
	if req.PerPod == appcfg.PerPodSplit {
		return pickSplitSelection(req, availability, pressure, opts)
	}
	return pickLaunchSelection(req, availability, pressure, opts)
}

// podsOnNode counts the pods of a consumer that land on each node given its
// rows, which is what its CPU and memory demand is multiplied by.
func podsOnNode(req Requirement, rows []Allocation) map[string]int64 {
	cards := map[string]int64{}
	for _, row := range rows {
		cards[row.NodeName]++
	}
	out := make(map[string]int64, len(cards))
	for node, n := range cards {
		switch req.PerPod {
		case appcfg.PerPodAll:
			out[node] = 1
		case appcfg.PerPodSplit:
			k := int64(len(rows)) / int64(maxInt32(req.Replicas, 1))
			if k < 1 {
				k = 1
			}
			out[node] = n / k
		default:
			out[node] = n
		}
	}
	return out
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// withAdded returns a copy of the snapshot in which the given demand is
// already running on the node, so the next consumer is judged against what
// the previous ones leave.
func (p PressureSnapshot) withAdded(nodeName string, added AddedResources) PressureSnapshot {
	usage, ok := p.UsageByNode[nodeName]
	if !ok {
		return p
	}
	next := PressureSnapshot{Threshold: p.Threshold, UsageByNode: make(map[string]prometheus.NodeResourceUsage, len(p.UsageByNode))}
	for k, v := range p.UsageByNode {
		next.UsageByNode[k] = v
	}
	if usage.CPUCapacity > 0 {
		usage.CPUUtilization += float64(added.CPU) / float64(usage.CPUCapacity)
	}
	usage.MemoryAvailable -= added.Memory
	usage.DiskAvailable -= added.Disk
	next.UsageByNode[nodeName] = usage
	return next
}

func cloneNodes(nodes []Node) []Node {
	out := make([]Node, len(nodes))
	for i, node := range nodes {
		out[i] = node
		out[i].Devices = make([]Device, len(node.Devices))
		for j, device := range node.Devices {
			out[i].Devices[j] = device
			out[i].Devices[j].Bindings = append([]Allocation(nil), device.Bindings...)
		}
	}
	return out
}

func restrictToNode(nodes []Node, name string) []Node {
	for _, node := range nodes {
		if node.NodeName == name {
			return []Node{node}
		}
	}
	return nil
}

type consumerPlacementError struct {
	consumer string
	reason   string
}

func (e *consumerPlacementError) Error() string {
	return fmt.Sprintf("%s %s: %s", ErrConsumerUnplaceable, e.consumer, e.reason)
}

func (e *consumerPlacementError) Unwrap() error { return ErrConsumerUnplaceable }

func placementReason(req Requirement, availability *AvailabilityResult, pressure PressureSnapshot) string {
	if req.PerPod == appcfg.PerPodSplit {
		return fmt.Sprintf("cannot give each of the %d replicas its own group of cards on one node that together hold the required GPU memory", req.Replicas)
	}
	if availability == nil {
		return "insufficient_resources"
	}
	added := AddedResources{CPU: req.RequiredCPU, Memory: req.RequiredMemory, Disk: req.RequiredDisk}
	var pressured, unfit []string
	for _, node := range availability.Nodes {
		if dimensions := pressure.PressuredDimensions(Node{NodeName: node.NodeName}, added); len(dimensions) > 0 {
			resources := make([]string, 0, len(dimensions))
			for _, dimension := range dimensions {
				resources = append(resources, dimension.Resource)
			}
			pressured = append(pressured, fmt.Sprintf("%s (%s)", node.NodeName, strings.Join(resources, ",")))
			continue
		}
		if node.Status != NodeStatusAvailable {
			unfit = append(unfit, fmt.Sprintf("%s (%s)", node.NodeName, node.Status))
		}
	}
	if len(pressured) == 0 {
		if availability.Reason != "" {
			return availability.Reason
		}
		return "insufficient_resources"
	}
	reason := "node-pressure: with the app's other consumers, " + strings.Join(pressured, ", ") + " would pass the pressure threshold"
	if len(unfit) > 0 {
		reason += "; no card fits on " + strings.Join(unfit, ", ")
	}
	return reason
}

// placeConsumers places every consumer in order on a working copy of the
// node view. Consumers listed in manual carry the user's own pick and are
// validated instead of picked. Either every consumer gets cards or none does.
func placeConsumers(appConfig *appcfg.ApplicationConfig, reqs []Requirement, manual map[string][]BindingSelection, nodes []Node, pressure PressureSnapshot) ([]Allocation, *BindingValidationResult, error) {
	gpuReqs := make([]Requirement, 0, len(reqs))
	for _, req := range reqs {
		if req.Mode != utils.CPUType {
			gpuReqs = append(gpuReqs, req)
		}
	}
	if len(gpuReqs) == 0 {
		return nil, nil, nil
	}
	if len(gpuReqs) > 1 && len(manual) == 0 {
		names := make([]string, 0, len(nodes))
		for _, node := range nodes {
			names = append(names, node.NodeName)
		}
		sort.Strings(names)
		for _, name := range names {
			view := restrictToNode(cloneNodes(nodes), name)
			if rows, validation, err := placeConsumersOn(appConfig, gpuReqs, nil, view, pressure); err == nil && validation == nil {
				return rows, nil, nil
			}
		}
	}
	return placeConsumersOn(appConfig, gpuReqs, manual, cloneNodes(nodes), pressure)
}

func placeConsumersOn(appConfig *appcfg.ApplicationConfig, reqs []Requirement, manual map[string][]BindingSelection, nodes []Node, pressure PressureSnapshot) ([]Allocation, *BindingValidationResult, error) {
	var out []Allocation
	for _, req := range reqs {
		availability := listAvailableForLaunch(req, nodes, pressure)
		selections, isManual := manual[req.Consumer]
		if !isManual {
			var ok bool
			selections, ok = pickConsumerSelection(req, availability, pressure)
			if !ok {
				klog.Infof("compute: GPU consumer %s/%s/%s not placed: %s", appConfig.OwnerName, appConfig.AppName, req.Consumer, explainPlacement(req, availability, pressure))
				return nil, nil, &consumerPlacementError{consumer: req.Consumer, reason: placementReason(req, availability, pressure)}
			}
		}
		rows, validation := bindAllocations(appConfig, req, selections, nodes, pressure)
		if !validation.OK {
			if isManual {
				validation.Consumer = req.Consumer
				return nil, validation, nil
			}
			return nil, nil, &consumerPlacementError{consumer: req.Consumer, reason: validation.Code}
		}
		attachBindings(nodes, rows)
		for node, pods := range podsOnNode(req, rows) {
			pressure = pressure.withAdded(node, AddedResources{
				CPU:    req.RequiredCPU * pods,
				Memory: req.RequiredMemory * pods,
			})
		}
		out = append(out, rows...)
	}
	return out, nil, nil
}

func allocateConsumersForInstall(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (*Allocation, error) {
	if err := RequireConsumerAwareHAMI(ctx, c); err != nil {
		return nil, err
	}
	reqs, err := ConsumerRequirements(appConfig)
	if err != nil {
		return nil, err
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	var placed []Allocation
	first, err := mutateAllocations(ctx, c, func(nodes []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		attachBindings(nodes, withoutAppAllocations(allocations, appConfig.AppName, appConfig.OwnerName))
		rows, _, err := placeConsumers(appConfig, reqs, nil, nodes, pressure)
		if err != nil {
			return nil, nil, err
		}
		placed = rows
		next := replaceAppAllocationsWith(allocations, appConfig.AppName, appConfig.OwnerName, rows)
		if len(rows) == 0 {
			return next, nil, nil
		}
		return next, &rows[0], nil
	})
	if err != nil {
		return nil, err
	}
	klog.Infof("compute: app %s/%s consumers placed on %s", appConfig.OwnerName, appConfig.AppName, describeAllocations(placed))
	if err := syncHAMIBindings(ctx, c, appConfig.AppName, appConfig.OwnerName, placed); err != nil {
		return nil, err
	}
	return first, nil
}

// applyConsumerBindings is ApplyBindingSelection for apps declaring workload
// accelerators: consumers named in manual keep the user's pick, every other
// consumer is placed automatically.
func applyConsumerBindings(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, manual map[string][]BindingSelection, dryRun bool) (*BindingApplyResult, error) {
	if err := RequireConsumerAwareHAMI(ctx, c); err != nil {
		return nil, err
	}
	reqs, err := ConsumerRequirements(appConfig)
	if err != nil {
		return nil, err
	}
	for name := range manual {
		if _, err := ConsumerRequirementFor(appConfig, name); err != nil {
			return nil, err
		}
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	result := &BindingApplyResult{TargetApp: appConfig.AppName, TargetOwner: appConfig.OwnerName}
	var placed []Allocation
	place := func(nodes []Node, allocations []Allocation) ([]Allocation, error) {
		attachBindings(nodes, withoutAppAllocations(allocations, appConfig.AppName, appConfig.OwnerName))
		rows, validation, err := placeConsumers(appConfig, reqs, manual, nodes, pressure)
		if err != nil {
			var placement *consumerPlacementError
			if errors.As(err, &placement) {
				result.Status = BindingApplyStatusUnavailable
				result.Validation = invalidBinding("consumer-unplaceable:" + placement.consumer)
				result.Consumer = placement.consumer
				return nil, errBindingUnavailable
			}
			return nil, err
		}
		if validation != nil {
			result.Status = BindingApplyStatusUnavailable
			result.Validation = validation
			result.Consumer = validation.Consumer
			return nil, errBindingUnavailable
		}
		placed = rows
		return rows, nil
	}
	if dryRun {
		nodes, err := loadNodeResources(ctx, c)
		if err != nil {
			return nil, err
		}
		allocations, err := loadAllocations(ctx, c)
		if err != nil {
			return nil, err
		}
		if _, err := place(nodes, allocations); err != nil {
			if errors.Is(err, errBindingUnavailable) {
				return result, nil
			}
			return nil, err
		}
		result.Status = BindingApplyStatusValid
		result.Validation = &BindingValidationResult{OK: true, Code: BindingValidationReasonValid}
		result.Allocations = placed
		return result, nil
	}
	if _, err := mutateAllocations(ctx, c, func(nodes []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		rows, err := place(nodes, allocations)
		if err != nil {
			return nil, nil, err
		}
		next := replaceAppAllocationsWith(allocations, appConfig.AppName, appConfig.OwnerName, rows)
		if len(rows) == 0 {
			return next, nil, nil
		}
		return next, &rows[0], nil
	}); err != nil {
		if errors.Is(err, errBindingUnavailable) {
			return result, nil
		}
		return nil, err
	}
	if len(placed) == 0 {
		result.Status = BindingApplyStatusNotRequired
		return result, nil
	}
	klog.Infof("compute: app %s/%s consumers bound to %s on resume", appConfig.OwnerName, appConfig.AppName, describeAllocations(placed))
	if err := syncHAMIBindings(ctx, c, appConfig.AppName, appConfig.OwnerName, placed); err != nil {
		return nil, err
	}
	result.Status = BindingApplyStatusApplied
	result.Allocations = placed
	return result, nil
}

// ApplyConsumerBindings applies a resume-time binding for an app declaring
// workload accelerators.
func ApplyConsumerBindings(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, manual map[string][]BindingSelection) (*BindingApplyResult, error) {
	return applyConsumerBindings(ctx, c, appConfig, manual, false)
}

// ValidateConsumerBindings is the write-free counterpart of
// ApplyConsumerBindings.
func ValidateConsumerBindings(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, manual map[string][]BindingSelection) (*BindingApplyResult, error) {
	return applyConsumerBindings(ctx, c, appConfig, manual, true)
}

// ConsumerAvailability lists the placement options of one consumer, counting
// the cards the app's other consumers hold.
func ConsumerAvailability(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, consumer string) (*AvailabilityResult, error) {
	req, err := ConsumerRequirementFor(appConfig, consumer)
	if err != nil {
		return nil, err
	}
	pressure, err := FetchPressureSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := loadNodeResources(ctx, c)
	if err != nil {
		return nil, err
	}
	allocations, err := loadAllocations(ctx, c)
	if err != nil {
		return nil, err
	}
	attachBindings(nodes, withoutConsumerAllocations(allocations, appConfig.AppName, appConfig.OwnerName, req.Consumer))
	return listAvailableForLaunch(req, nodes, pressure), nil
}

// LegacyRowsFitConsumer reports whether rows written for an app declaring
// spec.accelerator can be carried over to the single consumer its new
// version declares.
func LegacyRowsFitConsumer(req Requirement, rows []Allocation) bool {
	if len(rows) == 0 {
		return false
	}
	for _, row := range rows {
		if row.Consumer != "" || row.Mode != req.Mode {
			return false
		}
	}
	if req.PerPod != appcfg.PerPodSplit {
		return true
	}
	resolved := make([]resolvedSelection, 0, len(rows))
	for _, row := range rows {
		resolved = append(resolved, resolvedSelection{node: Node{NodeName: row.NodeName}})
	}
	return validateSplitSelection(req, resolved) == nil
}

// MigrateLegacyAllocations moves the app's consumer-less rows onto the given
// consumer and rebuilds its bindings under the consumer-aware names.
func MigrateLegacyAllocations(ctx context.Context, c client.Client, appName, owner, consumer string) error {
	var migrated []Allocation
	if _, err := mutateAllocations(ctx, c, func(_ []Node, allocations []Allocation) ([]Allocation, *Allocation, error) {
		rows := appAllocations(allocations, appName, owner)
		for i := range rows {
			rows[i].Consumer = consumer
		}
		migrated = rows
		return replaceAppAllocationsWith(allocations, appName, owner, rows), nil, nil
	}); err != nil {
		return err
	}
	return syncHAMIBindings(ctx, c, appName, owner, migrated)
}

// consumerSelections maps a resume body's flat binding onto the app's only
// consumer; an app with several consumers must name them.
func consumerSelections(appConfig *appcfg.ApplicationConfig, selections []BindingSelection) (map[string][]BindingSelection, error) {
	if len(selections) == 0 {
		return nil, nil
	}
	req, err := ConsumerRequirementFor(appConfig, "")
	if err != nil {
		return nil, err
	}
	return map[string][]BindingSelection{req.Consumer: selections}, nil
}

// PrepareConsumerUpgrade runs before an app is upgraded to a version
// declaring workload accelerators. It carries the previous mode choices over,
// moves rows written for spec.accelerator onto the single consumer that
// replaces it when they still fit, and re-places every consumer whose rows
// no longer fit. A stopped app keeps no rows; resume places it.
func PrepareConsumerUpgrade(ctx context.Context, c client.Client, prev, next *appcfg.ApplicationConfig) error {
	if err := RequireConsumerAwareHAMI(ctx, c); err != nil {
		return err
	}
	consumers := next.GPUConsumers()
	carried := map[string]string{}
	for _, consumer := range consumers {
		if mode := prev.SelectedGpuTypes[consumer.Name]; mode != "" && consumerDeclaresMode(consumer, mode) {
			carried[consumer.Name] = mode
		}
	}
	if !prev.HasGPUConsumers() && len(consumers) == 1 && consumerDeclaresMode(consumers[0], prev.SelectedGpuType) {
		carried[consumers[0].Name] = prev.SelectedGpuType
	}
	chosen, pending, err := AutoSelectConsumerModes(ctx, c, next, carried)
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return fmt.Errorf("choose a compute mode for GPU consumers %v before upgrading", pending)
	}
	next.SelectedGpuTypes = chosen

	reqs, err := ConsumerRequirements(next)
	if err != nil {
		return err
	}
	rows, err := FindAllocationsForApp(ctx, c, next.AppName, next.OwnerName)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	var gpuReqs []Requirement
	for _, req := range reqs {
		if req.Mode != utils.CPUType {
			gpuReqs = append(gpuReqs, req)
		}
	}
	if len(gpuReqs) == 1 && LegacyRowsFitConsumer(gpuReqs[0], rows) {
		klog.Infof("compute: app %s/%s rows move to GPU consumer %s on upgrade", next.OwnerName, next.AppName, gpuReqs[0].Consumer)
		return MigrateLegacyAllocations(ctx, c, next.AppName, next.OwnerName, gpuReqs[0].Consumer)
	}
	manual := map[string][]BindingSelection{}
	for _, req := range gpuReqs {
		var selections []BindingSelection
		for _, row := range rows {
			if row.Consumer == req.Consumer && row.Mode == req.Mode {
				selections = append(selections, BindingSelection{NodeName: row.NodeName, DeviceID: row.DeviceID, Memory: row.Memory})
			}
		}
		if len(selections) > 0 {
			manual[req.Consumer] = selections
		}
	}
	if len(manual) == len(gpuReqs) && len(rows) == countSelections(manual) {
		return nil
	}
	result, err := ApplyConsumerBindings(ctx, c, next, manual)
	if err != nil {
		return err
	}
	if result.Status == BindingApplyStatusUnavailable {
		code := ""
		if result.Validation != nil {
			code = result.Validation.Code
		}
		return fmt.Errorf("the upgraded GPU consumers cannot be placed: %s", code)
	}
	return nil
}

func countSelections(manual map[string][]BindingSelection) int {
	n := 0
	for _, selections := range manual {
		n += len(selections)
	}
	return n
}

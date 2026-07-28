package compute

import (
	"fmt"
	"math"
	"sort"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"k8s.io/apimachinery/pkg/api/resource"
)

func SimulatePreflight(demands []PreflightDemand, snapshot PreflightSnapshot) (PreflightReport, error) {
	state := clonePreflightSnapshot(snapshot)
	if err := validatePreflightInputs(demands, state); err != nil {
		return PreflightReport{}, err
	}
	for _, demand := range demands {
		failure, placement, err := simulateDemand(demand, &state)
		if err != nil {
			return PreflightReport{}, fmt.Errorf("simulate demand %q: %w", demand.ID, err)
		}
		if failure != nil {
			failure.FailedDemandID = demand.ID
			return *failure, nil
		}
		if err := applyDemand(demand, placement, &state); err != nil {
			return PreflightReport{}, fmt.Errorf("apply demand %q: %w", demand.ID, err)
		}
	}
	return PreflightReport{Installable: true}, nil
}

func simulateDemand(demand PreflightDemand, state *PreflightSnapshot) (*PreflightReport, PreflightPlacement, error) {
	added := addedResourcesFromRequirement(demand.Requirement)
	required := apputils.ResourceState{CPU: added.CPU, Memory: added.Memory, Disk: added.Disk}
	// Preflight intentionally checks current pressure on every dimension,
	// including dimensions the demand leaves at zero.
	dimensions := apputils.AllResourceDimensions

	resourcePressure, evalErr := apputils.EvaluatePhysicalCapacity(required, state.Cluster, dimensions)
	if failure, err := resourceFailure(PreflightReasonClusterCapacity, resourcePressure, evalErr); failure != nil || err != nil {
		return failure, PreflightPlacement{}, err
	}
	resourcePressure, evalErr = apputils.EvaluateClusterPressure(required, state.Cluster, dimensions)
	if failure, err := resourceFailure(PreflightReasonClusterPressure, resourcePressure, evalErr); failure != nil || err != nil {
		return failure, PreflightPlacement{}, err
	}
	if demand.Owner != PreflightSharedOwner {
		resourcePressure, evalErr = apputils.EvaluateOwnerPressure(required, state.Owners[demand.Owner], dimensions)
		if failure, err := resourceFailure(PreflightReasonOwnerQuota, resourcePressure, evalErr); failure != nil || err != nil {
			return failure, PreflightPlacement{}, err
		}
	}
	cpu := *resource.NewMilliQuantity(state.K8sAvailable.CPU, resource.DecimalSI)
	memory := *resource.NewQuantity(state.K8sAvailable.Memory, resource.BinarySI)
	resourcePressure, evalErr = apputils.EvaluateK8sRequest(required, dimensions, cpu, memory)
	if failure, err := resourceFailure(PreflightReasonK8sRequest, resourcePressure, evalErr); failure != nil || err != nil {
		return failure, PreflightPlacement{}, err
	}

	hardPlacement, ok := pickPreflightPlacement(demand, added, state.Nodes, PressureSnapshot{}, false)
	if !ok {
		return &PreflightReport{Reason: PreflightReasonUnschedulable}, PreflightPlacement{}, nil
	}
	placement, ok := pickPreflightPlacement(demand, added, state.Nodes, state.Pressure, true)
	if ok {
		return nil, placement, nil
	}
	pressure, err := placementPressure(added, hardPlacement, state)
	if err != nil {
		return nil, PreflightPlacement{}, err
	}
	return &PreflightReport{Reason: PreflightReasonNodePressure, Pressure: pressure}, PreflightPlacement{}, nil
}

func resourceFailure(reason string, pressure []apputils.ResourcePressure, err error) (*PreflightReport, error) {
	if err != nil {
		return nil, err
	}
	if len(pressure) == 0 {
		return nil, nil
	}
	return &PreflightReport{Reason: reason, Pressure: []PreflightPressure{{Source: reason, Dimensions: pressure}}}, nil
}

func pickPreflightPlacement(demand PreflightDemand, added AddedResources, nodes []Node, pressure PressureSnapshot, checkPressure bool) (PreflightPlacement, bool) {
	if demand.Requirement.Mode == utils.CPUType {
		for _, node := range nodes {
			if usableCPUNodeMemory(node.memoryCapacity) < demand.Requirement.RequiredMemory {
				continue
			}
			if checkPressure && pressure.WouldPressure(node, added) {
				continue
			}
			return PreflightPlacement{NodeNames: []string{node.NodeName}}, true
		}
		return PreflightPlacement{}, false
	}
	app := &appcfg.ApplicationConfig{
		AppID: demand.AppID, AppName: demand.Application, OwnerName: demand.Owner,
		SelectedGpuType: demand.Requirement.Mode,
	}
	allocations, ok := pickAllocations(app, demand.Requirement, nodes, pressure, allocationOptions{
		checkPressure: checkPressure,
		deterministic: true,
		pressureAdded: &added,
	})
	if !ok {
		return PreflightPlacement{}, false
	}
	return placementFromAllocations(allocations), true
}

func placementFromAllocations(allocations []Allocation) PreflightPlacement {
	nodeSet := make(map[string]struct{})
	for _, allocation := range allocations {
		nodeSet[allocation.NodeName] = struct{}{}
	}
	nodeNames := make([]string, 0, len(nodeSet))
	for nodeName := range nodeSet {
		nodeNames = append(nodeNames, nodeName)
	}
	sort.Strings(nodeNames)
	return PreflightPlacement{NodeNames: nodeNames, Allocations: allocations}
}

func placementPressure(added AddedResources, placement PreflightPlacement, state *PreflightSnapshot) ([]PreflightPressure, error) {
	out := make([]PreflightPressure, 0, len(placement.NodeNames))
	for _, nodeName := range placement.NodeNames {
		node, ok := findNode(state.Nodes, nodeName)
		if !ok {
			return nil, fmt.Errorf("placement node %q is missing", nodeName)
		}
		projected, err := placementAddedResources(added, nodeName, placement.Allocations, state.Nodes)
		if err != nil {
			return nil, err
		}
		if err := validateNodeProjection(state.Pressure.UsageByNode[nodeName], projected); err != nil {
			return nil, fmt.Errorf("node %q: %w", nodeName, err)
		}
		dimensions := state.Pressure.PressuredDimensions(node, projected)
		if len(dimensions) > 0 {
			out = append(out, PreflightPressure{Source: "node/" + nodeName, Dimensions: dimensions})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("pressure placement failed without pressured dimensions")
	}
	return out, nil
}

func validateNodeProjection(usage prometheus.NodeResourceUsage, added AddedResources) error {
	values := [][2]int64{
		{nodeUsedCPU(usage), added.CPU},
		{usage.MemoryCapacity - usage.MemoryAvailable, added.Memory},
		{usage.DiskCapacity - usage.DiskAvailable, added.Disk},
	}
	for _, value := range values {
		if _, ok := checkedAdd(value[0], value[1]); !ok {
			return fmt.Errorf("resource projection overflows")
		}
	}
	return nil
}

func placementAddedResources(added AddedResources, nodeName string, allocations []Allocation, nodes []Node) (AddedResources, error) {
	for _, allocation := range allocations {
		if allocation.NodeName != nodeName {
			continue
		}
		node, ok := findNode(nodes, nodeName)
		if !ok {
			return AddedResources{}, fmt.Errorf("node %q is missing", nodeName)
		}
		device, ok := findDevice(node, allocation.DeviceID)
		if !ok {
			return AddedResources{}, fmt.Errorf("device %q is missing", allocation.DeviceID)
		}
		var valid bool
		added.Memory, valid = checkedAdd(added.Memory, timeSliceAddedMemory(device))
		if !valid {
			return AddedResources{}, fmt.Errorf("host memory projection overflows")
		}
	}
	return added, nil
}

func applyDemand(demand PreflightDemand, placement PreflightPlacement, state *PreflightSnapshot) error {
	added := addedResourcesFromRequirement(demand.Requirement)
	if err := applyPlacement(added, placement, state); err != nil {
		return err
	}
	if err := addMetricUsage(state.Cluster, added); err != nil {
		return err
	}
	if demand.Owner != PreflightSharedOwner {
		if err := addMetricUsage(state.Owners[demand.Owner], added); err != nil {
			return err
		}
	}
	if added.CPU > state.K8sAvailable.CPU || added.Memory > state.K8sAvailable.Memory {
		return fmt.Errorf("kubernetes availability underflows")
	}
	state.K8sAvailable.CPU -= added.CPU
	state.K8sAvailable.Memory -= added.Memory
	return nil
}

func applyPlacement(added AddedResources, placement PreflightPlacement, state *PreflightSnapshot) error {
	for _, nodeName := range placement.NodeNames {
		projected, err := placementAddedResources(added, nodeName, placement.Allocations, state.Nodes)
		if err != nil {
			return err
		}
		if err := projectPressure(&state.Pressure, nodeName, projected); err != nil {
			return err
		}
	}
	for _, allocation := range placement.Allocations {
		for nodeIndex := range state.Nodes {
			if state.Nodes[nodeIndex].NodeName != allocation.NodeName {
				continue
			}
			for deviceIndex := range state.Nodes[nodeIndex].Devices {
				if state.Nodes[nodeIndex].Devices[deviceIndex].ID == allocation.DeviceID {
					state.Nodes[nodeIndex].Devices[deviceIndex].Bindings = append(
						state.Nodes[nodeIndex].Devices[deviceIndex].Bindings, allocation)
				}
			}
		}
	}
	return nil
}

func projectPressure(snapshot *PressureSnapshot, nodeName string, added AddedResources) error {
	usage, ok := snapshot.UsageByNode[nodeName]
	if !ok {
		return fmt.Errorf("node metrics for %q are missing", nodeName)
	}
	projectedCPU, ok := checkedAdd(nodeUsedCPU(usage), added.CPU)
	if !ok {
		return fmt.Errorf("cpu projection overflows")
	}
	if usage.CPUCapacity > 0 {
		usage.CPUUtilization = float64(projectedCPU) / float64(usage.CPUCapacity)
	}
	if added.Memory > usage.MemoryAvailable || added.Disk > usage.DiskAvailable {
		return fmt.Errorf("node availability underflows")
	}
	usage.MemoryAvailable -= added.Memory
	usage.DiskAvailable -= added.Disk
	snapshot.UsageByNode[nodeName] = usage
	return nil
}

func validatePreflightInputs(demands []PreflightDemand, snapshot PreflightSnapshot) error {
	if snapshot.Cluster == nil {
		return fmt.Errorf("cluster metrics are missing")
	}
	if snapshot.K8sAvailable.CPU < 0 || snapshot.K8sAvailable.Memory < 0 {
		return fmt.Errorf("kubernetes availability is invalid")
	}
	if err := ValidatePressureSnapshot(snapshot.Nodes, snapshot.Pressure); err != nil {
		return err
	}
	if _, err := apputils.EvaluateClusterPressure(apputils.ResourceState{}, snapshot.Cluster, apputils.AllResourceDimensions); err != nil {
		return fmt.Errorf("cluster metrics are invalid: %w", err)
	}
	for _, demand := range demands {
		if err := validatePreflightDemand(demand); err != nil {
			return err
		}
		if demand.Owner == PreflightSharedOwner {
			continue
		}
		metrics, ok := snapshot.Owners[demand.Owner]
		if !ok {
			return fmt.Errorf("owner metrics for %q are missing", demand.Owner)
		}
		if _, err := apputils.EvaluateOwnerPressure(apputils.ResourceState{}, metrics, apputils.AllResourceDimensions); err != nil {
			return fmt.Errorf("owner metrics for %q are invalid: %w", demand.Owner, err)
		}
	}
	return nil
}

func validatePreflightDemand(demand PreflightDemand) error {
	req := demand.Requirement
	if demand.ID == "" || demand.Owner == "" || demand.Application == "" || req.Mode == "" {
		return fmt.Errorf("demand identity is incomplete")
	}
	if req.RequiredCPU < 0 || req.RequiredGPU < 0 || req.LimitedGPU < 0 ||
		req.RequiredMemory < 0 || req.LimitedMemory < 0 || req.RequiredDisk < 0 {
		return fmt.Errorf("demand resources must not be negative")
	}
	if (req.LimitedGPU != 0 && req.LimitedGPU < req.RequiredGPU) ||
		(req.LimitedMemory != 0 && req.LimitedMemory < req.RequiredMemory) {
		return fmt.Errorf("demand limits must cover requirements")
	}
	return nil
}

func addMetricUsage(metrics *prometheus.ClusterMetrics, added AddedResources) error {
	if metrics == nil {
		return fmt.Errorf("metrics are missing")
	}
	metrics.CPU.Usage += float64(added.CPU) / 1000
	metrics.Memory.Usage += float64(added.Memory)
	metrics.Disk.Usage += float64(added.Disk)
	if math.IsInf(metrics.CPU.Usage, 0) || math.IsInf(metrics.Memory.Usage, 0) || math.IsInf(metrics.Disk.Usage, 0) {
		return fmt.Errorf("metric projection overflows")
	}
	return nil
}

func addedResourcesFromRequirement(req Requirement) AddedResources {
	return AddedResources{CPU: req.RequiredCPU, Memory: req.RequiredMemory, Disk: req.RequiredDisk}
}

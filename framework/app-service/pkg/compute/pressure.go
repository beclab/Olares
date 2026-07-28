package compute

import (
	"context"
	"fmt"
	"math"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

const defaultPressureThreshold = 0.90

const (
	PressureResourceCPU    = "cpu"
	PressureResourceMemory = "memory"
	PressureResourceDisk   = "disk"
)

type DimensionPressure = apputils.ResourcePressure

func FetchPressureSnapshot(ctx context.Context) (PressureSnapshot, error) {
	usage, err := prometheus.GetNodeResourceUsage(ctx)
	if err != nil {
		return PressureSnapshot{}, err
	}
	return PressureSnapshot{
		Threshold:   defaultPressureThreshold,
		UsageByNode: usage,
	}, nil
}

// Evaluate returns the per-dimension pressure breakdown (cpu, memory,
// disk) for adding `added` to `node`. WouldPressure is the boolean
// reduction of this, and PressuredDimensions filters it down to the
// dimensions that actually block the placement.
func (p PressureSnapshot) Evaluate(node Node, added AddedResources) []DimensionPressure {
	threshold := p.Threshold
	if threshold == 0 {
		threshold = defaultPressureThreshold
	}
	usage, known := p.UsageByNode[node.NodeName]
	usedCPU := nodeUsedCPU(usage)
	usedMemory := usage.MemoryCapacity - usage.MemoryAvailable
	if usedMemory < 0 {
		usedMemory = 0
	}
	usedDisk := usage.DiskCapacity - usage.DiskAvailable
	if usedDisk < 0 {
		usedDisk = 0
	}
	return []DimensionPressure{
		evaluateDimension(PressureResourceCPU, added.CPU, usedCPU, usage.CPUCapacity, threshold, known),
		evaluateDimension(PressureResourceMemory, added.Memory, usedMemory, usage.MemoryCapacity, threshold, known),
		evaluateDimension(PressureResourceDisk, added.Disk, usedDisk, usage.DiskCapacity, threshold, known),
	}
}

func nodeUsedCPU(usage prometheus.NodeResourceUsage) int64 {
	util := usage.CPUUtilization
	if math.IsNaN(util) || math.IsInf(util, 0) {
		util = 1.0
	}
	if util < 0 {
		util = 0
	}
	return int64(float64(usage.CPUCapacity) * util)
}

func evaluateDimension(resource string, required, used, capacity int64, threshold float64, known bool) DimensionPressure {
	d := DimensionPressure{
		Resource: resource,
		Required: required,
		Used:     used,
		Capacity: capacity,
	}
	if headroom := float64(capacity)*threshold - float64(used); headroom > 0 {
		d.Available = int64(headroom)
	}
	// A node we have metrics for but that reports non-positive capacity on a
	// dimension the app actually needs (e.g. a NotReady node or a stale/zeroed
	// metric) cannot host the app. Without this, exceedsPressure's `total <= 0`
	// short-circuit would report "no pressure" and the node would look like it
	// had infinite headroom.
	if known && required > 0 && capacity <= 0 {
		d.Pressured = true
		return d
	}
	projected, ok := checkedAdd(used, required)
	if !ok {
		d.Pressured = true
		return d
	}
	d.Pressured = exceedsPressure(projected, capacity, threshold)
	return d
}

func (p PressureSnapshot) WouldPressure(node Node, added AddedResources) bool {
	for _, d := range p.Evaluate(node, added) {
		if d.Pressured {
			return true
		}
	}
	return false
}

// PressuredDimensions returns only the dimensions that would exceed the
// pressure threshold when `added` is placed on `node`. It is non-empty
// exactly when WouldPressure is true, so callers can use its length as
// the rejection signal while also surfacing which resource(s) fell short
// and by how much.
func (p PressureSnapshot) PressuredDimensions(node Node, added AddedResources) []DimensionPressure {
	var out []DimensionPressure
	for _, d := range p.Evaluate(node, added) {
		if d.Pressured {
			out = append(out, d)
		}
	}
	return out
}

func ValidatePressureSnapshot(nodes []Node, pressure PressureSnapshot) error {
	threshold := pressure.Threshold
	if threshold != 0 && (math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold <= 0 || threshold > 1) {
		return fmt.Errorf("node pressure threshold is invalid")
	}
	for _, node := range nodes {
		usage, ok := pressure.UsageByNode[node.NodeName]
		if !ok || !validNodeUsage(usage) {
			return fmt.Errorf("node metrics for %q are invalid", node.NodeName)
		}
	}
	return nil
}

func validNodeUsage(usage prometheus.NodeResourceUsage) bool {
	return usage.CPUCapacity >= 0 && !math.IsNaN(usage.CPUUtilization) && !math.IsInf(usage.CPUUtilization, 0) &&
		usage.CPUUtilization >= 0 && usage.CPUUtilization <= 1 &&
		usage.MemoryCapacity >= 0 && usage.MemoryAvailable >= 0 && usage.MemoryAvailable <= usage.MemoryCapacity &&
		usage.DiskCapacity >= 0 && usage.DiskAvailable >= 0 && usage.DiskAvailable <= usage.DiskCapacity
}

func checkedAdd(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b || b < 0 && a < math.MinInt64-b {
		return 0, false
	}
	return a + b, true
}

func exceedsPressure(used, total int64, threshold float64) bool {
	if total <= 0 {
		return false
	}
	return float64(used) > float64(total)*threshold
}

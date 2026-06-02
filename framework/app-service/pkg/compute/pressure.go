package compute

import (
	"context"
	"math"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
)

const defaultPressureThreshold = 0.90

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

func (p PressureSnapshot) WouldPressure(node Node, added AddedResources) bool {
	threshold := p.Threshold
	if threshold == 0 {
		threshold = defaultPressureThreshold
	}
	usage, known := p.UsageByNode[node.NodeName]
	// A node we have metrics for but that reports non-positive capacity on a
	// dimension the app actually needs (e.g. a NotReady node or a stale/zeroed
	// metric) cannot host the app. Without this, exceedsPressure's `total <= 0`
	// short-circuit would report "no pressure" and nodePressureValidator would
	// treat the node as infinite headroom and schedule onto it.
	if known {
		if (added.CPU > 0 && usage.CPUCapacity <= 0) ||
			(added.Memory > 0 && usage.MemoryCapacity <= 0) ||
			(added.Disk > 0 && usage.DiskCapacity <= 0) {
			return true
		}
	}
	// Sanitise the CPU utilisation before converting to a used-core count. A
	// 0/0 monitoring ratio yields NaN (and a misbehaving exporter can produce
	// Inf or a negative value); int64(NaN) is implementation-defined in Go, so
	// an unsanitised value makes the whole pressure decision undefined. Assume
	// the node is fully used when the metric is non-finite, so a broken metric
	// never looks like free headroom.
	util := usage.CPUUtilization
	if math.IsNaN(util) || math.IsInf(util, 0) {
		util = 1.0
	}
	if util < 0 {
		util = 0
	}
	usedCPU := int64(float64(usage.CPUCapacity) * util)
	usedMemory := usage.MemoryCapacity - usage.MemoryAvailable
	if usedMemory < 0 {
		usedMemory = 0
	}
	usedDisk := usage.DiskCapacity - usage.DiskAvailable
	if usedDisk < 0 {
		usedDisk = 0
	}
	return exceedsPressure(usedCPU+added.CPU, usage.CPUCapacity, threshold) ||
		exceedsPressure(usedMemory+added.Memory, usage.MemoryCapacity, threshold) ||
		exceedsPressure(usedDisk+added.Disk, usage.DiskCapacity, threshold)
}

func exceedsPressure(used, total int64, threshold float64) bool {
	if total <= 0 {
		return false
	}
	return float64(used) > float64(total)*threshold
}

package compute

import (
	"context"

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
	usage := p.UsageByNode[node.NodeName]
	usedCPU := int64(float64(usage.CPUCapacity) * usage.CPUUtilization)
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

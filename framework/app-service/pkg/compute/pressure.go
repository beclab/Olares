package compute

import (
	"context"
	"math"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
)

const defaultPressureThreshold = 0.90

const (
	PressureResourceCPU    = "cpu"
	PressureResourceMemory = "memory"
	PressureResourceDisk   = "disk"
)

// DimensionPressure describes, for a single resource dimension, how the
// node's current usage plus the app's requested amount compares against
// the pressure threshold:
//
//   - Required is how much the app wants to add on this dimension.
//   - Used / Capacity are the node's current consumption and total.
//   - Available is the remaining headroom UNDER the threshold before the
//     request is added (capacity*threshold - used), clamped to >= 0. This
//     is the "how much can still be placed" figure the frontend renders.
//   - Pressured reports whether adding Required would push the dimension
//     past the threshold (or the node reports unusable capacity for a
//     dimension the app needs).
type DimensionPressure struct {
	Resource  string `json:"resource"`
	Required  int64  `json:"required"`
	Used      int64  `json:"used"`
	Capacity  int64  `json:"capacity"`
	Available int64  `json:"available"`
	Pressured bool   `json:"pressured"`
}

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
	return []DimensionPressure{
		evaluateDimension(PressureResourceCPU, added.CPU, usedCPU, usage.CPUCapacity, threshold, known),
		evaluateDimension(PressureResourceMemory, added.Memory, usedMemory, usage.MemoryCapacity, threshold, known),
		evaluateDimension(PressureResourceDisk, added.Disk, usedDisk, usage.DiskCapacity, threshold, known),
	}
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
	d.Pressured = exceedsPressure(used+required, capacity, threshold)
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

func exceedsPressure(used, total int64, threshold float64) bool {
	if total <= 0 {
		return false
	}
	return float64(used) > float64(total)*threshold
}

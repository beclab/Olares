package compute

import (
	"math"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
)

// TestProbe_WouldPressure_DegenerateNodeTreatedAsHeadroom is a bug-hunting
// probe. A node that reports zero (or negative) capacity -- e.g. a NotReady
// node or one whose prometheus metrics are missing -- should NOT be treated as
// having free headroom, otherwise nodePressureValidator (which passes as soon
// as ANY node has `!WouldPressure`) will happily schedule an app onto a node
// that physically cannot host it.
//
// Candidate bug S4-1: exceedsPressure() short-circuits `if total <= 0 { return
// false }`, so a zero/negative-capacity node yields WouldPressure==false ("no
// pressure") for every request, making it look like infinite headroom.
func TestProbe_WouldPressure_DegenerateNodeTreatedAsHeadroom(t *testing.T) {
	// Regression for fixed bug S4-1 (a node present in monitoring but reporting
	// zero capacity was treated as infinite headroom).
	snap := PressureSnapshot{
		Threshold: 0.9,
		UsageByNode: map[string]prometheus.NodeResourceUsage{
			"broken": {CPUCapacity: 0, MemoryCapacity: 0, DiskCapacity: 0},
		},
	}
	node := Node{NodeName: "broken"}

	if !snap.WouldPressure(node, AddedResources{CPU: 1_000_000, Memory: 1 << 40, Disk: 1 << 40}) {
		t.Errorf("a zero-capacity node was treated as having headroom for a huge request; nodePressureValidator would schedule onto it")
	}
}

// TestProbe_WouldPressure_NonFiniteUtilizationTreatedAsFull is a regression for
// fixed bug S4-2. A non-finite CPU utilisation (NaN from a 0/0 monitoring
// ratio, or Inf) used to flow into int64(float64(cap)*util), whose result is
// implementation-defined in Go, making the pressure decision undefined. The
// utilisation is now sanitised to "fully used", so such a node is never
// reported as free headroom for a CPU request.
func TestProbe_WouldPressure_NonFiniteUtilizationTreatedAsFull(t *testing.T) {
	for _, util := range []float64{math.NaN(), math.Inf(1)} {
		snap := PressureSnapshot{
			Threshold: 0.9,
			UsageByNode: map[string]prometheus.NodeResourceUsage{
				"bad": {CPUCapacity: 1000, CPUUtilization: util},
			},
		}
		node := Node{NodeName: "bad"}
		if !snap.WouldPressure(node, AddedResources{CPU: 100}) {
			t.Errorf("util=%v: node with non-finite utilisation reported as having headroom", util)
		}
	}
}

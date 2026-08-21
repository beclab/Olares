package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// An app's declared memory already covers the VRAM that a unified-memory card
// swaps into host memory, so a time-slice card's own physical memory must not
// be added on top. Charging the app for the whole card made a small-VRAM app on
// a big card look as expensive as one that needs the entire card, and blocked
// apps the node could comfortably host.
func TestDeviceFitsLevelIgnoresTimeSliceCardMemory(t *testing.T) {
	const gib = int64(1 << 30)
	node := Node{NodeName: "gpu-node"}
	device := Device{
		ID:          "gpu0",
		Memory:      10 * gib,
		Health:      deviceHealthYes,
		SupportType: SupportTypeTimeSlice,
	}
	// 80Gi of 100Gi is in use, so the 0.9 threshold leaves 10Gi of headroom.
	pressure := PressureSnapshot{
		Threshold: 0.9,
		UsageByNode: map[string]prometheus.NodeResourceUsage{
			node.NodeName: {
				CPUCapacity:     10_000,
				MemoryCapacity:  100 * gib,
				MemoryAvailable: 20 * gib,
				DiskCapacity:    100 * gib,
				DiskAvailable:   100 * gib,
				CPUUtilization:  0,
			},
		},
	}
	opts := allocationOptions{checkPressure: true}

	// 5Gi of declared memory fits in the 10Gi of headroom. Under the old
	// accounting the card's 10Gi was added and this was rejected.
	req := Requirement{
		Mode:           utils.NvidiaCardType,
		RequiredMemory: 5 * gib,
		RequiredGPU:    gib,
	}
	if fits, _ := deviceFitsLevelWithPressure(req, node, device, pressure, FitLevelRequired, false, opts); !fits {
		t.Fatal("a time-slice card must not be charged its own physical memory on top of the app's request")
	}

	// The app's own declared memory is now the only thing that can exhaust
	// the node's headroom.
	req.RequiredMemory = 11 * gib
	if fits, _ := deviceFitsLevelWithPressure(req, node, device, pressure, FitLevelRequired, false, opts); fits {
		t.Fatal("a request larger than the node's headroom must still be rejected")
	}
}

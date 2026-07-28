package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

func TestDeviceFitsLevelTimeSliceMemoryForZeroRequirement(t *testing.T) {
	const gib = int64(1 << 30)
	node := Node{NodeName: "gpu-node"}
	device := Device{
		ID:          "gpu0",
		Memory:      10 * gib,
		Health:      deviceHealthYes,
		SupportType: SupportTypeTimeSlice,
	}
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
	req := Requirement{
		Mode:           utils.NvidiaCardType,
		RequiredMemory: 5 * gib,
	}

	fits, _ := deviceFitsLevel(req, node, device, pressure, FitLevelRequired, false, 0)
	if !fits {
		t.Fatal("zero GPU requirement should preserve the legacy fit decision without time-slice host memory")
	}

	req.RequiredGPU = gib
	fits, _ = deviceFitsLevel(req, node, device, pressure, FitLevelRequired, false, 0)
	if fits {
		t.Fatal("positive GPU requirement must include time-slice host memory")
	}
}

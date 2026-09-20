package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// TestAutoPickOnlyChoosesDevicesTheResumeViewOffers pins the property that ties
// the two placement flows together: install and resume start from the same
// availability view, so a card install assigns itself must be one the resume
// picker would have let a user click. The cluster below mixes every reason a
// card is unusable — already taken exclusively, unhealthy, too little VRAM left
// — so a picker running against the raw node list instead of the view would
// have plenty of chances to reach past what the view exposes.
func TestAutoPickOnlyChoosesDevicesTheResumeViewOffers(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice", SelectedGpuType: utils.NvidiaCardType}
	req := Requirement{
		Mode:           utils.NvidiaCardType,
		RequiredGPU:    8 * gi,
		LimitedGPU:     8 * gi,
		RequiredMemory: gi,
		LimitedMemory:  gi,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a",
			Device{ID: "a-gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive,
				Bindings: []Allocation{{AppName: "other", Owner: "bob"}}},
			Device{ID: "a-gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
		),
		nvidiaNode("nvidia-b",
			Device{ID: "b-gpu0", Memory: 16 * gi, Health: deviceHealthNo, SupportType: SupportTypeExclusive},
			Device{ID: "b-gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
		),
		nvidiaNode("nvidia-c",
			Device{ID: "c-gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice,
				Bindings: []Allocation{{AppName: "other", Owner: "bob", Memory: 12 * gi}}},
		),
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemorySlice),
	}

	offered := map[string]bool{}
	for _, node := range listAvailableForLaunch(req, nodes, PressureSnapshot{}).Nodes {
		for _, device := range node.Devices {
			if device.Operable {
				offered[node.NodeName+"/"+device.DeviceID] = true
			}
		}
	}
	if len(offered) != 2 {
		t.Fatalf("expected the two free exclusive cards to be selectable, got %#v", offered)
	}

	// The picker breaks ties at random, so repeat enough to see every card it
	// is willing to choose.
	chosen := map[string]bool{}
	for i := 0; i < 50; i++ {
		picked, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
		if !ok {
			t.Fatalf("expected a placement on one of the free exclusive cards")
		}
		for _, allocation := range picked {
			key := allocation.NodeName + "/" + allocation.DeviceID
			if !offered[key] {
				t.Fatalf("install picked %s, which the resume view does not offer as selectable", key)
			}
			chosen[key] = true
		}
	}
	if len(chosen) != len(offered) {
		t.Fatalf("expected the picker to spread across every offered card, got %#v", chosen)
	}
}

// TestAutoPickPassesManualBindingValidation covers the other half of the
// unification: AllocateForInstall now runs its own pick through the validation
// a user-submitted binding goes through, so the picker must never produce a
// selection that validation would turn away.
func TestAutoPickPassesManualBindingValidation(t *testing.T) {
	cases := []struct {
		name  string
		req   Requirement
		nodes []Node
	}{
		{
			name: "single exclusive card",
			req:  Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 8 * gi, LimitedGPU: 8 * gi, RequiredMemory: gi, LimitedMemory: gi},
			nodes: []Node{nvidiaNode("nvidia-a",
				Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive})},
		},
		{
			name: "memory slice card",
			req:  Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 8 * gi, LimitedGPU: 8 * gi, RequiredMemory: gi, LimitedMemory: gi},
			nodes: []Node{nvidiaNode("nvidia-a",
				Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice})},
		},
		{
			name: "multi card on one node",
			req:  Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 24 * gi, LimitedGPU: 24 * gi, SupportMultiCards: true},
			nodes: []Node{nvidiaNode("nvidia-a",
				Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
				Device{ID: "gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
			)},
		},
		{
			name: "multi card across nodes",
			req:  Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 24 * gi, LimitedGPU: 24 * gi, SupportMultiCards: true, SupportMultiNodes: true},
			nodes: []Node{
				nvidiaNode("nvidia-a", Device{ID: "a-gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice}),
				nvidiaNode("nvidia-b", Device{ID: "b-gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice}),
			},
		},
		{
			name:  "gb10 chip",
			req:   Requirement{Mode: utils.GB10ChipType, RequiredMemory: 24 * gi, LimitedMemory: 48 * gi},
			nodes: []Node{computeNode("spark-ab12", utils.GB10ChipType, 96*gi, SupportTypeMemorySlice)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			availability := listAvailableForLaunch(tc.req, tc.nodes, PressureSnapshot{})
			selections, ok := pickLaunchSelection(tc.req, availability, PressureSnapshot{}, allocationOptions{checkPressure: true})
			if !ok {
				t.Fatalf("expected the auto-picker to find a placement")
			}
			if result := ValidateBindingSelection(tc.req, selections, tc.nodes, PressureSnapshot{}); !result.OK {
				t.Fatalf("install's own pick %#v was rejected by the manual binding validation: %#v", selections, result)
			}
		})
	}
}

// TestNonNvidiaNodeExposesEveryDevice covers nvidia-gb10 chips and discrete
// Intel GPUs, the non-nvidia modes that can put several cards on one node. The
// availability view used to report only the node's first card, which hid the
// rest from the resume picker; now that install picks from the same view, that
// would also have made a busy first card look like a full node.
func TestNonNvidiaNodeExposesEveryDevice(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "ollama", OwnerName: "alice", SelectedGpuType: utils.GB10ChipType}
	req := Requirement{Mode: utils.GB10ChipType, RequiredMemory: 24 * gi, LimitedMemory: 24 * gi}
	nodes := []Node{multiDeviceNode("spark-ab12", utils.GB10ChipType, SupportTypeMemorySlice,
		Device{ID: "spark-0", Memory: 96 * gi, Bindings: []Allocation{{AppName: "other", Owner: "bob", Memory: 90 * gi}}},
		Device{ID: "spark-1", Memory: 96 * gi},
	)}

	result := listAvailableForLaunch(req, nodes, PressureSnapshot{})
	if len(result.Nodes) != 1 || len(result.Nodes[0].Devices) != 2 {
		t.Fatalf("expected both cards to be listed for manual selection, got %#v", result.Nodes)
	}

	picked, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
	if !ok || len(picked) != 1 {
		t.Fatalf("expected a single placement on the free card, got ok=%v picked=%#v", ok, picked)
	}
	if picked[0].DeviceID != "spark-1" || picked[0].Memory != 24*gi {
		t.Fatalf("expected a 24Gi slice of the free second card, got %#v", picked[0])
	}
}

func multiDeviceNode(name, mode, supportType string, devices ...Device) Node {
	node := Node{
		NodeName:       name,
		GPUTypes:       []string{mode},
		memoryCapacity: 128 * gi,
		Devices:        devices,
	}
	for i := range node.Devices {
		node.Devices[i].NodeName = name
		node.Devices[i].Mode = mode
		if node.Devices[i].Health == "" {
			node.Devices[i].Health = deviceHealthYes
		}
		if node.Devices[i].SupportType == "" {
			node.Devices[i].SupportType = supportType
		}
	}
	return node
}

package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

const gi = int64(1024 * 1024 * 1024)

func TestCalculateInstallComputePlanUsesPureInputs(t *testing.T) {
	app := &appcfg.ApplicationConfig{
		AppName: "ollama",
		Resources: []appcfg.ResourceMode{
			resourceMode(utils.NvidiaCardType, "8Gi", "1Gi"),
			resourceMode(utils.AppleMChipType, "", "1Gi"),
			resourceMode(utils.StrixHaloChipType, "", "8Gi"),
		},
	}
	nodes := []Node{
		nvidiaNode("nvidia-a", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
		computeNode("strix-a", utils.StrixHaloChipType, 4*gi, SupportTypeExclusive),
	}

	plan := calculateInstallComputePlan(app, nodes)
	assertModeStatus(t, plan, utils.NvidiaCardType, StatusInstallable)
	assertModeStatus(t, plan, utils.AppleMChipType, StatusNoMatchingNode)
	assertModeStatus(t, plan, utils.StrixHaloChipType, StatusInsufficientResources)
}

func TestInstallComputePlanIgnoresCurrentBindings(t *testing.T) {
	app := &appcfg.ApplicationConfig{
		AppName: "llm",
		Resources: []appcfg.ResourceMode{
			resourceMode(utils.NvidiaCardType, "8Gi", "1Gi"),
		},
	}
	nodes := []Node{
		nvidiaNode("nvidia-a", Device{
			ID:          "gpu0",
			Memory:      16 * gi,
			Health:      deviceHealthYes,
			SupportType: SupportTypeExclusive,
			Bindings:    []Allocation{{AppName: "other", Owner: "other", Memory: 16 * gi}},
		}),
	}

	plan := calculateInstallComputePlan(app, nodes)
	assertModeStatus(t, plan, utils.NvidiaCardType, StatusInstallable)
}

func TestLegacyAppConfigIsMappedToComputeModes(t *testing.T) {
	cpuApp := &appcfg.ApplicationConfig{
		AppName: "legacy-cpu",
		Requirement: appcfg.AppRequirement{
			Memory:        quantity("1Gi"),
			LimitedMemory: quantity("2Gi"),
		},
	}
	plan := calculateInstallComputePlan(cpuApp, []Node{
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemoryShared),
	})
	assertModeStatus(t, plan, utils.CPUType, StatusInstallable)

	gpuApp := &appcfg.ApplicationConfig{
		AppName: "legacy-gpu",
		Requirement: appcfg.AppRequirement{
			GPU:           quantity("2Gi"),
			LimitedGPU:    quantity("16Gi"),
			Memory:        quantity("20Gi"),
			LimitedMemory: quantity("70Gi"),
		},
	}
	mode, ok := gpuApp.SelectedResourceMode()
	if !ok {
		t.Fatalf("expected legacy GPU app to resolve a selected compute mode")
	}
	req := RequirementFromMode(mode)
	if req.Mode != utils.NvidiaCardType || req.RequiredGPU != 2*gi || req.LimitedGPU != 16*gi || req.LimitedMemory != 70*gi {
		t.Fatalf("legacy GPU requirement was not mapped correctly: %#v", req)
	}
}

func TestCPUModeFallsBackToGPUTypeNodesWithoutUsingGPUMemory(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "cpu-app"}
	req := Requirement{
		Mode:           utils.CPUType,
		RequiredMemory: 8 * gi,
		LimitedMemory:  8 * gi,
	}
	nodes := []Node{
		computeNode("cpu-small", utils.CPUType, 4*gi, SupportTypeMemoryShared),
		nvidiaNode("nvidia-a", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
	}

	result := EvaluateInstallMode(req, nodes)
	if result.Status != StatusInstallable {
		t.Fatalf("expected CPU app to fall back to GPU-type node, got %#v", result)
	}
	picked, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
	if ok || len(picked) != 0 {
		t.Fatalf("CPU mode must not create compute allocations, got ok=%v picked=%#v", ok, picked)
	}
}

func TestListAvailableForLaunchFiltersNotMatchAndMarksOperable(t *testing.T) {
	req := Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       24 * gi,
		LimitedGPU:        24 * gi,
		RequiredMemory:    gi,
		LimitedMemory:     gi,
		SupportMultiCards: true,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a",
			Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
			Device{ID: "gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive},
		),
		nvidiaNode("nvidia-b", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemoryShared),
	}

	result := listAvailableForLaunch(req, nodes, PressureSnapshot{})
	if len(result.Nodes) != 2 {
		t.Fatalf("expected not_match nodes to be hidden from launch result, got %d", len(result.Nodes))
	}
	if !result.Schedulable {
		t.Fatalf("expected launch result to be schedulable")
	}
	if result.Scope != AvailabilityScopeSingleNode {
		t.Fatalf("expected single-node scope, got %s", result.Scope)
	}
	if result.Nodes[0].Status != NodeStatusAvailable || !result.Nodes[0].Devices[0].Operable || !result.Nodes[0].Devices[1].Operable {
		t.Fatalf("expected available node devices to be operable: %#v", result.Nodes[0])
	}
	if result.Nodes[1].Status != NodeStatusNotEnough || result.Nodes[1].Devices[0].Operable {
		t.Fatalf("expected not_enough node to be visible but not operable: %#v", result.Nodes[1])
	}
}

func TestListAvailableForLaunchReportsNoMatchingNode(t *testing.T) {
	req := Requirement{Mode: utils.NvidiaCardType, RequiredGPU: gi, LimitedGPU: gi}
	result := listAvailableForLaunch(req, []Node{
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemoryShared),
	}, PressureSnapshot{})

	if result.Schedulable || result.Reason != "no-matching-node" || len(result.Nodes) != 0 {
		t.Fatalf("expected no matching node result, got %#v", result)
	}
}

func TestValidateBindingSelectionTopologyAndMemorySlice(t *testing.T) {
	req := Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       12 * gi,
		LimitedGPU:        12 * gi,
		RequiredMemory:    gi,
		LimitedMemory:     gi,
		SupportMultiCards: true,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a",
			Device{ID: "gpu0", Memory: 8 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice},
			Device{ID: "gpu1", Memory: 8 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice},
		),
		nvidiaNode("nvidia-b", Device{ID: "gpu0", Memory: 8 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice}),
	}

	result := ValidateBindingSelection(req, []BindingSelection{
		{NodeName: "nvidia-a", DeviceID: "gpu0", Memory: 8 * gi},
		{NodeName: "nvidia-b", DeviceID: "gpu0", Memory: 8 * gi},
	}, nodes, PressureSnapshot{})
	if result.OK || result.Code != "multi-node-not-supported" {
		t.Fatalf("expected multi-node rejection, got %#v", result)
	}

	result = ValidateBindingSelection(req, []BindingSelection{
		{NodeName: "nvidia-a", DeviceID: "gpu0"},
		{NodeName: "nvidia-a", DeviceID: "gpu1", Memory: 8 * gi},
	}, nodes, PressureSnapshot{})
	if result.OK || result.Code != "memory-required:gpu0" {
		t.Fatalf("expected memory slice amount rejection, got %#v", result)
	}

	result = ValidateBindingSelection(req, []BindingSelection{
		{NodeName: "nvidia-a", DeviceID: "gpu0", Memory: 6 * gi},
		{NodeName: "nvidia-a", DeviceID: "gpu1", Memory: 6 * gi},
	}, nodes, PressureSnapshot{})
	if !result.OK {
		t.Fatalf("expected valid aggregate selection, got %#v", result)
	}
}

func TestAvailabilityCrossNodeSumsRawAvailableDespitePerNodePressure(t *testing.T) {
	req := Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       24 * gi,
		LimitedGPU:        24 * gi,
		RequiredMemory:    2 * gi,
		LimitedMemory:     2 * gi,
		SupportMultiCards: true,
		SupportMultiNodes: true,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
		nvidiaNode("nvidia-b", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
	}
	// Pressure on nvidia-a — its single card's FitLevel would be empty, but the
	// cross-node summary should still count the device's raw available memory.
	pressure := PressureSnapshot{
		Threshold: 0.9,
		UsageByNode: map[string]prometheus.NodeResourceUsage{
			"nvidia-a": {MemoryCapacity: 4 * gi, MemoryAvailable: 0},
		},
	}

	result := listAvailableForLaunch(req, nodes, pressure)
	if !result.Schedulable {
		t.Fatalf("expected cross-node cluster total (32Gi) to satisfy 24Gi request, got %#v", result)
	}
	if result.Scope != AvailabilityScopeCrossNode {
		t.Fatalf("expected cross-node scope, got %s", result.Scope)
	}
}

func TestAvailabilityNonNvidiaNodeBecomesNotAvailableUnderPressure(t *testing.T) {
	req := Requirement{
		Mode:           utils.StrixHaloChipType,
		RequiredMemory: 4 * gi,
		LimitedMemory:  4 * gi,
	}
	nodes := []Node{computeNode("strix-a", utils.StrixHaloChipType, 32*gi, SupportTypeExclusive)}

	healthyResult := listAvailableForLaunch(req, nodes, PressureSnapshot{})
	if len(healthyResult.Nodes) != 1 || healthyResult.Nodes[0].Status != NodeStatusAvailable {
		t.Fatalf("expected strix node to be available without pressure, got %#v", healthyResult)
	}

	pressure := PressureSnapshot{
		Threshold: 0.9,
		UsageByNode: map[string]prometheus.NodeResourceUsage{
			"strix-a": {MemoryCapacity: 4 * gi, MemoryAvailable: 0},
		},
	}
	pressuredResult := listAvailableForLaunch(req, nodes, pressure)
	if len(pressuredResult.Nodes) != 1 || pressuredResult.Nodes[0].Status != NodeStatusNotAvailable {
		t.Fatalf("expected strix node to be not_available under memory pressure, got %#v", pressuredResult)
	}
	if pressuredResult.Schedulable {
		t.Fatalf("expected scheduling to fail under pressure, got %#v", pressuredResult)
	}
}

func TestValidateBindingSelectionNonNvidiaRequiresSingleSelection(t *testing.T) {
	req := Requirement{
		Mode:           utils.StrixHaloChipType,
		RequiredMemory: 4 * gi,
		LimitedMemory:  4 * gi,
	}
	nodes := []Node{
		computeNode("strix-a", utils.StrixHaloChipType, 32*gi, SupportTypeExclusive),
		computeNode("strix-b", utils.StrixHaloChipType, 32*gi, SupportTypeExclusive),
	}
	result := ValidateBindingSelection(req, []BindingSelection{
		{NodeName: "strix-a", DeviceID: "strix-a-device"},
		{NodeName: "strix-b", DeviceID: "strix-b-device"},
	}, nodes, PressureSnapshot{})
	if result.OK || result.Code != "non-nvidia-must-single-selection" {
		t.Fatalf("expected non-nvidia multi selection to be rejected, got %#v", result)
	}
}

func TestPickAggregateAllocationsAcrossNodes(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	req := Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       24 * gi,
		LimitedGPU:        24 * gi,
		SupportMultiCards: true,
		SupportMultiNodes: true,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
		nvidiaNode("nvidia-b", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
	}
	picked, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
	if !ok || len(picked) != 2 {
		t.Fatalf("expected 2 cross-node allocations to satisfy 24Gi, got ok=%v picked=%#v", ok, picked)
	}
	nodesPicked := map[string]struct{}{}
	for _, alloc := range picked {
		nodesPicked[alloc.NodeName] = struct{}{}
	}
	if len(nodesPicked) != 2 {
		t.Fatalf("expected allocations to span both nodes, got %#v", nodesPicked)
	}
}

// TestLegacyComputeMode covers the synthesis of the single ResourceMode for
// a legacy manifest (no spec.resources matrix). The contract is:
//
//  1. An explicit SelectedGpuType wins outright — used during install once
//     the auto-selector (or the user) has picked a mode and the chart is
//     reloaded with that pick.
//  2. With SelectedGpuType empty, a non-zero Requirement.GPU collapses to
//     nvidia (the only legacy-supported GPU type); zero/unset collapses to
//     cpu. This is the "no choice yet" state hit by the install pre-check
//     and the first GetAppConfig call inside the install handler.
func TestLegacyComputeMode(t *testing.T) {
	cases := []struct {
		name string
		cfg  *appcfg.ApplicationConfig
		want string
	}{
		{
			name: "no selection, no gpu requirement -> cpu",
			cfg:  &appcfg.ApplicationConfig{AppName: "no-gpu"},
			want: utils.CPUType,
		},
		{
			name: "no selection, non-zero requiredGpu -> nvidia",
			cfg: &appcfg.ApplicationConfig{
				AppName: "legacy-gpu",
				Requirement: appcfg.AppRequirement{
					GPU: quantity("4Gi"),
				},
			},
			want: utils.NvidiaCardType,
		},
		{
			name: "no selection, zero requiredGpu -> cpu",
			cfg: &appcfg.ApplicationConfig{
				AppName: "legacy-zero-gpu",
				Requirement: appcfg.AppRequirement{
					GPU: quantity("0"),
				},
			},
			want: utils.CPUType,
		},
		{
			name: "explicit SelectedGpuType wins over requirement-derived fallback",
			cfg: &appcfg.ApplicationConfig{
				AppName:         "legacy-gpu-explicit",
				SelectedGpuType: utils.NvidiaCardType,
				Requirement: appcfg.AppRequirement{
					GPU: quantity("4Gi"),
				},
			},
			want: utils.NvidiaCardType,
		},
		{
			name: "explicit SelectedGpuType wins even when requirement says cpu",
			cfg: &appcfg.ApplicationConfig{
				AppName:         "legacy-pinned-on-strix",
				SelectedGpuType: utils.StrixHaloChipType,
			},
			want: utils.StrixHaloChipType,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			modes := tt.cfg.ComputeResourceModes()
			if len(modes) != 1 {
				t.Fatalf("expected a single synthesized mode for legacy app, got %#v", modes)
			}
			if got := modes[0].Mode; got != tt.want {
				t.Fatalf("expected mode %s, got %s", tt.want, got)
			}
		})
	}
}

func resourceMode(mode, requiredGPU, requiredMemory string) appcfg.ResourceMode {
	return appcfg.ResourceMode{
		Mode: mode,
		ResourceRequirement: appcfg.ResourceRequirement{
			RequiredGPU:    requiredGPU,
			LimitedGPU:     requiredGPU,
			RequiredMemory: requiredMemory,
			LimitedMemory:  requiredMemory,
		},
	}
}

func nvidiaNode(name string, devices ...Device) Node {
	node := Node{
		NodeName:       name,
		GPUType:        utils.NvidiaCardType,
		memoryCapacity: 64 * gi,
		Devices:        devices,
	}
	for i := range node.Devices {
		node.Devices[i].NodeName = name
	}
	return node
}

func computeNode(name, gpuType string, memory int64, supportType string) Node {
	return Node{
		NodeName:       name,
		GPUType:        gpuType,
		memoryCapacity: 64 * gi,
		Devices: []Device{{
			ID:          name + "-device",
			NodeName:    name,
			Memory:      memory,
			Health:      deviceHealthYes,
			SupportType: supportType,
		}},
	}
}

func assertModeStatus(t *testing.T, plan []ModePlanResult, mode, status string) {
	t.Helper()
	for _, item := range plan {
		if item.ComputeType == mode {
			if item.Status != status {
				t.Fatalf("expected %s status %s, got %s", mode, status, item.Status)
			}
			return
		}
	}
	t.Fatalf("mode %s not found in plan %#v", mode, plan)
}

func quantity(value string) *resource.Quantity {
	q := resource.MustParse(value)
	return &q
}

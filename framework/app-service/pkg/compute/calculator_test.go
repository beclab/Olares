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
		Accelerator: []appcfg.ResourceMode{
			resourceMode(utils.NvidiaCardType, "8Gi", "1Gi"),
			resourceMode(utils.AppleMChipType, "", "1Gi"),
			resourceMode(utils.AMDType, "", "8Gi"),
		},
	}
	nodes := []Node{
		nvidiaNode("nvidia-a", Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}),
		computeNode("amd-a", utils.AMDType, 4*gi, SupportTypeExclusive),
	}

	plan := calculateInstallComputePlan(app, nodes)
	assertModeStatus(t, plan, utils.NvidiaCardType, StatusInstallable)
	assertModeStatus(t, plan, utils.AppleMChipType, StatusNoMatchingNode)
	assertModeStatus(t, plan, utils.AMDType, StatusInsufficientResources)
}

func TestInstallComputePlanIgnoresCurrentBindings(t *testing.T) {
	app := &appcfg.ApplicationConfig{
		AppName: "llm",
		Accelerator: []appcfg.ResourceMode{
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
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemorySlice),
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
		computeNode("cpu-small", utils.CPUType, 4*gi, SupportTypeMemorySlice),
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
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemorySlice),
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
		computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemorySlice),
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

// TestAllocationsFromResolvedSelectionMultiCardWholeCard is a regression test
// for the multi-card binding bug: when the frontend submitted two whole-card
// (Exclusive / TimeSlice) selections, allocationsFromResolvedSelection folded
// them into a single RequiredGPU budget and dropped every card past the one
// that first covered the budget, so only one HAMI GPUBinding was created for a
// two-card request. Every selected whole card must yield its own allocation.
func TestAllocationsFromResolvedSelectionMultiCardWholeCard(t *testing.T) {
	cases := []struct {
		name        string
		supportType string
	}{
		{name: "exclusive", supportType: SupportTypeExclusive},
		{name: "timeslice", supportType: SupportTypeTimeSlice},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := &appcfg.ApplicationConfig{AppName: "trainer", OwnerName: "alice"}
			req := Requirement{
				Mode:              utils.NvidiaCardType,
				RequiredGPU:       12 * gi,
				LimitedGPU:        12 * gi,
				SupportMultiCards: true,
			}
			nodes := []Node{
				nvidiaNode("nvidia-a",
					Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: tc.supportType},
					Device{ID: "gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: tc.supportType},
				),
			}
			resolved, err := resolveSelection([]BindingSelection{
				{NodeName: "nvidia-a", DeviceID: "gpu0"},
				{NodeName: "nvidia-a", DeviceID: "gpu1"},
			}, nodes)
			if err != nil {
				t.Fatalf("resolveSelection failed: %v", err)
			}
			allocations := allocationsFromResolvedSelection(app, req, resolved)
			if len(allocations) != 2 {
				t.Fatalf("expected one allocation per selected card, got %d: %#v", len(allocations), allocations)
			}
			devices := map[string]struct{}{}
			for _, a := range allocations {
				devices[a.DeviceID] = struct{}{}
				if a.Memory != 0 {
					t.Fatalf("whole-card allocation should record Memory=0, got %#v", a)
				}
				if a.AppName != "trainer" || a.Owner != "alice" || a.NodeName != "nvidia-a" {
					t.Fatalf("unexpected allocation identity: %#v", a)
				}
			}
			for _, id := range []string{"gpu0", "gpu1"} {
				if _, ok := devices[id]; !ok {
					t.Fatalf("expected an allocation for %s, got %#v", id, allocations)
				}
			}
		})
	}
}

// TestAllocationsFromResolvedSelectionMultiCardMemorySlice pins the unchanged
// memory-slice path: each selected card keeps its own explicit slice amount.
func TestAllocationsFromResolvedSelectionMultiCardMemorySlice(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "trainer", OwnerName: "alice"}
	req := Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       12 * gi,
		LimitedGPU:        12 * gi,
		SupportMultiCards: true,
	}
	nodes := []Node{
		nvidiaNode("nvidia-a",
			Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice},
			Device{ID: "gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice},
		),
	}
	resolved, err := resolveSelection([]BindingSelection{
		{NodeName: "nvidia-a", DeviceID: "gpu0", Memory: 6 * gi},
		{NodeName: "nvidia-a", DeviceID: "gpu1", Memory: 6 * gi},
	}, nodes)
	if err != nil {
		t.Fatalf("resolveSelection failed: %v", err)
	}
	allocations := allocationsFromResolvedSelection(app, req, resolved)
	if len(allocations) != 2 {
		t.Fatalf("expected one allocation per memory-slice card, got %d: %#v", len(allocations), allocations)
	}
	for _, a := range allocations {
		if a.Memory != 6*gi {
			t.Fatalf("expected each memory-slice allocation to keep its 6Gi slice, got %#v", a)
		}
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
		Mode:           utils.AMDType,
		RequiredMemory: 4 * gi,
		LimitedMemory:  4 * gi,
	}
	nodes := []Node{computeNode("strix-a", utils.AMDType, 32*gi, SupportTypeExclusive)}

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
		Mode:           utils.AMDType,
		RequiredMemory: 4 * gi,
		LimitedMemory:  4 * gi,
	}
	nodes := []Node{
		computeNode("strix-a", utils.AMDType, 32*gi, SupportTypeExclusive),
		computeNode("strix-b", utils.AMDType, 32*gi, SupportTypeExclusive),
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

// TestPickSingleAllocationGB10MemorySlice guards the GB10 single-card
// allocation path: a GB10 node's device is decoded with the MemorySlice
// support type by default (shareModeToSupportType), so PickAllocations must
// offer MemorySlice in its support-type order. GB10 shares the non-nvidia
// [Exclusive, MemorySlice] order; dropping MemorySlice from it would filter
// every GB10 candidate out and surface as "no available compute resource for
// type nvidia-gb10" at install time.
func TestPickSingleAllocationGB10MemorySlice(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "ollama", OwnerName: "alice"}
	req := Requirement{
		Mode:           utils.GB10ChipType,
		RequiredMemory: 24 * gi,
		LimitedMemory:  48 * gi,
	}
	nodes := []Node{
		computeNode("spark-ab12", utils.GB10ChipType, 96*gi, SupportTypeMemorySlice),
	}

	picked, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
	if !ok || len(picked) != 1 {
		t.Fatalf("expected a single GB10 allocation, got ok=%v picked=%#v", ok, picked)
	}
	if picked[0].NodeName != "spark-ab12" || picked[0].Mode != utils.GB10ChipType {
		t.Fatalf("unexpected GB10 allocation: %#v", picked[0])
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
				SelectedGpuType: utils.AMDType,
			},
			want: utils.AMDType,
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
		GPUTypes:       []string{utils.NvidiaCardType},
		memoryCapacity: 64 * gi,
		Devices:        devices,
	}
	for i := range node.Devices {
		node.Devices[i].NodeName = name
		if node.Devices[i].Mode == "" {
			node.Devices[i].Mode = utils.NvidiaCardType
		}
	}
	return node
}

func computeNode(name, gpuType string, memory int64, supportType string) Node {
	return Node{
		NodeName:       name,
		GPUTypes:       []string{gpuType},
		memoryCapacity: 64 * gi,
		Devices: []Device{{
			ID:          name + "-device",
			NodeName:    name,
			Mode:        gpuType,
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

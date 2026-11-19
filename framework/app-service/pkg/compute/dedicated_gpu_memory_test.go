package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

const gib = int64(1 << 30)

// dedicatedModeRequirement is the requirement a manifest produces for a
// discrete card: a GPU-memory quota well above the pod memory request, so any
// check that reads the wrong one is off by a visible margin.
func dedicatedModeRequirement(mode string) Requirement {
	return RequirementFromMode(appcfg.ResourceMode{
		Mode: mode,
		ResourceRequirement: appcfg.ResourceRequirement{
			RequiredCPU:    "1",
			LimitedCPU:     "2",
			RequiredMemory: "2Gi",
			LimitedMemory:  "4Gi",
			RequiredDisk:   "1Gi",
			LimitedDisk:    "2Gi",
			RequiredGPU:    "16Gi",
			LimitedGPU:     "20Gi",
		},
	})
}

func exclusiveDevice(mode string, memory int64, bindings ...Allocation) Device {
	return Device{
		ID:          "card0",
		NodeName:    "n1",
		Mode:        mode,
		Memory:      memory,
		Health:      deviceHealthYes,
		SupportType: SupportTypeExclusive,
		Bindings:    bindings,
	}
}

func singleDeviceNode(device Device) Node {
	return Node{
		NodeName: device.NodeName,
		GPUTypes: []string{device.Mode},
		Devices:  []Device{device},
	}
}

// The discrete-GPU modes own VRAM separate from system RAM, so their fit target
// is the manifest's GPU-memory quota. Only the unified-memory modes may fall
// back to the pod memory request. Before this was fixed every non-nvidia mode
// took the fallback, so a discrete Intel/AMD card was sized against pod memory
// and the declared VRAM was parsed and then ignored.
func TestFitTargetPerMode(t *testing.T) {
	tests := []struct {
		name             string
		mode             string
		wantRequired     int64
		wantLimit        int64
		wantDedicatedGPU bool
	}{
		{"nvidia uses gpu quota", utils.NvidiaCardType, 16 * gib, 20 * gib, true},
		{"amd discrete uses gpu quota", utils.AMDGPUType, 16 * gib, 20 * gib, true},
		{"intel discrete uses gpu quota", utils.IntelGPUType, 16 * gib, 20 * gib, true},
		{"amd integrated uses pod memory", utils.AMDType, 2 * gib, 4 * gib, false},
		{"intel integrated uses pod memory", utils.IntelType, 2 * gib, 4 * gib, false},
		{"gb10 uses pod memory", utils.GB10ChipType, 2 * gib, 4 * gib, false},
		{"apple m uses pod memory", utils.AppleMChipType, 2 * gib, 4 * gib, false},
		{"moore soc uses pod memory", utils.MooreSocType, 2 * gib, 4 * gib, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dedicatedModeRequirement(tt.mode)
			if got := utils.HasDedicatedGPUMemory(tt.mode); got != tt.wantDedicatedGPU {
				t.Fatalf("HasDedicatedGPUMemory(%s) = %v, want %v", tt.mode, got, tt.wantDedicatedGPU)
			}
			if got := targetForMode(req, FitLevelRequired); got != tt.wantRequired {
				t.Errorf("targetForMode(%s, required) = %d, want %d", tt.mode, got, tt.wantRequired)
			}
			if got := targetForMode(req, FitLevelLimit); got != tt.wantLimit {
				t.Errorf("targetForMode(%s, limit) = %d, want %d", tt.mode, got, tt.wantLimit)
			}
			if got := requiredTargetForMode(req); got != tt.wantRequired {
				t.Errorf("requiredTargetForMode(%s) = %d, want %d", tt.mode, got, tt.wantRequired)
			}
		})
	}
}

// Install-time capacity has to reject a discrete card smaller than the app's
// declared VRAM. It used to accept one whenever the card cleared the much
// smaller pod memory request, so an app needing 16Gi was reported installable
// against an 8Gi card and only failed later at runtime.
func TestInstallCapacityFitsSizesDiscreteCardsByGPUQuota(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		cardMemory int64
		want       bool
	}{
		{"amd discrete card below gpu quota", utils.AMDGPUType, 8 * gib, false},
		{"amd discrete card above gpu quota", utils.AMDGPUType, 24 * gib, true},
		{"intel discrete card below gpu quota", utils.IntelGPUType, 8 * gib, false},
		{"intel discrete card above gpu quota", utils.IntelGPUType, 24 * gib, true},
		{"nvidia card below gpu quota", utils.NvidiaCardType, 8 * gib, false},
		{"nvidia card above gpu quota", utils.NvidiaCardType, 24 * gib, true},
		// The unified-memory modes have no VRAM of their own, so they keep
		// being sized against the 2Gi pod memory request.
		{"intel integrated above pod memory", utils.IntelType, 8 * gib, true},
		{"intel integrated below pod memory", utils.IntelType, gib, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dedicatedModeRequirement(tt.mode)
			node := singleDeviceNode(exclusiveDevice(tt.mode, tt.cardMemory))
			if got := installCapacityFits(req, []Node{node}); got != tt.want {
				t.Fatalf("installCapacityFits(%s, card=%dGi) = %v, want %v",
					tt.mode, tt.cardMemory/gib, got, tt.want)
			}
		})
	}
}

func TestDedicatedVRAMMultiCardPlacement(t *testing.T) {
	for _, mode := range []string{utils.NvidiaCardType, utils.IntelGPUType, utils.AMDGPUType} {
		t.Run(mode, func(t *testing.T) {
			for _, tc := range []struct {
				name       string
				multiCards bool
				multiNodes bool
				crossNode  bool
				cardMemory int64
				wantScope  string
				wantOK     bool
			}{
				{"single card only", false, false, false, 16 * gib, AvailabilityScopeCard, false},
				{"multiple cards on one node", true, false, false, 16 * gib, AvailabilityScopeSingleNode, true},
				{"cross node disabled", true, false, true, 16 * gib, AvailabilityScopeSingleNode, false},
				{"multiple nodes imply multiple cards", false, true, true, 16 * gib, AvailabilityScopeCrossNode, true},
				{"aggregate memory insufficient", true, false, false, 8 * gib, AvailabilityScopeSingleNode, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					req := RequirementFromMode(appcfg.ResourceMode{
						Mode:              mode,
						SupportMultiCards: tc.multiCards,
						SupportMultiNodes: tc.multiNodes,
						ResourceRequirement: appcfg.ResourceRequirement{
							RequiredGPU: "24Gi", LimitedGPU: "24Gi", RequiredMemory: "2Gi", LimitedMemory: "2Gi",
						},
					})
					first := exclusiveDevice(mode, tc.cardMemory)
					second := exclusiveDevice(mode, tc.cardMemory)
					second.ID = "card1"
					nodes := []Node{singleDeviceNode(first)}
					if tc.crossNode {
						second.NodeName = "n2"
						nodes = append(nodes, singleDeviceNode(second))
					} else {
						nodes[0].Devices = append(nodes[0].Devices, second)
					}
					if got := installCapacityFits(req, nodes); got != tc.wantOK {
						t.Fatalf("install capacity fits = %v, want %v", got, tc.wantOK)
					}
					view := listAvailableForLaunch(req, nodes, PressureSnapshot{})
					if view.Scope != tc.wantScope || view.Schedulable != tc.wantOK {
						t.Fatalf("availability = %#v, want scope=%s schedulable=%v", view, tc.wantScope, tc.wantOK)
					}
					app := &appcfg.ApplicationConfig{AppName: "trainer", OwnerName: "alice"}
					allocations, ok := PickAllocations(app, req, nodes, PressureSnapshot{})
					if ok != tc.wantOK {
						t.Fatalf("auto-pick = %v, want %v", ok, tc.wantOK)
					}
					if ok && (len(allocations) != 2 || allocations[0].Memory != 0 || allocations[1].Memory != 0) {
						t.Fatalf("expected two whole-card allocations, got %#v", allocations)
					}
					selections := []BindingSelection{
						{NodeName: first.NodeName, DeviceID: first.ID},
						{NodeName: second.NodeName, DeviceID: second.ID},
					}
					bound, validation := bindAllocations(app, req, selections, nodes, PressureSnapshot{})
					if validation.OK != tc.wantOK || (validation.OK && len(bound) != 2) {
						t.Fatalf("manual binding = %#v, allocations=%#v, want valid=%v", validation, bound, tc.wantOK)
					}
				})
			}
		})
	}
}

// The node/device availability view feeds both the resume picker and install's
// auto-pick, so it has to reach the same verdict as the capacity check: a
// discrete card too small for the declared VRAM is not-enough, not available.
func TestAvailabilityStatusForDiscreteCards(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		cardMemory int64
		wantStatus string
	}{
		{"amd discrete card below gpu quota", utils.AMDGPUType, 8 * gib, NodeStatusNotEnough},
		{"amd discrete card above gpu quota", utils.AMDGPUType, 24 * gib, NodeStatusAvailable},
		{"intel discrete card below gpu quota", utils.IntelGPUType, 8 * gib, NodeStatusNotEnough},
		{"intel discrete card above gpu quota", utils.IntelGPUType, 24 * gib, NodeStatusAvailable},
		{"intel integrated sized by pod memory", utils.IntelType, 8 * gib, NodeStatusAvailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dedicatedModeRequirement(tt.mode)
			node := singleDeviceNode(exclusiveDevice(tt.mode, tt.cardMemory))
			result := listAvailableForLaunch(req, []Node{node}, PressureSnapshot{})
			if len(result.Nodes) != 1 {
				t.Fatalf("expected 1 classified node, got %d", len(result.Nodes))
			}
			if result.Nodes[0].Status != tt.wantStatus {
				t.Fatalf("node status = %s, want %s", result.Nodes[0].Status, tt.wantStatus)
			}
			if wantSchedulable := tt.wantStatus == NodeStatusAvailable; result.Schedulable != wantSchedulable {
				t.Fatalf("schedulable = %v, want %v (reason %q)",
					result.Schedulable, wantSchedulable, result.Reason)
			}
		})
	}
}

// A manually submitted binding goes through the same VRAM judgement, so a
// discrete card that cannot hold the declared GPU memory has to be rejected
// with the VRAM code rather than being waved through on pod memory.
func TestValidateBindingSelectionForDiscreteCards(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		cardMemory int64
		wantOK     bool
		wantCode   string
	}{
		{"amd discrete card below gpu quota", utils.AMDGPUType, 8 * gib, false, "device-vram-insufficient"},
		{"amd discrete card above gpu quota", utils.AMDGPUType, 24 * gib, true, BindingValidationReasonValid},
		{"intel discrete card below gpu quota", utils.IntelGPUType, 8 * gib, false, "device-vram-insufficient"},
		{"intel discrete card above gpu quota", utils.IntelGPUType, 24 * gib, true, BindingValidationReasonValid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dedicatedModeRequirement(tt.mode)
			node := singleDeviceNode(exclusiveDevice(tt.mode, tt.cardMemory))
			selections := []BindingSelection{{NodeName: node.NodeName, DeviceID: node.Devices[0].ID}}
			result := ValidateBindingSelection(req, selections, []Node{node}, PressureSnapshot{})
			if result.OK != tt.wantOK || result.Code != tt.wantCode {
				t.Fatalf("validation = (ok=%v, code=%q), want (ok=%v, code=%q)",
					result.OK, result.Code, tt.wantOK, tt.wantCode)
			}
		})
	}
}

// A discrete-GPU app can legitimately declare no GPU-memory quota: the field is
// only mandatory for modes listed under spec.resources, and an auto-resource
// sentinel resolves to zero when the chart carries no nvidia.com/gpumem. Its
// fit target is then zero, which must not let it onto a card another app
// already holds exclusively — the capacity comparison alone would allow that,
// since an exhausted card reports zero available.
func TestValidateBindingSelectionRejectsBoundExclusiveCardWithoutGPUQuota(t *testing.T) {
	for _, mode := range []string{utils.AMDGPUType, utils.IntelGPUType, utils.NvidiaCardType, utils.IntelType} {
		t.Run(mode, func(t *testing.T) {
			req := RequirementFromMode(appcfg.ResourceMode{
				Mode: mode,
				ResourceRequirement: appcfg.ResourceRequirement{
					RequiredCPU: "1", RequiredDisk: "1Gi",
				},
			})
			device := exclusiveDevice(mode, 24*gib, Allocation{
				AppName: "incumbent", Mode: mode, NodeName: "n1", DeviceID: "card0",
			})
			node := singleDeviceNode(device)
			selections := []BindingSelection{{NodeName: node.NodeName, DeviceID: device.ID}}
			result := ValidateBindingSelection(req, selections, []Node{node}, PressureSnapshot{})
			if result.OK || result.Code != "exclusive-already-bound:"+device.ID {
				t.Fatalf("validation = (ok=%v, code=%q), want rejection with exclusive-already-bound",
					result.OK, result.Code)
			}
		})
	}
}

// The allocation records what the app took from the card, which the compute
// listing shows back to the user. For a discrete card that is the VRAM quota;
// recording the pod memory request instead reported a host-RAM number as the
// card's allocated video memory.
func TestAllocationRecordsGPUQuotaForDiscreteCards(t *testing.T) {
	for _, mode := range []string{utils.AMDGPUType, utils.IntelGPUType} {
		t.Run(mode, func(t *testing.T) {
			req := dedicatedModeRequirement(mode)
			node := singleDeviceNode(exclusiveDevice(mode, 24*gib))
			appConfig := &appcfg.ApplicationConfig{AppName: "app", OwnerName: "owner"}
			allocations, ok := PickAllocations(appConfig, req, []Node{node}, PressureSnapshot{})
			if !ok {
				t.Fatal("expected the app to be placed on a card larger than its quota")
			}
			if len(allocations) != 1 {
				t.Fatalf("expected 1 allocation, got %d", len(allocations))
			}
			// AvailableSupportTypes makes every discrete Intel/AMD card
			// Exclusive-only, so the app holds the card outright and the
			// allocation records Memory=0 ("whole card") the way an Exclusive
			// nvidia card does. Writing the quota here instead would claim a
			// partition of a card that cannot be partitioned, and would suggest
			// the remaining VRAM is up for grabs when exclusive-already-bound
			// refuses every further binding.
			if allocations[0].Memory != 0 {
				t.Fatalf("allocation memory = %d, want 0 (whole card) for an exclusive %s card",
					allocations[0].Memory, mode)
			}
		})
	}
}

// A discrete Intel/AMD app that declares no GPU-memory quota still has to get a
// binding out of resume. Its fit target is zero — requiredGPUMemory is only
// mandatory for modes listed under spec.resources, and the auto-resource
// sentinel resolves to zero whenever the rendered chart carries neither
// nvidia.com/gpumem nor the GPU-memory pod annotation — and the frontend sends
// no per-card memory for a whole-card mode. That combination used to fall
// through every branch of allocationsFromResolvedSelection and drop the card,
// which reached the user as "empty-compute-binding" on an idle, healthy card.
func TestResumeBindsDiscreteCardWithoutGPUQuota(t *testing.T) {
	for _, mode := range []string{utils.AMDGPUType, utils.IntelGPUType} {
		t.Run(mode, func(t *testing.T) {
			req := RequirementFromMode(appcfg.ResourceMode{
				Mode: mode,
				ResourceRequirement: appcfg.ResourceRequirement{
					RequiredCPU: "1", RequiredMemory: "2Gi", RequiredDisk: "1Gi",
				},
			})
			if req.RequiredGPU != 0 {
				t.Fatalf("precondition: expected a zero gpu quota, got %d", req.RequiredGPU)
			}
			device := exclusiveDevice(mode, 24*gib)
			node := singleDeviceNode(device)
			appConfig := &appcfg.ApplicationConfig{AppName: "app", OwnerName: "owner"}

			// Memory is left at zero exactly as the resume picker submits it.
			selections := []BindingSelection{{NodeName: node.NodeName, DeviceID: device.ID}}
			allocations, validation := bindAllocations(appConfig, req, selections, []Node{node}, PressureSnapshot{})
			if !validation.OK {
				t.Fatalf("resume binding refused: %s", validation.Code)
			}
			if len(allocations) != 1 {
				t.Fatalf("expected 1 allocation, got %d", len(allocations))
			}
			if allocations[0].DeviceID != device.ID {
				t.Fatalf("allocation bound to %q, want %q", allocations[0].DeviceID, device.ID)
			}
		})
	}
}

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

package compute

import (
	"reflect"
	"slices"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var allTestModes = []string{
	utils.CPUType,
	utils.NvidiaCardType,
	utils.GB10ChipType,
	utils.AppleMChipType,
	utils.IntelType,
	utils.AMDType,
	utils.IntelGPUType,
	utils.AMDGPUType,
	utils.MooreSocType,
	"some-future-mode",
}

// TestModeCapabilities pins the scheduling and injection capabilities of each mode.
func TestModeCapabilities(t *testing.T) {
	nvidiaRegister := DeviceRegister{Annotation: constants.NodeNvidiaRegistryKey, Decoder: registerHAMI, Handshake: constants.NodeHandshakeKey}
	tests := []struct {
		mode string
		want ModeCapability
	}{
		{utils.CPUType, ModeCapability{
			SupportTypes: []string{SupportTypeMemorySlice},
		}},
		{utils.NvidiaCardType, ModeCapability{
			SupportTypes:     []string{SupportTypeTimeSlice, SupportTypeMemorySlice, SupportTypeExclusive},
			HAMI:             true,
			ExtendedResource: constants.NvidiaGPU,
			MemoryResource:   constants.NvidiaGPUMem,
			DedicatedVRAM:    true,
			RuntimeClass:     "nvidia",
			MultiCard:        true,
			Register:         nvidiaRegister,
		}},
		{utils.GB10ChipType, ModeCapability{
			SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
			HAMI:             true,
			ExtendedResource: constants.NvidiaGPU,
			MemoryResource:   constants.NvidiaGPUMem,
			RuntimeClass:     "nvidia",
			Register:         nvidiaRegister,
		}},
		{utils.AppleMChipType, ModeCapability{
			SupportTypes: []string{SupportTypeExclusive},
		}},
		{utils.IntelType, ModeCapability{
			SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
			ExtendedResource: constants.IntelIGPU,
		}},
		{utils.AMDType, ModeCapability{
			SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
			ExtendedResource: constants.AMDGPU,
		}},
		{utils.IntelGPUType, ModeCapability{
			SupportTypes:       []string{SupportTypeExclusive, SupportTypeMemorySlice},
			HAMI:               true,
			ExtendedResource:   constants.IntelVGPU,
			MemoryResource:     constants.IntelVGPUMem,
			DedicatedVRAM:      true,
			MultiCard:          true,
			MemorySliceReserve: 256 * mib,
			Register: DeviceRegister{
				Annotation: constants.NodeIntelHAMIRegisterKey,
				Decoder:    registerHAMI,
				Handshake:  constants.NodeIntelHandshakeKey,
				Cards:      constants.NodeIntelCardsKey,
			},
		}},
		{utils.AMDGPUType, ModeCapability{
			SupportTypes:     []string{SupportTypeExclusive},
			ExtendedResource: constants.AMDGPU,
			DedicatedVRAM:    true,
			MultiCard:        true,
			Register:         DeviceRegister{Annotation: constants.NodeAmdRegisterKey, Decoder: registerTuple},
		}},
		{utils.MooreSocType, ModeCapability{
			SupportTypes: []string{SupportTypeMemorySlice, SupportTypeExclusive},
		}},
		{"some-future-mode", ModeCapability{
			SupportTypes: []string{SupportTypeExclusive},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got := CapabilityFor(tt.mode)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CapabilityFor(%q) = %+v, want %+v", tt.mode, got, tt.want)
			}
			if got.HAMI != IsHAMIMode(tt.mode) {
				t.Fatalf("IsHAMIMode(%q) disagrees with the table", tt.mode)
			}
			if (got.ExtendedResource != "") != UsesGPUExtendedResource(tt.mode) {
				t.Fatalf("UsesGPUExtendedResource(%q) disagrees with the table", tt.mode)
			}
			if !reflect.DeepEqual(AvailableSupportTypes(tt.mode), tt.want.SupportTypes) {
				t.Fatalf("AvailableSupportTypes(%q) = %v, want %v", tt.mode, AvailableSupportTypes(tt.mode), tt.want.SupportTypes)
			}
		})
	}
}

// TestModeCapabilityDedicatedVRAMMatchesUtils keeps the table on the set that
// TestDedicatedGPUMemoryTypesMatchOAC ties to the manifest validator.
func TestModeCapabilityDedicatedVRAMMatchesUtils(t *testing.T) {
	for _, mode := range allTestModes {
		if got, want := CapabilityFor(mode).DedicatedVRAM, utils.HasDedicatedGPUMemory(mode); got != want {
			t.Errorf("%s: DedicatedVRAM = %v, utils.HasDedicatedGPUMemory = %v", mode, got, want)
		}
	}
}

func TestModeCapabilityRowsAreConsistent(t *testing.T) {
	for _, mode := range allTestModes {
		c := CapabilityFor(mode)
		if len(c.SupportTypes) == 0 {
			t.Errorf("%s: no support types", mode)
		}
		for _, supportType := range c.SupportTypes {
			if _, ok := supportTypeToShareMode(supportType); !ok {
				t.Errorf("%s: support type %q has no share-mode code", mode, supportType)
			}
		}
		if c.HAMI && c.Register.Decoder != registerHAMI {
			t.Errorf("%s: HAMi mode must decode a HAMi register", mode)
		}
		if c.Register.Decoder == registerHAMI && c.Register.Handshake == "" {
			t.Errorf("%s: a HAMi register needs its handshake annotation", mode)
		}
		if c.MemorySliceReserve != 0 && !slices.Contains(c.SupportTypes, SupportTypeMemorySlice) {
			t.Errorf("%s: memory-slice reserve without MemorySlice", mode)
		}
		if c.Register.Decoder != registerNone && c.Register.Annotation == "" {
			t.Errorf("%s: register decoder without an annotation", mode)
		}
		if c.MemoryResource != "" && c.ExtendedResource == "" {
			t.Errorf("%s: memory resource without a device resource", mode)
		}
		if c.MultiCard && !c.DedicatedVRAM {
			t.Errorf("%s: multi-card aggregation is sized in GPU memory, so it needs DedicatedVRAM", mode)
		}
		if c.DedicatedVRAM && !c.MultiCard {
			t.Errorf("%s: dedicated-VRAM modes must allow multi-card aggregation", mode)
		}
	}
}

func TestCapabilityForReturnsACopy(t *testing.T) {
	first := CapabilityFor(utils.NvidiaCardType)
	first.SupportTypes[0] = "mutated"
	if got := CapabilityFor(utils.NvidiaCardType).SupportTypes[0]; got != SupportTypeTimeSlice {
		t.Fatalf("mutating a returned row leaked into the table: %q", got)
	}
}

func TestSupportTypeOrder(t *testing.T) {
	withTimeSlice := []string{SupportTypeExclusive, SupportTypeMemorySlice, SupportTypeTimeSlice}
	without := []string{SupportTypeExclusive, SupportTypeMemorySlice}
	for _, mode := range allTestModes {
		want := without
		if mode == utils.NvidiaCardType {
			want = withTimeSlice
		}
		if got := supportTypeOrder(mode); !reflect.DeepEqual(got, want) {
			t.Errorf("supportTypeOrder(%s) = %v, want %v", mode, got, want)
		}
	}
}

func TestCardSupportTypes(t *testing.T) {
	tests := []struct {
		name string
		mode string
		card []string
		want []string
	}{
		{"no card list keeps the mode's", utils.NvidiaCardType, nil,
			[]string{SupportTypeTimeSlice, SupportTypeMemorySlice, SupportTypeExclusive}},
		{"card narrows to one", utils.NvidiaCardType, []string{SupportTypeExclusive},
			[]string{SupportTypeExclusive}},
		{"mode order wins over card order", utils.NvidiaCardType, []string{SupportTypeExclusive, SupportTypeMemorySlice},
			[]string{SupportTypeMemorySlice, SupportTypeExclusive}},
		{"card cannot add a type", utils.GB10ChipType, []string{SupportTypeTimeSlice, SupportTypeExclusive},
			[]string{SupportTypeExclusive}},
		{"nothing in common keeps the mode's", utils.AMDGPUType, []string{SupportTypeMemorySlice},
			[]string{SupportTypeExclusive}},
		{"intel card without libvxpu", utils.IntelGPUType, []string{SupportTypeExclusive},
			[]string{SupportTypeExclusive}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cardSupportTypes(tt.mode, tt.card); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("cardSupportTypes(%s, %v) = %v, want %v", tt.mode, tt.card, got, tt.want)
			}
		})
	}
}

func TestResolveSupportType(t *testing.T) {
	all := AvailableSupportTypes(utils.NvidiaCardType)
	tests := []struct {
		name      string
		available []string
		shareMode string
		want      string
	}{
		{"unset falls back to the first", all, "", SupportTypeTimeSlice},
		{"exclusive", all, "0", SupportTypeExclusive},
		{"memory slice", all, "1", SupportTypeMemorySlice},
		{"time slice", all, "2", SupportTypeTimeSlice},
		{"unknown code falls back", all, "9", SupportTypeTimeSlice},
		{"type the device lacks falls back", AvailableSupportTypes(utils.GB10ChipType), "2", SupportTypeMemorySlice},
		{"narrowed card ignores a stale annotation", []string{SupportTypeExclusive}, "1", SupportTypeExclusive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveSupportType(tt.available, tt.shareMode); got != tt.want {
				t.Fatalf("resolveSupportType(%v, %q) = %q, want %q", tt.available, tt.shareMode, got, tt.want)
			}
		})
	}
}

func TestRegisteredDeviceAppliesCardSupportTypes(t *testing.T) {
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{
		Name:        "gpu-a",
		Annotations: map[string]string{shareModeAnnotationKey("GPU-0"): "1"},
	}}
	tests := []struct {
		name          string
		card          []string
		wantAvailable []string
		wantType      string
	}{
		{"register without support types", nil, AvailableSupportTypes(utils.NvidiaCardType), SupportTypeMemorySlice},
		{"register narrows the card", []string{SupportTypeExclusive}, []string{SupportTypeExclusive}, SupportTypeExclusive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := registeredDevice(node, utils.NvidiaCardType, registeredCard{
				ID: "GPU-0", Memory: 24 * gi, Health: deviceHealthYes, SupportTypes: tt.card,
			})
			if !reflect.DeepEqual(device.AvailableSupportTypes, tt.wantAvailable) || device.SupportType != tt.wantType {
				t.Fatalf("device = %+v, want available %v and type %s", device, tt.wantAvailable, tt.wantType)
			}
		})
	}
}

func TestAggregatesCards(t *testing.T) {
	tests := []struct {
		req  Requirement
		want bool
	}{
		{Requirement{Mode: utils.NvidiaCardType}, false},
		{Requirement{Mode: utils.NvidiaCardType, SupportMultiCards: true}, true},
		{Requirement{Mode: utils.NvidiaCardType, SupportMultiNodes: true}, true},
		{Requirement{Mode: utils.GB10ChipType, SupportMultiCards: true}, false},
		{Requirement{Mode: utils.IntelGPUType, SupportMultiNodes: true}, true},
		{Requirement{Mode: utils.AMDGPUType, SupportMultiCards: true}, true},
	}
	for _, tt := range tests {
		if got := aggregatesCards(tt.req); got != tt.want {
			t.Errorf("aggregatesCards(%+v) = %v, want %v", tt.req, got, tt.want)
		}
	}
}

func TestMemorySliceReserve(t *testing.T) {
	card := func(mode string, bound ...int64) Device {
		d := Device{ID: "card", Mode: mode, Memory: 32 * gi, SupportType: SupportTypeMemorySlice}
		for _, memory := range bound {
			d.Bindings = append(d.Bindings, Allocation{Memory: memory})
		}
		return d
	}
	reserve := int64(256 * mib)

	if got, want := deviceAvailableMemory(card(utils.IntelGPUType, 8*gi)), 24*gi-2*reserve; got != want {
		t.Errorf("intel card with one 8Gi binding: available %d, want %d", got, want)
	}
	if got := deviceAvailableMemory(card(utils.NvidiaCardType, 8*gi)); got != 24*gi {
		t.Errorf("nvidia books no reserve: available %d, want %d", got, 24*gi)
	}
	if got := deviceAvailableMemory(card(utils.IntelGPUType, 16*gi, 16*gi)); got != 0 {
		t.Errorf("full intel card: available %d, want 0", got)
	}

	nodes := []Node{{NodeName: "n", GPUTypes: []string{utils.IntelGPUType}, Devices: []Device{card(utils.IntelGPUType, 16*gi)}}}
	req := Requirement{Mode: utils.IntelGPUType, RequiredGPU: 16 * gi, LimitedGPU: 16 * gi}
	if res := ValidateBindingSelection(req, []BindingSelection{{NodeName: "n", DeviceID: "card", Memory: 16 * gi}}, nodes, PressureSnapshot{}); res.OK {
		t.Errorf("a quota that only fits without the reserve was accepted: %+v", res)
	}
	req.RequiredGPU, req.LimitedGPU = 15*gi, 15*gi
	if res := ValidateBindingSelection(req, []BindingSelection{{NodeName: "n", DeviceID: "card", Memory: 15 * gi}}, nodes, PressureSnapshot{}); !res.OK {
		t.Errorf("a quota that fits with the reserve was rejected: %+v", res)
	}
}

func TestAttachBindingsReportsReserve(t *testing.T) {
	nodes := []Node{{NodeName: "n", Devices: []Device{
		{ID: "intel-slice", Mode: utils.IntelGPUType, Memory: 32 * gi, SupportType: SupportTypeMemorySlice},
		{ID: "intel-exclusive", Mode: utils.IntelGPUType, Memory: 32 * gi, SupportType: SupportTypeExclusive},
		{ID: "nvidia-slice", Mode: utils.NvidiaCardType, Memory: 24 * gi, SupportType: SupportTypeMemorySlice},
	}}}
	attachBindings(nodes, []Allocation{
		{NodeName: "n", DeviceID: "intel-slice", Memory: 8 * gi},
		{NodeName: "n", DeviceID: "intel-slice", Memory: 4 * gi},
		{NodeName: "n", DeviceID: "intel-exclusive"},
		{NodeName: "n", DeviceID: "nvidia-slice", Memory: 8 * gi},
	})
	want := []int64{2 * 256 * mib, 0, 0}
	for i, device := range nodes[0].Devices {
		if device.Reserved != want[i] {
			t.Errorf("%s reserved = %d, want %d", device.ID, device.Reserved, want[i])
		}
	}
}

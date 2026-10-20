package compute

import (
	"slices"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// registerDecoder names the format a mode's devices are read from.
type registerDecoder int

const (
	// registerNone: the mode publishes no device list, so the whole node is
	// one device drawing from system memory.
	registerNone registerDecoder = iota
	// registerHAMI: HAMi's device list, one card per
	// "id,count,devmem,devcore,type,numa,health[,index,mode[,arch]]" entry.
	registerHAMI
	// registerTuple: the bytetrade "kind,card,driver,name,arch,codename,mem"
	// list. Only discrete entries become devices; a list without any falls
	// back to the registerNone device.
	registerTuple
)

// DeviceRegister is where a mode's devices are published on the node.
type DeviceRegister struct {
	Annotation string
	Decoder    registerDecoder
	// Handshake is the registerHAMI liveness annotation paired with Annotation.
	Handshake string
	// Cards optionally names a JSON annotation that narrows each card's
	// support types (see decodeRegisteredCardInfo).
	Cards string
}

// ModeCapability is everything app-service decides per accelerator mode.
// Scheduling, validation and webhook rules read these fields instead of
// comparing mode names, so a new mode is one more row in modeCapabilities.
type ModeCapability struct {
	// SupportTypes a device of this mode may take; the first is the default
	// for a device whose share mode was never configured.
	SupportTypes []string
	// HAMI: placement goes through HAMi and a GPUBinding per allocation.
	HAMI bool
	// ExtendedResource is the per-pod device count resource. Empty means the
	// pod reaches its node through a nodeSelector instead.
	ExtendedResource string
	// MemoryResource is the per-container memory limit the webhook injects
	// next to ExtendedResource when the allocation carries a memory quota.
	MemoryResource string
	// DedicatedVRAM: devices own memory separate from system RAM, so an app
	// is sized by its GPU-memory quota rather than its pod memory request.
	DedicatedVRAM bool
	// RuntimeClass the webhook sets on the pod, if any.
	RuntimeClass string
	// MultiCard: one app may aggregate several devices, on one node or across
	// nodes, when its manifest declares supportMultiCards / supportMultiNodes.
	MultiCard bool
	// MemorySliceReserve is device memory, in bytes, HAMi books per
	// memory-slice binding on top of its quota, for allocations the memory
	// hook cannot count. A card's free memory is reduced by it once per
	// existing binding and once for the binding being placed.
	MemorySliceReserve int64
	Register           DeviceRegister
}

var modeCapabilities = map[string]ModeCapability{
	utils.CPUType: {
		SupportTypes: []string{SupportTypeMemorySlice},
	},
	utils.NvidiaCardType: {
		SupportTypes:     []string{SupportTypeTimeSlice, SupportTypeMemorySlice, SupportTypeExclusive},
		HAMI:             true,
		ExtendedResource: constants.NvidiaGPU,
		MemoryResource:   constants.NvidiaGPUMem,
		DedicatedVRAM:    true,
		RuntimeClass:     hamiRuntimeClass,
		MultiCard:        true,
		Register:         nvidiaRegister,
	},
	utils.GB10ChipType: {
		SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
		HAMI:             true,
		ExtendedResource: constants.NvidiaGPU,
		MemoryResource:   constants.NvidiaGPUMem,
		RuntimeClass:     hamiRuntimeClass,
		Register:         nvidiaRegister,
	},
	utils.AppleMChipType: {
		SupportTypes: []string{SupportTypeExclusive},
	},
	utils.IntelType: {
		SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
		ExtendedResource: constants.IntelIGPU,
	},
	utils.AMDType: {
		SupportTypes:     []string{SupportTypeMemorySlice, SupportTypeExclusive},
		ExtendedResource: constants.AMDGPU,
	},
	utils.IntelGPUType: {
		SupportTypes:       []string{SupportTypeExclusive, SupportTypeMemorySlice},
		HAMI:               true,
		ExtendedResource:   constants.IntelVGPU,
		MemoryResource:     constants.IntelVGPUMem,
		DedicatedVRAM:      true,
		MultiCard:          true,
		MemorySliceReserve: intelMemorySliceReserve,
		Register: DeviceRegister{
			Annotation: constants.NodeIntelHAMIRegisterKey,
			Decoder:    registerHAMI,
			Handshake:  constants.NodeIntelHandshakeKey,
			Cards:      constants.NodeIntelCardsKey,
		},
	},
	utils.AMDGPUType: {
		SupportTypes:     []string{SupportTypeExclusive},
		ExtendedResource: constants.AMDGPU,
		DedicatedVRAM:    true,
		MultiCard:        true,
		Register:         DeviceRegister{Annotation: constants.NodeAmdRegisterKey, Decoder: registerTuple},
	},
	utils.MooreSocType: {
		SupportTypes: []string{SupportTypeMemorySlice, SupportTypeExclusive},
	},
}

// unknownModeCapability keeps a mode missing from the table schedulable as a
// whole node, without extended resources or memory slicing.
var unknownModeCapability = ModeCapability{
	SupportTypes: []string{SupportTypeExclusive},
}

const hamiRuntimeClass = "nvidia"

var nvidiaRegister = DeviceRegister{
	Annotation: constants.NodeNvidiaRegistryKey,
	Decoder:    registerHAMI,
	Handshake:  constants.NodeHandshakeKey,
}

// intelMemorySliceReserve matches the default of HAMi's intel.memoryReserve.
const intelMemorySliceReserve = 256 * mib

// CapabilityFor returns the capability row for a mode. The SupportTypes slice
// is a copy, so callers may modify it.
func CapabilityFor(mode string) ModeCapability {
	capability, ok := modeCapabilities[mode]
	if !ok {
		capability = unknownModeCapability
	}
	capability.SupportTypes = slices.Clone(capability.SupportTypes)
	return capability
}

// cardSupportTypes narrows a mode's support types to those a register entry
// says one card can take, keeping the mode's order so its default stays first
// when the card allows it. An empty card list, or one sharing nothing with the
// mode, leaves the mode's list unchanged: a card can only give up support
// types, never gain one, and never be left with none.
func cardSupportTypes(mode string, card []string) []string {
	modeTypes := AvailableSupportTypes(mode)
	if len(card) == 0 {
		return modeTypes
	}
	narrowed := make([]string, 0, len(modeTypes))
	for _, supportType := range modeTypes {
		if slices.Contains(card, supportType) {
			narrowed = append(narrowed, supportType)
		}
	}
	if len(narrowed) == 0 {
		return modeTypes
	}
	return narrowed
}

// resolveSupportType reads a device's share-mode annotation against the
// support types the device may take, falling back to the first of them when
// the annotation is unset or names a type the device does not offer.
func resolveSupportType(available []string, shareMode string) string {
	for _, supportType := range available {
		if code, _ := supportTypeToShareMode(supportType); code == shareMode {
			return supportType
		}
	}
	return available[0]
}

// aggregatesCards reports whether the request is placed across several
// devices rather than on one.
func aggregatesCards(req Requirement) bool {
	return CapabilityFor(req.Mode).MultiCard && (req.SupportMultiCards || req.SupportMultiNodes)
}

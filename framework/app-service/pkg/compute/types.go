package compute

import (
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

const (
	StatusInstallable           = "installable"
	StatusInsufficientResources = "insufficient-resources"
	StatusNoMatchingNode        = "no-matching-node"

	SupportTypeTimeSlice    = "TimeSlice"
	SupportTypeMemorySlice  = "MemorySlice"
	SupportTypeExclusive    = "Exclusive"
	SupportTypeMemoryShared = "MemoryShared"

	allocationConfigMapNamespace = "os-framework"
	allocationConfigMapName      = "app-gpu-allocations"
	allocationConfigMapKey       = "allocations.json"

	managedByLabelKey      = "app.bytetrade.io/managed-by"
	allocationModeLabelKey = "gpu.bytetrade.io/mode"
	managedByAppService    = "app-service"

	hamiGPUBindingAPIVersion = "gpu.bytetrade.io/v1alpha1"
	hamiGPUBindingKind       = "GPUBinding"
	hamiGPUBindingListKind   = "GPUBindingList"
)

type Node struct {
	NodeName       string `json:"nodeName"`
	GPUType        string `json:"gpuType"`
	Health         string `json:"health,omitempty"`
	memoryCapacity int64
	Devices        []Device `json:"devices"`
}

type Device struct {
	ID                    string       `json:"id"`
	NodeName              string       `json:"nodeName"`
	CardModel             string       `json:"cardModel,omitempty"`
	Memory                int64        `json:"memory"`
	Health                string       `json:"health"`
	SupportType           string       `json:"supportType"`
	AvailableSupportTypes []string     `json:"availableSupportTypes,omitempty"`
	Bindings              []Allocation `json:"bindings,omitempty"`
}

type Allocation struct {
	AppID    string `json:"appId,omitempty"`
	AppName  string `json:"appName"`
	Owner    string `json:"owner,omitempty"`
	Mode     string `json:"mode"`
	NodeName string `json:"nodeName"`
	DeviceID string `json:"deviceId"`
	Memory   int64  `json:"memory"`
}

type Requirement struct {
	Mode              string `json:"mode"`
	RequiredCPU       int64  `json:"requiredCpu"`
	RequiredGPU       int64  `json:"requiredGpu"`
	LimitedGPU        int64  `json:"limitedGpu"`
	RequiredMemory    int64  `json:"requiredMemory"`
	LimitedMemory     int64  `json:"limitedMemory"`
	RequiredDisk      int64  `json:"requiredDisk"`
	SupportMultiCards bool   `json:"supportMultiCards"`
	SupportMultiNodes bool   `json:"supportMultiNodes"`
}

type ModePlanResult struct {
	ComputeType string `json:"computeType"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
}

const (
	NodeStatusNotMatch     = "not_match"
	NodeStatusNotEnough    = "not_enough"
	NodeStatusNotAvailable = "not_available"
	NodeStatusAvailable    = "available"

	AvailabilityScopeCard          = "card"
	AvailabilityScopeSingleNode    = "single-node-cards"
	AvailabilityScopeCrossNode     = "cross-node-cards"
	AvailabilityScopeNode          = "node"
	FitLevelLimit                  = "limit"
	FitLevelRequired               = "required"
	BindingApplyStatusNotRequired  = "not-required"
	BindingApplyStatusRequired     = "required"
	BindingApplyStatusUnavailable  = "unavailable"
	BindingApplyStatusApplied      = "applied"
	BindingValidationReasonValid   = "valid"
	BindingValidationReasonInvalid = "invalid"
	deviceHealthYes                = "yes"
	deviceHealthNo                 = "no"
)

type DeviceOption struct {
	NodeName    string `json:"nodeName"`
	DeviceID    string `json:"deviceId"`
	CardModel   string `json:"cardModel,omitempty"`
	SupportType string `json:"supportType"`
	Capacity    int64  `json:"capacity"`
	Available   int64  `json:"available"`
	FitLevel    string `json:"fitLevel,omitempty"`
	Health      string `json:"health"`
	Operable    bool   `json:"operable"`
}

type NodeOption struct {
	NodeName string         `json:"nodeName"`
	GPUType  string         `json:"gpuType"`
	Status   string         `json:"status"`
	Devices  []DeviceOption `json:"devices"`
}

type AvailabilityResult struct {
	Schedulable bool         `json:"schedulable"`
	Requirement Requirement  `json:"requirement"`
	Scope       string       `json:"scope"`
	Nodes       []NodeOption `json:"nodes"`
	Reason      string       `json:"reason,omitempty"`
}

type BindingSelection struct {
	NodeName string `json:"nodeName"`
	DeviceID string `json:"deviceId"`
	Memory   int64  `json:"memory,omitempty"`
}

type BindingValidationResult struct {
	OK     bool   `json:"ok"`
	Code   string `json:"code,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type BindingApplyResult struct {
	Status       string                   `json:"status"`
	Allocations  []Allocation             `json:"allocations,omitempty"`
	Availability *AvailabilityResult      `json:"availability,omitempty"`
	Validation   *BindingValidationResult `json:"validation,omitempty"`
	TargetApp    string                   `json:"-"`
	TargetOwner  string                   `json:"-"`
}

type PressureSnapshot struct {
	Threshold   float64
	UsageByNode map[string]prometheus.NodeResourceUsage
}

type AddedResources struct {
	CPU    int64
	Memory int64
	Disk   int64
}

// AddedResourcesFromAppConfig translates the app's selected
// ResourceMode (or its legacy scalar Requirement) into a per-node
// AddedResources budget suitable for PressureSnapshot.WouldPressure.
//
// When the app declares an explicit resource mode (>= 0.12.0 manifest
// format), the values come from that mode. For legacy manifests the
// values come from appConfig.Requirement after ResolveRequirement has
// applied any supportedGpu overrides.
//
// Returns the zero value when appConfig is nil so callers can pass it
// straight into WouldPressure even when the app has no declared
// requirement at all.
func AddedResourcesFromAppConfig(appConfig *appcfg.ApplicationConfig) AddedResources {
	if appConfig == nil {
		return AddedResources{}
	}
	if req, ok := SelectedRequirement(appConfig); ok {
		return AddedResources{
			CPU:    req.RequiredCPU,
			Memory: req.RequiredMemory,
			Disk:   req.RequiredDisk,
		}
	}
	var cpu, mem, disk int64
	if r := appConfig.Requirement.CPU; r != nil {
		cpu = r.MilliValue()
	}
	if r := appConfig.Requirement.Memory; r != nil {
		mem = r.Value()
	}
	if r := appConfig.Requirement.Disk; r != nil {
		disk = r.Value()
	}
	return AddedResources{CPU: cpu, Memory: mem, Disk: disk}
}

func IsHAMIMode(mode string) bool {
	switch mode {
	case utils.NvidiaCardType, utils.GB10ChipType:
		return true
	default:
		return false
	}
}

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

	SupportTypeTimeSlice   = "TimeSlice"
	SupportTypeMemorySlice = "MemorySlice"
	SupportTypeExclusive   = "Exclusive"

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

// Node is a physical cluster node in the compute view. A node may support
// several accelerator modes at once (e.g. an Olares One exposing both nvidia
// and intel), so GPUTypes is a list rather than a single value; cpu is implicit
// on every node and is not listed here. Each Device carries the Mode it serves,
// so callers can filter a node's devices to a single mode (see Node.viewForMode
// / matchingNodes) without losing the multi-mode view this struct exposes to
// the API.
type Node struct {
	NodeName       string   `json:"nodeName"`
	GPUTypes       []string `json:"gpuTypes"`
	Health         string   `json:"health,omitempty"`
	memoryCapacity int64
	Devices        []Device `json:"devices"`
}

type Device struct {
	ID                    string       `json:"id"`
	NodeName              string       `json:"nodeName"`
	Mode                  string       `json:"mode"`
	CardModel             string       `json:"cardModel,omitempty"`
	Memory                int64        `json:"memory"`
	Health                string       `json:"health"`
	SupportType           string       `json:"supportType"`
	AvailableSupportTypes []string     `json:"availableSupportTypes,omitempty"`
	Bindings              []Allocation `json:"bindings,omitempty"`
}

// SupportsMode reports whether the node can host a pod of the given mode. cpu
// is universal — every node can run cpu workloads — while non-cpu modes must be
// advertised in GPUTypes.
func (n Node) SupportsMode(mode string) bool {
	if mode == utils.CPUType {
		return true
	}
	for _, m := range n.GPUTypes {
		if m == mode {
			return true
		}
	}
	return false
}

// viewForMode returns a copy of the node projected onto a single mode: only the
// devices serving that mode are retained. The device-centric scheduler /
// availability code operates on these single-mode views so it never has to
// reason about a node's other modes.
func (n Node) viewForMode(mode string) Node {
	view := n
	devices := make([]Device, 0, len(n.Devices))
	for _, d := range n.Devices {
		if d.Mode == mode {
			devices = append(devices, d)
		}
	}
	view.Devices = devices
	return view
}

// primaryGPUType returns a single representative mode for display on nodes that
// don't match a requested mode (e.g. the availability "not-match" listing).
func (n Node) primaryGPUType() string {
	if len(n.GPUTypes) > 0 {
		return n.GPUTypes[0]
	}
	return utils.CPUType
}

type Allocation struct {
	AppID    string `json:"appId,omitempty"`
	AppName  string `json:"appName"`
	Owner    string `json:"owner,omitempty"`
	Mode     string `json:"mode"`
	NodeName string `json:"nodeName"`
	DeviceID string `json:"deviceId"`
	Memory   int64  `json:"memory"`
	// Spec carries the bound app's currently-selected resource-mode
	// requirement (require/limit GPU & memory, multi-card / multi-node
	// support). It is resolved at read time for the compute-resources
	// listing (see AttachBoundAppSpecs) and is never persisted to the
	// allocation config map.
	Spec *Requirement `json:"spec,omitempty"`
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
	FitLevelLimit                  = "limit"
	FitLevelRequired               = "required"
	BindingApplyStatusNotRequired  = "not-required"
	BindingApplyStatusRequired     = "required"
	BindingApplyStatusUnavailable  = "unavailable"
	BindingApplyStatusApplied      = "applied"
	BindingApplyStatusValid        = "valid"
	BindingValidationReasonValid   = "valid"
	BindingValidationReasonInvalid = "invalid"
	deviceHealthYes                = "yes"
	deviceHealthNo                 = "no"
)

type DeviceOption struct {
	NodeName    string       `json:"nodeName"`
	DeviceID    string       `json:"deviceId"`
	CardModel   string       `json:"cardModel,omitempty"`
	SupportType string       `json:"supportType"`
	Capacity    int64        `json:"capacity"`
	Available   int64        `json:"available"`
	FitLevel    string       `json:"fitLevel,omitempty"`
	Health      string       `json:"health"`
	Operable    bool         `json:"operable"`
	Bindings    []Allocation `json:"bindings,omitempty"`
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
	// NodePressure is populated only when the binding is rejected because
	// a selected node would be pushed past its resource pressure
	// threshold (Code "node-pressure:<node>"). It breaks the rejection
	// down per resource so the caller can tell whether cpu, memory, or
	// both fell short, how much the app needs, and how much headroom the
	// node still has.
	NodePressure *NodePressureDetail `json:"nodePressure,omitempty"`
}

// NodePressureDetail carries the per-resource breakdown behind a
// "node-pressure" binding rejection for a single node.
type NodePressureDetail struct {
	NodeName   string              `json:"nodeName"`
	Dimensions []DimensionPressure `json:"dimensions"`
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

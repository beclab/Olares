// Package nodestatus builds the node-local view every olaresd serves about
// itself: who it is, how it is doing, and what it can be asked to do.
//
// Health, connectivity and phase are three separate answers. A node that is
// restarting is not unhealthy, and a node that cannot be reached has not been
// proven to be off. Anything this node cannot confirm is reported as unknown
// rather than guessed, so that a caller aggregating several nodes never turns
// a missing fact into a false one.
package nodestatus

import (
	"strings"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// Health is what the node knows about its own well-being.
type Health string

const (
	HealthHealthy  Health = "healthy"
	HealthDegraded Health = "degraded"
	HealthUnknown  Health = "unknown"
)

// Connectivity is how reachable a node is. A node only ever reports itself as
// online: the other values are for the caller that failed to reach it, and
// Offline requires a confirmed shutdown rather than a timeout.
type Connectivity string

const (
	ConnectivityOnline      Connectivity = "online"
	ConnectivityStale       Connectivity = "stale"
	ConnectivityUnreachable Connectivity = "unreachable"
	ConnectivityOffline     Connectivity = "offline"
)

// Phase is what the node is currently doing.
type Phase string

const (
	PhaseRunning      Phase = "running"
	PhaseMaintenance  Phase = "maintenance"
	PhaseRestarting   Phase = "restarting"
	PhaseShuttingDown Phase = "shutting_down"
	PhaseUnknown      Phase = "unknown"
)

// ConditionOlaresStateAbnormal explains a degraded node in the structured
// condition list, so the summary above it stays short.
const ConditionOlaresStateAbnormal = "OlaresStateAbnormal"

// Condition is one structured fact behind the summarized health.
type Condition struct {
	Type    string `json:"type"`
	Status  bool   `json:"status"`
	Message string `json:"message,omitempty"`
}

// Identity is who this node is. Fields left empty mean the node could not
// resolve them; they are reported as unknown rather than filled in.
type Identity struct {
	NodeName   string
	Role       inventory.Role
	Hostname   string
	DeviceType string
}

// Status is the node-local response body.
type Status struct {
	NodeName   string         `json:"nodeName"`
	Hostname   string         `json:"hostname"`
	Role       inventory.Role `json:"role"`
	DeviceType string         `json:"deviceType"`

	Health       Health       `json:"health"`
	Connectivity Connectivity `json:"connectivity"`
	Phase        Phase        `json:"phase"`

	// TerminusState is the raw state olaresd already reports on
	// /system/status, kept verbatim so clients can map it themselves during
	// the compatibility window and so an unknown value is never swallowed.
	TerminusState clistate.TerminusState `json:"terminusState"`

	CPU  string   `json:"cpu"`
	GPUs []string `json:"gpus"`

	// Memory and Disk are the totals as the host formatted them, e.g.
	// "128 G". They are strings on purpose and are named without a unit: the
	// daemon has never held the byte counts behind them, so presenting one
	// would mean inventing it. Empty means this node could not work it out.
	Memory string `json:"memory"`
	Disk   string `json:"disk"`

	// OlaresdVersion is the version of the olaresd that answered. Version
	// drift between nodes is the usual explanation for one node offering a
	// capability its neighbour does not, so the detail page shows it.
	OlaresdVersion string `json:"olaresdVersion"`

	// ObservedAt is when the state below was last refreshed, not when this
	// response was produced. It is null until the first refresh, because a
	// response that timestamps itself is a stale answer that looks fresh.
	ObservedAt   *time.Time            `json:"observedAt"`
	Conditions   []Condition           `json:"conditions"`
	Capabilities map[string]Capability `json:"capabilities"`
}

// Build assembles the node-local status from this node's identity and the
// state snapshot observed at observedAt.
func Build(id Identity, st clistate.State, caps map[string]Capability, observedAt time.Time) Status {
	role := id.Role
	if role == "" {
		role = inventory.RoleUnknown
	}
	deviceType := id.DeviceType
	if deviceType == "" {
		deviceType = DeviceTypeUnknown
	}
	if caps == nil {
		caps = map[string]Capability{}
	}

	gpus := st.GPUList
	if gpus == nil {
		gpus = []string{}
	}

	health := healthFor(st.TerminusState)

	conditions := make([]Condition, 0, len(st.Pressure)+1)
	if health == HealthDegraded {
		conditions = append(conditions, Condition{
			Type:    ConditionOlaresStateAbnormal,
			Status:  true,
			Message: st.TerminusState.Describe(),
		})
	}
	for _, p := range st.Pressure {
		conditions = append(conditions, Condition{Type: p.Type, Status: true, Message: p.Message})
	}

	var observed *time.Time
	if !observedAt.IsZero() {
		at := observedAt
		observed = &at
	}

	var olaresdVersion string
	if st.OlaresdVersion != nil {
		olaresdVersion = *st.OlaresdVersion
	}

	return Status{
		NodeName:       id.NodeName,
		Hostname:       id.Hostname,
		Role:           role,
		DeviceType:     deviceType,
		Health:         health,
		Connectivity:   ConnectivityOnline,
		Phase:          phaseFor(st.TerminusState),
		TerminusState:  st.TerminusState,
		CPU:            st.CpuInfo,
		GPUs:           gpus,
		Memory:         st.Memory,
		Disk:           st.Disk,
		OlaresdVersion: olaresdVersion,
		ObservedAt:     observed,
		Conditions:     conditions,
		Capabilities:   caps,
	}
}

func healthFor(s clistate.TerminusState) Health {
	switch s {
	case clistate.TerminusRunning:
		return HealthHealthy
	case clistate.SystemError, clistate.NetworkNotReady, clistate.InvalidIpAddress,
		clistate.InstallFailed, clistate.InitializeFailed, clistate.IPChangeFailed:
		return HealthDegraded
	default:
		return HealthUnknown
	}
}

func phaseFor(s clistate.TerminusState) Phase {
	switch s {
	case clistate.Restarting:
		return PhaseRestarting
	case clistate.Shutdown:
		return PhaseShuttingDown
	case clistate.Installing, clistate.Initializing, clistate.Uninstalling, clistate.Upgrading,
		clistate.IPChanging, clistate.DiskModifing, clistate.AddingNode, clistate.RemovingNode,
		clistate.SelfRepairing:
		return PhaseMaintenance
	case clistate.TerminusRunning, clistate.SystemError, clistate.NetworkNotReady,
		clistate.InvalidIpAddress, clistate.InstallFailed, clistate.InitializeFailed,
		clistate.IPChangeFailed:
		// Olares is up on this node, healthy or not, so it is running a phase
		// somebody can act on. States where it is absent (not-installed,
		// uninitialized) fall through: the host is powered on, but calling
		// that "running" would describe a system that is not there.
		return PhaseRunning
	default:
		return PhaseUnknown
	}
}

// Device type slugs. Only values this daemon can actually derive from the host
// are named here; everything else is slugified from what the host reports.
const (
	DeviceTypeUnknown   = "unknown"
	DeviceTypeOlaresOne = "olares-one"
)

// DeviceType maps the device name the host declares (/etc/machine.info, read
// by utils.GetDeviceName) onto a stable slug. It describes the device only:
// what the device can do is answered by the capability probes, so an
// unrecognized name is a display concern and never a functional one.
func DeviceType(deviceName *string) string {
	if deviceName == nil {
		return DeviceTypeUnknown
	}
	name := strings.TrimSpace(*deviceName)
	if name == "" {
		return DeviceTypeUnknown
	}
	// The Olares One installer identifies the device by the same
	// whitespace-insensitive comparison (build/base-package/joincluster.sh).
	if strings.EqualFold(strings.Join(strings.Fields(name), ""), "olaresone") {
		return DeviceTypeOlaresOne
	}
	return slugify(name)
}

func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			b.WriteRune(r)
		default:
			dash = true
		}
	}
	if b.Len() == 0 {
		return DeviceTypeUnknown
	}
	return b.String()
}

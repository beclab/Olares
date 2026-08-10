// Package clusterop is the master olaresd's cluster-level power orchestrator:
// it turns "reboot the cluster" or "shut the cluster down" into an ordered,
// recorded sequence of node-local commands.
//
// Two properties shape everything here. The compute nodes go first and the
// control node last, because the control node is the one running this code and
// nothing after it can be observed. And the final state of an operation that
// reached the control node is command_issued, never succeeded: the machine
// that would confirm the result is the machine being powered off.
package clusterop

import (
	"encoding/json"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// Type is the power operation a caller asked for. Each value is declared by
// the module that carries it out, next to the code that does the work; this
// package holds no list of them, because what this daemon can do is whatever
// registered itself.
type Type string

// ParseType accepts only what this daemon can actually carry out, which is
// exactly what a module has registered itself for.
func ParseType(s string) (Type, error) {
	return DefaultRegistry().Parse(s)
}

// Status is the state of a whole cluster operation.
type Status string

const (
	StatusPending         Status = "pending"
	StatusRunning         Status = "running"
	StatusSucceeded       Status = "succeeded"
	StatusPartiallyFailed Status = "partially_failed"
	StatusFailed          Status = "failed"

	// StatusCommandIssued is where a cluster power operation ends once the
	// control node has been told to go down.
	StatusCommandIssued Status = "command_issued"
)

// Terminal reports whether the operation has stopped moving. It is also what
// releases the single-operation lock, so command_issued has to be terminal:
// waiting for a confirmation that can never arrive would block the cluster's
// next power operation forever.
func (s Status) Terminal() bool {
	switch s {
	case StatusSucceeded, StatusPartiallyFailed, StatusFailed, StatusCommandIssued:
		return true
	default:
		return false
	}
}

// StepStatus is the state of one stage of an operation.
type StepStatus string

const (
	StepPending       StepStatus = "pending"
	StepRunning       StepStatus = "running"
	StepSucceeded     StepStatus = "succeeded"
	StepFailed        StepStatus = "failed"
	StepCommandIssued StepStatus = "command_issued"
	StepSkipped       StepStatus = "skipped"
)

// Step names.
const (
	StepPrecheck      = "precheck"
	StepWorkerCommand = "worker-power-command"
	StepWorkerRestart = "worker-restart"
	StepMasterCommand = "master-power-command"
)

// NodeStatus is the per-node outcome inside an operation.
type NodeStatus string

const (
	NodePending NodeStatus = "pending"

	// NodeCommandIssued means the node accepted the power command. For a
	// shutdown that is as far as observation goes.
	NodeCommandIssued NodeStatus = "command_issued"

	// NodeRestarted means the node went away and came back Ready.
	NodeRestarted NodeStatus = "restarted"

	NodeFailed  NodeStatus = "failed"
	NodeSkipped NodeStatus = "skipped"
)

// Stable error codes. They are part of the contract with user-service and
// TermiPass: the human-readable message may change, these may not.
const (
	CodeInventoryUnavailable   = "inventory_unavailable"
	CodeNoMasterNode           = "no_master_node"
	CodeNodeUnaddressable      = "node_unaddressable"
	CodeNodeUnreachable        = "node_unreachable"
	CodePowerUnsupported       = "power_unsupported"
	CodeDispatchFailed         = "dispatch_failed"
	CodeNodeDidNotGoDown       = "node_did_not_go_down"
	CodeRestartTimeout         = "restart_timeout"
	CodeHostPowerFailed        = "host_power_failed"
	CodePrecheckFailed         = "precheck_failed"
	CodeWorkerCommandFailed    = "worker_command_failed"
	CodeWorkerRestartFailed    = "worker_restart_failed"
	CodeStatePersistenceFailed = "state_persistence_failed"
	CodeRequestInProgress      = "request_in_progress"

	// CodeUnsupportedTopology refuses a cluster this daemon cannot sequence.
	// One control node and its compute nodes is the only supported shape:
	// with two control nodes the second is powered off by the first one's
	// run, with nothing left to record what happened to it.
	CodeUnsupportedTopology = "unsupported_topology"

	// CodeSelfUnresolved refuses to guess which node this daemon runs on.
	// Guessing wrong makes the machine that is orchestrating the operation
	// into one of the machines it powers off first.
	CodeSelfUnresolved = "self_unresolved"

	// CodeNodeIdentityUnknown marks a node the directory could not describe.
	// It is distinct from power_unsupported, which is a claim the node made.
	CodeNodeIdentityUnknown = "node_identity_unknown"

	// CodeNodeNotReady refuses a compute node the cluster already considers
	// down, rather than dialling it and waiting out the restart timeout.
	CodeNodeNotReady = "node_not_ready"

	// CodeBootIDUnavailable refuses a reboot whose result could not be
	// proved: with no boot id to compare against, a node that never went
	// down is indistinguishable from one that restarted.
	CodeBootIDUnavailable = "boot_id_unavailable"

	// CodeDaemonRestarted marks an operation that was still moving when
	// olaresd stopped. Nothing observed how it ended, so it is reported as
	// failed rather than left running and blocking the next one.
	CodeDaemonRestarted = "daemon_restarted"
)

// Step is one stage of an operation.
type Step struct {
	Name       string     `json:"name"`
	Status     StepStatus `json:"status"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	Code       string     `json:"code,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// NodeResult is what happened to one node during an operation.
type NodeResult struct {
	NodeName   string         `json:"nodeName"`
	Role       inventory.Role `json:"role"`
	Status     NodeStatus     `json:"status"`
	StartedAt  *time.Time     `json:"startedAt"`
	FinishedAt *time.Time     `json:"finishedAt"`
	Code       string         `json:"code,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// Operation is the persisted record of one cluster power operation.
//
// It deliberately holds no credential of any kind. The record is written to
// the daemon's state directory and read back by anything that can read the
// disk; the access token and the JWS that authorized the operation live in
// memory for the length of the run and nowhere else.
type Operation struct {
	ID        string `json:"id"`
	Type      Type   `json:"type"`
	RequestID string `json:"requestId"`
	Scope     string `json:"scope"`
	Target    string `json:"target,omitempty"`
	ClusterID string `json:"clusterId"`

	// Owner is the Olares ID the operation was created for. It is part of the
	// idempotency key, so one owner's retry can never join another's run.
	Owner string `json:"owner"`

	// ParamsDigest is the caller intent hash for module params. The raw params
	// are never written to the operation record.
	ParamsDigest string `json:"paramsDigest,omitempty"`

	Status     Status     `json:"status"`
	Code       string     `json:"code,omitempty"`
	Error      string     `json:"error,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`

	// HostBootID is the boot the control node was on when it was told to
	// reboot. The daemon that comes up next compares it with the boot the
	// machine is on then: a different one is the only proof available that
	// the machine really went down, rather than olaresd merely restarting.
	// It is an opaque kernel identifier and authorizes nothing.
	HostBootID string `json:"hostBootId,omitempty"`

	// CommandIssuedUntil is the persisted transient lock deadline. The command
	// may outlive this daemon, so the deadline cannot live only in Manager.
	CommandIssuedUntil time.Time `json:"commandIssuedUntil,omitempty"`

	// ModuleState is optional restart evidence for a module. It must never
	// contain raw params or secrets.
	ModuleState json.RawMessage `json:"moduleState,omitempty"`

	Steps []Step       `json:"steps"`
	Nodes []NodeResult `json:"nodes"`
}

// Clone returns a copy that shares no slice, and no *time.Time inside those
// slices or the operation itself, with the original — so a value handed to a
// caller cannot be written back into the stored record, whether the caller
// reassigns a field or writes through a timestamp pointer it kept.
func (o Operation) Clone() Operation {
	out := o
	out.StartedAt = cloneTime(o.StartedAt)
	out.FinishedAt = cloneTime(o.FinishedAt)
	if o.ModuleState != nil {
		out.ModuleState = append(json.RawMessage(nil), o.ModuleState...)
	}
	if o.Steps != nil {
		out.Steps = append([]Step(nil), o.Steps...)
		for i := range out.Steps {
			out.Steps[i].StartedAt = cloneTime(out.Steps[i].StartedAt)
			out.Steps[i].FinishedAt = cloneTime(out.Steps[i].FinishedAt)
		}
	}
	if o.Nodes != nil {
		out.Nodes = append([]NodeResult(nil), o.Nodes...)
		for i := range out.Nodes {
			out.Nodes[i].StartedAt = cloneTime(out.Nodes[i].StartedAt)
			out.Nodes[i].FinishedAt = cloneTime(out.Nodes[i].FinishedAt)
		}
	}
	return out
}

// cloneTime copies the value a *time.Time points to, so two copies of a
// record can never alias the same timestamp memory.
func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

// PhaseFor derives the cluster phase from the operation currently in flight.
//
// ok is false when there is nothing in flight, and that is not the same as
// "running": the caller keeps whatever phase it already had. Health is a
// separate question and is never answered here — a cluster rebooting on
// purpose is not a cluster in trouble.
func PhaseFor(op *Operation) (nodestatus.Phase, bool) {
	if op == nil || op.Status.Terminal() {
		return "", false
	}
	return phaseOf(DefaultRegistry(), op)
}

// phaseOf asks the module that carries this operation out what it makes the
// cluster look like while it is happening. The module answers for its own
// operation and nothing else: whether the operation still is happening is the
// caller's question, and the two have different answers once the command has
// been issued and the machine has not gone down yet.
//
// An operation whose module is not registered imposes no phase. There is
// nothing left that knows what it was doing.
func phaseOf(registry *ModuleRegistry, op *Operation) (nodestatus.Phase, bool) {
	if op == nil {
		return "", false
	}
	module, ok := registry.Lookup(op.Type)
	if !ok {
		return "", false
	}
	return module.Phase(*op)
}

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

// StepStatus is the state of one step of an operation.
type StepStatus string

const (
	StepPending       StepStatus = "pending"
	StepRunning       StepStatus = "running"
	StepSucceeded     StepStatus = "succeeded"
	StepFailed        StepStatus = "failed"
	StepCommandIssued StepStatus = "command_issued"
	StepSkipped       StepStatus = "skipped"
)

// Step names. A power operation has a fixed set; an upgrade names its steps
// after the stages of the plan it is executing, so those are not listed here.
const (
	StepPrecheck      = "precheck"
	StepWorkerCommand = "worker-power-command"
	StepWorkerRestart = "worker-restart"
	StepMasterCommand = "master-power-command"

	// StepPlan is where an upgrade reads the plan it is about to run.
	StepPlan = "plan"

	// StepUpgradeReadiness is where every compute node is asked whether it
	// holds the version being rolled out, after it has been prepared.
	StepUpgradeReadiness = "upgrade-readiness"
)

// NodeStatus is the per-node outcome inside an operation.
type NodeStatus string

const (
	NodePending NodeStatus = "pending"

	// NodeRunning means the node has started the work and has not finished.
	// Only an upgrade uses it: a power command is accepted, not performed.
	NodeRunning NodeStatus = "running"

	// NodeCommandIssued means the node accepted the power command. For a
	// shutdown that is as far as observation goes.
	NodeCommandIssued NodeStatus = "command_issued"

	// NodeRestarted means the node went away and came back Ready.
	NodeRestarted NodeStatus = "restarted"

	// NodeSucceeded means the node finished the work it was given and said
	// so. An upgrade stage can report this because the node is still there
	// afterwards to be asked.
	NodeSucceeded NodeStatus = "succeeded"

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

	// CodeModuleFailed marks an operation whose module stopped without
	// leaving a usable outcome. Nothing is known about what it did, but
	// leaving the record running would hold the cluster's single-operation
	// lock until the daemon restarts.
	CodeModuleFailed = "module_failed"

	// CodeDaemonRestarted marks an operation that was still moving when
	// olaresd stopped. Nothing observed how it ended, so it is reported as
	// failed rather than left running and blocking the next one.
	CodeDaemonRestarted = "daemon_restarted"
)

// Stable codes an upgrade can end on.
const (
	// CodePlanUnavailable means the control node could not work out what the
	// upgrade consists of. Nothing has run at that point.
	CodePlanUnavailable = "plan_unavailable"

	// CodeStageFailed is the stage-level roll-up of a node that failed.
	CodeStageFailed = "stage_failed"

	// CodeStageBusy means a node was asked for a stage while it was still
	// running one for another operation. It is a refusal, not a failure of
	// the work: the node did not start anything.
	CodeStageBusy = "stage_busy"

	// CodeStageTimeout means a node was still running its stage when the
	// stage's deadline passed. It is not a claim that the work stopped.
	CodeStageTimeout = "stage_timeout"

	// CodeUpgradeUnsupported refuses a node that cannot run upgrade stages:
	// an olaresd from before staged upgrades, or one in a container that
	// cannot touch the host.
	CodeUpgradeUnsupported = "upgrade_unsupported"

	// CodeVersionMismatch means a node is not holding the version this
	// upgrade is for, so it would derive a different plan.
	CodeVersionMismatch = "version_mismatch"

	// CodeUpgradeCancelled means the upgrade stopped between two stages
	// because it was asked to, rather than because anything went wrong. The
	// stages that had already run stay run: an upgrade cannot be undone, and
	// the record says how far it got.
	CodeUpgradeCancelled = "upgrade_cancelled"
)

// Step is one step of an operation.
type Step struct {
	Name       string     `json:"name"`
	Status     StepStatus `json:"status"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	Code       string     `json:"code,omitempty"`
	Error      string     `json:"error,omitempty"`

	// Placement and MaxParallel describe where and how the step ran. They are
	// set by an upgrade, whose steps come from a plan and differ from one
	// another; a power operation's steps are fixed and leave them empty.
	//
	// Placement is deliberately not called Scope: Operation.Scope already means
	// whether an operation covers the cluster or one node, and a field of the
	// same name meaning which nodes a step ran on would be two answers to two
	// different questions wearing one word.
	Placement   string `json:"placement,omitempty"`
	MaxParallel int    `json:"maxParallel,omitempty"`

	// Nodes is the per-node outcome within this step.
	//
	// A power operation leaves it empty and uses Operation.Nodes: each node
	// takes part in one step, so a flat list says everything. An upgrade runs
	// every node through many steps, and a flat list would have to overwrite
	// the previous stage's result to record the next one — turning "node B
	// failed at the containerd migration" into "node B is running the version
	// flip" the moment the operation moved on without it.
	Nodes []NodeResult `json:"nodes,omitempty"`
}

// Clone returns a copy that shares no slice with the original.
func (s Step) Clone() Step {
	out := s
	if s.Nodes != nil {
		out.Nodes = append([]NodeResult(nil), s.Nodes...)
	}
	return out
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

	// StopRequested says the operator asked this operation to stop at its
	// next safe point.
	//
	// It is on the record rather than in the manager's memory for the same
	// reason everything else about an upgrade is: an upgrade restarts olaresd,
	// so a cancellation held in memory would be forgotten by the process that
	// comes back and the run would carry on as though nobody had asked. See
	// Manager.RequestStop.
	StopRequested bool `json:"stopRequested,omitempty"`

	// SupersededBy names the operation that took this one's request id over,
	// and is empty on every operation still answering for its own.
	//
	// It exists because a request id identifies at most one operation, and
	// everything reads it that way: the loader refuses two records claiming
	// one, the pruner removes the mapping when it removes a record, and Create
	// hands a repeat request whatever the id resolves to. A retry produces a
	// second record for the same request — see RetryableModule — so the older
	// one has to stop claiming it, while staying on disk as the history of
	// what was attempted. This is how it says so.
	SupersededBy string `json:"supersededBy,omitempty"`

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
			out.Steps[i] = out.Steps[i].Clone()
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
// nothing left that knows what it was doing, and neither has a module that
// panicked when asked — see safePhase.
func phaseOf(registry *ModuleRegistry, op *Operation) (nodestatus.Phase, bool) {
	if op == nil {
		return "", false
	}
	module, ok := registry.Lookup(op.Type)
	if !ok {
		return "", false
	}
	return safePhase(module, *op)
}

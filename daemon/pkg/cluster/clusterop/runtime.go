package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Sentinel errors a Runtime mutation can return to a module. They are the
// only way a module discovers that a mutation was rejected rather than
// applied: no method on the interface panics or silently no-ops. This list
// is not exhaustive and may grow — a module must always use errors.Is rather
// than assume these are the only checked-mutation failures it can see.
var (
	// ErrOperationTerminal rejects a mutation against an operation that is
	// no longer active — a normal terminal status (succeeded, failed,
	// partially_failed), or a command_issued operation whose grace deadline
	// has already passed — including one the manager forced there after a
	// state_persistence_failed settlement. A module's own goroutine keeps
	// running after any of that happens, and every further mutation must
	// see it and stop rather than resurrect a settled record.
	ErrOperationTerminal = errors.New("cluster operation is already terminal")

	// ErrStepNotFound rejects finishing a step that was never started.
	ErrStepNotFound = errors.New("operation step not found")

	// ErrStepSettled rejects finishing a step that has already settled.
	ErrStepSettled = errors.New("operation step already settled")

	// ErrNodeNotFound rejects updating a node that was never recorded.
	ErrNodeNotFound = errors.New("operation node not found")

	// ErrConcurrentUpdate rejects UpdateNode's compare-and-replace when some
	// other checked mutation committed a change to the same node between
	// this call's snapshot and its own attempt to commit. UpdateNode never
	// retries and never overwrites that concurrent change: the caller reads
	// the new state (Operation) and decides whether to try again.
	ErrConcurrentUpdate = errors.New("operation node changed concurrently")

	// ErrInvalidModuleState rejects SetModuleState with bytes that are not
	// valid JSON. A nil or empty value is not rejected — it means "clear
	// the recorded state" — but anything else must be a JSON document, since
	// ModuleState is read back and unmarshalled by a module's own Recover.
	ErrInvalidModuleState = errors.New("invalid module state")

	// ErrInvalidOutcome rejects Complete with a status, or a
	// status/deadline combination, that is not one an operation may
	// legally end a Run on.
	ErrInvalidOutcome = errors.New("invalid operation outcome")
)

// errStatePersistenceFailed is returned by a checked mutation whose write to
// the store itself failed. The operation record already carries the stable,
// public CodeStatePersistenceFailed a caller can act on; this value only
// tells the module in the same call stack that this particular mutation did
// not take effect. It is deliberately unexported: it is bookkeeping between
// Manager's own methods, not part of the contract a module reacts to.
var errStatePersistenceFailed = errors.New("cluster operation state could not be recorded")

// errOperationNotFound guards a checked mutation issued for an operation id
// the manager no longer holds. newRuntime is only ever handed the id of an
// operation the manager just created or loaded, so this should not happen in
// practice; it exists so a checked mutation never mutates through a stale
// id. It is deliberately unexported for the same reason as
// errStatePersistenceFailed above.
var errOperationNotFound = errors.New("cluster operation not found")

// operationRuntime is the only thing an OperationModule sees of the
// operation it is running. It holds no Store, no reference to Manager.ops,
// and no mutex of its own: every mutation is validated and persisted by the
// manager under its own lock, so a module can never observe or leave the
// record in a state the manager itself did not produce.
type operationRuntime struct {
	m   *Manager
	id  string
	ctx context.Context
}

// newRuntime binds a Runtime to one operation for the lifetime of one Run (or
// future Recover) call.
func newRuntime(m *Manager, id string, ctx context.Context) Runtime {
	return &operationRuntime{m: m, id: id, ctx: ctx}
}

func (rt *operationRuntime) Operation() (Operation, bool) {
	return rt.m.Get(rt.id)
}

func (rt *operationRuntime) CanContinue() bool {
	return rt.m.canContinue(rt.id)
}

func (rt *operationRuntime) Now() time.Time {
	return rt.m.deps.Now()
}

func (rt *operationRuntime) Context() context.Context {
	return rt.ctx
}

func (rt *operationRuntime) StartStep(name string) error {
	return rt.m.checkedUpdate(rt.id, rt.m.rejectSettled, func(op *Operation) {
		at := rt.m.deps.Now()
		op.Steps = append(op.Steps, Step{Name: name, Status: StepRunning, StartedAt: &at})
	})
}

func (rt *operationRuntime) FinishStep(name string, status StepStatus, code, reason string) error {
	validate := func(op *Operation) error {
		if err := rt.m.rejectSettled(op); err != nil {
			return err
		}
		step := findStep(op, name)
		if step == nil {
			return ErrStepNotFound
		}
		if step.Status != StepRunning {
			return ErrStepSettled
		}
		return nil
	}
	return rt.m.checkedUpdate(rt.id, validate, func(op *Operation) {
		step := findStep(op, name)
		at := rt.m.deps.Now()
		step.Status = status
		step.Code = code
		// A module's own reason text never reaches the record: only a
		// reviewed message keyed by code does. See safeReason.
		step.Error = safeReason(code, reason)
		step.FinishedAt = &at
	})
}

func (rt *operationRuntime) InitNodes(nodes []NodeResult) error {
	// Built as a non-nil slice — even for a nil/empty nodes — and with every
	// timestamp pointer copied, before the operation is even looked up: a
	// caller that keeps and later mutates its slice, or a *time.Time inside
	// it, must never be able to reach back into stored state through it, and
	// an empty result must persist as "[]", not JSON null.
	cloned := make([]NodeResult, len(nodes))
	for i, n := range nodes {
		cloned[i] = n
		cloned[i].StartedAt = cloneTime(n.StartedAt)
		cloned[i].FinishedAt = cloneTime(n.FinishedAt)
	}
	return rt.m.checkedUpdate(rt.id, rt.m.rejectSettled, func(op *Operation) {
		op.Nodes = cloned
	})
}

// UpdateNode never runs mutate with the manager's lock held. It takes a
// snapshot of the named node, calls mutate against that private copy, and
// then commits with an optimistic compare-and-replace: if the stored node no
// longer matches the snapshot, something else changed it in between and this
// call fails with ErrConcurrentUpdate rather than overwriting it. This keeps
// three properties a lock-held callback could not: mutate may call back into
// this same Runtime (including Operation) without deadlocking, a panic
// inside mutate touches only a local copy and never the manager, and a
// concurrent writer's update is never silently lost.
func (rt *operationRuntime) UpdateNode(name string, mutate func(*NodeResult)) error {
	before, err := rt.m.snapshotNode(rt.id, name)
	if err != nil {
		return err
	}

	after := before
	mutate(&after)

	return rt.m.replaceNode(rt.id, name, before, after)
}

func (rt *operationRuntime) SetHostBootID(bootID string) error {
	return rt.m.checkedUpdate(rt.id, rt.m.rejectSettled, func(op *Operation) {
		op.HostBootID = bootID
	})
}

func (rt *operationRuntime) SetModuleState(raw json.RawMessage) error {
	// Checked, and rejected, before the manager is ever locked: an invalid
	// payload must not touch state at all, let alone be treated as (or
	// trigger a settlement resembling) a persistence failure.
	if len(raw) > 0 && !json.Valid(raw) {
		return ErrInvalidModuleState
	}
	// A nil or empty raw means "clear"; anything else is copied before the
	// manager is even locked, so a caller mutating its own slice afterward
	// cannot reach into persisted recovery evidence through it.
	var cp json.RawMessage
	if len(raw) > 0 {
		cp = append(json.RawMessage(nil), raw...)
	}
	return rt.m.checkedUpdate(rt.id, rt.m.rejectSettled, func(op *Operation) {
		op.ModuleState = cp
	})
}

func (rt *operationRuntime) SetCommandIssuedUntil(until time.Time) error {
	return rt.m.checkedUpdate(rt.id, rt.m.rejectSettled, func(op *Operation) {
		op.CommandIssuedUntil = until
	})
}

func (rt *operationRuntime) Complete(outcome Outcome) error {
	if !outcome.valid(rt.m.deps.Now()) {
		return ErrInvalidOutcome
	}
	return rt.m.complete(rt.id, outcome)
}

// findStep returns the most recently started step of this name, matching the
// existing finishStep/failStep search order: a name may recur across an
// operation's history, and the latest occurrence is the one still open.
func findStep(op *Operation, name string) *Step {
	for i := len(op.Steps) - 1; i >= 0; i-- {
		if op.Steps[i].Name == name {
			return &op.Steps[i]
		}
	}
	return nil
}

func findNode(op *Operation, name string) *NodeResult {
	for i := range op.Nodes {
		if op.Nodes[i].NodeName == name {
			return &op.Nodes[i]
		}
	}
	return nil
}

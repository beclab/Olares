package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Sentinel errors a Runtime mutation can return. They are the only way an
// OperationModule discovers that a mutation was rejected rather than
// applied: no method on the interface panics or silently no-ops.
var (
	// ErrOperationTerminal rejects a mutation against an operation that has
	// already reached a terminal status — including one the manager forced
	// terminal after a state_persistence_failed settlement. A module's own
	// goroutine keeps running after that happens, and every further
	// mutation must see it and stop rather than resurrect a settled record.
	ErrOperationTerminal = errors.New("cluster operation is already terminal")

	// ErrStepNotFound rejects finishing a step that was never started.
	ErrStepNotFound = errors.New("operation step not found")

	// ErrStepSettled rejects finishing a step that has already settled.
	ErrStepSettled = errors.New("operation step already settled")

	// ErrNodeNotFound rejects updating a node that was never recorded.
	ErrNodeNotFound = errors.New("operation node not found")

	// ErrInvalidOutcome rejects Complete with a status that is not one of
	// the outcomes a run may legally end on.
	ErrInvalidOutcome = errors.New("invalid operation outcome")
)

// errStatePersistenceFailed is returned by a checked mutation whose write to
// the store itself failed. The operation record already carries the stable,
// public CodeStatePersistenceFailed a caller can act on; this value only
// tells the module in the same call stack that this particular mutation did
// not take effect.
var errStatePersistenceFailed = errors.New("cluster operation state could not be recorded")

// errOperationNotFound guards a checked mutation issued for an operation id
// the manager no longer holds. newRuntime is only ever handed the id of an
// operation the manager just created or loaded, so this should not happen in
// practice; it exists so a checked mutation never mutates through a stale id.
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
	return rt.m.checkedUpdate(rt.id, rejectTerminal, func(op *Operation) {
		at := rt.m.deps.Now()
		op.Steps = append(op.Steps, Step{Name: name, Status: StepRunning, StartedAt: &at})
	})
}

func (rt *operationRuntime) FinishStep(name string, status StepStatus, code, reason string) error {
	validate := func(op *Operation) error {
		if op.Status.Terminal() {
			return ErrOperationTerminal
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
		step.Error = reason
		step.FinishedAt = &at
	})
}

func (rt *operationRuntime) InitNodes(nodes []NodeResult) error {
	// Copied before the operation is even looked up: a caller that keeps and
	// later mutates its slice must never be able to reach back into stored
	// state through it.
	cloned := append([]NodeResult(nil), nodes...)
	return rt.m.checkedUpdate(rt.id, rejectTerminal, func(op *Operation) {
		op.Nodes = cloned
	})
}

func (rt *operationRuntime) UpdateNode(name string, mutate func(*NodeResult)) error {
	validate := func(op *Operation) error {
		if op.Status.Terminal() {
			return ErrOperationTerminal
		}
		if findNode(op, name) == nil {
			return ErrNodeNotFound
		}
		return nil
	}
	return rt.m.checkedUpdate(rt.id, validate, func(op *Operation) {
		if node := findNode(op, name); node != nil {
			mutate(node)
		}
	})
}

func (rt *operationRuntime) SetHostBootID(bootID string) error {
	return rt.m.checkedUpdate(rt.id, rejectTerminal, func(op *Operation) {
		op.HostBootID = bootID
	})
}

func (rt *operationRuntime) SetModuleState(raw json.RawMessage) error {
	// Copied before the manager is even locked: SetModuleState must not let
	// a caller mutate persisted recovery evidence through a slice it kept.
	cp := append(json.RawMessage(nil), raw...)
	return rt.m.checkedUpdate(rt.id, rejectTerminal, func(op *Operation) {
		op.ModuleState = cp
	})
}

func (rt *operationRuntime) SetCommandIssuedUntil(until time.Time) error {
	return rt.m.checkedUpdate(rt.id, rejectTerminal, func(op *Operation) {
		op.CommandIssuedUntil = until
	})
}

func (rt *operationRuntime) Complete(outcome Outcome) error {
	if !outcome.valid() {
		return ErrInvalidOutcome
	}
	return rt.m.complete(rt.id, outcome)
}

// rejectTerminal is the validate function shared by every checked mutation
// that has no state to check beyond "the operation must still be moving".
func rejectTerminal(op *Operation) error {
	if op.Status.Terminal() {
		return ErrOperationTerminal
	}
	return nil
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

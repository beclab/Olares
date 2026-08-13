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

	// ErrRecoveryCannotExtendDeadline rejects a recovery Runtime naming a
	// future CommandIssuedUntil for a record whose own deadline has already
	// passed. rejectSettledDuringRecovery deliberately lets recovery touch
	// such a record however long ago its grace ran out — that is what lets
	// an old reboot still be confirmed — but confirming or releasing a
	// command is not evidence that the cluster should go back on hold for
	// it. Clearing the deadline (the zero value) is always allowed; only
	// naming a new one in the future, for a record already past its old
	// one, is refused.
	ErrRecoveryCannotExtendDeadline = errors.New("recovery cannot extend an expired command deadline")
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

	// settled is the rule this runtime's mutations are checked against. A
	// run and a recovery are allowed different things — see
	// Manager.rejectSettledDuringRecovery — and which one this is decided
	// when the runtime was built, not by whoever calls a method on it.
	settled settledCheck

	// recovery marks a runtime built by newRecoveryRuntime. It gates the one
	// restriction that is specific to recovery rather than to settled: see
	// SetCommandIssuedUntil.
	recovery bool
}

// newRuntime binds a Runtime to one operation for the lifetime of one Run
// call.
func newRuntime(m *Manager, id string, ctx context.Context) Runtime {
	return &operationRuntime{m: m, id: id, ctx: ctx, settled: m.rejectSettled}
}

// newRecoveryRuntime binds a Runtime to an operation whose command outlived
// the daemon that issued it. It differs from newRuntime in exactly three
// ways, all of which exist because the thing it settles was left
// outstanding: an operation still at command_issued stays mutable however
// long ago its grace deadline passed, the whole confirmation can be recorded
// as one write, and it may not use that latitude to name a new deadline for
// a record whose own deadline has already run out.
func newRecoveryRuntime(m *Manager, id string, ctx context.Context) recoveryRuntime {
	return &operationRuntime{m: m, id: id, ctx: ctx, settled: m.rejectSettledDuringRecovery, recovery: true}
}

// recoveryRuntime is what a RecoverableModule in this package is handed. The
// extra method is deliberately a description of a final state rather than
// access to anything: no manager, no store, no lock, and no callback that
// would run while the record is held.
type recoveryRuntime interface {
	Runtime

	// Settle records a stage, the nodes it was waiting on, and the outcome
	// the operation ends on, as one change. Either all of it is written or
	// none of it is, so nothing can read a confirmed stage on an operation
	// that still says its command is outstanding.
	Settle(settlement) error
}

// managedRuntime is the manager's own Runtime, as the modules built into
// this package need it. Sequencing a power operation takes side effects
// Runtime deliberately does not expose — the node directory, the fan-out,
// the cluster's own view of itself, this machine's power point — because a
// module has no business reaching past the operation it was handed. A
// built-in module asks for them by asserting this unexported interface, so
// what it can reach is still decided here rather than by whoever wrote the
// module.
type managedRuntime interface {
	Runtime
	managed() (*Manager, string)
}

func (rt *operationRuntime) managed() (*Manager, string) { return rt.m, rt.id }

// managerOf names the manager and the operation behind rt. ok is false for
// any Runtime this manager did not create, which is the answer a module has
// to act on rather than work around: nothing else can carry out a power
// operation.
func managerOf(rt Runtime) (m *Manager, id string, ok bool) {
	managed, ok := rt.(managedRuntime)
	if !ok {
		return nil, "", false
	}
	m, id = managed.managed()
	return m, id, m != nil
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
	return rt.m.checkedUpdate(rt.id, rt.settled, func(op *Operation) {
		at := rt.m.deps.Now()
		op.Steps = append(op.Steps, Step{Name: name, Status: StepRunning, StartedAt: &at})
	})
}

func (rt *operationRuntime) FinishStep(name string, status StepStatus, code, reason string) error {
	// Checked before anything else looks at it: a stage is read out of the
	// same record by the same callers as the operation, so what it may
	// carry is decided the same way. See safeCode.
	code = safeCode(code)
	validate := func(op *Operation) error {
		if err := rt.settled(op); err != nil {
			return err
		}
		step := findStep(op, name)
		if step == nil {
			return ErrStepNotFound
		}
		// A command_issued step is still open: it says a command was handed
		// out and its result is not known yet, which is exactly what a
		// RecoverableModule comes back to settle. Every other non-running
		// status is a stage that has already reported what it did.
		if step.Status != StepRunning && step.Status != StepCommandIssued {
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
		sanitizeNodeCode(&cloned[i])
	}
	return rt.m.checkedUpdate(rt.id, rt.settled, func(op *Operation) {
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
	before, err := rt.m.snapshotNode(rt.id, name, rt.settled)
	if err != nil {
		return err
	}

	// The copy mutate is handed is a deep one: a NodeResult keeps its times
	// behind pointers, and a callback writing a timestamp in place would
	// otherwise change the very snapshot the commit below compares against.
	after := cloneNode(before)
	mutate(&after)
	sanitizeNodeCode(&after)

	return rt.m.replaceNode(rt.id, name, before, after, rt.settled)
}

// sanitizeNodeCode holds a per-node result to the same rule the operation
// and its stages are held to: a code a caller could not act on is not kept.
// The message goes with it — a node reading module_failed must not carry a
// sentence written for the code that was refused.
func sanitizeNodeCode(n *NodeResult) {
	code := safeCode(n.Code)
	if code == n.Code {
		return
	}
	n.Code = code
	n.Error = reasonFor(code)
}

func (rt *operationRuntime) SetHostBootID(bootID string) error {
	return rt.m.checkedUpdate(rt.id, rt.settled, func(op *Operation) {
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
	return rt.m.checkedUpdate(rt.id, rt.settled, func(op *Operation) {
		op.ModuleState = cp
	})
}

func (rt *operationRuntime) SetCommandIssuedUntil(until time.Time) error {
	validate := func(op *Operation) error {
		if err := rt.settled(op); err != nil {
			return err
		}
		return rt.checkCommandIssuedUntilTransition(op.CommandIssuedUntil, until)
	}
	return rt.m.checkedUpdate(rt.id, validate, func(op *Operation) {
		op.CommandIssuedUntil = until
	})
}

// checkCommandIssuedUntilTransition is the one rule shared by every path
// that can write CommandIssuedUntil through a recovery runtime —
// SetCommandIssuedUntil, Complete, and Settle all call through here rather
// than each carrying its own copy of the check. It is a no-op for an
// ordinary run: this restriction exists only because recovery is allowed to
// touch a record however long its own grace deadline has passed, and that
// latitude must not become a way to re-arm the cluster's lock.
//
// A record whose CommandIssuedUntil is the zero value has never held a
// deadline at all — nothing to have "expired" — so recovery may still name
// one for the first time, the same way an ordinary run does; that is what
// lets a RecoverableModule hand a running record its own command_issued
// grace period during recovery. What is refused is naming a new future
// deadline for a record whose own, previously set, deadline has already run
// out: confirming or releasing such a command is not evidence the cluster
// should go back on hold for it. Clearing a deadline (next is the zero
// value) is always allowed, at any time, for the same reason
// rejectSettledDuringRecovery lets recovery touch an expired command_issued
// record in the first place — see confirmReboot.
func (rt *operationRuntime) checkCommandIssuedUntilTransition(current, next time.Time) error {
	if !rt.recovery {
		return nil
	}
	now := rt.m.deps.Now()
	if next.IsZero() || !next.After(now) {
		return nil
	}
	if current.IsZero() || current.After(now) {
		return nil
	}
	return ErrRecoveryCannotExtendDeadline
}

func (rt *operationRuntime) Complete(outcome Outcome) error {
	if !outcome.valid(rt.m.deps.Now()) {
		return ErrInvalidOutcome
	}
	validate := rt.settled
	if rt.recovery {
		validate = func(op *Operation) error {
			if err := rt.settled(op); err != nil {
				return err
			}
			return rt.checkCommandIssuedUntilTransition(op.CommandIssuedUntil, outcome.CommandIssuedUntil)
		}
	}
	return rt.m.complete(rt.id, outcome, validate)
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

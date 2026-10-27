package clusterop

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"k8s.io/klog/v2"
)

// startupRecoveryTimeout bounds the one recovery this daemon waits for while
// it is still starting: a record that was left mid-flight, handed to its
// module before the framework assumes the worst about it.
//
// It is short because nothing is being watched here. A module built into
// this package returns from that call immediately — the case that can take
// minutes, confirming a command that outlived this daemon, runs in the
// background instead (see Manager.resume). A module that does not return is
// a module holding up a daemon that has not yet served a single request, and
// this is what stops it.
const startupRecoveryTimeout = 10 * time.Second

// settleTerminated is the shape MarkInterrupted and settleUnknownModule
// share: both close out a record nothing is going to keep moving, the same
// way — every open step and pending node gets the same code, so a caller
// never sees a "failed" operation with a step still reading "running".
func settleTerminated(op *Operation, now time.Time, code, message string) {
	op.Status = StatusFailed
	op.Code = code
	op.Error = message
	op.UpdatedAt = now
	at := now
	op.FinishedAt = &at
	for i := range op.Steps {
		if op.Steps[i].Status == StepRunning || op.Steps[i].Status == StepPending {
			op.Steps[i].Status = StepFailed
			op.Steps[i].Code = code
			op.Steps[i].FinishedAt = &at
		}
	}
	for i := range op.Nodes {
		if op.Nodes[i].Status == NodePending {
			op.Nodes[i].Status = NodeFailed
			op.Nodes[i].Code = code
			op.Nodes[i].FinishedAt = &at
		}
	}
}

// MarkInterrupted settles an operation that was still moving when olaresd
// stopped watching it. Nothing observed how it ended, so it is not reported
// as anything but failed — and settling it is what stops it from holding
// the cluster's single-operation lock forever.
//
// It is the framework's default for a module with no recovery interface of
// its own, and for a RecoverableModule that was given a chance to look at
// the operation and did not settle it (see Manager.recoverLoadedOperations).
// It is exported so a caller outside this package can settle a record the
// same way, but nothing here calls it as anything but a fallback: it does
// not ask whether a module could have said more.
func MarkInterrupted(op *Operation, now time.Time) {
	settleTerminated(op, now, CodeDaemonRestarted, "olaresd restarted while this operation was in progress")
}

// settleUnknownModule is MarkInterrupted's counterpart for a type nothing in
// this build can carry out any more — a module removed since the record was
// written. There is no module left to ask, so nothing is retried and
// nothing is guessed at: the record is settled exactly the way a fresh
// request for that type is refused today, the same reviewed sentence and
// stable code. It stays on disk and queryable by id, like every other
// settled operation — only its module is gone.
func settleUnknownModule(op *Operation, now time.Time) {
	settleTerminated(op, now, CodeUnsupportedOperation, reasonFor(CodeUnsupportedOperation))
}

// recoverLoadedOperations decides, for every operation this daemon just
// loaded from disk, what may safely happen to it before this daemon starts
// serving requests. It must run after every loaded record already has an
// entry in m.ops/m.order/m.byRequest: a module's Recover mutates through
// the same Runtime any other caller uses, and a checked mutation looks its
// operation up there.
//
// It returns the id of every command_issued record whose module offers
// recovery — the one case that may have to wait on something this daemon
// cannot know without watching the cluster, such as rebootModule.Recover
// polling for up to Timeouts.Ready — so the caller can hand each to
// Manager.resume once every other loaded record has already been decided,
// rather than block construction on it.
func (m *Manager) recoverLoadedOperations() (unfinished []string) {
	for _, id := range m.order {
		// Read through the lock like every other reader: a Recover that ran
		// out of time below is still running, still holds a Runtime, and
		// commits through the same map this is reading.
		op, ok := m.Get(id)
		if !ok {
			continue
		}
		module, ok := m.registry.Lookup(op.Type)
		if !ok {
			// Nothing in this build can carry this type out any more. A
			// record that already reported something is left exactly as it
			// is; one still moving is settled rather than left to hold the
			// cluster's lock until this daemon restarts again.
			if !op.Status.Terminal() {
				now := m.deps.Now()
				m.updateAt(id, now, func(op *Operation) { settleUnknownModule(op, now) })
			}
			continue
		}
		recoverable, ok := module.(RecoverableModule)
		if !ok {
			if !op.Status.Terminal() {
				now := m.deps.Now()
				m.updateAt(id, now, func(op *Operation) { MarkInterrupted(op, now) })
			}
			continue
		}
		if op.Status == StatusCommandIssued {
			// This may run for as long as a module needs to watch the
			// cluster, so it is not started here; see Manager.resume.
			unfinished = append(unfinished, id)
			continue
		}
		if op.Status.Terminal() {
			// Already reported something, and not a command a module can
			// still confirm. Retained exactly as it is.
			continue
		}
		// The module gets to look at a run that stopped mid-flight before
		// this daemon assumes the worst about it, but only for as long as
		// a daemon that has not started serving yet can afford to wait.
		// Every module built into this package returns immediately for
		// anything other than a command_issued record it can prove
		// happened; one that does not return is not allowed to hold up
		// olaresd's start.
		m.recoverWithinStartupLimit(recoverable, id)
		if current, ok := m.Get(id); ok && !current.Status.Terminal() {
			// Recover looked and did not settle anything — the common case
			// for an interruption that is not a command_issued record it
			// can prove happened, and also what a Recover that ran out of
			// time leaves behind. Nothing else is going to settle it, so
			// this is the same fallback a module with no recovery
			// interface at all would have gotten. Once it lands the record
			// is failed, and a Recover still running is refused every
			// further mutation like any other caller holding a Runtime for
			// a settled operation.
			now := m.deps.Now()
			m.updateAt(id, now, func(op *Operation) { MarkInterrupted(op, now) })
		}
	}
	return unfinished
}

// recoverWithinStartupLimit runs one module's Recover while this daemon is
// still being built, and stops waiting for it after startupRecoveryTimeout.
//
// The call is not cancelled, because it cannot be: a module that ignores the
// context it was handed — or waits on something that will never answer — is
// exactly the case this bounds, and there is no way to take a goroutine back.
// What is bounded is how long the constructor waits, which is what decides
// whether olaresd starts. The goroutine left behind reaches the record only
// through its Runtime, so once the caller settles the record every mutation
// it attempts afterwards is refused the same way any other caller's would be.
func (m *Manager) recoverWithinStartupLimit(recoverable RecoverableModule, id string) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.safeRecover(m.deps.Base, recoverable, id)
	}()

	timer := time.NewTimer(m.deps.recoveryTimeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		klog.Errorf("clusterop: module recovery for operation %s did not return within %s; "+
			"settling it as interrupted and carrying on", id, m.deps.recoveryTimeout)
	}
}

// safeRecover calls a module's Recover inside a panic boundary: Recover is
// other people's code, running against a real operation record, and a panic
// in it must end that one goroutine rather than the daemon. The panic and
// its stack go to this node's log only; nothing here ever forwards them
// into a checked mutation that could write them into a persisted record.
//
// A panic leaves the operation exactly as Recover last committed it through
// its Runtime — never forced to any particular status by this wrapper —
// because a module that panics partway through a multi-step recovery may
// already have confirmed a stage or a node, and none of that is a change
// this rolls back. A command_issued record a panicking Recover never got to
// confirm is left exactly where offering no recovery at all leaves one: at
// command_issued, until its own grace deadline releases the cluster's lock.
// The one caller that runs synchronously during load,
// recoverLoadedOperations, still applies its own MarkInterrupted fallback
// afterward for a record a panic left non-terminal.
func (m *Manager) safeRecover(ctx context.Context, recoverable RecoverableModule, id string) {
	op, ok := m.Get(id)
	if !ok {
		return
	}
	rt := newRecoveryRuntime(m, id, ctx)
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module recovery for operation %s panicked: %v\n%s",
				id, r, debug.Stack())
		}
	}()
	recoverable.Recover(ctx, rt, op)
}

// safeRun calls a module's Run inside a panic boundary, the Run counterpart
// to safeRecover. A module that panics has stopped the same way one that
// returns no usable Outcome has: nothing is known about what it did beyond
// whatever it already committed through its own Runtime before the panic,
// and that is exactly what recoverFromRunPanic decides between. Any lock
// the module already released through its Runtime — a step, a node — stays
// released either way.
func safeRun(module OperationModule, ctx context.Context, rt Runtime, req RunRequest) (outcome Outcome) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked in Run: %v\n%s", module.Type(), r, debug.Stack())
			outcome = recoverFromRunPanic(rt)
		}
	}()
	return module.Run(ctx, rt, req)
}

// recoverFromRunPanic decides what a panicking Run leaves behind, once the
// panic itself has already been contained and logged by safeRun.
//
// A module that already moved the record away from running before it
// panicked — a command_issued handoff through Complete, a step or node
// closed through Settle, a terminal status reported through Complete — has
// committed real state, and none of that is this wrapper's to second-guess
// or overwrite: it reports the outcome as already recorded, so
// Manager.settle (the only caller of safeRun) leaves the record exactly as
// the module left it.
//
// A module that panicked before committing anything is still exactly where
// m.run left it: running. That case is settled failed with the same stable
// CodeModuleFailed an outcome the record cannot hold already settles on
// (see Manager.settle), and — because nothing else is ever going to — any
// step or node the module started but never finished is closed alongside
// it, the same way MarkInterrupted closes one for an interruption nothing
// observed the end of. Without that a caller could see an operation
// reported failed with a step still reading running.
func recoverFromRunPanic(rt Runtime) Outcome {
	m, id, ok := managerOf(rt)
	if !ok {
		// A Runtime this package did not build carries no way to check what,
		// if anything, was already committed through it. Best effort: the
		// same generic failure a foreign Runtime's own Complete is left to
		// accept or refuse, exactly as safeRun always returned before this
		// check existed.
		return failedWith(CodeModuleFailed, reasonFor(CodeModuleFailed))
	}
	// Decided and written in one turn of the manager's lock. Reading
	// "did the module already commit something?" and then writing the
	// settlement would leave a window: a goroutine the module left running
	// can commit a real command_issued handoff in between, and this would
	// overwrite a command that really went out with a generic failure.
	now := m.deps.Now()
	err := m.checkedUpdateAt(id, now, stillRunning, func(op *Operation) {
		settleTerminated(op, now, CodeModuleFailed, reasonFor(CodeModuleFailed))
	})
	if err != nil && !errors.Is(err, ErrOperationTerminal) && !errors.Is(err, errOperationNotFound) &&
		!errors.Is(err, errStatePersistenceFailed) {
		klog.Warningf("clusterop: settle a panicking run for operation %s: %v", id, err)
	}
	return Outcome{}.alreadyRecorded()
}

// stillRunning refuses a mutation for a record the module has already moved
// off running — a command_issued handoff, or a terminal status it reported
// itself. It is the validate half of the framework's own fallbacks, which
// have nothing to add to a record that already says what the module
// committed. ErrOperationTerminal is the answer because that is what every
// other refusal of a settled record returns; the caller does not act on it
// beyond leaving the record alone.
func stillRunning(op *Operation) error {
	if op.Status != StatusRunning {
		return ErrOperationTerminal
	}
	return nil
}

// moduleAlreadyCommitted reports whether the module moved this record away
// from running before it stopped — a command_issued handoff, or a terminal
// status it reported itself. Manager.settle asks it before forcing a
// settlement on a Run whose final outcome the record cannot hold, because it
// has nothing to add to a record that already says what the module
// committed, and overwriting a command_issued record would report a failure
// for a command that really was issued.
//
// It is a read, and the settlement it guards is a separate write. That is
// safe for the caller it has: by then the module's Run has returned, so the
// answer is not being changed underneath it by the same call stack. The
// other fallback — a Run that panicked, where a goroutine the module left
// behind may still be committing — does not read and then write; it validates
// stillRunning inside the same locked mutation. See recoverFromRunPanic.
//
// A record the manager no longer holds counts as committed: there is nothing
// left to settle.
func (m *Manager) moduleAlreadyCommitted(id string) bool {
	current, ok := m.Get(id)
	return !ok || current.Status != StatusRunning
}

// builtInPowerOperation marks an operation this package implements itself.
// Only reboot and shutdown do, and only this package can say so: the marker
// is an unexported method, and there is no exported type here holding one to
// embed, so a module written anywhere else cannot claim it.
//
// Two things follow from the marker and nothing else does:
//
//   - the legacy /command/power-node endpoint serves the operation, and only
//     that endpoint does. It predates module validation and hands a request
//     straight to a module without asking whether the module accepts it, so
//     it may only reach node code written here; and a power operation
//     reaching a machine any other way would be a machine powered outside
//     the sequence the master records. See ExecutePowerNode and ExecuteNode.
//   - the operation's errors reach a caller unchanged, because their code
//     and message are reviewed text from this package rather than whatever a
//     module chose to put in them. See ExecutePowerNode.
type builtInPowerOperation interface {
	OperationModule
	NodeOperationModule

	builtInPowerOperation()
}

// builtInPowerTypes is what the marker resolves to at request time. It is
// written only by registerBuiltInPowerOperation, from the two power modules'
// package initializers, and read only afterwards.
//
// The check is by type rather than by asking the module in hand, because the
// module in hand comes from whichever registry the caller passed: resolving
// it there would make the answer depend on what that registry happens to
// hold. The types themselves are decided here, once, and a later module
// cannot take one — DefaultRegistry refuses a second module for a type it
// already has.
var builtInPowerTypes = map[Type]struct{}{}

// registerBuiltInPowerOperation registers a power operation this package
// implements. It takes builtInPowerOperation rather than OperationModule so
// that only a module carrying the marker can get in, which is what makes the
// set below unforgeable from outside.
func registerBuiltInPowerOperation(module builtInPowerOperation) {
	MustRegisterModule(module)
	builtInPowerTypes[module.Type()] = struct{}{}
}

func isBuiltInPowerOperation(typ Type) bool {
	_, ok := builtInPowerTypes[typ]
	return ok
}

// ExecuteNode carries a node-scope command out through the module
// registered for req.Type. The generic node endpoint calls it, so the policy
// for an unsupported type, a module with no node capability, and a module
// that fails or panics while executing is decided exactly once, here.
//
// A module's own failure does not travel: a module chooses the code and the
// message of the error it returns, and both would reach an HTTP caller. It
// could pick a code that means something else here, or put an address or a
// token in a message somebody reads. What it said goes to this node's log,
// and what comes back is the one thing actually known — that the module
// failed.
//
// The built-in power operations are refused before they are reached at all,
// which makes the two endpoints serve disjoint sets. The master dispatches a
// reboot or a shutdown to /command/power-node, so nothing legitimate arrives
// here as one; what would is a machine — the control node included, since
// the generic endpoint answers for every role — powered off outside the
// sequence, the record and the single-operation lock that a cluster power
// operation is.
func ExecuteNode(ctx context.Context, registry *ModuleRegistry, req NodeRequest) error {
	if isBuiltInPowerOperation(req.Type) {
		return unsupportedNodeOperationError()
	}
	err, fromModule := runNodeModule(ctx, registry, req)
	if err == nil || !fromModule {
		return err
	}
	klog.Errorf("clusterop: module %s failed on this node: %v", req.Type, err)
	return moduleFailedNodeOperationError()
}

// ExecutePowerNode is ExecuteNode for the legacy /command/power-node
// endpoint, which an older worker serves and which a newer master still
// dispatches reboot and shutdown to.
//
// It refuses anything but a built-in power operation. That endpoint asks no
// module whether it accepts the request that arrived — it never has, and its
// request shape cannot grow one — so reaching any other module through it
// would be a way around the validation the generic endpoint performs.
//
// Both halves of "built-in" are checked. The type must be one of the two
// this package registered for itself, and the module the registry actually
// holds for it must be one this package wrote — it has to carry the
// unexported marker. The type alone would be enough only for as long as
// nothing else can be registered under that name: the module set this runs
// against is whichever one the caller passed, and a module registered under
// "reboot" in some other set would otherwise be handed a request nothing
// judged, and would power a machine outside the sequence the master records.
func ExecutePowerNode(ctx context.Context, registry *ModuleRegistry, req NodeRequest) error {
	if !isBuiltInPowerOperation(req.Type) || !registryHoldsBuiltInPower(registry, req.Type) {
		return unsupportedNodeOperationError()
	}
	err, _ := runNodeModule(ctx, registry, req)
	return err
}

// registryHoldsBuiltInPower reports whether the module registry holds for
// typ is a power operation this package implements itself. A registry that
// is nil, does not hold the type, or holds something else under it all
// answer the same way: there is nothing here the legacy endpoint may serve.
func registryHoldsBuiltInPower(registry *ModuleRegistry, typ Type) bool {
	if registry == nil {
		return false
	}
	module, ok := registry.Lookup(typ)
	if !ok {
		return false
	}
	if _, ok := module.(builtInPowerOperation); !ok {
		klog.Errorf("clusterop: %q is registered to a module this daemon did not write; "+
			"the power endpoint serves only the operations it implements itself", typ)
		return false
	}
	return true
}

// runNodeModule finds the module for req.Type and runs it. fromModule
// reports whether err is the module's own answer rather than this package's
// refusal to reach it at all, which is what decides whether the error may be
// repeated to a caller.
func runNodeModule(ctx context.Context, registry *ModuleRegistry,
	req NodeRequest) (err error, fromModule bool) {
	// A nil registry answers no differently than one that just does not
	// hold req.Type: there is nothing here that can act on a single node
	// for it either way. Checked before Lookup, which is a pointer method
	// that would otherwise panic on a nil receiver before the recover below
	// is even in place to catch it.
	if registry == nil {
		return unsupportedNodeOperationError(), false
	}
	module, ok := registry.Lookup(req.Type)
	if !ok {
		return unsupportedNodeOperationError(), false
	}
	nodeModule, ok := module.(NodeOperationModule)
	if !ok {
		return unsupportedNodeOperationError(), false
	}
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked in ExecuteNode: %v\n%s", req.Type, r, debug.Stack())
			err, fromModule = moduleFailedNodeOperationError(), false
		}
	}()
	return nodeModule.ExecuteNode(ctx, req), true
}

// unsupportedNodeOperationError is what ExecuteNode returns for a type this
// build does not know, or a module that knows the type but declares no node
// capability. It is the same *PowerError shape, code and reviewed sentence
// PowerHost already refuses an unknown operation with, so a caller mapping
// one error type maps both.
func unsupportedNodeOperationError() error {
	return &PowerError{Code: CodeUnsupportedOperation, Message: reasonFor(CodeUnsupportedOperation)}
}

// moduleFailedNodeOperationError is what a panicking ExecuteNode settles on:
// nothing is known about what it did, so this reuses the one stable code
// the rest of this package already uses for that.
func moduleFailedNodeOperationError() error {
	return &PowerError{Code: CodeModuleFailed, Message: reasonFor(CodeModuleFailed)}
}

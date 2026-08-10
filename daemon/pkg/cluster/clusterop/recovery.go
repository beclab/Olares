package clusterop

import (
	"context"
	"runtime/debug"
	"time"

	"k8s.io/klog/v2"
)

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
		op, ok := m.ops[id]
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
				m.update(id, func(op *Operation) { settleUnknownModule(op, m.deps.Now()) })
			}
			continue
		}
		recoverable, ok := module.(RecoverableModule)
		if !ok {
			if !op.Status.Terminal() {
				m.update(id, func(op *Operation) { MarkInterrupted(op, m.deps.Now()) })
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
		// this daemon assumes the worst about it. Every module built into
		// this package returns immediately for anything other than a
		// command_issued record it can prove happened, so this is not
		// expected to block construction; a module written otherwise is a
		// bug in that module, the same way a Store call that never
		// returned would be a bug in the store.
		m.safeRecover(m.deps.Base, recoverable, id)
		if current, ok := m.ops[id]; ok && !current.Status.Terminal() {
			// Recover looked and did not settle anything — the common case
			// for an interruption that is not a command_issued record it
			// can prove happened. Nothing else is going to, so this is the
			// same fallback a module with no recovery interface at all
			// would have gotten.
			m.update(id, func(op *Operation) { MarkInterrupted(op, m.deps.Now()) })
		}
	}
	return unfinished
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
// returns no usable Outcome has: nothing is known about what it did, and
// leaving the record running would hold the cluster's single-operation
// lock until this daemon restarts. CodeModuleFailed is the same stable code
// an outcome the record cannot hold already settles on (see Manager.settle
// in orchestrate.go), so a caller sees one framework-owned failure mode
// either way, never the panic's own text. Any lock the module already
// released through its Runtime — a step, a node — stays released; only
// Complete has not been called, and settle (the only caller of safeRun)
// calls it next with the outcome this returns.
func safeRun(module OperationModule, ctx context.Context, rt Runtime, req RunRequest) (outcome Outcome) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked in Run: %v\n%s", module.Type(), r, debug.Stack())
			outcome = failedWith(CodeModuleFailed, reasonFor(CodeModuleFailed))
		}
	}()
	return module.Run(ctx, rt, req)
}

// ExecuteNode carries a node-scope command out through the module
// registered for req.Type. Both node-execution HTTP handlers (task 6) call
// this instead of a module's own ExecuteNode directly, so the policy for an
// unsupported type, a module with no node capability, and a module that
// panics while executing is decided exactly once, here.
func ExecuteNode(ctx context.Context, registry *ModuleRegistry, req NodeRequest) (err error) {
	module, ok := registry.Lookup(req.Type)
	if !ok {
		return unsupportedNodeOperationError()
	}
	nodeModule, ok := module.(NodeOperationModule)
	if !ok {
		return unsupportedNodeOperationError()
	}
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked in ExecuteNode: %v\n%s", req.Type, r, debug.Stack())
			err = moduleFailedNodeOperationError()
		}
	}()
	return nodeModule.ExecuteNode(ctx, req)
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

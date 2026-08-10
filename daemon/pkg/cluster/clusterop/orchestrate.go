package clusterop

import (
	"context"
	"errors"

	"k8s.io/klog/v2"
)

// run carries one operation out through the module registered for its type.
//
// Nothing here knows what any operation does. The module is looked up first,
// because a type nothing can carry out has to settle the record rather than
// leave it moving and holding the cluster's single-operation lock; after that
// the module is handed a Runtime bound to this operation and the outcome it
// returns is what the record ends on.
func (m *Manager) run(ctx context.Context, id string, opType Type, req RunRequest) {
	rt := newRuntime(m, id, ctx)

	module, ok := m.registry.Lookup(opType)
	if !ok {
		m.settle(rt, id, Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "this daemon does not perform that operation",
		})
		return
	}

	if !m.update(id, func(op *Operation) {
		op.Status = StatusRunning
		at := m.deps.Now()
		op.StartedAt = &at
	}) {
		return
	}

	m.settle(rt, id, safeRun(module, ctx, rt, req))
}

// settle records the outcome a module ended on.
//
// A module that already recorded its own outcome says so, and this leaves the
// record alone rather than writing it again and reading the refusal as
// agreement.
//
// An outcome the record cannot hold is the one case that must not be passed
// on: the module has stopped, and an operation left running holds the
// cluster's single-operation lock until the daemon restarts. It is settled as
// failed, which is all that is known.
func (m *Manager) settle(rt Runtime, id string, outcome Outcome) {
	if outcome.recorded {
		return
	}
	err := rt.Complete(outcome)
	if err == nil {
		return
	}
	if errors.Is(err, ErrInvalidOutcome) {
		klog.Warningf("clusterop: operation %s reported an unusable outcome %+v", id, outcome)
		err = rt.Complete(failedWith(CodeModuleFailed, reasonFor(CodeModuleFailed)))
		if err == nil {
			return
		}
	}
	if m.settlementFailedToPersist(id, err) {
		// applyLocked already logged this at error level the moment the
		// write itself failed; warning about it again here would read as a
		// second, unrelated problem rather than the expected fallout of the
		// first.
		return
	}
	klog.Warningf("clusterop: settle operation %s: %v", id, err)
}

// settlementFailedToPersist reports whether err is Complete refusing this
// settlement because the record's state could no longer be recorded, rather
// than for any ordinary reason a run can already be terminal. errStatePersistenceFailed
// is this call's own write failing; ErrOperationTerminal alongside
// m.persistFailed[id] is a previous call's write having already failed and
// forced the record there first.
func (m *Manager) settlementFailedToPersist(id string, err error) bool {
	if errors.Is(err, errStatePersistenceFailed) {
		return true
	}
	if !errors.Is(err, ErrOperationTerminal) {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.persistFailed[id]
}

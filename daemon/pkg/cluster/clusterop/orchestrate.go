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

	m.settle(rt, id, module.Run(ctx, rt, req))
}

// settle records the outcome a module ended on.
//
// ErrOperationTerminal is not a failure here. A module may settle its own
// operation and keep going — a node shutdown holds the record at
// command_issued until the machine stops answering — and it hands back the
// outcome it settled on so that a manager which has not recorded one yet
// still can. Whichever of the two wrote it, the record is not written twice.
func (m *Manager) settle(rt Runtime, id string, outcome Outcome) {
	if err := rt.Complete(outcome); err != nil && !errors.Is(err, ErrOperationTerminal) {
		klog.Warningf("clusterop: settle operation %s: %v", id, err)
	}
}

package clusterop

import (
	"time"

	"k8s.io/klog/v2"
)

// settlement is an operation's last state, described in one piece: the stage
// that was still open, the nodes that stage was waiting on, and the outcome
// the record ends on.
//
// It exists because a confirmation is not three facts, it is one. A reboot
// confirmed by the next daemon has to move the stage, the control node and
// the operation together — a reader that catches the record in between is
// told that the master's power command succeeded on an operation whose own
// status still says the command is outstanding, which is not a state the
// cluster was ever in.
//
// It is deliberately data. Nothing in it runs, so nothing a module wrote can
// execute while the manager is locked, and it names things by value — a step
// name, a node name, the status each should move to — so a module never holds
// the manager, the store, or any part of the record itself.
type settlement struct {
	// step is the stage to close, if there is one.
	step stepSettlement

	// nodes are the nodes to move. Each says what it expects to find, so a
	// settlement built from a stale reading of the record changes nothing.
	nodes []nodeSettlement

	// outcome is what the operation itself ends on.
	outcome Outcome
}

type stepSettlement struct {
	// name is empty when the settlement closes no stage.
	name   string
	status StepStatus
}

type nodeSettlement struct {
	name string

	// from is the status this node must still be in. A node that has moved
	// on since the settlement was built is left alone rather than
	// overwritten.
	from NodeStatus

	// to is what it becomes.
	to NodeStatus
}

// Settle records the whole settlement as one change: one validation, one
// mutation, one write. See Manager.settleAtomically for what happens if that
// write fails.
func (rt *operationRuntime) Settle(s settlement) error {
	if !s.outcome.valid(rt.m.deps.Now()) {
		return ErrInvalidOutcome
	}
	return rt.m.settleAtomically(rt.id, s, rt.settled)
}

// settleAtomically applies s to a copy of the record, writes the copy, and
// only then lets anyone see it. That order is the whole point: if the write
// fails, no part of the settlement ever became visible, so a confirmation
// that could not be recorded does not leave a half-confirmed operation
// behind. A failed write settles the record the same way every other failed
// write does — state_persistence_failed, and no further mutation accepted.
func (m *Manager) settleAtomically(id string, s settlement, settled settledCheck) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, ok := m.ops[id]
	if !ok {
		return errOperationNotFound
	}
	if m.persistFailed[id] {
		return ErrOperationTerminal
	}
	if err := settled(op); err != nil {
		return err
	}
	if s.step.name != "" {
		step := findStep(op, s.step.name)
		if step == nil {
			return ErrStepNotFound
		}
		if step.Status != StepRunning && step.Status != StepCommandIssued {
			return ErrStepSettled
		}
	}

	at := m.deps.Now()
	next := op.Clone()
	if s.step.name != "" {
		step := findStep(&next, s.step.name)
		step.Status = s.step.status
		step.FinishedAt = &at
	}
	for _, want := range s.nodes {
		node := findNode(&next, want.name)
		// A node that is not there, or that has already moved on, is not
		// something to force: the settlement describes what this operation
		// was waiting on, and anything else was decided by whoever wrote it.
		if node == nil || node.Status != want.from {
			continue
		}
		node.Status = want.to
		node.FinishedAt = &at
	}
	next.Status = s.outcome.Status
	next.Code = s.outcome.Code
	next.Error = s.outcome.persistedReason()
	next.CommandIssuedUntil = s.outcome.CommandIssuedUntil
	next.UpdatedAt = at
	// The moment the outcome was established, not the moment the command
	// went out: a record promoted from command_issued already carries the
	// latter, and it is not when the operation finished.
	next.FinishedAt = &at

	if err := m.deps.Store.Save(next); err != nil {
		klog.Errorf("clusterop: persist operation %s: %v", id, err)
		m.persistFailed[id] = true
		m.forceStatePersistenceFailedLocked(op, at)
		return errStatePersistenceFailed
	}

	m.ops[id] = &next
	if m.activeID == id && !m.operationActive(&next, at) {
		m.activeID = ""
	}
	return nil
}

// forceStatePersistenceFailedLocked is the settlement a record gets when its
// own state could not be written. It says only that, and nothing about what
// the operation was in the middle of doing, because that is the one thing no
// longer knowable from here.
func (m *Manager) forceStatePersistenceFailedLocked(op *Operation, at time.Time) {
	op.Status = StatusFailed
	op.Code = CodeStatePersistenceFailed
	op.Error = reasonFor(CodeStatePersistenceFailed)
	op.UpdatedAt = at
	op.FinishedAt = &at
	m.activeID = op.ID
}

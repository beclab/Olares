package clusterop

import (
	"fmt"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// Record-keeping for an upgrade: how a stage's progress is written into the
// operation, and how the operation is read back to decide what still needs
// doing. It is separate from the orchestration in upgrade.go because it is a
// separate concern — everything here is about the record, and nothing here
// talks to a node.
//
// The record is what makes an upgrade resumable. A run that restarts reads it
// to find where it got to, so these functions are also the definition of what
// "already done" means.

// stageSettled reports whether this operation already carried a stage to a
// conclusion that does not need repeating.
func (m *Manager) stageSettled(id, stageName string) bool {
	op, ok := m.Get(id)
	if !ok {
		return false
	}
	step := lastStepNamed(&op, stageName)
	if step == nil {
		return false
	}
	return step.Status == StepSucceeded || step.Status == StepSkipped
}

func (m *Manager) settleFromState(id, stageName, nodeName string, state UpgradeStageState) bool {
	if state.Phase == UpgradeStagePhaseSucceeded {
		// The caller still has to verify installed versions or a pending reboot.
		return true
	}
	code := state.Code
	if code == "" {
		code = CodeStageFailed
	}
	reason := reasonFor(code)
	if code == CodeStageFailed && state.Error != "" {
		if detail := commands.CommandErrorSummary(state.Error); detail != "" {
			reason = detail
		}
	}
	m.settleStageNode(id, stageName, nodeName, NodeFailed, code, reason)
	return false
}

// beginStage records a stage before it starts, with every node it will touch.
// Written first so that a daemon that dies mid-stage leaves behind which nodes
// were involved rather than an empty step.
func (m *Manager) beginStage(id string, stage UpgradeStage, targets []inventory.Node) []inventory.Node {
	var pending []inventory.Node
	m.update(id, func(op *Operation) {
		at := m.deps.Now()
		existing := lastStepNamed(op, stage.Name)
		completed := make(map[string]NodeResult)
		// Older daemons recorded success before confirming the boot ID or
		// binary versions. Recheck such an interrupted stage on recovery.
		if existing != nil && (existing.Status != StepRunning || (!stage.AwaitRestart && stage.Name != StageBootstrap)) {
			for _, node := range existing.Nodes {
				if node.Status == NodeSucceeded {
					completed[node.NodeName] = node
				}
			}
		}
		nodes := make([]NodeResult, 0, len(targets))
		for _, t := range targets {
			if done, ok := completed[t.NodeName]; ok {
				nodes = append(nodes, done)
				continue
			}
			nodes = append(nodes, NodeResult{NodeName: t.NodeName, Role: t.Role, Status: NodePending})
			pending = append(pending, t)
		}
		fresh := Step{
			Name:        stage.Name,
			Status:      StepRunning,
			Placement:   string(stage.Placement),
			MaxParallel: stage.MaxParallel,
			StartedAt:   &at,
			Nodes:       nodes,
		}
		// A stage a previous attempt left behind is overwritten rather than
		// joined by a second one; see lastStepNamed.
		if existing != nil {
			*existing = fresh
			return
		}
		op.Steps = append(op.Steps, fresh)
	})
	return pending
}

func (m *Manager) recordSkippedStage(id string, stage UpgradeStage, why string) {
	m.update(id, func(op *Operation) {
		at := m.deps.Now()
		fresh := Step{
			Name:        stage.Name,
			Status:      StepSkipped,
			Placement:   string(stage.Placement),
			MaxParallel: stage.MaxParallel,
			StartedAt:   &at,
			FinishedAt:  &at,
			Error:       why,
		}
		if existing := lastStepNamed(op, stage.Name); existing != nil {
			*existing = fresh
			return
		}
		op.Steps = append(op.Steps, fresh)
	})
}

func (m *Manager) setStageNode(id, stageName, nodeName string, fn func(*NodeResult)) {
	m.update(id, func(op *Operation) {
		// The latest attempt at this stage, which is the one running; see
		// lastStepNamed.
		step := lastStepNamed(op, stageName)
		if step == nil {
			return
		}
		for j := range step.Nodes {
			if step.Nodes[j].NodeName == nodeName {
				fn(&step.Nodes[j])
				return
			}
		}
	})
}

func (m *Manager) settleStageNode(id, stageName, nodeName string, status NodeStatus, code, reason string) {
	m.setStageNode(id, stageName, nodeName, func(n *NodeResult) {
		at := m.deps.Now()
		n.Status = status
		n.Code = code
		n.Error = reason
		n.FinishedAt = &at
	})
}

func (m *Manager) skipRemainingStageNodes(id, stageName string) {
	m.update(id, func(op *Operation) {
		step := lastStepNamed(op, stageName)
		if step == nil {
			return
		}
		at := m.deps.Now()
		for j := range step.Nodes {
			if step.Nodes[j].Status == NodePending {
				step.Nodes[j].Status = NodeSkipped
				step.Nodes[j].Error = "the stage stopped before this node was reached"
				step.Nodes[j].FinishedAt = &at
			}
		}
	})
}

// failUpgradeAt ends an upgrade that broke in the middle.
//
// Before notify-start, a failure is reported as failed even if preparation
// stages succeeded. After commit, a prior successful plan stage counts toward
// a partially_failed upgrade.
//
// There is no rollback: Olares has no downgrade path, so the honest report is
// which stages ran, and the repair is to fix the node and run the upgrade
// again — the stages that succeeded will find their work already done.
func (m *Manager) failUpgradeAt(id string, up UpgradePlan, stage UpgradeStage) Outcome {
	status := StatusFailed
	reason := reasonFor(CodeStageFailed)
	if op, ok := m.Get(id); ok {
		for _, s := range op.Steps {
			if op.Committed && s.Name != stage.Name && s.Status == StepSucceeded && up.HasStage(s.Name) {
				status = StatusPartiallyFailed
				break
			}
		}
		if failed := lastStepNamed(&op, stage.Name); failed != nil {
			for _, node := range failed.Nodes {
				if node.Status == NodeFailed {
					reason = fmt.Sprintf("stage %s on node %s: %s", stage.Name, node.NodeName,
						commands.CommandErrorSummary(node.Error))
					break
				}
			}
		}
	}
	// Retain the failing node and its bounded diagnostic in the watcher's error.
	return settledWith(status, CodeStageFailed, reason)
}

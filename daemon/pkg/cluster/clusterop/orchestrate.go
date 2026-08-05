package clusterop

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"k8s.io/klog/v2"
)

// plan is the node layout one operation will act on: every compute node first,
// the control node last.
type plan struct {
	master  inventory.Node
	workers []inventory.Node

	// baseline is the boot each node was on before the operation began. Only
	// a reboot has one; see Manager.baseline.
	baseline map[string]inventory.Observation
}

// errPrecheckFailed ends a run before anything was dispatched. The reasons are
// already on the operation record, per node.
var errPrecheckFailed = errors.New("precheck failed")

func (m *Manager) run(ctx context.Context, id string, opType Type, creds Credentials) {
	if !m.update(id, func(op *Operation) {
		op.Status = StatusRunning
		at := m.deps.Now()
		op.StartedAt = &at
	}) {
		return
	}

	op, ok := m.Get(id)
	if !ok {
		return
	}
	if op.Scope == ScopeNode {
		m.runNode(ctx, id, opType, op.Target, creds)
		return
	}

	p, err := m.precheck(ctx, id, opType, creds)
	if err != nil {
		klog.Warningf("clusterop: %s %s stopped at the precheck: %v", opType, id, err)
		return
	}

	switch opType {
	case TypeReboot:
		m.runReboot(ctx, id, p, creds)
	case TypeShutdown:
		m.runShutdown(ctx, id, p, creds)
	}
}

func (m *Manager) runNode(ctx context.Context, id string, opType Type, target string, creds Credentials) {
	m.startStep(id, StepPrecheck)
	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for the precheck", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		m.fail(id, StatusFailed, CodeInventoryUnavailable, reason)
		return
	}
	var node inventory.Node
	for _, candidate := range nodes {
		if candidate.NodeName == target {
			node = candidate
			break
		}
	}
	if node.NodeName == "" {
		m.failStep(id, StepPrecheck, CodeNodeIdentityUnknown, "the node directory could not identify this node")
		m.fail(id, StatusFailed, CodeNodeIdentityUnknown, "the node directory could not identify this node")
		return
	}
	m.update(id, func(op *Operation) {
		op.Nodes = []NodeResult{{NodeName: node.NodeName, Role: node.Role, Status: NodePending}}
	})

	if node.Role == inventory.RoleMaster {
		if opType != TypeReboot {
			m.failStep(id, StepPrecheck, CodePowerUnsupported, "the control node cannot be shut down by a node operation")
			m.fail(id, StatusFailed, CodePowerUnsupported, "the control node cannot be shut down by a node operation")
			return
		}
		if code, reason := m.checkControlNode(opType); code != "" {
			m.failStep(id, StepPrecheck, code, reason)
			m.fail(id, StatusFailed, code, reason)
			return
		}
		m.finishStep(id, StepPrecheck, StepSucceeded, "", "")
		m.powerMaster(ctx, id, plan{master: node}, opType, false)
		return
	}

	baseline, err := m.baseline(ctx, opType)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the reboot baseline", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		m.fail(id, StatusFailed, CodeInventoryUnavailable, reason)
		return
	}
	required := requiredCapability(opType)
	if code, reason := m.checkWorker(ctx, node, opType, required, baseline, creds); code != "" {
		m.setNode(id, node.NodeName, func(n *NodeResult) {
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
		})
		m.failStep(id, StepPrecheck, code, reason)
		m.fail(id, StatusFailed, code, reason)
		return
	}
	m.finishStep(id, StepPrecheck, StepSucceeded, "", "")
	issued, ok := m.commandWorkers(ctx, id, plan{workers: []inventory.Node{node}, baseline: baseline}, opType, creds)
	if !ok {
		return
	}
	if opType == TypeShutdown {
		m.update(id, func(op *Operation) { op.Status = StatusCommandIssued })
		return
	}
	m.startStep(id, StepWorkerRestart)
	if m.awaitRestarts(ctx, id, issued, baseline) > 0 {
		m.failStep(id, StepWorkerRestart, CodeWorkerRestartFailed, "the node did not come back")
		m.fail(id, StatusFailed, CodeWorkerRestartFailed, "the node did not come back")
		return
	}
	m.finishStep(id, StepWorkerRestart, StepSucceeded, "", "")
	m.update(id, func(op *Operation) { op.Status = StatusSucceeded })
}

// precheck refuses the whole operation unless every compute node can be told
// to do what was asked. There is no force: a shutdown that skipped a node the
// master could not reach would leave that node running while the user is told
// the cluster is off.
func (m *Manager) precheck(ctx context.Context, id string, opType Type, creds Credentials) (plan, error) {
	m.startStep(id, StepPrecheck)

	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for the precheck", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		m.fail(id, StatusFailed, CodeInventoryUnavailable, reason)
		return plan{}, errPrecheckFailed
	}

	// The topology gate runs before anything is inspected: a cluster this
	// daemon cannot sequence must not have a single node dialled by a run it
	// is going to refuse anyway.
	p, code, err := splitCluster(nodes)
	if err != nil {
		m.failStep(id, StepPrecheck, code, err.Error())
		m.fail(id, StatusFailed, code, err.Error())
		return plan{}, errPrecheckFailed
	}

	m.update(id, func(op *Operation) {
		op.Nodes = make([]NodeResult, 0, len(p.workers)+1)
		for _, w := range p.workers {
			op.Nodes = append(op.Nodes, NodeResult{NodeName: w.NodeName, Role: w.Role, Status: NodePending})
		}
		op.Nodes = append(op.Nodes, NodeResult{NodeName: p.master.NodeName, Role: p.master.Role, Status: NodePending})
	})

	// The control node is checked before any other node is dialled. It acts
	// last, so a refusal found later is found with every compute node already
	// powered off and nothing left able to reach them.
	if code, reason := m.checkControlNode(opType); code != "" {
		m.setNode(id, p.master.NodeName, func(n *NodeResult) {
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
			at := m.deps.Now()
			n.FinishedAt = &at
		})
		m.skipRemainingNodes(id, "blocked by the precheck")
		m.failStep(id, StepPrecheck, CodePrecheckFailed, "the control node cannot perform this operation")
		m.fail(id, StatusFailed, CodePrecheckFailed, "the control node cannot perform this operation")
		return plan{}, errPrecheckFailed
	}

	// A reboot is only provable against the boot each node is on now, so the
	// baseline is read here, before anything is told to go down.
	baseline, err := m.baseline(ctx, opType)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the reboot baseline", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		m.fail(id, StatusFailed, CodeInventoryUnavailable, reason)
		return plan{}, errPrecheckFailed
	}
	p.baseline = baseline

	required := requiredCapability(opType)
	var blocked bool
	for _, w := range p.workers {
		code, reason := m.checkWorker(ctx, w, opType, required, baseline, creds)
		if code == "" {
			continue
		}
		blocked = true
		m.setNode(id, w.NodeName, func(n *NodeResult) {
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
			at := m.deps.Now()
			n.FinishedAt = &at
		})
	}

	if blocked {
		// Nothing has been dispatched, so every other node is untouched
		// rather than merely unfinished.
		m.skipRemainingNodes(id, "blocked by the precheck")
		m.failStep(id, StepPrecheck, CodePrecheckFailed, "one or more nodes cannot perform this operation")
		m.fail(id, StatusFailed, CodePrecheckFailed, "one or more nodes cannot perform this operation")
		return plan{}, errPrecheckFailed
	}

	m.finishStep(id, StepPrecheck, StepSucceeded, "", "")
	return p, nil
}

// checkControlNode asks this machine's own execution point whether it would
// carry the operation out.
//
// It deliberately does not read the control node's declared capabilities. A
// control node does not offer power.shutdown to the cluster — shutting it down
// alone is not a thing to offer, it is the last step of a cluster operation —
// so treating that silence as a refusal would block every cluster shutdown.
// What answers here is the same check PowerHost performs immediately before it
// runs the command, so the precheck and the execution point cannot disagree.
func (m *Manager) checkControlNode(opType Type) (code, reason string) {
	err := m.deps.LocalPowerSupport(opType)
	if err == nil {
		return "", ""
	}
	return powerReason(CodePowerUnsupported, "ask this machine whether it can power itself", err)
}

// checkWorker answers whether one compute node can be told to do this. The
// control node is deliberately not checked the same way: it never declares a
// single-node shutdown, because turning it off is the last step of a cluster
// operation rather than something to offer on its own.
//
// The order of the checks is the order in which the answers are useful. A node
// the directory could not describe has made no claim about what it supports,
// and a node the cluster already considers down explains every later failure,
// so both are reported before anything is dialled.
func (m *Manager) checkWorker(ctx context.Context, w inventory.Node, opType Type, required string,
	baseline map[string]inventory.Observation, creds Credentials) (code, reason string) {
	if w.NodeName == "" || w.Role == "" || w.Role == inventory.RoleUnknown {
		return CodeNodeIdentityUnknown, "the node directory could not identify this node"
	}
	if !w.Ready {
		return CodeNodeNotReady, "the cluster does not consider this node ready"
	}
	if w.IP == "" {
		return CodeNodeUnaddressable, "node has no internal address"
	}
	if opType == TypeReboot {
		if obs, ok := baseline[w.NodeName]; !ok || obs.BootID == "" {
			return CodeBootIDUnavailable, "the cluster does not report which boot this node is on"
		}
	}
	status, err := m.deps.Inspect(ctx, w, creds)
	if err != nil {
		return CodeNodeUnreachable, suppress(CodeNodeUnreachable, "read node "+w.NodeName+" for the precheck", err)
	}
	if c, ok := status.Capabilities[required]; !ok || !c.Supported {
		return CodePowerUnsupported, fmt.Sprintf("node does not support %s", required)
	}
	return "", ""
}

// baseline records which boot every node is on before any of them is told to
// leave it. A shutdown proves nothing about a restart and needs none.
func (m *Manager) baseline(ctx context.Context, opType Type) (map[string]inventory.Observation, error) {
	if opType != TypeReboot {
		return nil, nil
	}
	return m.deps.Observe(ctx)
}

func requiredCapability(t Type) string {
	if t == TypeShutdown {
		return nodestatus.CapPowerShutdown
	}
	return nodestatus.CapPowerReboot
}

// splitCluster names the machine this daemon runs on as the control node and
// treats everything else as a compute node to be powered first.
//
// Exactly one control node, and it has to be this machine. There is no
// fallback to the first master in the list: a cluster power operation is
// sequenced by the node that goes last, so a daemon that guessed wrong would
// be orchestrating from a machine it had already scheduled for shutdown, and
// on a two-master cluster the second control node is powered off by the first
// one's run with nothing left to record what became of it.
func splitCluster(nodes []inventory.Node) (plan, string, error) {
	var masters []int
	for i, n := range nodes {
		if n.Role == inventory.RoleMaster {
			masters = append(masters, i)
		}
	}
	if len(masters) != 1 {
		return plan{}, CodeUnsupportedTopology, fmt.Errorf(
			"this daemon powers a cluster with exactly one control node, found %d", len(masters))
	}

	var selves []int
	for i, n := range nodes {
		if n.IsSelf {
			selves = append(selves, i)
		}
	}
	if len(selves) != 1 {
		return plan{}, CodeSelfUnresolved, fmt.Errorf(
			"the node directory identifies %d nodes as this machine", len(selves))
	}
	if selves[0] != masters[0] {
		return plan{}, CodeSelfUnresolved, errors.New("this machine is not the cluster's control node")
	}

	var p plan
	p.master = nodes[masters[0]]
	if p.master.NodeName == "" {
		return plan{}, CodeSelfUnresolved, errors.New("the control node has no name in the node directory")
	}
	for i, n := range nodes {
		if i != masters[0] {
			p.workers = append(p.workers, n)
		}
	}
	return p, "", nil
}

func (m *Manager) runReboot(ctx context.Context, id string, p plan, creds Credentials) {
	issued, ok := m.commandWorkers(ctx, id, p, TypeReboot, creds)
	if !ok {
		return
	}

	if len(issued) > 0 {
		m.startStep(id, StepWorkerRestart)
		failures := m.awaitRestarts(ctx, id, issued, p.baseline)
		if failures > 0 {
			m.failStep(id, StepWorkerRestart, CodeWorkerRestartFailed, "one or more nodes did not come back")
			m.stopBeforeMaster(id, len(issued) > failures, CodeWorkerRestartFailed,
				"the control node was left running because a compute node did not come back")
			return
		}
		m.finishStep(id, StepWorkerRestart, StepSucceeded, "", "")
	}

	m.powerMaster(ctx, id, p, TypeReboot, len(issued) > 0)
}

func (m *Manager) runShutdown(ctx context.Context, id string, p plan, creds Credentials) {
	issued, ok := m.commandWorkers(ctx, id, p, TypeShutdown, creds)
	if !ok {
		return
	}
	m.powerMaster(ctx, id, p, TypeShutdown, len(issued) > 0)
}

// commandWorkers hands the power command to every compute node at once. Any
// node that does not accept it stops the operation: the control node stays up,
// because it is the only thing left that can report what happened.
func (m *Manager) commandWorkers(ctx context.Context, id string, p plan, opType Type, creds Credentials) ([]inventory.Node, bool) {
	if len(p.workers) == 0 {
		return nil, true
	}

	m.startStep(id, StepWorkerCommand)
	at := m.deps.Now()
	for _, w := range p.workers {
		name := w.NodeName
		m.setNode(id, name, func(n *NodeResult) { n.StartedAt = &at })
	}

	var requestID, scope, target, clusterID string
	if op, ok := m.Get(id); ok {
		requestID = op.RequestID
		scope = op.Scope
		target = op.Target
		clusterID = op.ClusterID
	}
	if !m.canContinue(id) {
		return nil, false
	}
	outcomes := m.deps.Dispatch(ctx, p.workers, PeerRequest{
		Type:        opType,
		OperationID: id,
		RequestID:   requestID,
		Scope:       scope,
		Target:      target,
		ClusterID:   clusterID,
	}, creds)

	accepted := map[string]bool{}
	for _, o := range outcomes {
		outcome := o
		// Detail is what the fan-out saw at the wire, so it is logged rather
		// than recorded: see reasons.
		var reason string
		if outcome.Code != "" {
			var detail error
			if outcome.Err != "" {
				detail = errors.New(outcome.Err)
			}
			reason = suppress(outcome.Code, "dispatch the power command to node "+outcome.NodeName, detail)
		}
		m.setNode(id, outcome.NodeName, func(n *NodeResult) {
			done := m.deps.Now()
			n.FinishedAt = &done
			if outcome.Code == "" {
				n.Status = NodeCommandIssued
				return
			}
			n.Status = NodeFailed
			n.Code = outcome.Code
			n.Error = reason
		})
		if outcome.Code == "" {
			accepted[outcome.NodeName] = true
		}
	}

	var issued []inventory.Node
	for _, w := range p.workers {
		if accepted[w.NodeName] {
			issued = append(issued, w)
		}
	}
	if !m.canContinue(id) {
		return nil, false
	}

	if len(issued) == len(p.workers) {
		m.finishStep(id, StepWorkerCommand, StepCommandIssued, "", "")
		return issued, true
	}

	m.failStep(id, StepWorkerCommand, CodeWorkerCommandFailed, "one or more nodes did not accept the power command")
	m.stopBeforeMaster(id, len(issued) > 0, CodeWorkerCommandFailed,
		"the control node was left running because a compute node did not accept the power command")
	return nil, false
}

// rebootProgress reads one node's restart out of what the cluster reports.
//
// down is "this node has left the boot it was told to leave": absent from the
// directory, NotReady, or already on another boot. up additionally requires it
// to be Ready on a boot that is not the one it started on — a node whose
// kubelet merely flapped comes back Ready on the same boot, and calling that a
// restart would take the control node down next on a cluster that is not in
// the state the operation claims.
func rebootProgress(obs inventory.Observation, present bool, baseline string) (down, up bool) {
	if !present {
		return true, false
	}
	rebooted := obs.BootID != baseline
	down = rebooted || !obs.Ready
	up = rebooted && obs.Ready
	return down, up
}

// awaitRestarts waits for every node that was told to reboot. They restart at
// the same time, so they are watched at the same time — and they are watched
// through the master's own view of the cluster rather than by dialling each
// node, because the node-local status endpoint needs a user credential and a
// cluster mid-reboot is exactly when the service that issues those is down.
func (m *Manager) awaitRestarts(ctx context.Context, id string, nodes []inventory.Node,
	baseline map[string]inventory.Observation) int {
	pending := map[string]bool{}
	wentDown := map[string]bool{}
	for _, n := range nodes {
		pending[n.NodeName] = true
	}

	var failed int
	settle := func(name, code, reason string) {
		delete(pending, name)
		m.setNode(id, name, func(n *NodeResult) {
			done := m.deps.Now()
			n.FinishedAt = &done
			if code == "" {
				n.Status = NodeRestarted
				return
			}
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
		})
		if code != "" {
			failed++
		}
	}

	start := m.deps.Now()
	downBy := start.Add(m.deps.Timeouts.Down)
	readyBy := start.Add(m.deps.Timeouts.Ready)
	for len(pending) > 0 {
		seen, err := m.deps.Observe(ctx)
		if err != nil {
			// The cluster is unreadable from here; that is indistinguishable
			// from every node being away, so the deadlines below decide.
			klog.V(4).Infof("clusterop: observe cluster: %v", err)
			seen = nil
		}
		for name := range pending {
			obs, present := seen[name]
			down, up := rebootProgress(obs, present, baseline[name].BootID)
			if down {
				wentDown[name] = true
			}
			if up && wentDown[name] {
				settle(name, "", "")
			}
		}

		now := m.deps.Now()
		for name := range pending {
			switch {
			case !wentDown[name] && !now.Before(downBy):
				settle(name, CodeNodeDidNotGoDown,
					fmt.Sprintf("node was still up %s after the reboot command", m.deps.Timeouts.Down))
			case !now.Before(readyBy):
				settle(name, CodeRestartTimeout,
					fmt.Sprintf("node did not come back ready on a new boot within %s", m.deps.Timeouts.Ready))
			}
		}
		if len(pending) == 0 {
			break
		}
		if err := m.deps.Sleep(ctx, m.deps.Timeouts.Poll); err != nil {
			for name := range pending {
				settle(name, CodeRestartTimeout, "the wait for this node was cancelled")
			}
		}
	}
	return failed
}

// powerMaster is the last thing an operation does. The record is written
// before the command is issued, because after it there may be no process left
// to write anything.
func (m *Manager) powerMaster(ctx context.Context, id string, p plan, opType Type, workersDone bool) {
	m.startStep(id, StepMasterCommand)
	at := m.deps.Now()
	m.setNode(id, p.master.NodeName, func(n *NodeResult) { n.StartedAt = &at })

	if opType == TypeReboot {
		// Written before the command, because after it there may be no
		// process left to write anything: this is what the next daemon
		// compares against to find out whether the machine really went down.
		boot, err := m.deps.HostBootID()
		if err != nil {
			klog.Warningf("clusterop: read this machine's boot id: %v", err)
		}
		m.update(id, func(op *Operation) { op.HostBootID = boot })
	}

	if !m.canContinue(id) {
		return
	}
	if err := m.deps.PowerSelf(ctx, opType); err != nil {
		code, reason := powerReason(CodeHostPowerFailed, "power the control node", err)
		m.setNode(id, p.master.NodeName, func(n *NodeResult) {
			done := m.deps.Now()
			n.FinishedAt = &done
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
		})
		m.failStep(id, StepMasterCommand, code, reason)
		status := StatusFailed
		if workersDone {
			status = StatusPartiallyFailed
		}
		m.fail(id, status, code, reason)
		return
	}

	m.setNode(id, p.master.NodeName, func(n *NodeResult) {
		done := m.deps.Now()
		n.FinishedAt = &done
		n.Status = NodeCommandIssued
	})
	m.finishStep(id, StepMasterCommand, StepCommandIssued, "", "")
	m.update(id, func(op *Operation) {
		// Not succeeded: the machine that would confirm the result is the one
		// carrying out the command.
		op.Status = StatusCommandIssued
	})
}

// stopBeforeMaster settles an operation that got part way through the compute
// nodes. The control node is recorded as skipped rather than failed: nothing
// was asked of it.
func (m *Manager) stopBeforeMaster(id string, partial bool, code, reason string) {
	m.skipRemainingNodes(id, "the operation stopped before this node was reached")
	status := StatusFailed
	if partial {
		status = StatusPartiallyFailed
	}
	m.fail(id, status, code, reason)
}

func (m *Manager) skipRemainingNodes(id string, reason string) {
	m.update(id, func(op *Operation) {
		at := m.deps.Now()
		for i := range op.Nodes {
			if op.Nodes[i].Status == NodePending {
				op.Nodes[i].Status = NodeSkipped
				op.Nodes[i].Error = reason
				op.Nodes[i].FinishedAt = &at
			}
		}
	})
}

func (m *Manager) fail(id string, status Status, code, reason string) {
	m.update(id, func(op *Operation) {
		op.Status = status
		op.Code = code
		op.Error = reason
		at := m.deps.Now()
		for i := range op.Steps {
			if op.Steps[i].Status == StepPending || op.Steps[i].Status == StepRunning {
				op.Steps[i].Status = StepFailed
				op.Steps[i].Code = code
				op.Steps[i].FinishedAt = &at
			}
		}
	})
}

func (m *Manager) startStep(id, name string) {
	m.update(id, func(op *Operation) {
		at := m.deps.Now()
		op.Steps = append(op.Steps, Step{Name: name, Status: StepRunning, StartedAt: &at})
	})
}

func (m *Manager) finishStep(id, name string, status StepStatus, code, reason string) {
	m.update(id, func(op *Operation) {
		at := m.deps.Now()
		for i := len(op.Steps) - 1; i >= 0; i-- {
			if op.Steps[i].Name != name {
				continue
			}
			op.Steps[i].Status = status
			op.Steps[i].Code = code
			op.Steps[i].Error = reason
			op.Steps[i].FinishedAt = &at
			return
		}
	})
}

func (m *Manager) failStep(id, name, code, reason string) {
	m.finishStep(id, name, StepFailed, code, reason)
}

func (m *Manager) setNode(id, name string, fn func(*NodeResult)) {
	m.update(id, func(op *Operation) {
		for i := range op.Nodes {
			if op.Nodes[i].NodeName == name {
				fn(&op.Nodes[i])
				return
			}
		}
	})
}

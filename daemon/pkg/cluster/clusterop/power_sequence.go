package clusterop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"k8s.io/klog/v2"
)

// powerSpec is everything the sequencing below used to decide by switching on
// the operation type. A power module describes its operation once and the
// sequence reads that description, so what separates a reboot from a
// shutdown lives with the module that carries it out rather than being
// spread across the stages it passes through.
type powerSpec struct {
	// opType is what goes on the wire to every compute node, and what this
	// machine's own power point is asked for last.
	opType Type

	// capability is what a compute node must declare before it is told to do
	// this. The control node is deliberately not checked this way; see
	// checkControlNode.
	capability string

	// provesBootChange marks an operation whose result is proved by which
	// boot a machine is on. Everything that proof needs follows from it: the
	// baseline is read before anything is told to go down, a node that
	// cannot say which boot it is on is refused, and the control node's own
	// boot is recorded before it is commanded.
	provesBootChange bool

	// awaitRestart waits for every commanded compute node to come back Ready
	// on a new boot. Nothing comes back from a shutdown, so nothing waits.
	awaitRestart bool

	// awaitDown holds a node-scope operation open until the target stops
	// answering. It is what a shutdown can observe instead of a restart.
	awaitDown bool

	// grace is how long the record stays locked between the command being
	// issued and the machine going down. It is read from the injected
	// timeouts rather than chosen, so no module can hold the cluster's
	// single-operation lock for longer than the wait it is bounded by.
	grace func(Timeouts) time.Duration

	// refuseControlNode is why this operation may not be aimed at the
	// control node on its own. nil means it may.
	refuseControlNode *refusal
}

// refusal is a stable code and the fixed sentence that goes with it.
type refusal struct {
	code   string
	reason string
}

// The two sentences a precheck refuses a whole operation with. They are
// named because the stage and the operation are given the same one: a caller
// shown one sentence and an operator reading the other cannot tell which
// describes what happened.
const (
	controlNodeRefused = "the control node cannot perform this operation"
	nodesRefused       = "one or more nodes cannot perform this operation"
)

// stopped is the precheck's way of saying "and this is what the operation
// ends on". The pointer is what distinguishes a refusal from a plan.
func stopped(o Outcome) *Outcome { return &o }

// plan is the node layout one operation will act on: every compute node first,
// the control node last.
type plan struct {
	master  inventory.Node
	workers []inventory.Node

	// baseline is the boot each node was on before the operation began. Only
	// an operation that proves a boot change has one; see Manager.baseline.
	baseline map[string]inventory.Observation
}

// outcomeStatePersistenceFailed ends a run whose state could no longer be
// recorded. The manager has already settled the record that way itself (see
// applyLocked), so completing on this is refused as terminal; it exists so
// the sequence still returns an outcome that says what stopped it.
var outcomeStatePersistenceFailed = Outcome{
	Status: StatusFailed,
	Code:   CodeStatePersistenceFailed,
}

// validatePowerScope refuses a request this module cannot carry out as
// asked, before anything is recorded or started.
//
// An empty scope is one it can: records written before scope existed carry
// none, and every run path has always read anything that is not "node" as
// covering the whole cluster.
//
// Whether the scope and the target describe the same operation is asked here
// rather than by the route, because it is a question about what this module
// does: powerNode is the only path that reads Target, and it is reached only
// for node scope, so a whole-cluster request naming one node would silently
// power every machine the caller did not name. The route's own check is a
// different one and still runs: that the owner signed this exact scope and
// this exact target.
func validatePowerScope(req CreateRequest) error {
	switch req.Scope {
	case "", ScopeCluster:
		if req.Target != "" {
			return fmt.Errorf("a %s-wide operation cannot name a target node", ScopeCluster)
		}
		return nil
	case ScopeNode:
		if req.Target == "" {
			return fmt.Errorf("a %s operation must name the node it acts on", ScopeNode)
		}
		return nil
	default:
		return fmt.Errorf("unsupported cluster operation scope %q", req.Scope)
	}
}

// runPowerOperation is the whole of a power module's Run. Nothing in it
// knows which power operation it is carrying out: spec says what this one
// needs, and the stages are the same ones every cluster power operation has
// always gone through.
func runPowerOperation(ctx context.Context, rt Runtime, req RunRequest, spec powerSpec) Outcome {
	m, id, ok := managerOf(rt)
	if !ok {
		// Only the manager's own Runtime carries the seams a power
		// operation is made of. Anything else cannot reach a machine, and
		// says so rather than reporting an operation it never performed.
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "this runtime cannot carry out a power operation",
		}
	}
	op, ok := rt.Operation()
	if !ok {
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "the operation this run was started for is gone",
		}
	}

	if op.Scope == ScopeNode {
		return m.powerNode(ctx, rt, id, spec, op.Target, req.Creds)
	}

	p, failure := m.precheck(ctx, id, spec, req.Creds)
	if failure != nil {
		klog.Warningf("clusterop: %s %s stopped at the precheck: %s", spec.opType, id, failure.Code)
		return *failure
	}
	return m.powerCluster(ctx, id, p, spec, req.Creds)
}

// powerCluster is the cluster-wide sequence: every compute node, then the
// control node. The control node is the machine running this code, so
// anything sequenced after it does not happen.
func (m *Manager) powerCluster(ctx context.Context, id string, p plan, spec powerSpec,
	creds Credentials) Outcome {
	issued, failure := m.commandWorkers(ctx, id, p, spec, creds)
	if failure != nil {
		return *failure
	}

	if spec.awaitRestart && len(issued) > 0 {
		m.startStep(id, StepWorkerRestart)
		failures := m.awaitRestarts(ctx, id, issued, p.baseline)
		if failures > 0 {
			m.failStep(id, StepWorkerRestart, CodeWorkerRestartFailed, "one or more nodes did not come back")
			return m.stopBeforeMaster(id, len(issued) > failures, CodeWorkerRestartFailed,
				"the control node was left running because a compute node did not come back")
		}
		m.finishStep(id, StepWorkerRestart, StepSucceeded, "", "")
	}

	return m.powerMaster(ctx, id, p, spec, len(issued) > 0)
}

// powerNode carries out an operation aimed at one node. The control node is
// the only one this daemon can act on directly; every other node is
// commanded through the same fan-out a cluster operation uses.
func (m *Manager) powerNode(ctx context.Context, rt Runtime, id string, spec powerSpec,
	target string, creds Credentials) Outcome {
	m.startStep(id, StepPrecheck)
	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for the precheck", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		return failedWith(CodeInventoryUnavailable, reason)
	}
	var node inventory.Node
	for _, candidate := range nodes {
		if candidate.NodeName == target {
			node = candidate
			break
		}
	}
	if node.NodeName == "" {
		const reason = "the node directory could not identify this node"
		m.failStep(id, StepPrecheck, CodeNodeIdentityUnknown, reason)
		return failedWith(CodeNodeIdentityUnknown, reason)
	}
	m.update(id, func(op *Operation) {
		op.Nodes = []NodeResult{{NodeName: node.NodeName, Role: node.Role, Status: NodePending}}
	})

	if node.Role == inventory.RoleMaster {
		if spec.refuseControlNode != nil {
			m.failStep(id, StepPrecheck, spec.refuseControlNode.code, spec.refuseControlNode.reason)
			return failedWith(spec.refuseControlNode.code, spec.refuseControlNode.reason)
		}
		if code, reason := m.checkControlNode(spec.opType); code != "" {
			m.failStep(id, StepPrecheck, code, reason)
			return failedWith(code, reason)
		}
		m.finishStep(id, StepPrecheck, StepSucceeded, "", "")
		return m.powerMaster(ctx, id, plan{master: node}, spec, false)
	}

	baseline, err := m.baseline(ctx, spec)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the reboot baseline", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		return failedWith(CodeInventoryUnavailable, reason)
	}
	if code, reason := m.checkWorker(ctx, node, spec, baseline, creds); code != "" {
		m.setNode(id, node.NodeName, func(n *NodeResult) {
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
		})
		m.failStep(id, StepPrecheck, code, reason)
		return failedWith(code, reason)
	}
	m.finishStep(id, StepPrecheck, StepSucceeded, "", "")

	issued, failure := m.commandWorkers(ctx, id,
		plan{workers: []inventory.Node{node}, baseline: baseline}, spec, creds)
	if failure != nil {
		return *failure
	}

	if spec.awaitDown {
		// Settled before the wait rather than after it: the command has
		// been issued, and the record has to say so while the node is on
		// its way down. Clearing the deadline once it has gone is what
		// releases the cluster for the next operation.
		settled := Outcome{
			Status:             StatusCommandIssued,
			CommandIssuedUntil: m.deps.Now().Add(spec.grace(m.deps.Timeouts)),
		}
		if err := rt.Complete(settled); err != nil {
			klog.Warningf("clusterop: record the issued command for %s: %v", id, err)
			return settled
		}
		m.awaitNodeDown(ctx, id, node.NodeName)
		// Recorded above, before the wait. Saying so is what keeps the
		// sequence from being written a second time on the way out.
		return settled.alreadyRecorded()
	}

	m.startStep(id, StepWorkerRestart)
	if m.awaitRestarts(ctx, id, issued, baseline) > 0 {
		const reason = "the node did not come back"
		m.failStep(id, StepWorkerRestart, CodeWorkerRestartFailed, reason)
		return failedWith(CodeWorkerRestartFailed, reason)
	}
	m.finishStep(id, StepWorkerRestart, StepSucceeded, "", "")
	return Outcome{Status: StatusSucceeded}
}

// precheck refuses the whole operation unless every compute node can be told
// to do what was asked. There is no force: a shutdown that skipped a node the
// master could not reach would leave that node running while the user is told
// the cluster is off.
func (m *Manager) precheck(ctx context.Context, id string, spec powerSpec,
	creds Credentials) (plan, *Outcome) {
	m.startStep(id, StepPrecheck)

	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for the precheck", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		return plan{}, stopped(failedWith(CodeInventoryUnavailable, reason))
	}

	// The topology gate runs before anything is inspected: a cluster this
	// daemon cannot sequence must not have a single node dialled by a run it
	// is going to refuse anyway.
	p, code, err := splitCluster(nodes)
	if err != nil {
		// splitCluster's errors are sentences written for this, and describe
		// only the shape of the node directory — how many control nodes it
		// lists, how many rows claim to be this machine. That is why they can
		// be shown; nothing here forwards an error from anywhere else.
		m.failStep(id, StepPrecheck, code, err.Error())
		return plan{}, stopped(failedWith(code, err.Error()))
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
	if code, reason := m.checkControlNode(spec.opType); code != "" {
		m.setNode(id, p.master.NodeName, func(n *NodeResult) {
			n.Status = NodeFailed
			n.Code = code
			n.Error = reason
			at := m.deps.Now()
			n.FinishedAt = &at
		})
		m.skipRemainingNodes(id, "blocked by the precheck")
		m.failStep(id, StepPrecheck, CodePrecheckFailed, controlNodeRefused)
		return plan{}, stopped(failedWith(CodePrecheckFailed, controlNodeRefused))
	}

	// An operation proved by a boot change is only provable against the boot
	// each node is on now, so the baseline is read here, before anything is
	// told to go down.
	baseline, err := m.baseline(ctx, spec)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the reboot baseline", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		return plan{}, stopped(failedWith(CodeInventoryUnavailable, reason))
	}
	p.baseline = baseline

	var blocked bool
	for _, w := range p.workers {
		code, reason := m.checkWorker(ctx, w, spec, baseline, creds)
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
		m.failStep(id, StepPrecheck, CodePrecheckFailed, nodesRefused)
		return plan{}, stopped(failedWith(CodePrecheckFailed, nodesRefused))
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
func (m *Manager) checkWorker(ctx context.Context, w inventory.Node, spec powerSpec,
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
	if spec.provesBootChange {
		if obs, ok := baseline[w.NodeName]; !ok || obs.BootID == "" {
			return CodeBootIDUnavailable, "the cluster does not report which boot this node is on"
		}
	}
	status, err := m.deps.Inspect(ctx, w, creds)
	if err != nil {
		return CodeNodeUnreachable, suppress(CodeNodeUnreachable, "read node "+w.NodeName+" for the precheck", err)
	}
	if c, ok := status.Capabilities[spec.capability]; !ok || !c.Supported {
		return CodePowerUnsupported, fmt.Sprintf("node does not support %s", spec.capability)
	}
	return "", ""
}

// baseline records which boot every node is on before any of them is told to
// leave it. An operation that proves nothing about a restart needs none.
func (m *Manager) baseline(ctx context.Context, spec powerSpec) (map[string]inventory.Observation, error) {
	if !spec.provesBootChange {
		return nil, nil
	}
	return m.deps.Observe(ctx)
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

// commandWorkers hands the power command to every compute node at once. Any
// node that does not accept it stops the operation: the control node stays up,
// because it is the only thing left that can report what happened.
func (m *Manager) commandWorkers(ctx context.Context, id string, p plan, spec powerSpec,
	creds Credentials) ([]inventory.Node, *Outcome) {
	if len(p.workers) == 0 {
		return nil, nil
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
		stopped := outcomeStatePersistenceFailed
		return nil, &stopped
	}
	outcomes := m.deps.Dispatch(ctx, p.workers, PeerRequest{
		Type:        spec.opType,
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
		stopped := outcomeStatePersistenceFailed
		return nil, &stopped
	}

	if len(issued) == len(p.workers) {
		m.finishStep(id, StepWorkerCommand, StepCommandIssued, "", "")
		return issued, nil
	}

	m.failStep(id, StepWorkerCommand, CodeWorkerCommandFailed, "one or more nodes did not accept the power command")
	failure := m.stopBeforeMaster(id, len(issued) > 0, CodeWorkerCommandFailed,
		"the control node was left running because a compute node did not accept the power command")
	return nil, &failure
}

// restartProgress reads one node's restart out of what the cluster reports.
//
// down is "this node has left the boot it was told to leave": absent from the
// directory, NotReady, or already on another boot. up additionally requires it
// to be Ready on a boot that is not the one it started on — a node whose
// kubelet merely flapped comes back Ready on the same boot, and calling that a
// restart would take the control node down next on a cluster that is not in
// the state the operation claims.
func restartProgress(obs inventory.Observation, present bool, baseline string) (down, up bool) {
	if nodeUnavailable(obs, present) {
		return true, false
	}
	restarted := obs.BootID != baseline
	down = restarted
	up = restarted && obs.Ready
	return down, up
}

func nodeUnavailable(obs inventory.Observation, present bool) bool {
	return !present || !obs.Ready
}

// awaitNodeDown holds a node-scope operation open until the target stops
// answering, and clears the deadline when it does. Nothing here decides what
// going down means for the operation: the record already says the command was
// issued, and this only reports whether the machine acted on it.
func (m *Manager) awaitNodeDown(ctx context.Context, id, target string) {
	deadline := m.deps.Now().Add(m.deps.Timeouts.Down)
	for m.deps.Now().Before(deadline) {
		seen, err := m.deps.Observe(ctx)
		if err == nil {
			obs, present := seen[target]
			if nodeUnavailable(obs, present) {
				m.update(id, func(op *Operation) {
					op.CommandIssuedUntil = time.Time{}
				})
				return
			}
		}
		if err := m.deps.Sleep(ctx, m.deps.Timeouts.Poll); err != nil {
			return
		}
	}
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
			down, up := restartProgress(obs, present, baseline[name].BootID)
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
func (m *Manager) powerMaster(ctx context.Context, id string, p plan, spec powerSpec,
	workersDone bool) Outcome {
	m.startStep(id, StepMasterCommand)
	at := m.deps.Now()
	m.setNode(id, p.master.NodeName, func(n *NodeResult) { n.StartedAt = &at })

	if spec.provesBootChange {
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
		return outcomeStatePersistenceFailed
	}
	if err := m.deps.PowerSelf(ctx, spec.opType); err != nil {
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
		return settledWith(status, code, reason)
	}

	m.setNode(id, p.master.NodeName, func(n *NodeResult) {
		done := m.deps.Now()
		n.FinishedAt = &done
		n.Status = NodeCommandIssued
	})
	m.finishStep(id, StepMasterCommand, StepCommandIssued, "", "")

	// Not succeeded: the machine that would confirm the result is the one
	// carrying out the command.
	return Outcome{
		Status:             StatusCommandIssued,
		CommandIssuedUntil: m.deps.Now().Add(spec.grace(m.deps.Timeouts)),
	}
}

// stopBeforeMaster settles an operation that got part way through the compute
// nodes. The control node is recorded as skipped rather than failed: nothing
// was asked of it.
func (m *Manager) stopBeforeMaster(id string, partial bool, code, reason string) Outcome {
	m.skipRemainingNodes(id, "the operation stopped before this node was reached")
	status := StatusFailed
	if partial {
		status = StatusPartiallyFailed
	}
	return settledWith(status, code, reason)
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

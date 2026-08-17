package clusterop

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"k8s.io/klog/v2"
)

// UpgradeStagePhase is how a single node's run of a single stage ended.
type UpgradeStagePhase string

const (
	UpgradeStagePhaseRunning   UpgradeStagePhase = "running"
	UpgradeStagePhaseSucceeded UpgradeStagePhase = "succeeded"
	UpgradeStagePhaseFailed    UpgradeStagePhase = "failed"
)

// Terminal reports whether the node has stopped working on the stage.
func (p UpgradeStagePhase) Terminal() bool {
	return p == UpgradeStagePhaseSucceeded || p == UpgradeStagePhaseFailed
}

// UpgradeStageRequest is what one node is told to run.
//
// It names a stage of a plan rather than a list of tasks. The node resolves it
// against the plan its own olares-cli derives. Sending the tasks instead would
// make the control node's idea of the upgrade authoritative over the binary
// that implements it, which is backwards: the tasks only exist in that binary.
//
// What makes both sides derive the same plan is that both run the same
// olares-cli version, which the orchestrator establishes before dispatching
// anything — the node prepare stage installs it and checkWorkersCanUpgrade
// confirms it.
type UpgradeStageRequest struct {
	OperationID string `json:"operationId"`
	Stage       string `json:"stage"`
	Version     string `json:"version"`

	// ClusterID is what the receiving node compares with its own kube-system
	// UID, so a token cannot be carried into another cluster. It is the only
	// part of the request that authorizes anything; everything else describes
	// the work.
	ClusterID string `json:"clusterId"`
}

// UpgradeStageState is one node's record of one stage. It is persisted by the node
// that ran it, because an upgrade stage restarts olaresd — installing the new
// daemon is part of the work — and a state held in memory would be lost by the
// process whose own replacement it is describing.
type UpgradeStageState struct {
	OperationID string            `json:"operationId"`
	Stage       string            `json:"stage"`
	Version     string            `json:"version"`
	Phase       UpgradeStagePhase `json:"phase"`
	Code        string            `json:"code,omitempty"`
	Error       string            `json:"error,omitempty"`
	StartedAt   time.Time         `json:"startedAt"`
	FinishedAt  *time.Time        `json:"finishedAt,omitempty"`
}

// UpgradeStageRunner runs upgrade stages on one machine and remembers what happened.
//
// The control node holds one for itself and reaches every other node's through
// HTTP. Both sides are the same interface on purpose: the control node is a
// node, its stages are recorded the same way, and an upgrade that resumes
// after a restart has one kind of record to read rather than two.
type UpgradeStageRunner interface {
	// Start begins a stage and returns as soon as it is under way. It is
	// idempotent per (operationId, stage): asked again for something it is
	// already running or has finished, it reports that rather than starting a
	// second copy. This is what makes an interrupted upgrade resumable —
	// after a restart the orchestrator asks again, and a node that already did
	// the work says so instead of doing it twice.
	Start(ctx context.Context, req UpgradeStageRequest) (UpgradeStageState, error)

	// Status reports a stage this runner was asked to run.
	Status(operationID, stageName string) (UpgradeStageState, bool)
}

// upgradeDeps is the upgrade's own seams. It is only ever reached from a run
// the module already admitted, and upgradeModule.Run is what checks they are
// all there — so nothing below it checks again.
func (m *Manager) upgradeDeps() *UpgradeDeps { return m.deps.Upgrade }

// runUpgrade walks the plan, stage by stage, and does not begin a stage until
// the previous one has finished on every node it was scheduled on.
//
// That barrier is the whole ordering model. A stage boundary exists exactly
// where the plan changes scope, so an author who writes two tasks in order and
// says they run in different places has already said that the second waits for
// the first everywhere. There is no dependency graph here to disagree with it.
// It takes no credentials: see UpgradeDeps.Auth for why an upgrade cannot be
// carried by the signature that authorized it, and why that is what lets a
// resumed run continue without one.
func (m *Manager) runUpgrade(ctx context.Context, id string) Outcome {
	run, failure := m.upgradePrecheck(ctx, id)
	if failure != nil {
		return *failure
	}

	for _, stage := range run.plan.Stages {
		// Asked between stages and nowhere else: this is the point where every
		// node is between pieces of work rather than inside one, so it is the
		// only point at which stopping leaves the cluster somewhere anybody
		// can describe. See Manager.RequestStop.
		if m.StopRequested(id) {
			klog.Infof("clusterop: upgrade %s stopping before stage %s at the operator's request", id, stage.Name)
			return failedWith(CodeUpgradeCancelled, reasonFor(CodeUpgradeCancelled))
		}

		// A stage this operation already finished is not run again. This is
		// what resuming means: after olaresd restarts — which an upgrade
		// causes on the node it is orchestrating from — the run starts over
		// from the plan and walks past everything already recorded as done.
		if m.stageSettled(id, stage.Name) {
			klog.V(4).Infof("clusterop: upgrade %s resuming past completed stage %s", id, stage.Name)
			continue
		}

		// Two ways for a stage to have nothing to do, both recorded as skipped
		// rather than omitted: a plan whose stages silently disappear is one
		// nobody can check against the plan it came from, and a stage recorded
		// as succeeded when it did nothing cannot be told from one that did.
		if stage.Empty() {
			// This version put no tasks here. Dispatching it would start an
			// olares-cli on every node in its fanout to resolve an empty task
			// list — and, worse, would write the same record a stage that ran
			// real work writes.
			m.recordSkippedStage(id, stage, "this version contributes no tasks to this stage")
			continue
		}

		targets := stage.Placement.Targets(run.nodes)
		if len(targets) == 0 {
			// A workers-scoped stage on a single-node cluster.
			m.recordSkippedStage(id, stage, "this cluster has no nodes in this stage's scope")
			continue
		}

		if failure := m.runStage(ctx, id, run, stage, targets); failure != nil {
			return *failure
		}
		if !m.canContinue(id) {
			return outcomeStatePersistenceFailed
		}
	}

	return Outcome{Status: StatusSucceeded}
}

// upgradeRun is everything the stage loop needs, named so that the plan of
// stages and the layout of nodes cannot be mistaken for each other — they are
// both called some kind of "plan" in this package and they are not the same
// thing at all.
type upgradeRun struct {
	plan  UpgradePlan
	nodes plan
	token string
}

// upgradePrecheck settles everything that has to be true before the first
// stage runs, in the order the answers become available: what the cluster is,
// what may reach it, whether every node can take part, what the upgrade
// consists of, and then getting the release onto the nodes that need it.
//
// The order is not arbitrary and cannot be shuffled. The token is minted
// before the first probe because the probe is carried by it; the plan is read
// before nodes are asked whether they can run it because that question has no
// meaning without one; and the nodes are prepared before being asked what
// version they hold, because preparing them is what changes the answer.
func (m *Manager) upgradePrecheck(ctx context.Context, id string) (upgradeRun, *Outcome) {
	m.startStep(id, StepPrecheck)

	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for the upgrade precheck", err)
		m.failStep(id, StepPrecheck, CodeInventoryUnavailable, reason)
		return upgradeRun{}, stopped(failedWith(CodeInventoryUnavailable, reason))
	}
	p, code, err := splitCluster(nodes)
	if err != nil {
		m.failStep(id, StepPrecheck, code, err.Error())
		return upgradeRun{}, stopped(failedWith(code, err.Error()))
	}

	// The token is minted here rather than later because the readiness probe
	// below is the first thing that talks to a node, and it is carried by the
	// token like every other hop of an upgrade. A resumed run re-reads it
	// rather than minting a new one, so the workers a previous run already
	// dispatched to still recognize this one.
	token, err := m.upgradeToken(ctx, id, len(p.workers) > 0)
	if err != nil {
		reason := suppress(CodePlanUnavailable, "prepare the upgrade authorization", err)
		m.failStep(id, StepPrecheck, CodePlanUnavailable, reason)
		return upgradeRun{}, stopped(failedWith(CodePlanUnavailable, reason))
	}

	if failure := m.checkWorkersReachable(ctx, id, p, token); failure != nil {
		return upgradeRun{}, failure
	}

	// The plan is read before the nodes are asked whether they can run it,
	// because "can you run this" has no answer until there is a this. The two
	// were the other way round until it turned out the only useful question
	// to ask a node is whether it holds the version being rolled out.
	up, failure := m.readUpgradePlan(ctx, id)
	if failure != nil {
		return upgradeRun{}, failure
	}

	// Each compute node fetches the target version's binaries before it is
	// asked whether it can run a stage of it, because that is what makes the
	// answer yes.
	if failure := m.prepareNodes(ctx, id, p, up, token); failure != nil {
		return upgradeRun{}, failure
	}
	if failure := m.checkWorkersCanUpgrade(ctx, id, p, up, token); failure != nil {
		return upgradeRun{}, failure
	}

	return upgradeRun{plan: up, nodes: p, token: token}, nil
}

// checkWorkersReachable refuses the upgrade unless every compute node is
// Ready and speaks the protocol.
//
// Unlike a shutdown, where a node that is already down is arguably fine, an
// upgrade that skips a node leaves the cluster running two versions with
// nothing recording which node is on which.
//
// It records an answer for every compute node before it refuses, rather than
// stopping at the first one. Which nodes cannot take part is the only thing
// that makes the refusal actionable — an olaresd from before staged upgrades
// has to be updated by hand, and being told that "a node" needs it does not
// say which. Asking all of them costs one round trip each against a cluster
// that is not going to be upgraded either way.
func (m *Manager) checkWorkersReachable(ctx context.Context, id string, p plan, token string) *Outcome {
	m.setPrecheckNodes(id, p)

	var (
		code   string
		reason string
	)
	for _, w := range p.workers {
		c, r := m.checkUpgradeWorker(ctx, id, w, token)
		if c == "" {
			m.settlePrecheckNode(id, w.NodeName, NodeSucceeded, "", "")
			continue
		}
		m.settlePrecheckNode(id, w.NodeName, NodeFailed, c, r)
		if code == "" {
			code, reason = c, r
		}
	}
	if code != "" {
		m.failStep(id, StepPrecheck, code, reason)
		return stopped(failedWith(code, reason))
	}
	m.finishStep(id, StepPrecheck, StepSucceeded, "", "")
	return nil
}

// setPrecheckNodes puts every compute node on the precheck step, so the answers
// below have somewhere to land.
func (m *Manager) setPrecheckNodes(id string, p plan) {
	m.update(id, func(op *Operation) {
		step := lastStepNamed(op, StepPrecheck)
		if step == nil {
			return
		}
		step.Nodes = make([]NodeResult, 0, len(p.workers))
		for _, w := range p.workers {
			step.Nodes = append(step.Nodes,
				NodeResult{NodeName: w.NodeName, Role: w.Role, Status: NodePending})
		}
	})
}

func (m *Manager) settlePrecheckNode(id, nodeName string, status NodeStatus, code, reason string) {
	m.settleStageNode(id, StepPrecheck, nodeName, status, code, reason)
}

// readUpgradePlan asks this node's olares-cli what the upgrade consists of,
// and refuses one it cannot make sense of rather than starting to walk it.
func (m *Manager) readUpgradePlan(ctx context.Context, id string) (UpgradePlan, *Outcome) {
	m.startStep(id, StepPlan)
	up, err := m.upgradeDeps().Plan(ctx)
	if err == nil {
		err = up.Validate()
	}
	if err != nil {
		reason := suppress(CodePlanUnavailable, "read the upgrade plan", err)
		m.failStep(id, StepPlan, CodePlanUnavailable, reason)
		return UpgradePlan{}, stopped(failedWith(CodePlanUnavailable, reason))
	}
	m.finishStep(id, StepPlan, StepSucceeded, "", "")
	klog.Infof("clusterop: upgrade %s to %s in %d stages", id, up.Version, len(up.Stages))
	return up, nil
}

// prepareNodes brings every compute node to the target version's binaries.
//
// It runs as an ordinary stage — same dispatch, same persistence, same
// polling, same resume — because it needs all of those for the same reason a
// plan stage does, and more acutely: installing the new olaresd on a node
// restarts the daemon that is running the work. That node's stage record is
// left saying "running", its replacement settles it as interrupted, and this
// orchestrator dispatches it again; the second run finds the packages
// downloaded and the binaries already at the target version and finishes.
func (m *Manager) prepareNodes(ctx context.Context, id string, p plan, up UpgradePlan, token string) *Outcome {
	if len(p.workers) == 0 {
		return nil
	}
	return m.runStage(ctx, id, upgradeRun{plan: up, nodes: p, token: token}, NodePrepareStage(), p.workers)
}

// checkWorkersCanUpgrade confirms every compute node now holds the version
// being rolled out, after the prepare stage has put it there.
//
// This is the check that the prepare stage worked, asked of the node rather
// than inferred from the stage having returned success, and it is the only
// thing standing between a node on the previous olares-cli and that node
// resolving every stage name against the previous version's plan. Nothing
// downstream compares the two plans, so a mismatch that gets past here is not
// caught at all.
func (m *Manager) checkWorkersCanUpgrade(ctx context.Context, id string, p plan, up UpgradePlan,
	token string) *Outcome {
	if len(p.workers) == 0 {
		return nil
	}
	m.startStep(id, StepUpgradeReadiness)

	for _, w := range p.workers {
		code, reason := m.checkWorkerVersion(ctx, id, w, up, token)
		if code == "" {
			continue
		}
		klog.Warningf("clusterop: upgrade %s: node %s cannot run stages: %s", id, w.NodeName, code)
		m.failStep(id, StepUpgradeReadiness, code, reason)
		return stopped(failedWith(code, reason))
	}
	m.finishStep(id, StepUpgradeReadiness, StepSucceeded, "", "")
	return nil
}

func (m *Manager) checkWorkerVersion(ctx context.Context, id string, w inventory.Node, up UpgradePlan,
	token string) (code, reason string) {
	ready, err := m.upgradeDeps().Readiness(ctx, w, id, token)
	if err != nil {
		return CodeNodeUnreachable, suppress(CodeNodeUnreachable,
			"read node "+w.NodeName+" after preparing it", err)
	}
	if !ready.Supported {
		// It answered the precheck and does not now. The node was left in a
		// state the prepare stage did not intend.
		return CodeUpgradeUnsupported, reasonFor(CodeUpgradeUnsupported)
	}
	// A node that cannot say which olares-cli it holds is refused like one
	// holding the wrong version. This used to be allowed through, on the
	// grounds that the plan digest would catch it at the stage; with the
	// digest gone this is the only check there is, and "it would not say" is
	// not an answer to build an upgrade on.
	if ready.CLIVersion != up.Version {
		return CodeVersionMismatch, reasonFor(CodeVersionMismatch)
	}
	return "", ""
}

// upgradeToken is the credential workers check. A single-node cluster needs
// none: nothing leaves the machine, so there is no hop to authorize and no
// reason to put a secret in the cluster for one.
func (m *Manager) upgradeToken(ctx context.Context, id string, hasWorkers bool) (string, error) {
	if !hasWorkers {
		return "", nil
	}
	return m.upgradeDeps().Auth(ctx, id)
}

// checkUpgradeWorker asks the cheap questions about one compute node, before
// anything has been read or downloaded anywhere.
//
// The last of them is whether the node speaks the staged-upgrade protocol at
// all. An olaresd from before staged upgrades serves no stage route and does
// not declare upgrade.stages — it has never heard of the key — and the missing
// capability is the only way to tell. It is asked here, at the very start,
// rather than later: the alternative is discovering it from a 404 after the
// orchestrator has already spent twenty minutes downloading a release onto a
// node that was never going to be able to use it.
//
// What is deliberately not asked here is which olares-cli the node holds. At
// this point it holds the previous one — that is the normal state, and fixing
// it is what the node prepare stage is for. See checkWorkersCanUpgrade.
func (m *Manager) checkUpgradeWorker(ctx context.Context, id string, w inventory.Node, token string) (code, reason string) {
	if w.NodeName == "" || w.Role == "" || w.Role == inventory.RoleUnknown {
		return CodeNodeIdentityUnknown, "the node directory could not identify this node"
	}
	if !w.Ready {
		return CodeNodeNotReady, "the cluster does not consider this node ready"
	}
	if w.IP == "" {
		return CodeNodeUnaddressable, "node has no internal address"
	}
	// An olaresd from before upgrade stages existed serves no readiness route,
	// so it answers with a transport error rather than a refusal. Both mean
	// the same thing here and are reported the same way: this node cannot
	// take part, found before anything has run.
	ready, err := m.upgradeDeps().Readiness(ctx, w, id, token)
	if err != nil {
		klog.Warningf("clusterop: node %s did not answer the upgrade readiness probe: %v", w.NodeName, err)
		return CodeUpgradeUnsupported, reasonFor(CodeUpgradeUnsupported)
	}
	if !ready.Supported {
		return CodeUpgradeUnsupported, reasonFor(CodeUpgradeUnsupported)
	}
	return "", ""
}

// runStage runs one stage on its nodes and reports whether the upgrade may go on.
func (m *Manager) runStage(ctx context.Context, id string, run upgradeRun, stage UpgradeStage,
	targets []inventory.Node) *Outcome {
	m.beginStage(id, stage, targets)

	req := m.stageRequest(id, run.plan, stage)
	if !m.canContinue(id) {
		stop := outcomeStatePersistenceFailed
		return &stop
	}

	// Read before anything is told to do anything, because it is what a
	// machine having gone down is proved against later.
	baseline := m.restartBaseline(ctx, stage)

	failures := m.runStageOnNodes(ctx, id, stage, targets, req, run.token, baseline)
	if failures > 0 {
		m.skipRemainingStageNodes(id, stage.Name)
		m.failStep(id, stage.Name, CodeStageFailed, "one or more nodes did not complete this stage")
		return stopped(m.failUpgradeAt(id, run.plan, stage))
	}
	m.finishStep(id, stage.Name, StepSucceeded, "", "")
	return nil
}

// restartBaseline is the boot each node is on before a stage that may take
// machines down. A stage that takes none needs none.
func (m *Manager) restartBaseline(ctx context.Context, stage UpgradeStage) map[string]inventory.Observation {
	if !stage.AwaitRestart {
		return nil
	}
	seen, err := m.deps.Observe(ctx)
	if err != nil {
		// Not fatal, and deliberately so: without a baseline every node
		// reads as "has not gone down", which the wait below settles as
		// "this stage did not reboot it" once the down window passes. That
		// is the same answer this stage produced before it waited at all.
		klog.Warningf("clusterop: read the restart baseline for stage %s: %v", stage.Name, err)
		return nil
	}
	return seen
}

// awaitNodeRestart holds one node's slot in the stage until the machine is
// back, and reports whether it came.
//
// It is not the power operations' awaitRestarts, and the difference is the
// whole point: a power operation commanded every node and can insist each one
// went down, while a stage's tasks decide for themselves — the GPU driver
// upgrade reboots only the machines whose driver it changed. So a node that
// never stops answering is not a node that failed to reboot; it is a node this
// stage had no reason to reboot, and it is done once the window in which it
// would have gone down has passed.
//
// The control node is the one case this cannot observe. It is scheduled last
// and takes the orchestrator down with it, so what confirms that reboot is the
// daemon coming back and resuming the operation.
func (m *Manager) awaitNodeRestart(ctx context.Context, id string, stage UpgradeStage,
	node inventory.Node, baseline map[string]inventory.Observation) bool {
	name := node.NodeName
	start := m.deps.Now()
	downBy := start.Add(m.deps.Timeouts.Down)
	readyBy := start.Add(m.deps.Timeouts.Ready)

	var wentDown bool
	for {
		seen, err := m.deps.Observe(ctx)
		if err != nil {
			// Unreadable from here is indistinguishable from the node being
			// away, so the deadlines below decide.
			klog.V(4).Infof("clusterop: observe cluster while waiting for %s: %v", name, err)
			seen = nil
		}
		obs, present := seen[name]
		down, up := restartProgress(obs, present, baseline[name].BootID)
		if down {
			wentDown = true
		}
		if up && wentDown {
			klog.Infof("clusterop: node %s came back for stage %s of %s", name, stage.Name, id)
			return true
		}

		now := m.deps.Now()
		switch {
		case !wentDown && !now.Before(downBy):
			// It never went down, so this stage had no reason to reboot it.
			return true
		case !now.Before(readyBy):
			m.settleStageNode(id, stage.Name, name, NodeFailed,
				CodeRestartTimeout, reasonFor(CodeRestartTimeout))
			return false
		}

		if err := m.deps.Sleep(ctx, m.deps.Timeouts.Poll); err != nil {
			m.settleStageNode(id, stage.Name, name, NodeFailed,
				CodeRestartTimeout, "the wait for this node was cancelled")
			return false
		}
	}
}

// runStageOnNodes gives the stage to its nodes, at most MaxParallel of them at
// a time, and reports how many did not finish it.
//
// A bounded stage also stops handing the work out once something has failed.
// The reason a stage is bounded is that running it costs the cluster capacity,
// so starting it on another node after one has broken takes more of the
// cluster down for an upgrade that is already not going to finish. An
// unbounded stage has no such cost and has already started everywhere.
func (m *Manager) runStageOnNodes(ctx context.Context, id string, stage UpgradeStage,
	targets []inventory.Node, req UpgradeStageRequest, token string,
	baseline map[string]inventory.Observation) int {
	limit := stage.MaxParallel
	if limit <= 0 || limit > len(targets) {
		limit = len(targets)
	}

	var (
		mu       sync.Mutex
		failures int
		wg       sync.WaitGroup
	)
	slots := make(chan struct{}, limit)

	for _, node := range targets {
		// The slot is taken before the count is read, and that order is the
		// whole of the "stop after a failure" behaviour: waiting for a slot
		// is what makes an earlier node's result available to look at. Read
		// first and a serial stage would launch the next node while the
		// previous one was still running, which is neither serial nor a stop.
		slots <- struct{}{}
		if stage.Bounded() {
			mu.Lock()
			stop := failures > 0
			mu.Unlock()
			if stop {
				<-slots
				break
			}
		}
		wg.Add(1)
		go func(n inventory.Node) {
			defer wg.Done()
			defer func() { <-slots }()
			if m.runStageOnNode(ctx, id, stage, n, req, token, baseline) {
				return
			}
			mu.Lock()
			failures++
			mu.Unlock()
		}(node)
	}

	wg.Wait()
	return failures
}

// runStageOnNode starts the stage on one node and waits for it to finish —
// and, for a stage that may take the machine down, for the machine to come
// back.
//
// The second wait is here, inside the node's slot, rather than after every
// node has been given the stage. That is what makes MaxParallel mean anything
// for a stage like this one: rebooting is exactly the work a cluster wants
// done one machine at a time, and a barrier at the end of the stage would let
// every node be told to reboot first and only then start counting them back.
func (m *Manager) runStageOnNode(ctx context.Context, id string, stage UpgradeStage,
	node inventory.Node, req UpgradeStageRequest, token string,
	baseline map[string]inventory.Observation) bool {
	at := m.deps.Now()
	m.setStageNode(id, stage.Name, node.NodeName, func(n *NodeResult) {
		n.Status = NodeRunning
		n.StartedAt = &at
	})

	state, err := m.startStageOn(ctx, node, req, token)
	if err != nil {
		code, reason := stageDispatchReason(node, err)
		m.settleStageNode(id, stage.Name, node.NodeName, NodeFailed, code, reason)
		return false
	}
	if state.Phase.Terminal() {
		if !m.settleFromState(id, stage.Name, node.NodeName, state) {
			return false
		}
	} else if !m.awaitStage(ctx, id, stage, node, req, token) {
		return false
	}

	if !stage.AwaitRestart {
		return true
	}
	return m.awaitNodeRestart(ctx, id, stage, node, baseline)
}

// awaitStage polls one node until its stage settles or the deadline passes.
//
// A poll that fails is not a failed stage: an upgrade stage restarts olaresd
// on the node running it, so being unable to reach a node for a while is the
// expected middle of the work rather than the end of it. Only the deadline
// ends the wait.
func (m *Manager) awaitStage(ctx context.Context, id string, stage UpgradeStage,
	node inventory.Node, req UpgradeStageRequest, token string) bool {
	limit := stage.Timeout(m.deps.Timeouts.Stage)
	deadline := m.deps.Now().Add(limit)

	for {
		if err := m.deps.Sleep(ctx, m.deps.Timeouts.Poll); err != nil {
			m.settleStageNode(id, stage.Name, node.NodeName, NodeFailed, CodeStageTimeout,
				"the wait for this node was cancelled")
			return false
		}

		state, err := m.stageStatusOf(ctx, node, id, stage.Name, token)
		switch {
		case err != nil:
			klog.V(4).Infof("clusterop: read stage %s on %s: %v", stage.Name, node.NodeName, err)
		case state.Phase.Terminal() && state.Code == CodeDaemonRestarted:
			// Not a failure: the stage restarted the olaresd that was
			// running it. Several stages are expected to — node-prepare
			// installs the new daemon, reboot-nodes takes the machine down —
			// and the node marks its own record failed on the way back up
			// because nothing about the half-finished run can be recovered.
			//
			// What is recoverable is the stage, because every task in one is
			// reentrant. So it is asked again, inside this operation rather
			// than by failing it and leaving a fresh operation to start the
			// whole upgrade over. The deadline still bounds it, so a node
			// that restarts endlessly ends as a timeout rather than a loop.
			klog.Infof("clusterop: node %s restarted while running stage %s of %s; asking it again",
				node.NodeName, stage.Name, id)
			if _, err := m.startStageOn(ctx, node, req, token); err != nil {
				// Reported and not settled: the deadline decides. A node
				// that is on its way down refuses the request it is about
				// to be able to serve.
				klog.V(4).Infof("clusterop: re-dispatch stage %s to %s: %v",
					stage.Name, node.NodeName, err)
			}
		case state.Phase.Terminal():
			return m.settleFromState(id, stage.Name, node.NodeName, state)
		}

		if !m.deps.Now().Before(deadline) {
			m.settleStageNode(id, stage.Name, node.NodeName, NodeFailed, CodeStageTimeout,
				fmt.Sprintf("this node did not finish the stage within %s", limit))
			return false
		}
		if !m.canContinue(id) {
			return false
		}
	}
}

// startStageOn runs the stage locally when the target is this machine, and
// asks the node otherwise. The local path is not an optimization: an upgrade
// stage on the control node may restart the very daemon that would have made
// the HTTP call, and the in-process runner records its state before that can
// happen.
func (m *Manager) startStageOn(ctx context.Context, node inventory.Node, req UpgradeStageRequest,
	token string) (UpgradeStageState, error) {
	if node.IsSelf {
		return m.upgradeDeps().Local.Start(ctx, req)
	}
	return m.upgradeDeps().Start(ctx, node, req, token)
}

func (m *Manager) stageStatusOf(ctx context.Context, node inventory.Node, operationID, stageName string,
	token string) (UpgradeStageState, error) {
	if node.IsSelf {
		state, ok := m.upgradeDeps().Local.Status(operationID, stageName)
		if !ok {
			return UpgradeStageState{}, fmt.Errorf("no record of this stage on this node")
		}
		return state, nil
	}
	return m.upgradeDeps().Status(ctx, node, operationID, stageName, token)
}

func (m *Manager) stageRequest(id string, up UpgradePlan, stage UpgradeStage) UpgradeStageRequest {
	req := UpgradeStageRequest{
		OperationID: id,
		Stage:       stage.Name,
		Version:     up.Version,
	}
	if op, ok := m.Get(id); ok {
		req.ClusterID = op.ClusterID
	}
	return req
}

// stageDispatchReason keeps the transport's own classification rather than
// flattening every failure into one code: a node that never answered and a
// node that answered with a refusal are different problems for whoever has to
// act on the record.
func stageDispatchReason(node inventory.Node, err error) (code, reason string) {
	code = CodeDispatchFailed
	var dispatch *StageDispatchError
	if errors.As(err, &dispatch) && dispatch.Code != "" {
		code = dispatch.Code
	}
	return code, suppress(code, "start an upgrade stage on node "+node.NodeName, err)
}

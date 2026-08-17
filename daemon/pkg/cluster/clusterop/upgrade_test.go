package clusterop

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// upgradeHarness drives a whole cluster upgrade with no cluster, no network,
// no clock and no olares-cli.
type upgradeHarness struct {
	t    *testing.T
	deps Deps

	mu sync.Mutex
	// started records every (node, stage) the orchestrator dispatched, in
	// order, which is what the ordering assertions read.
	started []string
	// stageStates is the state each (node, stage) reports.
	stageStates map[string]*UpgradeStageState
	// failAt makes one (node, stage) fail.
	failAt map[string]bool
	// unreachable makes a node refuse to accept a stage at all.
	unreachable map[string]bool
	// restartOnce makes a node report, once, that its own stage restarted the
	// olaresd running it — what node-prepare and reboot-nodes really do.
	restartOnce map[string]bool
	// restartPending is the half of that the status read answers with.
	restartPending map[string]bool
	// observed is what the cluster reports about each node, for the stages
	// that wait for machines to come back.
	observed map[string]inventory.Observation
	// readiness overrides what a node answers the readiness probe with. An
	// olaresd from before upgrade stages existed serves no such route, which
	// is expressed here as an error rather than a negative answer.
	readiness map[string]UpgradeReadiness

	plan UpgradePlan

	// onStart, when set, runs while a node is inside a stage. It is how a test
	// observes how many of them are in there at the same time.
	onStart func(node, stage string)
}

// readyOn is what a node running this version's olaresd answers.
func readyOn(version string) UpgradeReadiness {
	return UpgradeReadiness{Supported: true, CLIVersion: version}
}

func stageKey(node, stage string) string { return node + "/" + stage }

func newUpgradeHarness(t *testing.T, plan UpgradePlan, nodes []inventory.Node) *upgradeHarness {
	t.Helper()
	h := &upgradeHarness{
		t:              t,
		stageStates:    map[string]*UpgradeStageState{},
		failAt:         map[string]bool{},
		unreachable:    map[string]bool{},
		readiness:      map[string]UpgradeReadiness{},
		restartOnce:    map[string]bool{},
		restartPending: map[string]bool{},
		observed:       map[string]inventory.Observation{},
		plan:           plan,
	}

	start := func(node string, req UpgradeStageRequest) (UpgradeStageState, error) {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.unreachable[node] {
			return UpgradeStageState{}, errors.New("node did not answer")
		}
		key := stageKey(node, req.Stage)
		if existing, ok := h.stageStates[key]; ok && existing.Phase == UpgradeStagePhaseSucceeded {
			return *existing, nil
		}
		h.started = append(h.started, key)
		if h.restartOnce[key] {
			// Accepted and then interrupted: the stage took the daemon that
			// was running it down with it.
			delete(h.restartOnce, key)
			h.restartPending[key] = true
			state := &UpgradeStageState{
				OperationID: req.OperationID, Stage: req.Stage,
				Version: req.Version, Phase: UpgradeStagePhaseRunning,
			}
			h.stageStates[key] = state
			return *state, nil
		}
		if h.onStart != nil {
			h.mu.Unlock()
			h.onStart(node, req.Stage)
			h.mu.Lock()
		}

		state := &UpgradeStageState{
			OperationID: req.OperationID,
			Stage:       req.Stage,
			Version:     req.Version,
			Phase:       UpgradeStagePhaseSucceeded,
		}
		if h.failAt[key] {
			state.Phase = UpgradeStagePhaseFailed
			state.Code = CodeStageFailed
		}
		h.stageStates[key] = state
		return *state, nil
	}

	h.deps = Deps{
		Inventory: func(context.Context) ([]inventory.Node, error) { return nodes, nil },
		Observe: func(context.Context) (map[string]inventory.Observation, error) {
			h.mu.Lock()
			defer h.mu.Unlock()
			if len(h.observed) == 0 {
				return nil, nil
			}
			out := make(map[string]inventory.Observation, len(h.observed))
			for k, v := range h.observed {
				out[k] = v
			}
			return out, nil
		},
		// Required by the power operations this manager also serves; an
		// upgrade never reaches it, and says so if it does.
		Inspect: func(context.Context, inventory.Node, Credentials) (nodestatus.Status, error) {
			t.Error("an upgrade read a node's status instead of its readiness")
			return nodestatus.Status{}, nil
		},
		LocalPowerSupport: func(Type) error { return nil },
		HostBootID:        func() (string, error) { return "boot", nil },
		PowerSelf: func(context.Context, Type) error {
			t.Error("an upgrade powered a machine")
			return nil
		},
		Dispatch: func(context.Context, []inventory.Node, PeerRequest, Credentials) []DispatchOutcome {
			t.Error("an upgrade dispatched a power command")
			return nil
		},

		Upgrade: &UpgradeDeps{
			Plan:  func(context.Context) (UpgradePlan, error) { return h.plan, nil },
			Auth:  func(context.Context, string) (string, error) { return "test-token", nil },
			Local: &harnessRunner{h: h, node: selfNameOf(nodes), start: start},
			Start: func(_ context.Context, n inventory.Node, req UpgradeStageRequest, token string) (UpgradeStageState, error) {
				if token != "test-token" {
					return UpgradeStageState{}, fmt.Errorf("worker got token %q", token)
				}
				return start(n.NodeName, req)
			},
			Status: func(_ context.Context, n inventory.Node, opID, stageName, _ string) (UpgradeStageState, error) {
				h.mu.Lock()
				defer h.mu.Unlock()
				key := stageKey(n.NodeName, stageName)
				if h.restartPending[key] {
					delete(h.restartPending, key)
					return UpgradeStageState{
						OperationID: opID, Stage: stageName,
						Phase: UpgradeStagePhaseFailed, Code: CodeDaemonRestarted,
					}, nil
				}
				state, ok := h.stageStates[key]
				if !ok {
					return UpgradeStageState{}, errors.New("no such stage")
				}
				return *state, nil
			},
			// By default every node is an olaresd that understands stages and
			// holds the version being rolled out. The tests that care about
			// the opposite replace this.
			Readiness: func(_ context.Context, n inventory.Node, _, _ string) (UpgradeReadiness, error) {
				h.mu.Lock()
				defer h.mu.Unlock()
				if h.unreachable[n.NodeName] {
					return UpgradeReadiness{}, errors.New("node did not answer")
				}
				if r, ok := h.readiness[n.NodeName]; ok {
					return r, nil
				}
				return readyOn(h.plan.Version), nil
			},
		},

		Sleep:    func(context.Context, time.Duration) error { return nil },
		Timeouts: Timeouts{Poll: time.Millisecond, Stage: time.Minute},
	}
	return h
}

func selfNameOf(nodes []inventory.Node) string {
	for _, n := range nodes {
		if n.IsSelf {
			return n.NodeName
		}
	}
	return ""
}

// harnessRunner stands in for the control node's own stage runner.
type harnessRunner struct {
	h     *upgradeHarness
	node  string
	start func(string, UpgradeStageRequest) (UpgradeStageState, error)
}

func (r *harnessRunner) Start(_ context.Context, req UpgradeStageRequest) (UpgradeStageState, error) {
	return r.start(r.node, req)
}

func (r *harnessRunner) Status(operationID, stageName string) (UpgradeStageState, bool) {
	r.h.mu.Lock()
	defer r.h.mu.Unlock()
	state, ok := r.h.stageStates[stageKey(r.node, stageName)]
	if !ok {
		return UpgradeStageState{}, false
	}
	return *state, true
}

func (h *upgradeHarness) order() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.started...)
}

func twoNodeCluster() []inventory.Node {
	return []inventory.Node{
		{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.12", Ready: true},
		{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.11", Ready: true, IsSelf: true},
	}
}

func samplePlan() UpgradePlan {
	return UpgradePlan{
		Version: "1.12.8",
		Stages: []UpgradeStage{
			{Name: "01-master", Placement: PlacementAdmin, Tasks: []string{"PublishMarketEnsureApps"}},
			{Name: "02-all-nodes", Placement: PlacementAllNodes, MaxParallel: 1, Tasks: []string{"MigrateContainerdConfigV3"}},
			{Name: "03-master", Placement: PlacementAdmin, Tasks: []string{"UpgradeSystemComponents"}},
		},
	}
}

// newTestStore is an operation store in a directory this test owns.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStore(filepath.Join(t.TempDir(), "operations"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func runUpgradeOperation(t *testing.T, h *upgradeHarness) Operation {
	t.Helper()
	h.deps.Store = newTestStore(t)
	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "req-1", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return waitForTerminal(t, m, op.ID)
}

func waitForTerminal(t *testing.T, m *Manager, id string) Operation {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		op, ok := m.Get(id)
		if ok && op.Status.Terminal() {
			return op
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("operation %s never settled", id)
	return Operation{}
}

// The barrier is the whole point: a stage does not start on any node until the
// previous stage has finished on every node it was scheduled on.
func TestUpgradeRunsStagesInOrderWorkersBeforeAdmin(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	op := runUpgradeOperation(t, h)

	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}
	want := []string{
		// Every compute node fetches the release before any of the plan runs.
		// The control node is absent: the upgrade watcher installed its
		// binaries before this orchestrator existed to run.
		"worker-1/" + StageNodePrepare,
		"master-1/01-master",
		// The all-nodes stage puts the compute node first: the control node
		// is running the orchestrator and may restart olaresd.
		"worker-1/02-all-nodes",
		"master-1/02-all-nodes",
		"master-1/03-master",
	}
	got := h.order()
	if len(got) != len(want) {
		t.Fatalf("dispatched %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dispatched %v, want %v", got, want)
		}
	}
}

// Each stage records its own per-node results, so a node that failed at one
// stage is still recorded as having failed there after the run moves on.
func TestUpgradeRecordsPerStagePerNodeResults(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	op := runUpgradeOperation(t, h)

	byName := map[string]Step{}
	for _, s := range op.Steps {
		byName[s.Name] = s
	}
	stage, ok := byName["02-all-nodes"]
	if !ok {
		t.Fatalf("no record of stage 02-all-nodes in %+v", op.Steps)
	}
	if stage.Placement != string(PlacementAllNodes) || stage.MaxParallel != 1 {
		t.Errorf("stage fanout = %q maxParallel = %d", stage.Placement, stage.MaxParallel)
	}
	if len(stage.Nodes) != 2 {
		t.Fatalf("stage recorded %d nodes, want both", len(stage.Nodes))
	}
	for _, n := range stage.Nodes {
		if n.Status != NodeSucceeded {
			t.Errorf("node %s = %s (%s)", n.NodeName, n.Status, n.Code)
		}
	}
}

// A failed stage stops the upgrade there. Nothing later runs, and the record
// says partially_failed because earlier stages really did change the cluster.
func TestUpgradeStopsAtAFailedStage(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.failAt[stageKey("worker-1", "02-all-nodes")] = true

	op := runUpgradeOperation(t, h)

	if op.Status != StatusPartiallyFailed {
		t.Fatalf("status = %s, want partially_failed (%s: %s)", op.Status, op.Code, op.Error)
	}
	for _, key := range h.order() {
		if key == "master-1/03-master" {
			t.Error("the upgrade carried on to a later stage after one failed")
		}
	}
	// The control node was never asked to run the stage the worker failed:
	// serial means stop, not carry on into the next machine.
	for _, key := range h.order() {
		if key == "master-1/02-all-nodes" {
			t.Error("a serial stage carried on to the next node after a failure")
		}
	}
}

// A stage with a limit takes that many nodes at a time and no more.
func TestBoundedStageRunsAtMostMaxParallelNodes(t *testing.T) {
	nodes := []inventory.Node{
		{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.12", Ready: true},
		{NodeName: "worker-2", Role: inventory.RoleWorker, IP: "10.0.0.13", Ready: true},
		{NodeName: "worker-3", Role: inventory.RoleWorker, IP: "10.0.0.14", Ready: true},
		{NodeName: "worker-4", Role: inventory.RoleWorker, IP: "10.0.0.15", Ready: true},
		{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.11", Ready: true, IsSelf: true},
	}
	plan := UpgradePlan{
		Version: "1.12.8",
		Stages: []UpgradeStage{
			{Name: "01-all-nodes", Placement: PlacementAllNodes, MaxParallel: 2,
				Tasks: []string{"MigrateContainerdConfigV3"}},
		},
	}

	h := newUpgradeHarness(t, plan, nodes)

	var mu sync.Mutex
	var inFlight, peak int

	// Each node waits inside the stage until a second one joins it, so the
	// result does not depend on how quickly goroutines get scheduled: with a
	// limit of two both arrive and leave at once, and with a limit of one the
	// wait times out and the peak stays at one.
	h.onStart = func(_, stage string) {
		if stage != "01-all-nodes" {
			return
		}
		mu.Lock()
		inFlight++
		if inFlight > peak {
			peak = inFlight
		}
		mu.Unlock()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			joined := inFlight >= 2
			mu.Unlock()
			if joined {
				break
			}
			time.Sleep(time.Millisecond)
		}

		mu.Lock()
		inFlight--
		mu.Unlock()
	}

	if op := runUpgradeOperation(t, h); op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}

	mu.Lock()
	defer mu.Unlock()
	t.Logf("peak=%d started=%v", peak, h.order())
	if peak > 2 {
		t.Errorf("%d nodes were inside a maxParallel=2 stage at once", peak)
	}
	if peak < 2 {
		t.Errorf("peak concurrency was %d: the limit was not being used at all", peak)
	}
}

// An upgrade that failed can be asked for again.
//
// Its request id comes from the target version, so the same id arrives every
// time the watcher retries. Without RetryAfterFailure the first failure would
// be handed back for ever and the only way out would be deleting the record
// off disk.
func TestFailedUpgradeIsRetriedUnderTheSameRequestID(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.failAt[stageKey("worker-1", "02-all-nodes")] = true
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	req := CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}

	first, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	failed := waitForTerminal(t, m, first.ID)
	if failed.Status != StatusPartiallyFailed {
		t.Fatalf("first run = %s, want partially_failed", failed.Status)
	}

	// Whatever broke has been dealt with; the same request arrives again.
	h.mu.Lock()
	delete(h.failAt, stageKey("worker-1", "02-all-nodes"))
	h.mu.Unlock()

	second, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if second.ID == first.ID {
		t.Fatal("the retry was answered with the failed operation instead of a new one")
	}
	if got := waitForTerminal(t, m, second.ID); got.Status != StatusSucceeded {
		t.Fatalf("retry = %s (%s: %s)", got.Status, got.Code, got.Error)
	}

	// The failure is still on record, and the request id now finds the retry.
	if _, ok := m.Get(first.ID); !ok {
		t.Error("the superseded operation was discarded")
	}
	if op, ok := m.GetByRequest(req.RequestID); !ok || op.ID != second.ID {
		t.Errorf("request id resolves to %+v, want the retry %s", op.ID, second.ID)
	}
}

// A cancellation outlives the daemon that was asked for it.
//
// Cancelling an upgrade and having it carry on is the one outcome a cancel
// must not produce, and an upgrade restarts olaresd as part of its own work —
// so a stop held only in the manager's memory would be forgotten by the
// process that resumes the run.
func TestAStopRequestSurvivesADaemonRestart(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	// Asked while a node is inside the first stage, so the record is
	// certainly still running: a stop is only meaningful before the operation
	// settles, and a test that raced it into a settled one would prove
	// nothing.
	inside := make(chan struct{}, 1)
	var once sync.Once
	h.onStart = func(_, _ string) {
		once.Do(func() { inside <- struct{}{} })
	}

	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	select {
	case <-inside:
	case <-time.After(5 * time.Second):
		t.Fatal("no stage ever started")
	}
	if !m.RequestStop(op.ID) {
		t.Fatal("a running upgrade refused to be stopped")
	}

	settled := waitForTerminal(t, m, op.ID)
	if settled.Code != CodeUpgradeCancelled {
		t.Errorf("stopped upgrade settled as %s/%s, want %s",
			settled.Status, settled.Code, CodeUpgradeCancelled)
	}

	restarted, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("restart: %v", err)
	}
	if !restarted.StopRequested(op.ID) {
		t.Error("the daemon that came back did not know the upgrade had been cancelled")
	}
}

// A cancelled upgrade stays cancelled.
//
// Every other way an upgrade ends is worth another attempt, and the watcher
// makes one on a timer. A cancel is not: the operator asked for it to stop,
// and answering that by starting the whole thing again is the one outcome a
// cancel must not produce. Removing the upgrade target normally stops the
// watcher too, so on the real path both halves agree — this pins the half that
// does not depend on the other one having worked, which is what a cluster
// upgrade did when it was cancelled on a live two-node cluster: it settled as
// cancelled and the next tick started it over.
func TestACancelledUpgradeIsNotStartedAgain(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	inside := make(chan struct{}, 1)
	var once sync.Once
	h.onStart = func(_, _ string) { once.Do(func() { inside <- struct{}{} }) }

	req := CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}
	op, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	select {
	case <-inside:
	case <-time.After(5 * time.Second):
		t.Fatal("no stage ever started")
	}
	if !m.RequestStop(op.ID) {
		t.Fatal("a running upgrade refused to be stopped")
	}
	settled := waitForTerminal(t, m, op.ID)
	if settled.Code != CodeUpgradeCancelled {
		t.Fatalf("settled as %s/%s, want %s", settled.Status, settled.Code, CodeUpgradeCancelled)
	}

	// The watcher's next tick.
	again, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if again.ID != op.ID {
		t.Fatalf("the cancelled upgrade was started again as %s", again.ID)
	}
	if again.Code != CodeUpgradeCancelled {
		t.Errorf("second Create answered with %s/%s, want the cancelled record",
			again.Status, again.Code)
	}
}

// One step per name, however many attempts it took.
//
// A resumed run walks its steps again, and appending a second record for each
// left the operation describing two of everything: one abandoned mid-flight,
// one real. It was not merely untidy. The writers searched the step list from
// opposite ends, so the per-node results of the resumed stage were written
// onto the abandoned copy while the live one still reported every node as
// pending — which is exactly what a two-node cluster showed after its
// orchestrator was killed in the middle of a stage.
func TestAResumedStageKeepsOneStepPerName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	at := time.Now()
	interrupted := Operation{
		ID: "op-onestep-1", Type: TypeUpgrade, RequestID: "req-1", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
		Status: StatusRunning, CreatedAt: at, UpdatedAt: at, StartedAt: &at,
		Steps: []Step{
			{Name: StepPrecheck, Status: StepSucceeded},
			{Name: StepPlan, Status: StepSucceeded},
			// Caught mid-flight, with the nodes it had reached recorded.
			{
				Name: "02-all-nodes", Status: StepRunning,
				Placement: string(PlacementAllNodes),
				Nodes: []NodeResult{
					{NodeName: "worker-1", Status: NodeRunning},
					{NodeName: "master-1", Status: NodePending},
				},
			},
		},
	}
	if err := store.Save(interrupted); err != nil {
		t.Fatalf("Save: %v", err)
	}

	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = store

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op := waitForTerminal(t, m, interrupted.ID)
	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}

	seen := map[string]int{}
	for _, s := range op.Steps {
		seen[s.Name]++
	}
	for name, n := range seen {
		if n > 1 {
			t.Errorf("step %q recorded %d times, want once", name, n)
		}
	}

	var stage *Step
	for i := range op.Steps {
		if op.Steps[i].Name == "02-all-nodes" {
			stage = &op.Steps[i]
		}
	}
	if stage == nil {
		t.Fatal("the resumed stage is not on the record at all")
	}
	if stage.Status != StepSucceeded {
		t.Errorf("resumed stage = %s, want succeeded", stage.Status)
	}
	for _, n := range stage.Nodes {
		if n.Status != NodeSucceeded {
			t.Errorf("node %s = %s on the stage that succeeded, want succeeded",
				n.NodeName, n.Status)
		}
	}
}

// A daemon restarts after a retry and comes back.
//
// The retry leaves two records for one request id, and everything that reads
// that id assumes at most one: the loader refused to start at all, which — now
// that a daemon unable to record operations exits — meant olaresd crash-looped
// from the moment anybody retried an upgrade. The superseded record has to
// stop claiming the request before the retry starts claiming it, on disk and
// not only in memory.
func TestDaemonStartsAgainAfterAnUpgradeWasRetried(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.failAt[stageKey("worker-1", "02-all-nodes")] = true
	store := newTestStore(t)
	h.deps.Store = store

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	req := CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}

	first, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitForTerminal(t, m, first.ID)

	h.mu.Lock()
	delete(h.failAt, stageKey("worker-1", "02-all-nodes"))
	h.mu.Unlock()

	second, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	waitForTerminal(t, m, second.ID)

	// The daemon restarts over the records both attempts left behind.
	restarted, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("a daemon that had retried an upgrade could not start: %v", err)
	}
	op, ok := restarted.GetByRequest(req.RequestID)
	if !ok {
		t.Fatal("the request id resolves to nothing after a restart")
	}
	if op.ID != second.ID {
		t.Errorf("request id resolves to %s, want the retry %s", op.ID, second.ID)
	}
	if _, ok := restarted.Get(first.ID); !ok {
		t.Error("the superseded attempt was lost across the restart")
	}
}

// A succeeded upgrade is not run again just because someone asked twice.
func TestSucceededUpgradeIsNotRerun(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	req := CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}

	first, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := waitForTerminal(t, m, first.ID); got.Status != StatusSucceeded {
		t.Fatalf("first run = %s", got.Status)
	}

	second, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if second.ID != first.ID {
		t.Error("a completed upgrade was started again")
	}
}

// Nothing is dispatched at all when a node the upgrade needs is not Ready:
// skipping it would leave the cluster running two versions with nothing
// recording which node is on which.
func TestUpgradeRefusesANotReadyNode(t *testing.T) {
	nodes := twoNodeCluster()
	nodes[0].Ready = false
	h := newUpgradeHarness(t, samplePlan(), nodes)

	op := runUpgradeOperation(t, h)

	if op.Status != StatusFailed || op.Code != CodeNodeNotReady {
		t.Fatalf("status = %s code = %s, want failed/%s", op.Status, op.Code, CodeNodeNotReady)
	}
	if len(h.order()) != 0 {
		t.Errorf("dispatched %v after refusing the precheck", h.order())
	}
}

// A workers-scoped stage on a single-node cluster is recorded as skipped, not
// dropped: a plan whose stages disappear cannot be checked against the plan it
// came from.
func TestUpgradeSkipsAWorkerStageOnASingleNode(t *testing.T) {
	plan := samplePlan()
	plan.Stages = append(plan.Stages, UpgradeStage{
		Name: "04-workers", Placement: PlacementWorkers, Tasks: []string{"RestartAgent"},
	})
	single := []inventory.Node{
		{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.11", Ready: true, IsSelf: true},
	}
	h := newUpgradeHarness(t, plan, single)

	op := runUpgradeOperation(t, h)

	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}
	var found bool
	for _, s := range op.Steps {
		if s.Name == "04-workers" {
			found = true
			if s.Status != StepSkipped {
				t.Errorf("stage 04-workers = %s, want skipped", s.Status)
			}
		}
	}
	if !found {
		t.Error("the worker stage vanished from the record instead of being skipped")
	}
}

// A node that cannot be reached to start a stage fails that stage rather than
// being quietly left behind.
func TestUpgradeFailsAnUnreachableNode(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.unreachable["worker-1"] = true

	op := runUpgradeOperation(t, h)

	if op.Status != StatusPartiallyFailed && op.Status != StatusFailed {
		t.Fatalf("status = %s, want a failure", op.Status)
	}
	for _, s := range op.Steps {
		if s.Name != "02-all-nodes" {
			continue
		}
		for _, n := range s.Nodes {
			if n.NodeName == "worker-1" && n.Status != NodeFailed {
				t.Errorf("unreachable node recorded as %s", n.Status)
			}
		}
	}
}

// An olaresd from before staged upgrades existed serves no stage route and
// serves no readiness route at all. The upgrade has to be refused on that,
// before any stage runs — dialling it later would 404 in the middle, with
// earlier stages already applied.
func TestUpgradeRefusesANodeThatCannotRunStages(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	// An older olaresd: no readiness route, so the probe fails at the wire.
	h.deps.Upgrade.Readiness = func(_ context.Context, n inventory.Node, _, _ string) (UpgradeReadiness, error) {
		if n.NodeName == "worker-1" {
			return UpgradeReadiness{}, errors.New("node answered 404 Not Found")
		}
		return readyOn(h.plan.Version), nil
	}

	op := runUpgradeOperation(t, h)

	if op.Status != StatusFailed || op.Code != CodeUpgradeUnsupported {
		t.Fatalf("status = %s code = %s, want failed/%s", op.Status, op.Code, CodeUpgradeUnsupported)
	}
	if len(h.order()) != 0 {
		t.Errorf("dispatched %v before finding out a node could not take part", h.order())
	}
}

// A node that still reports the previous olares-cli after it was prepared is
// refused before any stage of the plan runs. It would otherwise resolve every
// stage name against the previous version's plan, and nothing downstream
// compares the two.
func TestUpgradeRefusesANodeHoldingAnotherVersionAfterPrepare(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.readiness["worker-1"] = readyOn("1.12.7")

	op := runUpgradeOperation(t, h)

	if op.Status != StatusFailed || op.Code != CodeVersionMismatch {
		t.Fatalf("status = %s code = %s, want failed/%s", op.Status, op.Code, CodeVersionMismatch)
	}
	// Preparing it is expected and is what the check is checking; running any
	// of the plan on it is not.
	for _, key := range h.order() {
		if !strings.HasSuffix(key, StageNodePrepare) {
			t.Errorf("dispatched plan work %q to a node that is not on the target version", key)
		}
	}
}

// A node that could not be brought to the target release stops the upgrade
// before any of the plan runs anywhere.
func TestUpgradeThatFailsPreparingANodeIsNotPartiallyApplied(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.failAt[stageKey("worker-1", StageNodePrepare)] = true

	op := runUpgradeOperation(t, h)

	if op.Status != StatusFailed {
		t.Fatalf("status = %s, want failed (%s: %s)", op.Status, op.Code, op.Error)
	}
	for _, key := range h.order() {
		if !strings.HasSuffix(key, StageNodePrepare) {
			t.Errorf("dispatched plan work %q after failing to prepare a node", key)
		}
	}
}

// A node that declares stages but will not say which olares-cli it holds is
// refused.
//
// It used to be allowed through, on the grounds that it had made no claim to
// contradict and that the plan digest would catch it at the stage. There is no
// digest now: comparing versions is the only thing establishing that a stage
// name means the same work on two machines, so a node that does not answer
// has not passed the only check there is.
func TestUpgradeRefusesANodeThatDoesNotReportItsCLIVersion(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.readiness["worker-1"] = UpgradeReadiness{Supported: true}

	op := runUpgradeOperation(t, h)
	if op.Status != StatusFailed || op.Code != CodeVersionMismatch {
		t.Fatalf("status = %s code = %s, want failed/%s", op.Status, op.Code, CodeVersionMismatch)
	}
	for _, key := range h.order() {
		if !strings.HasSuffix(key, StageNodePrepare) {
			t.Errorf("dispatched plan work %q to a node that would not state its version", key)
		}
	}
}

// A single-node cluster has no hop to check, so none of the above applies: the
// control node runs its own stages in process.
func TestUpgradeOnASingleNodeNeedsNoReadinessCheck(t *testing.T) {
	single := []inventory.Node{
		{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.11", Ready: true, IsSelf: true},
	}
	h := newUpgradeHarness(t, samplePlan(), single)
	h.deps.Upgrade.Readiness = func(context.Context, inventory.Node, string, string) (UpgradeReadiness, error) {
		t.Error("a single-node upgrade probed a node over the network")
		return UpgradeReadiness{}, nil
	}
	h.deps.Upgrade.Auth = func(context.Context, string) (string, error) {
		t.Error("a single-node upgrade minted a token for a hop that does not exist")
		return "", nil
	}

	op := runUpgradeOperation(t, h)
	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}
}

// The token that authorizes a worker is not the owner's signature: no
// credential from the create request reaches the hop.
func TestUpgradeAuthorizesWorkersWithTheOperationToken(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())

	var sawToken string
	h.deps.Upgrade.Start = func(_ context.Context, n inventory.Node, req UpgradeStageRequest, token string) (UpgradeStageState, error) {
		sawToken = token
		h.mu.Lock()
		defer h.mu.Unlock()
		h.started = append(h.started, stageKey(n.NodeName, req.Stage))
		state := &UpgradeStageState{OperationID: req.OperationID, Stage: req.Stage, Phase: UpgradeStagePhaseSucceeded}
		h.stageStates[stageKey(n.NodeName, req.Stage)] = state
		return *state, nil
	}

	op := runUpgradeOperation(t, h)
	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}
	if sawToken != "test-token" {
		t.Errorf("worker was given %q, want the operation token", sawToken)
	}
}

// An upgrade must not begin while a reboot is in flight, and vice versa. They
// share one lock because they are the same kind of thing to the cluster.
func TestUpgradeAndPowerOperationsExcludeEachOther(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = newTestStore(t)
	// Hold the first operation open so the second one meets it.
	release := make(chan struct{})
	h.deps.Upgrade.Plan = func(context.Context) (UpgradePlan, error) {
		<-release
		return h.plan, nil
	}

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer close(release)

	if _, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "req-1", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}); err != nil {
		t.Fatalf("Create upgrade: %v", err)
	}

	_, err = m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "req-2", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("a reboot was accepted during an upgrade: %v", err)
	}
}

// While an upgrade is running the cluster is in maintenance, which is what the
// phase already means. It is not "restarting": nothing has gone down.
func TestUpgradeReportsMaintenancePhase(t *testing.T) {
	op := &Operation{Type: TypeUpgrade, Status: StatusRunning}
	phase, ok := PhaseFor(op)
	if !ok || string(phase) != "maintenance" {
		t.Fatalf("phase = %q ok = %t, want maintenance", phase, ok)
	}
}

// A node that restarted running its own stage is asked again, in the same
// operation.
//
// Several stages are meant to restart olaresd on the node running them:
// node-prepare installs the new daemon, reboot-nodes takes the machine down.
// The node marks its own record failed on the way back up, because nothing
// about the half-finished run can be recovered — but the stage can be, since
// every task in one is reentrant. Reading that as a failure meant the first
// real upgrade of any worker failed at node-prepare and only got through
// because the watcher started a whole new operation.
func TestNodeRestartedByItsOwnStageIsAskedAgain(t *testing.T) {
	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.restartOnce[stageKey("worker-1", "02-all-nodes")] = true
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	settled := waitForTerminal(t, m, op.ID)
	if settled.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s), want the restart to have been ridden out",
			settled.Status, settled.Code, settled.Error)
	}
	if settled.ID != op.ID {
		t.Errorf("the upgrade continued as %s, want the same operation %s", settled.ID, op.ID)
	}

	// Asked twice, and only that node.
	var asked int
	for _, key := range h.order() {
		if key == stageKey("worker-1", "02-all-nodes") {
			asked++
		}
	}
	if asked != 2 {
		t.Errorf("the restarted stage was dispatched %d times, want 2", asked)
	}

	step := stepNamed(t, settled, "02-all-nodes")
	for _, n := range step.Nodes {
		if n.Status != NodeSucceeded {
			t.Errorf("node %s = %s/%s, want succeeded", n.NodeName, n.Status, n.Code)
		}
	}
}

// A stage whose tasks may take the machine down is not over when olares-cli
// exits.
//
// The reboot task issues the command and returns a moment before the machine
// stops answering. Without a wait the orchestrator reads that as the node
// being finished: on a serial stage it reboots the next node while the
// previous one is still coming up, and an upgrade whose last node never came
// back still ends as succeeded — which is what a two-node cluster did when
// its worker failed to boot.
func TestARebootStageWaitsForNodesToComeBack(t *testing.T) {
	plan := samplePlan()
	for i := range plan.Stages {
		if plan.Stages[i].Name == "02-all-nodes" {
			plan.Stages[i].AwaitRestart = true
		}
	}
	h := newUpgradeHarness(t, plan, twoNodeCluster())
	// Both nodes are up on the boot they started on, and stay there: nothing
	// went down, so nothing has to come back.
	h.observed = map[string]inventory.Observation{
		"worker-1": {BootID: "boot-a", Ready: true},
		"master-1": {BootID: "boot-a", Ready: true},
	}
	h.deps.Timeouts = Timeouts{Poll: time.Millisecond, Stage: time.Minute,
		Down: 10 * time.Millisecond, Ready: time.Second}
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := waitForTerminal(t, m, op.ID); got.Status != StatusSucceeded {
		t.Fatalf("a stage that rebooted nothing = %s (%s: %s)", got.Status, got.Code, got.Error)
	}
}

// The same stage, with a node that goes down and does not return.
func TestARebootStageFailsANodeThatNeverReturns(t *testing.T) {
	plan := samplePlan()
	for i := range plan.Stages {
		if plan.Stages[i].Name == "02-all-nodes" {
			plan.Stages[i].AwaitRestart = true
		}
	}
	h := newUpgradeHarness(t, plan, twoNodeCluster())
	// worker-1 has left the cluster's view and never comes back; master-1
	// never went down.
	h.observed = map[string]inventory.Observation{
		"master-1": {BootID: "boot-a", Ready: true},
	}
	h.deps.Timeouts = Timeouts{Poll: time.Millisecond, Stage: time.Minute,
		Down: time.Millisecond, Ready: 50 * time.Millisecond}
	h.deps.Store = newTestStore(t)

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := waitForTerminal(t, m, op.ID)
	if settled.Status == StatusSucceeded {
		t.Fatal("the upgrade succeeded with a node that never came back from its reboot")
	}

	step := stepNamed(t, settled, "02-all-nodes")
	var found bool
	for _, n := range step.Nodes {
		if n.NodeName == "worker-1" && n.Code == CodeRestartTimeout {
			found = true
		}
	}
	if !found {
		t.Errorf("worker-1 is not recorded as having failed to come back: %+v", step.Nodes)
	}
}

func stepNamed(t *testing.T, op Operation, name string) Step {
	t.Helper()
	for i := len(op.Steps) - 1; i >= 0; i-- {
		if op.Steps[i].Name == name {
			return op.Steps[i]
		}
	}
	t.Fatalf("no step named %q on the record", name)
	return Step{}
}

// A serial stage that reboots waits for one machine before touching the next.
//
// The wait has to live inside the node's slot. Put it after every node has
// been given the stage instead and MaxParallel stops meaning anything for the
// one kind of work it matters most for: every node is told to reboot, they all
// go down together, and only then does anything start counting them back.
func TestASerialRebootStageWaitsBeforeTheNextNode(t *testing.T) {
	plan := samplePlan()
	for i := range plan.Stages {
		if plan.Stages[i].Name == "02-all-nodes" {
			plan.Stages[i].AwaitRestart = true
			plan.Stages[i].MaxParallel = 1
		}
	}
	h := newUpgradeHarness(t, plan, twoNodeCluster())
	h.observed = map[string]inventory.Observation{
		"worker-1": {BootID: "boot-a", Ready: true},
		"master-1": {BootID: "boot-a", Ready: true},
	}
	// Down is the window in which a node that is going to reboot has to
	// disappear; master-1 never does, so it is what that node waits out.
	h.deps.Timeouts = Timeouts{Poll: time.Millisecond, Stage: time.Minute,
		Down: 200 * time.Millisecond, Ready: 5 * time.Second}
	h.deps.Store = newTestStore(t)

	var mu sync.Mutex
	workerBack := false
	startedAfterWorkerBack := true

	h.onStart = func(node, stage string) {
		if stage != "02-all-nodes" {
			return
		}
		switch node {
		case "worker-1":
			// It reboots: gone from the cluster's view, then back on a new
			// boot a moment later.
			h.mu.Lock()
			delete(h.observed, "worker-1")
			h.mu.Unlock()
			go func() {
				time.Sleep(30 * time.Millisecond)
				h.mu.Lock()
				h.observed["worker-1"] = inventory.Observation{BootID: "boot-b", Ready: true}
				h.mu.Unlock()
				mu.Lock()
				workerBack = true
				mu.Unlock()
			}()
		case "master-1":
			mu.Lock()
			startedAfterWorkerBack = workerBack
			mu.Unlock()
		}
	}

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := waitForTerminal(t, m, op.ID); got.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", got.Status, got.Code, got.Error)
	}

	mu.Lock()
	defer mu.Unlock()
	if !startedAfterWorkerBack {
		t.Error("the control node was given the reboot stage while the compute node was still down")
	}
}

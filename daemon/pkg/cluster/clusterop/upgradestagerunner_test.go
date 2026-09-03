package clusterop

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newRunner(t *testing.T, dir string, exec UpgradeStageExec) *LocalUpgradeStageRunner {
	t.Helper()
	r, err := NewLocalUpgradeStageRunner(dir, exec)
	if err != nil {
		t.Fatalf("NewLocalUpgradeStageRunner: %v", err)
	}
	return r
}

func sampleStageRequest() UpgradeStageRequest {
	return UpgradeStageRequest{
		OperationID: "op-1", Stage: "02-all-nodes",
		Version: "1.12.8",
	}
}

// waitForStage polls the runner's own record, which is the only place the
// answer lives.
func waitForStage(t *testing.T, r *LocalUpgradeStageRunner, req UpgradeStageRequest) UpgradeStageState {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		state, ok := r.Status(req.OperationID, req.Stage)
		if ok && state.Phase.Terminal() {
			return state
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("the stage never settled")
	return UpgradeStageState{}
}

func TestUpgradeStageRunnerRecordsSuccess(t *testing.T) {
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error { return nil })
	req := sampleStageRequest()

	state, err := r.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if state.Phase != UpgradeStagePhaseRunning {
		t.Errorf("Start returned %s, want running: it must not block on the work", state.Phase)
	}
	if got := waitForStage(t, r, req); got.Phase != UpgradeStagePhaseSucceeded {
		t.Errorf("stage = %s (%s)", got.Phase, got.Code)
	}
}

// The detail of a failure stays on the machine. What the control node reads is
// the stable code and its fixed text.
func TestUpgradeStageRunnerSuppressesFailureDetail(t *testing.T) {
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error {
		return errors.New("helm upgrade os-framework failed at /root/.olares/wizard/config")
	})
	req := sampleStageRequest()
	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}

	got := waitForStage(t, r, req)
	if got.Phase != UpgradeStagePhaseFailed || got.Code != CodeStageFailed {
		t.Fatalf("stage = %s/%s", got.Phase, got.Code)
	}
	if got.Error != reasonFor(CodeStageFailed) {
		t.Errorf("error text = %q, want the fixed text for the code", got.Error)
	}
}

// Asked twice for a stage it already finished, the runner says so instead of
// doing the work again. This is what makes a resumed upgrade skip ahead.
func TestUpgradeStageRunnerDoesNotRepeatFinishedWork(t *testing.T) {
	var mu sync.Mutex
	var runs int
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error {
		mu.Lock()
		runs++
		mu.Unlock()
		return nil
	})
	req := sampleStageRequest()

	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}
	waitForStage(t, r, req)

	state, err := r.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("second Start: %v", err)
	}
	if state.Phase != UpgradeStagePhaseSucceeded {
		t.Errorf("second Start = %s, want the finished record", state.Phase)
	}
	mu.Lock()
	defer mu.Unlock()
	if runs != 1 {
		t.Errorf("the stage ran %d times", runs)
	}
}

// An upgrade stage installs the new olaresd and restarts it, killing the
// olares-cli it launched. The record left behind says "running" and describes
// work that stopped part way, so a new process must settle it rather than
// report progress that is not happening.
func TestUpgradeStageRunnerSettlesAStageInterruptedByARestart(t *testing.T) {
	// Seeded where the runner will look, through the same constant it uses,
	// so this stands in for a record the previous process left behind.
	opsDir := t.TempDir()
	store, err := newUpgradeStageStore(filepath.Join(opsDir, stageRecordsDir))
	if err != nil {
		t.Fatalf("newUpgradeStageStore: %v", err)
	}
	req := sampleStageRequest()
	if err := store.Save(UpgradeStageState{
		OperationID: req.OperationID, Stage: req.Stage,
		Phase: UpgradeStagePhaseRunning, StartedAt: time.Now(),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	restarted := newRunner(t, opsDir, func(context.Context, UpgradeStageRequest) error { return nil })

	state, ok := restarted.Status(req.OperationID, req.Stage)
	if !ok {
		t.Fatal("the record did not survive the restart")
	}
	if state.Phase != UpgradeStagePhaseFailed || state.Code != CodeDaemonRestarted {
		t.Fatalf("stage = %s/%s, want failed/%s", state.Phase, state.Code, CodeDaemonRestarted)
	}
}

// Having been settled as interrupted, the same stage runs again when asked:
// upgrade tasks are reentrant, and re-running is the only way forward.
func TestUpgradeStageRunnerRetriesAFailedStage(t *testing.T) {
	// Seeded where the runner will look, through the same constant it uses,
	// so this stands in for a record the previous process left behind.
	opsDir := t.TempDir()
	store, err := newUpgradeStageStore(filepath.Join(opsDir, stageRecordsDir))
	if err != nil {
		t.Fatalf("newUpgradeStageStore: %v", err)
	}
	req := sampleStageRequest()
	if err := store.Save(UpgradeStageState{
		OperationID: req.OperationID, Stage: req.Stage, Phase: UpgradeStagePhaseFailed,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var ran bool
	var mu sync.Mutex
	r := newRunner(t, opsDir, func(context.Context, UpgradeStageRequest) error {
		mu.Lock()
		ran = true
		mu.Unlock()
		return nil
	})
	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}
	waitForStage(t, r, req)

	mu.Lock()
	defer mu.Unlock()
	if !ran {
		t.Error("a failed stage was not retried when the orchestrator asked again")
	}
}

// Two ids that arrive from the network must not be able to name a file
// outside the stage directory.
func TestUpgradeStageStoreKeepsHostileIdsInsideItsDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "stages")
	store, err := newUpgradeStageStore(dir)
	if err != nil {
		t.Fatalf("newUpgradeStageStore: %v", err)
	}
	state := UpgradeStageState{
		OperationID: "../../etc", Stage: "../passwd", Phase: UpgradeStagePhaseRunning,
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, ok, err := store.Load(state.OperationID, state.Stage)
	if err != nil || !ok {
		t.Fatalf("Load: %v ok=%t", err, ok)
	}
	if got.OperationID != state.OperationID {
		t.Errorf("round trip lost the ids: %+v", got)
	}

	entries, err := filepath.Glob(filepath.Join(dir, "*"+stageRecordSuffix))
	if err != nil || len(entries) != 1 {
		t.Fatalf("records = %v (%v)", entries, err)
	}
	if filepath.Dir(entries[0]) != dir {
		t.Errorf("record written to %s, outside %s", entries[0], dir)
	}
}

func TestUpgradeStageRunnerRefusesAnUnnamedStage(t *testing.T) {
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error {
		t.Error("a stage with no id was run")
		return nil
	})
	if _, err := r.Start(context.Background(), UpgradeStageRequest{OperationID: "op-1"}); err == nil {
		t.Fatal("a request naming no stage was accepted")
	}
}

// One stage at a time on a machine, whichever operation asked.
//
// Idempotency per (operation, stage) is not enough. A stage the control node
// gave up waiting for is still running here — a timeout ends the waiting, not
// the work — and the retry that follows carries a new operation id. Without
// this the node would start a second olares-cli beside the first: two helm
// upgrades, two containerd restarts, or two driver installs at once.
func TestANodeRunsOneUpgradeStageAtATime(t *testing.T) {
	release := make(chan struct{})
	var running sync.WaitGroup
	running.Add(1)

	var started int
	var mu sync.Mutex
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error {
		mu.Lock()
		started++
		mu.Unlock()
		running.Done()
		<-release
		return nil
	})
	defer close(release)

	first := sampleStageRequest()
	if _, err := r.Start(context.Background(), first); err != nil {
		t.Fatalf("Start: %v", err)
	}
	running.Wait()

	// The control node timed the first one out and retried, which makes a new
	// operation with the same stage.
	second := first
	second.OperationID = "op-2"
	state, err := r.Start(context.Background(), second)
	if err != nil {
		t.Fatalf("second Start: %v", err)
	}
	if state.Phase != UpgradeStagePhaseFailed || state.Code != CodeStageBusy {
		t.Fatalf("second stage = %s/%s, want failed/%s", state.Phase, state.Code, CodeStageBusy)
	}

	mu.Lock()
	defer mu.Unlock()
	if started != 1 {
		t.Errorf("the node ran %d stages at once, want 1", started)
	}
}

// The hold is on other work, not on the same stage being asked about again,
// which is what a resumed run does.
func TestTheSameStageMayStillBeAskedAboutWhileItRuns(t *testing.T) {
	release := make(chan struct{})
	var running sync.WaitGroup
	running.Add(1)
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error {
		running.Done()
		<-release
		return nil
	})
	defer close(release)

	req := sampleStageRequest()
	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}
	running.Wait()

	state, err := r.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("re-ask: %v", err)
	}
	if state.Phase != UpgradeStagePhaseRunning {
		t.Errorf("re-asking for the running stage got %s/%s, want running", state.Phase, state.Code)
	}
}

// A stage that leaves the machine going down records both halves of that: the
// flag, and the boot it ran on.
//
// The flag alone does not survive being read later. reboot-nodes takes the
// control node down, so its record is read again by the run that resumes after
// the machine returns, and by then "is a restart coming" has two true answers
// depending on whether you mean then or now. The boot id settles it — the
// control node compares it with the boot the cluster reports and needs no
// observation of its own from before the restart.
func TestAStageRecordsTheBootItRanOnWhenItLeavesTheMachineGoingDown(t *testing.T) {
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error { return nil })
	r.restartPending = func() bool { return true }
	r.bootID = func() (string, error) { return "boot-a", nil }

	req := sampleStageRequest()
	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}
	settled := waitForStage(t, r, req)

	if !settled.RestartPending {
		t.Error("the stage did not record that the machine was going down")
	}
	if settled.BootID != "boot-a" {
		t.Errorf("BootID = %q, want the boot the stage ran on", settled.BootID)
	}

	// The machine reboots, so the tmpfs marker is gone and asking again would
	// now answer no. The record still says yes, because it is a fact about
	// the run rather than a question re-answered on every read — and the
	// control node needs that yes to know a restart is owed at all. Deciding
	// whether it has since happened is its job, from the boot id above; a
	// node that quietly withdrew the flag would have the reboot barrier
	// released the moment olaresd answered, long before the cluster would
	// schedule anything here.
	r.restartPending = func() bool { return false }

	again, err := r.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !again.RestartPending || again.BootID != "boot-a" {
		t.Errorf("re-read as pending=%v boot=%q, want the record unchanged",
			again.RestartPending, again.BootID)
	}
}

// A stage that changes nothing about the machine records no boot: there is no
// restart to prove, and a boot id here would only invite one to be waited for.
func TestAStageThatLeavesTheMachineUpRecordsNoBoot(t *testing.T) {
	r := newRunner(t, t.TempDir(), func(context.Context, UpgradeStageRequest) error { return nil })
	r.restartPending = func() bool { return false }
	r.bootID = func() (string, error) { return "boot-a", nil }

	req := sampleStageRequest()
	if _, err := r.Start(context.Background(), req); err != nil {
		t.Fatalf("Start: %v", err)
	}
	settled := waitForStage(t, r, req)
	if settled.RestartPending || settled.BootID != "" {
		t.Errorf("recorded pending=%v boot=%q, want neither", settled.RestartPending, settled.BootID)
	}
}

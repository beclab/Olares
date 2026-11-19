package clusterop

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// An upgrade restarts olaresd on the node orchestrating it: installing the new
// daemon is one of the stages. Treating that restart the way a power operation
// treats one — settle it as failed, nothing observed how it ended — would fail
// nearly every upgrade at the point where it was working correctly.
func TestUpgradeResumesAfterTheDaemonRestarts(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// An upgrade caught mid-flight, with its first stage already done.
	at := time.Now()
	interrupted := Operation{
		ID: "op-resume-1", Type: TypeUpgrade, RequestID: "req-1", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
		Status: StatusRunning, CreatedAt: at, UpdatedAt: at, StartedAt: &at,
		Steps: []Step{
			{Name: StepPrecheck, Status: StepSucceeded},
			{Name: StepPlan, Status: StepSucceeded},
			{Name: "01-master", Status: StepSucceeded, Placement: string(PlacementAdmin)},
			{Name: "02-all-nodes", Status: StepRunning, Placement: string(PlacementAllNodes)},
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
		t.Fatalf("status = %s (%s: %s), want the upgrade to have finished", op.Status, op.Code, op.Error)
	}

	// The stage that was already done is not run again.
	for _, key := range h.order() {
		if key == "master-1/01-master" {
			t.Error("a completed stage was run a second time after the restart")
		}
	}
	// The one it was in the middle of is, and so is everything after it.
	want := map[string]bool{
		"worker-1/02-all-nodes": false,
		"master-1/02-all-nodes": false,
		"master-1/03-master":    false,
	}
	for _, key := range h.order() {
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, ran := range want {
		if !ran {
			t.Errorf("%s did not run after the restart", key)
		}
	}
}

// A power operation caught by the same restart is still failed: nothing
// watched how it ended, and unlike an upgrade there is no per-stage record
// saying what was already done.
func TestPowerOperationsAreStillFailedByADaemonRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	at := time.Now()
	if err := store.Save(Operation{
		ID: "op-power-1", Type: TypeReboot, RequestID: "req-2", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
		Status: StatusRunning, CreatedAt: at, UpdatedAt: at,
		Steps: []Step{{Name: StepWorkerCommand, Status: StepRunning}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = store
	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	op, ok := m.Get("op-power-1")
	if !ok {
		t.Fatal("the operation was not read back")
	}
	if op.Status != StatusFailed || op.Code != CodeDaemonRestarted {
		t.Fatalf("status = %s code = %s, want failed/%s", op.Status, op.Code, CodeDaemonRestarted)
	}
}

// A resumed run has no credentials — the signature that authorized the
// original request is long gone, and could not have covered an hour of work
// anyway. It re-reads the operation's token from the cluster instead.
func TestResumedUpgradeReReadsTheOperationToken(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	at := time.Now()
	if err := store.Save(Operation{
		ID: "op-resume-2", Type: TypeUpgrade, RequestID: "req-3", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
		Status: StatusRunning, CreatedAt: at, UpdatedAt: at,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = store

	var minted int
	h.deps.Upgrade.Auth = func(_ context.Context, operationID string) (string, error) {
		minted++
		if operationID != "op-resume-2" {
			t.Errorf("token asked for %q", operationID)
		}
		return "test-token", nil
	}

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op := waitForTerminal(t, m, "op-resume-2")
	if op.Status != StatusSucceeded {
		t.Fatalf("status = %s (%s: %s)", op.Status, op.Code, op.Error)
	}
	if minted != 1 {
		t.Errorf("the resumed run asked for a token %d times, want exactly one", minted)
	}
}

// A cancelled upgrade that resumes does not prepare any node first.
//
// The stop is recorded on the operation and survives the restart, but it was
// only asked about between plan stages — and the precheck runs before those,
// installing binaries and restarting olaresd on every worker. So an operator
// who cancelled got the machines touched anyway. Worse, that work failing
// settles the run as a stage failure rather than as cancelled, and a stage
// failure is retried: the cancel would be undone by the work it forbade.
func TestACancelledUpgradeResumesWithoutTouchingAnyNode(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	at := time.Now()
	cancelled := Operation{
		ID: "op-cancelled-1", Type: TypeUpgrade, RequestID: "req-1", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
		Status: StatusRunning, StopRequested: true,
		CreatedAt: at, UpdatedAt: at, StartedAt: &at,
	}
	if err := store.Save(cancelled); err != nil {
		t.Fatalf("Save: %v", err)
	}

	h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
	h.deps.Store = store

	m, err := NewManager(h.deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	op := waitForTerminal(t, m, cancelled.ID)

	if op.Code != CodeUpgradeCancelled {
		t.Errorf("resumed as %s/%s, want %s", op.Status, op.Code, CodeUpgradeCancelled)
	}
	if ran := h.order(); len(ran) > 0 {
		t.Errorf("a cancelled upgrade still reached %v", ran)
	}
}

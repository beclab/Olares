package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// testRuntimeDeps stubs every Manager dependency as a no-op, backed by a
// store rooted at dir and a fake, manually-advancing clock. Runtime tests
// never need a real cluster.
func testRuntimeDeps(t *testing.T, dir string) (Deps, *fakeClock) {
	t.Helper()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	clock := newClock()
	return Deps{
		Store: store,
		Now:   clock.Now,
		Sleep: clock.Sleep,
		Inventory: func(context.Context) ([]inventory.Node, error) {
			return nil, nil
		},
		Inspect: func(context.Context, inventory.Node, Credentials) (nodestatus.Status, error) {
			return nodestatus.Status{}, nil
		},
		Dispatch: func(context.Context, []inventory.Node, PeerRequest, Credentials) []DispatchOutcome {
			return nil
		},
		Observe: func(context.Context) (map[string]inventory.Observation, error) {
			return nil, nil
		},
		LocalPowerSupport: func(Type) error { return nil },
		HostBootID:        func() (string, error) { return "boot-1", nil },
		PowerSelf:         func(context.Context, Type) error { return nil },
	}, clock
}

// buildTestRuntime wires a Manager with every dependency stubbed to a no-op
// and seeds one operation directly into it, the same way
// TestPersistenceFailureStopsBeforePoweringTheControlNode seeds one in
// manager_test.go. A Runtime test drives state transitions on that record; it
// never needs a real cluster.
func buildTestRuntime(t *testing.T, status Status, dir string, ctx context.Context) (Runtime, *Manager, string) {
	t.Helper()
	deps, _ := testRuntimeDeps(t, dir)
	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	at := m.deps.Now()
	op := &Operation{
		ID:        "op-runtime",
		Type:      TypeReboot,
		RequestID: "request-runtime",
		Owner:     "alice@olares.com",
		Status:    status,
		CreatedAt: at,
		UpdatedAt: at,
		Steps:     []Step{},
		Nodes:     []NodeResult{{NodeName: "node-a", Role: inventory.RoleWorker, Status: NodePending}},
	}
	if status.Terminal() {
		finished := m.deps.Now()
		op.FinishedAt = &finished
	}
	if err := m.deps.Store.Save(*op); err != nil {
		t.Fatalf("Store.Save: %v", err)
	}
	m.ops[op.ID] = op
	m.order = append(m.order, op.ID)
	m.byRequest[op.RequestID] = op.ID
	if !status.Terminal() {
		m.activeID = op.ID
	}

	return newRuntime(m, op.ID, ctx), m, op.ID
}

func newTestRuntime(t *testing.T, status Status) (Runtime, *Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "operations")
	return buildTestRuntime(t, status, dir, context.Background())
}

func newTestRuntimeAt(t *testing.T, status Status, dir string) (Runtime, *Manager, string) {
	t.Helper()
	return buildTestRuntime(t, status, dir, context.Background())
}

func newTestRuntimeWithContext(t *testing.T, status Status, ctx context.Context) (Runtime, *Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "operations")
	return buildTestRuntime(t, status, dir, ctx)
}

// newRuntimeWithOperation is the minimal constructor the task brief's own
// examples use directly.
func newRuntimeWithOperation(t *testing.T, status Status) Runtime {
	t.Helper()
	rt, _, _ := newTestRuntime(t, status)
	return rt
}

// buildCommandIssuedRuntime seeds a command_issued operation with a still
// open step and node, and a CommandIssuedUntil deadline graceOffset away
// from "now" — positive to build one still inside its operationActive grace
// window, negative to build one whose grace has already expired.
func buildCommandIssuedRuntime(t *testing.T, graceOffset time.Duration) (Runtime, *Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "operations")
	deps, _ := testRuntimeDeps(t, dir)
	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	at := m.deps.Now()
	until := at.Add(graceOffset)
	op := &Operation{
		ID:                 "op-runtime",
		Type:               TypeReboot,
		RequestID:          "request-runtime",
		Owner:              "alice@olares.com",
		Status:             StatusCommandIssued,
		CreatedAt:          at,
		UpdatedAt:          at,
		FinishedAt:         &at,
		CommandIssuedUntil: until,
		Steps:              []Step{{Name: StepMasterCommand, Status: StepRunning, StartedAt: &at}},
		Nodes:              []NodeResult{{NodeName: "node-a", Role: inventory.RoleWorker, Status: NodePending}},
	}
	if err := m.deps.Store.Save(*op); err != nil {
		t.Fatalf("Store.Save: %v", err)
	}
	m.ops[op.ID] = op
	m.order = append(m.order, op.ID)
	m.byRequest[op.RequestID] = op.ID
	if m.operationActive(op, m.deps.Now()) {
		m.activeID = op.ID
	}

	return newRuntime(m, op.ID, context.Background()), m, op.ID
}

func TestRuntimeRejectsMissingStep(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	err := rt.FinishStep("missing", StepSucceeded, "", "")
	if !errors.Is(err, ErrStepNotFound) {
		t.Fatalf("FinishStep() = %v, want ErrStepNotFound", err)
	}
}

func TestRuntimeRejectsTerminalRollback(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusSucceeded)
	if err := rt.StartStep("late"); !errors.Is(err, ErrOperationTerminal) {
		t.Fatalf("StartStep() = %v, want ErrOperationTerminal", err)
	}
}

func TestRuntimeRejectsDuplicateStepCompletion(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	if err := rt.StartStep("work"); err != nil {
		t.Fatal(err)
	}
	if err := rt.FinishStep("work", StepSucceeded, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := rt.FinishStep("work", StepSucceeded, "", ""); !errors.Is(err, ErrStepSettled) {
		t.Fatalf("second FinishStep() = %v, want ErrStepSettled", err)
	}
}

// A command_issued step is the one stage a RecoverableModule exists to
// settle: the command was handed out, and what it did is only knowable
// later. Refusing to finish it would leave a confirmed operation whose own
// record still says the command is outstanding.
func TestRuntimeFinishStepSettlesACommandIssuedStep(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.StartStep(StepMasterCommand); err != nil {
		t.Fatal(err)
	}
	if err := rt.FinishStep(StepMasterCommand, StepCommandIssued, "", ""); err != nil {
		t.Fatal(err)
	}

	if err := rt.FinishStep(StepMasterCommand, StepSucceeded, "", ""); err != nil {
		t.Fatalf("FinishStep() on a command_issued step = %v, want it settled", err)
	}
	got, _ := m.Get(id)
	if got.Steps[0].Status != StepSucceeded {
		t.Fatalf("Steps[0].Status = %q, want succeeded", got.Steps[0].Status)
	}
}

// FinishedAt says when the operation settled. A command_issued record
// already carries one, and promoting it to a final status once the outcome
// is known has to move it to the moment that outcome was established — the
// alternative reports a reboot as having finished before the machine was
// even told to go down.
func TestRuntimeCompleteRestampsFinishedAtOnACommandIssuedRecord(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Hour)
	before, _ := m.Get(id)
	if before.FinishedAt == nil {
		t.Fatal("a command_issued record must already say when it settled")
	}

	if err := rt.Complete(Outcome{Status: StatusSucceeded}); err != nil {
		t.Fatal(err)
	}

	after, _ := m.Get(id)
	if after.FinishedAt == nil || !after.FinishedAt.After(*before.FinishedAt) {
		t.Fatalf("FinishedAt = %v, want the moment the reboot was confirmed (after %v)",
			after.FinishedAt, before.FinishedAt)
	}
}

func TestRuntimeUpdateNodeRejectsUnknownNode(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	err := rt.UpdateNode("missing", func(*NodeResult) {})
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("UpdateNode() = %v, want ErrNodeNotFound", err)
	}
}

func TestRuntimeUpdateNodeMutatesKnownNode(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.UpdateNode("node-a", func(n *NodeResult) { n.Status = NodeRestarted }); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(id)
	if got.Nodes[0].Status != NodeRestarted {
		t.Fatalf("Nodes[0].Status = %q, want restarted", got.Nodes[0].Status)
	}
}

// TestRuntimeUpdateNodeCallbackCanReadOperationWithoutDeadlock is I1: mutate
// must run with no manager lock held, so a module callback that calls back
// into the same Runtime (the most natural thing for it to do) must not
// deadlock against the mutation that is invoking it.
func TestRuntimeUpdateNodeCallbackCanReadOperationWithoutDeadlock(t *testing.T) {
	rt, _, _ := newTestRuntime(t, StatusRunning)
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := rt.UpdateNode("node-a", func(n *NodeResult) {
			if _, ok := rt.Operation(); !ok {
				t.Error("Operation() reported no operation from inside UpdateNode's callback")
			}
			if !rt.CanContinue() {
				t.Error("CanContinue() reported false from inside UpdateNode's callback")
			}
			n.Status = NodeRestarted
		})
		if err != nil {
			t.Errorf("UpdateNode() = %v", err)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("UpdateNode deadlocked when its callback called back into the Runtime")
	}
}

// TestRuntimeUpdateNodeCallbackPanicLeavesOperationUnchanged is I1: a
// panicking callback must not have touched manager memory or the on-disk
// record, because mutate only ever mutates a private copy before UpdateNode
// commits it.
func TestRuntimeUpdateNodeCallbackPanicLeavesOperationUnchanged(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	rt, m, id := newTestRuntimeAt(t, StatusRunning, dir)

	before, _ := m.Get(id)
	beforeDisk, ok, err := m.deps.Store.Load(id)
	if err != nil || !ok {
		t.Fatalf("Store.Load: ok=%v err=%v", ok, err)
	}

	func() {
		defer func() { _ = recover() }()
		_ = rt.UpdateNode("node-a", func(n *NodeResult) {
			panic("module bug")
		})
	}()

	after, _ := m.Get(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("a panicking callback changed in-memory state: before=%+v after=%+v", before, after)
	}
	afterDisk, ok, err := m.deps.Store.Load(id)
	if err != nil || !ok {
		t.Fatalf("Store.Load: ok=%v err=%v", ok, err)
	}
	if !reflect.DeepEqual(beforeDisk, afterDisk) {
		t.Fatalf("a panicking callback changed the on-disk record: before=%+v after=%+v", beforeDisk, afterDisk)
	}
}

// TestRuntimeUpdateNodeRejectsConcurrentChange is I1: if another checked
// mutation commits a change to the same node between this call's snapshot
// and its own commit, UpdateNode must refuse to overwrite it rather than
// silently discarding what the other writer recorded.
func TestRuntimeUpdateNodeRejectsConcurrentChange(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)

	before, err := m.snapshotNode(id, "node-a", m.rejectSettled)
	if err != nil {
		t.Fatal(err)
	}

	// Another writer commits between our snapshot and our own commit.
	winner := before
	winner.Status = NodeCommandIssued
	if err := m.replaceNode(id, "node-a", before, winner, m.rejectSettled); err != nil {
		t.Fatal(err)
	}

	loser := before
	loser.Status = NodeRestarted
	if err := m.replaceNode(id, "node-a", before, loser, m.rejectSettled); !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("replaceNode() with a stale snapshot = %v, want ErrConcurrentUpdate", err)
	}

	if !rt.CanContinue() {
		t.Fatal("a rejected concurrent update incorrectly tripped state_persistence_failed")
	}
	got, _ := m.Get(id)
	if got.Nodes[0].Status != NodeCommandIssued {
		t.Fatalf("Nodes[0].Status = %q, want the concurrent writer's committed value preserved", got.Nodes[0].Status)
	}
}

// UpdateNode promises mutate runs against a private copy. A NodeResult
// carries its times behind pointers, so "private" has to mean the times too:
// otherwise a module writing a timestamp through the pointer it was handed
// reaches straight into the stored record, with no lock held, no validation
// and nothing written to disk.
func TestRuntimeUpdateNodeMutateCannotWriteThroughTheTimeItWasHanded(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	at := time.Unix(1700000000, 0).UTC()
	if err := rt.InitNodes([]NodeResult{{
		NodeName: "node-a", Role: inventory.RoleWorker, Status: NodeCommandIssued, FinishedAt: &at,
	}}); err != nil {
		t.Fatal(err)
	}

	var duringMutate time.Time
	if err := rt.UpdateNode("node-a", func(n *NodeResult) {
		*n.FinishedAt = at.Add(time.Hour)
		current, _ := rt.Operation()
		duringMutate = *current.Nodes[0].FinishedAt
	}); err != nil {
		t.Fatalf("UpdateNode() = %v", err)
	}

	if !duringMutate.Equal(at) {
		t.Errorf("stored FinishedAt during mutate = %v, want the record untouched at %v", duringMutate, at)
	}
	got, _ := m.Get(id)
	if !got.Nodes[0].FinishedAt.Equal(at.Add(time.Hour)) {
		t.Errorf("committed FinishedAt = %v, want the mutation applied through the commit", got.Nodes[0].FinishedAt)
	}
}

// The compare-and-replace exists to catch a writer whose change would
// otherwise be lost. Two timestamps that say the same moment are not such a
// change, and refusing an update because somebody rewrote a pointer would be
// a conflict a caller can neither see nor act on.
func TestRuntimeUpdateNodeComparesNodesByValueRatherThanPointerIdentity(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	at := time.Unix(1700000000, 0).UTC()
	if err := rt.InitNodes([]NodeResult{{
		NodeName: "node-a", Role: inventory.RoleWorker, Status: NodeCommandIssued, FinishedAt: &at,
	}}); err != nil {
		t.Fatal(err)
	}

	before, err := m.snapshotNode(id, "node-a", m.rejectSettled)
	if err != nil {
		t.Fatal(err)
	}
	same := *before.FinishedAt
	rewritten := before
	rewritten.FinishedAt = &same
	if err := m.replaceNode(id, "node-a", before, rewritten, m.rejectSettled); err != nil {
		t.Fatalf("replaceNode() rewriting a timestamp to the same moment = %v", err)
	}

	next := before
	next.Status = NodeRestarted
	if err := m.replaceNode(id, "node-a", before, next, m.rejectSettled); err != nil {
		t.Fatalf("replaceNode() against an unchanged node = %v, want it committed", err)
	}
	got, _ := m.Get(id)
	if got.Nodes[0].Status != NodeRestarted {
		t.Errorf("Nodes[0].Status = %q, want restarted", got.Nodes[0].Status)
	}
}

// TestRuntimeUpdateNodeConcurrentCallersSettleWithoutCorruption is the Minor
// "real concurrent interleaving" coverage: many goroutines race
// UpdateNode against the same node. Every call must return either nil or
// ErrConcurrentUpdate, at least one must succeed, and the final record must
// still be a single, uncorrupted node. Run with -race.
func TestRuntimeUpdateNodeConcurrentCallersSettleWithoutCorruption(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	const attempts = 50

	var wg sync.WaitGroup
	var mu sync.Mutex
	var succeeded, conflicted int

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := rt.UpdateNode("node-a", func(n *NodeResult) {
				// Shaped like a code (see safeCode), so each writer's
				// value is kept as written and the record can be asked
				// afterwards what actually landed.
				n.Code = fmt.Sprintf("attempt_%d", i)
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, ErrConcurrentUpdate):
				conflicted++
			default:
				t.Errorf("UpdateNode() = %v, want nil or ErrConcurrentUpdate", err)
			}
		}(i)
	}
	wg.Wait()

	if succeeded == 0 {
		t.Fatal("no concurrent UpdateNode call ever succeeded")
	}
	if succeeded+conflicted != attempts {
		t.Fatalf("succeeded(%d)+conflicted(%d) != attempts(%d)", succeeded, conflicted, attempts)
	}
	got, ok := m.Get(id)
	if !ok || len(got.Nodes) != 1 || got.Nodes[0].NodeName != "node-a" {
		t.Fatalf("operation's node list is corrupted: %+v", got)
	}
}

// TestRuntimeConcurrentStartStepDoesNotRace is further Minor concurrency
// coverage across a different checked mutation. Run with -race.
func TestRuntimeConcurrentStartStepDoesNotRace(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := rt.StartStep(fmt.Sprintf("step-%d", i)); err != nil {
				t.Errorf("StartStep() = %v", err)
			}
		}(i)
	}
	wg.Wait()

	got, _ := m.Get(id)
	if len(got.Steps) != n {
		t.Fatalf("Steps = %d, want %d", len(got.Steps), n)
	}
}

func TestRuntimeCompleteRejectsInvalidOutcome(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	before, _ := m.Get(id)

	err := rt.Complete(Outcome{Status: Status("bogus")})
	if !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("Complete() = %v, want ErrInvalidOutcome", err)
	}

	after, _ := m.Get(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("invalid outcome mutated the operation: before=%+v after=%+v", before, after)
	}
}

// TestOutcomeValidRejectsCrossFieldMismatch is a Minor fix: command_issued
// requires a future deadline, and every other status must carry none.
func TestOutcomeValidRejectsCrossFieldMismatch(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	cases := []struct {
		name    string
		outcome Outcome
		want    bool
	}{
		{"succeeded with no deadline", Outcome{Status: StatusSucceeded}, true},
		{"succeeded with a deadline", Outcome{Status: StatusSucceeded, CommandIssuedUntil: now.Add(time.Hour)}, false},
		{"failed with a deadline", Outcome{Status: StatusFailed, CommandIssuedUntil: now.Add(time.Hour)}, false},
		{"partially_failed with a deadline", Outcome{Status: StatusPartiallyFailed, CommandIssuedUntil: now.Add(time.Hour)}, false},
		{"command_issued with a future deadline", Outcome{Status: StatusCommandIssued, CommandIssuedUntil: now.Add(time.Hour)}, true},
		{"command_issued with no deadline", Outcome{Status: StatusCommandIssued}, false},
		{"command_issued with a past deadline", Outcome{Status: StatusCommandIssued, CommandIssuedUntil: now.Add(-time.Hour)}, false},
		{"command_issued with the deadline exactly now", Outcome{Status: StatusCommandIssued, CommandIssuedUntil: now}, false},
		{"unknown status", Outcome{Status: Status("bogus")}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.outcome.valid(now); got != tc.want {
				t.Fatalf("valid(%v) = %v, want %v", now, got, tc.want)
			}
		})
	}
}

func TestRuntimeCompleteRejectsCommandIssuedWithoutFutureDeadline(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	if err := rt.Complete(Outcome{Status: StatusCommandIssued}); !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("Complete() = %v, want ErrInvalidOutcome", err)
	}
}

func TestRuntimeCompleteRejectsSucceededWithADeadline(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	err := rt.Complete(Outcome{Status: StatusSucceeded, CommandIssuedUntil: time.Now().Add(time.Hour)})
	if !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("Complete() = %v, want ErrInvalidOutcome", err)
	}
}

func TestRuntimeCompleteRejectsSecondCompletion(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	if err := rt.Complete(Outcome{Status: StatusSucceeded}); err != nil {
		t.Fatal(err)
	}
	if err := rt.Complete(Outcome{Status: StatusSucceeded}); !errors.Is(err, ErrOperationTerminal) {
		t.Fatalf("second Complete() = %v, want ErrOperationTerminal", err)
	}
}

func TestRuntimeCompleteSetsTerminalFieldsAndReleasesLock(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.Complete(Outcome{Status: StatusSucceeded, Code: "", Error: ""}); err != nil {
		t.Fatal(err)
	}

	got, ok := m.Get(id)
	if !ok {
		t.Fatal("operation missing after Complete")
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", got.Status)
	}
	if got.FinishedAt == nil {
		t.Fatal("FinishedAt not set")
	}
	if m.activeID != "" {
		t.Fatalf("activeID = %q, want released", m.activeID)
	}
}

func TestRuntimeCompleteWithCommandIssuedKeepsLock(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	until := time.Now().Add(time.Hour)
	if err := rt.Complete(Outcome{Status: StatusCommandIssued, CommandIssuedUntil: until}); err != nil {
		t.Fatal(err)
	}

	got, ok := m.Get(id)
	if !ok || got.CommandIssuedUntil.IsZero() {
		t.Fatalf("CommandIssuedUntil not persisted: %+v", got)
	}
	if m.activeID != id {
		t.Fatalf("activeID = %q, want %q (lock held until the deadline)", m.activeID, id)
	}
}

// --- I2: an active command_issued operation stays mutable through its grace
// window, and stops being mutable once the deadline has passed. ---

func TestRuntimeAllowsClearingCommandIssuedUntilWithinGrace(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Hour)
	if err := rt.SetCommandIssuedUntil(time.Time{}); err != nil {
		t.Fatalf("SetCommandIssuedUntil() = %v, want nil while still within grace", err)
	}
	got, _ := m.Get(id)
	if !got.CommandIssuedUntil.IsZero() {
		t.Fatalf("CommandIssuedUntil = %v, want cleared", got.CommandIssuedUntil)
	}
}

func TestRuntimeAllowsFinishStepAndUpdateNodeWithinGrace(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Hour)
	if err := rt.FinishStep(StepMasterCommand, StepSucceeded, "", ""); err != nil {
		t.Fatalf("FinishStep() = %v, want nil while still within grace", err)
	}
	if err := rt.UpdateNode("node-a", func(n *NodeResult) { n.Status = NodeRestarted }); err != nil {
		t.Fatalf("UpdateNode() = %v, want nil while still within grace", err)
	}
	got, _ := m.Get(id)
	if got.Steps[0].Status != StepSucceeded {
		t.Fatalf("Steps[0].Status = %q, want succeeded", got.Steps[0].Status)
	}
	if got.Nodes[0].Status != NodeRestarted {
		t.Fatalf("Nodes[0].Status = %q, want restarted", got.Nodes[0].Status)
	}
}

func TestRuntimeAllowsCompleteFromCommandIssuedToSucceededWithinGrace(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Hour)
	if err := rt.Complete(Outcome{Status: StatusSucceeded}); err != nil {
		t.Fatalf("Complete() = %v, want nil while still within grace", err)
	}
	got, _ := m.Get(id)
	if got.Status != StatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", got.Status)
	}
	if m.activeID != "" {
		t.Fatalf("activeID = %q, want released after Complete", m.activeID)
	}
}

func TestRuntimeRejectsMutationsAfterCommandIssuedGraceExpires(t *testing.T) {
	cases := []struct {
		name string
		call func(Runtime) error
	}{
		{"SetCommandIssuedUntil", func(rt Runtime) error { return rt.SetCommandIssuedUntil(time.Time{}) }},
		{"FinishStep", func(rt Runtime) error { return rt.FinishStep(StepMasterCommand, StepSucceeded, "", "") }},
		{"UpdateNode", func(rt Runtime) error { return rt.UpdateNode("node-a", func(*NodeResult) {}) }},
		{"Complete", func(rt Runtime) error { return rt.Complete(Outcome{Status: StatusSucceeded}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt, _, _ := buildCommandIssuedRuntime(t, -time.Hour)
			if err := tc.call(rt); !errors.Is(err, ErrOperationTerminal) {
				t.Fatalf("%s() = %v, want ErrOperationTerminal after the grace deadline", tc.name, err)
			}
		})
	}
}

// A recovery runtime is allowed to touch a command_issued record however
// long its grace deadline has passed — that is the whole point of
// rejectSettledDuringRecovery — but it must never use that latitude to put
// the cluster back on hold for a deadline that has already run out. Naming
// a new deadline is not evidence of anything; it would just be this daemon
// re-arming a lock nothing asked it to hold.
func TestRecoveryCannotExtendAnExpiredDeadline(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, -time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())
	before, _ := m.Get(id)

	err := recovery.SetCommandIssuedUntil(m.deps.Now().Add(time.Hour))
	if !errors.Is(err, ErrRecoveryCannotExtendDeadline) {
		t.Fatalf("SetCommandIssuedUntil() = %v, want ErrRecoveryCannotExtendDeadline", err)
	}

	got, _ := m.Get(id)
	if !got.CommandIssuedUntil.Equal(before.CommandIssuedUntil) {
		t.Errorf("CommandIssuedUntil = %v, want it left exactly as it was found (%v)",
			got.CommandIssuedUntil, before.CommandIssuedUntil)
	}
}

// Recovery clearing an expired deadline is exactly the confirm-a-command
// case this package exists for: it releases the lock, it never re-arms it.
func TestRecoveryCanClearAnExpiredDeadline(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, -time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())

	if err := recovery.SetCommandIssuedUntil(time.Time{}); err != nil {
		t.Fatalf("SetCommandIssuedUntil(zero) = %v, want nil", err)
	}
	got, _ := m.Get(id)
	if !got.CommandIssuedUntil.IsZero() {
		t.Errorf("CommandIssuedUntil = %v, want it cleared", got.CommandIssuedUntil)
	}
}

// A run's own runtime gets no new restriction from this: recovery is the
// only caller this guards against, because it is the only one allowed to
// touch a record whose deadline has already passed.
func TestRunRuntimeMaySetAFutureDeadlineRegardless(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Hour)
	if err := rt.SetCommandIssuedUntil(m.deps.Now().Add(2 * time.Hour)); err != nil {
		t.Fatalf("SetCommandIssuedUntil() = %v, want nil for an ordinary run", err)
	}
	got, _ := m.Get(id)
	if got.CommandIssuedUntil.IsZero() {
		t.Error("CommandIssuedUntil was not persisted")
	}
}

// The deadline guard has to be the same rule no matter which Runtime method
// writes CommandIssuedUntil — Complete and Settle carry it just as much as
// SetCommandIssuedUntil does, because a RecoverableModule normally reaches
// the field through one of those two, not through SetCommandIssuedUntil
// directly. See checkCommandIssuedUntilTransition in runtime.go.

func TestRecoveryCannotExtendAnExpiredDeadlineViaComplete(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, -time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())
	before, _ := m.Get(id)

	err := recovery.Complete(Outcome{Status: StatusCommandIssued, CommandIssuedUntil: m.deps.Now().Add(time.Hour)})
	if !errors.Is(err, ErrRecoveryCannotExtendDeadline) {
		t.Fatalf("Complete() = %v, want ErrRecoveryCannotExtendDeadline", err)
	}
	got, _ := m.Get(id)
	if got.Status != StatusCommandIssued || !got.CommandIssuedUntil.Equal(before.CommandIssuedUntil) {
		t.Errorf("operation = %+v, want it left exactly as it was found (CommandIssuedUntil %v)",
			got, before.CommandIssuedUntil)
	}
}

func TestRecoveryCannotExtendAnExpiredDeadlineViaSettle(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, -time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())
	before, _ := m.Get(id)

	s := settlement{outcome: Outcome{Status: StatusCommandIssued, CommandIssuedUntil: m.deps.Now().Add(time.Hour)}}
	if err := recovery.Settle(s); !errors.Is(err, ErrRecoveryCannotExtendDeadline) {
		t.Fatalf("Settle() = %v, want ErrRecoveryCannotExtendDeadline", err)
	}
	got, _ := m.Get(id)
	if got.Status != StatusCommandIssued || !got.CommandIssuedUntil.Equal(before.CommandIssuedUntil) {
		t.Errorf("operation = %+v, want it left exactly as it was found (CommandIssuedUntil %v)",
			got, before.CommandIssuedUntil)
	}
}

// A record whose CommandIssuedUntil has never been set is not "expired" —
// there was nothing to re-arm — so recovery may still give a running record
// its own first command_issued grace period, through any of the three
// paths that write the field.

func TestRecoveryMaySetAFirstDeadlineFromZeroViaSetCommandIssuedUntil(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if got, _ := m.Get(id); !got.CommandIssuedUntil.IsZero() {
		t.Fatalf("fixture already has a deadline: %v", got.CommandIssuedUntil)
	}
	recovery := newRecoveryRuntime(m, id, context.Background())

	future := rt.Now().Add(time.Hour)
	if err := recovery.SetCommandIssuedUntil(future); err != nil {
		t.Fatalf("SetCommandIssuedUntil() = %v, want nil for a first-time deadline from zero", err)
	}
	got, _ := m.Get(id)
	if !got.CommandIssuedUntil.Equal(future) {
		t.Errorf("CommandIssuedUntil = %v, want %v", got.CommandIssuedUntil, future)
	}
}

func TestRecoveryMaySettleCommandIssuedFromRunningWithAFreshDeadline(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	recovery := newRecoveryRuntime(m, id, context.Background())
	future := rt.Now().Add(time.Hour)

	if err := recovery.Complete(Outcome{Status: StatusCommandIssued, CommandIssuedUntil: future}); err != nil {
		t.Fatalf("Complete() = %v, want nil for a first-time deadline from zero", err)
	}
	got, _ := m.Get(id)
	if got.Status != StatusCommandIssued || !got.CommandIssuedUntil.Equal(future) {
		t.Errorf("operation = %+v, want command_issued with CommandIssuedUntil %v", got, future)
	}
}

// Late reboot confirmation — the whole reason recovery is allowed to touch
// an expired command_issued record at all — clears the deadline (the zero
// value) rather than naming a new one, so it must keep succeeding through
// Settle exactly as confirmReboot relies on. See
// TestARebootIsConfirmedAfterItsGraceDeadlineHasLongPassed for the same
// guarantee end to end through rebootModule.
func TestRecoveryMaySettleWithZeroDeadlineEvenWhenExpired(t *testing.T) {
	_, m, id := buildCommandIssuedRuntime(t, -time.Hour)
	recovery := newRecoveryRuntime(m, id, context.Background())

	s := settlement{outcome: Outcome{Status: StatusSucceeded}}
	if err := recovery.Settle(s); err != nil {
		t.Fatalf("Settle() = %v, want nil: clearing the deadline must still succeed for a late confirmation", err)
	}
	got, _ := m.Get(id)
	if got.Status != StatusSucceeded {
		t.Errorf("Status = %q, want succeeded", got.Status)
	}
	if !got.CommandIssuedUntil.IsZero() {
		t.Errorf("CommandIssuedUntil = %v, want cleared", got.CommandIssuedUntil)
	}
}

// --- I3: a module's own error text never reaches the persisted record. ---

func TestRuntimeFinishStepPersistsSafeReasonNotRawDetail(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.StartStep("work"); err != nil {
		t.Fatal(err)
	}

	secret := "token=super-secret raw detail from a lower-level error"
	if err := rt.FinishStep("work", StepFailed, CodeNodeUnreachable, secret); err != nil {
		t.Fatal(err)
	}

	got, _ := m.Get(id)
	want := reasonFor(CodeNodeUnreachable)
	if got.Steps[0].Error != want {
		t.Fatalf("Steps[0].Error = %q, want the safe reason %q", got.Steps[0].Error, want)
	}
	if strings.Contains(got.Steps[0].Error, "super-secret") {
		t.Fatalf("Steps[0].Error leaked raw detail: %q", got.Steps[0].Error)
	}

	onDisk, ok, err := m.deps.Store.Load(id)
	if err != nil || !ok {
		t.Fatalf("Store.Load: ok=%v err=%v", ok, err)
	}
	if strings.Contains(onDisk.Steps[0].Error, "super-secret") {
		t.Fatalf("on-disk record leaked raw detail: %q", onDisk.Steps[0].Error)
	}
}

func TestRuntimeFinishStepSuccessKeepsErrorEmpty(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.StartStep("work"); err != nil {
		t.Fatal(err)
	}
	// A module bug that supplies detail text with no code must not leak it
	// either: a blank code always persists an empty error.
	if err := rt.FinishStep("work", StepSucceeded, "", "unexpected leftover detail"); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(id)
	if got.Steps[0].Error != "" {
		t.Fatalf("Steps[0].Error = %q, want empty for a blank code", got.Steps[0].Error)
	}
}

func TestRuntimeCompletePersistsSafeReasonNotRawDetail(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	secret := "Authorization: Bearer super-secret-token"
	if err := rt.Complete(Outcome{Status: StatusFailed, Code: CodeHostPowerFailed, Error: secret}); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(id)
	want := reasonFor(CodeHostPowerFailed)
	if got.Error != want {
		t.Fatalf("Error = %q, want the safe reason %q", got.Error, want)
	}
	if strings.Contains(got.Error, "super-secret-token") {
		t.Fatalf("Error leaked raw detail: %q", got.Error)
	}

	onDisk, ok, err := m.deps.Store.Load(id)
	if err != nil || !ok {
		t.Fatalf("Store.Load: ok=%v err=%v", ok, err)
	}
	if strings.Contains(onDisk.Error, "super-secret-token") {
		t.Fatalf("on-disk record leaked raw detail: %q", onDisk.Error)
	}
}

func TestRuntimeCompleteSuccessKeepsErrorEmpty(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.Complete(Outcome{Status: StatusSucceeded, Error: "leftover detail nobody meant to persist"}); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(id)
	if got.Error != "" {
		t.Fatalf("Error = %q, want empty on success", got.Error)
	}
}

func TestRuntimeSetModuleStateCopiesInput(t *testing.T) {
	rt, _, _ := newTestRuntime(t, StatusRunning)
	raw := json.RawMessage(`{"phase":"a"}`)
	if err := rt.SetModuleState(raw); err != nil {
		t.Fatal(err)
	}

	raw[2] = 'X' // mutate the caller's slice after handing it to the runtime

	op, _ := rt.Operation()
	if string(op.ModuleState) != `{"phase":"a"}` {
		t.Fatalf("ModuleState = %s, want unaffected by a later caller mutation", op.ModuleState)
	}
}

// TestRuntimeSetModuleStateRejectsInvalidJSON is a Minor fix: invalid JSON
// must be rejected before the manager is ever locked, so it cannot trip
// state_persistence_failed or touch the operation at all.
func TestRuntimeSetModuleStateRejectsInvalidJSON(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	before, _ := m.Get(id)

	err := rt.SetModuleState(json.RawMessage(`{not valid`))
	if !errors.Is(err, ErrInvalidModuleState) {
		t.Fatalf("SetModuleState() = %v, want ErrInvalidModuleState", err)
	}

	after, _ := m.Get(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("invalid module state mutated the operation: before=%+v after=%+v", before, after)
	}
	if !m.canContinue(id) {
		t.Fatal("invalid module state incorrectly tripped state_persistence_failed")
	}
}

func TestRuntimeSetModuleStateClearsOnEmptyInput(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.SetModuleState(json.RawMessage(`{"phase":"a"}`)); err != nil {
		t.Fatal(err)
	}
	if err := rt.SetModuleState(nil); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(id)
	if got.ModuleState != nil {
		t.Fatalf("ModuleState = %s, want cleared by a nil input", got.ModuleState)
	}
}

func TestRuntimeInitNodesCopiesInputSlice(t *testing.T) {
	rt, _, _ := newTestRuntime(t, StatusRunning)
	nodes := []NodeResult{{NodeName: "node-a", Role: inventory.RoleWorker, Status: NodePending}}
	if err := rt.InitNodes(nodes); err != nil {
		t.Fatal(err)
	}

	nodes[0].NodeName = "tampered"

	op, _ := rt.Operation()
	if len(op.Nodes) != 1 || op.Nodes[0].NodeName != "node-a" {
		t.Fatalf("Nodes = %+v, want unaffected by a later caller mutation", op.Nodes)
	}
}

// TestRuntimeInitNodesDeepCopiesNodeTimestamps is a Minor fix: InitNodes must
// copy each node's *time.Time fields, not just the NodeResult struct.
func TestRuntimeInitNodesDeepCopiesNodeTimestamps(t *testing.T) {
	rt, _, _ := newTestRuntime(t, StatusRunning)
	started := time.Now()
	original := started
	nodes := []NodeResult{{NodeName: "node-a", StartedAt: &started}}
	if err := rt.InitNodes(nodes); err != nil {
		t.Fatal(err)
	}

	*nodes[0].StartedAt = started.Add(time.Hour) // mutate through the caller's retained pointer

	op, _ := rt.Operation()
	if !op.Nodes[0].StartedAt.Equal(original) {
		t.Fatalf("StartedAt = %v, want unaffected by a mutation through the caller's pointer", op.Nodes[0].StartedAt)
	}
}

// TestRuntimeInitNodesPersistsEmptySliceNotNull is a Minor fix: InitNodes(nil)
// must round-trip through the store as "nodes":[], never as null.
func TestRuntimeInitNodesPersistsEmptySliceNotNull(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	if err := rt.InitNodes(nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(m.deps.Store.(*Store).Dir(), id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["nodes"]) != "[]" {
		t.Fatalf("nodes = %s, want []", raw["nodes"])
	}
}

func TestRuntimeRejectsMutationsOnTerminalOperation(t *testing.T) {
	cases := []struct {
		name string
		call func(Runtime) error
	}{
		{"StartStep", func(rt Runtime) error { return rt.StartStep("late") }},
		{"FinishStep", func(rt Runtime) error { return rt.FinishStep("late", StepSucceeded, "", "") }},
		{"InitNodes", func(rt Runtime) error { return rt.InitNodes(nil) }},
		{"UpdateNode", func(rt Runtime) error { return rt.UpdateNode("node-a", func(*NodeResult) {}) }},
		{"SetHostBootID", func(rt Runtime) error { return rt.SetHostBootID("boot-2") }},
		{"SetModuleState", func(rt Runtime) error { return rt.SetModuleState(json.RawMessage(`{}`)) }},
		{"SetCommandIssuedUntil", func(rt Runtime) error { return rt.SetCommandIssuedUntil(time.Now()) }},
		{"Complete", func(rt Runtime) error { return rt.Complete(Outcome{Status: StatusSucceeded}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := newRuntimeWithOperation(t, StatusSucceeded)
			if err := tc.call(rt); !errors.Is(err, ErrOperationTerminal) {
				t.Fatalf("%s() = %v, want ErrOperationTerminal", tc.name, err)
			}
		})
	}
}

func TestRuntimeValidationFailureDoesNotMutateOperation(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)
	before, _ := m.Get(id)

	if err := rt.FinishStep("missing", StepSucceeded, "", ""); !errors.Is(err, ErrStepNotFound) {
		t.Fatalf("FinishStep() = %v, want ErrStepNotFound", err)
	}

	after, _ := m.Get(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("a rejected mutation changed the operation: before=%+v after=%+v", before, after)
	}
}

// TestRuntimeRejectsMutationAfterPersistenceFailure proves the checked
// mutation follows the same state_persistence_failed settlement update
// already applies: the failing call itself reports the low-level failure,
// and the operation is then locked into ErrOperationTerminal rather than
// retried, exactly as canContinue()/update() already behave.
func TestRuntimeRejectsMutationAfterPersistenceFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	rt, m, id := newTestRuntimeAt(t, StatusRunning, dir)

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	if err := rt.StartStep("work"); !errors.Is(err, errStatePersistenceFailed) {
		t.Fatalf("StartStep() after broken store = %v, want errStatePersistenceFailed", err)
	}

	got, ok := m.Get(id)
	if !ok || got.Status != StatusFailed || got.Code != CodeStatePersistenceFailed {
		t.Fatalf("operation = %+v, want failed/%s", got, CodeStatePersistenceFailed)
	}

	if err := rt.StartStep("again"); !errors.Is(err, ErrOperationTerminal) {
		t.Fatalf("StartStep() after settled persistence failure = %v, want ErrOperationTerminal", err)
	}
}

func TestRuntimeCanContinueReflectsPersistenceFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	rt, _, _ := newTestRuntimeAt(t, StatusRunning, dir)
	if !rt.CanContinue() {
		t.Fatal("CanContinue() = false before any failure")
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := rt.StartStep("work"); err == nil {
		t.Fatal("StartStep() succeeded against a removed store")
	}
	if rt.CanContinue() {
		t.Fatal("CanContinue() = true after a persistence failure")
	}
}

// TestRuntimeNowDelegatesToManagerClock brackets rt.Now() between two direct
// calls to the manager's own clock, so it actually proves Now() reads that
// clock rather than merely being unable to observe it running backwards
// (which, against a monotonically increasing clock, is a vacuous check).
func TestRuntimeNowDelegatesToManagerClock(t *testing.T) {
	rt, m, _ := newTestRuntime(t, StatusRunning)

	before := m.deps.Now()
	got := rt.Now()
	after := m.deps.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("Now() = %v, want between %v and %v (the manager's own clock)", got, before, after)
	}
	// The fake clock's epoch is in 2023; the real wall clock is not, so this
	// tells apart Now() delegating to the manager clock from it reading the
	// wall clock directly.
	if time.Since(got) < 24*time.Hour {
		t.Fatalf("Now() = %v, want the injected clock rather than the wall clock", got)
	}
}

func TestRuntimeContextReturnsBoundContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "marker")
	rt, _, _ := newTestRuntimeWithContext(t, StatusRunning, ctx)
	if got := rt.Context().Value(key{}); got != "marker" {
		t.Fatalf("Context() = %v, want the context newRuntime was given", got)
	}
}

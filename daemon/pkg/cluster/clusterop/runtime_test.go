package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// buildTestRuntime wires a Manager with every dependency stubbed to a no-op
// and seeds one operation directly into it, the same way
// TestPersistenceFailureStopsBeforePoweringTheControlNode seeds one in
// manager_test.go. A Runtime test drives state transitions on that record; it
// never needs a real cluster.
func buildTestRuntime(t *testing.T, status Status, dir string, ctx context.Context) (Runtime, *Manager, string) {
	t.Helper()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	clock := newClock()
	deps := Deps{
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
	}
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

func TestRuntimeNowDelegatesToManagerClock(t *testing.T) {
	rt, m, _ := newTestRuntime(t, StatusRunning)
	if got, want := rt.Now(), m.deps.Now(); got.After(want) {
		t.Fatalf("Now() = %v, want no later than the manager clock %v", got, want)
	}
	// The fake clock's epoch is in 2023; the real wall clock is not, so this
	// tells apart Now() delegating to the manager clock from it reading the
	// wall clock directly.
	if time.Since(rt.Now()) < 24*time.Hour {
		t.Fatalf("Now() = %v, want the injected clock rather than the wall clock", rt.Now())
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

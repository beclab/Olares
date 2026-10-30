package domain

import (
	"context"
	"errors"
	"testing"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// worldAt is a test helper that builds a minimal World whose Current()
// returns state. Handlers under test only read Current() (never the
// full AM), so this is enough for pure-engine tests.
func worldAt(state State) World {
	return World{
		AM: &appsv1.ApplicationManager{
			Status: appsv1.ApplicationManagerStatus{State: state},
		},
	}
}

// TestEngineStep_HappyPath asserts that Engine.Step dispatches an action
// against a declared Transition and returns the handler's Result verbatim
// (Next is filled in, no error).
func TestEngineStep_HappyPath(t *testing.T) {
	tab := NewTransitionTable()
	tab.SetSpec(appsv1.Pending, StateSpec{})
	tab.SetSpec(appsv1.Downloading, StateSpec{})
	tab.AddReachable(appsv1.Pending, appsv1.Downloading)
	tab.Add(Transition{
		From:    appsv1.Pending,
		On:      Tick{},
		Handler: PlaceholderHandler(appsv1.Downloading),
	})
	e := NewEngine(tab)

	res, err := e.Step(context.Background(), worldAt(appsv1.Pending), Tick{})
	if err != nil {
		t.Fatalf("Step err=%v want nil", err)
	}
	if res.Next != appsv1.Downloading {
		t.Errorf("Next=%q want %q", res.Next, appsv1.Downloading)
	}
}

// TestEngineStep_NoTransition asserts Step returns ErrNoTransition when
// no edge exists for (state, action-kind). This is the "reconcile is a
// no-op" signal for terminal states.
func TestEngineStep_NoTransition(t *testing.T) {
	tab := NewTransitionTable()
	e := NewEngine(tab)

	_, err := e.Step(context.Background(), worldAt(appsv1.Running), Tick{})
	if err == nil {
		t.Fatal("Step err=nil want ErrNoTransition")
	}
	var target ErrNoTransition
	if !errors.As(err, &target) {
		t.Fatalf("Step err=%T want ErrNoTransition", err)
	}
	if target.From != appsv1.Running || target.Action != KindTick {
		t.Errorf("ErrNoTransition=%+v want {Running, Tick}", target)
	}
}

// TestEngineStep_NilAction returns ErrNoTransition (with an empty
// Action tag) rather than panicking. The Runner uses this as a
// defensive guard so a mis-wired call site produces a diagnosable
// error instead of a nil-panic in the reconcile worker.
func TestEngineStep_NilAction(t *testing.T) {
	tab := NewTransitionTable()
	e := NewEngine(tab)

	_, err := e.Step(context.Background(), worldAt(appsv1.Running), nil)
	if err == nil {
		t.Fatal("Step err=nil want ErrNoTransition")
	}
	var target ErrNoTransition
	if !errors.As(err, &target) {
		t.Fatalf("Step err=%T want ErrNoTransition", err)
	}
}

// TestEngineStep_UndeclaredTransitionRejected is the CONTRACT.md rule
// 4 enforcement test: a handler that returns Result.Next pointing at a
// state that is not declared in the reachable map MUST cause
// ErrUndeclaredTransition. This is the runtime guard that prevents the
// legacy "six drifting tables" problem from recurring.
func TestEngineStep_UndeclaredTransitionRejected(t *testing.T) {
	tab := NewTransitionTable()
	tab.SetSpec(appsv1.Pending, StateSpec{})
	tab.SetSpec(appsv1.Downloading, StateSpec{})
	tab.SetSpec(appsv1.Running, StateSpec{})
	tab.AddReachable(appsv1.Pending, appsv1.Downloading)
	// Handler smuggles a jump to Running — which is NOT reachable.
	tab.Add(Transition{
		From:    appsv1.Pending,
		On:      Tick{},
		Handler: PlaceholderHandler(appsv1.Running),
	})
	e := NewEngine(tab)

	_, err := e.Step(context.Background(), worldAt(appsv1.Pending), Tick{})
	if err == nil {
		t.Fatal("Step err=nil want ErrUndeclaredTransition")
	}
	var target ErrUndeclaredTransition
	if !errors.As(err, &target) {
		t.Fatalf("Step err=%T want ErrUndeclaredTransition", err)
	}
	if target.From != appsv1.Pending || target.To != appsv1.Running {
		t.Errorf("ErrUndeclaredTransition=%+v want {Pending, Running}", target)
	}
}

// TestEngineStep_SelfWriteBypassesReachability documents that returning
// Next == current state is always legal (Runner needs this for
// idempotent Status re-assertion; see legacy
// IsStateTransitionAllowed's same-state relaxation).
func TestEngineStep_SelfWriteBypassesReachability(t *testing.T) {
	tab := NewTransitionTable()
	tab.SetSpec(appsv1.Pending, StateSpec{})
	// No reachable edges from Pending; only Tick registered.
	tab.Add(Transition{
		From:    appsv1.Pending,
		On:      Tick{},
		Handler: PlaceholderHandler(appsv1.Pending),
	})
	e := NewEngine(tab)

	res, err := e.Step(context.Background(), worldAt(appsv1.Pending), Tick{})
	if err != nil {
		t.Fatalf("Step err=%v want nil (self-write must be allowed)", err)
	}
	if res.Next != appsv1.Pending {
		t.Errorf("Next=%q want %q", res.Next, appsv1.Pending)
	}
}

// TestEngineStep_EmptyNextStaysPut documents that Next == "" (aka
// InitialState) is the handler's way to say "do not change state".
// Engine.Step does NOT trigger ErrUndeclaredTransition on empty Next.
func TestEngineStep_EmptyNextStaysPut(t *testing.T) {
	tab := NewTransitionTable()
	tab.SetSpec(appsv1.Running, StateSpec{})
	tab.Add(Transition{
		From:    appsv1.Running,
		On:      Tick{},
		Handler: func(_ context.Context, _ World) Result { return Result{} },
	})
	e := NewEngine(tab)

	res, err := e.Step(context.Background(), worldAt(appsv1.Running), Tick{})
	if err != nil {
		t.Fatalf("Step err=%v", err)
	}
	if res.Next != InitialState {
		t.Errorf("Next=%q want empty", res.Next)
	}
}

// TestEngineStep_HandlerErrorSurfaces documents that Result.Err is
// returned verbatim from Step and takes precedence over the Next
// validation. The Runner surfaces this to controller-runtime, which
// re-enqueues with backoff.
func TestEngineStep_HandlerErrorSurfaces(t *testing.T) {
	tab := NewTransitionTable()
	tab.SetSpec(appsv1.Running, StateSpec{})
	boom := errors.New("boom")
	tab.Add(Transition{
		From: appsv1.Running,
		On:   Tick{},
		Handler: func(_ context.Context, _ World) Result {
			return Result{Err: boom}
		},
	})
	e := NewEngine(tab)

	_, err := e.Step(context.Background(), worldAt(appsv1.Running), Tick{})
	if !errors.Is(err, boom) {
		t.Errorf("err=%v want %v", err, boom)
	}
}

// TestActionKind_Uniqueness asserts that every declared Action type
// (plus every TriggerOp{op} for op in AllOps) has a unique ActionKind
// string. Duplicate kinds would collapse two edges into the same
// dispatch slot at Table.Add time, silently dropping edges.
func TestActionKind_Uniqueness(t *testing.T) {
	actions := []Action{
		Tick{},
		WorkloadReady{},
		WorkloadFailed{},
		TimeoutFired{},
	}
	for _, op := range AllOps {
		actions = append(actions, TriggerOp{Op: op})
	}

	seen := map[ActionKind]Action{}
	for _, a := range actions {
		if prev, dup := seen[a.Kind()]; dup {
			t.Errorf("duplicate ActionKind %q shared by %T and %T",
				a.Kind(), prev, a)
		}
		seen[a.Kind()] = a
	}
}

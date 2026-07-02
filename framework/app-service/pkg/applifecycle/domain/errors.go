package domain

import "fmt"

// ErrNoTransition is returned by Engine.Step when the TransitionTable has
// no edge from (world.Current(), action.Kind()). This is not necessarily a
// bug — most Reconcile ticks against a terminal state legitimately have no
// matching edge — so the Runner treats it as "reconcile is a no-op" rather
// than as a Reconcile failure.
type ErrNoTransition struct {
	From   State
	Action ActionKind
}

// Error implements error.
func (e ErrNoTransition) Error() string {
	return fmt.Sprintf("no transition declared from %q on action %q", e.From, e.Action)
}

// ErrGuardRejected is returned by Engine.Step when a matching Transition
// exists but every candidate's Guard returned false. Like ErrNoTransition,
// the Runner treats this as a no-op reconcile.
type ErrGuardRejected struct {
	From   State
	Action ActionKind
}

// Error implements error.
func (e ErrGuardRejected) Error() string {
	return fmt.Sprintf("all guards rejected transition from %q on action %q", e.From, e.Action)
}

// ErrUndeclaredTransition is returned by Engine.Step when the handler
// returned a Result.Next that is not declared in the TransitionTable. This
// IS a bug — the handler produced a state jump the table does not permit
// — and the Runner escalates it as a Reconcile failure (which
// controller-runtime will re-enqueue with backoff).
//
// The whole point of enforcing this at runtime is to root out the drift
// between the six independent tables in the legacy
// pkg/appstate/state_transition.go: a handler cannot silently write an
// unauthorized state; the write is rejected before it reaches the API
// server.
type ErrUndeclaredTransition struct {
	From State
	To   State
}

// Error implements error.
func (e ErrUndeclaredTransition) Error() string {
	return fmt.Sprintf("handler returned undeclared transition %q -> %q; declare it in TransitionTable or fix the handler", e.From, e.To)
}

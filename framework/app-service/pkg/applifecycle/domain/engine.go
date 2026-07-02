package domain

import "context"

// Engine is the pure dispatcher over a TransitionTable. It selects the
// matching Transition for (world.Current(), action), evaluates any
// Guard, invokes the Handler, and validates the returned Next state
// against the table. See CONTRACT.md for the invariants Engine enforces.
//
// Engine holds ONLY the table; it is stateless across Step calls and
// safe to share across goroutines (concurrent Step calls do not race).
type Engine struct {
	table *TransitionTable
}

// NewEngine returns an Engine backed by table.
func NewEngine(table *TransitionTable) *Engine {
	return &Engine{table: table}
}

// Table exposes the underlying TransitionTable for read-only queries by
// the Runner (TTL lookups, AllowsOp checks in apiserver, etc.).
func (e *Engine) Table() *TransitionTable {
	return e.table
}

// Step dispatches action against the state machine and returns the
// handler's Result. Semantics:
//
//   - If no Transition is declared for (world.Current(), action.Kind()),
//     returns ErrNoTransition. The Runner treats this as a no-op
//     reconcile.
//   - If every candidate's Guard returns false, returns ErrGuardRejected.
//     The Runner treats this as a no-op reconcile.
//   - If the Handler returns a non-empty Result.Next that is not
//     declared in the table, returns ErrUndeclaredTransition. This IS a
//     bug and the Runner escalates it as a Reconcile failure.
//   - Otherwise returns the Handler's Result verbatim (including
//     Result.Err, which the Runner also treats as a Reconcile failure).
//
// Step is pure: the same (world, action) always produces the same
// Result (modulo whatever the Handler decides to do, which itself must
// be pure per CONTRACT.md).
func (e *Engine) Step(ctx context.Context, w World, action Action) (Result, error) {
	if action == nil {
		return Result{}, ErrNoTransition{From: w.Current(), Action: ""}
	}

	kind := action.Kind()
	candidates := e.table.Candidates(w.Current(), kind)
	if len(candidates) == 0 {
		return Result{}, ErrNoTransition{From: w.Current(), Action: kind}
	}

	for _, tr := range candidates {
		if tr.Guard != nil && !tr.Guard(w) {
			continue
		}
		res := tr.Handler(ctx, w)
		if res.Err != nil {
			return res, res.Err
		}
		if res.Next != "" && res.Next != w.Current() && !e.table.Allows(w.Current(), res.Next) {
			return Result{}, ErrUndeclaredTransition{From: w.Current(), To: res.Next}
		}
		return res, nil
	}

	return Result{}, ErrGuardRejected{From: w.Current(), Action: kind}
}

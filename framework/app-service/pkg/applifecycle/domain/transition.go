package domain

import (
	"context"
	"fmt"
	"time"
)

// Driver classifies who fires the transition. Kept for observability and
// for the Runner to know whether a state should participate in the
// timeout ticker sweep. The distinction mirrors the legacy split between
// ReconcileDrivenTransitions (fired by reconcile ticks) and
// operation-driven transitions (fired by apiserver requests).
type Driver int

const (
	// DriverOp indicates the transition is fired by a user-initiated
	// operation (apiserver POST /install, /uninstall, etc.).
	DriverOp Driver = iota

	// DriverTick indicates the transition is fired by a Reconcile tick
	// observing the world (e.g. Installing -> Running when workload
	// becomes Ready).
	DriverTick

	// DriverCallback indicates the transition is fired by a workload
	// completion callback (WorkloadReady / WorkloadFailed) originating
	// from the Runner after a CallWorkload effect resolves.
	DriverCallback

	// DriverTimeout indicates the transition is fired by the TTL timer
	// expiring on the current state (TimeoutFired action).
	DriverTimeout
)

// HandlerFunc is a pure function of (context, World) that returns a
// Result. It MUST NOT mutate any input. See CONTRACT.md for the full
// set of forbidden behaviors.
type HandlerFunc func(ctx context.Context, w World) Result

// Guard is an optional predicate a Transition can carry to disambiguate
// multiple candidates that share the same (From, ActionKind). Guards are
// pure functions of World. Evaluated in insertion order; the first
// candidate whose guard returns true wins.
type Guard func(w World) bool

// Transition is a declarative record of one edge in the state machine.
// Transitions are values, not code: they are compiled at process start
// (in table_default.go) and never mutated after Engine construction.
type Transition struct {
	// From is the source state.
	From State

	// On is the input event that fires this transition. Kept as a
	// concrete Action value so table_default.go can be read as a
	// literal expression of the state machine. Table.Add dispatches
	// on On.Kind() at insertion time.
	On Action

	// Guard is optional. When non-nil, evaluated against the World
	// before Handler is invoked; if it returns false the transition
	// is skipped and the next candidate is tried.
	Guard Guard

	// Handler is the pure function invoked when this transition fires.
	// Its Result.Next is validated against Table.Allows before the
	// Runner applies the effects.
	Handler HandlerFunc
}

// StateSpec is metadata attached to a state (not to an edge). Kept
// separate from Transition so that a state's TTL / classification does
// not have to be repeated on every outgoing edge.
type StateSpec struct {
	// TTL is the state timeout; when elapsed without a terminal
	// transition, the Runner fires a TimeoutFired action. Zero means
	// "use the default TTL" (defaultTTL, 10 minutes).
	TTL time.Duration

	// Operating is true when the state represents work in progress
	// (Downloading, Installing, ..., Uninstalling, and their
	// *Canceling counterparts). Matches legacy OperatingStates.
	Operating bool

	// Canceling is true when the state is a *Canceling variant.
	// Matches legacy CancelingStates. Every Canceling state is also
	// Operating.
	Canceling bool

	// Cancelable is true when the state has an outgoing CancelOp
	// transition. Matches legacy CancelableStates.
	Cancelable bool

	// TerminalReinstallable is true when the state is a "clean
	// terminal" from which a fresh Install can be started against
	// the same AM. Matches legacy IsTerminalReinstallable.
	TerminalReinstallable bool

	// SafelyDeletable is true when the AM CR can be deleted directly
	// from this state without leaking cluster resources. Strict
	// subset of TerminalReinstallable. Matches legacy IsSafelyDeletable.
	SafelyDeletable bool
}

// TransitionTable is the single source of truth for the state machine.
// It merges the six independent tables in the legacy
// pkg/appstate/state_transition.go into one queryable structure:
//
//   - Allows            replaces StateTransitions
//   - AllowsOp          replaces OperationAllowedInState / IsOperationAllowed
//   - TTL               replaces StateToDurationMap / StateToDuration
//   - IsOperating       replaces OperatingStates
//   - IsCanceling       replaces CancelingStates
//   - IsCancelable      replaces CancelableStates
//   - IsTerminalReinstallable / IsSafelyDeletable replace the same-named funcs
//
// All accessors are read-only after construction. Table.Add MUST be
// called only during table setup (in table_default.go's init).
type TransitionTable struct {
	// edges holds every registered Transition, keyed by
	// (From, ActionKind). The []Transition inside preserves insertion
	// order so Guards resolve deterministically.
	edges map[State]map[ActionKind][]Transition

	// reachable[from][to] is true when the table has any registered
	// edge from -> to. Populated by Add so Allows is O(1).
	reachable map[State]map[State]bool

	// specs is the per-state metadata (TTL, classification).
	specs map[State]StateSpec

	// timeouts is the automatic deadline for states that dispatch
	// CallWorkload. Absence means the state does not time out.
	// This is deliberately separate from specs.TTL: legacy TTL()
	// still falls back to 10m for every unset state, which would
	// time out Running and Stopped if Runner used it.
	timeouts map[State]time.Duration
}

// NewTransitionTable returns an empty TransitionTable ready for Add.
func NewTransitionTable() *TransitionTable {
	return &TransitionTable{
		edges:     make(map[State]map[ActionKind][]Transition),
		reachable: make(map[State]map[State]bool),
		specs:     make(map[State]StateSpec),
		timeouts:  make(map[State]time.Duration),
	}
}

// SetSpec records the metadata for a state. Idempotent; a second call
// overwrites the first. Typically called from table_default.go's init.
func (t *TransitionTable) SetSpec(state State, spec StateSpec) {
	t.specs[state] = spec
}

// Add registers a Transition. Panics on invalid input to fail fast at
// startup rather than silently drop edges. Table construction happens
// once at process boot so panicking is safe and preferable to error
// returns.
func (t *TransitionTable) Add(tr Transition) {
	if tr.On == nil {
		panic("domain: Transition.On is nil; every edge must declare an Action")
	}
	if tr.Handler == nil {
		panic("domain: Transition.Handler is nil; use PlaceholderHandler while the real handler is not yet written")
	}
	kind := tr.On.Kind()
	byKind := t.edges[tr.From]
	if byKind == nil {
		byKind = make(map[ActionKind][]Transition)
		t.edges[tr.From] = byKind
	}
	byKind[kind] = append(byKind[kind], tr)
}

// SetOrAddHandler either replaces the handler on the first existing
// Transition matching (from, on.Kind()), or adds a fresh Transition
// if none exists. This is the recommended entry point for the
// handlers/ package during startup: it swaps a PlaceholderHandler
// for a real handler when one already exists (e.g. Installing on
// WorkloadReady), and adds a new edge when none is registered yet
// (e.g. Installing on Tick, which the default table does not
// declare because no legacy Installing → Installing self-loop exists
// but the phase-driven design needs one).
//
// Panics on nil handler (fail-fast at startup, per the same rationale
// as Add's nil checks). Does not touch reachability — if the added
// edge introduces a new (from, to) pair, the caller must also invoke
// AddReachable(from, to).
func (t *TransitionTable) SetOrAddHandler(from State, on Action, handler HandlerFunc) {
	if on == nil {
		panic("domain: SetOrAddHandler: on is nil; every edge must declare an Action")
	}
	if handler == nil {
		panic(fmt.Sprintf("domain: SetOrAddHandler: handler is nil for (%q, %q)", from, on.Kind()))
	}
	kind := on.Kind()
	if trs := t.edges[from][kind]; len(trs) > 0 {
		trs[0].Handler = handler
		t.edges[from][kind] = trs
		return
	}
	t.Add(Transition{From: from, On: on, Handler: handler})
}

// AddReachable declares a from -> to edge as legal without registering
// a full Transition. Useful when the target state is reached by the
// same (from, action-kind) as another edge but via a different Guard,
// or by an Effect chain (WorkloadReady callback) whose Handler is
// declared elsewhere. Also called by Add automatically for the
// (From, To) pair carried by the handler; but the target To is only
// known at handler runtime, so callers register the set of legal
// targets explicitly via AddReachable during table setup.
func (t *TransitionTable) AddReachable(from, to State) {
	byTo := t.reachable[from]
	if byTo == nil {
		byTo = make(map[State]bool)
		t.reachable[from] = byTo
	}
	byTo[to] = true
}

// Candidates returns every transition registered for (from, kind) in
// insertion order. Empty slice when no edge is declared. Callers
// evaluate Guards themselves; see Engine.Step.
func (t *TransitionTable) Candidates(from State, kind ActionKind) []Transition {
	if byKind, ok := t.edges[from]; ok {
		return byKind[kind]
	}
	return nil
}

// Allows reports whether a from -> to transition is declared. Same
// semantics as legacy IsStateTransitionValid.
func (t *TransitionTable) Allows(from, to State) bool {
	if byTo, ok := t.reachable[from]; ok {
		return byTo[to]
	}
	return false
}

// AllowsSelfWrite reports whether a from -> to transition is legal
// under the same relaxed semantics as legacy IsStateTransitionAllowed:
// same-state writes (from == to) are always allowed to support
// idempotent retries and terminal-state re-assertions.
func (t *TransitionTable) AllowsSelfWrite(from, to State) bool {
	if from == to {
		return true
	}
	return t.Allows(from, to)
}

// AllowsOp reports whether operation op is legal in state from. Matches
// legacy IsOperationAllowed / OperationAllowedInState.
func (t *TransitionTable) AllowsOp(from State, op OpType) bool {
	_, ok := t.edges[from][OpKind(op)]
	return ok
}

// SetTimeout enables an automatic deadline for state. A non-positive
// duration disables it. Only states that dispatch CallWorkload are
// enabled by the default table.
func (t *TransitionTable) SetTimeout(state State, d time.Duration) {
	if d <= 0 {
		delete(t.timeouts, state)
		return
	}
	t.timeouts[state] = d
}

// Timeout reports whether state has an automatic deadline and, if so,
// its duration. States that are absent do not time out. Unlike TTL,
// there is no default.
func (t *TransitionTable) Timeout(state State) (time.Duration, bool) {
	d, ok := t.timeouts[state]
	if !ok || d <= 0 {
		return 0, false
	}
	return d, true
}

// TTL returns the timeout for state. Falls back to defaultTTL when
// unset. Matches legacy StateToDuration.
func (t *TransitionTable) TTL(state State) time.Duration {
	if s, ok := t.specs[state]; ok && s.TTL > 0 {
		return s.TTL
	}
	return defaultTTL
}

// IsCancelable reports whether state has an outgoing CancelOp
// transition. Matches legacy IsCancelable / CancelableStates.
func (t *TransitionTable) IsCancelable(state State) bool {
	return t.specs[state].Cancelable
}

// IsOperating reports whether state represents work in progress.
// Matches legacy OperatingStates.
func (t *TransitionTable) IsOperating(state State) bool {
	return t.specs[state].Operating
}

// IsCanceling reports whether state is a *Canceling variant. Matches
// legacy IsCanceling / CancelingStates.
func (t *TransitionTable) IsCanceling(state State) bool {
	return t.specs[state].Canceling
}

// IsTerminalReinstallable matches legacy IsTerminalReinstallable.
func (t *TransitionTable) IsTerminalReinstallable(state State) bool {
	return t.specs[state].TerminalReinstallable
}

// IsSafelyDeletable matches legacy IsSafelyDeletable.
func (t *TransitionTable) IsSafelyDeletable(state State) bool {
	return t.specs[state].SafelyDeletable
}

// States returns every state that has a StateSpec registered. Order is
// unspecified (map iteration). Useful for exhaustive property tests.
func (t *TransitionTable) States() []State {
	out := make([]State, 0, len(t.specs))
	for s := range t.specs {
		out = append(out, s)
	}
	return out
}

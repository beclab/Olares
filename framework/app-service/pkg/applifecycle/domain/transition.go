package domain

import (
	"context"
	"fmt"
	"time"
)

// HandlerFunc is a pure function of (context, World) that returns a
// Result. It MUST NOT mutate any input. See CONTRACT.md for the full
// set of forbidden behaviors.
type HandlerFunc func(ctx context.Context, w World) Result

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

	// Handler is the pure function invoked when this transition fires.
	// Its Result.Next is validated against Table.Allows before the
	// Runner applies the effects.
	Handler HandlerFunc
}

// StateSpec is metadata attached to a state (not to an edge). Kept
// separate from Transition so that a state's classification does
// not have to be repeated on every outgoing edge. Deadlines live on
// TransitionTable.timeouts, not here.
type StateSpec struct {
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

	// Reinstallable is true when the state is a clean terminal from
	// which a fresh Install can be started against the same AM.
	Reinstallable bool

	// SafelyDeletable is true when the AM CR can be deleted directly
	// from this state without leaking cluster resources. Strict
	// subset of Reinstallable. Matches legacy IsSafelyDeletable.
	SafelyDeletable bool
}

// TransitionTable is the single source of truth for the state machine.
// It merges the legacy transition and classification tables.
// All accessors are read-only after construction. Table.Add MUST be
// called only during table setup (in table_default.go's init).
type TransitionTable struct {
	// edges holds the one Transition for each (From, ActionKind).
	// A missing key means that pair has no handler.
	edges map[State]map[ActionKind]Transition

	// reachable[from][to] is true when the table has any registered
	// edge from -> to. Populated by Add so Allows is O(1).
	reachable map[State]map[State]bool

	// specs is the per-state classification metadata.
	specs map[State]StateSpec

	// timeouts is the automatic deadline for states that dispatch
	// CallWorkload. Absence means the state does not time out.
	// There is no default: an unset state such as Running or Stopped
	// must not expire.
	timeouts map[State]time.Duration
}

// NewTransitionTable returns an empty TransitionTable ready for Add.
func NewTransitionTable() *TransitionTable {
	return &TransitionTable{
		edges:     make(map[State]map[ActionKind]Transition),
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
		byKind = make(map[ActionKind]Transition)
		t.edges[tr.From] = byKind
	}
	if _, exists := byKind[kind]; exists {
		panic(fmt.Sprintf("domain: duplicate transition from %q on %q", tr.From, kind))
	}
	byKind[kind] = tr
}

// SetOrAddHandler either replaces the handler on the Transition
// matching (from, on.Kind()), or adds a fresh Transition if none
// exists. This is the recommended entry point for the handlers/
// package during startup: it swaps a PlaceholderHandler for a real
// handler when one already exists (e.g. Installing on WorkloadReady),
// and adds a new edge when none is registered yet (e.g. Installing
// on Tick, which the default table does not declare because no
// legacy Installing → Installing self-loop exists but the
// phase-driven design needs one).
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
	if byKind := t.edges[from]; byKind != nil {
		if tr, ok := byKind[kind]; ok {
			tr.Handler = handler
			byKind[kind] = tr
			return
		}
	}
	t.Add(Transition{From: from, On: on, Handler: handler})
}

// AddReachable declares a from -> to edge as legal without registering
// a full Transition. The single handler for (from, action kind) may
// return this target, or the target is reached by an effect chain
// whose handler is declared elsewhere. Also called by Add automatically
// for the (From, To) pair carried by the handler; but the target To is
// only known at handler runtime, so callers register the set of legal
// targets explicitly via AddReachable during table setup.
func (t *TransitionTable) AddReachable(from, to State) {
	byTo := t.reachable[from]
	if byTo == nil {
		byTo = make(map[State]bool)
		t.reachable[from] = byTo
	}
	byTo[to] = true
}

// Edge returns the Transition registered for (from, kind). The bool
// is false when that pair has no handler.
func (t *TransitionTable) Edge(from State, kind ActionKind) (Transition, bool) {
	byKind, ok := t.edges[from]
	if !ok {
		return Transition{}, false
	}
	tr, ok := byKind[kind]
	return tr, ok
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

// AllowsOp reports whether operation op is legal in state from.
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
// its duration. States that are absent do not time out.
func (t *TransitionTable) Timeout(state State) (time.Duration, bool) {
	d, ok := t.timeouts[state]
	if !ok || d <= 0 {
		return 0, false
	}
	return d, true
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

// IsCanceling reports whether state is a *Canceling variant.
func (t *TransitionTable) IsCanceling(state State) bool {
	return t.specs[state].Canceling
}

func (t *TransitionTable) IsReinstallable(state State) bool {
	return t.specs[state].Reinstallable
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

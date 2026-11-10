package domain

// Result is the return value of every HandlerFunc. It is the ONLY way a
// handler communicates decisions and side effects back to the Runner.
// Handlers MUST NOT mutate cluster state directly; every change is
// expressed either as a state transition (Next) or as an Effect. See
// CONTRACT.md.
type Result struct {
	// Next is the target state after this handler runs. Leave empty
	// (InitialState) to signal "stay in the current state". Engine.Step
	// validates Next against the TransitionTable before returning; a
	// jump to an undeclared state produces ErrUndeclaredTransition.
	Next State

	// Effects is the ordered list of side effects the Runner must apply
	// after the handler returns. The Runner applies them in-order;
	// notably, PatchAM effects merge in any AppendOpRecord effects that
	// precede them so the write to etcd is atomic.
	Effects []Effect

	// Message and Reason mirror ApplicationManagerStatus.{Message,
	// Reason} and are conveniences used by table_default.go's generic
	// PlaceholderHandler to build a PatchAM effect without repetition.
	// Handlers can populate them alongside an explicit PatchAM effect;
	// the Runner reconciles the two by preferring explicit PatchAM
	// contents over the convenience fields.
	Message string
	Reason  string

	// Err carries an unrecoverable error from the handler. When set,
	// Engine.Step returns it verbatim; the Runner returns it to
	// controller-runtime which re-enqueues the AM according to its
	// standard error backoff. Effects are NOT applied when Err is
	// non-nil.
	Err error
}

// WithEffects returns a copy of r with additional effects appended.
// Utility used by handlers that build a Result incrementally.
func (r Result) WithEffects(effs ...Effect) Result {
	r.Effects = append(append([]Effect(nil), r.Effects...), effs...)
	return r
}

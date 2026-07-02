package runner

import (
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// PendingActionQueue holds at most one queued Action per AM. The
// Runner pops the queued Action at the start of each Reconcile,
// defaulting to Tick{} when nothing is queued.
//
// The queue is intentionally single-slot per AM. Multiple pending
// actions collapse to "most recent wins" because:
//
//   - CallWorkload emits at most one OnSuccess or OnFailure per
//     completion, and the AM's own goroutine dispatches them serially
//     (the Runner enforces at-most-one CallWorkload per AM in flight).
//   - TimeoutFired is derived from the state's TTL, which the Runner
//     computes at Reconcile time — the queue never needs to hold
//     multiple timeouts.
//   - User TriggerOp actions arrive via the apiserver, which stores
//     them in the AM.Spec first; the Runner re-derives them from
//     spec on the next Reconcile, so the queue does not need to
//     hold them.
//
// If a follow-up needs richer buffering (e.g. per-op ordered queue),
// swap this type without changing Runner's callers — the Pop/Deliver
// contract is the only public surface.
type PendingActionQueue struct {
	mu      sync.Mutex
	pending map[string]domain.Action
}

// NewPendingActionQueue returns an empty queue.
func NewPendingActionQueue() *PendingActionQueue {
	return &PendingActionQueue{pending: make(map[string]domain.Action)}
}

// Deliver queues an Action for the AM with the given name. Overwrites
// any existing pending action for that AM — "most recent wins", per
// the single-slot rationale in the type doc.
func (q *PendingActionQueue) Deliver(name string, action domain.Action) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending[name] = action
}

// Pop returns the pending Action for name (and clears it), or nil if
// none was queued. Callers substitute Tick{} for nil.
func (q *PendingActionQueue) Pop(name string) domain.Action {
	q.mu.Lock()
	defer q.mu.Unlock()
	a, ok := q.pending[name]
	if !ok {
		return nil
	}
	delete(q.pending, name)
	return a
}

// Restore puts action back in the slot for name after a Reconcile
// failed to consume it (optimistic-concurrency conflict, busy
// download slot, or a hard dispatch error). It reports whether the
// action was inserted. It is a no-op when the slot already holds
// something: a CallWorkload goroutine that completed during the
// failed Reconcile carries fresher intent than the action we are
// unwinding, and "most recent wins" still applies.
//
// Without this, Pop's destructive read drops the callback and forces
// phase recovery to redispatch work that already completed. Restoring
// it preserves the callback within this Runner and avoids duplicate
// at-least-once execution.
func (q *PendingActionQueue) Restore(name string, action domain.Action) bool {
	if action == nil {
		return false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.pending[name]; ok {
		return false
	}
	q.pending[name] = action
	return true
}

// Peek returns the pending Action for name without clearing it.
// Exposed for tests and metrics.
func (q *PendingActionQueue) Peek(name string) (domain.Action, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	a, ok := q.pending[name]
	return a, ok
}

// Requeuer is the seam through which the Runner pokes controller-
// runtime's workqueue after a CallWorkload goroutine has delivered
// an Action into PendingActionQueue.
//
// In production the adapter wraps
// controller-runtime's `source.Channel` — dropping a struct on the
// channel schedules a synthetic Reconcile against the AM.
//
// In tests a fake Requeuer records every Requeue call so the test
// can drive the follow-up Reconcile deterministically.
type Requeuer interface {
	// Requeue schedules a Reconcile for the AM named `name`.
	// Non-blocking; must be safe to call from a Runner-owned
	// goroutine.
	Requeue(name string)
}

// RequeuerFunc adapts a plain function into the Requeuer interface.
// Convenient for tests: `runner.RequeuerFunc(func(n string){...})`.
type RequeuerFunc func(name string)

// Requeue implements Requeuer.
func (f RequeuerFunc) Requeue(name string) { f(name) }

// NoopRequeuer is a Requeuer that discards every call. Used in
// synchronous tests that drive Reconcile manually and don't need
// the async workqueue path.
type NoopRequeuer struct{}

// Requeue implements Requeuer.
func (NoopRequeuer) Requeue(_ string) {}

// RecordingRequeuer records every Requeue call for test assertions.
// Safe for concurrent use.
type RecordingRequeuer struct {
	mu    sync.Mutex
	calls []string
}

// Requeue implements Requeuer.
func (r *RecordingRequeuer) Requeue(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, name)
}

// Calls returns a copy of the recorded requeue names, in call order.
func (r *RecordingRequeuer) Calls() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.calls))
	copy(out, r.calls)
	return out
}

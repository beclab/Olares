package apiserver

import (
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// OpDispatcher is the sole seam between HTTP handlers and the
// state-machine Runner. HTTP handlers depend on this interface —
// NOT on *runner.Runner — so they can be exercised with a fake in
// unit tests and so the Runner's public surface stays free of
// apiserver-shaped helper methods.
//
// Trigger is intentionally single-shot and side-effect-only. It
// requeues the AM so controller-runtime schedules a Reconcile
// immediately, without waiting on informer resync. The Action
// argument records what the caller requested; the Runner does not
// read it from the in-memory queue. The operation is already
// persisted on the AM (Spec.OpType and the operation-request-id
// annotation), and an empty-queue Reconcile re-derives the
// TriggerOp from that. Delivering the same action as well races
// with that watch reconcile and can drop the Tick that starts the
// work.
//
// Trigger MUST NOT block, MUST NOT return an error, and MUST be
// safe to call from any goroutine. If Requeue can fail in
// production (e.g. workqueue backpressure), the adapter is
// responsible for degrading gracefully — the HTTP layer must be
// able to respond to the caller regardless.
type OpDispatcher interface {
	// Trigger requeues the AM named `name`. Non-blocking. action
	// is the caller's record of the requested operation.
	Trigger(name string, action domain.Action)
}

// RunnerDispatcher adapts a Requeuer into an OpDispatcher.
//
// In production, main.go instantiates one Runner and one Requeuer
// (a source.Channel wrapper), passes the same Requeuer into
// runner.Config so async CallWorkload callbacks share the
// workqueue, and passes the Runner+Requeuer here so HTTP triggers
// wake that same Runner. The Runner argument is the wiring
// invariant: a nil Runner means there is nothing to wake.
type RunnerDispatcher struct {
	requeuer runner.Requeuer
}

// NewRunnerDispatcher constructs a RunnerDispatcher from a Runner
// and a Requeuer. Panics when either argument is nil — a nil
// requeuer means HTTP triggers would be dropped on the floor with
// no observable failure, so failing loudly at wire-up time is
// safer than degrading silently at request time.
func NewRunnerDispatcher(r *runner.Runner, req runner.Requeuer) *RunnerDispatcher {
	if r == nil {
		panic("apiserver.NewRunnerDispatcher: runner is nil")
	}
	if req == nil {
		panic("apiserver.NewRunnerDispatcher: requeuer is nil")
	}
	return &RunnerDispatcher{requeuer: req}
}

// Trigger implements OpDispatcher. The persisted AM operation is the
// source of truth, so this only wakes the Runner.
func (d *RunnerDispatcher) Trigger(name string, _ domain.Action) {
	d.requeuer.Requeue(name)
}

// DispatchCall is one recorded Trigger invocation, exposed for
// test assertions.
type DispatchCall struct {
	Name   string
	Action domain.Action
}

// FakeDispatcher is the test double every HTTP handler unit test
// should use. It records every Trigger call in-order, so a test
// can assert:
//
//   - that a handler triggered EXACTLY the expected op (no extra
//     rogue triggers)
//   - that the (name, action) pair matches the expected AM +
//     TriggerOp{Op}
//   - that the handler did NOT trigger when it should have
//     rejected the request (e.g. auth failure, AllowsOp
//     rejection)
//
// FakeDispatcher is safe for concurrent use — the HTTP layer never
// dispatches concurrently for a single request, but production-
// shaped tests (e.g. httptest with a real router) may.
//
// Zero value is ready to use.
type FakeDispatcher struct {
	mu    sync.Mutex
	calls []DispatchCall
}

// Trigger implements OpDispatcher.
func (f *FakeDispatcher) Trigger(name string, action domain.Action) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, DispatchCall{Name: name, Action: action})
}

// Calls returns a copy of every Trigger invocation, in call order.
// The returned slice is safe to inspect without holding any lock.
func (f *FakeDispatcher) Calls() []DispatchCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]DispatchCall, len(f.calls))
	copy(out, f.calls)
	return out
}

// Reset clears the recorded call history. Useful when a single
// test reuses a dispatcher across multiple sub-tests.
func (f *FakeDispatcher) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = nil
}

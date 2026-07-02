// Package apiserver defines the thin translator layer between the
// public HTTP API (pkg/apiserver) and the state-machine Runner
// (pkg/applifecycle/runner).
//
// # Purpose
//
// Legacy HTTP handlers used to do three things at once when a user
// hit /apps/{name}/install (or /uninstall, /cancel, /suspend, ...):
//
//  1. write the user's intent into ApplicationManager.Spec (op-type,
//     annotations)
//  2. reach past the state machine and set
//     ApplicationManager.Status.State directly (e.g. to Uninstalling,
//     PendingCanceling, ApplyingEnv, ...)
//  3. rely on the reconcile controller to notice the new state and
//     start the appropriate side effects
//
// Step (2) is the source of every "apiserver wrote X, controller
// wrote Y, they raced" bug the refactor is meant to eliminate. In the
// model there is ONE writer for Status — the Runner —
// and every user intent is persisted on the AM. The Runner
// re-derives the TriggerOp from that on the next Reconcile.
//
// This package packages the missing translator step so HTTP handlers
// can express:
//
//	// spec is user-owned — HTTP handler still writes it directly.
//	ctrlClient.Update(ctx, amWithNewOpTypeAndAnnotations)
//
//	// wake the Runner — it re-derives the operation from the AM.
//	dispatcher.Trigger(name, domain.TriggerOp{Op: appv1.CancelOp})
//
// Trigger only Requeue()s the AM into the controller-runtime
// workqueue. The in-memory queue is reserved for workload
// callbacks. A second copy of the user operation there races with
// the watch reconcile and can skip the Tick that starts the work.
//
// # What lives here vs. what does not
//
// This package OWNS:
//
//   - OpDispatcher: the single interface HTTP handlers depend on.
//   - RunnerDispatcher: production adapter that binds a
//     *runner.Runner + runner.Requeuer into an OpDispatcher.
//   - FakeDispatcher: in-memory recorder used by handler unit tests.
//
// This package DOES NOT OWN:
//
//   - AM Spec / annotation writes. Those stay on the apiserver's
//     controller-runtime client, matching the "spec = user-owned"
//     invariant.
//   - HTTP request/response marshalling. Handlers live in
//     pkg/apiserver and call into this package only for the
//     translator step.
//   - Any Reconcile-time behaviour. That is the Runner's job; this
//     package just hands it the Action.
//
// # Wiring
//
// In production, main.go constructs a single *runner.Runner and a
// single controller-runtime source.Channel-backed Requeuer, wraps
// them via NewRunnerDispatcher, and passes the resulting
// OpDispatcher into pkg/apiserver's handlerBuilder via
// WithOpDispatcher (see pkg/apiserver/handler.go).
//
// The Runner's internal async-callback path (CallWorkload → Deliver
// → Requeue) MUST use the same Requeuer instance so a single
// workqueue drives every reconcile, whether the trigger comes from
// an HTTP request or a background goroutine.
//
// # Testing
//
// Handler tests inject FakeDispatcher and assert on Calls(). Runner
// tests keep using runner.RecordingRequeuer / NoopRequeuer — the two
// test doubles serve different layers of the stack and are not
// interchangeable.
package apiserver

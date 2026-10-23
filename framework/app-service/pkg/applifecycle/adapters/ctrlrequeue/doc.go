// Package ctrlrequeue implements runner.Requeuer on top of a
// controller-runtime source.Channel.
//
// # Purpose
//
// The Runner's async-callback path (CallWorkload → OnSuccess /
// OnFailure) and the apiserver's OpDispatcher.Trigger both need a
// way to poke controller-runtime's workqueue after they deliver an
// Action into the PendingActionQueue — otherwise the delivered
// Action sits there until the next informer resync or periodic tick
// picks up the AM, defeating the point of dispatching in the first
// place.
//
// This package packages the standard controller-runtime idiom for
// that pattern: a chan of TypedGenericEvent[*ApplicationManager]
// consumed by a source.Channel wired into the reconciler builder.
// Writers Requeue(name) and the reconciler picks up a
// reconcile.Request keyed by that name on the next queue tick.
//
// # Why one Requeuer shared across all writers
//
// Wiring invariant: exactly ONE Requeuer instance drives the
// Runner's reconciler. That means:
//
//   - runner.Config.Requeuer is set to this instance so async
//     CallWorkload callbacks re-enter through the same workqueue
//     the periodic Reconcile does.
//   - apiserver.NewRunnerDispatcher receives the same instance so
//     HTTP-triggered Actions land in the same queue.
//
// If the two paths ever ended up on different Requeuers, an async
// callback could enqueue a Reconcile that never picks up the HTTP-
// delivered Action (or vice versa) — silent starvation with no
// observable failure. main.go enforces the shared-instance
// invariant at wire-up time.
//
// # Backpressure
//
// The underlying channel is buffered (see defaultBufferSize). If
// writers overrun the buffer — highly unlikely under any realistic
// load, since each Requeue costs a single AM re-enqueue and the
// reconciler is designed to drain within seconds — Requeue drops
// the event rather than blocking. A dropped event is a latency
// regression, not a correctness one: the next unrelated Reconcile
// or informer resync still picks up any state the writer intended
// to communicate through PendingActionQueue.
//
// # Contract with controller-runtime
//
// The GenericEvent path was chosen (over `source.Kind` for a real
// object) because our writers know only the AM name, not the full
// object. Enqueuing by name is exactly what controller-runtime's
// EnqueueRequestsFromMapFunc supports, so we synthesize a stub
// ApplicationManager with Name set and let the map-func extract
// the name into a reconcile.Request. This mirrors the way the
// upstream controller-runtime docs demonstrate wiring external
// webhooks — the well-trodden idiom.
package ctrlrequeue

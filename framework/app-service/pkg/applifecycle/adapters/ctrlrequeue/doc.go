// Package ctrlrequeue implements runner.Requeuer on top of a
// controller-runtime source.Channel.
//
// # Purpose
//
// The Runner's async-callback path (CallWorkload → OnSuccess /
// OnFailure) delivers an Action into the PendingActionQueue and
// then needs to poke controller-runtime's workqueue. The
// ApplicationManager has not changed, so the informer will not
// enqueue. Without Requeue the delivered Action sits until the
// next informer event.
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
// Runner's reconciler. runner.Config.Requeuer and the Source
// registered by ctrladapter.Setup must be that same instance.
// CallWorkload callbacks re-enter through the workqueue the
// informer-driven Reconcile uses. A second Requeuer would enqueue
// a Reconcile that never pops the in-memory Action.
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

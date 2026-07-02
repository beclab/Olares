// Package runner is the "impure shell" that wires together every
// external I/O boundary and the pure domain engine.
//
// It is the ONLY package permitted to:
//
//   - call ports.AMStore / WorkloadOps / ImagePuller / etc.
//   - spawn goroutines
//   - call time.After via Clock (via the injected port)
//   - translate handler-returned domain.Effect into real side effects
//
// Handlers stay pure per CONTRACT.md. All impurity concentrates here.
//
// # High-level Reconcile flow
//
//	                   ┌──────────────────────────────────────┐
//	                   │  controller-runtime (or test)        │
//	                   │  Reconcile(ctx, req)                 │
//	                   └──────────────┬───────────────────────┘
//	                                  │
//	┌─────────────────────────────────▼──────────────────────────────────┐
//	│  1. Materialise World                                              │
//	│       AM     ← AMStore.Get(req.Name)                               │
//	│       AppExists  ← AMStore.GetApplication(req.Name) NotFound?      │
//	│       Now         ← Clock.Now()                                    │
//	│       ClusterStats.DownloadingCount ← AMStore.List() reduce         │
//	│  2. Pop pending Action (default: Tick{})                           │
//	│  3. result, err ← Engine.Step(world, action)                       │
//	│  4. Interpret err                                                  │
//	│       ErrNoTransition / ErrGuardRejected → success no-op           │
//	│       ErrUndeclaredTransition / handler err → propagate            │
//	│  5. Dispatch result.Effects in order (see effects.go)              │
//	│  6. Return reconcile.Result{RequeueAfter: ...}                     │
//	└────────────────────────────────────────────────────────────────────┘
//
// # Effect dispatch (see effects.go for detail)
//
//   - AppendOpRecord ─── buffered; merged into the immediately-following
//     PatchAM so the write to etcd is atomic
//   - PatchAM       ─── AMStore.PatchStatus with the current RV; conflict
//     surfaces as reconcile.Result{Requeue: true}
//   - EmitEvent     ─── EventRecorder.RecordAppEvent (blocking; adapter
//     owns any internal queue)
//   - ScheduleTick  ─── accumulates the minimum RequeueAfter seen; the
//     Reconcile result carries it back to controller-
//     runtime
//   - CallWorkload  ─── spawned as a Runner-owned goroutine that (a)
//     routes to the correct port (see pending.go),
//     then (b) enqueues OnSuccess / OnFailure via
//     PendingActionQueue.Deliver(req.Name, action) and
//     (c) requeues the AM into the controller-runtime
//     workqueue via the injected Requeuer
//
// # Async callback delivery
//
// Long-running effects (CallWorkload) do not block Reconcile. When the
// workload terminates the runner-owned goroutine hands the resulting
// Action to PendingActionQueue and pokes the workqueue via
// Requeuer.Requeue(req). The next Reconcile pops that Action instead
// of the default Tick{} and dispatches it through the same pipeline.
//
// This matches the "controller persistence" pattern and keeps
// Reconcile itself synchronous — the only concurrency the runner adds
// is one goroutine per in-flight CallWorkload.
package runner

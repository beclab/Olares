// Package ctrladapter is the one-directional bridge between
// controller-runtime's reconcile.Reconciler and the applifecycle
// Runner.
//
// # Why a separate package
//
// pkg/applifecycle/runner deliberately avoids importing
// controller-runtime so the Runner and its handler contract stay
// framework-agnostic (this is the same reason runner.Request +
// runner.ReconcileOutcome exist instead of just re-exporting
// ctrl.Request / ctrl.Result). This sub-package is where those two
// types meet — everything ctrl-shaped lives here, everything
// state-machine-shaped lives one directory up. The seam is one
// file, ~20 lines wide.
//
// # What lives here
//
//   - Reconciler: satisfies reconcile.Reconciler by delegating
//     Reconcile calls to *runner.Runner and mapping
//     ReconcileOutcome → reconcile.Result.
//   - Setup: registers a Reconciler with a controller-runtime
//     Manager, watching BOTH the ApplicationManager informer
//     (For) AND the shared Requeuer.Source() (WatchesRawSource)
//     so HTTP-triggered Actions and async CallWorkload callbacks
//     land in the same workqueue as regular informer events.
//
// # What does NOT live here
//
//   - Runner construction: main.go owns runner.New (it holds the
//     port adapters + configuration).
//   - Port adapters: pkg/applifecycle/adapters/* (helm, image,
//     nats, ...). Reconciler doesn't know they exist.
//   - Requeuer construction: pkg/applifecycle/adapters/ctrlrequeue
//     owns that. Setup consumes it.
//   - Dispatcher construction: pkg/applifecycle/apiserver owns
//     the RunnerDispatcher that HTTP handlers depend on.
//
// # Coexistence with the legacy controller
//
// Named "app-lifecycle-runner" so metrics + logs are visibly
// distinct from the legacy "application-manager-controller".
// Both run side-by-side during the shadow-mode migration; while
// the Runner is wired with fake ports (or a read-only k8sstore),
// the legacy controller remains the sole writer of
// ApplicationManager.Status. Only after every port is real and
// the routes flip does the legacy controller retire.
package ctrladapter

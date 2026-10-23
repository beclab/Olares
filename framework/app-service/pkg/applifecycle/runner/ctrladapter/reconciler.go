package ctrladapter

import (
	"context"
	"errors"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/ctrlrequeue"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// controllerName is the Named() value passed to the builder.
// Distinct from the legacy "application-manager-controller" so
// metrics + logs unambiguously identify which reconciler is
// touching an AM during the shadow-mode window when both run.
const controllerName = "app-lifecycle-runner"

// Reconciler is the reconcile.Reconciler shell around a
// *runner.Runner. Reconciler holds NO state of its own — every
// per-AM decision goes through Runner.Reconcile, so this type is
// purely a signature adapter.
type Reconciler struct {
	// Runner is the state-machine Runner. Required; a nil Runner
	// makes Reconcile return an error rather than panicking so
	// controller-runtime's error backoff surfaces the
	// misconfiguration in logs instead of the process dying.
	Runner *runner.Runner
}

// Reconcile implements reconcile.Reconciler. It delegates to
// Runner.Reconcile and maps the returned ReconcileOutcome onto a
// reconcile.Result:
//
//   - outcome.RequeueAfter  → Result.RequeueAfter (used by every
//     ScheduleTick effect the state machine emits).
//   - outcome.Requeue       → Result.Requeue (set on
//     optimistic-concurrency conflicts; the workqueue's
//     rate-limiter provides backoff on repeated conflicts, which
//     is exactly the safety net we want here — a Millisecond-
//     valued RequeueAfter would bypass that rate-limiting).
//
// The Result.Requeue field is deprecated in newer controller-
// runtime releases in favour of RequeueAfter, but it is the ONLY
// documented way to trigger a rate-limited immediate requeue and
// remains supported in v0.22.x; we accept the deprecation warning
// on that one field until upstream provides a replacement.
//
// Namespace is intentionally ignored: ApplicationManager is
// cluster-scoped, so req.Namespace is always "" and
// runner.Request only carries a name.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r.Runner == nil {
		return reconcile.Result{}, errors.New("ctrladapter.Reconciler: Runner is nil")
	}
	outcome, err := r.Runner.Reconcile(ctx, runner.Request{Name: req.Name})
	if err != nil {
		return reconcile.Result{}, err
	}
	return outcomeToResult(outcome), nil
}

// outcomeToResult is the pure mapping ReconcileOutcome →
// reconcile.Result. Extracted from Reconciler.Reconcile so the
// mapping can be unit-tested as a table without spinning up a
// Runner + fake ports.
//
// Split rationale: everything the Reconciler adds ON TOP of the
// Runner is contained in this one function. If the mapping ever
// grows a wrinkle (e.g. deriving Requeue from a specific
// RequeueAfter sentinel), the change lives here and stays covered
// by outcomeToResult tests alone.
func outcomeToResult(o runner.ReconcileOutcome) reconcile.Result {
	return reconcile.Result{
		Requeue:      o.Requeue,
		RequeueAfter: o.RequeueAfter,
	}
}

// Setup wires a Reconciler into the given controller-runtime
// Manager. The registered controller watches TWO sources:
//
//   - For(&ApplicationManager{}) — informer-driven events (Create /
//     Update / Delete on any AM in the cluster).
//   - WatchesRawSource(requeuer.Source()) — synthetic events from
//     the shared Requeuer, i.e. every apiserver dispatcher.Trigger
//     call AND every Runner-internal CallWorkload callback that
//     needs to poke the workqueue after delivering an Action.
//
// The `requeuer` argument MUST be the same instance passed to
// runner.Config.Requeuer AND to apiserver.NewRunnerDispatcher.
// Any drift here means synthetic Actions land in a queue nothing
// is reading (silent starvation). See ctrlrequeue.doc.go for the
// full invariant.
//
// Setup returns an error on nil Runner / nil Requeuer instead of
// panicking so the caller (main.go) can bail out through the
// existing setupLog + os.Exit(1) path uniformly with every other
// controller registration.
func Setup(mgr ctrl.Manager, r *runner.Runner, requeuer *ctrlrequeue.Requeuer) error {
	if mgr == nil {
		return errors.New("ctrladapter.Setup: manager is nil")
	}
	if r == nil {
		return errors.New("ctrladapter.Setup: runner is nil")
	}
	if requeuer == nil {
		return errors.New("ctrladapter.Setup: requeuer is nil")
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&appv1alpha1.ApplicationManager{}).
		WatchesRawSource(requeuer.Source()).
		Complete(&Reconciler{Runner: r})
}

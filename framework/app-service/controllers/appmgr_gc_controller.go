package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ApplicationManagerGCReconciler reclaims ApplicationManager objects that have
// been sitting in an IsSafelyDeletable terminal state longer than
// constants.AppMgrTerminalRetention.
//
// Background:
//   - The state machine itself never deletes the AM CR — terminal states are
//     kept as a reinstall anchor and a UI/CLI history record (so users can
//     still see "this install failed because X").
//   - The deletion responsibility lives in two places: lazy cleanup in
//     checkAppNameConflict at install time, and this controller for the
//     "no one ever tried to reinstall the same app name" steady-state case.
//
// Safety rails:
//   - Only AMs whose Status.State is IsSafelyDeletable are considered. That
//     predicate already guarantees the state machine has run a full cleanup
//     and confirmed the app namespace is gone before transitioning into the
//     state.
//   - On top of that, before deleting we double-check the namespace really is
//     IsNotFound. This protects against the rare "InstallFailed-cleanup-timeout"
//     case (NS finalizer stuck past 5min): in that scenario the AM is still
//     in InstallFailed and IsSafelyDeletable, but the namespace is technically
//     still around. Deleting the AM there would orphan the NS from any
//     state-machine retry, blocking future same-name installs until manual
//     cleanup. We skip and requeue instead, letting InstallFailedApp.Exec
//     keep retrying until the NS finalizer releases.
type ApplicationManagerGCReconciler struct {
	client.Client
}

// SetupWithManager wires the GC controller into the manager.
//
// Event predicates:
//   - Create: only enqueued when the object already lands in an
//     IsSafelyDeletable state. New AMs created by the normal install flow
//     start at "" / Pending so they get filtered out; the predicate exists
//     specifically for informer initial-sync after a controller restart,
//     where existing safely-deletable AMs are delivered as Create events
//     and would otherwise never be reconciled until something updates them.
//   - Update: enqueued only when Status.State actually changes into an
//     IsSafelyDeletable state. Skipping no-op state Updates keeps the
//     workqueue quiet while still reacting promptly to terminal transitions.
//   - Delete: never enqueued — Delete is the action the GC controller
//     produces, not one it reacts to.
func (r *ApplicationManagerGCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("app-manager-gc-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              r,
	})
	if err != nil {
		return fmt.Errorf("app manager gc setup failed %w", err)
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&appv1alpha1.ApplicationManager{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, h *appv1alpha1.ApplicationManager) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: h.GetName()}}}
			}),
		predicate.TypedFuncs[*appv1alpha1.ApplicationManager]{
			CreateFunc: func(e event.TypedCreateEvent[*appv1alpha1.ApplicationManager]) bool {
				return appstate.IsSafelyDeletable(e.Object.Status.State)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*appv1alpha1.ApplicationManager]) bool {
				if e.ObjectOld.Status.State == e.ObjectNew.Status.State {
					return false
				}
				return appstate.IsSafelyDeletable(e.ObjectNew.Status.State)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*appv1alpha1.ApplicationManager]) bool {
				return false
			},
		},
	))
	if err != nil {
		return fmt.Errorf("appmgr gc watch failed %w", err)
	}
	return nil
}

func (r *ApplicationManagerGCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var am appv1alpha1.ApplicationManager
	if err := r.Get(ctx, req.NamespacedName, &am); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !appstate.IsSafelyDeletable(am.Status.State) {
		return ctrl.Result{}, nil
	}

	entered := terminalStateEntryTime(&am)
	if entered.IsZero() {
		// Without a timestamp we cannot compute age; treat as freshly-entered
		// and requeue once after the full retention.
		klog.V(4).Infof("appmgr-gc: %s in %s has no StatusTime/UpdateTime, deferring full retention", am.Name, am.Status.State)
		return ctrl.Result{RequeueAfter: constants.AppMgrTerminalRetention}, nil
	}

	age := time.Since(entered)
	if age < constants.AppMgrTerminalRetention {
		return ctrl.Result{RequeueAfter: constants.AppMgrTerminalRetention - age}, nil
	}

	// Double-check the namespace is gone before deleting the AM. The state
	// machine guarantees this for the happy path, but the InstallFailed
	// cleanup-timeout edge case can leave NS lingering while the AM is in
	// InstallFailed and IsSafelyDeletable says "deletable". Deleting the AM
	// while the NS is still around would orphan it from InstallFailedApp.Exec's
	// retry loop — refuse and let Exec continue trying.
	if nsGone, err := isAppNamespaceGone(ctx, r.Client, &am); err != nil {
		klog.Warningf("appmgr-gc: %s namespace check failed (will retry): %v", am.Name, err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if !nsGone {
		klog.Infof("appmgr-gc: %s in %s eligible by age (%s) but namespace %q still present; deferring to InstallFailedApp.Exec",
			am.Name, am.Status.State, age, am.Spec.AppNamespace)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	klog.Infof("appmgr-gc: deleting %s (state=%s, owner=%s, age=%s)",
		am.Name, am.Status.State, am.Spec.AppOwner, age)
	if err := r.Delete(ctx, &am); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	return ctrl.Result{}, nil
}

// terminalStateEntryTime returns when the AM most recently entered its current
// terminal state. StatusTime is the canonical "when did state become X" stamp
// and is updated alongside every transition that goes through updateStatus;
// UpdateTime is the broader "this AM was touched" stamp and serves as a
// fallback for legacy data that may predate StatusTime.
func terminalStateEntryTime(am *appv1alpha1.ApplicationManager) time.Time {
	if am.Status.StatusTime != nil && !am.Status.StatusTime.IsZero() {
		return am.Status.StatusTime.Time
	}
	if am.Status.UpdateTime != nil && !am.Status.UpdateTime.IsZero() {
		return am.Status.UpdateTime.Time
	}
	return time.Time{}
}

// isAppNamespaceGone reports whether the app's namespace truly no longer
// exists. Protected namespaces (user-space and friends) are never created or
// torn down by the app lifecycle, so they always read as "gone" for the
// purposes of GC eligibility. An empty AppNamespace means the install never
// got far enough to claim one (e.g. failure in pre-helm validation) — also
// treat as gone.
func isAppNamespaceGone(ctx context.Context, c client.Client, am *appv1alpha1.ApplicationManager) (bool, error) {
	if am.Spec.AppNamespace == "" {
		return true, nil
	}
	if apputils.IsProtectedNamespace(am.Spec.AppNamespace) {
		return true, nil
	}
	var ns corev1.Namespace
	err := c.Get(ctx, types.NamespacedName{Name: am.Spec.AppNamespace}, &ns)
	if apierrors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

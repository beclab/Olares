package controllers

import (
	"context"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const overlayMACFinalizer = "app.bytetrade.io/overlay-mac-claim"

// OverlayMACLifecycleReconciler keeps an Application's MAC claim until all
// macvlan pods have disappeared. The Application owner reference on the claim
// then lets Kubernetes garbage-collect the allocation safely.
type OverlayMACLifecycleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.bytetrade.io,resources=applications,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=app.bytetrade.io,resources=applications/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
func (r *OverlayMACLifecycleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	app := &appv1alpha1.Application{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name}, app); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if app.DeletionTimestamp.IsZero() || !containsFinalizer(app.Finalizers, overlayMACFinalizer) {
		return ctrl.Result{}, nil
	}

	pods := &corev1.PodList{}
	if err := r.List(ctx, pods,
		client.InNamespace(app.Spec.Namespace),
		client.MatchingLabels{
			constants.ApplicationNameLabel:        app.Spec.Name,
			constants.ApplicationOwnerLabel:       app.Spec.Owner,
			constants.ApplicationMacvlanInitLabel: "true",
		},
	); err != nil {
		return ctrl.Result{}, err
	}
	if len(pods.Items) != 0 {
		// A terminating Pod can still own a macvlan link and DHCP lease. Wait
		// for it to disappear from the API before allowing owner GC.
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	copy := app.DeepCopy()
	copy.Finalizers = removeFinalizer(copy.Finalizers, overlayMACFinalizer)
	if err := r.Update(ctx, copy); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *OverlayMACLifecycleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Application{}).
		Complete(r)
}

func containsFinalizer(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func removeFinalizer(values []string, value string) []string {
	result := make([]string, 0, len(values))
	for _, existing := range values {
		if existing != value {
			result = append(result, existing)
		}
	}
	return result
}

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const overlayMACFinalizer = "app.bytetrade.io/overlay-mac-claim"
const (
	overlayMACPendingPhase        = "Pending"
	overlayMACBoundPhase          = "Bound"
	overlayMACPendingGracePeriod  = time.Minute
	overlayMACPendingRequeue      = 30 * time.Second
	overlayMACApplicationUIDLabel = "app.bytetrade.io/overlay-mac-application-uid"
)

var overlayMACAllocationGVR = schema.GroupVersionResource{
	Group:    "app.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "overlaymacallocations",
}

// OverlayMACLifecycleReconciler keeps an Application's MAC claim until all
// macvlan pods have disappeared. The Application owner reference on the claim
// then lets Kubernetes garbage-collect the allocation safely.
type OverlayMACLifecycleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	DynamicClient dynamic.Interface
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
	if !containsFinalizer(app.Finalizers, overlayMACFinalizer) {
		return ctrl.Result{}, nil
	}
	if app.DeletionTimestamp.IsZero() {
		if r.DynamicClient == nil {
			return ctrl.Result{}, fmt.Errorf("dynamic client is required for pending overlay MAC reconciliation")
		}
		if err := r.reconcilePendingAllocations(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: overlayMACPendingRequeue}, nil
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

func (r *OverlayMACLifecycleReconciler) reconcilePendingAllocations(ctx context.Context, app *appv1alpha1.Application) error {
	allocations, err := r.DynamicClient.Resource(overlayMACAllocationGVR).List(ctx, metav1.ListOptions{
		LabelSelector: overlayMACApplicationUIDLabel + "=" + string(app.UID),
	})
	if err != nil {
		return err
	}
	for i := range allocations.Items {
		allocation := &allocations.Items[i]
		phase, _, _ := unstructured.NestedString(allocation.Object, "spec", "phase")
		applicationUID, _, _ := unstructured.NestedString(allocation.Object, "spec", "applicationUID")
		applicationRef, _, _ := unstructured.NestedString(allocation.Object, "spec", "applicationRef")
		if phase != overlayMACPendingPhase ||
			applicationUID != string(app.UID) ||
			applicationRef != app.Name {
			continue
		}
		mac, _, _ := unstructured.NestedString(allocation.Object, "spec", "mac")
		instanceKey, _, _ := unstructured.NestedString(allocation.Object, "spec", "instanceKey")
		if applicationReferencesMAC(app, instanceKey, mac) {
			copy := allocation.DeepCopy()
			if err := unstructured.SetNestedField(copy.Object, overlayMACBoundPhase, "spec", "phase"); err != nil {
				return err
			}
			if _, err := r.DynamicClient.Resource(overlayMACAllocationGVR).Update(ctx, copy, metav1.UpdateOptions{}); err != nil {
				return err
			}
			continue
		}
		createdAt := allocation.GetCreationTimestamp()
		if createdAt.IsZero() || time.Since(createdAt.Time) < overlayMACPendingGracePeriod {
			continue
		}
		uid := allocation.GetUID()
		if uid == "" {
			return fmt.Errorf("pending overlay MAC allocation %s has no UID", allocation.GetName())
		}
		if err := r.DynamicClient.Resource(overlayMACAllocationGVR).Delete(ctx, allocation.GetName(), metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func applicationReferencesMAC(app *appv1alpha1.Application, instanceKey, mac string) bool {
	if app.Spec.Settings == nil {
		return false
	}
	if app.Spec.Settings["overlayMacvlanMac"] == mac {
		return true
	}
	raw := app.Spec.Settings["overlayMacvlanMacByOrdinal"]
	if raw == "" {
		return false
	}
	values := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return false
	}
	parts := strings.Split(instanceKey, "/")
	ordinal := parts[len(parts)-1]
	return values[ordinal] == mac
}

func (r *OverlayMACLifecycleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("overlay-mac-lifecycle").
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

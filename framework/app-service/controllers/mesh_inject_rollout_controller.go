package controllers

import (
	"context"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MeshInjectRolloutReconciler watches Linkerd control-plane readiness and, on
// the first Ready transition (false→true), enqueues a rate-limited sweep of
// Applications / Shared namespaces / gateway dataplane workloads. It never
// mutates pods on the request path.
type MeshInjectRolloutReconciler struct {
	Client       client.Client
	Kube         kubernetes.Interface
	Worker       *meshinagent.RolloutWorker
	requeueAfter time.Duration
}

// Reconcile polls mesh control-plane readiness and persists epoch in a ConfigMap.
func (r *MeshInjectRolloutReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	after := r.requeueAfter
	if after <= 0 {
		after = 30 * time.Second
	}
	if r.Client == nil || r.Kube == nil {
		return ctrl.Result{RequeueAfter: after}, nil
	}
	worker := r.Worker
	if worker == nil {
		worker = meshinagent.DefaultWorker()
	}

	ready := meshinagent.IsMeshControlPlaneReady(ctx, r.Kube)
	wasReady, epoch, err := meshinagent.LoadMeshInjectRolloutState(ctx, r.Client)
	if err != nil {
		klog.Errorf("mesh-in-rollout: mesh-inject load state failed: %v", err)
		return ctrl.Result{RequeueAfter: after}, err
	}

	switch {
	case ready && !wasReady:
		newEpoch := meshinagent.NextMeshReadyEpoch(epoch)
		if err := meshinagent.StoreMeshInjectRolloutState(ctx, r.Client, true, newEpoch); err != nil {
			return ctrl.Result{RequeueAfter: after}, err
		}
		klog.Infof("mesh-in-rollout: mesh control plane first ready epoch=%s → sweep", newEpoch)
		if worker != nil {
			worker.EnqueueMeshReadySweep(ctx, newEpoch)
		}
	case !ready && wasReady:
		if err := meshinagent.StoreMeshInjectRolloutState(ctx, r.Client, false, epoch); err != nil {
			return ctrl.Result{RequeueAfter: after}, err
		}
		klog.Warningf("mesh-in-rollout: mesh control plane became not-ready (epoch=%s retained)", epoch)
	case ready && wasReady:
		// Already tracked.
	default:
		if epoch == "" || epoch == "0" {
			_ = meshinagent.StoreMeshInjectRolloutState(ctx, r.Client, false, "0")
		}
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

// SetupWithManager watches os-mesh Linkerd deployments and polls periodically.
func (r *MeshInjectRolloutReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mesh-inject-rollout").
		For(&appsv1.Deployment{}, builder.WithPredicates(meshNamespaceDeployPredicate())).
		Complete(r)
}

func meshNamespaceDeployPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object != nil && e.Object.GetNamespace() == mesh.LinkerdNamespace
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew != nil && e.ObjectNew.GetNamespace() == mesh.LinkerdNamespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object != nil && e.Object.GetNamespace() == mesh.LinkerdNamespace
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return e.Object != nil && e.Object.GetNamespace() == mesh.LinkerdNamespace
		},
	}
}

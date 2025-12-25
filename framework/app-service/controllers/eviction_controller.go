package controllers

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// EvictionManagerController manages the status of app entrance
type EvictionManagerController struct {
	client.Client
}

func (r *EvictionManagerController) SetUpWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("eviction-manager-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&corev1.Pod{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, pod *corev1.Pod) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				}}}
			}),
		predicate.TypedFuncs[*corev1.Pod]{
			CreateFunc: func(e event.TypedCreateEvent[*corev1.Pod]) bool {
				return true
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*corev1.Pod]) bool {
				return true
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*corev1.Pod]) bool {
				return false
			},
		},
	))
	if err != nil {
		klog.Errorf("entrance-status-manager-controller failed to watch err=%v", err)
		return err
	}
	return nil
}

func (r *EvictionManagerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	klog.Infof("reconcile entrance-status-manager request name=%v", req.Name)
	var pod corev1.Pod
	err := r.Get(ctx, req.NamespacedName, &pod)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	nss := append(security.UnderLayerNamespaces, security.GPUSystemNamespaces...)
	ignoredNs := sets.NewString(nss...)

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	if podNamespace == "" || ignoredNs.Has(podNamespace) {
		return ctrl.Result{}, nil
	}

	if pod.Status.Reason != "Evicted" {
		return ctrl.Result{}, nil
	}
	klog.Infof("pod.Name=%s, pod.Namespace=%s,pod.Status.Reason=%s,message=%s", podName, podNamespace, pod.Status.Reason, pod.Status.Message)

	var nodes corev1.NodeList
	err = r.List(ctx, &nodes, &client.ListOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}
	canScheduleNodes := 0
	for _, node := range nodes.Items {
		if utils.IsNodeReady(&node) && !node.Spec.Unschedulable {
			canScheduleNodes++
		}
	}
	if canScheduleNodes > 1 {
		return ctrl.Result{}, nil
	}
	appName := pod.GetLabels()[constants.ApplicationNameLabel]
	owner := pod.GetLabels()[constants.ApplicationOwnerLabel]
	if appName != "" || owner != "" {
		return ctrl.Result{}, nil
	}
	_, err = r.setDeployOrStsReplicas(ctx, podName, podNamespace, int32(0))
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Delete(ctx, &pod)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EvictionManagerController) setDeployOrStsReplicas(ctx context.Context, podName, namespace string, replicas int32) (bool, error) {
	var pod corev1.Pod
	key := types.NamespacedName{Name: podName, Namespace: namespace}
	err := r.Get(ctx, key, &pod)
	if err != nil {
		return false, err
	}
	if len(pod.OwnerReferences) == 0 {
		return true, nil
	}
	var kind, name string
	ownerRef := pod.OwnerReferences[0]
	switch ownerRef.Kind {
	case "ReplicaSet":
		key = types.NamespacedName{Namespace: namespace, Name: ownerRef.Name}
		var rs appsv1.ReplicaSet
		err = r.Get(ctx, key, &rs)
		if err != nil {
			return false, err
		}
		if len(rs.OwnerReferences) > 0 && rs.OwnerReferences[0].Kind == deployment {
			kind = deployment
			name = rs.OwnerReferences[0].Name
		}
	case statefulSet:
		kind = statefulSet
		name = ownerRef.Name
	}
	if kind == "" {
		return true, nil
	}
	switch kind {
	case deployment:
		var deploy appsv1.Deployment
		key = types.NamespacedName{Name: name, Namespace: namespace}
		err = r.Get(ctx, key, &deploy)
		if err != nil {
			return false, err
		}
		deployCopy := deploy.DeepCopy()
		deployCopy.Spec.Replicas = &replicas

		err = r.Patch(ctx, deployCopy, client.MergeFrom(&deploy))
		if err != nil {
			return false, err
		}
	case statefulSet:
		var sts appsv1.StatefulSet
		key = types.NamespacedName{Name: name, Namespace: namespace}
		err = r.Get(ctx, key, &sts)
		if err != nil {
			return false, err
		}
		stsCopy := sts.DeepCopy()
		stsCopy.Spec.Replicas = &replicas

		err = r.Patch(ctx, stsCopy, client.MergeFrom(&sts))
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

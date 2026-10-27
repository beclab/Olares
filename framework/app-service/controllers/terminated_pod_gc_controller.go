package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	envTerminatedPodGCMinAge     = "TERMINATED_POD_GC_MIN_AGE"
	envTerminatedPodGCReadyDelay = "TERMINATED_POD_GC_READY_DELAY"
)

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch

// TerminatedPodGCController removes terminal Pods of any Deployment in the
// cluster once that Deployment has as many stably ready replicas as it wants. A
// Deployment scaled to zero satisfies that trivially, so leftovers of a stopped
// workload are collected too.
//
// Unknown Pods are deliberately excluded. Their workload might still be running
// on an unreachable node, and there is nothing left for this controller to do
// anyway: Kubernetes counts an Unknown Pod as active, so a ReplicaSet whose
// desired count is below its active count deletes such a Pod itself, and while a
// ReplicaSet is satisfied no replacement exists for the readiness gate to
// accept. Pods stuck on a lost node are reclaimed by taint eviction and the
// built-in PodGC.
//
// Reaching a Deployment through the ownership chain keeps Pods that no
// Deployment governs out of scope, so completed Job Pods and bare Pods are
// never touched.
type TerminatedPodGCController struct {
	client.Client
	minPodAge  time.Duration
	readyDelay time.Duration
	now        func() time.Time
}

func (r *TerminatedPodGCController) SetupWithManager(mgr ctrl.Manager) error {
	r.minPodAge = parseTimeWithDefault(envTerminatedPodGCMinAge, time.Minute)
	r.readyDelay = parseTimeWithDefault(envTerminatedPodGCReadyDelay, time.Minute)
	if r.now == nil {
		r.now = time.Now
	}

	klog.Infof(
		"terminated-pod-gc-controller initialized, minPodAge=%v, readyDelay=%v",
		r.minPodAge,
		r.readyDelay,
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named("terminated-pod-gc-controller").
		For(&appsv1.Deployment{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForPod),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

func (r *TerminatedPodGCController) requestsForPod(ctx context.Context, obj client.Object) []reconcile.Request {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}

	podOwner := metav1.GetControllerOf(pod)
	if podOwner == nil || podOwner.Kind != replicaSet {
		return nil
	}

	var rs appsv1.ReplicaSet
	if err := r.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: podOwner.Name}, &rs); err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("failed to resolve ReplicaSet for pod namespace=%s name=%s: %v", pod.Namespace, pod.Name, err)
		}
		return nil
	}
	if podOwner.UID != rs.UID {
		return nil
	}

	rsOwner := metav1.GetControllerOf(&rs)
	if rsOwner == nil || rsOwner.Kind != deployment {
		return nil
	}

	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      rsOwner.Name,
	}}}
}

func (r *TerminatedPodGCController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deploy); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if deploy.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, err
	}

	var podList corev1.PodList
	if err := r.List(
		ctx,
		&podList,
		client.InNamespace(deploy.Namespace),
		client.MatchingLabelsSelector{Selector: selector},
	); err != nil {
		return ctrl.Result{}, err
	}

	// Resolving ownership costs a ReplicaSet lookup per Pod, so skip it for the
	// overwhelmingly common case of a workload with nothing to collect.
	if !anyCollectablePod(podList.Items) {
		return ctrl.Result{}, nil
	}

	pods, err := r.podsControlledByDeployment(ctx, &deploy, podList.Items)
	if err != nil {
		return ctrl.Result{}, err
	}

	var terminalPods []*corev1.Pod
	for _, pod := range pods {
		if isCollectablePod(pod) {
			terminalPods = append(terminalPods, pod)
		}
	}
	if len(terminalPods) == 0 {
		return ctrl.Result{}, nil
	}

	// Collect only once the Deployment carries its full complement of replicas
	// that have stayed ready long enough to be trusted. A Deployment scaled to
	// zero needs none, which is what makes stopped apps collectable.
	now := r.currentTime()
	var stablyReady int32
	var nextCheck time.Duration
	for _, pod := range pods {
		readySince, ready := runningAndReadySince(pod)
		if !ready {
			continue
		}
		remaining := r.readyDelay - now.Sub(readySince)
		if remaining <= 0 {
			stablyReady++
			continue
		}
		nextCheck = shorterPositiveDuration(nextCheck, remaining)
	}
	if stablyReady < deploymentReplicas(&deploy) {
		return ctrl.Result{RequeueAfter: nextCheck}, nil
	}
	nextCheck = 0

	for _, pod := range terminalPods {
		remaining := r.minPodAge - now.Sub(terminalSince(pod))
		if remaining > 0 {
			nextCheck = shorterPositiveDuration(nextCheck, remaining)
			continue
		}

		uid := pod.UID
		err := r.Delete(ctx, pod, &client.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		klog.Infof(
			"deleted terminal pod namespace=%s name=%s phase=%s deployment=%s",
			pod.Namespace,
			pod.Name,
			pod.Status.Phase,
			deploy.Name,
		)
	}

	return ctrl.Result{RequeueAfter: nextCheck}, nil
}

func (r *TerminatedPodGCController) podsControlledByDeployment(
	ctx context.Context,
	deploy *appsv1.Deployment,
	pods []corev1.Pod,
) ([]*corev1.Pod, error) {
	replicaSets := make(map[string]*appsv1.ReplicaSet)
	missingReplicaSets := make(map[string]struct{})
	result := make([]*corev1.Pod, 0, len(pods))

	for i := range pods {
		pod := &pods[i]
		owner := metav1.GetControllerOf(pod)
		if owner == nil || owner.Kind != replicaSet {
			continue
		}
		if _, missing := missingReplicaSets[owner.Name]; missing {
			continue
		}

		rs, found := replicaSets[owner.Name]
		if !found {
			var fetched appsv1.ReplicaSet
			err := r.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: owner.Name}, &fetched)
			if apierrors.IsNotFound(err) {
				missingReplicaSets[owner.Name] = struct{}{}
				continue
			}
			if err != nil {
				return nil, err
			}
			rs = &fetched
			replicaSets[owner.Name] = rs
		}
		if owner.UID != rs.UID {
			continue
		}

		rsOwner := metav1.GetControllerOf(rs)
		if rsOwner == nil || rsOwner.Kind != deployment || rsOwner.UID != deploy.UID {
			continue
		}
		result = append(result, pod)
	}

	return result, nil
}

func (r *TerminatedPodGCController) currentTime() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

func deploymentReplicas(deploy *appsv1.Deployment) int32 {
	if deploy.Spec.Replicas == nil {
		return 1
	}
	return *deploy.Spec.Replicas
}

func isTerminalPod(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded
}

func isCollectablePod(pod *corev1.Pod) bool {
	return isTerminalPod(pod) && pod.DeletionTimestamp == nil
}

func anyCollectablePod(pods []corev1.Pod) bool {
	for i := range pods {
		if isCollectablePod(&pods[i]) {
			return true
		}
	}
	return false
}

func runningAndReadySince(pod *corev1.Pod) (time.Time, bool) {
	if pod.DeletionTimestamp != nil || pod.Status.Phase != corev1.PodRunning {
		return time.Time{}, false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			if condition.LastTransitionTime.IsZero() {
				return pod.CreationTimestamp.Time, true
			}
			return condition.LastTransitionTime.Time, true
		}
	}
	return time.Time{}, false
}

func terminalSince(pod *corev1.Pod) time.Time {
	var since time.Time
	for _, status := range pod.Status.InitContainerStatuses {
		since = laterTime(since, terminatedAt(status))
	}
	for _, status := range pod.Status.ContainerStatuses {
		since = laterTime(since, terminatedAt(status))
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady &&
			condition.Status == corev1.ConditionFalse &&
			!condition.LastTransitionTime.IsZero() {
			since = laterTime(since, condition.LastTransitionTime.Time)
		}
	}
	if since.IsZero() {
		return pod.CreationTimestamp.Time
	}
	return since
}

func terminatedAt(status corev1.ContainerStatus) time.Time {
	if status.State.Terminated == nil || status.State.Terminated.FinishedAt.IsZero() {
		return time.Time{}
	}
	return status.State.Terminated.FinishedAt.Time
}

func laterTime(left, right time.Time) time.Time {
	if right.After(left) {
		return right
	}
	return left
}

func shorterPositiveDuration(current, candidate time.Duration) time.Duration {
	if candidate <= 0 {
		return current
	}
	if current <= 0 || candidate < current {
		return candidate
	}
	return current
}

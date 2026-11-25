package controllers

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	envTerminatedPodGCMinAge     = "TERMINATED_POD_GC_MIN_AGE"
	envTerminatedPodGCReadyDelay = "TERMINATED_POD_GC_READY_DELAY"
	overlayTerminalPodRequeue    = 30 * time.Second
)

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applications,verbs=get;list;watch

// TerminatedPodGCController removes terminal Pods of any Deployment in the
// cluster once that Deployment has as many stably ready replicas as it wants. A
// Deployment scaled to zero satisfies that trivially, so leftovers of a stopped
// workload are collected too.
// Platform-managed overlay singletons are an exception: their fixed MAC blocks
// replacement admission until the old Pod disappears. On a Ready node, collect
// Pods rejected during recovery without waiting for a replacement, provided
// kubelet reports all containers terminated and the sandbox unready. Sandbox
// unready is not a CNI DEL receipt; teardown ordering remains a hardware test.
// Admission keeps the MAC reserved until the old Pod disappears.
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
	APIReader  client.Reader
	minPodAge  time.Duration
	readyDelay time.Duration
	now        func() time.Time
}

func (r *TerminatedPodGCController) SetupWithManager(mgr ctrl.Manager) error {
	r.APIReader = mgr.GetAPIReader()
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
		Watches(
			&appv1alpha1.Application{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForOverlayApplication),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
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

// An Application can become available or enable Overlay after its Pod enters
// a terminal state. Reconcile its workloads without polling unverified Pods.
func (r *TerminatedPodGCController) requestsForOverlayApplication(ctx context.Context, obj client.Object) []reconcile.Request {
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app.Spec.Namespace == "" || app.Spec.Name == "" || app.Spec.Owner == "" ||
		app.Spec.Settings["enableOverlayGateway"] != "true" {
		return nil
	}

	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(app.Spec.Namespace)); err != nil {
		klog.Errorf("failed to list Deployments for overlay application %s: %v", app.Name, err)
		return nil
	}
	var requests []reconcile.Request
	for _, deploy := range deployments.Items {
		labels := deploy.Spec.Template.Labels
		if labels[constants.ApplicationNameLabel] == app.Spec.Name && labels[constants.ApplicationOwnerLabel] == app.Spec.Owner {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&deploy)})
		}
	}
	return requests
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
	fullyReady := stablyReady >= deploymentReplicas(&deploy)
	if fullyReady {
		nextCheck = 0
	}

	var overlayApp *appv1alpha1.Application
	if slices.ContainsFunc(terminalPods, hasOverlayNetworkMarkers) {
		overlayApp, err = r.enabledOverlayApplication(ctx, &deploy)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	for _, pod := range terminalPods {
		overlayPod := overlayApp != nil && hasOverlayNetworkMarkers(pod)
		if overlayPod {
			// Use live reads for recovery safety, not informer snapshots. A
			// terminal Pod update must be reconsidered before it can be deleted.
			nextCheck = shorterPositiveDuration(nextCheck, overlayTerminalPodRequeue)
			var current corev1.Pod
			if err := r.liveReader().Get(ctx, client.ObjectKeyFromObject(pod), &current); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
			if current.UID != pod.UID || current.ResourceVersion != pod.ResourceVersion || !isCollectablePod(&current) {
				continue
			}
			if !podSandboxStopped(pod) || !allPodContainersTerminated(pod) {
				continue
			}
			ready, err := r.terminalPodNodeReady(ctx, pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			if !ready {
				continue
			}
		}
		if !fullyReady {
			// Limit the readiness exception to the kubelet admission failure
			// observed when restarting an existing workload.
			if !overlayPod || overlayApp.DeletionTimestamp != nil ||
				pod.Status.Phase != corev1.PodFailed || pod.Status.Reason != "UnexpectedAdmissionError" {
				continue
			}
			overlay, err := r.isOverlaySingleton(ctx, &deploy, pod, overlayApp)
			if err != nil {
				return ctrl.Result{}, err
			}
			if !overlay {
				continue
			}
		}
		remaining := r.minPodAge - now.Sub(terminalSince(pod, overlayPod))
		if remaining > 0 {
			nextCheck = shorterPositiveDuration(nextCheck, remaining)
			continue
		}

		uid := pod.UID
		preconditions := &metav1.Preconditions{UID: &uid}
		if overlayPod {
			rv := pod.ResourceVersion
			preconditions.ResourceVersion = &rv
		}
		err := r.Delete(ctx, pod, &client.DeleteOptions{Preconditions: preconditions})
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

func (r *TerminatedPodGCController) liveReader() client.Reader {
	if r.APIReader != nil {
		return r.APIReader
	}
	return r.Client
}

func (r *TerminatedPodGCController) enabledOverlayApplication(ctx context.Context, deploy *appsv1.Deployment) (*appv1alpha1.Application, error) {
	name, owner := deploy.Spec.Template.Labels[constants.ApplicationNameLabel], deploy.Spec.Template.Labels[constants.ApplicationOwnerLabel]
	if name == "" || owner == "" {
		return nil, nil
	}
	appName, err := apputils.FmtAppMgrName(name, owner, deploy.Namespace)
	if err != nil {
		return nil, err
	}
	var app appv1alpha1.Application
	if err := r.liveReader().Get(ctx, types.NamespacedName{Name: appName}, &app); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if app.UID == "" || app.Spec.Namespace != deploy.Namespace || app.Spec.Name != name ||
		app.Spec.Owner != owner || app.Spec.Settings["enableOverlayGateway"] != "true" {
		return nil, nil
	}
	return &app, nil
}

// Recheck the live ownership chain and template identity before bypassing the
// replacement readiness gate.
func (r *TerminatedPodGCController) isOverlaySingleton(ctx context.Context, deploy *appsv1.Deployment, pod *corev1.Pod, app *appv1alpha1.Application) (bool, error) {
	var current appsv1.Deployment
	if err := r.liveReader().Get(ctx, client.ObjectKeyFromObject(deploy), &current); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if current.UID != deploy.UID || current.DeletionTimestamp != nil {
		return false, nil
	}
	deploy = &current
	if deploymentReplicas(deploy) != 1 || deploy.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
		return false, nil
	}
	name, owner := pod.Labels[constants.ApplicationNameLabel], pod.Labels[constants.ApplicationOwnerLabel]
	if name != app.Spec.Name || owner != app.Spec.Owner || deploy.Spec.Template.Labels[constants.ApplicationNameLabel] != name ||
		deploy.Spec.Template.Labels[constants.ApplicationOwnerLabel] != owner {
		return false, nil
	}
	controller := metav1.GetControllerOf(pod)
	if controller == nil || controller.APIVersion != "apps/v1" || controller.Kind != replicaSet {
		return false, nil
	}
	var rs appsv1.ReplicaSet
	if err := r.liveReader().Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: controller.Name}, &rs); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	parent := metav1.GetControllerOf(&rs)
	if rs.UID != controller.UID || parent == nil || parent.APIVersion != "apps/v1" || parent.Kind != deployment || parent.UID != deploy.UID || parent.Name != deploy.Name {
		return false, nil
	}
	// The Application may have been disabled or replaced while checking the Pod.
	verifiedApp, err := r.enabledOverlayApplication(ctx, deploy)
	return verifiedApp != nil && verifiedApp.UID == app.UID && verifiedApp.DeletionTimestamp == nil, err
}

func (r *TerminatedPodGCController) terminalPodNodeReady(ctx context.Context, pod *corev1.Pod) (bool, error) {
	if pod.Spec.NodeName == "" {
		return false, nil
	}
	var node corev1.Node
	if err := r.liveReader().Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if node.DeletionTimestamp != nil {
		return false, nil
	}
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue, nil
		}
	}
	return false, nil
}

func allPodContainersTerminated(pod *corev1.Pod) bool {
	terminated := func(names []string, statuses []corev1.ContainerStatus) bool {
		finished := make(map[string]bool, len(statuses))
		for _, status := range statuses {
			if status.State.Terminated == nil || status.State.Running != nil || status.State.Waiting != nil {
				return false
			}
			finished[status.Name] = true
		}
		for _, name := range names {
			if !finished[name] {
				return false
			}
		}
		return true
	}
	names := func(containers []corev1.Container) []string {
		result := make([]string, 0, len(containers))
		for _, container := range containers {
			result = append(result, container.Name)
		}
		return result
	}
	ephemeral := make([]string, 0, len(pod.Spec.EphemeralContainers))
	for _, container := range pod.Spec.EphemeralContainers {
		ephemeral = append(ephemeral, container.Name)
	}
	return len(pod.Spec.Containers) > 0 &&
		terminated(names(pod.Spec.Containers), pod.Status.ContainerStatuses) &&
		terminated(names(pod.Spec.InitContainers), pod.Status.InitContainerStatuses) &&
		terminated(ephemeral, pod.Status.EphemeralContainerStatuses)
}

func podSandboxStopped(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReadyToStartContainers {
			return condition.Status == corev1.ConditionFalse && !condition.LastTransitionTime.IsZero()
		}
	}
	return false
}

// Labels can be changed after admission. Network selections and the immutable
// injected init container identify candidates for MAC safety checks. Only a
// verified Overlay-enabled Application can opt a candidate into those checks.
func hasOverlayNetworkMarkers(pod *corev1.Pod) bool {
	if pod.Labels[constants.ApplicationMacvlanInitLabel] == "true" {
		return true
	}
	for _, container := range pod.Spec.InitContainers {
		if container.Name == constants.MacvlanReplyInitContainerName {
			return true
		}
	}

	overlayNetwork := constants.OverlayMacvlanNetworkNamespace + "/" + constants.OverlayMacvlanNetworkName
	networks := pod.Annotations["k8s.v1.cni.cncf.io/networks"]
	if networks == overlayNetwork || pod.Annotations["v1.multus-cni.io/default-network"] == overlayNetwork {
		return true
	}
	if networks == "" {
		return false
	}
	var selections []struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if json.Unmarshal([]byte(networks), &selections) == nil {
		for _, selection := range selections {
			if selection.Name == constants.OverlayMacvlanNetworkName && selection.Namespace == constants.OverlayMacvlanNetworkNamespace {
				return true
			}
		}
	}
	return false
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

func terminalSince(pod *corev1.Pod, overlay bool) time.Time {
	var since time.Time
	for _, status := range pod.Status.InitContainerStatuses {
		since = laterTime(since, terminatedAt(status))
	}
	for _, status := range pod.Status.ContainerStatuses {
		since = laterTime(since, terminatedAt(status))
	}
	if overlay {
		for _, status := range pod.Status.EphemeralContainerStatuses {
			since = laterTime(since, terminatedAt(status))
		}
	}
	for _, condition := range pod.Status.Conditions {
		if (condition.Type == corev1.PodReady || (overlay && condition.Type == corev1.PodReadyToStartContainers)) &&
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

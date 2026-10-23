package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	lifecycleapi "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/google/uuid"

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

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applicationmanagers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applicationmanagers/status,verbs=get;update;patch

const (
	envPendingPodSuspendAppTimeout = "PENDING_POD_SUSPEND_APP_TIMEOUT"
	envPendingPodMinRequeueDelay   = "PENDING_POD_MIN_REQUEUE_DELAY"
)

// PodAbnormalSuspendAppController watches Pods belonging to applications and suspends the app
// when a Pod is Evicted, or when Pending and Unschedulable beyond a timeout.
type PodAbnormalSuspendAppController struct {
	client.Client
	// Ops delivers StopOp into the applifecycle Runner. The controller
	// writes Spec and annotations only; the Runner owns the
	// Running→Stopping transition and its state-change event.
	Ops                 lifecycleapi.OpDispatcher
	pendingTimeout      time.Duration
	minRequeueDelay     time.Duration
	controllerStartedAt time.Time
}

func (r *PodAbnormalSuspendAppController) SetUpWithManager(mgr ctrl.Manager) error {
	r.controllerStartedAt = time.Now()
	r.pendingTimeout = parseTimeWithDefault(envPendingPodSuspendAppTimeout, 3*time.Minute)
	r.minRequeueDelay = parseTimeWithDefault(envPendingPodMinRequeueDelay, 3*time.Second)
	if r.minRequeueDelay > r.pendingTimeout {
		r.minRequeueDelay = r.pendingTimeout
	}
	c, err := controller.New("pod-abnormal-suspend-app-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	klog.Infof("pod-abnormal-suspend-app-controller initialized, startedAt=%s, pendingTimeout=%v, minRequeueDelay=%v", r.controllerStartedAt.Format(time.RFC3339), r.pendingTimeout, r.minRequeueDelay)

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&corev1.Pod{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, pod *corev1.Pod) []reconcile.Request {
				if !hasRequiredAppLabels(pod) {
					return nil
				}
				klog.Infof("enqueue pod for abnormal check name=%s namespace=%s app=%s owner=%s", pod.Name, pod.Namespace, pod.GetLabels()[constants.ApplicationNameLabel], pod.GetLabels()[constants.ApplicationOwnerLabel])
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				}}}
			}),
		predicate.TypedFuncs[*corev1.Pod]{
			CreateFunc: func(e event.TypedCreateEvent[*corev1.Pod]) bool {
				return hasRequiredAppLabels(e.Object)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*corev1.Pod]) bool {
				return hasRequiredAppLabels(e.ObjectNew)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*corev1.Pod]) bool {
				return false
			},
		},
	))
	if err != nil {
		klog.Errorf("pod-abnormal-suspend-app-controller failed to watch err=%v", err)
		return err
	}
	err = mgr.GetFieldIndexer().IndexField(context.Background(),
		&corev1.Event{},
		"involvedObject.name",
		func(obj client.Object) []string {
			event := obj.(*corev1.Event)
			if event.InvolvedObject.Name != "" {
				return []string{event.InvolvedObject.Name}
			}
			return nil
		},
	)
	if err != nil {
		klog.Errorf("pod-abnormal-suspend-app-controller failed to set index err=%v", err)
		return err
	}
	return nil
}

func hasRequiredAppLabels(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	name := pod.GetLabels()[constants.ApplicationNameLabel]
	owner := pod.GetLabels()[constants.ApplicationOwnerLabel]
	return name != "" && owner != ""
}

func (r *PodAbnormalSuspendAppController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	klog.Infof("reconcile pod name=%s namespace=%s phase=%s reason=%s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status.Reason)

	appName := pod.GetLabels()[constants.ApplicationNameLabel]
	owner := pod.GetLabels()[constants.ApplicationOwnerLabel]
	if appName == "" || owner == "" {
		klog.Infof("ignore pod name=%s namespace=%s due to missing app labels", pod.Name, pod.Namespace)
		return ctrl.Result{}, nil
	}

	if pod.DeletionTimestamp != nil {
		klog.Infof("ignore pod name=%s namespace=%s because it is being deleted", pod.Name, pod.Namespace)
		return ctrl.Result{}, nil
	}

	if pod.Status.Reason == "Evicted" {
		klog.Infof("pod evicted name=%s namespace=%s, attempting to suspend app=%s owner=%s", pod.Name, pod.Namespace, appName, owner)
		ok, err := r.trySuspendApp(ctx, owner, appName, constants.AppStopDueToEvicted, "evicted pod: "+pod.Namespace+"/"+pod.Name, pod.Namespace)
		if err != nil {
			klog.Errorf("suspend attempt failed for app=%s owner=%s: %v", appName, owner, err)
			return ctrl.Result{}, err
		}
		if !ok {
			klog.Infof("app not suspendable yet app=%s owner=%s, requeue after 5s", appName, owner)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		// An Evicted pod is a one-shot signal: once acted on, drop it so a later
		// resync cannot replay it against a healthy app. Nothing else reclaims
		// it — the eviction controller skips pods carrying app labels, and a
		// suspended app runs at replicas=0.
		klog.Infof("deleting evicted pod name=%s namespace=%s app=%s owner=%s", pod.Name, pod.Namespace, appName, owner)
		if err := r.deleteEvictedPod(ctx, &pod); err != nil {
			klog.Errorf("failed to delete evicted pod name=%s namespace=%s: %v", pod.Name, pod.Namespace, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if isScheduled(&pod) {
		klog.Infof("ignore pod name=%s namespace=%s because it is scheduled", pod.Name, pod.Namespace)
		return ctrl.Result{}, nil
	}

	pendingKind, err := utils.GetPendingKind(r.Client, &pod)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, found := pendingUnschedulableSince(&pod)
	insufficientGPU := pendingKind == utils.PendingKindInSufficientGPU
	if found || insufficientGPU {
		reason := constants.AppUnschedulable
		if insufficientGPU {
			reason = constants.AppHamiSchedulable
		} else {
			now := time.Now()
			pendingSince := pendingTimeoutStart(&pod, r.controllerStartedAt)
			elapsed := now.Sub(pendingSince)
			if elapsed < 0 {
				elapsed = 0
			}
			klog.Infof("pod pending unschedulable name=%s namespace=%s since=%s elapsed=%v timeout=%v", pod.Name, pod.Namespace, pendingSince.Format(time.RFC3339), elapsed, r.pendingTimeout)
			if elapsed < r.pendingTimeout {
				delay := r.pendingTimeout - elapsed
				if delay > r.minRequeueDelay {
					delay = r.minRequeueDelay
				}
				klog.Infof("requeue pod name=%s namespace=%s after %v until timeout", pod.Name, pod.Namespace, delay)
				return ctrl.Result{RequeueAfter: delay}, nil
			}
		}

		klog.Infof("attempting to suspend app=%s owner=%s due to pending unschedulable timeout", appName, owner)
		ok, err := r.trySuspendApp(ctx, owner, appName, reason, "pending unschedulable timeout on pod: "+pod.Namespace+"/"+pod.Name, pod.Namespace)
		if err != nil {
			klog.Errorf("suspend attempt failed for app=%s owner=%s: %v", appName, owner, err)
			return ctrl.Result{}, err
		}
		if !ok {
			klog.Infof("app not suspendable yet app=%s owner=%s, requeue after 5s", appName, owner)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	return ctrl.Result{}, nil
}

func pendingTimeoutStart(pod *corev1.Pod, controllerStartedAt time.Time) time.Time {
	if pod.CreationTimestamp.IsZero() || pod.CreationTimestamp.Time.Before(controllerStartedAt) {
		return controllerStartedAt
	}
	return pod.CreationTimestamp.Time
}

func (r *PodAbnormalSuspendAppController) deleteEvictedPod(ctx context.Context, pod *corev1.Pod) error {
	if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func isScheduled(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	if pod.Status.Phase != corev1.PodPending {
		return true
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodScheduled {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func pendingUnschedulableSince(pod *corev1.Pod) (time.Time, bool) {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodScheduled && c.Status == corev1.ConditionFalse && c.Reason == corev1.PodReasonUnschedulable {
			if !c.LastTransitionTime.IsZero() {
				return c.LastTransitionTime.Time, true
			}
			// fallback to creation time if transition is missing
			return pod.CreationTimestamp.Time, true
		}
	}
	return time.Time{}, false
}

// trySuspendApp attempts to suspend the app and returns (true, nil) if a suspend request was issued.
// If the app is not suspendable yet, returns (false, nil) to trigger a short requeue.
func (r *PodAbnormalSuspendAppController) trySuspendApp(ctx context.Context, owner, appName, reason, message, podNamespace string) (bool, error) {
	// Resolve to the live AM (shared cluster-wide or per-user) — a pod-abnormal
	// suspend on a shared app needs to find the {app}-shared-{app} AM no matter
	// which user the pod is running under.
	name, _, err := apputils.ResolveAppMgrName(ctx, appName, owner)
	if err != nil {
		klog.Errorf("failed to resolve app manager name app=%s owner=%s: %v", appName, owner, err)
		return false, err
	}
	var am appv1alpha1.ApplicationManager
	if err := r.Get(ctx, types.NamespacedName{Name: name}, &am); err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("applicationmanager not found name=%s for app=%s owner=%s", name, appName, owner)
			return false, nil
		}
		klog.Errorf("failed to get applicationmanager name=%s for app=%s owner=%s: %v", name, appName, owner, err)
		return false, err
	}
	if stopAlreadyInFlight(&am) {
		return true, nil
	}
	// Status still describes the previous operation. Requeue instead of
	// replacing an unconsumed Uninstall, Upgrade, or ApplyEnv. After the
	// Runner consumes it, AllowsOp decides whether this pod can still stop
	// the app.
	if deferSuspend(&am) {
		op, _ := domain.UnconsumedOperation(&am)
		klog.Infof("app %s owner %s has unconsumed %s request, deferring suspend", appName, owner, op)
		return false, nil
	}

	if !domain.DefaultTable.AllowsOp(am.Status.State, appv1alpha1.StopOp) {
		klog.Infof("operation StopOp not allowed in state=%s for app=%s owner=%s", am.Status.State, appName, owner)
		return false, nil
	}
	if r.Ops == nil {
		return false, fmt.Errorf("pod abnormal suspend: op dispatcher is nil")
	}
	if err := r.suspendResolved(ctx, &am, reason, message, podNamespace); err != nil {
		return false, err
	}
	klog.Infof("suspend requested for app=%s owner=%s, reason=%s", am.Spec.AppName, am.Spec.AppOwner, message)
	return true, nil
}

// stopAlreadyInFlight reports that a stop is accepted or finished.
// Status stays on the previous operation until the Runner copies the
// durable request into Status.OpID, so Spec.OpType plus an unconsumed
// request id counts too. Another evicted pod must not mint a new id.
func stopAlreadyInFlight(am *appv1alpha1.ApplicationManager) bool {
	return am.Status.OpType == appv1alpha1.StopOp ||
		am.Status.State == appv1alpha1.Stopped ||
		am.Status.State == appv1alpha1.Stopping ||
		domain.OperationRequestPending(am, appv1alpha1.StopOp)
}

// deferSuspend reports a non-stop request the Runner has not consumed.
// trySuspendApp requeues those instead of minting a Stop that erases them.
func deferSuspend(am *appv1alpha1.ApplicationManager) bool {
	op, pending := domain.UnconsumedOperation(am)
	return pending && op != appv1alpha1.StopOp
}

// suspendResolved writes the stop intent and delivers StopOp. Status
// stays where it is: OnRunningStop (and the other StopOp entries)
// perform the transition and publish the event, using the reason
// annotations so an eviction is not reported as a user stop.
func (r *PodAbnormalSuspendAppController) suspendResolved(ctx context.Context, am *appv1alpha1.ApplicationManager, reason, message, podNamespace string) error {
	patched := am.DeepCopy()
	if patched.Annotations == nil {
		patched.Annotations = make(map[string]string)
	}
	patched.Spec.OpType = appv1alpha1.StopOp
	if strings.HasSuffix(podNamespace, "-shared") {
		patched.Annotations[api.AppStopAllKey] = "true"
	}
	if reason == constants.AppHamiSchedulable || reason == constants.AppUnschedulable {
		patched.Annotations[api.AppStopByControllerDuePendingPod] = "true"
	}
	patched.Annotations[api.AppStopReasonKey] = reason
	patched.Annotations[api.AppStopMessageKey] = message
	patched.Annotations[domain.AnnotationOperationRequestID] = uuid.NewString()
	if err := r.Patch(ctx, patched, client.MergeFrom(am)); err != nil {
		klog.Errorf("failed to patch applicationmanager stop intent name=%s: %v", am.Name, err)
		return err
	}
	r.Ops.Trigger(am.Name, domain.TriggerOp{Op: appv1alpha1.StopOp})
	return nil
}

func parseTimeWithDefault(name string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(name)
	if v == "" {
		klog.Infof("%s not set, using default %v", name, defaultValue)
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		klog.Warningf("invalid %s value %q, using default %v", name, v, defaultValue)
		return defaultValue
	}
	klog.Infof("%s set to %v", name, d)
	return d
}

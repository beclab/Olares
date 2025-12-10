package controllers

import (
	"context"
	"os"
	"strconv"
	"time"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/appstate"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

// PodAbnormalSuspendAppController watches Pods belonging to applications and suspends the app
// when a Pod is Evicted, or when Pending and Unschedulable beyond a timeout.
type PodAbnormalSuspendAppController struct {
	client.Client
	pendingTimeout time.Duration
}

func (r *PodAbnormalSuspendAppController) SetUpWithManager(mgr ctrl.Manager) error {
	r.pendingTimeout = parsePendingTimeout(os.Getenv(envPendingPodSuspendAppTimeout))

	c, err := controller.New("pod-abnormal-suspend-app-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	klog.Infof("pod-abnormal-suspend-app-controller initialized, pendingTimeout=%v", r.pendingTimeout)

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
		ok, err := r.trySuspendApp(ctx, owner, appName, constants.AppStopDueToEvicted, "evicted pod: "+pod.Namespace+"/"+pod.Name)
		if err != nil {
			klog.Errorf("suspend attempt failed for app=%s owner=%s: %v", appName, owner, err)
			return ctrl.Result{}, err
		}
		if !ok {
			klog.Infof("app not suspendable yet app=%s owner=%s, requeue after 5s", appName, owner)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, nil
	}

	if isScheduled(&pod) {
		klog.Infof("ignore pod name=%s namespace=%s because it is scheduled", pod.Name, pod.Namespace)
		return ctrl.Result{}, nil
	}

	if pendingSince, found := pendingUnschedulableSince(&pod); found {
		elapsed := time.Since(pendingSince)
		klog.Infof("pod pending unschedulable name=%s namespace=%s since=%s elapsed=%v timeout=%v", pod.Name, pod.Namespace, pendingSince.Format(time.RFC3339), elapsed, r.pendingTimeout)
		if elapsed < r.pendingTimeout {
			delay := r.pendingTimeout - elapsed
			klog.Infof("requeue pod name=%s namespace=%s after %v until timeout", pod.Name, pod.Namespace, delay)
			return ctrl.Result{RequeueAfter: r.pendingTimeout - elapsed}, nil
		}

		klog.Infof("attempting to suspend app=%s owner=%s due to pending unschedulable timeout", appName, owner)
		ok, err := r.trySuspendApp(ctx, owner, appName, constants.AppUnschedulable, "pending unschedulable timeout on pod: "+pod.Namespace+"/"+pod.Name)
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
func (r *PodAbnormalSuspendAppController) trySuspendApp(ctx context.Context, owner, appName, reason, message string) (bool, error) {
	name, err := apputils.FmtAppMgrName(appName, owner, "")
	if err != nil {
		klog.Errorf("failed to format app manager name app=%s owner=%s: %v", appName, owner, err)
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
	if am.Status.State == appv1alpha1.Stopped {
		return true, nil
	}

	if !appstate.IsOperationAllowed(am.Status.State, appv1alpha1.StopOp) {
		klog.Infof("operation StopOp not allowed in state=%s for app=%s owner=%s", am.Status.State, appName, owner)
		return false, nil
	}

	am.Spec.OpType = appv1alpha1.StopOp
	if err := r.Update(ctx, &am); err != nil {
		klog.Errorf("failed to update applicationmanager spec to StopOp name=%s app=%s owner=%s: %v", am.Name, appName, owner, err)
		return false, err
	}

	opID := strconv.FormatInt(time.Now().Unix(), 10)
	now := metav1.Now()
	status := appv1alpha1.ApplicationManagerStatus{
		OpType:     appv1alpha1.StopOp,
		OpID:       opID,
		State:      appv1alpha1.Stopping,
		StatusTime: &now,
		UpdateTime: &now,
		Reason:     reason,
		Message:    message,
	}
	if _, err := apputils.UpdateAppMgrStatus(name, status); err != nil {
		return false, err
	}

	utils.PublishAppEvent(utils.EventParams{
		Owner:      am.Spec.AppOwner,
		Name:       am.Spec.AppName,
		OpType:     string(status.OpType),
		OpID:       opID,
		State:      appv1alpha1.Stopping.String(),
		Progress:   message,
		RawAppName: am.Spec.RawAppName,
		Type:       am.Spec.Type.String(),
		Title:      apputils.AppTitle(am.Spec.Config),
		Reason:     reason,
		Message:    message,
	})
	klog.Infof("suspend requested for app=%s owner=%s, reason=%s", am.Spec.AppName, am.Spec.AppOwner, message)
	return true, nil
}

func parsePendingTimeout(v string) time.Duration {
	if v == "" {
		klog.Infof("%s not set, using default 3m", envPendingPodSuspendAppTimeout)
		return 3 * time.Minute
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		klog.Warningf("invalid %s value %q, using default 3m", envPendingPodSuspendAppTimeout, v)
		return 3 * time.Minute
	}
	klog.Infof("%s set to %v", envPendingPodSuspendAppTimeout, d)
	return d
}

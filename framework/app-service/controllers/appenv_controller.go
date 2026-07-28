package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"

	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AppEnvController struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=appenvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=appenvs/status,verbs=get;update;patch
//+kubebuilder:groups=app.bytetrade.io,resources=applicationmanagers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applicationmanagers/status,verbs=get;update;patch

func (r *AppEnvController) SetupWithManager(mgr ctrl.Manager) error {
	// When an app becomes Running again, re-enqueue its AppEnv so any change
	// that was deferred while the app was Stopped (see reconcileAppEnv) gets
	// applied. Without this the pending change would sit forever, since the
	// AppEnv itself does not change on resume.
	appMgrBecameRunning := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldAM, ok1 := e.ObjectOld.(*appv1alpha1.ApplicationManager)
			newAM, ok2 := e.ObjectNew.(*appv1alpha1.ApplicationManager)
			if !ok1 || !ok2 {
				return false
			}
			return oldAM.Status.State != appv1alpha1.Running && newAM.Status.State == appv1alpha1.Running
		},
		CreateFunc:  func(event.CreateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&sysv1alpha1.AppEnv{}).
		Watches(
			&appv1alpha1.ApplicationManager{},
			handler.EnqueueRequestsFromMapFunc(appEnvRequestForAppMgr),
			builder.WithPredicates(appMgrBecameRunning),
		).
		Complete(r)
}

// appEnvRequestForAppMgr maps an ApplicationManager to its backing AppEnv so a
// pending env change can be re-evaluated when the app's state changes.
func appEnvRequestForAppMgr(_ context.Context, obj client.Object) []reconcile.Request {
	am, ok := obj.(*appv1alpha1.ApplicationManager)
	if !ok {
		return nil
	}
	if am.Spec.AppName == "" || am.Spec.AppOwner == "" || am.Spec.AppNamespace == "" {
		return nil
	}
	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Namespace: am.Spec.AppNamespace,
		Name:      apputils.FormatAppEnvName(am.Spec.AppName, am.Spec.AppOwner),
	}}}
}

func (r *AppEnvController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling AppEnv: %s", req.NamespacedName)

	var appEnv sysv1alpha1.AppEnv
	if err := r.Get(ctx, req.NamespacedName, &appEnv); err != nil {
		//todo: more detailed logic
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileAppEnv(ctx, &appEnv)
}

func (r *AppEnvController) reconcileAppEnv(ctx context.Context, appEnv *sysv1alpha1.AppEnv) (ctrl.Result, error) {
	klog.Infof("Processing AppEnv change: %s/%s", appEnv.Namespace, appEnv.Name)

	// Check if this AppEnv was triggered by an environment variable change
	if appEnv.Annotations != nil && appEnv.Annotations[constants.AppEnvSyncAnnotation] != "" {
		klog.Infof("AppEnv %s/%s triggered by environment variable change: %s",
			appEnv.Namespace, appEnv.Name, appEnv.Annotations[constants.AppEnvSyncAnnotation])

		// Clear the annotation immediately - the update will trigger another reconcile
		if err := r.clearSyncAnnotation(ctx, appEnv); err != nil {
			klog.Errorf("Failed to clear sync annotation for AppEnv %s/%s: %v", appEnv.Namespace, appEnv.Name, err)
			return ctrl.Result{}, err
		}

		// Return immediately - the annotation update will trigger another reconcile
		return ctrl.Result{}, nil
	}

	// This reconcile is not triggered by annotation, proceed with normal sync
	if err := r.syncEnvValues(ctx, appEnv); err != nil {
		klog.Errorf("Failed to sync environment values for AppEnv %s/%s: %v", appEnv.Namespace, appEnv.Name, err)
		return ctrl.Result{}, err
	}

	if appEnv.NeedApply {
		appMgr, err := r.getAppMgr(ctx, appEnv)
		if err != nil {
			klog.Errorf("Failed to get app manager for AppEnv %s/%s: %v", appEnv.Namespace, appEnv.Name, err)
			return ctrl.Result{}, err
		}

		// An applyEnv runs a helm upgrade that re-renders the workload. When the
		// app is stopped, the outcome depends on how its replicas are controlled:
		//   - No workloadReplicas (replicas hardcoded in the chart): stop only
		//     patched the live workload to replicas=0, so the upgrade re-renders
		//     the hardcoded count and resurrects the app. Defer the apply; the
		//     change stays pending (NeedApply=true) and is applied by the next
		//     upgrade / when the app runs.
		//   - workloadReplicas: stop scaled the release to zero (replicaCount=0 is
		//     pinned in the release values), so the env upgrade reuses that and
		//     stays at zero. We let it proceed so the new env is baked into the
		//     release and picked up on the next resume; the applyEnv flow records
		//     the pre-op state and lands back in Stopped without waiting for pods.
		if appMgr.Status.State == appv1alpha1.Stopped && !appMgrHasWorkloadReplicas(appMgr) {
			klog.Infof("app %s owner %s is stopped without workloadReplicas, deferring applyEnv (env change kept pending)", appEnv.AppName, appEnv.AppOwner)
			return ctrl.Result{}, nil
		}

		// check for active user batch lease to avoid mid-batch apply
		userNamespace := utils.UserspaceName(appEnv.AppOwner)
		lease := &coordinationv1.Lease{}
		if err := r.Get(ctx, types.NamespacedName{Name: "env-batch-lock", Namespace: userNamespace}, lease); err == nil {
			if isLeaseActive(lease) {
				klog.Infof("User batch lease is active for app: %s owner: %s, requeueing", appEnv.AppName, appEnv.AppOwner)
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}
		}
		if err := r.triggerApplyEnv(ctx, appEnv); err != nil {
			klog.Errorf("Failed to trigger ApplyEnv for AppEnv %s/%s: %v", appEnv.Namespace, appEnv.Name, err)
			return ctrl.Result{}, err
		}
		if err := r.markEnvApplied(ctx, appEnv); err != nil {
			klog.Errorf("Failed to mark AppEnv %s/%s as applied: %v", appEnv.Namespace, appEnv.Name, err)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AppEnvController) syncEnvValues(ctx context.Context, appEnv *sysv1alpha1.AppEnv) error {
	original := appEnv.DeepCopy()

	// Get SystemEnv values
	var systemEnvList sysv1alpha1.SystemEnvList
	if err := r.List(ctx, &systemEnvList); err != nil {
		return fmt.Errorf("failed to list SystemEnvs: %v", err)
	}
	systemEnvMap := make(map[string]*sysv1alpha1.SystemEnv)
	for _, sysEnv := range systemEnvList.Items {
		systemEnvMap[sysEnv.EnvName] = &sysEnv
	}

	// Get UserEnv values from user-space-{appOwner} namespace
	var userEnvList sysv1alpha1.UserEnvList
	userNamespace := utils.UserspaceName(appEnv.AppOwner)
	if err := r.List(ctx, &userEnvList, client.InNamespace(userNamespace)); err != nil {
		return fmt.Errorf("failed to list UserEnvs in namespace %s: %v", userNamespace, err)
	}
	userEnvMap := make(map[string]*sysv1alpha1.UserEnv)
	for _, userEnv := range userEnvList.Items {
		userEnvMap[userEnv.EnvName] = &userEnv
	}

	updated := false
	for i := range appEnv.Envs {
		envVar := &appEnv.Envs[i]
		if envVar.ValueFrom != nil {
			var refValue string
			var refType string
			var refSource string

			// Check if both UserEnv and SystemEnv exist with the same name
			var userEnv *sysv1alpha1.UserEnv
			var sysEnv *sysv1alpha1.SystemEnv
			if userEnv = userEnvMap[envVar.ValueFrom.EnvName]; userEnv != nil {
				refValue = userEnv.GetEffectiveValue()
				refType = userEnv.Type
				refSource = "UserEnv"
			}
			if sysEnv = systemEnvMap[envVar.ValueFrom.EnvName]; sysEnv != nil {
				if userEnv != nil {
					// Both exist - this is unexpected, log a warning
					klog.Warningf("AppEnv %s/%s references environment variable %s which exists in both UserEnv and SystemEnv. UserEnv value will be used.",
						appEnv.Namespace, appEnv.Name, envVar.ValueFrom.EnvName)
				} else {
					refValue = sysEnv.GetEffectiveValue()
					refType = sysEnv.Type
					refSource = "SystemEnv"
				}
			}

			// do not check for non-empty value as an existing refed env may also contain empty value
			if userEnv != nil || sysEnv != nil {
				if envVar.Value != refValue || envVar.Type != refType || envVar.ValueFrom.Status != constants.EnvRefStatusSynced {
					envVar.Value = refValue
					envVar.Type = refType
					envVar.ValueFrom.Status = constants.EnvRefStatusSynced
					updated = true
					if envVar.ApplyOnChange {
						appEnv.NeedApply = true
					}
					klog.V(4).Infof("AppEnv %s/%s environment variable %s synced from %s with value: %s",
						appEnv.Namespace, appEnv.Name, envVar.ValueFrom.EnvName, refSource, refValue)
				}
			} else {
				if envVar.ValueFrom.Status != constants.EnvRefStatusNotFound {
					envVar.ValueFrom.Status = constants.EnvRefStatusNotFound
					updated = true
				}
			}
		}
	}

	if updated {
		if err := r.Patch(ctx, appEnv, client.MergeFrom(original)); err != nil {
			return fmt.Errorf("failed to update AppEnv %s/%s: %v", appEnv.Namespace, appEnv.Name, err)
		}
	}

	return nil
}

// getAppMgr loads the ApplicationManager backing the given AppEnv.
func (r *AppEnvController) getAppMgr(ctx context.Context, appEnv *sysv1alpha1.AppEnv) (*appv1alpha1.ApplicationManager, error) {
	appMgrName, err := apputils.FmtAppMgrName(appEnv.AppName, appEnv.AppOwner, appEnv.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to format app manager name: %v", err)
	}

	var appMgr appv1alpha1.ApplicationManager
	if err := r.Get(ctx, types.NamespacedName{Name: appMgrName}, &appMgr); err != nil {
		return nil, fmt.Errorf("failed to get ApplicationManager %s: %v", appMgrName, err)
	}

	return &appMgr, nil
}

// appMgrHasWorkloadReplicas reports whether the app declares per-workload
// replica counts (and is therefore stopped via a helm scale-to-zero rather than
// a direct replicas=0 patch).
func appMgrHasWorkloadReplicas(appMgr *appv1alpha1.ApplicationManager) bool {
	if appMgr == nil || appMgr.Spec.Config == "" {
		return false
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(appMgr.Spec.Config), &cfg); err != nil {
		klog.Warningf("failed to unmarshal app config for %s when checking workloadReplicas: %v", appMgr.Name, err)
		return false
	}
	return cfg.HasWorkloadReplicas()
}

func (r *AppEnvController) triggerApplyEnv(ctx context.Context, appEnv *sysv1alpha1.AppEnv) error {
	klog.Infof("Triggering ApplyEnv for app: %s owner: %s", appEnv.AppName, appEnv.AppOwner)

	appMgrName, err := apputils.FmtAppMgrName(appEnv.AppName, appEnv.AppOwner, appEnv.Namespace)
	if err != nil {
		return fmt.Errorf("failed to format app manager name: %v", err)
	}

	var targetAppMgr appv1alpha1.ApplicationManager
	if err := r.Get(ctx, types.NamespacedName{Name: appMgrName}, &targetAppMgr); err != nil {
		return fmt.Errorf("failed to get ApplicationManager %s: %v", appMgrName, err)
	}

	state := targetAppMgr.Status.State
	if !appstate.IsOperationAllowed(state, appv1alpha1.ApplyEnvOp) {
		// trigger backoff retry and this is the expected behaviour
		return fmt.Errorf("app %s is currently in state %s, applyEnv not allowed", appEnv.AppName, state)
	}

	appMgrCopy := targetAppMgr.DeepCopy()
	appMgrCopy.Spec.OpType = appv1alpha1.ApplyEnvOp
	// Record the state right before applyEnv so the applyEnv flow can detect an
	// app that was Stopped (scaled to zero) and land it back in Stopped without
	// waiting for pods, instead of forcing it to start up.
	if appMgrCopy.Annotations == nil {
		appMgrCopy.Annotations = make(map[string]string)
	}
	appMgrCopy.Annotations[api.AppPreUpgradeStateKey] = string(state)

	if err := r.Patch(ctx, appMgrCopy, client.MergeFrom(&targetAppMgr)); err != nil {
		return fmt.Errorf("failed to update ApplicationManager Spec.OpType: %v", err)
	}

	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)

	status := appv1alpha1.ApplicationManagerStatus{
		OpType:     appv1alpha1.ApplyEnvOp,
		State:      appv1alpha1.ApplyingEnv,
		OpID:       opID,
		Message:    "waiting for applying env",
		StatusTime: &now,
		UpdateTime: &now,
	}

	_, err = apputils.UpdateAppMgrStatus(targetAppMgr.Name, status)
	if err != nil {
		return fmt.Errorf("failed to update ApplicationManager Status: %v", err)
	}

	klog.Infof("Successfully triggered ApplyEnv for app: %s owner: %s", appEnv.AppName, appEnv.AppOwner)
	return nil
}

func (r *AppEnvController) clearSyncAnnotation(ctx context.Context, appEnv *sysv1alpha1.AppEnv) error {
	if appEnv.Annotations == nil || appEnv.Annotations[constants.AppEnvSyncAnnotation] == "" {
		return nil
	}

	original := appEnv.DeepCopy()
	delete(appEnv.Annotations, constants.AppEnvSyncAnnotation)

	klog.Infof("Clearing environment sync annotation from AppEnv %s/%s", appEnv.Namespace, appEnv.Name)
	return r.Patch(ctx, appEnv, client.MergeFrom(original))
}

func (r *AppEnvController) markEnvApplied(ctx context.Context, appEnv *sysv1alpha1.AppEnv) error {
	if !appEnv.NeedApply {
		return nil
	}
	original := appEnv.DeepCopy()
	appEnv.NeedApply = false
	return r.Patch(ctx, appEnv, client.MergeFrom(original))
}

// isLeaseActive returns true if now < RenewTime + LeaseDurationSeconds
func isLeaseActive(l *coordinationv1.Lease) bool {
	if l == nil || l.Spec.RenewTime == nil || l.Spec.LeaseDurationSeconds == nil {
		return false
	}
	exp := l.Spec.RenewTime.Add(time.Duration(*l.Spec.LeaseDurationSeconds) * time.Second)
	return time.Now().Before(exp)
}

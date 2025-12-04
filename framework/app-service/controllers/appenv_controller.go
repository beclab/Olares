package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/appstate"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&sysv1alpha1.AppEnv{}).
		Complete(r)
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

	am, err := apputils.UpdateAppMgrStatus(targetAppMgr.Name, status)
	if err != nil {
		return fmt.Errorf("failed to update ApplicationManager Status: %v", err)
	}
	utils.PublishAppEvent(utils.EventParams{
		Owner:      am.Spec.AppOwner,
		Name:       am.Spec.AppName,
		OpType:     string(am.Status.OpType),
		OpID:       opID,
		State:      appv1alpha1.ApplyingEnv.String(),
		RawAppName: am.Spec.RawAppName,
		Type:       "app",
		Title:      apputils.AppTitle(am.Spec.Config),
	})

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

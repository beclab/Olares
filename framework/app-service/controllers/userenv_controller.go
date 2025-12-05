package controllers

import (
	"context"
	"fmt"

	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/security"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UserEnvController struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=userenvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=userenvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=appenvs,verbs=get;list;watch;update;patch

func (r *UserEnvController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var userEnv sysv1alpha1.UserEnv
	if err := r.Get(ctx, req.NamespacedName, &userEnv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileUserEnv(ctx, &userEnv)
}

func (r *UserEnvController) reconcileUserEnv(ctx context.Context, userEnv *sysv1alpha1.UserEnv) (ctrl.Result, error) {
	// Extract username from UserEnv namespace
	isUserNs, username := security.IsUserInternalNamespaces(userEnv.Namespace)
	if !isUserNs {
		klog.Warningf("UserEnv %s/%s is not in a user namespace, skipping reconciliation", userEnv.Namespace, userEnv.Name)
		return ctrl.Result{}, nil
	}

	klog.Infof("Processing UserEnv change: %s of user: %s", userEnv.EnvName, username)

	var appEnvList sysv1alpha1.AppEnvList
	if err := r.List(ctx, &appEnvList); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list AppEnvs: %v", err)
	}

	refCount := 0
	annotatedCount := 0
	failedCount := 0

	for i := range appEnvList.Items {
		appEnv := &appEnvList.Items[i]
		if appEnv.AppOwner != username {
			continue
		}
		if r.isReferenced(appEnv, userEnv.EnvName) {
			refCount++

			annotated, err := r.annotateAppEnvForSync(ctx, appEnv, userEnv)
			if annotated {
				annotatedCount++
			}
			if err != nil {
				klog.Errorf("Failed to annotate AppEnv %s/%s for sync: %v", appEnv.Namespace, appEnv.Name, err)
				failedCount++
				continue
			}
		}
	}

	if refCount > 0 {
		klog.Infof("UserEnv %s reconciliation completed: %d total references, %d annotated for sync, %d failed",
			userEnv.EnvName, refCount, annotatedCount, failedCount)
	}

	if failedCount > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to annotate %d AppEnvs referencing environment variable %s", failedCount, userEnv.EnvName)
	}

	return ctrl.Result{}, nil
}

func (r *UserEnvController) isReferenced(appEnv *sysv1alpha1.AppEnv, userEnvName string) bool {
	for _, envVar := range appEnv.Envs {
		if envVar.ValueFrom != nil && envVar.ValueFrom.EnvName == userEnvName {
			return true
		}
	}
	return false
}

func (r *UserEnvController) annotateAppEnvForSync(ctx context.Context, appEnv *sysv1alpha1.AppEnv, userEnv *sysv1alpha1.UserEnv) (bool, error) {
	// Check if annotation already exists
	if appEnv.Annotations != nil && appEnv.Annotations[constants.AppEnvSyncAnnotation] != "" {
		klog.V(4).Infof("AppEnv %s/%s already has sync annotation, skipping", appEnv.Namespace, appEnv.Name)
		return false, nil
	}

	// Add annotation to trigger AppEnvController sync
	original := appEnv.DeepCopy()
	if appEnv.Annotations == nil {
		appEnv.Annotations = make(map[string]string)
	}
	appEnv.Annotations[constants.AppEnvSyncAnnotation] = userEnv.EnvName

	klog.Infof("Annotating AppEnv %s/%s for sync due to environment variable %s change",
		appEnv.Namespace, appEnv.Name, userEnv.EnvName)

	return true, r.Patch(ctx, appEnv, client.MergeFrom(original))
}

func (r *UserEnvController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sysv1alpha1.UserEnv{}).
		Complete(r)
}

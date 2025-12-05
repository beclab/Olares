package controllers

import (
	"context"
	"fmt"

	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemEnvController struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=systemenvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=systemenvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=appenvs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=appenvs/status,verbs=get;update;patch

func (r *SystemEnvController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sysv1alpha1.SystemEnv{}).
		Complete(r)
}

func (r *SystemEnvController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling SystemEnv: %s", req.NamespacedName)

	var systemEnv sysv1alpha1.SystemEnv
	if err := r.Get(ctx, req.NamespacedName, &systemEnv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileSystemEnv(ctx, &systemEnv)
}

func (r *SystemEnvController) reconcileSystemEnv(ctx context.Context, systemEnv *sysv1alpha1.SystemEnv) (ctrl.Result, error) {
	klog.Infof("Processing SystemEnv change: %s", systemEnv.EnvName)

	var appEnvList sysv1alpha1.AppEnvList
	if err := r.List(ctx, &appEnvList); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list AppEnvs: %v", err)
	}

	refCount := 0
	annotatedCount := 0
	failedCount := 0

	for i := range appEnvList.Items {
		appEnv := &appEnvList.Items[i]
		if r.isReferenced(appEnv, systemEnv.EnvName) {
			refCount++

			annotated, err := r.annotateAppEnvForSync(ctx, appEnv, systemEnv)
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
		klog.Infof("SystemEnv %s reconciliation completed: %d total references, %d annotated for sync, %d failed",
			systemEnv.EnvName, refCount, annotatedCount, failedCount)
	}

	if failedCount > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to annotate %d AppEnvs referencing environment variable %s", failedCount, systemEnv.EnvName)
	}

	return ctrl.Result{}, nil
}

func (r *SystemEnvController) isReferenced(appEnv *sysv1alpha1.AppEnv, systemEnvName string) bool {
	for _, envVar := range appEnv.Envs {
		if envVar.ValueFrom != nil && envVar.ValueFrom.EnvName == systemEnvName {
			return true
		}
	}
	return false
}

func (r *SystemEnvController) annotateAppEnvForSync(ctx context.Context, appEnv *sysv1alpha1.AppEnv, systemEnv *sysv1alpha1.SystemEnv) (bool, error) {
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
	appEnv.Annotations[constants.AppEnvSyncAnnotation] = systemEnv.EnvName

	klog.Infof("Annotating AppEnv %s/%s for sync due to environment variable %s change",
		appEnv.Namespace, appEnv.Name, systemEnv.EnvName)

	return true, r.Patch(ctx, appEnv, client.MergeFrom(original))
}

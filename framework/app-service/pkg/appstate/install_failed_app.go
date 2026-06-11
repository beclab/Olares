package appstate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &InstallFailedApp{}

type InstallFailedApp struct {
	*baseOperationApp
}

func NewInstallFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallFailedApp{
				&baseOperationApp{
					ttl: ttl,
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
			}
		})
}

func (p *InstallFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if !apputils.IsProtectedNamespace(p.manager.Spec.AppNamespace) {
		var pvcs corev1.PersistentVolumeClaimList
		err := p.client.List(ctx, &pvcs, client.InNamespace(p.manager.Spec.AppNamespace))
		if err != nil {
			klog.Errorf("failed to list pvcs %v", err)
			return nil, err
		}
		for _, pvc := range pvcs.Items {
			var curPvc corev1.PersistentVolumeClaim
			err = p.client.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, &curPvc)
			if err != nil && !apierrors.IsNotFound(err) {
				return nil, err
			}
			err = p.client.Delete(ctx, &curPvc)
			if err != nil && !apierrors.IsNotFound(err) {
				return nil, err
			}
		}
		var ns corev1.Namespace
		err = p.client.Get(ctx, types.NamespacedName{Name: p.manager.Spec.AppNamespace}, &ns)
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, err
		}
		if err == nil {
			e := p.client.Delete(ctx, &ns)
			if e != nil {
				klog.Errorf("failed to delete ns %s, err=%v", p.manager.Spec.AppNamespace, e)
				return nil, e
			}
		}
	}

	// A failed install tears down the namespace above, so any allocation that
	// AllocateForInstall reserved for this app is now backing a workload that no
	// longer exists. Release it (guarded, so it is a no-op once cleaned up) to
	// avoid leaking the GPU/compute reservation.
	if err := p.cleanupComputeAllocation(ctx); err != nil {
		klog.Errorf("cleanup compute allocation for install-failed app %s failed %v", p.manager.Spec.AppName, err)
		return nil, err
	}

	return nil, nil
}

func (p *InstallFailedApp) cleanupComputeAllocation(ctx context.Context) error {
	if p.manager.Spec.Config == "" {
		return nil
	}
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup of install-failed app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	// A failed install never owns a shared server, so only its own allocation
	// row (if any) needs releasing; never touch a shared server here.
	cleaned, err := compute.EnsureAllocationsDeletedForComputeTarget(ctx, p.client, &appCfg, false)
	if err != nil {
		return err
	}
	if cleaned {
		klog.Infof("released leaked compute allocation for install-failed app %s", p.manager.Spec.AppName)
	}
	return nil
}

func (p *InstallFailedApp) Cancel(ctx context.Context) error {
	return nil
}

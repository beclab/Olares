package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

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

	return nil, nil
}

func (p *InstallFailedApp) Cancel(ctx context.Context) error {
	return nil
}

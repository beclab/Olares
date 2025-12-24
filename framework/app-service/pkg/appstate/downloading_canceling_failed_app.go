package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FIXME: impossible state

var _ OperationApp = &DownloadingCancelFailedApp{}

type DownloadingCancelFailedApp struct {
	*baseOperationApp
	imageClient images.ImageManager
}

func NewDownloadingCancelFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &DownloadingCancelFailedApp{
				baseOperationApp: &baseOperationApp{
					ttl: ttl,
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
				imageClient: images.NewImageManager(c),
			}

		})
}

func (p *DownloadingCancelFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if !apputils.IsProtectedNamespace(p.manager.Spec.AppNamespace) {
		var ns corev1.Namespace
		err := p.client.Get(ctx, types.NamespacedName{Name: p.manager.Spec.AppNamespace}, &ns)
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
	var im appsv1.ImageManager
	err := p.client.Get(ctx, types.NamespacedName{Name: p.manager.Name}, &im)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if im.Status.State != appsv1.DownloadingCanceled.String() {
		err = p.imageClient.UpdateStatus(ctx, p.manager.Name, appsv1.DownloadingCanceled.String(), appsv1.DownloadingCanceled.String())
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (p *DownloadingCancelFailedApp) Cancel(ctx context.Context) error {
	return nil
}

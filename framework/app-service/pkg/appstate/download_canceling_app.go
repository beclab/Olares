package appstate

import (
	"context"
	"fmt"
	"time"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/images"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ CancelOperationApp = &DownloadingCancelingApp{}

type DownloadingCancelingApp struct {
	*baseOperationApp
	imageClient images.ImageManager
}

func (p *DownloadingCancelingApp) IsAppCreated() bool {
	return false
}

func NewDownloadingCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &DownloadingCancelingApp{
				baseOperationApp: &baseOperationApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
					ttl: ttl,
				},
				imageClient: images.NewImageManager(c),
			}
		})
}

func (p *DownloadingCancelingApp) exec(ctx context.Context) error {
	err := p.imageClient.UpdateStatus(ctx, p.manager.Name, appsv1.DownloadingCanceled.String(), appsv1.DownloadingCanceled.String())
	if err != nil {
		klog.Errorf("update im name=%s to downloadingCanceled state failed %v", p.manager.Name, err)
		return err
	}
	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}

	if !apputils.IsProtectedNamespace(p.manager.Spec.AppNamespace) {
		var ns corev1.Namespace
		err = p.client.Get(ctx, types.NamespacedName{Name: p.manager.Spec.AppNamespace}, &ns)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("failed to get namespace %s, %v", p.manager.Spec.AppNamespace, err)
			return err
		}
		if err == nil {
			if delErr := p.client.Delete(ctx, &ns); delErr != nil && !apierrors.IsNotFound(delErr) {
				klog.Errorf("failed to delete namespace %s, %v", p.manager.Spec.AppNamespace, delErr)
				return delErr
			}
		}
	}
	return nil
}

func (p *DownloadingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	err := p.exec(ctx)
	if err != nil {
		updateErr := p.updateStatus(ctx, p.manager, appsv1.DownloadingCancelFailed, nil, err.Error(), "")
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.DownloadingCancelFailed.String(), updateErr)
			return nil, updateErr
		}
	}

	return &downloadingCancelInProgressApp{
		DownloadingCancelingApp:           p,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}, nil
}

func (p *DownloadingCancelingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.DownloadingCancelFailed, nil, appsv1.DownloadingCancelFailed.String(), "")
	if err != nil {
		klog.Errorf("update state to %s failed %v", appsv1.DownloadingCancelFailed.String(), err)
		return err
	}
	return nil
}

var _ PollableStatefulInProgressApp = &downloadingCancelInProgressApp{}

type downloadingCancelInProgressApp struct {
	*DownloadingCancelingApp
	*basePollableStatefulInProgressApp
}

func (p *downloadingCancelInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

func (p *downloadingCancelInProgressApp) poll(ctx context.Context) error {
	if apputils.IsProtectedNamespace(p.manager.Spec.AppNamespace) {
		return nil
	}

	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			var ns corev1.Namespace
			err := p.client.Get(ctx, types.NamespacedName{Name: p.manager.Spec.AppNamespace}, &ns)
			klog.Infof("downloading cancel poll namespace %s err %v", p.manager.Spec.AppNamespace, err)
			if apierrors.IsNotFound(err) {
				return nil
			}

		case <-ctx.Done():
			return fmt.Errorf("app %s execute cancel operation failed %w", p.manager.Spec.AppName, ctx.Err())
		}
	}
}

func (p *downloadingCancelInProgressApp) WaitAsync(ctx context.Context) {
	appFactory.waitForPolling(ctx, p, func(err error) {
		if err != nil {
			updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.DownloadingCancelFailed, nil, appsv1.DownloadingCancelFailed.String(), "")
			if updateErr != nil {
				klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.DownloadingCancelFailed.String(), updateErr)
				return
			}

			return
		}

		updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.DownloadingCanceled, nil, appsv1.DownloadingCanceled.String(), "")
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InstallingCanceled.String(), updateErr)
			return
		}

	})
}

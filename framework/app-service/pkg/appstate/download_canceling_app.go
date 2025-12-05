package appstate

import (
	"context"
	"time"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/images"
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

func (p *DownloadingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	err := p.imageClient.UpdateStatus(ctx, p.manager.Name, appsv1.DownloadingCanceled.String(), appsv1.DownloadingCanceled.String())
	if err != nil {
		klog.Errorf("update im name=%s to downloadingCanceled state failed %v", p.manager.Name, err)
		return nil, err
	}
	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}
	message := constants.OperationCanceledByUserTpl
	if p.manager.Status.Message == constants.OperationCanceledByTerminusTpl {
		message = constants.OperationCanceledByTerminusTpl
	}
	opRecord := makeRecord(p.manager, appsv1.DownloadingCanceled, message)

	// FIXME: should check if the image downloading is canceled successfully
	updateErr := p.updateStatus(ctx, p.manager, appsv1.DownloadingCanceled, opRecord, message, "")
	if updateErr != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.DownloadingCanceled.String(), updateErr)
		return nil, updateErr
	}
	return nil, nil
}

func (p *DownloadingCancelingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.DownloadingCancelFailed, nil, appsv1.DownloadingCancelFailed.String(), "")
	if err != nil {
		klog.Errorf("update state to %s failed %v", appsv1.DownloadingCancelFailed.String(), err)
		return err
	}
	return nil
}

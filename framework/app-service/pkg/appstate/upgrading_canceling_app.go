package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/images"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ CancelOperationApp = &UpgradingCancelingApp{}

type UpgradingCancelingApp struct {
	*baseOperationApp
	imageClient images.ImageManager
}

func (p *UpgradingCancelingApp) IsAppCreated() bool {
	return true
}

func NewUpgradingCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &UpgradingCancelingApp{
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

func (p *UpgradingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	var err error
	klog.Infof("execute upgrading cancel operation appName=%s", p.manager.Spec.AppName)
	err = p.imageClient.UpdateStatus(ctx, p.manager.Name, appsv1.DownloadingCanceled.String(), appsv1.DownloadingCanceled.String())
	if err != nil {
		klog.Errorf("update im name=%s to downloadingCanceled state failed %v", p.manager.Name, err)
		return nil, err
	}

	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}

	err = p.updateStatus(ctx, p.manager, appsv1.Stopping, nil, appsv1.Stopping.String(), "")
	if err != nil {
		klog.Errorf("update appmgr state to suspending state failed %v", err)
		return nil, err
	}
	return nil, nil
}

func (p *UpgradingCancelingApp) Cancel(ctx context.Context) error {
	return nil
}

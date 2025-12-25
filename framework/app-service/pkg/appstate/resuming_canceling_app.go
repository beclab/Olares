package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"k8s.io/klog/v2"
)

var _ CancelOperationApp = &ResumingCancelingApp{}

type ResumingCancelingApp struct {
	*baseOperationApp
}

func (p *ResumingCancelingApp) IsAppCreated() bool {
	return true
}

func NewResumingCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &ResumingCancelingApp{
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

func (p *ResumingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}

	err := p.updateStatus(ctx, p.manager, appsv1.Stopping, nil, appsv1.Stopping.String(), "")
	if err != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Stopping.String(), err)
		return nil, err
	}

	return nil, nil
}

func (p *ResumingCancelingApp) Cancel(ctx context.Context) error {
	return p.updateStatus(ctx, p.manager, appsv1.ResumingCancelFailed, nil, appsv1.ResumingCancelFailed.String(), "")
}

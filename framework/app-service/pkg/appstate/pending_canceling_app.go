package appstate

import (
	"context"
	"time"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ CancelOperationApp = &PendingCancelingApp{}

type PendingCancelingApp struct {
	*baseOperationApp
}

func (p *PendingCancelingApp) IsAppCreated() bool {
	return false
}

func NewPendingCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &PendingCancelingApp{
				&baseOperationApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
					ttl: ttl,
				},
			}
		})
}

func (p *PendingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}

	err := p.updateStatus(ctx, p.manager, appsv1.PendingCanceled, nil, appsv1.PendingCanceled.String(), "")
	if err != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.PendingCanceled, err)
		return nil, err
	}

	return nil, nil
}

func (p *PendingCancelingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.PendingCancelFailed, nil, appsv1.PendingCancelFailed.String(), "")
	if err != nil {
		klog.Errorf("update manager %s to state %s failed %v", p.manager.Name, appsv1.PendingCancelFailed, err)
		return err
	}

	return nil
}

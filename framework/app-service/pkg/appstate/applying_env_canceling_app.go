package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ CancelOperationApp = &ApplyingEnvCancelingApp{}

type ApplyingEnvCancelingApp struct {
	*baseOperationApp
}

func (p *ApplyingEnvCancelingApp) IsAppCreated() bool {
	return true
}

func NewApplyingEnvCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &ApplyingEnvCancelingApp{
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

func (p *ApplyingEnvCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	var err error
	klog.Infof("execute applying env cancel operation appName=%s", p.manager.Spec.AppName)

	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not cancelable", p.manager.Name)
	}
	err = p.updateStatus(ctx, p.manager, appsv1.Stopping, nil, appsv1.Stopping.String(), "")
	if err != nil {
		klog.Errorf("update appmgr state to running state failed %v", err)
		err = p.updateStatus(ctx, p.manager, appsv1.ApplyingEnvCancelFailed, nil, "Failed to update status after canceling", "")
		if err != nil {
			klog.Errorf("update appmgr state to suspending failed %v", err)
			return nil, err
		}
		return nil, nil
	}
	return nil, nil
}

func (p *ApplyingEnvCancelingApp) Cancel(ctx context.Context) error {
	return nil
}

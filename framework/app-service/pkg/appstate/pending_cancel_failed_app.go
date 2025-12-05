package appstate

import (
	"time"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FIXME: impossible state

var _ StatefulApp = &PendingCancelFailedApp{}

type PendingCancelFailedApp struct {
	*DoNothingApp
}

func NewPendingCancelFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &PendingCancelFailedApp{
				DoNothingApp: &DoNothingApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
			}
		})
}

//func (p *PendingCancelFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
//	// FIXME: should set a max retry count for cancel operation
//	err := p.updateStatus(ctx, p.manager, appsv1.PendingCanceling, nil, appsv1.PendingCanceling.String())
//	if err != nil {
//		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.PendingCanceling, err)
//	}
//	return nil, err
//}

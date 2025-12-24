package appstate

import (
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulApp = &UpgradingCancelFailedApp{}

type UpgradingCancelFailedApp struct {
	*DoNothingApp
}

func NewUpgradingCancelFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &UpgradingCancelFailedApp{
				DoNothingApp: &DoNothingApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
			}
		})
}

//func (p *UpgradingCancelFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
//	err := p.updateStatus(ctx, p.manager, appsv1.UpgradingCanceling, nil, appsv1.UpgradingCanceling.String())
//	if err != nil {
//		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.UpgradingCanceling, err)
//	}
//	return nil, err
//}

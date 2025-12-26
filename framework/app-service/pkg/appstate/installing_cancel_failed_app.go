package appstate

import (
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &InstallingCancelFailedApp{}

type InstallingCancelFailedApp struct {
	UninstallFailedApp
}

func NewInstallingCancelFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallingCancelFailedApp{
				UninstallFailedApp: UninstallFailedApp{
					&baseOperationApp{
						ttl: ttl,
						baseStatefulApp: &baseStatefulApp{
							manager: manager,
							client:  c,
						},
					},
				},
			}
		})
}

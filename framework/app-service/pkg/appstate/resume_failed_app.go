package appstate

import (
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &ResumeFailedApp{}

type ResumeFailedApp struct {
	SuspendFailedApp
}

func NewResumeFailedApp(deps Deps,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	return deps.Factory.New(deps, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &ResumeFailedApp{
				SuspendFailedApp: SuspendFailedApp{
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

package appstate

import (
	"context"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulApp = &ApplyingEnvCancelFailedApp{}

type ApplyingEnvCancelFailedApp struct {
	*baseStatefulApp
}

func NewApplyingEnvCancelFailedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return &ApplyingEnvCancelFailedApp{
		baseStatefulApp: &baseStatefulApp{
			manager: manager,
			client:  c,
		},
	}, nil
}

func (p *ApplyingEnvCancelFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

func (p *ApplyingEnvCancelFailedApp) IsTimeout() bool {
	return false
}

func (p *ApplyingEnvCancelFailedApp) Cancel(ctx context.Context) error {
	return nil
}

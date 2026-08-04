package appstate

import (
	"context"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

var _ StatefulApp = &ApplyingEnvCancelFailedApp{}

type ApplyingEnvCancelFailedApp struct {
	*baseStatefulApp
}

func NewApplyingEnvCancelFailedApp(deps Deps,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return &ApplyingEnvCancelFailedApp{
		baseStatefulApp: &baseStatefulApp{
			manager: manager,
			client:  deps.Client,
			deps:    deps,
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

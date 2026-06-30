package appstate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &StoppedApp{}

// StoppedApp is the steady-state handler for the Stopped state.
//
// The compute allocation is normally released during the Stopping -> Stopped
// transition (SuspendingApp) or by SuspendFailedApp, but a handful of paths can
// land an app in Stopped without that cleanup having run: apps that were stopped
// before compute accounting existed, or any out-of-band write that sets the
// state directly. StoppedApp re-runs a guarded cleanup so a stopped app never
// keeps a leaked GPU/compute reservation while its workload is scaled to zero.
type StoppedApp struct {
	*baseOperationApp
}

func NewStoppedApp(c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &StoppedApp{
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

func (p *StoppedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if err := p.cleanupComputeAllocation(ctx); err != nil {
		klog.Errorf("cleanup compute allocation for stopped app %s failed %v", p.manager.Spec.AppName, err)
		return nil, err
	}
	return nil, nil
}

func (p *StoppedApp) cleanupComputeAllocation(ctx context.Context) error {
	if p.manager.Spec.Config == "" {
		return nil
	}
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup of stopped app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	// Mirror SuspendingApp / SuspendFailedApp: stop-all also frees the shared
	// server's allocation, a client-only stop must not.
	stopServer := p.manager.Annotations[api.AppStopAllKey] == "true"
	cleaned, err := compute.EnsureAllocationsDeletedForComputeTarget(ctx, p.client, &appCfg, stopServer)
	if err != nil {
		return err
	}
	if cleaned {
		klog.Infof("released leaked compute allocation for stopped app %s", p.manager.Spec.AppName)
	}
	return nil
}

func (p *StoppedApp) Cancel(ctx context.Context) error {
	return nil
}

package appstate

import (
	"context"
	"encoding/json"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &UninstalledApp{}

type UninstalledApp struct {
	*baseOperationApp
}

func NewUninstalledApp(ctx context.Context, client client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	var err error
	var app appsv1.Application
	err = client.Get(ctx, types.NamespacedName{Name: manager.Name}, &app)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get application %s failed %v", manager.Name, err)
		return nil, NewStateError(err.Error())
	}

	r := &UninstalledApp{
		baseOperationApp: &baseOperationApp{
			ttl: 0,
			baseStatefulApp: &baseStatefulApp{
				manager: manager,
				client:  client,
			},
		},
	}

	if err == nil {
		// app is not expected to exist
		return nil, NewErrorUnknownState(func() func(ctx context.Context) error {
			return func(ctx context.Context) error {
				// Force delete the app if it does not exist.
				// forceDeleteApp also releases the compute allocation.
				err = r.forceDeleteApp(ctx)
				if err != nil {
					klog.Errorf("delete app %s failed %v", manager.Spec.AppName, err)
					return err
				}

				return nil
			}
		}, nil)
	}

	return r, nil
}

// Exec runs a guarded compute cleanup for the terminal Uninstalled state.
//
// The Uninstalling -> Uninstalled transition (and forceDeleteApp) already
// release the allocation, so in the normal flow this is a cheap no-op (a single
// ConfigMap read). It exists to catch apps that reached Uninstalled without that
// cleanup running, e.g. apps uninstalled before compute accounting existed.
func (p *UninstalledApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if err := p.cleanupComputeAllocation(ctx); err != nil {
		klog.Errorf("cleanup compute allocation for uninstalled app %s failed %v", p.manager.Spec.AppName, err)
		return nil, err
	}
	return nil, nil
}

func (p *UninstalledApp) cleanupComputeAllocation(ctx context.Context) error {
	if p.manager.Spec.Config == "" {
		return nil
	}
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup of uninstalled app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	uninstallAll := p.manager.Annotations[api.AppUninstallAllKey] == "true"
	cleaned, err := compute.EnsureAllocationsDeletedForComputeTarget(ctx, p.client, &appCfg, uninstallAll)
	if err != nil {
		return err
	}
	if cleaned {
		klog.Infof("released leaked compute allocation for uninstalled app %s", p.manager.Spec.AppName)
	}
	return nil
}

func (p *UninstalledApp) Cancel(ctx context.Context) error {
	return nil
}

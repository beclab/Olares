package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &SuspendFailedApp{}

type SuspendFailedApp struct {
	*baseOperationApp
}

func NewSuspendFailedApp(deps Deps,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	return deps.Factory.New(deps, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &SuspendFailedApp{
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

func (p *SuspendFailedApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	err := p.StateReconcile(ctx)
	if err != nil {
		klog.Errorf("stop-failed-app %s state reconcile failed %v", p.manager.Spec.AppName, err)
	}
	return nil, err
}

func (p *SuspendFailedApp) StateReconcile(ctx context.Context) error {
	stopServer := p.manager.Annotations[api.AppStopAllKey] == "true"
	if stopServer {
		err := suspendV2AppAll(ctx, p.client, p.manager)
		if err != nil {
			klog.Errorf("suspend v2 app %s %s failed %v", p.manager.Spec.Type, p.manager.Spec.AppName, err)
			return fmt.Errorf("suspend v2 app %s failed %w", p.manager.Spec.AppName, err)
		}
	} else {
		err := suspendV1AppOrV2Client(ctx, p.client, p.manager)
		if err != nil {
			klog.Errorf("suspend app %s %s failed %v", p.manager.Spec.Type, p.manager.Spec.AppName, err)
			return fmt.Errorf("suspend app %s failed %w", p.manager.Spec.AppName, err)
		}
	}

	if p.manager.Spec.Type == "middleware" && userspace.IsKbMiddlewares(p.manager.Spec.AppName) {
		op := p.deps.NewMiddlewareOp(ctx, kbopv1alpha1.StopType, p.manager, p.client)
		err := op.Stop()
		if err != nil {
			klog.Errorf("stop-failed-middleware %s state reconcile failed %v", p.manager.Spec.AppName, err)
			return err
		}
	}

	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup of app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	if err := compute.DeleteAllocationsForComputeTarget(ctx, p.client, &appCfg, stopServer); err != nil {
		klog.Errorf("delete compute allocation for suspend-failed app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

func (p *SuspendFailedApp) Cancel(ctx context.Context) error {
	return nil
}

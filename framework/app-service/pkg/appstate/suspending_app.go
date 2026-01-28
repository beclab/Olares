package appstate

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubeblocks"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &SuspendingApp{}

type SuspendingApp struct {
	*baseOperationApp
}

func NewSuspendingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &SuspendingApp{
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

func (p *SuspendingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	err := p.exec(ctx)
	if err != nil {
		klog.Errorf("suspend app %s failed %v", p.manager.Spec.AppName, err)
		opRecord := makeRecord(p.manager, appsv1.StopFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, err.Error()))
		updateErr := p.updateStatus(ctx, p.manager, appsv1.StopFailed, opRecord, err.Error(), "")
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.StopFailed, err)
			return nil, updateErr
		}

		return nil, nil
	}

	opRecord := makeRecord(p.manager, appsv1.Stopped, fmt.Sprintf(constants.StopOperationCompletedTpl, p.manager.Spec.AppName))
	// Read latest status directly from apiserver to avoid cache staleness
	reason := p.manager.Status.Reason
	if cli, err := utils.GetClient(); err == nil {
		if am, err := cli.AppV1alpha1().ApplicationManagers().Get(ctx, p.manager.Name, metav1.GetOptions{}); err == nil && am != nil {
			if am.Status.Reason != "" {
				reason = am.Status.Reason
			}
		}
	}
	updateErr := p.updateStatus(ctx, p.manager, appsv1.Stopped, opRecord, fmt.Sprintf(constants.StopOperationCompletedTpl, p.manager.Spec.AppName), reason)
	if updateErr != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Stopped.String(), err)
		return nil, updateErr
	}

	return nil, nil
}

func (p *SuspendingApp) exec(ctx context.Context) error {
	// Check if stop-all is requested for V2 apps to also stop server-side shared charts
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

	if stopServer {
		// For V2 cluster-scoped apps, when server is down, stop all other users' clients
		// because they share the same server and cannot function without it
		klog.Infof("stopping other users' clients for v2 app %s", p.manager.Spec.AppName)

		var appManagerList appsv1.ApplicationManagerList
		if err := p.client.List(ctx, &appManagerList); err != nil {
			klog.Errorf("failed to list application managers: %v", err)
		} else {
			// find all ApplicationManagers with same AppName but different AppOwner
			for _, am := range appManagerList.Items {
				// Skip if same owner (already handled) or different app
				if am.Spec.AppName != p.manager.Spec.AppName || am.Spec.AppOwner == p.manager.Spec.AppOwner {
					continue
				}

				if am.Spec.Type != appsv1.App && am.Spec.Type != appsv1.Middleware {
					continue
				}

				if am.Status.State == appsv1.Stopped || am.Status.State == appsv1.Stopping {
					klog.Infof("app %s owner %s already in stopped/stopping state, skip", am.Spec.AppName, am.Spec.AppOwner)
					continue
				}

				if !IsOperationAllowed(am.Status.State, appsv1.StopOp) {
					klog.Infof("app %s owner %s not allowed do stop operation, skip", am.Spec.AppName, am.Spec.AppOwner)
					continue
				}
				opID := strconv.FormatInt(time.Now().Unix(), 10)
				now := metav1.Now()
				status := appsv1.ApplicationManagerStatus{
					OpType:     appsv1.StopOp,
					OpID:       opID,
					State:      appsv1.Stopping,
					StatusTime: &now,
					UpdateTime: &now,
					Reason:     p.manager.Status.Reason,
					Message:    p.manager.Status.Message,
				}
				if _, err := apputils.UpdateAppMgrStatus(am.Name, status); err != nil {
					return err
				}

				klog.Infof("stopping client for user %s, app %s", am.Spec.AppOwner, am.Spec.AppName)

			}
		}
	}

	if p.manager.Spec.Type == appsv1.Middleware && userspace.IsKbMiddlewares(p.manager.Spec.AppName) {
		err := p.execMiddleware(ctx)
		if err != nil {
			klog.Errorf("suspend middleware %s failed %v", p.manager.Spec.AppName, err)
			return err
		}
	}
	return nil
}

func (p *SuspendingApp) Cancel(ctx context.Context) error {
	// FIXME: cancel suspend operation if timeout
	return nil
}

func (p *SuspendingApp) execMiddleware(ctx context.Context) error {
	op := kubeblocks.NewOperation(ctx, kbopv1alpha1.StopType, p.manager, p.client)
	err := op.Stop()
	if err != nil {
		klog.Errorf("failed to stop middleware %s,err=%v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

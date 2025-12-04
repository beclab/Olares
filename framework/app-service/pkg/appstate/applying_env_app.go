package appstate

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	appsv1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/appcfg"
	"bytetrade.io/web3os/app-service/pkg/appinstaller"
	"bytetrade.io/web3os/app-service/pkg/appinstaller/versioned"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &ApplyingEnvApp{}

type ApplyingEnvApp struct {
	*baseOperationApp
}

func NewApplyingEnvApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &ApplyingEnvApp{
				baseOperationApp: &baseOperationApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
					ttl: ttl,
				},
			}
		})
}

func (a *ApplyingEnvApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	klog.Infof("Starting ApplyEnv operation for app: %s", a.manager.Name)

	opCtx, cancel := context.WithCancel(context.Background())
	return appFactory.execAndWatch(opCtx, a,
		func(c context.Context) (StatefulInProgressApp, error) {
			in := applyingEnvInProgressApp{
				ApplyingEnvApp: a,
				baseStatefulInProgressApp: &baseStatefulInProgressApp{
					done:   c.Done,
					cancel: cancel,
				},
			}

			go func() {
				defer cancel()

				err := a.exec(c)
				if err != nil {
					a.finally = func() {
						klog.Info("ApplyEnv operation failed, update app status to ApplyEnvFailed, ", a.manager.Name)
						opRecord := makeRecord(a.manager, appsv1.ApplyEnvFailed,
							fmt.Sprintf(constants.OperationFailedTpl, a.manager.Spec.OpType, err.Error()))

						updateErr := a.updateStatus(context.Background(), a.manager, appsv1.ApplyEnvFailed, opRecord, err.Error(), "")
						if updateErr != nil {
							klog.Errorf("update appmgr state to ApplyEnvFailed state failed %v", updateErr)
							return
						}
					}
					return
				}

				a.finally = func() {
					klog.Info("ApplyEnv operation success, update app status to Initializing, ", a.manager.Name)
					updateErr := a.updateStatus(context.Background(), a.manager, appsv1.Initializing, nil, "Environment variables applied, waiting for application to initialize", "")
					if updateErr != nil {
						klog.Errorf("update appmgr state to Initializing state failed %v", updateErr)
					}
				}
			}()

			return &in, nil
		})
}

func (a *ApplyingEnvApp) exec(ctx context.Context) error {
	var err error

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("Failed to get kube config: %v", err)
		return err
	}

	token := a.manager.Annotations[api.AppTokenKey]

	var appCfg *appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(a.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("Failed to unmarshal app config: %v", err)
		return err
	}

	helmOps, err := versioned.NewHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{Source: a.manager.Spec.Source, MarketSource: a.manager.GetMarketSource()})
	if err != nil {
		klog.Errorf("Failed to create HelmOps: %v", err)
		return err
	}

	if err := helmOps.ApplyEnv(); err != nil {
		klog.Errorf("Failed to upgrade chart with environment variables: %v", err)
		return err
	}

	klog.Infof("ApplyEnv operation completed successfully for app: %s", a.manager.Name)
	return nil
}

func (a *ApplyingEnvApp) Cancel(ctx context.Context) error {
	err := a.updateStatus(ctx, a.manager, appsv1.ApplyingEnvCanceling, nil, constants.OperationCanceledByTerminusTpl, "")
	if err != nil {
		klog.Errorf("update appmgr state to upgradingCanceling state failed %v", err)
		return err
	}
	return nil
}

var _ StatefulInProgressApp = &applyingEnvInProgressApp{}

type applyingEnvInProgressApp struct {
	*ApplyingEnvApp
	*baseStatefulInProgressApp
}

func (p *applyingEnvInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

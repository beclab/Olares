package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

var _ OperationApp = &InitializingApp{}

type InitializingApp struct {
	*baseOperationApp
}

func NewInitializingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InitializingApp{
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

func (p *InitializingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	var err error
	token := p.manager.Annotations[api.AppTokenKey]

	var appCfg *appcfg.ApplicationConfig
	err = json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return nil, err
	}
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return nil, err
	}

	opCtx, cancel := context.WithCancel(context.Background())

	ops, err := versioned.NewHelmOps(opCtx, kubeConfig, appCfg, token,
		appinstaller.Opt{Source: p.manager.Spec.Source, MarketSource: p.manager.GetMarketSource()})
	if err != nil {
		klog.Errorf("make helm ops failed %v", err)
		cancel()
		return nil, err
	}

	return appFactory.execAndWatch(opCtx, p,
		func(c context.Context) (StatefulInProgressApp, error) {
			in := initializingInProgressApp{
				InitializingApp: p,
				baseStatefulInProgressApp: &baseStatefulInProgressApp{
					done:   c.Done,
					cancel: cancel,
				},
			}

			go func() {
				defer cancel()

				ok, err := ops.WaitForLaunch()
				if !ok {
					klog.Errorf("wait for launch failed %v", err)
					if err != nil {
						klog.Error("wait for launch error: ", err, ", ", p.manager.Name)
						p.finally = func() {
							klog.Info("update app manager status to initializing canceling, ", p.manager.Name)
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InitializingCanceling, nil, appsv1.InitializingCanceling.String(), constants.AppStopDueToInitFailed)
							if updateErr != nil {
								klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InitializingCanceling, updateErr)
								return
							}

						}
					}
					return
				}

				p.finally = func() {
					klog.Info("update app manager status to running, ", p.manager.Name)
					message := fmt.Sprintf(constants.InstallOperationCompletedTpl, p.manager.Spec.Type.String(), p.manager.Spec.AppName)
					if p.manager.Spec.OpType == appsv1.UpgradeOp {
						message = fmt.Sprintf(constants.UpgradeOperationCompletedTpl, p.manager.Spec.Type.String(), p.manager.Spec.AppName)
					}
					if p.manager.Spec.OpType == appsv1.ApplyEnvOp {
						message = fmt.Sprintf(constants.ApplyEnvOperationCompletedTpl, p.manager.Spec.Type.String(), p.manager.Spec.AppName)
					}
					opRecord := makeRecord(p.manager, appsv1.Running, message)
					updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Running, opRecord, appsv1.Running.String(), "")
					if updateErr != nil {
						klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Running, updateErr)
						return
					}
				}
			}()

			return &in, nil
		},
	)

}

func (p *InitializingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.InitializingCanceling, nil, constants.OperationCanceledByTerminusTpl, "")
	if err != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InitializingCanceling, err)
		return err
	}
	return nil
}

var _ StatefulInProgressApp = &initializingInProgressApp{}

type initializingInProgressApp struct {
	*InitializingApp
	*baseStatefulInProgressApp
}

// override to avoid duplicate exec
func (p *initializingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

package appstate

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &ApplyingEnvApp{}

type ApplyingEnvApp struct {
	*baseOperationApp
	// landedState overrides the default success transition (Initializing) when
	// applyEnv runs against an app that was Stopped. A stopped (workloadReplicas)
	// app keeps its release scaled to zero through the env upgrade, so there are
	// no pods to wait for: it lands back in Stopped with the new env baked in.
	landedState  appsv1.ApplicationManagerState
	landedReason string
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

				if a.landedState != "" {
					landed := a.landedState
					reason := a.landedReason
					if reason == "" {
						reason = landed.String()
					}
					a.finally = func() {
						klog.Infof("ApplyEnv operation success, app %s landed in state %s (reason=%s)", a.manager.Name, landed, reason)
						updateErr := a.updateStatus(context.Background(), a.manager, landed, nil, landed.String(), reason)
						if updateErr != nil {
							klog.Errorf("update appmgr state to %s state failed %v", landed, updateErr)
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

	kubeConfig, err := getKubeConfig()
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

	// When the app was Stopped before applyEnv, its release is scaled to zero
	// (workloadReplicas apps; non-workloadReplicas stopped apps are deferred by
	// the AppEnv controller and never reach here). The env upgrade reuses the
	// zeroed replica values, so there are no pods to wait for: skip WaitForStartUp
	// and land back in Stopped with the new env baked into the release.
	preState := a.manager.Annotations[api.AppPreUpgradeStateKey]
	skipWaitForStartUp := preState == appsv1.Stopped.String()

	helmOps, err := newHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{
		Source:             a.manager.Spec.Source,
		MarketSource:       appcfg.GetMarketSource(a.manager),
		SkipWaitForStartUp: skipWaitForStartUp,
	})
	if err != nil {
		klog.Errorf("Failed to create HelmOps: %v", err)
		return err
	}

	if err := helmOps.ApplyEnv(); err != nil {
		klog.Errorf("Failed to upgrade chart with environment variables: %v", err)
		return err
	}

	if skipWaitForStartUp {
		klog.Infof("app %s applyEnv from Stopped state, landing back in Stopped", a.manager.Spec.AppName)
		a.landedState = appsv1.Stopped
		a.landedReason = constants.AppStopByUser
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

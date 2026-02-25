package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &InstallingApp{}

type InstallingApp struct {
	*baseOperationApp
}

func NewInstallingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallingApp{
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

func (p *InstallingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
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
	err = apputils.SetExposePorts(ctx, appCfg, nil)
	if err != nil {
		klog.Errorf("set expose ports failed %v", err)
		return nil, err
	}

	updatedConfig, err := json.Marshal(appCfg)
	if err != nil {
		klog.Errorf("marshal appConfig failed %v", err)
		return nil, err
	}
	managerCopy := p.manager.DeepCopy()
	managerCopy.Spec.Config = string(updatedConfig)
	err = p.client.Patch(ctx, managerCopy, client.MergeFrom(p.manager))
	if err != nil {
		klog.Errorf("update ApplicationManager config failed %v", err)
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
			in := installingInProgressApp{
				InstallingApp: p,
				baseStatefulInProgressApp: &baseStatefulInProgressApp{
					done:   c.Done,
					cancel: cancel,
				},
			}

			go func() {
				defer cancel()
				err = ops.Install()
				if err != nil {
					klog.Errorf("install app %s failed %v", p.manager.Spec.AppName, err)
					if errors.Is(err, errcode.ErrServerSidePodPending) {
						p.finally = func() {
							klog.Infof("app %s server side pods is pending, set stop-all annotation and update app state to stopping", p.manager.Spec.AppName)

							var am appsv1.ApplicationManager
							if err := p.client.Get(context.TODO(), types.NamespacedName{Name: p.manager.Name}, &am); err != nil {
								klog.Errorf("failed to get application manager: %v", err)
								return
							}

							if am.Annotations == nil {
								am.Annotations = make(map[string]string)
							}
							am.Annotations[api.AppStopAllKey] = "true"

							if err := p.client.Update(ctx, &am); err != nil {
								klog.Errorf("failed to set stop-all annotation: %v", err)
								return
							}

							updateErr := p.updateStatus(ctx, &am, appsv1.Stopping, nil, err.Error(), constants.AppUnschedulable)
							if updateErr != nil {
								klog.Errorf("update status failed %v", updateErr)
								return
							}
						}

						return
					}

					if errors.Is(err, errcode.ErrPodPending) {
						p.finally = func() {
							klog.Infof("app %s pods is still pending, update app state to stopping", p.manager.Spec.AppName)
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Stopping, nil, err.Error(), constants.AppUnschedulable)
							if updateErr != nil {
								klog.Errorf("update status failed %v", updateErr)
								return
							}
						}

						return
					}

					p.finally = func() {
						klog.Errorf("app %s install failed, update app state to installFailed", p.manager.Spec.AppName)
						opRecord := makeRecord(p.manager, appsv1.InstallFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, err.Error()))
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallFailed, opRecord, err.Error(), "")
						if updateErr != nil {
							klog.Errorf("update status failed %v", updateErr)
						}
					}

					return
				} // end of err != nil

				if p.manager.Spec.Type == appsv1.Middleware {
					ok, err := ops.WaitForLaunch()
					if !ok {
						klog.Errorf("wait for middleware %s launch failed %v", p.manager.Spec.AppName, err)
						if err != nil {
							p.finally = func() {
								klog.Info("update app manager status to installing canceling, ", p.manager.Name)
								updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallingCanceling, nil, appsv1.InstallingCanceling.String(), constants.AppStopDueToInitFailed)
								if updateErr != nil {
									klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InstallingCanceling, updateErr)
									return
								}

							}
						}
						return
					}
					p.finally = func() {
						message := fmt.Sprintf(constants.InstallOperationCompletedTpl, p.manager.Spec.Type.String(), p.manager.Spec.AppName)
						opRecord := makeRecord(p.manager, appsv1.Running, message)
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Running, opRecord, appsv1.Running.String(), "")
						if updateErr != nil {
							klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Running, updateErr)
							return
						}
					}
				} else {
					p.finally = func() {
						klog.Infof("app %s install successfully, update app state to initializing", p.manager.Spec.AppName)
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Initializing, nil, appsv1.Initializing.String(), "")
						if updateErr != nil {
							klog.Errorf("update status failed %v", updateErr)
							return
						}

					}
				}
			}()

			return &in, nil
		},
	)
}

func (p *InstallingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.InstallingCanceling, nil, constants.OperationCanceledByTerminusTpl, "")
	if err != nil {
		klog.Errorf("update appmgr state to installingCanceling state failed %v", err)
		return err
	}

	return nil
}

var _ StatefulInProgressApp = &installingInProgressApp{}

type installingInProgressApp struct {
	*InstallingApp
	*baseStatefulInProgressApp
}

// override to avoid duplicate exec
func (p *installingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

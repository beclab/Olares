package appstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ CancelOperationApp = &InstallingCancelingApp{}

type InstallingCancelingApp struct {
	*baseOperationApp
}

func (p *InstallingCancelingApp) IsAppCreated() bool {
	return true
}

func NewInstallingCancelingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallingCancelingApp{
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

func (p *InstallingCancelingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	klog.Infof("execute installing cancel operation appName=%s", p.manager.Spec.AppName)

	err := p.handleInstallCancel(ctx)

	if err != nil {
		klog.Error("execute installing cancel operation failed", err)

		state := appsv1.InstallingCancelFailed
		updateErr := p.updateStatus(ctx, p.manager, state, nil, state.String(), "")

		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, state, updateErr)
			return nil, updateErr
		}

		return nil, err
	}

	return &installingCancelInProgressApp{
		InstallingCancelingApp:            p,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}, nil
}

func (p *InstallingCancelingApp) handleInstallCancel(ctx context.Context) error {
	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not running", p.manager.Name)
	}

	token := p.manager.Annotations[api.AppTokenKey]
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}

	// get app config from app mgr of the app
	var appCfg *appcfg.ApplicationConfig
	err = json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return err
	}

	ops, err := versioned.NewHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{})
	if err != nil {
		klog.Errorf("make helm ops failed %v", err)
		return err
	}

	// find if there is a shared chart installed by this app owner
	sharedInstalled := false
	isAdmin, err := kubesphere.IsAdmin(ctx, kubeConfig, appCfg.OwnerName)
	if err != nil {
		klog.Errorf("Failed to check if user is admin: %v", err)
		return err
	}

	if isAdmin {
		for _, chart := range appCfg.SubCharts {
			if chart.Shared {
				client, err := kubernetes.NewForConfig(kubeConfig)
				if err != nil {
					klog.Errorf("Failed to create Kubernetes client: %v", err)
					return err
				}

				sharedChartNamespace, err := client.CoreV1().Namespaces().Get(ctx, chart.Namespace(appCfg.OwnerName), metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						klog.Infof("Shared chart namespace %s not found, skipping uninstall", chart.Namespace(appCfg.OwnerName))
						break
					}
					klog.Errorf("Failed to get shared chart namespace %s: %v", chart.Namespace(appCfg.OwnerName), err)
					return err
				}

				appNameOfSharedChart := sharedChartNamespace.Labels[constants.ApplicationNameLabel]
				installUserOfSharedChart := sharedChartNamespace.Labels[constants.ApplicationInstallUserLabel]
				if appNameOfSharedChart == appCfg.AppName && installUserOfSharedChart == appCfg.OwnerName {
					sharedInstalled = true
					klog.Infof("Found shared chart %s installed by user %s", chart.Name, appCfg.OwnerName)
					break
				}
			}
		} // end of for loop over subcharts
	} // end of isAdmin check

	if sharedInstalled {
		klog.Info("Shared chart found, uninstalling all")
		err = ops.UninstallAll()
	} else {
		err = ops.Uninstall()
	}
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		klog.Errorf("execute uninstall failed %v", err)
		return err
	}

	return nil
}

func (p *InstallingCancelingApp) Cancel(ctx context.Context) error {
	ok := appFactory.cancelOperation(p.manager.Name)
	if !ok {
		klog.Errorf("app %s operation is not ", p.manager.Name)
	}

	state := appsv1.InstallingCancelFailed
	updateErr := p.updateStatus(ctx, p.manager, state, nil, state.String(), "")

	if updateErr != nil {
		klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, state, updateErr)
		return updateErr
	}

	return nil
}

var _ PollableStatefulInProgressApp = &installingCancelInProgressApp{}

type installingCancelInProgressApp struct {
	*InstallingCancelingApp
	*basePollableStatefulInProgressApp
}

func (p *installingCancelInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}

func (p *installingCancelInProgressApp) poll(ctx context.Context) error {
	if apputils.IsProtectedNamespace(p.manager.Spec.AppNamespace) {
		return nil
	}

	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			var ns corev1.Namespace
			err := p.client.Get(ctx, types.NamespacedName{Name: p.manager.Spec.AppNamespace}, &ns)
			klog.Infof("poll namespace %s err %v", p.manager.Spec.AppNamespace, err)
			if apierrors.IsNotFound(err) {
				return nil
			}

		case <-ctx.Done():
			return fmt.Errorf("app %s execute cancel operation failed %w", p.manager.Spec.AppName, ctx.Err())
		}
	}
}

func (p *installingCancelInProgressApp) WaitAsync(ctx context.Context) {
	appFactory.waitForPolling(ctx, p, func(err error) {
		if err != nil {
			updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallingCancelFailed, nil, appsv1.InstallingCancelFailed.String(), "")
			if updateErr != nil {
				klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InstallingCancelFailed.String(), updateErr)
				return
			}

			return
		}
		message := constants.OperationCanceledByUserTpl
		if p.manager.Status.Message == constants.OperationCanceledByTerminusTpl {
			message = constants.OperationCanceledByTerminusTpl
		}
		opRecord := makeRecord(p.manager, appsv1.InstallingCanceled, message)
		updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallingCanceled, opRecord, message, "")
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InstallingCanceled.String(), updateErr)
			return
		}

	})
}

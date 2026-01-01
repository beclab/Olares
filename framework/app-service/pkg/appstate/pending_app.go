package appstate

import (
	"context"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulInProgressApp = &PendingApp{}

type PendingApp struct {
	*baseOperationApp
}

func NewPendingApp(ctx context.Context, c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {

	// Application's meta.name == ApplicationMannager's meta.name
	var app appsv1.Application
	err := c.Get(ctx, types.NamespacedName{Name: manager.Name}, &app)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get application error: ", err)
		return nil, NewStateError(err.Error())
	}

	// manager of pending state, application is not created yet
	if err == nil {
		return nil, NewErrorUnknownState(
			func() func(ctx context.Context) error {
				return func(ctx context.Context) error {
					return removeUnknownApplication(c, manager.Name)(ctx)
				}
			},
			nil, // TODO: clean up, delete all, application and application manager
		)
	}

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &PendingApp{
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

func (p *PendingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	if success, err := appFactory.addLimitedStatefulApp(ctx,
		// limit
		func() (bool, error) {
			clientset, err := utils.GetClient()
			if err != nil {
				klog.Errorf("failed to get clientset %v", err)
				return false, err
			}
			apps, err := clientset.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
			if err != nil {
				klog.Errorf("list application managers error: %v", err)
				return false, err
			}

			count := 0
			for _, app := range apps.Items {
				if app.Status.State == appsv1.Downloading {
					count++
				}
			}

			return count < 1, nil
		},

		// add
		func() error {
			p.manager.Status.State = appsv1.Downloading
			now := metav1.Now()
			p.manager.Status.StatusTime = &now
			p.manager.Status.UpdateTime = &now
			p.manager.Status.OpGeneration += 1
			err := p.client.Update(ctx, p.manager)
			if err != nil {
				klog.Error("update app manager status error, ", err, ", ", p.manager.Name)
				return err
			}
			return nil
		},
	); err != nil {
		klog.Errorf("add pending app %s to in progress map failed: %v", p.manager.Spec.AppName, err)
		return nil, err
	} else if !success {
		klog.Info("2 downloading apps are in progress, waiting for the next round")
		return nil, NewWaitingInLine(2)
	}

	return nil, nil
}

func (p *PendingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(context.TODO(), p.manager, appsv1.PendingCanceled, nil, constants.OperationCanceledByUserTpl, "")
	if err != nil {
		klog.Infof("Failed to update applicationmanagers status name=%s err=%v", p.manager.Name, err)
	}

	return err
}

func (p *PendingApp) Cleanup(ctx context.Context) {}
func (p *PendingApp) Done() <-chan struct{}       { return nil }

func removeUnknownApplication(client client.Client, name string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var app appsv1.Application
		err := client.Get(ctx, types.NamespacedName{Name: name}, &app)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Error("get application error: ", err)
			return err
		}

		if apierrors.IsNotFound(err) {
			return nil
		}

		// delete the whole namespace if the namespace is not system namespace
		if !apputils.IsProtectedNamespace(app.Spec.Namespace) {
			ns := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: app.Spec.Namespace,
				},
			}

			// application will be removed automatically when the ns is removed
			err = client.Delete(ctx, &ns)
			if err != nil {
				klog.Errorf("delete namespace %s failed %v ", app.Spec.Namespace, err)
				return err
			}

		} else {
			kubeConfig, err := ctrl.GetConfig()
			if err != nil {
				return err
			}
			actionConfig, _, err := helm.InitConfig(kubeConfig, app.Spec.Namespace)
			if err != nil {
				klog.Errorf("helm init config failed %v", err)
				return err
			}

			err = helm.UninstallCharts(actionConfig, app.Spec.Name)
			if err != nil {
				klog.Errorf("uninstall release %s failed %v", app.Spec.Name, err)
				return err
			}

		}

		return nil
	}
}

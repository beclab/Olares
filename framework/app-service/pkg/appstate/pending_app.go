package appstate

import (
	"context"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulInProgressApp = &PendingApp{}

type PendingApp struct {
	*baseOperationApp
}

func NewPendingApp(ctx context.Context, deps Deps,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {

	// Application's meta.name == ApplicationMannager's meta.name
	var app appsv1.Application
	err := deps.Client.Get(ctx, types.NamespacedName{Name: manager.Name}, &app)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get application error: ", err)
		return nil, NewStateError(err.Error())
	}

	// manager of pending state, application is not created yet
	if err == nil {
		return nil, NewErrorUnknownState(
			func() func(ctx context.Context) error {
				return func(ctx context.Context) error {
					return removeUnknownApplication(deps.Client, manager.Name)(ctx)
				}
			},
			nil, // TODO: clean up, delete all, application and application manager
		)
	}

	return deps.Factory.New(deps, manager, ttl,
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
	if success, err := p.deps.Factory.addLimitedStatefulApp(ctx,
		// limit: at most 1 concurrent Downloading app cluster-wide.
		// The count goes through the Deps.CountDownloading seam: production
		// reads a live clientset (utils.GetClient), while tests override it to
		// count via the injected controller-runtime fake client so this branch
		// can be driven without a live kubeconfig.
		func() (bool, error) {
			count, err := p.deps.CountDownloading(ctx)
			if err != nil {
				klog.Errorf("count downloading application managers error: %v", err)
				return false, err
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
			p.manager.Status.Reason = appsv1.Downloading.String()
			p.manager.Status.Message = "start to download"
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
	// Move to the canceling state and let PendingCancelingApp perform the
	// actual cleanup (stop the in-progress op, delete the namespace) before it
	// settles on PendingCanceled. This mirrors every other operating state's
	// cancel path (Operating -> *Canceling -> *Canceled) and matches the
	// declared StateTransitions[Pending] = {Downloading, PendingCanceling}.
	err := p.updateStatus(context.TODO(), p.manager, appsv1.PendingCanceling, nil, constants.InstallCanceledByTimeout, constants.InstallCancelBySystem)
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
			kubeConfig, err := getKubeConfig()
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

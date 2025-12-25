package appstate

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulApp = &RunningApp{}

type RunningApp struct {
	baseStatefulApp
}

func NewRunningApp(ctx context.Context, c client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {
	// check state
	var app appsv1.Application
	err := c.Get(ctx, types.NamespacedName{Name: manager.Name}, &app)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get application %s failed %v", manager.Name, err)
		return nil, NewStateError(err.Error())
	}

	r := &RunningApp{
		baseStatefulApp: baseStatefulApp{
			manager: manager,
			client:  c,
		},
	}

	sapp, serr := appFactory.New(c, manager, 0,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return r
		})

	if serr != nil {
		klog.Errorf("create running app %s failed %v", manager.Name, serr)
		return sapp, serr
	}

	if apierrors.IsNotFound(err) && manager.Spec.Type == appsv1.App {
		klog.Infof("application %s not found, force delete app", manager.Name)

		return nil, NewErrorUnknownState(func() func(ctx context.Context) error {
			return func(ctx context.Context) error {
				// Force delete the app if it does not exist
				err = r.forceDeleteApp(ctx)
				if err != nil {
					klog.Errorf("delete app %s failed %v", manager.Spec.AppName, err)
					return err
				}

				return nil
			}
		}, nil)
	}

	return sapp, nil
}

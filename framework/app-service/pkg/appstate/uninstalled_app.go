package appstate

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ StatefulApp = &UninstalledApp{}

type UninstalledApp struct {
	baseStatefulApp
}

func NewUninstalledApp(ctx context.Context, client client.Client,
	manager *appsv1.ApplicationManager) (StatefulApp, StateError) {

	var err error
	var app appsv1.Application
	err = client.Get(ctx, types.NamespacedName{Name: manager.Name}, &app)

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get application %s failed %v", manager.Name, err)
		return nil, NewStateError(err.Error())
	}

	r := &UninstalledApp{
		baseStatefulApp: baseStatefulApp{
			manager: manager,
			client:  client,
		},
	}

	if err == nil {
		// app is not expected to exist
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

	return r, nil
}

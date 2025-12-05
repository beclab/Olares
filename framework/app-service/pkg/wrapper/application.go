package wrapper

import (
	"context"
	"errors"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationHelper struct {
	*v1alpha1.Application
	Client any
}

func (a *ApplicationHelper) GetApplicationManger(ctx context.Context) (*v1alpha1.ApplicationManager, error) {
	if a.Client == nil {
		return nil, errors.New("client is nil")
	}

	if a.Application == nil {
		return nil, errors.New("application is nil")
	}

	switch c := a.Client.(type) {
	case *versioned.Clientset:
		return c.AppV1alpha1().ApplicationManagers().Get(ctx, a.Name, metav1.GetOptions{})
	case client.Client:
		var appManager v1alpha1.ApplicationManager
		key := types.NamespacedName{Name: a.Name}
		err := c.Get(ctx, key, &appManager)
		if err != nil {
			return nil, err
		}

		return &appManager, nil
	default:
		return nil, errors.New("unsupported client type")
	}
}

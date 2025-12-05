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

type ApplicationManagerHelper struct {
	*v1alpha1.ApplicationManager
	Client any
}

func (a *ApplicationManagerHelper) GetApplication(ctx context.Context) (*v1alpha1.Application, error) {
	if a.Client == nil {
		return nil, errors.New("client is nil")
	}

	if a.ApplicationManager == nil {
		return nil, errors.New("application manager is nil")
	}

	switch c := a.Client.(type) {
	case *versioned.Clientset:
		return c.AppV1alpha1().Applications().Get(ctx, a.Name, metav1.GetOptions{})
	case client.Client:
		var app v1alpha1.Application
		key := types.NamespacedName{Name: a.Name}
		err := c.Get(ctx, key, &app)
		if err != nil {
			return nil, err
		}
		return &app, nil
	default:
		return nil, errors.New("unsupported client type")
	}
}

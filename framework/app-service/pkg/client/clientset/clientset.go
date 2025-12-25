package clientset

import (
	v1alpha1client "github.com/beclab/Olares/framework/app-service/pkg/client/clientset/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned"
	"k8s.io/client-go/rest"
)

// ClientSet provider two client for interaction with kubernetes and application.
type ClientSet struct {
	KubeClient v1alpha1client.Client
	AppClient  *versioned.Clientset
}

// New constructs a new ClientSet.
func New(config *rest.Config) (*ClientSet, error) {
	kubeClient, err := v1alpha1client.NewKubeClient("", config)
	if err != nil {
		return nil, err
	}

	appClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClientSet{
		KubeClient: kubeClient,
		AppClient:  appClient,
	}, nil
}

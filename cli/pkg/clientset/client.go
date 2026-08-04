package clientset

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	CtrlRuntime() ctrlclient.Client
	Config() *rest.Config
}

type kubeClient struct {
	// kubernetes client
	k8s kubernetes.Interface

	// controller-runtime client
	ctrl ctrlclient.Client

	// +optional
	master string

	config *rest.Config
}

// NewKubeClient creates a Kubernetes and kubesphere client
func NewKubeClient() (Client, error) {
	var err error

	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	ctrlClient, err := ctrlclient.New(config, ctrlclient.Options{})
	if err != nil {
		return nil, err
	}

	client := kubeClient{
		k8s:    k8sClient,
		ctrl:   ctrlClient,
		master: config.Host,
		config: config,
	}

	return &client, nil
}

func (k *kubeClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubeClient) CtrlRuntime() ctrlclient.Client {
	return k.ctrl
}

func (k *kubeClient) Config() *rest.Config {
	return k.config
}

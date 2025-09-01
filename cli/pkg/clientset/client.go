package clientset

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	Config() *rest.Config
}

type kubeClient struct {
	// kubernetes client
	k8s kubernetes.Interface

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

	client := kubeClient{
		k8s:    k8sClient,
		master: config.Host,
		config: config,
	}

	return &client, nil
}

func (k *kubeClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubeClient) Config() *rest.Config {
	return k.config
}

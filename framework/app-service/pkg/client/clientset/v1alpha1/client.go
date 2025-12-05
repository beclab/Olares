package v1alpha1

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	Kubernetes() kubernetes.Interface

	// rest client instead of kubersphere  client
	// KubeSphere() kubesphere.Interface

	Config() *rest.Config
}

type kubeClient struct {
	// kubernetes client
	k8s kubernetes.Interface

	// kubeSphere client
	//ks kubesphere.Interface

	// +optional
	master string

	config *rest.Config
}

// NewKubeClientOrDie creates a KubernetesClient and panic if there is an error
func NewKubeClientOrDie(kubeConfig string, config *rest.Config) Client {
	var err error

	if kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			panic(err)
		}
	}

	k := kubeClient{
		k8s: kubernetes.NewForConfigOrDie(config),
		//ks:     kubesphere.NewForConfigOrDie(config),
		master: config.Host,
		config: config,
	}
	return &k
}

// NewKubeClient creates a KubernetesClient
func NewKubeClient(kubeConfig string, config *rest.Config) (Client, error) {
	var err error

	if kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
	}

	k := kubeClient{
		k8s: kubernetes.NewForConfigOrDie(config),
		//ks:     kubesphere.NewForConfigOrDie(config),
		master: config.Host,
		config: config,
	}
	return &k, nil
}

func (k *kubeClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

// func (k *kubeClient) KubeSphere() kubesphere.Interface {
// 	return k.ks
// }

func (k *kubeClient) Config() *rest.Config {
	return k.config
}

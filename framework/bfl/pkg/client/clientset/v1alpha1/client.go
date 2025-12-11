package v1alpha1

import (
	"sync"

	"bytetrade.io/web3os/bfl/pkg/constants"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	aruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = aruntime.NewScheme()
)

func init() {
	utilruntime.Must(iamV1alpha2.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

// KubeClient global singleton client for kubernetes and kubesphere
var KubeClient ClientInterface

var syncOnce sync.Once

type ClientInterface interface {
	Kubernetes() kubernetes.Interface

	Config() *rest.Config

	CtrlClient() client.Client
}

type kubeClient struct {
	// kubernetes client
	k8s kubernetes.Interface

	ctrlClient client.Client

	config *rest.Config
}

func init() {
	syncOnce.Do(func() {
		client, err := NewKubeClient(nil)
		if err != nil {
			panic(err)
		}

		KubeClient = client
	})
}

// NewKubeClientOrDie creates a KubernetesClient and panic if there is an error
func NewKubeClientOrDie() ClientInterface {
	var err error
	config, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	ctrlClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	k := kubeClient{
		k8s:        kubernetes.NewForConfigOrDie(config),
		config:     config,
		ctrlClient: ctrlClient,
	}
	return &k
}

func NewKubeClientWithToken(token string) (ClientInterface, error) {
	config := rest.Config{
		Host:        constants.KubeSphereAPIHost,
		BearerToken: token,
	}
	return NewKubeClient(&config)
}

// NewKubeClient creates a Kubernetes and kubesphere client
func NewKubeClient(config *rest.Config) (ClientInterface, error) {
	var err error

	if config == nil {
		config, err = ctrl.GetConfig()
		if err != nil {
			return nil, err
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	ctrlClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	client := kubeClient{
		k8s:        k8sClient,
		config:     config,
		ctrlClient: ctrlClient,
	}

	return &client, nil
}

func (k *kubeClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubeClient) Config() *rest.Config {
	return k.config
}

func (k *kubeClient) CtrlClient() client.Client {
	return k.ctrlClient
}

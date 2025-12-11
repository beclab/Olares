package client

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	// velerov2alpha1api "olares.com/backup-server/pkg/apis/velero/v2alpha1"
	sysv1 "olares.com/backup-server/pkg/generated/clientset/versioned"
)

type Factory interface {
	ClientConfig() (*rest.Config, error)

	DynamicClient() (dynamic.Interface, error)

	KubeClient() (kubernetes.Interface, error)

	Sysv1Client() (sysv1.Interface, error)
}

var _ Factory = &factory{}

type factory struct {
	config *rest.Config

	client dynamic.Interface
}

func NewFactory() (Factory, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, errors.Errorf("new rest kubeconfig: %v", err)
	}

	config.Burst = 15
	config.QPS = 50

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Errorf("new dynamic client: %v", err)
	}

	f := &factory{
		config: config,
		client: client,
	}
	return f, nil
}

func (f *factory) ClientConfig() (*rest.Config, error) {
	return f.config, nil
}

func (f *factory) DynamicClient() (dynamic.Interface, error) {
	return f.client, nil
}

func (f *factory) KubeClient() (kubernetes.Interface, error) {
	c, err := kubernetes.NewForConfig(f.config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func (f *factory) Sysv1Client() (sysv1.Interface, error) {
	c, err := sysv1.NewForConfig(f.config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func (f *factory) NewUnstructuredResources() *unstructured.UnstructuredList {
	r := new(unstructured.UnstructuredList)
	r.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "List"})

	return r
}

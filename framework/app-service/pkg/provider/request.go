package provider

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvr = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "providerregistries",
}

type RegistryInterface interface {
	Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type registryRequest struct {
	client dynamic.Interface
}

var _ RegistryInterface = &registryRequest{}

func NewRegistryRequest(client dynamic.Interface) *registryRequest {
	return &registryRequest{
		client: client,
	}
}

func (pr *registryRequest) Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return pr.client.Resource(gvr).Namespace(namespace).Get(ctx, name, opts)
}

func (pr *registryRequest) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return pr.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

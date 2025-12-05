package provider

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvrForApplicationPermission = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "applicationpermissions",
}

type ApplicationPermissionInterface interface {
	Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type applicationPermissionRequest struct {
	client dynamic.Interface
}

var _ ApplicationPermissionInterface = &applicationPermissionRequest{}

func NewApplicationPermissionRequest(client dynamic.Interface) *applicationPermissionRequest {
	return &applicationPermissionRequest{
		client: client,
	}
}

func (ap *applicationPermissionRequest) Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return ap.client.Resource(gvrForApplicationPermission).Namespace(namespace).Get(ctx, name, opts)
}

func (ap *applicationPermissionRequest) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return ap.client.Resource(gvrForApplicationPermission).Namespace(namespace).List(ctx, opts)
}

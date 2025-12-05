package tapr

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvr = schema.GroupVersionResource{
	Group:    "apr.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "middlewarerequests",
}

var gvk = schema.GroupVersionKind{
	Group:   "apr.bytetrade.io",
	Version: "v1alpha1",
	Kind:    "MiddlewareRequest",
}

// MiddlewareRequestInterface is an interface that contains operation for middleware request.
type MiddlewareRequestInterface interface {
	Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	Create(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error)
	Update(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type middlewareRequest struct {
	client dynamic.Interface
}

var _ MiddlewareRequestInterface = &middlewareRequest{}

// NewMiddlewareRequest constructs a new middleware request client.
func NewMiddlewareRequest(client dynamic.Interface) (*middlewareRequest, error) {
	return &middlewareRequest{
		client: client,
	}, nil
}

// Get takes the name and namespace of the middlewarerequest,
// and returns the unstructured middlewarerequest resource, and an error is there is any.
func (mr *middlewareRequest) Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Get(ctx, name, opts)
}

func (mr *middlewareRequest) Create(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, opts)
}

// Update takes the representation of a middlewarerequest and updates it. Returns the unstructured middlewarerequest and an error if there is any.
func (mr *middlewareRequest) Update(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Update(ctx, obj, opts)
}

// Delete takes name and namespace of the middlewarerequest and deletes it. Returns an error is on occurs.
func (mr *middlewareRequest) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return mr.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, opts)
}

// List returns the list of middlewarerequest that match the namespace and list options.
func (mr *middlewareRequest) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return mr.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

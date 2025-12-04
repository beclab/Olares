package middlewareinstaller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvr = schema.GroupVersionResource{
	Group:    "psmdb.percona.com",
	Version:  "v1",
	Resource: "perconaservermongodbs",
}

var gvk = schema.GroupVersionKind{
	Group:   "psmdb.percona.com",
	Version: "v1",
	Kind:    "PerconaServerMongoDB",
}

// MiddlewareMongodbInterface is an interface that contains operation for middleware mongodb.
type MiddlewareMongodbInterface interface {
	Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	Create(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error)
	Update(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type middlewareMongodb struct {
	client dynamic.Interface
}

var _ MiddlewareMongodbInterface = &middlewareMongodb{}

// NewMiddlewareMongodb constructs a new middleware request client.
func NewMiddlewareMongodb(client dynamic.Interface) (*middlewareMongodb, error) {
	return &middlewareMongodb{
		client: client,
	}, nil
}

// Get takes the name and namespace of the middlewaremongodb,
// and returns the unstructured middlewaremongodb resource, and an error is there is any.
func (mr *middlewareMongodb) Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Get(ctx, name, opts)
}

func (mr *middlewareMongodb) Create(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, opts)
}

// Update takes the representation of a middlewaremongodb and updates it. Returns the unstructured middlewaremongodb and an error if there is any.
func (mr *middlewareMongodb) Update(ctx context.Context, namespace string, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return mr.client.Resource(gvr).Namespace(namespace).Update(ctx, obj, opts)
}

// Delete takes name and namespace of the middlewaremongodb and deletes it. Returns an error is on occurs.
func (mr *middlewareMongodb) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return mr.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, opts)
}

// List returns the list of middlewaremongodb that match the namespace and list options.
func (mr *middlewareMongodb) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return mr.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

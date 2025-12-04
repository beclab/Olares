package tapr

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// CreateOrUpdateMiddlewareRequest is used for create or update a middleware request.
func CreateOrUpdateMiddlewareRequest(config *rest.Config, namespace string, request []byte) (*unstructured.Unstructured, error) {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return createOrUpdateMiddlewareRequest(client, namespace, request)

}

func createOrUpdateMiddlewareRequest(client dynamic.Interface, namespace string, request []byte) (*unstructured.Unstructured, error) {
	mr, err := NewMiddlewareRequest(client)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = decoder.Decode(request, &gvk, obj)
	if err != nil {
		return nil, err
	}

	ret, err := mr.Create(context.Background(), namespace, obj, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			existing, err := mr.Get(context.Background(), obj.GetNamespace(), obj.GetName(), metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			obj.SetResourceVersion(existing.GetResourceVersion())
			ret, err = mr.Update(context.Background(), namespace, obj, metav1.UpdateOptions{})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return ret, nil
}

func deleteMiddlewareRequest(ctx context.Context, client dynamic.Interface, namespace string, name string) error {
	mr, err := NewMiddlewareRequest(client)
	if err != nil {
		return err
	}
	return mr.Delete(ctx, namespace, name, metav1.DeleteOptions{})
}

// DeleteMiddlewareRequest is used for delete a middleware request.
func DeleteMiddlewareRequest(ctx context.Context, config *rest.Config, namespace, name string) error {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	return deleteMiddlewareRequest(ctx, client, namespace, name)
}

package dynamic_client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type ResourceClient[T any] struct {
	*resourceDynamicClient
}

func NewResourceClient[T any](gvr schema.GroupVersionResource) (*ResourceClient[T], error) {
	ri, err := newResourceDynamicClient()
	if err != nil {
		return nil, err
	}
	return &ResourceClient[T]{ri.GroupVersionResource(gvr)}, nil
}

func NewResourceClientOrDie[T any](gvr schema.GroupVersionResource) *ResourceClient[T] {
	c, err := NewResourceClient[T](gvr)
	if err != nil {
		panic(err)
	}

	return c
}

func (u *ResourceClient[T]) Update(ctx context.Context, resource *T, options metav1.UpdateOptions) (*T, error) {
	obj, err := ToUnstructured(resource)
	if err != nil {
		return nil, err
	}

	err = u.update(ctx, &unstructured.Unstructured{Object: obj}, options, resource)
	return resource, err
}

func (r *ResourceClient[T]) Get(ctx context.Context, name string, options metav1.GetOptions) (*T, error) {
	if r.informer != nil {
		var (
			obj runtime.Object
			err error
		)

		if r.namespace == "" {
			obj, err = r.informer.Lister().Get(name)
		} else {
			obj, err = r.informer.Lister().ByNamespace(r.namespace).Get(name)
		}

		if err != nil {
			klog.Error("lister get object error, ", err, ", ", name)
			return nil, err
		}

		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			klog.Error("convert to unstructured error, ", err, ", ", name)
			return nil, err
		}

		var v T
		err = r.unmarshal(unstructuredObj, &v)
		return &v, err
	}

	data, err := r.c.Resource(r.gvr).Namespace(r.namespace).Get(ctx, name, options)
	if err != nil {
		return nil, err
	}

	var v T
	err = r.unmarshal(data.UnstructuredContent(), v)
	return &v, err
}

func (r *ResourceClient[T]) List(ctx context.Context, options metav1.ListOptions) ([]*T, error) {
	if r.informer != nil {
		// cached listing
		list, err := r.informer.Lister().ByNamespace(r.namespace).List(labels.Everything())
		if err != nil {
			klog.Error("lister list object error, ", err, ", ", options)
			return nil, err
		}

		var ret []*T
		for _, item := range list {
			unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item)
			if err != nil {
				klog.Error("convert to unstructured error, ", err, ", ", item)
				return nil, err
			}

			var v T
			err = r.unmarshal(unstructuredObj, &v)
			if err != nil {
				klog.Error("convert object in list error, ", err, ", ", v)
				return nil, err
			}

			ret = append(ret, &v)
		}
		return ret, nil
	}

	data, err := r.c.Resource(r.gvr).Namespace(r.namespace).List(ctx, options)
	if err != nil {
		return nil, err
	}

	if len(data.Items) == 0 {
		return nil, nil
	}

	var ret []*T
	for _, item := range data.Items {
		var v T
		err = r.unmarshal(item.UnstructuredContent(), &v)
		if err != nil {
			klog.Error("convert object in list error, ", err, ", ", v)
			return nil, err
		}

		ret = append(ret, &v)
	}

	return ret, nil
}

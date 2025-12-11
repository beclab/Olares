package dynamic_client

import (
	"context"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	ctrl "sigs.k8s.io/controller-runtime"
)

var syncOnce sync.Once

type resourceDynamicClient struct {
	namespace string
	gvr       schema.GroupVersionResource
	c         dynamic.Interface
	informer  informers.GenericInformer
}

var resourceInformerFactory dynamicinformer.DynamicSharedInformerFactory
var clientCtx context.Context

func init() {
	syncOnce.Do(func() {
		config := ctrl.GetConfigOrDie()
		client := dynamic.NewForConfigOrDie(config)
		resourceInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(client, 0)
		clientCtx = context.Background()
	})
}

func newResourceDynamicClient() (*resourceDynamicClient, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &resourceDynamicClient{c: client}, nil
}

func (r *resourceDynamicClient) Namespace(ns string) *resourceDynamicClient {
	r.namespace = ns
	return r
}

func (r *resourceDynamicClient) GroupVersionResource(gvr schema.GroupVersionResource) *resourceDynamicClient {
	r.gvr = gvr
	r.informer = resourceInformerFactory.ForResource(gvr)

	// add a new resource informer, start to sync cache
	// factory will not start syncing duplicately
	resourceInformerFactory.Start(clientCtx.Done())
	resourceInformerFactory.WaitForCacheSync(clientCtx.Done())

	return r
}

func (r *resourceDynamicClient) unmarshal(v map[string]any, obj any) error {
	return UnstructuredConverter.FromUnstructured(v, obj)
}

func (r *resourceDynamicClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return r.c.Resource(r.gvr).Namespace(r.namespace).Delete(ctx, name, options)
}

func (r *resourceDynamicClient) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, v any) error {
	data, err := r.c.Resource(r.gvr).Namespace(r.namespace).Create(ctx, obj, options)
	if err != nil {
		return err
	}
	return r.unmarshal(data.UnstructuredContent(), v)
}

func (r *resourceDynamicClient) update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, v any) error {
	data, err := r.c.Resource(r.gvr).Namespace(r.namespace).Update(ctx, obj, options)
	if err != nil {
		return err
	}

	return r.unmarshal(data.UnstructuredContent(), v)
}

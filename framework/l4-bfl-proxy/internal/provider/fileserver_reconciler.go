package provider

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
)

// FileserverReconciler refreshes L4 endpoints when Files pods or Nodes change.
// It performs no Kubernetes writes: the Provider rebuilds the node-to-Pod-IP
// snapshot and publishes xDS through its existing debounce/retry loop.
type FileserverReconciler struct {
	cache        ctrlcache.Cache
	OnReconciled func()
}

func NewFileserverReconciler(c ctrlcache.Cache) *FileserverReconciler {
	return &FileserverReconciler{cache: c}
}

func (r *FileserverReconciler) Name() string { return "fileserver-reconciler" }

func (r *FileserverReconciler) notifyChanged() {
	if r.OnReconciled != nil {
		r.OnReconciled()
	}
}

func (r *FileserverReconciler) podHandler() toolscache.ResourceEventHandler {
	return toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			if tombstone, ok := obj.(toolscache.DeletedFinalStateUnknown); ok {
				obj = tombstone.Obj
			}
			pod, ok := obj.(*corev1.Pod)
			return ok && isFileServerPod(pod)
		},
		Handler: toolscache.ResourceEventHandlerFuncs{
			AddFunc:    func(_ interface{}) { r.notifyChanged() },
			UpdateFunc: func(_, _ interface{}) { r.notifyChanged() },
			DeleteFunc: func(_ interface{}) { r.notifyChanged() },
		},
	}
}

// Register before the Manager starts so its cache sync includes both informers.
func (r *FileserverReconciler) SetupWithManager(ctx context.Context) error {
	pods, err := r.cache.GetInformer(ctx, &corev1.Pod{})
	if err != nil {
		return fmt.Errorf("get fileserver pod informer: %w", err)
	}
	if _, err := pods.AddEventHandler(r.podHandler()); err != nil {
		return fmt.Errorf("add fileserver pod handler: %w", err)
	}
	nodes, err := r.cache.GetInformer(ctx, &corev1.Node{})
	if err != nil {
		return fmt.Errorf("get fileserver node informer: %w", err)
	}
	if _, err := nodes.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { r.notifyChanged() },
		UpdateFunc: func(_, _ interface{}) { r.notifyChanged() },
		DeleteFunc: func(_ interface{}) { r.notifyChanged() },
	}); err != nil {
		return fmt.Errorf("add fileserver node handler: %w", err)
	}
	return nil
}

func (r *FileserverReconciler) Start(ctx context.Context) error {
	r.notifyChanged()
	<-ctx.Done()
	return nil
}

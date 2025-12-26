package watchers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const WatcherName = "sys-event-watchers"
const applicationQueueDelay = 15 * time.Second

type Action int

const (
	UNKNOWN Action = iota
	ADD
	UPDATE
	DELETE
	SUSPEND
)

type SubscribeFunc interface {
	Do(ctx context.Context, o interface{}, a Action) error
}

type EnqueueObj struct {
	Obj       interface{}
	Action    Action
	Subscribe SubscribeFunc // func is unhashable
}

type Watchers struct {
	ctx             context.Context
	workqueue       workqueue.RateLimitingInterface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

func NewWatchers(ctx context.Context, kubeconfig *rest.Config, resync time.Duration) *Watchers {
	client := dynamic.NewForConfigOrDie(kubeconfig)
	return &Watchers{
		ctx:             ctx,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Watchers"),
		informerFactory: dynamicinformer.NewDynamicSharedInformerFactory(client, resync),
	}
}

func (l *Watchers) Run(workers int) error {
	defer func() {
		utilruntime.HandleCrash()
		l.workqueue.ShutDown()
	}()
	l.informerFactory.Start(l.ctx.Done())

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting ", WatcherName)

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	res := l.informerFactory.WaitForCacheSync(l.ctx.Done())
	for t, ok := range res {
		if !ok {
			return fmt.Errorf("failed to wait for caches to sync, %s", t.String())
		}
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(l.runWorker, time.Second, l.ctx.Done())
	}

	klog.Info("Started workers")
	<-l.ctx.Done()
	klog.Info("Shutting down workers, ", WatcherName)

	return nil
}

func (l *Watchers) Enqueue(obj EnqueueObj) {
	l.workqueue.Add(obj)
}

func (l *Watchers) runWorker() {
	for l.processNextWorkItem() {
	}
}

func (l *Watchers) processNextWorkItem() bool {
	obj, shutdown := l.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer l.workqueue.Done(obj)
		var eobj EnqueueObj
		var ok bool
		if eobj, ok = obj.(EnqueueObj); !ok {
			l.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := eobj.Subscribe.Do(l.ctx, eobj.Obj, eobj.Action); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			l.workqueue.AddAfter(eobj, applicationQueueDelay)
			return fmt.Errorf("error syncing '%v': %s, requeuing", eobj, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		l.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%v'", eobj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func AddToWatchers[R any](w *Watchers, gvr schema.GroupVersionResource, handler cache.ResourceEventHandler) error {
	informer := w.informerFactory.ForResource(gvr)
	klog.Info("add resource to watch, ", gvr.String())
	if handler != nil {
		convert := func(obj interface{}, newObj *R) error {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(*unstructured.Unstructured).Object, newObj)
			if err != nil {
				klog.Error("convert obj error, ", err)
				return err
			}
			return nil
		}

		newHandler := cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch h := handler.(type) {
				case cache.FilteringResourceEventHandler:
					var newObj R
					err := convert(obj, &newObj)
					if err != nil {
						return false
					}
					return h.FilterFunc(&newObj)
				}

				return true
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					var newObj R
					err := convert(obj, &newObj)
					if err != nil {
						return
					}
					switch h := handler.(type) {
					case cache.FilteringResourceEventHandler:
						h.Handler.(cache.ResourceEventHandlerFuncs).AddFunc(&newObj)
					case cache.ResourceEventHandlerFuncs:
						h.AddFunc(&newObj)
					}
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					var convNewObj, convOldObj R
					err := convert(newObj, &convNewObj)
					if err != nil {
						return
					}
					err = convert(oldObj, &convOldObj)
					if err != nil {
						return
					}

					switch h := handler.(type) {
					case cache.FilteringResourceEventHandler:
						h.Handler.(cache.ResourceEventHandlerFuncs).UpdateFunc(&convOldObj, &convNewObj)
					case cache.ResourceEventHandlerFuncs:
						h.UpdateFunc(&convOldObj, &convNewObj)
					}
				},
				DeleteFunc: func(obj interface{}) {
					var newObj R
					err := convert(obj, &newObj)
					if err != nil {
						return
					}
					switch h := handler.(type) {
					case cache.FilteringResourceEventHandler:
						h.Handler.(cache.ResourceEventHandlerFuncs).DeleteFunc(&newObj)
					case cache.ResourceEventHandlerFuncs:
						h.DeleteFunc(&newObj)
					}
				},
			},
		}
		informer.Informer().AddEventHandler(newHandler)
	}

	return nil
}

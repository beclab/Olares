package prodiverregistry

import (
	"encoding/json"
	"fmt"
	"time"

	clientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned/scheme"
	informers "bytetrade.io/web3os/system-server/pkg/generated/informers/externalversions/sys/v1alpha1"
	listers "bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "providerregistry-controller"

type Controller struct {
	sysClientset           clientset.Interface
	providerLister         listers.ProviderRegistryLister
	providerRegistrySynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewController(sysClientset clientset.Interface,
	prInformer informers.ProviderRegistryInformer) *Controller {
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))

	controller := &Controller{
		sysClientset:           sysClientset,
		providerLister:         prInformer.Lister(),
		providerRegistrySynced: prInformer.Informer().HasSynced,
		workqueue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ProviderRegistry"),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when providerregistry resources change
	prInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleAddObject,
		UpdateFunc: func(old, new interface{}) {
			if updated, err := diff(old, new); err != nil {
				klog.Error("diff error: ", err)
			} else if updated {
				controller.handleUpdateObject(new)
			}
		},
		DeleteFunc: controller.handleDeleteObject,
	})

	return controller
}

func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) handleAddObject(obj interface{}) {
	// filter obj
	klog.Info("handle add object")
	c.enqueue(obj)
}

func (c *Controller) handleUpdateObject(obj interface{}) {
	// filter obj
	klog.Info("handle update object ")

	c.enqueue(obj)
}

func (c *Controller) handleDeleteObject(obj interface{}) {
	// filter obj
	klog.Info("handle delete object")

	c.enqueue(obj)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting ProviderRegistry controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.providerRegistrySynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		//
		// TODO: set the provider or watcher's state to suspend when the pod replica of
		// the provider or watcher equals to zero
		//

		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		// if err := c.syncHandler(key); err != nil {
		// 	// Put the item back on the workqueue to handle any transient errors.
		// 	c.workqueue.AddRateLimited(key)
		// 	return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		// }
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func diff(old interface{}, new interface{}) (bool, error) {
	olddata, err := json.Marshal(old)
	if err != nil {
		return false, err
	}

	newdata, err := json.Marshal(new)
	if err != nil {
		return false, err
	}

	return string(olddata) != string(newdata), nil
}

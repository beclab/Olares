package pgcluster

import (
	"context"
	"fmt"
	"time"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	informers "bytetrade.io/web3os/tapr/pkg/generated/informers/externalversions"
	"bytetrade.io/web3os/tapr/pkg/generated/listers/apr/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "pgcluster-controller"

type Action int

const (
	UNKNOWN Action = iota
	ADD
	UPDATE
	DELETE
)

type controller struct {
	workqueue            workqueue.RateLimitingInterface
	informerFactory      informers.SharedInformerFactory
	synced               cache.InformerSynced
	informer             cache.SharedIndexInformer
	lister               v1alpha1.PGClusterLister
	aprClientSet         *aprclientset.Clientset
	k8sClientSet         *kubernetes.Clientset
	ctx                  context.Context
	cancel               context.CancelFunc
	requestLister        v1alpha1.MiddlewareRequestLister
	notifyClusterCreated func(cluster *aprv1.PGCluster)
}

type enqueueObj struct {
	action Action
	obj    interface{}
}

func NewController(kubeConfig *rest.Config, mainCtx context.Context, mrLister v1alpha1.MiddlewareRequestLister, notifyFn func(cluster *aprv1.PGCluster)) (*controller, v1alpha1.PGClusterLister) {
	clientset := aprclientset.NewForConfigOrDie(kubeConfig)

	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	informer := informerFactory.Apr().V1alpha1().PGClusters()

	ctrlr := &controller{
		aprClientSet:         clientset,
		k8sClientSet:         kubernetes.NewForConfigOrDie(kubeConfig),
		informerFactory:      informerFactory,
		informer:             informer.Informer(),
		lister:               informer.Lister(),
		requestLister:        mrLister,
		synced:               informer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pgcluster"),
		notifyClusterCreated: notifyFn,
	}

	_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrlr.handleAddObject,
		UpdateFunc: func(old, new interface{}) {
			ctrlr.handleUpdateObject(new)
		},
		DeleteFunc: ctrlr.handleDeleteObject,
	})

	if err != nil {
		klog.Error("create pgcluster controller error, ", err)
		panic(err)
	}

	ctrlr.ctx, ctrlr.cancel = context.WithCancel(mainCtx)
	return ctrlr, ctrlr.lister
}

func (c *controller) enqueue(obj enqueueObj) {
	// var key string
	// var err error
	// if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
	// 	utilruntime.HandleError(err)
	// 	return
	// }
	c.workqueue.Add(obj)
}

func (c *controller) handleAddObject(obj interface{}) {
	// filter obj
	klog.Info("handle add object")
	c.enqueue(enqueueObj{ADD, obj})
}

func (c *controller) handleUpdateObject(obj interface{}) {
	// filter obj
	klog.Info("handle update object ")

	c.enqueue(enqueueObj{UPDATE, obj})
}

func (c *controller) handleDeleteObject(obj interface{}) {
	// filter obj
	klog.Info("handle delete object")

	c.enqueue(enqueueObj{DELETE, obj})
}

func (c *controller) Run(workers int) error {
	defer func() {
		utilruntime.HandleCrash()
		c.workqueue.ShutDown()
		c.informerFactory.Shutdown()
	}()
	c.informerFactory.Start(c.ctx.Done())

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting ", controllerAgentName)

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(c.ctx.Done(), c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, c.ctx.Done())
	}

	klog.Info("Started workers")
	<-c.ctx.Done()
	klog.Info("Shutting down workers, ", controllerAgentName)

	return nil
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var eobj enqueueObj
		var ok bool
		if eobj, ok = obj.(enqueueObj); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(eobj); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(eobj)
			return fmt.Errorf("error syncing '%v': %s, requeuing", eobj, err.Error())
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully pg synced '%v'", eobj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *controller) syncHandler(obj enqueueObj) error {

	return c.handler(obj.action, obj.obj)
}

func (c *controller) Cancel() {
	c.cancel()
}

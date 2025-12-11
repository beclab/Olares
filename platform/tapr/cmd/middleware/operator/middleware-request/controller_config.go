package middlewarerequest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	workload_nats "bytetrade.io/web3os/tapr/pkg/workload/nats"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type configmapController struct {
	workqueue       workqueue.RateLimitingInterface
	informerFactory informers.SharedInformerFactory
	synced          cache.InformerSynced
	informer        cache.SharedIndexInformer
	lister          corelisters.ConfigMapLister
	k8sClientSet    *kubernetes.Clientset
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewConfigmapController(kubeConfig *rest.Config, mainCtx context.Context) (*configmapController, corelisters.ConfigMapLister) {
	clientSet := kubernetes.NewForConfigOrDie(kubeConfig)
	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	informer := informerFactory.Core().V1().ConfigMaps()
	lister := informer.Lister()

	ctr := &configmapController{
		k8sClientSet:    clientSet,
		informerFactory: informerFactory,
		informer:        informer.Informer(),
		lister:          lister,
		synced:          informer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nats-config"),
	}
	ctr.ctx, ctr.cancel = context.WithCancel(mainCtx)
	_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctr.handleAddObject,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctr.handleUpdateObject(newObj)
		},
	})
	if err != nil {
		klog.Errorf("create configmap controller err=%v", err)
		panic(err)
	}
	return ctr, lister
}

func (c *configmapController) enqueue(obj enqueueObj) {

	c.workqueue.Add(obj)
}

func (c *configmapController) handleAddObject(obj interface{}) {
	klog.Infof("handle configmap add object")
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		klog.Infof("Not a ConfigMap")
		return
	}
	if cm.Name != "nats-config" {
		return
	}
	c.enqueue(enqueueObj{ADD, obj})
}

func (c *configmapController) handleUpdateObject(obj interface{}) {
	return
}

func (c *configmapController) Run(workers int) error {
	defer func() {
		utilruntime.HandleCrash()
		c.workqueue.ShutDown()
		c.informerFactory.Shutdown()
	}()
	c.informerFactory.Start(c.ctx.Done())

	klog.Infof("Starting configmap controller")
	if ok := cache.WaitForCacheSync(c.ctx.Done(), c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	klog.Info("starting configmap controller workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, c.ctx.Done())
	}
	klog.Info("started configmap controller workers")
	<-c.ctx.Done()
	klog.Info("Shutting down config")
	return nil
}
func (c *configmapController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *configmapController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var eobj enqueueObj
		var ok bool
		if eobj, ok = obj.(enqueueObj); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		for e := c.syncHandler(eobj); e != nil; e = c.syncHandler(eobj) {
			c.workqueue.AddRateLimited(eobj)
			return fmt.Errorf("error syncing configmap '%v': %s, requeuing", eobj, e.Error())
		}
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced configmap %v", eobj)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *configmapController) syncHandler(obj enqueueObj) error {
	klog.Infof("configmap syncHandler.......")
	return c.handler(obj.action, obj.obj)
}

func (c *configmapController) handler(action Action, obj interface{}) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return errors.New("invalid configmap object")
	}

	if _, err := os.Stat(workload_nats.ConfPath); err == nil {
		return nil
	}

	natsConf, exists := cm.Data["nats.conf"]
	if !exists {
		klog.Infof("nats.conf not found in configmap data")
		return errors.New("nats.conf not found in configmap data")
	}
	err := os.MkdirAll(filepath.Dir(workload_nats.ConfPath), 0755)
	if err != nil {
		klog.Infof("mkdirall err=%v", err)
		return err
	}
	err = ioutil.WriteFile(workload_nats.ConfPath, []byte(natsConf), 0644)
	if err != nil {
		klog.Infof("writefile err=%v", err)
		return err
	}
	return nil
}

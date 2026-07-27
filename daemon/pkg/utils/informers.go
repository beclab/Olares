package utils

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	applisters "github.com/beclab/api/pkg/generated/listers/app.bytetrade.io/v1alpha1"

	appinformers "github.com/beclab/api/pkg/generated/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/beclab/api/pkg/generated/clientset/versioned"
)

// informerResyncPeriod is the shared informer resync period. The cache is kept
// fresh by watches; the periodic resync is only a safety net.
const informerResyncPeriod = 10 * time.Minute

// Shared informers back the read-heavy parts of the status loop so that
// olaresd reads Kubernetes objects from a local cache instead of doing a full
// cluster-wide List + deserialize on every 5s tick.
//
// The factories are started lazily on first use, once the cluster is reachable.
// Every accessor falls back to a live List while the cache is not yet synced,
// so the status loop never misjudges cluster health during cache warm-up.
var (
	informerMu         sync.Mutex
	informerCtx        context.Context
	informerCtxCancel  context.CancelFunc
	informerCtxCreator func() (context.Context, context.CancelFunc)

	informersStarted bool

	coreFactory k8sinformers.SharedInformerFactory
	podLister   corelisters.PodLister
	podsSynced  cache.InformerSynced
	nodeLister  corelisters.NodeLister
	nodesSynced cache.InformerSynced

	appFactory appinformers.SharedInformerFactory
	appLister  applisters.ApplicationLister
	appsSynced cache.InformerSynced

	dynFactory  dynamicinformer.DynamicSharedInformerFactory
	userLister  cache.GenericLister
	usersSynced cache.InformerSynced
)

// logLiveFallback records (at debug verbosity) that a read bypassed the
// informer cache and hit the API server directly, to help diagnose whether the
// cache is actually serving reads in production.
func logLiveFallback(resource string) {
	klog.V(4).Infof("informer cache for %s not synced, falling back to live List", resource)
}

// InitInformers stores the lifecycle context for the shared informers. It
// should be called once at startup before the informers are used. The factories
// themselves are started lazily on first access, once the cluster is reachable.
func InitInformers(ctx context.Context) {
	informerMu.Lock()
	defer informerMu.Unlock()
	informerCtxCreator = func() (context.Context, context.CancelFunc) {
		return context.WithCancel(ctx)
	}
	informerCtx, informerCtxCancel = informerCtxCreator()
}

// clientGeneration is bumped whenever cached k8s clients are invalidated so
// startInformersIfNeeded can detect that clients fetched outside informerMu
// became stale before factories were started.
var clientGeneration atomic.Uint64

// rebuildInformers stops running factories and clears listers so subsequent
// reads fall back to live Lists until the cache is rebuilt.
//
// Locking: safe to call while holding clientCacheMu. It takes informerMu only
// to swap state (never acquires clientCacheMu), then runs Shutdown outside
// informerMu. Callers that start informers must not hold informerMu while
// calling GetKubeClient, or clientCacheMu → informerMu here would deadlock.
func rebuildInformers() {
	informerMu.Lock()
	cancel := informerCtxCancel
	core, app, dyn := coreFactory, appFactory, dynFactory

	coreFactory = nil
	appFactory = nil
	dynFactory = nil
	podLister = nil
	podsSynced = nil
	nodeLister = nil
	nodesSynced = nil
	appLister = nil
	appsSynced = nil
	userLister = nil
	usersSynced = nil
	informersStarted = false

	if informerCtxCreator != nil {
		informerCtx, informerCtxCancel = informerCtxCreator()
	} else {
		informerCtx = nil
		informerCtxCancel = nil
	}
	informerMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if core != nil {
		core.Shutdown()
	}
	if app != nil {
		app.Shutdown()
	}
	if dyn != nil {
		dyn.Shutdown()
	}
}

// startInformersIfNeeded lazily creates and starts the shared informer factories
// once the cluster is reachable. Clients are fetched without holding
// informerMu so rebuildInformers can take informerMu while holding
// clientCacheMu without deadlocking.
//
// Factories are restarted after kubeconfig/hosts changes invalidate clients
// (see ensureFreshLocked → rebuildInformers).
func startInformersIfNeeded() {
	for {
		informerMu.Lock()
		if informerCtx == nil || informersStarted {
			informerMu.Unlock()
			return
		}
		informerMu.Unlock()

		gen := clientGeneration.Load()
		kubeClient, appClientSet, dynClient, err := informerClients()
		if err != nil {
			return
		}
		if clientGeneration.Load() != gen {
			// Clients were invalidated while we were fetching; retry with fresh ones.
			continue
		}

		informerMu.Lock()
		if informerCtx == nil || informersStarted {
			informerMu.Unlock()
			return
		}
		if clientGeneration.Load() != gen {
			informerMu.Unlock()
			continue
		}

		coreFactory = k8sinformers.NewSharedInformerFactory(kubeClient, informerResyncPeriod)
		podInformer := coreFactory.Core().V1().Pods()
		podLister = podInformer.Lister()
		podsSynced = podInformer.Informer().HasSynced
		nodeInformer := coreFactory.Core().V1().Nodes()
		nodeLister = nodeInformer.Lister()
		nodesSynced = nodeInformer.Informer().HasSynced

		appFactory = appinformers.NewSharedInformerFactory(appClientSet, informerResyncPeriod)
		applicationInformer := appFactory.App().V1alpha1().Applications()
		appLister = applicationInformer.Lister()
		appsSynced = applicationInformer.Informer().HasSynced

		dynFactory = dynamicinformer.NewDynamicSharedInformerFactory(dynClient, informerResyncPeriod)
		userInformer := dynFactory.ForResource(UserGVR)
		userLister = userInformer.Lister()
		usersSynced = userInformer.Informer().HasSynced

		stop := informerCtx.Done()
		coreFactory.Start(stop)
		appFactory.Start(stop)
		dynFactory.Start(stop)

		informersStarted = true
		informerMu.Unlock()
		klog.Info("shared informers started")
		return
	}
}

func informerClients() (kubernetes.Interface, *versioned.Clientset, dynamic.Interface, error) {
	kubeClient, err := GetKubeClient()
	if err != nil {
		return nil, nil, nil, err
	}
	appClientSet, err := GetAppClientSet()
	if err != nil {
		return nil, nil, nil, err
	}
	dynClient, err := GetDynamicClient()
	if err != nil {
		return nil, nil, nil, err
	}
	return kubeClient, &appClientSet, dynClient, nil
}

// syncedLister starts the informers if needed and returns the lister together
// with whether its cache has synced, under informerMu.
func syncedLister[T any](getLister func() (T, bool)) (T, bool) {
	startInformersIfNeeded()
	informerMu.Lock()
	defer informerMu.Unlock()
	return getLister()
}

// ListPods returns all pods across namespaces, preferring the synced informer
// cache and falling back to a live List while the cache is warming up (or when
// informers were never initialized, e.g. in tests).
//
// Cache reads are eventually consistent and may lag the API server by a watch
// cycle; the live fallback only covers the not-yet-synced case, not a
// synced-but-stale cache.
func ListPods(ctx context.Context) ([]*corev1.Pod, error) {
	lister, ok := syncedLister(func() (corelisters.PodLister, bool) {
		return podLister, podLister != nil && podsSynced != nil && podsSynced()
	})
	if ok {
		return lister.List(labels.Everything())
	}
	logLiveFallback("pods")

	client, err := GetKubeClient()
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	pods := make([]*corev1.Pod, 0, len(list.Items))
	for i := range list.Items {
		pods = append(pods, &list.Items[i])
	}
	return pods, nil
}

// ListNodes returns all nodes, preferring the synced informer cache and falling
// back to a live List while the cache is warming up.
func ListNodes(ctx context.Context) ([]*corev1.Node, error) {
	lister, ok := syncedLister(func() (corelisters.NodeLister, bool) {
		return nodeLister, nodeLister != nil && nodesSynced != nil && nodesSynced()
	})
	if ok {
		return lister.List(labels.Everything())
	}
	logLiveFallback("nodes")

	client, err := GetKubeClient()
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodes := make([]*corev1.Node, 0, len(list.Items))
	for i := range list.Items {
		nodes = append(nodes, &list.Items[i])
	}
	return nodes, nil
}

// ListApplications returns all Application CRs. Objects served from the informer
// cache are read-only and shared, so callers that mutate must DeepCopy first.
func ListApplications(ctx context.Context) ([]*appv1alpha1.Application, error) {
	lister, ok := syncedLister(func() (applisters.ApplicationLister, bool) {
		return appLister, appLister != nil && appsSynced != nil && appsSynced()
	})
	if ok {
		return lister.List(labels.Everything())
	}
	logLiveFallback("applications")

	clientset, err := GetAppClientSet()
	if err != nil {
		return nil, err
	}
	list, err := clientset.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	apps := make([]*appv1alpha1.Application, 0, len(list.Items))
	for i := range list.Items {
		apps = append(apps, &list.Items[i])
	}
	return apps, nil
}

// listUsersRaw returns all user CRs as unstructured objects, preferring the
// synced informer cache and falling back to a live List while warming up.
func listUsersRaw(ctx context.Context) ([]*unstructured.Unstructured, error) {
	lister, ok := syncedLister(func() (cache.GenericLister, bool) {
		return userLister, userLister != nil && usersSynced != nil && usersSynced()
	})
	if ok {
		objs, err := lister.List(labels.Everything())
		if err != nil {
			return nil, err
		}
		users := make([]*unstructured.Unstructured, 0, len(objs))
		for _, o := range objs {
			if u, isU := o.(*unstructured.Unstructured); isU {
				users = append(users, u)
			}
		}
		return users, nil
	}
	logLiveFallback("users")

	client, err := GetDynamicClient()
	if err != nil {
		return nil, err
	}
	list, err := client.Resource(UserGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	users := make([]*unstructured.Unstructured, 0, len(list.Items))
	for i := range list.Items {
		users = append(users, &list.Items[i])
	}
	return users, nil
}

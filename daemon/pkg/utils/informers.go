package utils

import (
	"context"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// informerResyncPeriod is the shared informer resync period. The cache is kept
// fresh by watches; the periodic resync is only a safety net.
const informerResyncPeriod = 10 * time.Minute

// Shared informers back the read-heavy parts of the status loop so that
// olaresd reads Kubernetes objects from a local cache instead of doing a full
// cluster-wide List + deserialize on every 5s tick.
//
// The factory is started lazily on first use, once the cluster is reachable.
// Every accessor falls back to a live List while the cache is not yet synced,
// so the status loop never misjudges cluster health during cache warm-up.
var (
	informerMu       sync.Mutex
	informerCtx      context.Context
	informerCancel   context.CancelFunc
	informerClient   kubernetes.Interface
	informerFactory  k8sinformers.SharedInformerFactory
	podLister        corelisters.PodLister
	podsSynced       cache.InformerSynced
	informersStarted bool
)

// logLiveFallback records (at debug verbosity) that a read bypassed the
// informer cache and hit the API server directly, to help diagnose whether the
// cache is actually serving reads in production.
func logLiveFallback(resource string) {
	klog.V(4).Infof("informer cache for %s not synced, falling back to live List", resource)
}

// InitInformers stores the lifecycle context for the shared informers. It
// should be called once at startup before the informers are used. The factory
// itself is started lazily on first access, once the cluster is reachable.
func InitInformers(ctx context.Context) {
	informerMu.Lock()
	defer informerMu.Unlock()
	informerCtx = ctx
}

// ensureStartedLocked lazily creates and starts the shared informer factory,
// rebuilding it if the underlying kube client was replaced (e.g. after a
// kubeconfig change invalidated the cached client). Callers must hold
// informerMu.
func ensureStartedLocked() {
	if informerCtx == nil {
		return
	}

	client, err := GetKubeClient()
	if err != nil {
		// Cluster not reachable yet; try again on the next access.
		return
	}

	if informersStarted {
		if client == informerClient {
			return
		}
		// The cached kube client was rebuilt; stop the stale factory and
		// recreate it against the new client so watches keep working.
		resetInformersLocked()
		klog.Info("kube client changed, restarting shared informers")
	}

	ctx, cancel := context.WithCancel(informerCtx)
	factory := k8sinformers.NewSharedInformerFactory(client, informerResyncPeriod)
	podInformer := factory.Core().V1().Pods()
	podLister = podInformer.Lister()
	podsSynced = podInformer.Informer().HasSynced

	factory.Start(ctx.Done())
	informerFactory = factory
	informerCancel = cancel
	informerClient = client
	informersStarted = true
	klog.Info("shared informers started")
}

// resetInformersLocked stops the running factory and clears the cached
// listers so the next access rebuilds them. Callers must hold informerMu.
func resetInformersLocked() {
	if informerCancel != nil {
		informerCancel()
		informerCancel = nil
	}
	informerFactory = nil
	informerClient = nil
	podLister = nil
	podsSynced = nil
	informersStarted = false
}

// podListerIfSynced returns the pod lister only when its cache has synced,
// otherwise nil so the caller falls back to a live List.
func podListerIfSynced() corelisters.PodLister {
	informerMu.Lock()
	defer informerMu.Unlock()

	ensureStartedLocked()

	if podLister == nil || podsSynced == nil || !podsSynced() {
		return nil
	}
	return podLister
}

// ListPods returns all pods across namespaces, preferring the synced informer
// cache and falling back to a live List while the cache is warming up (or when
// informers were never initialized, e.g. in tests).
//
// Cache reads are eventually consistent and may lag the API server by a watch
// cycle; the live fallback only covers the not-yet-synced case, not a
// synced-but-stale cache.
func ListPods(ctx context.Context) ([]*corev1.Pod, error) {
	if lister := podListerIfSynced(); lister != nil {
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

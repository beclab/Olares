package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/beclab/Olares/framework/cm-sidecar/internal/syncer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	envLabelSelector = "CONFIGMAP_LABEL_SELECTOR"
	envTargetDir     = "CONFIGMAP_TARGET_DIR"
	envNamespace     = "CONFIGMAP_NAMESPACE"

	defaultLabelSelector = "bytetrade.io/cm-sidecar"
	namespaceFile        = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	metricsAddr = "0"
	probeAddr   = ":8081"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	ctrl.SetLogger(klog.NewKlogr())

	targetDir := os.Getenv(envTargetDir)
	if targetDir == "" {
		klog.Fatalf("%s is required", envTargetDir)
	}

	selectorRaw := os.Getenv(envLabelSelector)
	if selectorRaw == "" {
		selectorRaw = defaultLabelSelector
	}
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		klog.Fatalf("parse %s %q: %v", envLabelSelector, selectorRaw, err)
	}

	namespace, err := resolveNamespace()
	if err != nil {
		klog.Fatalf("resolve namespace: %v", err)
	}

	klog.Infof("syncing configmaps matching %q in namespace %s into %s", selector, namespace, targetDir)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		klog.Fatalf("add corev1 scheme failed: %v", err)
	}

	// Pushing both filters into the cache keeps only the ConfigMaps this
	// sidecar cares about in memory, and makes a plain List in the reconciler
	// return exactly the desired set.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
		Cache: ctrlcache.Options{
			DefaultNamespaces: map[string]ctrlcache.Config{namespace: {}},
			ByObject: map[client.Object]ctrlcache.ByObject{
				&corev1.ConfigMap{}: {Label: selector},
			},
		},
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false,
	})
	if err != nil {
		klog.Fatalf("create manager failed: %v", err)
	}

	reconciler := &syncer.Reconciler{
		Client:    mgr.GetClient(),
		Namespace: namespace,
		Selector:  selector,
		Dir:       targetDir,
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("setup reconciler failed: %v", err)
	}

	// Runnables start after the cache has synced, so this initial sync sees the
	// full set of ConfigMaps, including the empty set.
	if err := mgr.Add(manager.RunnableFunc(reconciler.SyncOnce)); err != nil {
		klog.Fatalf("add initial sync failed: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Fatalf("set up health check failed: %v", err)
	}
	if err := mgr.AddReadyzCheck("cache-synced", cacheReadyCheck(mgr)); err != nil {
		klog.Fatalf("set up ready check failed: %v", err)
	}

	klog.Info("starting manager...")
	if err := mgr.Start(ctx); err != nil {
		klog.Fatalf("manager exited: %v", err)
	}
	klog.Info("manager stopped")
}

// resolveNamespace returns the namespace to watch, defaulting to the namespace
// the sidecar itself runs in.
func resolveNamespace() (string, error) {
	if ns := strings.TrimSpace(os.Getenv(envNamespace)); ns != "" {
		return ns, nil
	}

	data, err := os.ReadFile(namespaceFile)
	if err != nil {
		return "", fmt.Errorf("%s is unset and reading %s failed: %w", envNamespace, namespaceFile, err)
	}

	ns := strings.TrimSpace(string(data))
	if ns == "" {
		return "", fmt.Errorf("%s is unset and %s is empty", envNamespace, namespaceFile)
	}
	return ns, nil
}

// cacheReadyCheck reports ready only once the cache has synced, so a dependent
// container gated on this probe starts after the files are on disk.
func cacheReadyCheck(mgr manager.Manager) healthz.Checker {
	return func(req *http.Request) error {
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()

		if !mgr.GetCache().WaitForCacheSync(ctx) {
			return fmt.Errorf("cache not synced yet")
		}
		return nil
	}
}

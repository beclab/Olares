package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/beclab/l4-bfl-proxy/internal/envoy"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/beclab/l4-bfl-proxy/internal/provider"
	"github.com/beclab/l4-bfl-proxy/internal/runner"
	"github.com/beclab/l4-bfl-proxy/internal/translator"
	"github.com/beclab/l4-bfl-proxy/internal/xds/server"
	xdstranslator "github.com/beclab/l4-bfl-proxy/internal/xds/translator"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	userNamespacePrefix = "user-space"
	sslServerPort       = 443
	sslProxyServerPort  = 444
	xdsServerPort       = 8794
	resyncPeriod        = 60 * time.Minute
	metricsAddr         = ":9001"
	probeAddr           = ":8081"
)

func main() {
	klog.InitFlags(nil)
	tcpIdleTimeout := flag.Duration("xds-tcp-idle-timeout", time.Hour, "tcp proxy idle timeout")
	httpStreamIdleTimeout := flag.Duration("xds-http-stream-idle-timeout", 30*time.Minute, "http stream idle timeout")
	connectTimeout := flag.Duration("xds-connect-timeout", 5*time.Second, "upstream connect timeout")
	routeTimeout := flag.Duration("xds-route-timeout", 5*time.Minute, "route timeout")
	clusterIdleTimeout := flag.Duration("xds-cluster-idle-timeout", 10*time.Second, "upstream HTTP connection idle timeout (CommonHttpProtocolOptions.IdleTimeout); set below the backend keep-alive limit to avoid stale-connection resets")

	flag.Parse()
	defer klog.Flush()
	xdstranslator.SetTimeouts(*tcpIdleTimeout, *httpStreamIdleTimeout, *connectTimeout, *routeTimeout, *clusterIdleTimeout)

	ctrl.SetLogger(klog.NewKlogr())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	scheme := runtime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		klog.Fatalf("add iam scheme failed: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		klog.Fatalf("add app scheme failed: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		klog.Fatalf("add corev1 scheme failed: %v", err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		klog.Fatalf("add rbacv1 scheme failed: %v", err)
	}

	syncPeriod := resyncPeriod
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
		Cache: ctrlcache.Options{
			SyncPeriod: &syncPeriod,
		},
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false,
	})
	if err != nil {
		klog.Fatalf("create manager failed: %v", err)
	}

	providerResources := &message.ProviderResources{}
	xdsIR := &message.XdsIR{}
	xdsResources := &message.XdsResources{}

	prov := provider.New(mgr.GetCache(), providerResources, &provider.Config{
		UserNamespacePrefix: userNamespacePrefix,
		SSLServerPort:       sslServerPort,
		SSLProxyServerPort:  sslProxyServerPort,
	})
	if err := prov.SetupWithManager(ctx); err != nil {
		klog.Fatalf("setup provider failed: %v", err)
	}

	fsReconciler := provider.NewFileserverReconciler(mgr.GetCache(), mgr.GetClient())
	fsReconciler.OnReconciled = prov.NotifyChanged

	if err := fsReconciler.SetupWithManager(ctx); err != nil {
		klog.Fatalf("setup fileserver reconciler failed: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Fatalf("set up health check failed: %v", err)
	}
	if err := mgr.AddReadyzCheck("cache-synced", cacheReadyCheck(mgr)); err != nil {
		klog.Fatalf("set up ready check failed: %v", err)
	}

	runners := []runner.Runner{
		prov,
		fsReconciler,
		translator.New(providerResources, xdsIR, &translator.Config{
			SSLServerPort:      sslServerPort,
			SSLProxyServerPort: sslProxyServerPort,
		}),
		xdstranslator.New(xdsIR, xdsResources),
		server.New(xdsResources, &server.Config{
			Address: "127.0.0.1",
			Port:    xdsServerPort,
		}),
	}
	for _, r := range runners {
		rn := r
		if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
			klog.Infof("starting runner: %s", rn.Name())
			if err := rn.Start(ctx); err != nil {
				klog.Errorf("start runner %s failed: %v", rn.Name(), err)
				cancel()
				return err
			}
			return nil
		})); err != nil {
			klog.Fatalf("add runner %s: %v", rn.Name(), err)
		}
	}

	bootstrapCfg := envoy.DefaultBootstrapConfig(xdsServerPort)
	envoyCfg := envoy.DefaultEnvoyConfig()

	if err := envoy.WriteBootstrapConfig(envoyCfg.BootstrapPath, bootstrapCfg); err != nil {
		klog.Fatalf("write envoy bootstrap failed: %v", err)
	}

	envoyExited, err := envoy.StartEnvoy(ctx, cancel, envoyCfg)
	if err != nil {
		klog.Fatalf("start envoy failed: %v", err)
	}

	klog.Info("starting manager...")
	if err := mgr.Start(ctx); err != nil {
		klog.Fatalf("manager exited: %v", err)
	}
	klog.Info("manager stopped, waiting for envoy to drain...")
	// Block until Envoy fully exits so the graceful drain sequence
	// (SIGTERM → envoyDrainTimeout → SIGKILL) has time to complete
	// before the Go process itself returns.
	<-envoyExited
	klog.Info("envoy exited, shutting down")
}

// cacheReadyCheck returns a healthz.Checker that verifies the manager's cache has synced.
// This ensures the control plane has populated its cache with all resources from the API server
// before reporting ready. This prevents serving inconsistent xDS configuration to Envoy proxies
// when running multiple control plane replicas during periods of resource churn.
func cacheReadyCheck(mgr manager.Manager) healthz.Checker {
	return func(req *http.Request) error {
		// Use a short timeout to avoid blocking the health check indefinitely.
		// The readiness probe will retry periodically until the cache syncs.
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()

		// WaitForCacheSync returns true if the cache has synced, false if the context is cancelled.
		if !mgr.GetCache().WaitForCacheSync(ctx) {
			return fmt.Errorf("cache not synced yet")
		}

		return nil
	}
}

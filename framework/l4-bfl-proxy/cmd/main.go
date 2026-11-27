package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
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
	metricsAddr         = "0"
	probeAddr           = ":18081"
)

func main() {
	klog.InitFlags(nil)
	xdsOpts := xdstranslator.DefaultOptions()
	flag.DurationVar(&xdsOpts.TCPIdleTimeout, "xds-tcp-idle-timeout", xdsOpts.TCPIdleTimeout, "tcp proxy idle timeout")
	flag.DurationVar(&xdsOpts.HTTPStreamIdleTimeout, "xds-http-stream-idle-timeout", xdsOpts.HTTPStreamIdleTimeout, "http stream idle timeout")
	flag.DurationVar(&xdsOpts.ConnectTimeout, "xds-connect-timeout", xdsOpts.ConnectTimeout, "upstream connect timeout")
	flag.DurationVar(&xdsOpts.RouteTimeout, "xds-route-timeout", xdsOpts.RouteTimeout, "absolute per-request route timeout; 0 disables it so slow LLM upstreams (long TTFT/decode) are not cut off")
	flag.DurationVar(&xdsOpts.RouteIdleTimeout, "xds-route-idle-timeout", xdsOpts.RouteIdleTimeout, "per-request idle timeout (RouteAction.idle_timeout): reset a request when no bytes flow in either direction for this long; 0 inherits -xds-http-stream-idle-timeout")
	flag.DurationVar(&xdsOpts.ClusterIdleTimeout, "xds-cluster-idle-timeout", xdsOpts.ClusterIdleTimeout, "upstream HTTP connection idle timeout (CommonHttpProtocolOptions.IdleTimeout); set below the backend keep-alive limit to avoid stale-connection resets")
	flag.DurationVar(&xdsOpts.TCPKeepaliveTime, "xds-tcp-keepalive-time", xdsOpts.TCPKeepaliveTime, "idle time before the first TCP keepalive probe on downstream and upstream connections")
	flag.DurationVar(&xdsOpts.TCPKeepaliveInterval, "xds-tcp-keepalive-interval", xdsOpts.TCPKeepaliveInterval, "interval between TCP keepalive probes")
	tcpKeepaliveProbes := flag.Uint("xds-tcp-keepalive-probes", uint(xdsOpts.TCPKeepaliveProbes), "unanswered TCP keepalive probes before the connection is dropped")
	httpsBufferLimit := flag.Uint("xds-https-buffer-limit-bytes", uint(xdsOpts.HTTPSBufferLimitBytes), "HTTPS listener per-connection buffer limit; HTML responses rewritten by the Lua filter must fit in it")

	flag.Parse()
	defer klog.Flush()
	xdsOpts.TCPKeepaliveProbes = uint32(*tcpKeepaliveProbes)
	xdsOpts.HTTPSBufferLimitBytes = uint32(*httpsBufferLimit)
	xdstranslator.SetOptions(xdsOpts)

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
	if err := sysv1alpha1.AddToScheme(scheme); err != nil {
		klog.Fatalf("add sysv1alpha1 scheme failed: %v", err)
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

	xdsSrv := server.New(xdsResources, &server.Config{
		Address: "127.0.0.1",
		Port:    xdsServerPort,
	})

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Fatalf("set up health check failed: %v", err)
	}
	if err := mgr.AddReadyzCheck("cache-synced", cacheReadyCheck(mgr)); err != nil {
		klog.Fatalf("set up ready check failed: %v", err)
	}
	// Fail readiness when Envoy rejects the pushed xDS config (e.g. a bad
	// listener that leaves ports 443/444 not listening).
	//if err := mgr.AddReadyzCheck("envoy-config", xdsSrv.ReadyCheck); err != nil {
	//	klog.Fatalf("set up envoy-config ready check failed: %v", err)
	//}

	runners := []runner.Runner{
		prov,
		fsReconciler,
		translator.New(providerResources, xdsIR, &translator.Config{
			SSLServerPort:      sslServerPort,
			SSLProxyServerPort: sslProxyServerPort,
		}),
		xdstranslator.New(xdsIR, xdsResources),
		xdsSrv,
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

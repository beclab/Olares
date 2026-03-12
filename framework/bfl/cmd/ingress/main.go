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

	"bytetrade.io/web3os/bfl/internal/ingress/envoy"
	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"bytetrade.io/web3os/bfl/internal/ingress/provider"
	"bytetrade.io/web3os/bfl/internal/ingress/runner"
	"bytetrade.io/web3os/bfl/internal/ingress/translator"
	"bytetrade.io/web3os/bfl/internal/ingress/xds/server"
	xdstranslator "bytetrade.io/web3os/bfl/internal/ingress/xds/translator"
	"bytetrade.io/web3os/bfl/pkg/constants"
	v1alpha1App "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	xdsServerPort = 8794
	resyncPeriod  = 60 * time.Minute
	metricsAddr   = ":9001"
	probeAddr     = ":8081"
)

var (
	scheme = runtime.NewScheme()

	user           string
	bflServiceName string
	namespace      string
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1App.AddToScheme(scheme)
	_ = iamV1alpha2.AddToScheme(scheme)
}

func flags() error {
	flag.StringVar(&user, "user", "", "The bfl ingress owner username")
	flag.StringVar(&bflServiceName, "bfl-svc", "bfl", "BFL service name")
	flag.StringVar(&namespace, "namespace", "", "BFL owner namespace")
	tcpIdleTimeout := flag.Duration("xds-tcp-idle-timeout", time.Hour, "tcp proxy idle timeout")
	httpStreamIdleTimeout := flag.Duration("xds-http-stream-idle-timeout", 30*time.Minute, "http stream idle timeout")
	connectTimeout := flag.Duration("xds-connect-timeout", 5*time.Second, "upstream connect timeout")
	routeTimeout := flag.Duration("xds-route-timeout", 5*time.Minute, "route timeout")

	klog.InitFlags(nil)
	flag.Parse()
	xdstranslator.SetTimeouts(*tcpIdleTimeout, *httpStreamIdleTimeout, *connectTimeout, *routeTimeout)

	if namespace == "" {
		return fmt.Errorf("missing flag 'namespace'")
	}
	klog.Infof("flags:namespace: %s", namespace)
	if user == "" {
		return fmt.Errorf("missing flag 'user'")
	}

	constants.Username = user
	constants.Namespace = namespace
	constants.BFLServiceName = bflServiceName

	return nil
}

func main() {
	if err := flags(); err != nil {
		fmt.Fprintf(os.Stderr, "flag error: %v\n", err)
		os.Exit(1)
	}
	defer klog.Flush()

	ctrl.SetLogger(klog.NewKlogr())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	syncPeriod := resyncPeriod
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		Cache:                  cache.Options{SyncPeriod: &syncPeriod},
	})
	if err != nil {
		klog.Fatalf("create manager failed: %v", err)
	}

	providerResources := &message.ProviderResources{}
	xdsIR := &message.XdsIR{}
	xdsResources := &message.XdsResources{}

	prov := provider.New(mgr.GetCache(), providerResources, &provider.Config{
		Username:  constants.Username,
		Namespace: constants.Namespace,
	})
	if err := prov.SetupWithManager(ctx); err != nil {
		klog.Fatalf("setup provider failed: %v", err)
	}

	fsReconciler := provider.NewFileserverReconciler(mgr.GetCache(), mgr.GetClient())
	fsReconciler.OnReconciled = prov.NotifyChanged
	if err := fsReconciler.SetupWithManager(ctx); err != nil {
		klog.Fatalf("setup fileserver reconciler failed: %v", err)
	}
	svcReconciler := provider.NewBflSvcReconciler(mgr.GetCache(), mgr.GetClient())
	svcReconciler.OnReconciled = prov.NotifyChanged
	if err := svcReconciler.SetupWithManager(ctx); err != nil {
		klog.Fatalf("setup bflsvc reconciler failed: %v", err)
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
		svcReconciler,
		translator.New(providerResources, xdsIR, &translator.Config{
			AutheliaURL: os.Getenv("AUTHELIA_AUTH_URL"),
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

	if err := envoy.StartEnvoy(ctx, cancel, envoyCfg); err != nil {
		klog.Fatalf("start envoy failed: %v", err)
	}

	klog.Info("starting manager with envoy xDS pipeline...")
	klog.Infof("user=%s, namespace=%s, xds-port=%d", constants.Username, constants.Namespace, xdsServerPort)
	if err := mgr.Start(ctx); err != nil {
		klog.Fatalf("manager exited: %v", err)
	}
	klog.Info("manager stopped")
}

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

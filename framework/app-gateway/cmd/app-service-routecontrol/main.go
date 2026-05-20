// Command app-service-routecontrol turns SharedRouteRegistry CRs into HTTPRoute +
// NetworkPolicy objects in the same namespace as the backend Service
// (F-2, F-4). It is the only writer of HTTPRoute and the
// app-gateway-shared-ingress-np NetworkPolicy.
package main

import (
	"flag"
	"os"

	"go.uber.org/zap/zapcore"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gwapi "github.com/beclab/Olares/framework/app-gateway/pkg/api/v1alpha1"
	"github.com/beclab/Olares/framework/app-gateway/pkg/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(gwapi.AddToScheme(scheme))
}

func main() {
	var metricsAddr, probeAddr, gwNS, gwName, gwSection string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8090", "Address for the metrics endpoint.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8091", "Address for the health probe endpoint.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for the controller manager.")
	flag.StringVar(&gwNS, "gateway-namespace", "app-gateway", "Namespace of the parent Gateway referenced by HTTPRoute.")
	flag.StringVar(&gwName, "gateway-name", "app-gateway", "Name of the parent Gateway referenced by HTTPRoute.")
	flag.StringVar(&gwSection, "gateway-section", "http", "Listener sectionName on the parent Gateway.")
	opts := zap.Options{Development: true, TimeEncoder: zapcore.RFC3339TimeEncoder}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "app-service-routecontrol.app-gateway.olares.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controller.SRRReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		GatewayNamespace:  gwNS,
		GatewayName:       gwName,
		GatewaySectionRef: gwSection,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up SharedRouteRegistry controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "gateway", gwNS+"/"+gwName, "section", gwSection)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

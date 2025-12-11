package main

import (
	"flag"
	"fmt"
	"os"

	v1alpha1App "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/internal/ingress/controllers"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var (
	user                 string
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	enableNginx          bool
	bflServiceName       string
	namespace            string
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1App.AddToScheme(scheme)

	_ = iamV1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func flags() error {
	flag.StringVar(&user, "user", "", "The bfl ingress owner username")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8081", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-addr", ":8082", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableNginx, "enable-nginx", false, "Run nginx process")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&bflServiceName, "bfl-svc", "bfl", "BFL service name")
	flag.StringVar(&namespace, "namespace", "", "BFL owner namespace")

	flag.Parse()

	// required
	if namespace == "" {
		if namespace = utils.Namespace(); namespace == "" {
			return fmt.Errorf("namespace not found")
		}
	}
	if user == "" {
		return fmt.Errorf("missing flag 'user'")
	}

	constants.Username = user
	constants.Namespace = namespace
	constants.BFLServiceName = bflServiceName

	setupLog.Info("bfl-ingress flags",
		"metrics-addr", metricsAddr,
		"health-probe-addr", probeAddr,
		"enableNginx", enableNginx,
		"enableLeaderElection", enableLeaderElection,
		"username", constants.Username,
		"namespace", constants.Namespace,
		"name-ssl-configmap", constants.NameSSLConfigMapName,
		"bfl-service", constants.BFLServiceName)

	return nil
}

func main() {
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	if err := flags(); err != nil {
		setupLog.Error(err, "flag error")
		os.Exit(1)
	}

	// controller mgr
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "6001a005.bytetrade.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to setup health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to setup ready check")
		os.Exit(1)
	}

	// nginx controller
	ngx := controllers.NginxController{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Log:     ctrl.Log.WithName("controllers").WithName("NginxController"),
		Eventer: mgr.GetEventRecorderFor("nginx-controller"),
	}

	if err = ngx.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller NginxController")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if enableNginx {
		setupLog.Info("Starting nginx process, using default nginx.conf")
		err = ngx.RunNginx()
		if err != nil {
			setupLog.Error(err, "unable to run nginx process")
			os.Exit(1)
		}
	}

	setupLog.Info("Starting manager and nginx controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

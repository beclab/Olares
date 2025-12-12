package main

import (
	"os"

	controllers "bytetrade.io/web3os/osnode-init/pkg/controller"
	"bytetrade.io/web3os/osnode-init/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logLevel string

	metricsAddr string

	enableLeaderElection bool

	probeAddr string

	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	//+kubebuilder:scaffold:scheme
}

func run() error {
	c, err := ctrl.GetConfig()
	if err != nil {
		return errors.WithStack(err)
	}

	mgr, err := ctrl.NewManager(c, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "77yco38a.bytetrade.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return errors.Errorf("new manager: %v", err)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return errors.Errorf("unable to set up health check: %v", err)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return errors.Errorf("unable to setup ready check: %v", err)
	}

	if err = controllers.NewNodeInitController(mgr.GetClient(), mgr.GetScheme(), c).
		SetupWithManager(mgr); err != nil {
		return errors.Errorf("unable to create nodeInitController: %v", err)
	}

	// start system env watcher to update runtime constants
	if err = mgr.Add(controllers.NewSystemEnvWatcher(mgr.GetConfig())); err != nil {
		return errors.Errorf("unable to start system env watcher: %v", err)
	}

	log.Info("starting manager")

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return errors.Errorf("start manager: %v", err)
	}

	return nil
}

func main() {
	pflag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	pflag.StringVarP(&logLevel, "log-level", "l", "debug", "log level")
	pflag.Parse()

	log.InitLog(logLevel)

	hostIP := os.Getenv("NODE_IP")
	if hostIP == "" {
		log.Errorf("no env 'NODE_IP' provided")
		os.Exit(1)
	}
	controllers.NodeIP = hostIP

	if err := run(); err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
}

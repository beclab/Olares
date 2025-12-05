package main

import (
	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/controllers"
	"context"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"syscall"
)

var (
	scheme   = runtime.NewScheme()
	imageLog = ctrl.Log.WithName("image")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appv1alpha1.AddToScheme(scheme))
	utilruntime.Must(sysv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func main() {
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	config := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: ":7081",
		LeaderElection:         false,
	})
	if err != nil {
		imageLog.Error(err, "Unable to start image manager")
		os.Exit(1)
	}

	if err = (&controllers.ImageManagerController{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		imageLog.Error(err, "Unable to create image controller")
		os.Exit(1)
	}
	if err = (&controllers.AppImageInfoController{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		imageLog.Error(err, "Unable to create app image controller")
		os.Exit(1)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		imageLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		imageLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	ictx, cancelFunc := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		select {
		case <-c:
			cancelFunc()
			<-c
			os.Exit(1) // second signal. Exit directly.

		}
	}()

	if err = mgr.Start(ictx); err != nil {
		cancelFunc()
		imageLog.Error(err, "Unable to running image manager")
		os.Exit(1)
	}
	cancelFunc()
}

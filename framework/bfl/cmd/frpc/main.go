package main

import (
	"bytetrade.io/web3os/bfl/pkg/constants"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"bytetrade.io/web3os/bfl/internal/frpc/controllers"
	"bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	v1alpha2 "bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var (
	server     string
	port       int
	authMethod string
	username   string
	authToken  string

	defaultPort int = 7000
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func flags() error {
	flag.StringVar(&server, v1alpha2.FRPOptionServer, "", "The fprc server host")
	flag.IntVar(&port, v1alpha2.FRPOptionPort, defaultPort, "The frp-server port")
	flag.StringVar(&authMethod, v1alpha2.FRPOptionAuthMethod, "", "The frp auth method")
	flag.StringVar(&username, v1alpha2.FRPOptionUserName, "", "The olares user's username")
	flag.StringVar(&authToken, v1alpha2.FRPOptionAuthToken, "", "The token, if auth method is token")
	flag.Parse()

	if server == "" {
		return fmt.Errorf("missing flag 'server'")
	}

	if strings.HasPrefix(server, "http") {
		if serverURL, err := url.Parse(server); err != nil {
			return fmt.Errorf("invalid server url: %v", err)
		} else {
			server = serverURL.Host
		}
	}

	if username == "" {
		return errors.New("missing flag 'username'")
	}
	constants.Username = username

	if authMethod == v1alpha2.FRPAuthMethodToken && authToken == "" {
		return errors.New("auth method is selected as token but no token is provided")
	}

	return nil
}

func main() {
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	if err := flags(); err != nil {
		setupLog.Error(err, "invalid options")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Port:   9443,
	})
	if err != nil {
		setupLog.Error(err, "failed to start manager")
		os.Exit(1)
	}

	frpController := controllers.FrpcController{
		Client: mgr.GetClient(),
		Config: &controllers.FRPCConfig{
			Server:     server,
			Port:       port,
			AuthMethod: authMethod,
			UserName:   username,
			AuthToken:  authToken,
		},
		Log:            ctrl.Log.WithName("controllers").WithName("FrpcController"),
		ReconcileQueue: make(chan string, controllers.QueueSize),
	}

	if err = frpController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller FrpcController")
		os.Exit(1)
	}

	globalContext := ctrl.SetupSignalHandler()

	if err = frpController.RunFrpc(globalContext); err != nil {
		setupLog.Error(err, "unable to run frpc process")
		os.Exit(1)
	}

	if err = mgr.Start(globalContext); err != nil {
		setupLog.Error(err, "unable to start frpc")
		os.Exit(1)
	}
}

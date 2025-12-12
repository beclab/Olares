package apiserver

import (
	"context"
	"net/http"
	"time"

	"bytetrade.io/web3os/system-server/pkg/constants"
	sysclientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"
	permission "bytetrade.io/web3os/system-server/pkg/permission/v1alpha1"
	permissionv2alpha1 "bytetrade.io/web3os/system-server/pkg/permission/v2alpha1"
	providerv2alpha1 "bytetrade.io/web3os/system-server/pkg/providerregistry/v2alpha1"
	proxyv2alpha1 "bytetrade.io/web3os/system-server/pkg/serviceproxy/v2alpha1"

	"github.com/emicklei/go-restful/v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// APIServer represents an API server for system.
type APIServer struct {
	Server   *http.Server
	preStart func()

	// RESTful Server
	container *restful.Container

	serverCtx context.Context
}

// New constructs a new APIServer.
func New(ctx context.Context) (*APIServer, error) {
	server := &http.Server{
		Addr: constants.APIServerListenAddress,
	}

	return &APIServer{
		Server:    server,
		container: restful.NewContainer(),
		serverCtx: ctx,
	}, nil
}

// PrepareRun do prepares for API server.
func (s *APIServer) PrepareRun(
	kubeconfig *rest.Config,
	sysclientset *sysclientset.Clientset,
	permissionLister v1alpha1.ApplicationPermissionLister,
	providerLister v1alpha1.ProviderRegistryLister,
) error {

	proxyCfg := proxyv2alpha1.ServerOptions(constants.ProxyServerListenAddress)
	proxy := proxyv2alpha1.NewRBACProxyServer(s.serverCtx)
	if err := proxy.Init(proxyCfg); err != nil {
		klog.Errorf("failed to initialize proxy: %v", err)
		return err
	}

	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	// registry := prodiverregistry.NewRegistry(sysclientset, providerLister)
	ctrlSet := permission.PermissionControlSet{
		Ctrl: permission.NewPermissionControl(sysclientset, permissionLister),
		Mgr:  permission.NewAccessManager(),
	}

	// use the server context for goroutine in background
	utilruntime.Must(permission.AddPermissionControlToContainer(s.container, &ctrlSet, kubeconfig))
	utilruntime.Must(permissionv2alpha1.AddPermissionControlToContainer(s.container, permissionv2alpha1.Auth(proxy.Authenticator()), kubeconfig))
	utilruntime.Must(providerv2alpha1.AddProviderRegistryToContainer(s.container, permissionv2alpha1.Auth(proxy.Authenticator()), kubeconfig))
	s.Server.Handler = s.container

	s.preStart = func() {
		go func() {
			utilruntime.Must(proxy.Start(proxyCfg))
		}()
	}

	return nil
}

// Run running a server.
func (s *APIServer) Run() error {
	shutdownCtx, cancel := context.WithTimeout(s.serverCtx, 2*time.Minute)
	defer func() {
		cancel()
	}()

	go func() {
		<-s.serverCtx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
		klog.Info("shutdown apiserver for system-server")
	}()

	if s.preStart != nil {
		s.preStart()
	}

	klog.Info("starting apiserver for system-server,", "listen on ", constants.APIServerListenAddress)
	return s.Server.ListenAndServe()
}

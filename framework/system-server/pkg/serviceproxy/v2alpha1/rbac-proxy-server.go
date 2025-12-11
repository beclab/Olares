package v2alpha1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	permv2alpha1 "bytetrade.io/web3os/system-server/pkg/permission/v2alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"
	"github.com/brancz/kube-rbac-proxy/cmd/kube-rbac-proxy/app/options"
	"github.com/brancz/kube-rbac-proxy/pkg/authn"
	"github.com/brancz/kube-rbac-proxy/pkg/authz"
	"github.com/brancz/kube-rbac-proxy/pkg/filters"
	"github.com/brancz/kube-rbac-proxy/pkg/proxy"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ middleware.ProxyBalancer = (*server)(nil)
var serviceKey = "Provider-Service"

type server struct {
	proxy         *echo.Echo
	mainCtx       context.Context
	authenticator authenticator.Request
	authorizer    permv2alpha1.Authorizer
}

// AddTarget implements middleware.ProxyBalancer.
func (s *server) AddTarget(*middleware.ProxyTarget) bool {
	return true
}

// Next implements middleware.ProxyBalancer.
func (s *server) Next(ctx echo.Context) *middleware.ProxyTarget {
	if service := ctx.Get(serviceKey); service != nil {
		if svcStr, ok := service.(string); ok {
			klog.V(5).Infof("RBAC: using provider service %q", svcStr)
			var proxyPassStr string
			// if the service string is like "http://service.namespace.svc:port" or "https://service.namespace.svc:port"
			// we use it directly, otherwise we add "http://" prefix
			if strings.HasPrefix(svcStr, "http://") || strings.HasPrefix(svcStr, "https://") {
				proxyPassStr = svcStr
			} else {
				// otherwise we assume it is like "service.namespace.svc:port"
				proxyPassStr = fmt.Sprintf("http://%s", svcStr)
			}

			proxyPass, err := url.Parse(proxyPassStr)
			if err != nil {
				klog.Errorf("failed to parse URL %s: %v", proxyPassStr, err)
				return nil
			}
			return &middleware.ProxyTarget{URL: proxyPass}
		}
	}

	klog.V(5).Info("RBAC: no provider service found in context")
	return nil
}

// RemoveTarget implements middleware.ProxyBalancer.
func (s *server) RemoveTarget(string) bool {
	return true
}

func NewRBACProxyServer(ctx context.Context) *server {
	proxy := echo.New()
	proxy.Use(middleware.Recover())
	proxy.Use(middleware.Logger())

	s := &server{
		mainCtx: ctx,
		proxy:   proxy,
	}

	return s
}

func (s *server) Init(cfg *completedProxyRunOptions) error {
	var err error
	authnCfg := &permv2alpha1.AuthnConfig{
		AuthnConfig: *cfg.auth.Authentication,
		LLDAP: permv2alpha1.LLDAPConfig{
			Server: cfg.lldapServer,
			Port:   cfg.lldapPort,
		},
	}

	s.authenticator, err = permv2alpha1.UnionAllAuthenticators(s.mainCtx, authnCfg, cfg.kubeClient)
	if err != nil {
		klog.Errorf("failed to create authenticator: %v", err)
		return err
	}

	s.authorizer, err = permv2alpha1.UnionAllAuthorizers(s.mainCtx, cfg.auth.Authorization, cfg.kubeClient, cfg.informerFactory)
	if err != nil {
		klog.Errorf("failed to create authorizer: %v", err)
		return err
	}

	s.proxy.Use(s.rbac(cfg))

	config := middleware.DefaultProxyConfig
	config.Balancer = s

	transport, err := initTransport(cfg.upstreamCABundle, cfg.tls.UpstreamClientCertFile, cfg.tls.UpstreamClientKeyFile)
	if err != nil {
		klog.Error(err)
		return err
	}

	config.Transport = transport

	// proxy for http and https
	config.Skipper = func(c echo.Context) bool {
		return c.IsWebSocket() && !c.IsTLS()
	}
	s.proxy.Use(middleware.ProxyWithConfig(config))

	// proxy for websocket
	websocketConfig := config
	websocketTransport, err := initTransport(cfg.upstreamCABundle, cfg.tls.UpstreamClientCertFile, cfg.tls.UpstreamClientKeyFile)
	if err != nil {
		klog.Error(err)
		return err
	}
	websocketTransport.(*http.Transport).TLSClientConfig = nil

	websocketConfig.Skipper = nil
	websocketConfig.Transport = websocketTransport
	s.proxy.Use(middleware.ProxyWithConfig(websocketConfig))

	cfg.informerFactory.Start(s.mainCtx.Done())
	go func() {
		<-s.mainCtx.Done()
		cfg.informerFactory.Shutdown()
	}()
	return nil
}

func (s *server) Start(cfg *completedProxyRunOptions) error {
	klog.Info("starting proxy server for system-server,", "listen on ", cfg.insecureListenAddress)
	return s.proxy.Start(cfg.insecureListenAddress)
}

func (s *server) rbac(cfg *completedProxyRunOptions) func(next echo.HandlerFunc) echo.HandlerFunc {
	namespaceLister := cfg.informerFactory.Core().V1().Namespaces().Lister()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			handlerFunc := func(rw http.ResponseWriter, req *http.Request) {
				// here the request maybe is a copy of the request in echo context
				// so we must get the value and put it into the context
				if service, ok := permv2alpha1.ProviderServiceFrom(req.Context()); ok {
					c.Set(serviceKey, service)
				}

				err = next(c)
			}

			handlerFunc = permv2alpha1.RecoverHeader(handlerFunc)
			handlerFunc = permv2alpha1.MustHaveProviderService(handlerFunc)
			handlerFunc = permv2alpha1.WithUserHeader(permv2alpha1.UserFromServiceAccount(namespaceLister), handlerFunc)
			handlerFunc = filters.WithAuthHeaders(cfg.auth.Authentication.Header, handlerFunc)
			handlerFunc = permv2alpha1.WithAuthorization(s.authorizer, cfg.auth.Authorization, handlerFunc)
			handlerFunc = permv2alpha1.WithAuthentication(s.authenticator, cfg.auth.Authentication.Token.Audiences, handlerFunc)

			handlerFunc(c.Response(), c.Request())
			return err
		}
	}
}

func (s *server) Authenticator() authenticator.Request {
	return s.authenticator
}

func ServerOptions(listenAddress string) *completedProxyRunOptions {
	completed := &completedProxyRunOptions{
		insecureListenAddress: listenAddress,
		secureListenAddress:   "", // TODO: implement secure listen address
		proxyEndpointsPort:    0,
		upstreamForceH2C:      false,

		tls: &options.TLSConfig{},
		auth: &proxy.Config{
			Authentication: &authn.AuthnConfig{
				X509:   &authn.X509Config{},
				Header: &authn.AuthnHeaderConfig{},
				OIDC:   &authn.OIDCConfig{},
				Token:  &authn.TokenConfig{},
			},
			Authorization: &authz.Config{},
		},
	}

	var err error

	config := ctrl.GetConfigOrDie()
	completed.kubeClient = kubernetes.NewForConfigOrDie(config)

	completed.lldapServer = utils.GetEnvOrDefault("LLDAP_SERVER", "lldap-service.os-platform")
	completed.lldapPort, err = strconv.Atoi(utils.GetEnvOrDefault("LLDAP_PORT", "17170"))
	if err != nil {
		klog.Errorf("failed to parse LLDAP_PORT: %v", err)
		completed.lldapPort = 17170 // default value
	}

	completed.informerFactory = informers.NewSharedInformerFactory(completed.kubeClient, 0)

	return completed
}

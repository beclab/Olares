package v2alpha1

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	permv2alpha1 "bytetrade.io/web3os/system-server/pkg/permission/v2alpha1"
	"github.com/ghodss/yaml"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	k8sapiflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"

	"github.com/brancz/kube-rbac-proxy/cmd/kube-rbac-proxy/app"
	"github.com/brancz/kube-rbac-proxy/cmd/kube-rbac-proxy/app/options"
	"github.com/brancz/kube-rbac-proxy/pkg/authz"
	"github.com/brancz/kube-rbac-proxy/pkg/filters"
	rbac_proxy_tls "github.com/brancz/kube-rbac-proxy/pkg/tls"
)

func NewKubeRBACProxyCommand() *cobra.Command {
	var lldapServer string
	var lldapPort int

	o := options.NewProxyRunOptions()
	cmd := &cobra.Command{
		Use: "kube-rbac-proxy",
		Long: `The kube-rbac-proxy is a small HTTP proxy for a single upstream
that can perform RBAC authorization against the Kubernetes API using SubjectAccessReview.`,

		// stop printing usage when the command errors
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			klog.SetLogFilter(&app.SanitizingFilter{})

			// silence client-go warnings.
			// kube-apiserver loopback clients should not log self-issued warnings.
			rest.SetDefaultWarningHandler(rest.NoWarnings{})
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()

			fs := cmd.Flags()

			k8sapiflag.PrintFlags(fs)

			if err := o.Validate(); err != nil {
				return err
			}

			// set default options
			completedOptions, err := Complete(o)
			if err != nil {
				return err
			}

			completedOptions.lldapServer = lldapServer
			completedOptions.lldapPort = lldapPort

			return Run(completedOptions)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	fs := cmd.Flags()
	namedFlagSets := o.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name(), logs.SkipLoggingConfigurationFlags())

	// add lldap options
	lldapFlagSet := namedFlagSets.FlagSet("lldap")
	lldapFlagSet.StringVar(&lldapServer, "lldap-server", "lldap-service.os-platform", "The lldap server's service address.")
	lldapFlagSet.IntVar(&lldapPort, "lldap-port", 17170, "The lldap server's service port.")

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	k8sapiflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	return cmd
}

type configfile struct {
	AuthorizationConfig *authz.Config `json:"authorization,omitempty"`
}

func Complete(o *options.ProxyRunOptions) (*completedProxyRunOptions, error) {
	var err error
	completed := &completedProxyRunOptions{
		insecureListenAddress: o.InsecureListenAddress,
		secureListenAddress:   o.SecureListenAddress,
		proxyEndpointsPort:    o.ProxyEndpointsPort,
		upstreamForceH2C:      o.UpstreamForceH2C,

		allowPaths:  o.AllowPaths,
		ignorePaths: o.IgnorePaths,
	}

	completed.upstreamURL, err = url.Parse(o.Upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream URL: %w", err)
	}

	if upstreamCAPath := o.UpstreamCAFile; len(upstreamCAPath) > 0 {
		upstreamCAPEM, err := os.ReadFile(upstreamCAPath)
		if err != nil {
			return nil, err
		}

		upstreamCACertPool := x509.NewCertPool()
		if ok := upstreamCACertPool.AppendCertsFromPEM(upstreamCAPEM); !ok {
			return nil, errors.New("error parsing upstream CA certificate")
		}
		completed.upstreamCABundle = upstreamCACertPool
	}

	completed.auth = o.Auth
	completed.tls = o.TLS

	if configFileName := o.ConfigFileName; len(configFileName) > 0 {
		completed.auth.Authorization, err = parseAuthorizationConfigFile(configFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read the config file: %w", err)
		}
	}

	kubeconfig, err := initKubeConfig(o.KubeconfigLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if o.QPS > 0 {
		kubeconfig.QPS = o.QPS
	}
	if o.Burst > 0 {
		kubeconfig.Burst = o.Burst
	}

	completed.kubeClient, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Kubernetes client: %w", err)
	}

	completed.http2Disable = o.HTTP2Disable
	completed.http2Options = &http2.Server{
		IdleTimeout:                  90 * time.Second,
		MaxConcurrentStreams:         o.HTTP2MaxConcurrentStreams,
		MaxReadFrameSize:             o.HTTP2MaxSize,
		MaxUploadBufferPerStream:     int32(o.HTTP2MaxSize),
		MaxUploadBufferPerConnection: int32(o.HTTP2MaxSize) * int32(o.HTTP2MaxConcurrentStreams),
	}

	completed.informerFactory = informers.NewSharedInformerFactory(completed.kubeClient, 0)

	return completed, nil
}

func Run(cfg *completedProxyRunOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authnCfg := &permv2alpha1.AuthnConfig{
		AuthnConfig: *cfg.auth.Authentication,
		LLDAP: permv2alpha1.LLDAPConfig{
			Server: cfg.lldapServer,
			Port:   cfg.lldapPort,
		},
	}
	authenticator, err := permv2alpha1.UnionAllAuthenticators(ctx, authnCfg, cfg.kubeClient)
	if err != nil {
		klog.Errorf("failed to create authenticator: %v", err)
		return err
	}

	authorizer, err := permv2alpha1.UnionAllAuthorizers(ctx, cfg.auth.Authorization, cfg.kubeClient, cfg.informerFactory)
	if err != nil {
		klog.Errorf("failed to create authorizer: %v", err)
		return err
	}

	upstreamTransport, err := initTransport(cfg.upstreamCABundle, cfg.tls.UpstreamClientCertFile, cfg.tls.UpstreamClientKeyFile)
	if err != nil {
		return fmt.Errorf("failed to set up upstream TLS connection: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(cfg.upstreamURL)
	proxy.Transport = upstreamTransport

	if cfg.upstreamForceH2C {
		// Force http/2 for connections to the upstream i.e. do not start with HTTP1.1 UPGRADE req to
		// initialize http/2 session.
		// See https://github.com/golang/go/issues/14141#issuecomment-219212895 for more context
		proxy.Transport = &http2.Transport{
			// Allow http schema. This doesn't automatically disable TLS
			AllowHTTP: true,
			// Do disable TLS.
			// In combination with the schema check above. We could enforce h2c against the upstream server
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		}
	}

	namespaceLister := cfg.informerFactory.Core().V1().Namespaces().Lister()

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ignorePathFound := false
		for _, pathIgnored := range cfg.ignorePaths {
			ignorePathFound, err = path.Match(pathIgnored, req.URL.Path)
			if err != nil {
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			if ignorePathFound {
				break
			}
		}

		if !ignorePathFound {
			handlerFunc := proxy.ServeHTTP
			handlerFunc = permv2alpha1.WithUserHeader(permv2alpha1.UserFromServiceAccount(namespaceLister), handlerFunc)
			handlerFunc = filters.WithAuthHeaders(cfg.auth.Authentication.Header, handlerFunc)
			handlerFunc = permv2alpha1.WithAuthorization(authorizer, cfg.auth.Authorization, handlerFunc)
			handlerFunc = permv2alpha1.WithAuthentication(authenticator, cfg.auth.Authentication.Token.Audiences, handlerFunc)
			handlerFunc(w, req)

			return
		}

		proxy.ServeHTTP(w, req)
	})
	handler = filters.WithAllowPaths(cfg.allowPaths, handler)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	var gr run.Group
	{
		if cfg.secureListenAddress != "" {
			srv := &http.Server{
				Handler:   mux,
				TLSConfig: &tls.Config{},
			}

			if cfg.tls.CertFile == "" && cfg.tls.KeyFile == "" {
				klog.Info("Generating self signed cert as no cert is provided")
				host, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("failed to retrieve hostname for self-signed cert: %w", err)
				}
				certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(host, nil, nil)
				if err != nil {
					return fmt.Errorf("failed to generate self signed cert and key: %w", err)
				}
				cert, err := tls.X509KeyPair(certBytes, keyBytes)
				if err != nil {
					return fmt.Errorf("failed to load generated self signed cert and key: %w", err)
				}

				srv.TLSConfig.Certificates = []tls.Certificate{cert}
			} else {
				klog.Info("Reading certificate files")
				r, err := rbac_proxy_tls.NewCertReloader(cfg.tls.CertFile, cfg.tls.KeyFile, cfg.tls.ReloadInterval)
				if err != nil {
					return fmt.Errorf("failed to initialize certificate reloader: %w", err)
				}

				srv.TLSConfig.GetCertificate = r.GetCertificate

				ctx, cancel := context.WithCancel(context.Background())
				gr.Add(func() error {
					return r.Watch(ctx)
				}, func(error) {
					cancel()
				})
			}

			version, err := k8sapiflag.TLSVersion(cfg.tls.MinVersion)
			if err != nil {
				return fmt.Errorf("TLS version invalid: %w", err)
			}

			cipherSuiteIDs, err := k8sapiflag.TLSCipherSuites(cfg.tls.CipherSuites)
			if err != nil {
				return fmt.Errorf("failed to convert TLS cipher suite name to ID: %w", err)
			}

			srv.TLSConfig.CipherSuites = cipherSuiteIDs
			srv.TLSConfig.MinVersion = version
			srv.TLSConfig.ClientAuth = tls.RequestClientCert

			if cfg.http2Disable {
				// HTTP/2 is temporarily disabled due to CVE-2023-44487
				// Programs that must disable HTTP/2 can do so by setting
				// Transport.TLSNextProto (for clients) or Server.TLSNextProto
				// (for servers) to a non-nil, empty map.
				// https://pkg.go.dev/net/http
				srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

				// For reference:
				// https://github.com/kubernetes/kubernetes/blob/de054fbf9422d778568946de21a48c7330a6c1b7/staging/src/k8s.io/apiserver/pkg/server/secure_serving.go#L55-L59
				srv.TLSConfig.NextProtos = []string{"http/1.1"}
			} else {
				if err := http2.ConfigureServer(srv, cfg.http2Options); err != nil {
					return fmt.Errorf("failed to configure http2 server: %w", err)
				}
			}

			gr.Add(func() error {
				klog.Infof("Starting TCP socket on %v", cfg.secureListenAddress)
				l, err := net.Listen("tcp", cfg.secureListenAddress)
				if err != nil {
					return fmt.Errorf("failed to listen on secure address: %w", err)
				}
				defer l.Close()

				klog.Infof("Listening securely on %v", cfg.secureListenAddress)
				tlsListener := tls.NewListener(l, srv.TLSConfig)
				return srv.Serve(tlsListener)
			}, func(err error) {
				if err := srv.Shutdown(context.Background()); err != nil {
					klog.Errorf("failed to gracefully shutdown server: %+v", err)
				}
			})

			if cfg.proxyEndpointsPort != 0 {
				proxyEndpointsMux := http.NewServeMux()
				proxyEndpointsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })

				proxyEndpointsSrv := &http.Server{
					Handler:   proxyEndpointsMux,
					TLSConfig: srv.TLSConfig.Clone(),
				}

				if cfg.http2Disable {
					// HTTP/2 is temporarily disabled due to CVE-2023-44487
					// Programs that must disable HTTP/2 can do so by setting
					// Transport.TLSNextProto (for clients) or Server.TLSNextProto
					// (for servers) to a non-nil, empty map.
					// https://pkg.go.dev/net/http
					srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
					// For reference:
					// https://github.com/kubernetes/kubernetes/blob/de054fbf9422d778568946de21a48c7330a6c1b7/staging/src/k8s.io/apiserver/pkg/server/secure_serving.go#L55-L59
					srv.TLSConfig.NextProtos = []string{"http/1.1"}
				} else {
					if err := http2.ConfigureServer(proxyEndpointsSrv, cfg.http2Options); err != nil {
						return fmt.Errorf("failed to configure http2 server: %w", err)
					}
				}

				gr.Add(func() error {
					host, _, err := net.SplitHostPort(cfg.secureListenAddress)
					if err != nil {
						return fmt.Errorf("failed to split %q into host and port: %w", cfg.secureListenAddress, err)
					}
					endpointsAddr := net.JoinHostPort(host, strconv.Itoa(cfg.proxyEndpointsPort))

					klog.Infof("Starting TCP socket on %v", endpointsAddr)
					proxyListener, err := net.Listen("tcp", endpointsAddr)
					if err != nil {
						return fmt.Errorf("failed to listen on secure address: %w", err)
					}
					defer proxyListener.Close()

					klog.Infof("Listening securely on %v for proxy endpoints", endpointsAddr)
					tlsListener := tls.NewListener(proxyListener, srv.TLSConfig)
					return proxyEndpointsSrv.Serve(tlsListener)
				}, func(err error) {
					if err := proxyEndpointsSrv.Shutdown(context.Background()); err != nil {
						klog.Errorf("failed to gracefully shutdown proxy endpoints server: %+v", err)
					}
				})
			}
		}
	}
	{
		if cfg.insecureListenAddress != "" {
			srv := &http.Server{}
			if cfg.http2Disable {
				srv.Handler = mux
			} else {
				srv.Handler = h2c.NewHandler(mux, cfg.http2Options)
			}

			l, err := net.Listen("tcp", cfg.insecureListenAddress)
			if err != nil {
				return fmt.Errorf("failed to listen on insecure address: %w", err)
			}

			gr.Add(func() error {
				klog.Infof("Listening insecurely on %v", cfg.insecureListenAddress)
				return srv.Serve(l)
			}, func(err error) {
				if err := srv.Shutdown(context.Background()); err != nil {
					klog.Errorf("failed to gracefully shutdown server: %+v", err)
				}
				if err := l.Close(); err != nil {
					klog.Errorf("failed to gracefully close listener: %+v", err)
				}
			})
		}
	}
	{
		sig := make(chan os.Signal, 1)
		gr.Add(func() error {
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig
			klog.Info("received interrupt, shutting down")
			return nil
		}, func(err error) {
			close(sig)
		})
	}

	if len(cfg.secureListenAddress) == 0 && len(cfg.insecureListenAddress) == 0 {
		return fmt.Errorf("no listen address provided")
	}

	cfg.informerFactory.Start(ctx.Done())
	defer cfg.informerFactory.Shutdown()

	if err := gr.Run(); err != nil {
		return fmt.Errorf("failed to run groups: %w", err)
	}

	return nil
}

// Returns intiliazed config, allows local usage (outside cluster) based on provided kubeconfig or in-cluter
func initKubeConfig(kcLocation string) (*rest.Config, error) {
	if kcLocation != "" {
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kcLocation)
		if err != nil {
			return nil, fmt.Errorf("unable to build rest config based on provided path to kubeconfig file: %w", err)
		}
		return kubeConfig, nil
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot find Service Account in pod to build in-cluster rest config: %w", err)
	}

	return kubeConfig, nil
}

func parseAuthorizationConfigFile(filePath string) (*authz.Config, error) {
	klog.Infof("Reading config file: %s", filePath)
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource-attribute file: %w", err)
	}

	configFile := configfile{}

	if err := yaml.Unmarshal(b, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse config file content: %w", err)
	}

	return configFile.AuthorizationConfig, nil
}

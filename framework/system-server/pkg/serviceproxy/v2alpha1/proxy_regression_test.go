package v2alpha1

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/brancz/kube-rbac-proxy/cmd/kube-rbac-proxy/app/options"
	"github.com/brancz/kube-rbac-proxy/pkg/authz"
	"github.com/labstack/echo/v4"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

func writeProxyFixture(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func proxyKubeConfig(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"apiVersion":"v1","kind":"List","items":[]}`)
	}))
	t.Cleanup(server.Close)
	return writeProxyFixture(t, "kubeconfig", "apiVersion: v1\nkind: Config\nclusters:\n- name: fixture\n  cluster:\n    server: "+server.URL+"\ncontexts:\n- name: fixture\n  context:\n    cluster: fixture\n    user: fixture\ncurrent-context: fixture\nusers:\n- name: fixture\n  user: {}\n")
}
func proxyOptions(t *testing.T) *options.ProxyRunOptions {
	t.Helper()
	o := options.NewProxyRunOptions()
	o.Upstream = "http://127.0.0.1:1"
	o.KubeconfigLocation = proxyKubeConfig(t)
	return o
}
func completedFixture(t *testing.T) *completedProxyRunOptions {
	t.Helper()
	cfg, err := Complete(proxyOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	cfg.insecureListenAddress = ""
	cfg.secureListenAddress = ""
	return cfg
}

func TestProxyConfigurationContracts(t *testing.T) {
	o := proxyOptions(t)
	o.QPS = 25
	o.Burst = 40
	o.AllowPaths = []string{"/healthz"}
	o.IgnorePaths = []string{"/public"}
	o.UpstreamForceH2C = true
	o.ConfigFileName = writeProxyFixture(t, "authz.yaml", "authorization:\n  static:\n  - user:\n      name: fixture\n    path: /read\n    verb: get\n")
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer tlsServer.Close()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: tlsServer.Certificate().Raw})
	o.UpstreamCAFile = writeProxyFixture(t, "ca.pem", string(ca))
	cfg, err := Complete(o)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.upstreamURL.String() != o.Upstream || !cfg.upstreamForceH2C || cfg.upstreamCABundle == nil || len(cfg.auth.Authorization.Static) != 1 || cfg.http2Options.IdleTimeout != 90*time.Second || len(cfg.allowPaths) != 1 || len(cfg.ignorePaths) != 1 {
		t.Fatal("proxy configuration lost")
	}
	if cfg.kubeClient.RESTClient().GetRateLimiter().QPS() != 25 {
		t.Fatal("client QPS lost")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*options.ProxyRunOptions)
	}{
		{"bad-url", func(o *options.ProxyRunOptions) { o.Upstream = "://bad" }},
		{"missing-ca", func(o *options.ProxyRunOptions) { o.UpstreamCAFile = filepath.Join(t.TempDir(), "missing") }},
		{"invalid-ca", func(o *options.ProxyRunOptions) { o.UpstreamCAFile = writeProxyFixture(t, "ca", "invalid") }},
		{"missing-authz", func(o *options.ProxyRunOptions) { o.ConfigFileName = filepath.Join(t.TempDir(), "missing") }},
		{"invalid-authz", func(o *options.ProxyRunOptions) { o.ConfigFileName = writeProxyFixture(t, "authz", "authorization: [") }},
		{"missing-kubeconfig", func(o *options.ProxyRunOptions) { o.KubeconfigLocation = filepath.Join(t.TempDir(), "missing") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := proxyOptions(t)
			tc.mutate(options)
			if _, err := Complete(options); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	if _, err := initKubeConfig(""); err == nil {
		t.Fatal("local process unexpectedly used in-cluster identity")
	}
}

func TestProxyCommandRejectsPositionalArguments(t *testing.T) {
	cmd := NewKubeRBACProxyCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Fatal("positional argument accepted")
	}
	if err := cmd.Args(cmd, nil); err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"--upstream=://invalid", "--insecure-listen-address=127.0.0.1:0"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("invalid command configuration accepted")
	}
}

func TestProxyTransportTLS(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Fixture", "verified")
		w.WriteHeader(204)
	}))
	defer upstream.Close()
	pool := x509.NewCertPool()
	pool.AddCert(upstream.Certificate())
	transport, err := initTransport(pool, "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
	resp, err := client.Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	transport.(*http.Transport).CloseIdleConnections()
	if resp.StatusCode != 204 || resp.Header.Get("X-Fixture") != "verified" {
		t.Fatal("TLS upstream not reached")
	}
	if _, err := initTransport(pool, "missing-cert", "missing-key"); err == nil {
		t.Fatal("invalid client certificate accepted")
	}
	cert, key := upstream.TLS.Certificates[0].Certificate[0], upstream.TLS.Certificates[0].PrivateKey
	certPath := writeProxyFixture(t, "client.pem", string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})))
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := writeProxyFixture(t, "client.key", string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})))
	mtls, err := initTransport(pool, certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(mtls.(*http.Transport).TLSClientConfig.Certificates) != 1 {
		t.Fatal("client identity missing")
	}
	legacy, err := initTransport(nil, "", "")
	if err != nil || legacy.(*http.Transport).DialContext == nil {
		t.Fatal("legacy transport not constructed")
	}
}

type providerAuthorizer struct {
	service  string
	decision authorizer.Decision
	err      error
}

func (a providerAuthorizer) Authorize(context.Context, authorizer.Attributes) (string, authorizer.Decision, string, error) {
	return a.service, a.decision, "fixture", a.err
}
func TestProviderRBACBoundary(t *testing.T) {
	for _, tc := range []struct {
		name          string
		authenticated bool
		decision      authorizer.Decision
		service       string
		want          int
	}{
		{"anonymous", false, authorizer.DecisionAllow, "http://provider", 401},
		{"denied", true, authorizer.DecisionDeny, "http://provider", 403},
		{"no-service", true, authorizer.DecisionAllow, "", 400},
		{"allowed", true, authorizer.DecisionAllow, "http://provider", 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := completedFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			s := NewRBACProxyServer(ctx)
			s.authenticator = authenticator.RequestFunc(func(*http.Request) (*authenticator.Response, bool, error) {
				return &authenticator.Response{User: &user.DefaultInfo{Name: "alice"}}, tc.authenticated, nil
			})
			s.authorizer = providerAuthorizer{service: tc.service, decision: tc.decision}
			s.proxy.Use(s.rbac(cfg))
			s.proxy.Any("/*", func(c echo.Context) error {
				if c.Get(serviceKey) != tc.service || c.Request().Header.Get(constants.BflUserKey) != "alice" || c.Request().Header.Get("Authorization") != "fixture-upstream-auth" {
					t.Fatal("verified context not passed to provider")
				}
				return c.NoContent(204)
			})
			req := httptest.NewRequest("GET", "/read", nil)
			req.Header.Set("X-BFL-USER", "forged")
			req.Header.Set("Temp-Authorization", "fixture-upstream-auth")
			res := httptest.NewRecorder()
			s.proxy.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("status=%d want=%d", res.Code, tc.want)
			}
			if s.Authenticator() == nil {
				t.Fatal("authenticator accessor changed")
			}
		})
	}
}

func TestProxyInitializationAndShutdown(t *testing.T) {
	cfg := completedFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	s := NewRBACProxyServer(ctx)
	if err := s.Init(cfg); err != nil {
		cancel()
		t.Fatal(err)
	}
	if s.Authenticator() == nil || s.authorizer == nil {
		cancel()
		t.Fatal("proxy missing authorization")
	}
	cancel()
	cfg.informerFactory.Shutdown()
	cfg.appInformerFactory.Shutdown()
	for _, mode := range []string{"authn", "authz", "transport"} {
		t.Run(mode, func(t *testing.T) {
			cfg := completedFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch mode {
			case "authn":
				cfg.auth.Authentication.X509.ClientCAFile = "missing-ca"
			case "authz":
				cfg.auth.Authorization.Static = []authz.StaticAuthorizationConfig{{ResourceRequest: true, Path: "/invalid"}}
			case "transport":
				cfg.upstreamCABundle = x509.NewCertPool()
				cfg.tls.UpstreamClientCertFile = "missing-cert"
				cfg.tls.UpstreamClientKeyFile = "missing-key"
			}
			if err := NewRBACProxyServer(ctx).Init(cfg); err == nil {
				t.Fatal("invalid security setup accepted")
			}
		})
	}
	// Invalid listen address must be propagated by the Echo server.
	cfg.insecureListenAddress = "invalid-address"
	if err := s.Start(cfg); err == nil {
		t.Fatal("listen failure hidden")
	}
}

func TestProxyRunRejectsInvalidStartup(t *testing.T) {
	for _, mode := range []string{"no-listener", "authn", "authz", "transport", "bad-cert", "bad-min-version", "bad-cipher", "http1", "http2", "h2c"} {
		t.Run(mode, func(t *testing.T) {
			cfg := completedFixture(t)
			switch mode {
			case "authn":
				cfg.auth.Authentication.X509.ClientCAFile = "missing-ca"
			case "authz":
				cfg.auth.Authorization.Static = []authz.StaticAuthorizationConfig{{ResourceRequest: true, Path: "/invalid"}}
			case "transport":
				cfg.upstreamCABundle = x509.NewCertPool()
				cfg.tls.UpstreamClientCertFile = "missing-cert"
				cfg.tls.UpstreamClientKeyFile = "missing-key"
			case "bad-cert":
				cfg.secureListenAddress = "127.0.0.1:0"
				cfg.tls.CertFile = "missing-cert"
				cfg.tls.KeyFile = "missing-key"
			case "bad-min-version":
				cfg.secureListenAddress = "127.0.0.1:0"
				cfg.tls.MinVersion = "invalid"
			case "bad-cipher":
				cfg.secureListenAddress = "127.0.0.1:0"
				cfg.tls.CipherSuites = []string{"invalid"}
			case "http1", "http2", "h2c":
				// Occupying the insecure listener forces a bounded startup failure before
				// the signal loop. TLS and endpoints setup must preserve this error.
				occupied, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				defer occupied.Close()
				cfg.insecureListenAddress = occupied.Addr().String()
				cfg.http2Disable = mode == "http1"
				cfg.upstreamForceH2C = mode == "h2c"
				if mode != "h2c" {
					cfg.secureListenAddress = "127.0.0.1:0"
					cfg.proxyEndpointsPort = 8443
				}
			}
			err := Run(cfg)
			if err == nil {
				t.Fatal("invalid startup reported success")
			}
			if mode == "no-listener" && !strings.Contains(err.Error(), "no listen address") {
				t.Fatal(err)
			}
			if (mode == "http1" || mode == "http2" || mode == "h2c") && !strings.Contains(err.Error(), "failed to listen on insecure") {
				t.Fatal(err)
			}
		})
	}
}

func TestServerOptionsUsesPlatformConfiguration(t *testing.T) {
	t.Setenv("KUBECONFIG", proxyKubeConfig(t))
	t.Setenv("LLDAP_SERVER", "fixture-lldap")
	t.Setenv("LLDAP_PORT", "17200")
	cfg := ServerOptions("127.0.0.1:0")
	if cfg.lldapServer != "fixture-lldap" || cfg.lldapPort != 17200 || cfg.appLister == nil {
		t.Fatal("platform config lost")
	}
	t.Setenv("LLDAP_PORT", "invalid")
	if ServerOptions(":28080").lldapPort != 17170 {
		t.Fatal("LLDAP fallback port changed")
	}
}

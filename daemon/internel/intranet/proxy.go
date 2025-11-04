package intranet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"k8s.io/klog/v2"
)

var _ middleware.ProxyBalancer = (*proxyServer)(nil)

type key struct{}

var WSKey = key{}

type proxyServer struct {
	proxy     *echo.Echo
	dnsServer string
}

func NewProxyServer() (*proxyServer, error) {
	p := &proxyServer{
		proxy:     echo.New(),
		dnsServer: "10.233.0.3:53", // default k8s dns service
	}
	return p, nil
}

func (p *proxyServer) Start() error {
	klog.Info("Starting intranet proxy server...")
	config := middleware.DefaultProxyConfig
	config.Balancer = p
	config.Transport = p.initTransport()

	p.proxy.Use(middleware.Logger())
	p.proxy.Use(middleware.Recover())

	// add x-forwarded-proto header
	p.proxy.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Request().Header.Set("X-Forwarded-Proto", "http")
				return next(c)
			}
		},
	)

	// Handle HTTP to HTTPS redirection for non-intranet requests
	p.proxy.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if strings.HasSuffix(c.Request().Host, ".olares.local") ||
					strings.HasSuffix(c.Request().Host, "-olares.local") {
					if c.IsWebSocket() {
						ctx := c.Request().Context()
						ctx = context.WithValue(ctx, WSKey, true)
						r := c.Request().WithContext(ctx)
						c.SetRequest(r)
					}
					return next(c)
				}

				// not a intranet request, redirect to https
				redirect := middleware.HTTPSRedirect()
				return redirect(next)(c)
			}
		},
	)
	p.proxy.Use(middleware.ProxyWithConfig(config))

	go func() {
		err := p.proxy.Start(":80")
		if err != nil {
			klog.Error(err)
		}
	}()

	return nil
}

func (p *proxyServer) Close() error {
	if p.proxy != nil {
		return p.proxy.Close()
	}
	return nil
}

// AddTarget implements middleware.ProxyBalancer.
func (p *proxyServer) AddTarget(*middleware.ProxyTarget) bool {
	return true
}

// Next implements middleware.ProxyBalancer.
func (p *proxyServer) Next(c echo.Context) *middleware.ProxyTarget {
	scheme := "https://"
	if c.IsWebSocket() {
		scheme = "wss://"
	}

	var (
		proxyPass *url.URL
		err       error
	)
	requestHost := c.Request().Host
	if strings.HasSuffix(requestHost, "-olares.local") {
		// intranet request, and host parttern is appid-<username>-olares.local for windows and linux client
		tokens := strings.Split(requestHost, "-")
		if len(tokens) < 3 {
			klog.Error("invalid intranet request host, ", requestHost)
			return nil
		}
		requestHost = strings.Join(tokens, ".")
		c.Request().Host = requestHost
		proxyPass, err = url.Parse(scheme + requestHost + ":443")
	} else {
		proxyPass, err = url.Parse(scheme + c.Request().Host + ":443")
	}
	if err != nil {
		klog.Error("parse proxy target error, ", err)
		return nil
	}
	return &middleware.ProxyTarget{URL: proxyPass}
}

// RemoveTarget implements middleware.ProxyBalancer.
func (p *proxyServer) RemoveTarget(string) bool {
	return true
}

func (p *proxyServer) initTransport() http.RoundTripper {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: p.customDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 1800 * time.Second,
			DualStack: true,
		}),
		MaxIdleConns:          100,
		IdleConnTimeout:       10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	return transport
}

func (p *proxyServer) customDialContext(d *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		_, port, _ := net.SplitHostPort(addr)
		// Force proxying to localhost
		klog.Info("addr: ", addr, " port: ", port, " network: ", network)
		if port == "" {
			port = "443"
		}
		newAddr := net.JoinHostPort("127.0.0.1", port)

		isWs := false
		if v := ctx.Value(WSKey); v != nil {
			isWs = v.(bool)
		}
		if isWs {
			klog.Info("WebSocket connection detected, using upgraded dialer")
			return tlsDial(ctx, d, func(ctx context.Context, network, addr string) (net.Conn, error) {
				return d.DialContext(ctx, network, newAddr)
			}, network, addr, &tls.Config{InsecureSkipVerify: true})
		}

		return d.DialContext(ctx, network, newAddr)
	}
}

func tlsDial(ctx context.Context, netDialer *net.Dialer, dialFunc func(ctx context.Context, network, addr string) (net.Conn, error), network, addr string, config *tls.Config) (*tls.Conn, error) {
	if netDialer.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, netDialer.Timeout)
		defer cancel()
	}

	if !netDialer.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, netDialer.Deadline)
		defer cancel()
	}

	var (
		rawConn net.Conn
		err     error
	)

	if dialFunc != nil {
		rawConn, err = dialFunc(ctx, network, addr)
	} else {
		rawConn, err = netDialer.DialContext(ctx, network, addr)
	}
	if err != nil {
		return nil, err
	}

	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]

	if config == nil {
		return nil, fmt.Errorf("tls: config is nil")
	}
	// If no ServerName is set, infer the ServerName
	// from the hostname we're connecting to.
	if config.ServerName == "" {
		// Make a copy to avoid polluting argument or default.
		c := config.Clone()
		c.ServerName = hostname
		config = c
	}

	conn := tls.Client(rawConn, config)
	if err := conn.HandshakeContext(ctx); err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}

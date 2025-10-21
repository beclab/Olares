package intranet

import (
	"context"
	"crypto/tls"
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
	p.proxy.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if strings.HasSuffix(c.Request().Host, ".olares.local") {
					return next(c)
				}

				// not a intranet request, redirect to https
				redirect := middleware.HTTPSRedirect()
				return redirect(next)(c)
			}
		},
	)
	p.proxy.Use(middleware.ProxyWithConfig(config))

	go func(){
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
	proxyPass, err := url.Parse("https://" + c.Request().Host)
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
		addr = net.JoinHostPort("127.0.0.1", port)
		return d.DialContext(ctx, network, addr)
	}
}

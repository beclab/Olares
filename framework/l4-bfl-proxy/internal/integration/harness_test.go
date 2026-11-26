//go:build integration

// Package integration runs the translated xDS config against a real Envoy
// process. It requires an Envoy binary (ENVOY_BIN or envoy on PATH) and is
// excluded from the default build; run with `make test-integration` or
// `make test-integration-docker`.
package integration

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beclab/l4-bfl-proxy/internal/envoy"
	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/beclab/l4-bfl-proxy/internal/xds/server"
	xdstranslator "github.com/beclab/l4-bfl-proxy/internal/xds/translator"
)

const (
	appHost        = "app.test"
	httpIRName     = "it_user"
	upstreamName   = "it_upstream"
	defaultTimeout = 10 * time.Second
)

func envoyBinary(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("ENVOY_BIN"); p != "" {
		return p
	}
	p, err := exec.LookPath("envoy")
	if err != nil {
		t.Skip("envoy binary not found: set ENVOY_BIN or put envoy on PATH")
	}
	return p
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// testOptions shortens every timer so behaviours that take minutes in
// production can be observed in seconds.
func testOptions() xdstranslator.Options {
	o := xdstranslator.DefaultOptions()
	o.RouteTimeout = 0
	o.HTTPStreamIdleTimeout = 3 * time.Second
	o.ClusterIdleTimeout = 2 * time.Second
	o.TCPKeepaliveTime = 1 * time.Second
	o.TCPKeepaliveInterval = 1 * time.Second
	o.TCPKeepaliveProbes = 2
	return o
}

type proxy struct {
	t         *testing.T
	adminAddr string
	port      int
	logPath   string
}

type proxyConfig struct {
	opts      xdstranslator.Options
	upstream  int
	pushState bool
}

// startProxy translates a single-user HTTPS listener (plaintext, no ext_authz)
// routing appHost to 127.0.0.1:upstream, serves it over the production xDS
// server and boots Envoy with the production bootstrap.
func startProxy(t *testing.T, cfg proxyConfig) *proxy {
	t.Helper()
	bin := envoyBinary(t)

	prev := xdstranslator.CurrentOptions()
	xdstranslator.SetOptions(cfg.opts)
	t.Cleanup(func() { xdstranslator.SetOptions(prev) })

	listenPort := freePort(t)
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{{
			Name: httpIRName,
			Port: uint32(listenPort),
			VirtualHosts: []*ir.VirtualHostIR{{
				Name:                  "app",
				Domains:               []string{appHost},
				EnableWindowPushState: cfg.pushState,
				Routes: []*ir.HTTPRouteIR{{
					Name:       "app",
					PathPrefix: "/",
					Cluster:    upstreamName,
				}},
			}},
		}},
		Clusters: []*ir.ClusterIR{{
			Name: upstreamName,
			Host: "127.0.0.1",
			Port: uint32(cfg.upstream),
		}},
	}
	snap := xdstranslator.New(nil, nil).Translate(xdsIR)

	resources := &message.XdsResources{}
	resources.Store("it", snap)

	xdsPort := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	srvDone := make(chan struct{})
	go func() {
		defer close(srvDone)
		_ = server.New(resources, &server.Config{Address: "127.0.0.1", Port: xdsPort}).Start(ctx)
	}()

	dir := t.TempDir()
	adminPort := freePort(t)
	bootstrap := filepath.Join(dir, "envoy.yaml")
	bc := envoy.DefaultBootstrapConfig(xdsPort)
	bc.AdminPort = adminPort
	require.NoError(t, envoy.WriteBootstrapConfig(bootstrap, bc))

	logPath := filepath.Join(dir, "envoy.log")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)

	// A single worker keeps one upstream pool, so pool reuse is deterministic.
	cmd := exec.Command(bin, "-c", bootstrap, "--use-dynamic-base-id", "--log-level", "info", "--concurrency", "1",
		"--service-cluster", "l4-bfl-proxy", "--service-node", "l4-bfl-proxy")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	require.NoError(t, cmd.Start())

	p := &proxy{t: t, adminAddr: fmt.Sprintf("127.0.0.1:%d", adminPort), port: listenPort, logPath: logPath}

	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
		cancel()
		<-srvDone
		logFile.Close()
		if t.Failed() {
			if b, err := os.ReadFile(logPath); err == nil {
				t.Logf("envoy log:\n%s", b)
			}
		}
	})

	p.waitReady()
	return p
}

func (p *proxy) waitReady() {
	p.t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if p.stat("listener_manager.total_listeners_active") >= 1 &&
			p.stat("cluster_manager.warming_clusters") == 0 &&
			p.stat("cluster_manager.active_clusters") >= 2 {
			c, err := net.DialTimeout("tcp", p.addr(), time.Second)
			if err == nil {
				c.Close()
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	b, _ := os.ReadFile(p.logPath)
	p.t.Fatalf("envoy did not become ready; log:\n%s", b)
}

func (p *proxy) addr() string { return fmt.Sprintf("127.0.0.1:%d", p.port) }

func (p *proxy) log() string {
	b, _ := os.ReadFile(p.logPath)
	return string(b)
}

// stat returns an Envoy counter/gauge value, or -1 when it is not present.
func (p *proxy) stat(name string) int {
	resp, err := http.Get(fmt.Sprintf("http://%s/stats?filter=^%s$", p.adminAddr, strings.ReplaceAll(name, ".", `\.`)))
	if err != nil {
		return -1
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), ": ")
		if ok && k == name {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err == nil {
				return n
			}
		}
	}
	return -1
}

func (p *proxy) upstreamCxActive() int {
	return p.stat("cluster." + upstreamName + ".upstream_cx_active")
}
func (p *proxy) upstreamRqActive() int {
	return p.stat("cluster." + upstreamName + ".upstream_rq_active")
}
func (p *proxy) downstreamCxActive() int {
	return p.stat("http." + httpIRName + ".downstream_cx_active")
}

// requireNoLeak asserts Envoy holds no upstream request/connection and no
// downstream connection within the deadline.
func (p *proxy) requireNoLeak(within time.Duration) {
	p.t.Helper()
	eventually(p.t, within, func() bool {
		return p.upstreamRqActive() == 0 && p.upstreamCxActive() == 0 && p.downstreamCxActive() == 0
	}, func() string {
		return fmt.Sprintf("upstream_rq_active=%d upstream_cx_active=%d downstream_cx_active=%d",
			p.upstreamRqActive(), p.upstreamCxActive(), p.downstreamCxActive())
	})
}

func eventually(t *testing.T, within time.Duration, cond func() bool, msg func() string) {
	t.Helper()
	deadline := time.Now().Add(within)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s: %s", within, msg())
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// newClient returns a client with Go's own TCP keepalive disabled so the
// kernel timers observed in /proc belong to Envoy only.
func newClient(timeout time.Duration) *http.Client {
	d := &net.Dialer{KeepAlive: -1}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:       d.DialContext,
			DisableKeepAlives: true,
		},
	}
}

func (p *proxy) newRequest(t *testing.T, ctx context.Context, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+p.addr()+path, nil)
	require.NoError(t, err)
	req.Host = appHost
	return req
}

// upstream is a Go HTTP server that records connection lifecycle events.
type upstream struct {
	port   int
	srv    *http.Server
	opened atomic.Int32
	closed atomic.Int32
}

func startUpstream(t *testing.T, h http.Handler) *upstream {
	t.Helper()
	lc := net.ListenConfig{KeepAlive: -1}
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	u := &upstream{port: l.Addr().(*net.TCPAddr).Port}
	u.srv = &http.Server{
		Handler: h,
		ConnState: func(_ net.Conn, s http.ConnState) {
			switch s {
			case http.StateNew:
				u.opened.Add(1)
			case http.StateClosed, http.StateHijacked:
				u.closed.Add(1)
			}
		},
	}
	go func() { _ = u.srv.Serve(l) }()
	t.Cleanup(func() { _ = u.srv.Close() })
	return u
}

func (u *upstream) liveConns() int32 { return u.opened.Load() - u.closed.Load() }

// rawUpstream accepts TCP connections and hands each to fn, for behaviours a
// well-behaved http.Server cannot produce (hang, abrupt RST).
type rawUpstream struct {
	port int
	l    net.Listener
	mu   sync.Mutex
	eofs int
}

func startRawUpstream(t *testing.T, fn func(c *net.TCPConn, u *rawUpstream)) *rawUpstream {
	t.Helper()
	lc := net.ListenConfig{KeepAlive: -1}
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	u := &rawUpstream{port: l.Addr().(*net.TCPAddr).Port, l: l}
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go fn(c.(*net.TCPConn), u)
		}
	}()
	t.Cleanup(func() { _ = l.Close() })
	return u
}

// readRequestHead consumes the HTTP request head sent by Envoy.
func readRequestHead(r *bufio.Reader) error {
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		if line == "\r\n" {
			return nil
		}
	}
}

// waitPeerClose blocks until the peer closes the connection, then records it.
func (u *rawUpstream) waitPeerClose(r io.Reader) {
	_, _ = io.Copy(io.Discard, r)
	u.mu.Lock()
	u.eofs++
	u.mu.Unlock()
}

func (u *rawUpstream) peerCloses() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.eofs
}

// tcpTimerKinds returns the kernel timer kind ("tr" column of /proc/net/tcp)
// for every socket whose local or remote port matches. 2 means keepalive.
func tcpTimerKinds(t *testing.T, localPort, remotePort int) []int {
	t.Helper()
	b, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		t.Skip("/proc/net/tcp unavailable; keepalive check requires Linux")
	}
	var kinds []int
	lines := strings.Split(string(b), "\n")
	for _, line := range lines[1:] {
		f := strings.Fields(line)
		if len(f) < 6 {
			continue
		}
		lp := hexPort(f[1])
		rp := hexPort(f[2])
		if (localPort != 0 && lp != localPort) || (remotePort != 0 && rp != remotePort) {
			continue
		}
		if f[3] != "01" { // ESTABLISHED
			continue
		}
		tr, _, _ := strings.Cut(f[5], ":")
		k, _ := strconv.Atoi(tr)
		kinds = append(kinds, k)
	}
	return kinds
}

func hexPort(addr string) int {
	_, p, ok := strings.Cut(addr, ":")
	if !ok {
		return -1
	}
	b, err := hex.DecodeString(p)
	if err != nil || len(b) != 2 {
		return -1
	}
	return int(b[0])<<8 | int(b[1])
}

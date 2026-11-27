//go:build integration

package integration

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sse emulates an LLM server: it waits ttft before sending anything (not even
// headers), then streams tokens until done or the request is cancelled.
func sse(ttft time.Duration, tokens int, every time.Duration, cancelled *atomic.Int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(ttft):
		case <-r.Context().Done():
			cancelled.Add(1)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f := w.(http.Flusher)
		for i := 0; tokens < 0 || i < tokens; i++ {
			fmt.Fprintf(w, "data: token-%d\n\n", i)
			f.Flush()
			select {
			case <-time.After(every):
			case <-r.Context().Done():
				cancelled.Add(1)
				return
			}
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}
}

// TestSlowTTFTAndLongDecode: a request whose first byte arrives just under the
// idle timeout and whose total duration far exceeds it completes intact,
// because there is no absolute route timeout.
func TestSlowTTFTAndLongDecode(t *testing.T) {
	var cancelled atomic.Int32
	up := startUpstream(t, sse(2500*time.Millisecond, 10, time.Second, &cancelled))
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	start := time.Now()
	resp, err := newClient(time.Minute).Do(p.newRequest(t, context.Background(), "/v1/chat/completions"))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err, "stream must not be cut off")
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 10, strings.Count(string(body), "data: token-"))
	assert.Contains(t, string(body), "[DONE]")
	assert.Greater(t, elapsed, 3*testOptions().HTTPStreamIdleTimeout, "stream should outlive the idle timeout several times over")
	assert.Zero(t, cancelled.Load())
	p.requireNoLeak(5 * time.Second)
}

// TestAbsoluteRouteTimeoutCutsLongDecode documents the old failure mode: a
// non-zero route timeout truncates a healthy long-running stream.
func TestAbsoluteRouteTimeoutCutsLongDecode(t *testing.T) {
	var cancelled atomic.Int32
	up := startUpstream(t, sse(0, 10, time.Second, &cancelled))
	o := testOptions()
	o.RouteTimeout = 4 * time.Second
	p := startProxy(t, proxyConfig{opts: o, upstream: up.port})

	resp, err := newClient(time.Minute).Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Error(t, err, "absolute route timeout should reset the stream")
	assert.NotContains(t, string(body), "[DONE]")
	eventually(t, 3*time.Second, func() bool { return cancelled.Load() == 1 }, func() string { return "upstream not cancelled" })
}

// TestUpstreamHangBeforeHeaders: an upstream that accepts the request but never
// answers is released after the route idle timeout, on both sides.
func TestUpstreamHangBeforeHeaders(t *testing.T) {
	up := startRawUpstream(t, func(c *net.TCPConn, u *rawUpstream) {
		defer c.Close()
		r := bufio.NewReader(c)
		if readRequestHead(r) == nil {
			u.waitPeerClose(r)
		}
	})
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	start := time.Now()
	resp, err := newClient(time.Minute).Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
	assert.InDelta(t, testOptions().HTTPStreamIdleTimeout.Seconds(), elapsed.Seconds(), 1.5)
	eventually(t, 3*time.Second, func() bool { return up.peerCloses() == 1 }, func() string { return "envoy kept the hung upstream connection open" })
	p.requireNoLeak(3 * time.Second)
}

// TestUpstreamStallMidStream: a stream that stops producing bytes after the
// headers is reset after the idle timeout and the upstream sees cancellation.
func TestUpstreamStallMidStream(t *testing.T) {
	var cancelled atomic.Int32
	up := startUpstream(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: first\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		cancelled.Add(1)
	}))
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	start := time.Now()
	resp, err := newClient(time.Minute).Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, string(body), "data: first")
	assert.InDelta(t, testOptions().HTTPStreamIdleTimeout.Seconds(), elapsed.Seconds(), 1.5)
	eventually(t, 3*time.Second, func() bool { return cancelled.Load() == 1 }, func() string { return "upstream handler not cancelled" })
	p.requireNoLeak(3 * time.Second)
}

// TestClientDisconnectMidStream: closing the browser/tab mid-generation must
// cancel the upstream request immediately instead of letting it run on.
func TestClientDisconnectMidStream(t *testing.T) {
	var cancelled atomic.Int32
	up := startUpstream(t, sse(0, -1, 200*time.Millisecond, &cancelled))
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	ctx, cancel := context.WithCancel(context.Background())
	resp, err := newClient(time.Minute).Do(p.newRequest(t, ctx, "/"))
	require.NoError(t, err)
	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, line, "token-0")
	require.Equal(t, 1, p.upstreamRqActive())

	closedAt := time.Now()
	cancel()
	resp.Body.Close()

	eventually(t, 2*time.Second, func() bool { return cancelled.Load() == 1 }, func() string { return "upstream handler still running after client left" })
	t.Logf("upstream cancelled %s after client disconnect", time.Since(closedAt))
	p.requireNoLeak(2 * time.Second)
	eventually(t, 2*time.Second, func() bool { return up.liveConns() == 0 }, func() string {
		return fmt.Sprintf("upstream live conns=%d", up.liveConns())
	})
}

// TestUpstreamCrashMidStream: an app that dies mid-stream (RST) propagates to
// the client right away and leaves nothing behind in Envoy.
func TestUpstreamCrashMidStream(t *testing.T) {
	up := startRawUpstream(t, func(c *net.TCPConn, _ *rawUpstream) {
		r := bufio.NewReader(c)
		if readRequestHead(r) != nil {
			c.Close()
			return
		}
		_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nTransfer-Encoding: chunked\r\n\r\n")
		msg := "data: token-0\n\n"
		_, _ = fmt.Fprintf(c, "%x\r\n%s\r\n", len(msg), msg)
		time.Sleep(300 * time.Millisecond)
		_ = c.SetLinger(0)
		_ = c.Close()
	})
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	start := time.Now()
	resp, err := newClient(time.Minute).Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	assert.Error(t, err, "client must observe the upstream crash")
	assert.Contains(t, string(body), "token-0")
	assert.Less(t, time.Since(start), 2*time.Second, "crash must not wait for the idle timeout")
	p.requireNoLeak(2 * time.Second)
}

// TestUpstreamStopped: once an app is stopped, new requests fail fast with
// 503 rather than hanging, and no connection is left behind.
func TestUpstreamStopped(t *testing.T) {
	var cancelled atomic.Int32
	up := startUpstream(t, sse(0, 1, 0, &cancelled))
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	c := newClient(10 * time.Second)
	resp, err := c.Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, up.srv.Close())

	start := time.Now()
	resp, err = c.Do(p.newRequest(t, context.Background(), "/"))
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Less(t, time.Since(start), 2*time.Second)
	p.requireNoLeak(3 * time.Second)
}

// TestPoolReuseAndIdleReap: short requests share one pooled upstream
// connection, which is closed once idle for ClusterIdleTimeout.
func TestPoolReuseAndIdleReap(t *testing.T) {
	up := startUpstream(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	p := startProxy(t, proxyConfig{opts: testOptions(), upstream: up.port})

	c := newClient(5 * time.Second)
	for i := 0; i < 5; i++ {
		resp, err := c.Do(p.newRequest(t, context.Background(), "/"))
		require.NoError(t, err)
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
	assert.EqualValues(t, 1, up.opened.Load(), "sequential requests should reuse one pooled upstream connection")

	eventually(t, testOptions().ClusterIdleTimeout+3*time.Second, func() bool {
		return up.liveConns() == 0 && p.upstreamCxActive() == 0
	}, func() string {
		return fmt.Sprintf("upstream live conns=%d upstream_cx_active=%d", up.liveConns(), p.upstreamCxActive())
	})
}

// TestTCPKeepaliveArmed: while a request is idle in flight, the kernel
// keepalive timer is armed on Envoy's downstream and upstream sockets, so a
// peer that vanishes without FIN/RST is detected.
func TestTCPKeepaliveArmed(t *testing.T) {
	up := startRawUpstream(t, func(c *net.TCPConn, u *rawUpstream) {
		defer c.Close()
		r := bufio.NewReader(c)
		if readRequestHead(r) == nil {
			u.waitPeerClose(r)
		}
	})
	o := testOptions()
	o.HTTPStreamIdleTimeout = 30 * time.Second
	p := startProxy(t, proxyConfig{opts: o, upstream: up.port})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		resp, err := newClient(time.Minute).Do(p.newRequest(t, ctx, "/"))
		if err == nil {
			resp.Body.Close()
		}
	}()
	eventually(t, 5*time.Second, func() bool { return p.upstreamRqActive() == 1 }, func() string { return "request not in flight" })

	const keepaliveTimer = 2
	eventually(t, 5*time.Second, func() bool {
		return slices.Contains(tcpTimerKinds(t, p.port, 0), keepaliveTimer)
	}, func() string {
		return fmt.Sprintf("downstream socket timers=%v", tcpTimerKinds(t, p.port, 0))
	})
	eventually(t, 5*time.Second, func() bool {
		return slices.Contains(tcpTimerKinds(t, 0, up.port), keepaliveTimer)
	}, func() string {
		return fmt.Sprintf("upstream socket timers=%v", tcpTimerKinds(t, 0, up.port))
	})

	cancel()
	p.requireNoLeak(3 * time.Second)
}

// TestLargeHTMLWithLuaInjection: the Lua pushState filter buffers whole HTML
// bodies; with the 1MiB default a 40KB page is rewritten, while the former
// 32KB limit reproduces 500 response_payload_too_large.
func TestLargeHTMLWithLuaInjection(t *testing.T) {
	page := "<html><head></head><body>" + strings.Repeat("x", 40*1024) + "</body></html>"
	up := startUpstream(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, page)
	}))

	cases := []struct {
		name   string
		limit  uint32
		status int
	}{
		{"default_1MiB", 1 << 20, http.StatusOK},
		{"legacy_32KiB", 32768, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := testOptions()
			o.HTTPSBufferLimitBytes = tc.limit
			p := startProxy(t, proxyConfig{opts: o, upstream: up.port, pushState: true})

			resp, err := newClient(10 * time.Second).Do(p.newRequest(t, context.Background(), "/"))
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tc.status, resp.StatusCode)
			if tc.status == http.StatusOK {
				assert.Contains(t, string(body), "pushstate")
				assert.Greater(t, len(body), len(page))
			} else {
				eventually(t, 3*time.Second, func() bool {
					return strings.Contains(p.log(), "details=response_payload_too_large")
				}, func() string { return "access log lacks response_payload_too_large" })
			}
			p.requireNoLeak(3 * time.Second)
		})
	}
}

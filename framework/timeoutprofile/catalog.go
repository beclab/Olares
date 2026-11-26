// Package timeoutprofile holds the platform's default timeouts for every hop
// on the user access path, the traffic classes those defaults are grouped by,
// and the bounds an app entrance may override them within.
//
// The package has no dependencies so that any Olares module can consume it.
package timeoutprofile

import (
	"errors"
	"fmt"
	"time"
)

// CatalogVersion identifies the set of defaults returned by Default. Bump it
// whenever a value changes so components built from different commits can be
// told apart.
const CatalogVersion = "2026.11.1"

// Class groups traffic that needs the same idle behaviour.
type Class string

const (
	// Interactive covers pages and short API calls. It is the default class.
	Interactive Class = "interactive"
	// Streaming covers SSE and token streaming; bytes flow, but with gaps.
	Streaming Class = "streaming"
	// WebSocket covers upgraded long-lived connections.
	WebSocket Class = "websocket"
	// Transfer covers large uploads and downloads that report progress.
	Transfer Class = "transfer"
)

// Classes returns every class in a stable order.
func Classes() []Class {
	return []Class{Interactive, Streaming, WebSocket, Transfer}
}

// ClassProfile is the idle budget of one traffic class.
type ClassProfile struct {
	// EdgeIdle is the edge idle timeout; any byte in either direction resets it.
	EdgeIdle time.Duration
	// InnerIdleMax is the longest read/send idle an inner platform proxy may
	// use for this class. It must not exceed EdgeIdle, otherwise the edge
	// cuts streams the inner hop still considers healthy.
	InnerIdleMax time.Duration
}

// Edge holds connection-level protection enforced by the edge proxy. None of
// these values are overridable by apps.
type Edge struct {
	// RouteTimeout is the per-request total deadline at the edge. It must stay
	// zero: total budgets belong to inner hops, and a total deadline at the
	// edge cuts long downloads and streams regardless of progress.
	RouteTimeout time.Duration
	// RouteIdle is the per-route idle timeout. Zero leaves it unset so routes
	// inherit StreamIdle.
	RouteIdle time.Duration
	// RequestHeaders bounds how long a client may take to send request headers.
	RequestHeaders time.Duration
	// StreamIdle is the edge stream idle timeout inherited by every route
	// without RouteIdle; it also closes idle downstream connections.
	StreamIdle time.Duration
	// MaxStreamDuration is the hard upper bound of a single stream.
	MaxStreamDuration time.Duration
	// Connect bounds upstream TCP connect, including the TLS handshake when
	// the hop speaks TLS to its upstream.
	Connect time.Duration
	// FilesIdle is the idle timeout of the system Files routes. It matches the
	// longest idle used by the Files proxies behind the edge.
	FilesIdle time.Duration
	// AuthTimeout bounds a single call to the authorization service.
	AuthTimeout time.Duration
	// AuthFailureStatus is returned when the authorization service fails or
	// times out, so outages are not reported to users as access denied.
	AuthFailureStatus int
	// ConnectionBufferBytes caps per-connection buffering (backpressure, not
	// a body size limit). Filters that buffer a whole body, such as the edge
	// Lua HTML rewrite, fail with 500 response_payload_too_large beyond it.
	ConnectionBufferBytes uint32
	// UpstreamPoolIdle closes pooled upstream connections that carry no
	// request. It must stay below backend keep-alive limits to avoid resets
	// on reuse, and never affects in-flight requests.
	UpstreamPoolIdle time.Duration
	// TCPIdle is the idle timeout of raw TCP ports apps expose through
	// the edge. It does not apply to HTTPS traffic.
	TCPIdle time.Duration
	// TCP keepalive on edge downstream and upstream sockets, so peers that
	// vanish without FIN/RST are detected after about
	// TCPKeepaliveTime + TCPKeepaliveInterval*TCPKeepaliveProbes.
	TCPKeepaliveTime     time.Duration
	TCPKeepaliveInterval time.Duration
	TCPKeepaliveProbes   uint32
}

// GoServer holds the timeouts of platform net/http servers.
type GoServer struct {
	ReadHeader time.Duration
	Idle       time.Duration
}

// FastHTTPServer holds the timeouts of platform fasthttp servers, which
// cannot bound header reading separately from the body.
type FastHTTPServer struct {
	Read time.Duration
	Idle time.Duration
}

// IntranetProxy holds the timeouts of the host LAN proxy in front of the edge.
type IntranetProxy struct {
	ReadHeader   time.Duration
	Idle         time.Duration
	UpstreamDial time.Duration
}

// InnerProxy holds the defaults of platform nginx hops behind the edge.
type InnerProxy struct {
	Connect time.Duration
}

// Bounds limits what an app entrance may declare.
type Bounds struct {
	IdleMin  time.Duration
	IdleMax  time.Duration
	TotalMin time.Duration
	TotalMax time.Duration
}

// Catalog is the complete set of platform timeout defaults.
type Catalog struct {
	Version        string
	Edge           Edge
	Classes        map[Class]ClassProfile
	GoServer       GoServer
	FastHTTPServer FastHTTPServer
	IntranetProxy  IntranetProxy
	InnerProxy     InnerProxy
	Bounds         Bounds
}

// Default returns a fresh copy of the platform defaults.
func Default() Catalog {
	return Catalog{
		Version: CatalogVersion,
		Edge: Edge{
			RouteTimeout:          0,
			RouteIdle:             0,
			RequestHeaders:        10 * time.Second,
			StreamIdle:            1800 * time.Second,
			MaxStreamDuration:     24 * time.Hour,
			Connect:               5 * time.Second,
			FilesIdle:             1800 * time.Second,
			AuthTimeout:           15 * time.Second,
			AuthFailureStatus:     503,
			ConnectionBufferBytes: 1 << 20,
			UpstreamPoolIdle:      10 * time.Second,
			TCPIdle:               time.Hour,
			TCPKeepaliveTime:      60 * time.Second,
			TCPKeepaliveInterval:  15 * time.Second,
			TCPKeepaliveProbes:    4,
		},
		Classes: map[Class]ClassProfile{
			Interactive: {EdgeIdle: 600 * time.Second, InnerIdleMax: 600 * time.Second},
			Streaming:   {EdgeIdle: 600 * time.Second, InnerIdleMax: 600 * time.Second},
			WebSocket:   {EdgeIdle: 3600 * time.Second, InnerIdleMax: 3600 * time.Second},
			Transfer:    {EdgeIdle: 300 * time.Second, InnerIdleMax: 300 * time.Second},
		},
		GoServer:       GoServer{ReadHeader: 10 * time.Second, Idle: 120 * time.Second},
		FastHTTPServer: FastHTTPServer{Read: 30 * time.Second, Idle: 120 * time.Second},
		IntranetProxy: IntranetProxy{
			ReadHeader:   10 * time.Second,
			Idle:         120 * time.Second,
			UpstreamDial: 5 * time.Second,
		},
		InnerProxy: InnerProxy{Connect: 5 * time.Second},
		Bounds: Bounds{
			IdleMin:  30 * time.Second,
			IdleMax:  3600 * time.Second,
			TotalMin: 30 * time.Second,
			TotalMax: 24 * time.Hour,
		},
	}
}

// Profile returns the profile of class. Unknown or empty classes fall back to
// Interactive, so renderers reading untrusted input never fail open.
func (c Catalog) Profile(class Class) ClassProfile {
	if p, ok := c.Classes[class]; ok {
		return p
	}
	return c.Classes[Interactive]
}

// Validate checks the catalog for missing values and budget inversions.
func (c Catalog) Validate() error {
	var errs []error
	check := func(ok bool, format string, args ...any) {
		if !ok {
			errs = append(errs, fmt.Errorf(format, args...))
		}
	}

	check(c.Version != "", "version is empty")

	e := c.Edge
	check(e.RouteTimeout == 0, "edge route timeout must be 0 (disabled), got %s", e.RouteTimeout)
	check(e.RequestHeaders > 0, "edge request headers timeout must be positive, got %s", e.RequestHeaders)
	check(e.MaxStreamDuration > 0, "edge max stream duration must be positive, got %s", e.MaxStreamDuration)
	check(e.Connect > 0, "edge connect timeout must be positive, got %s", e.Connect)
	check(e.AuthTimeout > 0, "edge auth timeout must be positive, got %s", e.AuthTimeout)
	check(e.AuthFailureStatus >= 500 && e.AuthFailureStatus <= 599,
		"edge auth failure status must be a 5xx, got %d", e.AuthFailureStatus)
	check(e.ConnectionBufferBytes > 0, "edge connection buffer must be positive")
	check(e.UpstreamPoolIdle > 0, "edge upstream pool idle must be positive, got %s", e.UpstreamPoolIdle)
	check(e.TCPIdle > 0, "edge tcp idle must be positive, got %s", e.TCPIdle)
	check(e.TCPKeepaliveTime > 0 && e.TCPKeepaliveInterval > 0 && e.TCPKeepaliveProbes > 0,
		"edge tcp keepalive time, interval and probes must be positive")

	b := c.Bounds
	check(e.StreamIdle > 0 && e.StreamIdle <= b.IdleMax,
		"edge stream idle %s must be positive and not exceed idle max %s", e.StreamIdle, b.IdleMax)
	check(e.RouteIdle == 0 || (e.RouteIdle >= b.IdleMin && e.RouteIdle <= b.IdleMax),
		"edge route idle %s must be 0 (inherit stream idle) or within [%s, %s]", e.RouteIdle, b.IdleMin, b.IdleMax)
	check(b.IdleMin > 0 && b.IdleMin <= b.IdleMax, "idle bounds invalid: [%s, %s]", b.IdleMin, b.IdleMax)
	check(b.TotalMin > 0 && b.TotalMin <= b.TotalMax, "total bounds invalid: [%s, %s]", b.TotalMin, b.TotalMax)
	check(b.TotalMax <= e.MaxStreamDuration,
		"total max %s exceeds edge max stream duration %s", b.TotalMax, e.MaxStreamDuration)

	for _, class := range Classes() {
		p, ok := c.Classes[class]
		if !ok {
			errs = append(errs, fmt.Errorf("class %q missing", class))
			continue
		}
		check(p.EdgeIdle >= b.IdleMin && p.EdgeIdle <= b.IdleMax,
			"class %q edge idle %s outside [%s, %s]", class, p.EdgeIdle, b.IdleMin, b.IdleMax)
		check(p.InnerIdleMax > 0 && p.InnerIdleMax <= p.EdgeIdle,
			"class %q inner idle %s must be positive and not exceed edge idle %s", class, p.InnerIdleMax, p.EdgeIdle)
		check(e.RequestHeaders < p.EdgeIdle,
			"class %q edge idle %s must exceed request headers timeout %s", class, p.EdgeIdle, e.RequestHeaders)
		check(e.AuthTimeout < p.EdgeIdle,
			"class %q edge idle %s must exceed auth timeout %s", class, p.EdgeIdle, e.AuthTimeout)
	}

	if transfer, ok := c.Classes[Transfer]; ok {
		check(e.FilesIdle >= transfer.EdgeIdle && e.FilesIdle <= b.IdleMax,
			"edge files idle %s outside [%s, %s]", e.FilesIdle, transfer.EdgeIdle, b.IdleMax)
	}

	check(c.GoServer.ReadHeader > 0 && c.GoServer.Idle > 0, "go server timeouts must be positive")
	check(c.FastHTTPServer.Read > 0 && c.FastHTTPServer.Idle > 0, "fasthttp server timeouts must be positive")
	check(c.IntranetProxy.ReadHeader > 0 && c.IntranetProxy.Idle > 0, "intranet proxy timeouts must be positive")
	check(c.IntranetProxy.UpstreamDial > 0 && c.IntranetProxy.UpstreamDial <= e.Connect,
		"intranet proxy upstream dial %s must be positive and not exceed edge connect %s",
		c.IntranetProxy.UpstreamDial, e.Connect)
	check(c.InnerProxy.Connect > 0, "inner proxy connect timeout must be positive")

	return errors.Join(errs...)
}

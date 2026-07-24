package access

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/beclab/Olares/cli/pkg/olares"
)

// probe timeouts. LAN / intranet are local hops and should fail fast; the
// public probe gets a little more room for a real round-trip to the edge.
const (
	probeTimeoutLAN      = 2 * time.Second
	probeTimeoutHost     = 2 * time.Second
	probeTimeoutExternal = 3 * time.Second
)

// MaxProbeDuration is the worst-case wall time of a full ProbeLocation run
// (every method tried sequentially). Callers that bound a reprobe with a
// context should derive their budget from this rather than hard-coding a value
// that silently truncates the external probe when the timeouts change.
func MaxProbeDuration() time.Duration {
	return probeTimeoutLAN + probeTimeoutHost + probeTimeoutExternal
}

// vpnNet is the parsed VPNSubnet, computed once. nil only if VPNSubnet is ever
// made malformed (compile-time constant, so effectively never).
var vpnNet = func() *net.IPNet {
	_, n, err := net.ParseCIDR(VPNSubnet)
	if err != nil {
		return nil
	}
	return n
}()

// probeFn is the single-probe function ProbeLocation drives. It indirects
// through a package var purely so tests can substitute a deterministic stub
// (the real probeOnce dials the network); production never reassigns it.
var probeFn = probeOnce

// ProbeLocation determines where the CLI sits relative to id's Olares instance
// by trying each connection method in order and returning the first that yields
// any HTTP response:
//
//  1. lan       — http://<svc>.<local>.olares.local
//  2. host/cluster — https://<svc>.<terminus> resolved via the in-cluster DNS;
//     a connection source IP inside VPNSubnet means `host`, otherwise `cluster`
//  3. external  — https://<svc>.<terminus> via the system resolver
//
// "Reachable" means the probe established a connection and got back any HTTP
// status (including 3xx/4xx) — auth/permission is irrelevant here. When every
// probe fails it returns ("", *UnreachableError) carrying the last failure's
// classification for messaging.
//
// localPrefix is the dev-only URL label (pass "" in production); insecure
// mirrors the profile's TLS opt-in.
func ProbeLocation(ctx context.Context, id olares.ID, localPrefix string, insecure bool) (olares.Location, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	lastKind := KindOther
	localNetDown := 0

	// 1. LAN.
	lanURL := id.Endpoints(olares.LocationLAN, localPrefix).Desktop
	if err := probeFn(ctx, olares.LocationLAN, lanURL, insecure, nil, probeTimeoutLAN); err == nil {
		return olares.LocationLAN, nil
	} else {
		lastKind = classifyNetErr(err)
		if lastKind == KindLocalNetDown {
			localNetDown++
		}
	}

	// 2. host / cluster — same URL, intranet DNS, distinguished by source IP.
	intranetURL := id.Endpoints(olares.LocationHost, localPrefix).Desktop
	var srcIP net.IP
	if err := probeFn(ctx, olares.LocationHost, intranetURL, insecure, &srcIP, probeTimeoutHost); err == nil {
		return locationFromSrcIP(srcIP), nil
	} else {
		lastKind = classifyNetErr(err)
		if lastKind == KindLocalNetDown {
			localNetDown++
		}
	}

	// If both local-hop probes failed because the local network stack / route
	// is down (ENETUNREACH / EHOSTUNREACH), the public probe over that same
	// dead stack is hopeless — short-circuit instead of waiting it out.
	if localNetDown == 2 {
		return "", &UnreachableError{OlaresID: id.String(), LastKind: KindLocalNetDown}
	}

	// 3. external.
	extURL := id.Endpoints(olares.LocationExternal, localPrefix).Desktop
	if err := probeFn(ctx, olares.LocationExternal, extURL, insecure, nil, probeTimeoutExternal); err == nil {
		return olares.LocationExternal, nil
	} else {
		lastKind = classifyNetErr(err)
	}

	return "", &UnreachableError{OlaresID: id.String(), LastKind: lastKind}
}

// locationFromSrcIP maps the source address of a successful intranet probe to
// a position: an address inside VPNSubnet means the CLI is on the Olares host
// (and needs the in-cluster DNS resolver), otherwise it's inside a cluster pod
// (which already has cluster DNS). A nil source IP defaults to cluster — the
// safer of the two, since cluster uses the plain (system) resolver.
func locationFromSrcIP(srcIP net.IP) olares.Location {
	if srcIP != nil && vpnNet != nil && vpnNet.Contains(srcIP) {
		return olares.LocationHost
	}
	return olares.LocationCluster
}

// probeOnce performs a single reachability probe of rawURL using a transport
// configured for loc. When srcIP is non-nil, the dialer records the local
// (source) address of the established connection into it — used by the
// host/cluster discrimination. Returns nil on any HTTP response, or the
// transport error when the connection could not be established.
func probeOnce(ctx context.Context, loc olares.Location, rawURL string, insecure bool, srcIP *net.IP, timeout time.Duration) error {
	// Reuse the same transport builder runtime clients use, so a probe can't
	// drift from how loc actually connects (cluster resolver for host, the
	// insecure TLS opt-in, etc.). The per-probe timeout is enforced by reqCtx
	// + client.Timeout below rather than the dialer.
	tr := Transport(loc, insecure)
	if srcIP != nil {
		inner := tr.DialContext
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := inner(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			if la, ok := conn.LocalAddr().(*net.TCPAddr); ok {
				*srcIP = la.IP
			}
			return conn, nil
		}
	}
	defer tr.CloseIdleConnections()

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
		// Any HTTP status counts as "reachable"; don't chase redirects.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if resp, herr := doProbe(reqCtx, client, http.MethodHead, rawURL); herr == nil {
		resp.Body.Close()
		return nil
	} else if classifyNetErr(herr) != KindOther {
		// A definitive connection-level failure (DNS / refused / net-down /
		// timeout / TLS / caller-cancel): GET over the same path would fail
		// identically, so don't burn the remaining budget on it. Only the
		// "unclassified" bucket (e.g. EOF / RST, which some edges return for
		// HEAD specifically) is worth a single GET fallback.
		return herr
	}
	resp, err := doProbe(reqCtx, client, http.MethodGet, rawURL)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func doProbe(ctx context.Context, client *http.Client, method, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

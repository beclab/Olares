package access

import (
	"context"
	"net"
	"net/http"
	"os"
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

// connAddrs records both ends of the connection a probe established. The
// intranet probe classifies its position from these: `remote` says whether the
// in-cluster DNS actually handed back an intranet address, `src` whether we
// reached it over the overlay.
type connAddrs struct {
	src    net.IP
	remote net.IP
}

// serviceAccountTokenPath is projected into every pod that mounts a service
// account. It exists only inside the pod's mount namespace, never on the node
// filesystem, which is what makes it a usable "am I in a pod" signal. A var
// only so tests can point it at a temp file; production never reassigns it.
var serviceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// inClusterPod indirects through a package var so tests can force either
// answer; production never reassigns it.
var inClusterPod = runningInClusterPod

// runningInClusterPod reports whether this process is running inside a
// Kubernetes pod. Both signals are injected by the kubelet: the env var on
// every pod in a cluster, the token file on every pod that mounts a service
// account (the env var alone is enough, the file covers the exotic case of a
// pod with the env stripped).
func runningInClusterPod() bool {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}
	_, err := os.Stat(serviceAccountTokenPath)
	return err == nil
}

// ProbeLocation determines where the CLI sits relative to id's Olares instance
// by trying each connection method in order and returning the first that yields
// any HTTP response:
//
//  1. lan       — http://<svc>.<local>.olares.local
//  2. host/cluster — https://<svc>.<terminus> resolved via the in-cluster DNS.
//     Reaching it is not by itself evidence of an intranet position, so the
//     answer is classified from the connection's addresses — see
//     locationFromProbe. An inconclusive success falls through to (3), but is
//     remembered: if (3) then fails, it is returned as `host` rather than
//     discarded, since it remains the one route we've actually seen work.
//  3. external  — https://<svc>.<terminus> via the system resolver
//
// "Reachable" means the probe established a connection and got back any HTTP
// status (including 3xx/4xx) — auth/permission is irrelevant here. Only when
// no probe reached the instance at all does it return ("", *UnreachableError),
// carrying the last failure's classification for messaging.
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

	// 2. host / cluster — same URL, intranet DNS. A success here is NOT on its
	// own evidence of an intranet position: on any machine that happens to be
	// a k8s node, ClusterDNS answers, forwards the unknown <svc>.<terminus>
	// upstream, and the probe then succeeds over the public path. Fall through
	// to the external probe when the addresses don't corroborate a position —
	// that public path is exactly what `external` is.
	intranetURL := id.Endpoints(olares.LocationHost, localPrefix).Desktop
	var (
		addrs           connAddrs
		intranetReached bool
	)
	if err := probeFn(ctx, olares.LocationHost, intranetURL, insecure, &addrs, probeTimeoutHost); err == nil {
		if loc, ok := locationFromProbe(addrs); ok {
			return loc, nil
		}
		intranetReached = true
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

	if intranetReached {
		// The external probe was meant to confirm the public path the
		// inconclusive intranet probe appeared to ride, and it didn't — but
		// that intranet probe did get an HTTP response, so we have first-hand
		// evidence of a working route and no business calling the instance
		// unreachable. The two methods differ only in how the name is
		// resolved, so keep the resolver that answered: whatever made the
		// system one fail here (no/blocked resolver, a blip within the 3s
		// budget) would fail the same way for every later request.
		return olares.LocationHost, nil
	}
	return "", &UnreachableError{OlaresID: id.String(), LastKind: lastKind}
}

// locationFromProbe classifies a successful intranet probe, reporting false
// when the connection says nothing about our position.
//
// The decisive question is whether the in-cluster DNS answered from its own
// zone or merely forwarded the lookup upstream, and the resolved address is
// what settles it. Olares' CoreDNS hands back an intranet address (the
// l4-bfl-proxy pod IP, the control-plane node IP, or the user's overlay
// address, depending on which view matches the querying client), while an
// unrelated cluster's CoreDNS forwards and yields the instance's public
// address. So:
//
//   - source inside VPNSubnet — we reached it over the overlay, which only
//     happens from the Olares host;
//   - otherwise, inside a pod — `cluster`, whose runtime resolver is the pod's
//     own /etc/resolv.conf rather than an explicit ClusterDNS dial;
//   - otherwise, an intranet remote address — `host`: ClusterDNS gave us a
//     private route to the instance, so the runtime client should keep using
//     it instead of going out over the public path;
//   - otherwise inconclusive: the probe rode the public path.
func locationFromProbe(a connAddrs) (olares.Location, bool) {
	if isOverlayIP(a.src) {
		return olares.LocationHost, true
	}
	if inClusterPod() {
		return olares.LocationCluster, true
	}
	if isIntranetIP(a.remote) {
		return olares.LocationHost, true
	}
	return "", false
}

// isOverlayIP reports whether ip belongs to the headscale/tailscale overlay.
func isOverlayIP(ip net.IP) bool {
	return ip != nil && vpnNet != nil && vpnNet.Contains(ip)
}

// isIntranetIP reports whether ip is anything other than a public unicast
// address: RFC1918 / RFC4193, loopback, link-local, or the overlay's CGNAT
// range (which net.IP.IsPrivate deliberately excludes).
func isIntranetIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || isOverlayIP(ip)
}

// probeOnce performs a single reachability probe of rawURL using a transport
// configured for loc. When addrs is non-nil, the dialer records both ends of
// the established connection into it — used by the host/cluster
// discrimination. Returns nil on any HTTP response, or the transport error
// when the connection could not be established.
func probeOnce(ctx context.Context, loc olares.Location, rawURL string, insecure bool, addrs *connAddrs, timeout time.Duration) error {
	// Reuse the same transport builder runtime clients use, so a probe can't
	// drift from how loc actually connects (cluster resolver for host, the
	// insecure TLS opt-in, etc.). The per-probe timeout is enforced by reqCtx
	// + client.Timeout below rather than the dialer.
	tr := Transport(loc, insecure)
	if addrs != nil {
		inner := tr.DialContext
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := inner(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			if la, ok := conn.LocalAddr().(*net.TCPAddr); ok {
				addrs.src = la.IP
			}
			if ra, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
				addrs.remote = ra.IP
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

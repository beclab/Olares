package access

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/beclab/Olares/cli/pkg/olares"
)

// VPNSubnet is the CGNAT range Olares' headscale/tailscale overlay hands out
// (USER_SUBNET defaults to 100.64.0.0/20, comfortably inside this /10). A
// connection sourced from this range was made over the overlay, which only
// happens from the Olares host. See locationFromProbe.
const VPNSubnet = "100.64.0.0/10"

// ClusterSubnet covers the cluster's own addresses — both the Service CIDR and
// the Pod CIDR, which kubekey derives from this same /16 by default
// (10.233.0.0/18 for Services, 10.233.64.0/18 for Pods). A connection sourced
// from here was made from inside a cluster pod. Note that a cluster installed
// with custom CIDRs falls outside this constant, in which case the probe simply
// classifies from the remote address instead. See locationFromProbe.
const ClusterSubnet = "10.233.0.0/16"

// clusterResolver resolves names through the in-cluster DNS (olares.ClusterDNS)
// rather than the system resolver, so the public `<svc>.<terminus>` hostnames
// resolve to intranet IPs from the Olares host. Mirrors
// daemon/pkg/utils/cluster_api.go::GetClusterHttpClient.
//
// It dials only UDP/53 with no TCP fallback. That's sufficient for the small
// A-record answers these single-host lookups return (and matches the daemon
// reference); a truncated/oversized response would fail rather than retry over
// TCP, which we accept for probe/runtime simplicity.
func clusterResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", net.JoinHostPort(olares.ClusterDNS, "53"))
		},
	}
}

// Transport builds an *http.Transport configured for loc. The `host` Location
// gets a dialer whose resolver points at the in-cluster DNS; every other
// position uses the system resolver (cluster pods already inherit cluster DNS
// via /etc/resolv.conf, and external/lan want the public/LAN answer). insecure
// disables TLS verification (dev-only profile opt-in).
func Transport(loc olares.Location, insecure bool) *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	switch {
	case loc.UsesClusterResolver():
		tr.DialContext = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver:  clusterResolver(),
		}).DialContext
	case loc.UsesLocalDomain():
		// The .olares.local names resolve over mDNS, where a missing record
		// draws no response at all — the query just waits out the full ~5s
		// window. A dual-stack lookup therefore pays that 5s for the absent
		// AAAA even when the A record came back in milliseconds, which alone
		// exceeds probeTimeoutLAN. Pin the dial to IPv4 so only the A record
		// is looked up.
		d := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
		tr.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
			return d.DialContext(ctx, "tcp4", addr)
		}
	}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	return tr
}

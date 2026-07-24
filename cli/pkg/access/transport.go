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
// (USER_SUBNET defaults to 100.64.0.0/20, comfortably inside this /10). When a
// connection to the public hostname — resolved via the in-cluster DNS — has a
// source address in this range, the CLI is running on the Olares host itself;
// otherwise it's inside a cluster pod. See ProbeLocation.
const VPNSubnet = "100.64.0.0/10"

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
	if loc.UsesClusterResolver() {
		tr.DialContext = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver:  clusterResolver(),
		}).DialContext
	}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	return tr
}

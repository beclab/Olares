package access

import (
	"net"
	"testing"

	"github.com/beclab/Olares/cli/pkg/olares"
)

func TestMaxProbeDuration(t *testing.T) {
	want := probeTimeoutLAN + probeTimeoutHost + probeTimeoutExternal
	if got := MaxProbeDuration(); got != want {
		t.Errorf("MaxProbeDuration() = %v, want %v", got, want)
	}
}

func TestVPNSubnetParses(t *testing.T) {
	if vpnNet == nil {
		t.Fatal("vpnNet failed to parse VPNSubnet")
	}
	in := []string{"100.64.0.1", "100.96.5.5", "100.127.255.254"}
	for _, s := range in {
		if !vpnNet.Contains(net.ParseIP(s)) {
			t.Errorf("%s should be inside %s", s, VPNSubnet)
		}
	}
	out := []string{"10.0.0.1", "192.168.1.1", "100.128.0.1", "8.8.8.8"}
	for _, s := range out {
		if vpnNet.Contains(net.ParseIP(s)) {
			t.Errorf("%s should NOT be inside %s", s, VPNSubnet)
		}
	}
}

func TestLocationFromSrcIP(t *testing.T) {
	// Inside the VPN CGNAT range → we're on the host and need cluster DNS.
	if got := locationFromSrcIP(net.ParseIP("100.64.0.7")); got != olares.LocationHost {
		t.Errorf("VPN source IP → %q, want host", got)
	}
	// A regular pod address → cluster (system/inherited resolver).
	if got := locationFromSrcIP(net.ParseIP("10.233.1.5")); got != olares.LocationCluster {
		t.Errorf("pod source IP → %q, want cluster", got)
	}
	// Unknown source IP defaults to the safer cluster position.
	if got := locationFromSrcIP(nil); got != olares.LocationCluster {
		t.Errorf("nil source IP → %q, want cluster", got)
	}
}

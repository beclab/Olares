package access

import (
	"net"
	"os"
	"path/filepath"
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

func TestRunningInClusterPod(t *testing.T) {
	// Point the token path at a non-existent file so the env var is the only
	// signal in play, then at a real one to cover the file branch.
	missing := filepath.Join(t.TempDir(), "token")
	orig := serviceAccountTokenPath
	t.Cleanup(func() { serviceAccountTokenPath = orig })
	serviceAccountTokenPath = missing

	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	if runningInClusterPod() {
		t.Error("no env var and no token file → want false")
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "10.233.0.1")
	if !runningInClusterPod() {
		t.Error("KUBERNETES_SERVICE_HOST set → want true")
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	if err := os.WriteFile(missing, []byte("tok"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !runningInClusterPod() {
		t.Error("service account token present → want true")
	}
}

func TestLocationFromProbe(t *testing.T) {
	ip := net.ParseIP

	cases := []struct {
		name   string
		addrs  connAddrs
		inPod  bool
		want   olares.Location
		wantOK bool
	}{
		{
			name:   "overlay source address → host",
			addrs:  connAddrs{src: ip("100.64.0.7"), remote: ip("100.64.0.1")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			// The Olares node: ClusterDNS answers with a node-local address,
			// so only the remote end reveals the intranet hit.
			name:   "private remote address → host",
			addrs:  connAddrs{src: ip("192.168.31.104"), remote: ip("192.168.31.104")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			name:   "in a pod → cluster, whatever the addresses",
			addrs:  connAddrs{src: ip("10.233.1.5"), remote: ip("10.233.1.9")},
			inPod:  true,
			want:   olares.LocationCluster,
			wantOK: true,
		},
		{
			// ClusterDNS forwarded the lookup upstream: the probe rode the
			// public path and says nothing about our position.
			name:  "public remote address → inconclusive",
			addrs: connAddrs{src: ip("192.168.1.20"), remote: ip("42.193.109.3")},
		},
		{
			name:  "no addresses recorded → inconclusive",
			addrs: connAddrs{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setInClusterPod(t, tc.inPod)
			got, ok := locationFromProbe(tc.addrs)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("locationFromProbe(%+v) = (%q, %v), want (%q, %v)", tc.addrs, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestIsIntranetIP(t *testing.T) {
	intranet := []string{
		"192.168.31.104", // RFC1918
		"10.233.1.9",     // RFC1918 (pod / service range)
		"172.16.0.5",     // RFC1918
		"100.64.0.1",     // overlay CGNAT — not covered by net.IP.IsPrivate
		"127.0.0.1",      // loopback
		"169.254.25.10",  // link-local (nodelocaldns)
		"fd00::1",        // RFC4193
	}
	for _, s := range intranet {
		if !isIntranetIP(net.ParseIP(s)) {
			t.Errorf("%s should be intranet", s)
		}
	}

	public := []string{"42.193.109.3", "8.8.8.8", "2001:4860:4860::8888"}
	for _, s := range public {
		if isIntranetIP(net.ParseIP(s)) {
			t.Errorf("%s should NOT be intranet", s)
		}
	}

	if isIntranetIP(nil) {
		t.Error("nil should NOT be intranet")
	}
}

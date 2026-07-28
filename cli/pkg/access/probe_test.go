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

func TestClusterSubnetParses(t *testing.T) {
	if clusterNet == nil {
		t.Fatal("clusterNet failed to parse ClusterSubnet")
	}
	in := []string{"10.233.0.1", "10.233.64.1", "10.233.104.86", "10.233.255.254"}
	for _, s := range in {
		if !clusterNet.Contains(net.ParseIP(s)) {
			t.Errorf("%s should be inside %s", s, ClusterSubnet)
		}
	}
	out := []string{"10.232.0.1", "10.234.0.1", "192.168.1.1", "100.64.0.1"}
	for _, s := range out {
		if clusterNet.Contains(net.ParseIP(s)) {
			t.Errorf("%s should NOT be inside %s", s, ClusterSubnet)
		}
	}
}

func TestLocationFromProbe(t *testing.T) {
	cases := []struct {
		name   string
		addrs  connAddrs
		want   olares.Location
		wantOK bool
	}{
		{
			name:   "overlay src -> host",
			addrs:  connAddrs{src: net.ParseIP("100.64.0.7"), remote: net.ParseIP("100.64.0.1")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			name:   "pod src -> cluster",
			addrs:  connAddrs{src: net.ParseIP("10.233.104.86"), remote: net.ParseIP("192.168.31.104")},
			want:   olares.LocationCluster,
			wantOK: true,
		},
		{
			name:   "service-cidr src -> cluster",
			addrs:  connAddrs{src: net.ParseIP("10.233.1.5"), remote: net.ParseIP("10.233.0.3")},
			want:   olares.LocationCluster,
			wantOK: true,
		},
		{
			name:   "intranet remote (node IP) -> host",
			addrs:  connAddrs{src: net.ParseIP("192.168.50.202"), remote: net.ParseIP("192.168.31.104")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			name:   "intranet remote (RFC1918) with nil src -> host",
			addrs:  connAddrs{remote: net.ParseIP("10.0.0.5")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			name:   "overlay remote without overlay src -> host via intranet",
			addrs:  connAddrs{src: net.ParseIP("192.168.1.10"), remote: net.ParseIP("100.64.0.1")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			name:   "public remote + LAN src -> inconclusive",
			addrs:  connAddrs{src: net.ParseIP("192.168.50.202"), remote: net.ParseIP("42.193.109.3")},
			wantOK: false,
		},
		{
			name:   "nil addrs -> inconclusive",
			addrs:  connAddrs{},
			wantOK: false,
		},
		{
			// Overlay src wins even when remote looks public (DNS hijack /
			// unexpected answer); we still reached it over the overlay.
			name:   "overlay src beats public remote",
			addrs:  connAddrs{src: net.ParseIP("100.64.0.7"), remote: net.ParseIP("8.8.8.8")},
			want:   olares.LocationHost,
			wantOK: true,
		},
		{
			// Pod src wins over an intranet remote: the CLI is in a pod, so
			// runtime should use the inherited resolv.conf, not ClusterDNS.
			name:   "pod src beats intranet remote",
			addrs:  connAddrs{src: net.ParseIP("10.233.64.10"), remote: net.ParseIP("192.168.31.104")},
			want:   olares.LocationCluster,
			wantOK: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := locationFromProbe(tc.addrs)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (loc=%q)", ok, tc.wantOK, got)
			}
			if ok && got != tc.want {
				t.Errorf("loc = %q, want %q", got, tc.want)
			}
			if !ok && got != "" {
				t.Errorf("inconclusive loc = %q, want empty", got)
			}
		})
	}
}

func TestIsIntranetIP(t *testing.T) {
	yes := []string{"10.0.0.1", "192.168.1.1", "172.16.5.5", "127.0.0.1", "100.64.0.1", "169.254.1.1"}
	for _, s := range yes {
		if !isIntranetIP(net.ParseIP(s)) {
			t.Errorf("isIntranetIP(%s) = false, want true", s)
		}
	}
	no := []string{"8.8.8.8", "42.193.109.3", "1.1.1.1"}
	for _, s := range no {
		if isIntranetIP(net.ParseIP(s)) {
			t.Errorf("isIntranetIP(%s) = true, want false", s)
		}
	}
	if isIntranetIP(nil) {
		t.Error("isIntranetIP(nil) = true, want false")
	}
}

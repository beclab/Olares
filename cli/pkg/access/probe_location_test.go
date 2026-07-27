package access

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/olares"
)

// probeOutcome scripts one probe result: err==nil means "reachable"; srcIP /
// remoteIP (host probe only) are written back so locationFromProbe gets
// exercised.
type probeOutcome struct {
	err      error
	srcIP    net.IP
	remoteIP net.IP
}

// installProbeStub swaps probeFn for a deterministic, per-Location script and
// records the call order. The real probeFn is restored on cleanup.
func installProbeStub(t *testing.T, outcomes map[olares.Location]probeOutcome) *[]olares.Location {
	t.Helper()
	calls := &[]olares.Location{}
	orig := probeFn
	t.Cleanup(func() { probeFn = orig })
	probeFn = func(_ context.Context, loc olares.Location, _ string, _ bool, addrs *connAddrs, _ time.Duration) error {
		*calls = append(*calls, loc)
		out, ok := outcomes[loc]
		if !ok {
			// Default: unreachable (connection refused).
			return syscall.ECONNREFUSED
		}
		if out.err == nil && addrs != nil {
			addrs.src, addrs.remote = out.srcIP, out.remoteIP
		}
		return out.err
	}
	return calls
}

// setInClusterPod forces the "am I in a pod" signal for one test.
func setInClusterPod(t *testing.T, v bool) {
	t.Helper()
	orig := inClusterPod
	t.Cleanup(func() { inClusterPod = orig })
	inClusterPod = func() bool { return v }
}

func sameLocs(a, b []olares.Location) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestProbeLocationOrdering(t *testing.T) {
	id, _ := olares.ParseID("alice@olares.com")

	cases := []struct {
		name      string
		outcomes  map[olares.Location]probeOutcome
		inPod     bool
		want      olares.Location
		wantErr   bool
		wantCalls []olares.Location
	}{
		{
			name:      "lan wins first",
			outcomes:  map[olares.Location]probeOutcome{olares.LocationLAN: {}},
			want:      olares.LocationLAN,
			wantCalls: []olares.Location{olares.LocationLAN},
		},
		{
			name: "host: intranet reachable over the overlay",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("100.64.0.9"), remoteIP: net.ParseIP("100.64.0.1")},
			},
			want:      olares.LocationHost,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			// The Olares node itself: ClusterDNS resolves the public hostname
			// to a node-local address, so nothing rides the overlay and the
			// remote address is the only signal.
			name: "host: ClusterDNS resolved to an intranet address",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("192.168.31.104"), remoteIP: net.ParseIP("192.168.31.104")},
			},
			want:      olares.LocationHost,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			name: "cluster: inside a pod, reaching the l4 proxy",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("10.233.1.2"), remoteIP: net.ParseIP("10.233.1.9")},
			},
			inPod:     true,
			want:      olares.LocationCluster,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			// A bystander k8s node talking to someone else's Olares: its own
			// ClusterDNS knows nothing of that zone, forwards the lookup
			// upstream, and the probe lands on the instance's public address.
			// Must not be mistaken for an intranet position.
			name: "external: ClusterDNS forwarded upstream to a public address",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {srcIP: net.ParseIP("192.168.1.20"), remoteIP: net.ParseIP("42.193.109.3")},
				olares.LocationExternal: {},
			},
			want:      olares.LocationExternal,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			// Same, but with no addresses recorded at all.
			name: "external: intranet probe succeeds with no addresses recorded",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {},
				olares.LocationExternal: {},
			},
			want:      olares.LocationExternal,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			// Same inconclusive intranet success, but the confirming public
			// probe fails (system resolver blocked, or a blip inside its 3s
			// budget). We already got an HTTP response through ClusterDNS, so
			// declaring the instance unreachable would throw away the only
			// route we have evidence for.
			name: "host: inconclusive intranet success survives a failed external probe",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {srcIP: net.ParseIP("192.168.1.20"), remoteIP: net.ParseIP("42.193.109.3")},
				olares.LocationExternal: {err: syscall.ECONNREFUSED},
			},
			want:      olares.LocationHost,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			name: "external fallback",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {err: syscall.ECONNREFUSED},
				olares.LocationExternal: {},
			},
			want:      olares.LocationExternal,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			name: "all fail -> unreachable",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {err: syscall.ECONNREFUSED},
				olares.LocationExternal: {err: syscall.ECONNREFUSED},
			},
			wantErr:   true,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			name: "local net down on lan+host short-circuits external",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ENETUNREACH},
				olares.LocationHost: {err: syscall.EHOSTUNREACH},
			},
			wantErr:   true,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setInClusterPod(t, tc.inPod)
			calls := installProbeStub(t, tc.outcomes)
			got, err := ProbeLocation(context.Background(), id, "", false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got loc=%q", got)
				}
				if !IsUnreachable(err) {
					t.Errorf("expected *UnreachableError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("loc = %q, want %q", got, tc.want)
				}
			}
			if !sameLocs(*calls, tc.wantCalls) {
				t.Errorf("probe call sequence = %v, want %v", *calls, tc.wantCalls)
			}
		})
	}
}

// TestProbeLocationLocalNetDownLastKind verifies the short-circuit reports a
// local-network LastKind so the message leans toward "your network is down".
func TestProbeLocationLocalNetDownLastKind(t *testing.T) {
	id, _ := olares.ParseID("alice@olares.com")
	installProbeStub(t, map[olares.Location]probeOutcome{
		olares.LocationLAN:  {err: syscall.ENETUNREACH},
		olares.LocationHost: {err: syscall.ENETUNREACH},
	})
	_, err := ProbeLocation(context.Background(), id, "", false)
	var ue *UnreachableError
	if !errors.As(err, &ue) {
		t.Fatalf("expected *UnreachableError, got %v", err)
	}
	if ue.LastKind != KindLocalNetDown {
		t.Errorf("LastKind = %d, want KindLocalNetDown (%d)", ue.LastKind, KindLocalNetDown)
	}
}

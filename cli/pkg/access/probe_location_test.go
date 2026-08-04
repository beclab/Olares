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
// remoteIP (host probe only) are written into *connAddrs so locationFromProbe
// gets exercised.
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
			if out.srcIP != nil {
				addrs.src = out.srcIP
			}
			if out.remoteIP != nil {
				addrs.remote = out.remoteIP
			}
		}
		return out.err
	}
	return calls
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
			name: "host: intranet reachable with VPN source IP",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("100.64.0.9"), remoteIP: net.ParseIP("100.64.0.1")},
			},
			want:      olares.LocationHost,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			name: "cluster: intranet reachable with pod source IP",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("10.233.104.86"), remoteIP: net.ParseIP("192.168.31.104")},
			},
			want:      olares.LocationCluster,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			name: "host: LAN src with intranet remote (node IP)",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:  {err: syscall.ECONNREFUSED},
				olares.LocationHost: {srcIP: net.ParseIP("192.168.50.202"), remoteIP: net.ParseIP("192.168.31.104")},
			},
			want:      olares.LocationHost,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost},
		},
		{
			// ClusterDNS answered with a public address (hijacked / forwarded
			// upstream) — inconclusive, so ProbeLocation continues to external.
			name: "inconclusive host + external reachable -> external",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {srcIP: net.ParseIP("192.168.50.202"), remoteIP: net.ParseIP("42.193.109.3")},
				olares.LocationExternal: {},
			},
			want:      olares.LocationExternal,
			wantCalls: []olares.Location{olares.LocationLAN, olares.LocationHost, olares.LocationExternal},
		},
		{
			// Same inconclusive host probe, but external also fails — current
			// code surfaces UnreachableError (no host fallback).
			name: "inconclusive host + external fail -> unreachable",
			outcomes: map[olares.Location]probeOutcome{
				olares.LocationLAN:      {err: syscall.ECONNREFUSED},
				olares.LocationHost:     {srcIP: net.ParseIP("192.168.50.202"), remoteIP: net.ParseIP("42.193.109.3")},
				olares.LocationExternal: {err: syscall.ECONNREFUSED},
			},
			wantErr:   true,
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

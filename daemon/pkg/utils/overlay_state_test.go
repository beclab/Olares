package utils

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOverlayGatewayDesiredRoundTrip(t *testing.T) {
	orig := OverlayDesiredStateFile
	defer func() { OverlayDesiredStateFile = orig }()
	OverlayDesiredStateFile = filepath.Join(t.TempDir(), "overlay-gateway", "enabled")

	if OverlayGatewayDesired() {
		t.Fatal("fresh state dir must read as disabled")
	}
	if err := SetOverlayGatewayDesired(true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if !OverlayGatewayDesired() {
		t.Fatal("state must read as enabled after SetOverlayGatewayDesired(true)")
	}
	if err := SetOverlayGatewayDesired(true); err != nil {
		t.Fatalf("enable must be idempotent: %v", err)
	}
	if err := SetOverlayGatewayDesired(false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if OverlayGatewayDesired() {
		t.Fatal("state must read as disabled after SetOverlayGatewayDesired(false)")
	}
	if err := SetOverlayGatewayDesired(false); err != nil {
		t.Fatalf("disable must be idempotent: %v", err)
	}
}

func TestKernelSupportsAltname(t *testing.T) {
	cases := map[string]bool{
		"6.14.0-37-generic": true,
		"5.15.0-91-generic": true,
		"5.5.0":             true,
		"5.4.0-150-generic": false,
		"4.19.0":            false,
		"garbage":           false,
	}
	for release, want := range cases {
		if got := kernelSupportsAltname(release); got != want {
			t.Errorf("kernelSupportsAltname(%q) = %v, want %v", release, got, want)
		}
	}
}

func TestOverlayUdevRuleContent(t *testing.T) {
	got := overlayUdevRuleContent("d8:43:ae:af:5a:33", "/bin/ip")
	for _, want := range []string{`ACTION=="add"`, `SUBSYSTEM=="net"`, `ATTR{address}=="d8:43:ae:af:5a:33"`, `RUN+="/bin/ip link property add dev $env{INTERFACE} altname olares-lan"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("udev rule %q lacks %q", got, want)
		}
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatal("rule file must end with a newline")
	}
}

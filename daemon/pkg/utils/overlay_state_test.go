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

func TestOverlayLinkFileContent(t *testing.T) {
	got := overlayLinkFileContent("d8:43:ae:af:5a:33")
	for _, want := range []string{"[Match]", "MACAddress=d8:43:ae:af:5a:33", "Type=ether", "[Link]", "AlternativeName=olares-lan"} {
		if !strings.Contains(got, want) {
			t.Fatalf("link file %q lacks %q", got, want)
		}
	}
}

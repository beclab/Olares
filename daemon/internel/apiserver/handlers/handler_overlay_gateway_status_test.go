package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// stubOverlayNodeFacts replaces the node facts the status derivation reads and
// restores them when the test ends.
func stubOverlayNodeFacts(t *testing.T, desired bool, dev string, up bool, linkErr error, dhcp, kernelOK, ethernet bool) {
	t.Helper()
	origDesired, origLink, origDhcp, origKernel, origWSL, origDarwin, origEth :=
		overlayGatewayDesired, overlayParentLinkUp, isCniDhcpActive, kernelSupportsAltname, isWSL, isDarwin, isEthernetConnected
	t.Cleanup(func() {
		overlayGatewayDesired, overlayParentLinkUp, isCniDhcpActive, kernelSupportsAltname, isWSL, isDarwin, isEthernetConnected =
			origDesired, origLink, origDhcp, origKernel, origWSL, origDarwin, origEth
	})
	overlayGatewayDesired = func() bool { return desired }
	overlayParentLinkUp = func(context.Context) (string, bool, error) { return dev, up, linkErr }
	isCniDhcpActive = func(context.Context) bool { return dhcp }
	kernelSupportsAltname = func() bool { return kernelOK }
	isWSL = func() bool { return false }
	isDarwin = func() bool { return false }
	isEthernetConnected = func(context.Context) bool { return ethernet }
}

func TestGetOverlayGatewayStatusDerivation(t *testing.T) {
	cases := []struct {
		name            string
		desired         bool
		dev             string
		up              bool
		linkErr         error
		dhcp            bool
		kernelOK        bool
		ethernet        bool
		wantStatus      string
		wantDisable     bool
		wantReasonPart  string
		wantMessagePart string
	}{
		{name: "off and supported", kernelOK: true, ethernet: true, dhcp: true, wantStatus: OverlayGatewayOff},
		{name: "off with old kernel", kernelOK: false, ethernet: true, wantStatus: OverlayGatewayOff, wantDisable: true, wantReasonPart: "Kernel"},
		{name: "off without ethernet", kernelOK: true, ethernet: false, wantStatus: OverlayGatewayOff, wantDisable: true, wantReasonPart: "Ethernet"},
		{name: "desired but name missing", desired: true, linkErr: errors.New("Link not found"), kernelOK: true, ethernet: true, wantStatus: OverlayGatewayOff, wantMessagePart: "missing"},
		{name: "desired and parent down", desired: true, dev: "enp3s0", up: false, kernelOK: true, ethernet: true, wantStatus: OverlayGatewayOn, wantMessagePart: "enp3s0 is down"},
		{name: "desired and parent up", desired: true, dev: "enp3s0", up: true, dhcp: true, kernelOK: true, ethernet: true, wantStatus: OverlayGatewayOn},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubOverlayNodeFacts(t, tc.desired, tc.dev, tc.up, tc.linkErr, tc.dhcp, tc.kernelOK, tc.ethernet)
			s, err := (&Handlers{}).getOverlayGatewayStatus(context.Background())
			if err != nil {
				t.Fatalf("status: %v", err)
			}
			if s.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", s.Status, tc.wantStatus)
			}
			if s.Disable != tc.wantDisable {
				t.Fatalf("disable = %v, want %v (reason %q)", s.Disable, tc.wantDisable, s.DisableReason)
			}
			if tc.wantReasonPart != "" && !strings.Contains(s.DisableReason, tc.wantReasonPart) {
				t.Fatalf("disable reason %q must mention %q", s.DisableReason, tc.wantReasonPart)
			}
			if tc.wantMessagePart == "" && s.ErrorMessage != "" {
				t.Fatalf("unexpected error message %q", s.ErrorMessage)
			}
			if tc.wantMessagePart != "" && !strings.Contains(s.ErrorMessage, tc.wantMessagePart) {
				t.Fatalf("error message %q must mention %q", s.ErrorMessage, tc.wantMessagePart)
			}
			if s.CniDhcpActive != tc.dhcp {
				t.Fatalf("cni_dhcp_active = %v, want %v; it must be reported independently of the switch", s.CniDhcpActive, tc.dhcp)
			}
		})
	}
}

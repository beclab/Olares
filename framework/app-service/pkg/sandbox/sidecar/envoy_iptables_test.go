package sidecar

import (
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestGenerateIptablesCommandsSkipsLinkerdUID(t *testing.T) {
	cmd := generateIptablesCommands(nil, false)
	want := fmt.Sprintf("-A PROXY_OUTBOUND -m owner --uid-owner %d -j RETURN", constants.LinkerdProxyUID)
	if !strings.Contains(cmd, want) {
		t.Fatalf("iptables missing linkerd uid skip %q in:\n%s", want, cmd)
	}
	if !strings.Contains(cmd, "-A PROXY_OUTBOUND -m owner --uid-owner 1555 -j RETURN") {
		t.Fatalf("iptables missing envoy uid skip in:\n%s", cmd)
	}
}

func TestGenerateIptablesCommandsSkipsLinkerdInboundPorts(t *testing.T) {
	cmd := generateIptablesCommands(nil, false)
	want := fmt.Sprintf("-A PROXY_INBOUND -p tcp -m multiport --dports %d,%d,%d -j RETURN",
		constants.LinkerdTapPort, constants.LinkerdAdminPort, constants.LinkerdInboundPort)
	if !strings.Contains(cmd, want) {
		t.Fatalf("iptables missing linkerd inbound port skip %q in:\n%s", want, cmd)
	}
	catchAll := "-A PROXY_INBOUND -p tcp -j PROXY_IN_REDIRECT"
	idxSkip := strings.Index(cmd, want)
	idxCatch := strings.Index(cmd, catchAll)
	if idxSkip < 0 || idxCatch < 0 || idxSkip > idxCatch {
		t.Fatalf("linkerd inbound skip must precede catch-all redirect")
	}
}

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

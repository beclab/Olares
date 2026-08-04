package meshinagent

import (
	"regexp"
	"strings"
	"testing"
)

func TestIptablesInitScriptEveryRedirectHasDestination(t *testing.T) {
	script := IptablesInitScript()
	lines := strings.Split(script, "\n")
	redirectRe := regexp.MustCompile(`-j\s+REDIRECT`)
	destRe := regexp.MustCompile(`-d\s+"\$ip"`)
	var redirectLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if redirectRe.MatchString(trimmed) {
			redirectLines = append(redirectLines, trimmed)
			if !destRe.MatchString(trimmed) {
				t.Fatalf("REDIRECT without -d: %s", trimmed)
			}
		}
	}
	if len(redirectLines) < 2 {
		t.Fatalf("expected at least 80 and 443 REDIRECT lines, got %d", len(redirectLines))
	}
	has80, has443 := false, false
	for _, line := range redirectLines {
		if strings.Contains(line, "--dport 80") {
			has80 = true
		}
		if strings.Contains(line, "--dport 443") {
			has443 = true
		}
	}
	if !has80 || !has443 {
		t.Fatalf("missing dport coverage: 80=%v 443=%v lines=%v", has80, has443, redirectLines)
	}
}

func TestIptablesInitScriptEnvPreferredAndFailOpen(t *testing.T) {
	script := IptablesInitScript()
	for _, want := range []string{
		`MESH_IN_AGENT_GATEWAY_IPS`,
		`using gateway IPs from env`,
		`continuing without redirects`,
		`exit 0`,
		`attempt=1`,
		`while [ "$attempt" -le 3 ]`,
		`for ip in $GW_IPS`,
		`--dport 443`,
		`-d "$ip"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %q", want)
		}
	}
	if strings.Contains(script, "*:443") {
		t.Fatal("must not log blanket *:443")
	}
	if strings.Contains(script, "app-gateway-data.app-gateway.svc") {
		t.Fatal("must only resolve app-gateway-data from os-gateway")
	}
	if strings.Contains(script, "exit 1") {
		t.Fatal("resolve failure must not exit 1")
	}
	// No REDIRECT line may omit -d (guard against future port additions).
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, "-j REDIRECT") && !strings.Contains(line, `-d "$ip"`) {
			t.Fatalf("AC-2: REDIRECT without destination: %s", strings.TrimSpace(line))
		}
	}
}

func TestInitContainerSpecWithGatewayIPs(t *testing.T) {
	c := InitContainerSpecWithGatewayIPs("10.96.0.10,10.96.0.11")
	found := false
	for _, e := range c.Env {
		if e.Name == GatewayIPsEnv {
			found = true
			if e.Value != "10.96.0.10,10.96.0.11" {
				t.Fatalf("GatewayIPsEnv = %q", e.Value)
			}
		}
	}
	if !found {
		t.Fatal("expected MESH_IN_AGENT_GATEWAY_IPS env")
	}
	empty := InitContainerSpec()
	for _, e := range empty.Env {
		if e.Name == GatewayIPsEnv {
			t.Fatal("empty InitContainerSpec must omit GatewayIPsEnv")
		}
	}
}

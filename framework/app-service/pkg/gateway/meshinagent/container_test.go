package meshinagent

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestIptablesInitScript(t *testing.T) {
	for _, tc := range []struct {
		name, ips, dns, fail, check string
		wantSuccess                 bool
		wantRules                   []string
	}{
		{name: "dns exhaustion", dns: "empty"},
		{name: "dns recovery", dns: "retry", wantSuccess: true, wantRules: []string{"iptables 10.96.0.10 80 16080", "iptables 10.96.0.10 443 16443"}},
		{name: "snapshot preferred and dual stack", ips: "10.96.0.11,fd00::11 10.96.0.11", dns: "unused",
			wantSuccess: true, wantRules: []string{"iptables 10.96.0.11 80 16080", "iptables 10.96.0.11 443 16443",
				"ip6tables fd00::11 80 16080", "ip6tables fd00::11 443 16443"}},
		{name: "partial redirect retries", ips: "10.96.0.12", fail: "once", wantSuccess: true,
			wantRules: []string{"iptables 10.96.0.12 80 16080", "iptables 10.96.0.12 443 16443"}},
		{name: "persistent redirect failure", ips: "10.96.0.12", fail: "always",
			wantRules: []string{"iptables 10.96.0.12 80 16080"}},
		{name: "unverified rule is rejected", ips: "10.96.0.12", check: "reject",
			wantRules: []string{"iptables 10.96.0.12 80 16080", "iptables 10.96.0.12 80 16080", "iptables 10.96.0.12 80 16080"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write := func(name, code string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+code), 0700); err != nil {
					t.Fatal(err)
				}
			}
			write("dig", `
echo call >> "$TEST_DIR/dns"
test "$DNS" = retry && test "$(wc -l < "$TEST_DIR/dns")" -gt 4 && printf '10.96.0.10\n'
`)
			write("sleep", `echo retry >> "$TEST_DIR/retries"`)
			const firewall = `
tool=${0##*/}
mode= ip= port= dest=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -C|-I) mode=$1 ;;
    -d) shift; ip=$1 ;;
    --dport) shift; port=$1 ;;
    --to-ports) shift; dest=$1 ;;
  esac
  shift
done
test -n "$ip" && test -n "$port" && test -n "$dest" || exit 8
rule="$tool $ip $port $dest"
if [ "$mode" = -C ]; then
  test "$CHECK" != reject && grep -qFx "$rule" "$TEST_DIR/rules" 2>/dev/null
  exit $?
fi
test "$mode" = -I || exit 8
if [ "$port" = 443 ] && [ -n "$FAIL" ]; then
  if [ "$FAIL" = always ] || [ ! -e "$TEST_DIR/failed" ]; then
    touch "$TEST_DIR/failed"
    exit 1
  fi
fi
printf '%s\n' "$rule" >> "$TEST_DIR/rules"
`
			write("iptables", firewall)
			write("ip6tables", firewall)
			cmd := exec.Command("/bin/sh", "-c", IptablesInitScript())
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "TEST_DIR="+dir,
				"MESH_IN_AGENT_GATEWAY_IPS="+tc.ips, "DNS="+tc.dns, "FAIL="+tc.fail, "CHECK="+tc.check)
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantSuccess {
				t.Fatalf("exit = %v, output: %s", err, out)
			}
			data, readErr := os.ReadFile(filepath.Join(dir, "rules"))
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			if got := strings.FieldsFunc(strings.TrimSpace(string(data)), func(r rune) bool { return r == '\n' }); !slices.Equal(got, tc.wantRules) {
				t.Fatalf("installed rules = %v, want %v", got, tc.wantRules)
			}
			if tc.dns == "unused" {
				if _, err := os.Stat(filepath.Join(dir, "dns")); !os.IsNotExist(err) {
					t.Fatalf("DNS unexpectedly used for admission-provided IPs: %v", err)
				}
			}
			if tc.dns == "retry" || tc.fail == "once" || !tc.wantSuccess {
				if _, err := os.Stat(filepath.Join(dir, "retries")); err != nil {
					t.Fatalf("script did not retry: %v", err)
				}
			}
		})
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

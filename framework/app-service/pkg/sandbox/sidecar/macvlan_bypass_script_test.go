package sidecar

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// iptablesMock records rules per table/chain in files so the bypass script can
// be executed end-to-end without a real netfilter stack: -A appends, -I
// prepends, -C probes, -D removes, -S dumps (iptables-save style with -A lines).
const iptablesMock = `#!/bin/sh
if [ "${1:-}" = "--version" ]; then
  echo "iptables v1.8.8 (nf_tables)"
  exit 0
fi
table=filter
op=""
chain=""
while [ $# -gt 0 ]; do
  case "$1" in
    -t)
      table="$2"
      shift 2
      ;;
    -C|-D|-A|-S)
      op="$1"
      chain="$2"
      shift 2
      break
      ;;
    -I)
      op="-I"
      chain="$2"
      shift 2
      case "$1" in
        [0-9]*) shift ;;
      esac
      break
      ;;
    *)
      shift
      ;;
  esac
done
f="${IPT_STATE}/${table}_${chain}"
: > /dev/null
touch "$f"
rule="$*"
case "$op" in
  -C) grep -Fxq -- "$rule" "$f" ;;
  -D) grep -vFx -- "$rule" "$f" > "$f.tmp"; mv "$f.tmp" "$f" ;;
  -A) printf '%s\n' "$rule" >> "$f" ;;
  -I)
    if [ "${IPT_FAIL_INSERT:-0}" = "1" ]; then exit 1; fi
    printf '%s\n' "$rule" > "$f.tmp"
    cat "$f" >> "$f.tmp"
    mv "$f.tmp" "$f"
    ;;
  -S)
    printf -- '-P %s ACCEPT\n' "$chain"
    while IFS= read -r line || [ -n "$line" ]; do
      [ -z "$line" ] && continue
      printf -- '-A %s %s\n' "$chain" "$line"
    done < "$f"
    ;;
  *) exit 2 ;;
esac
`

type bypassScriptRun struct {
	t     *testing.T
	state string
	bin   string
}

func newBypassScriptRun(t *testing.T) *bypassScriptRun {
	t.Helper()
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	state := filepath.Join(root, "state")
	for _, d := range []string{bin, state} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(bin, "iptables-nft"), []byte(iptablesMock), 0o755); err != nil {
		t.Fatalf("write iptables-nft mock: %v", err)
	}
	return &bypassScriptRun{t: t, state: state, bin: bin}
}

// seed installs a pre-existing rule, e.g. the jump linkerd-init appends.
func (r *bypassScriptRun) seed(table, chain, rule string) {
	r.t.Helper()
	path := filepath.Join(r.state, table+"_"+chain)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		r.t.Fatalf("seed %s/%s: %v", table, chain, err)
	}
	defer f.Close()
	if _, err := f.WriteString(rule + "\n"); err != nil {
		r.t.Fatalf("seed write %s/%s: %v", table, chain, err)
	}
}

func (r *bypassScriptRun) run(extraEnv ...string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("/bin/sh", "-c", MacvlanBypassScript())
	cmd.Env = append(os.Environ(),
		"PATH="+r.bin+":"+os.Getenv("PATH"),
		"IPT_STATE="+r.state,
		"MACVLAN_IFACE=net1",
	)
	cmd.Env = append(cmd.Env, extraEnv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *bypassScriptRun) rules(table, chain string) []string {
	r.t.Helper()
	raw, err := os.ReadFile(filepath.Join(r.state, table+"_"+chain))
	if err != nil {
		r.t.Fatalf("read %s/%s: %v", table, chain, err)
	}
	lines := make([]string, 0, 4)
	for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func TestMacvlanBypassScriptWinsChainHeadOverProxyRules(t *testing.T) {
	r := newBypassScriptRun(t)
	// linkerd-init appends its inbound jump, mesh-in inserts a REDIRECT head.
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", "-p tcp -d 10.0.0.1 --dport 80 -j REDIRECT --to-ports 15001")
	r.seed("filter", "INPUT", "-p tcp -j DROP")
	r.seed("filter", "OUTPUT", "-p tcp -j DROP")

	out, err := r.run()
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "verified four iface ACCEPT heads on nft") {
		t.Fatalf("self-check must confirm nft chain heads, output:\n%s", out)
	}

	for _, tc := range []struct{ table, chain, want string }{
		{"nat", "PREROUTING", "-i net1 -j ACCEPT"},
		{"nat", "OUTPUT", "-o net1 -j ACCEPT"},
		{"filter", "INPUT", "-i net1 -j ACCEPT"},
		{"filter", "OUTPUT", "-o net1 -j ACCEPT"},
	} {
		rules := r.rules(tc.table, tc.chain)
		if rules[0] != tc.want {
			t.Fatalf("%s/%s head = %q, want %q (all: %v)", tc.table, tc.chain, rules[0], tc.want, rules)
		}
		if len(rules) != 2 {
			t.Fatalf("%s/%s must keep the pre-existing rule, got %v", tc.table, tc.chain, rules)
		}
	}
}

func TestMacvlanBypassScriptReinsertsStaleRuleAtHead(t *testing.T) {
	r := newBypassScriptRun(t)
	// A bypass rule left behind below a later -I 1 from another component.
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "PREROUTING", "-i net1 -j ACCEPT")

	out, err := r.run()
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}

	rules := r.rules("nat", "PREROUTING")
	if rules[0] != "-i net1 -j ACCEPT" {
		t.Fatalf("stale rule not moved to head: %v", rules)
	}
	if len(rules) != 2 {
		t.Fatalf("stale duplicate not cleaned up: %v", rules)
	}
}

func TestMacvlanBypassScriptIsIdempotentAcrossRuns(t *testing.T) {
	r := newBypassScriptRun(t)

	for i := 0; i < 3; i++ {
		if out, err := r.run(); err != nil {
			t.Fatalf("run %d failed: %v\n%s", i, err, out)
		}
	}

	if rules := r.rules("filter", "INPUT"); len(rules) != 1 {
		t.Fatalf("expected exactly one rule after repeated runs, got %v", rules)
	}
}

func TestMacvlanBypassScriptFailsClosedWhenInsertFails(t *testing.T) {
	r := newBypassScriptRun(t)

	out, err := r.run("IPT_FAIL_INSERT=1")
	if err == nil {
		t.Fatalf("script must exit non-zero when a rule cannot be installed, output:\n%s", out)
	}
	if !strings.Contains(out, "failed to install") {
		t.Fatalf("script must report the failing rule, output:\n%s", out)
	}
}

func TestMacvlanBypassScriptFailsClosedWhenNftBinaryMissing(t *testing.T) {
	r := newBypassScriptRun(t)

	out, err := r.run("IPTABLES_BIN=iptables-nft-missing")
	if err == nil {
		t.Fatalf("script must fail when iptables-nft is missing, output:\n%s", out)
	}
	if !strings.Contains(out, "iptables-nft-missing not found") {
		t.Fatalf("script must report missing nft binary, output:\n%s", out)
	}
}

func TestMacvlanBypassScriptFailsClosedWhenNotNftVariant(t *testing.T) {
	r := newBypassScriptRun(t)
	legacy := `#!/bin/sh
if [ "${1:-}" = "--version" ]; then
  echo "iptables v1.8.8 (legacy)"
  exit 0
fi
exit 0
`
	if err := os.WriteFile(filepath.Join(r.bin, "iptables-legacy-fake"), []byte(legacy), 0o755); err != nil {
		t.Fatalf("write legacy fake: %v", err)
	}

	out, err := r.run("IPTABLES_BIN=iptables-legacy-fake")
	if err == nil {
		t.Fatalf("script must fail when binary is not nf_tables, output:\n%s", out)
	}
	if !strings.Contains(out, "is not nf_tables") {
		t.Fatalf("script must reject non-nft variant, output:\n%s", out)
	}
}

package sidecar

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type nestedBypassScriptRun struct {
	t     *testing.T
	state string
	bin   string
}

func newNestedBypassScriptRun(t *testing.T) *nestedBypassScriptRun {
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
	return &nestedBypassScriptRun{t: t, state: state, bin: bin}
}

func (r *nestedBypassScriptRun) seed(table, chain, rule string) {
	r.t.Helper()
	path := filepath.Join(r.state, table+"_"+chain)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		r.t.Fatalf("seed %s/%s: %v", table, chain, err)
	}
	defer f.Close()
	if _, err := f.WriteString(rule + "\n"); err != nil {
		r.t.Fatalf("seed write: %v", err)
	}
}

func (r *nestedBypassScriptRun) run(ifaces string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("/bin/sh", "-c", NestedIfaceBypassScript())
	cmd.Env = append(os.Environ(),
		"PATH="+r.bin+":"+os.Getenv("PATH"),
		"IPT_STATE="+r.state,
		"MESH_BYPASS_IFACES="+ifaces,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *nestedBypassScriptRun) rules(table, chain string) []string {
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

func TestNestedIfaceBypassScriptWinsOverProxyInit(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", "-p tcp -j PROXY_INIT_OUTPUT")
	r.seed("filter", "INPUT", "-p tcp -j DROP")
	r.seed("filter", "OUTPUT", "-p tcp -j DROP")

	out, err := r.run("docker")
	if err != nil {
		t.Fatalf("script failed: %v\n%s", out, err)
	}
	if !strings.Contains(out, "verified ACCEPT heads") {
		t.Fatalf("missing verify log:\n%s", out)
	}
	if got := r.rules("nat", "PREROUTING"); got[0] != "-i docker -j ACCEPT" {
		t.Fatalf("PREROUTING head=%v", got)
	}
	if got := r.rules("nat", "PREROUTING"); len(got) < 2 || got[1] != "-p tcp -j PROXY_INIT_REDIRECT" {
		t.Fatalf("PROXY_INIT must remain after ACCEPT: %v", got)
	}
}

func TestNestedIfaceBypassScriptMultiIface(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", "-p tcp -j PROXY_INIT_OUTPUT")
	r.seed("filter", "INPUT", "-j DROP")
	r.seed("filter", "OUTPUT", "-j DROP")

	if out, err := r.run("docker,br0"); err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	pre := r.rules("nat", "PREROUTING")
	if len(pre) < 3 {
		t.Fatalf("expected two ACCEPT + PROXY_INIT, got %v", pre)
	}
	joined := strings.Join(pre, "\n")
	if !strings.Contains(joined, "-i docker -j ACCEPT") || !strings.Contains(joined, "-i br0 -j ACCEPT") {
		t.Fatalf("missing iface ACCEPT: %v", pre)
	}
	if pre[0] == "-p tcp -j PROXY_INIT_REDIRECT" {
		t.Fatalf("PROXY_INIT must not stay at head: %v", pre)
	}
}

func TestNestedIfaceBypassScriptFailsWithoutNft(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	_ = os.Remove(filepath.Join(r.bin, "iptables-nft"))
	out, err := r.run("docker")
	if err == nil {
		t.Fatalf("expected failure, out=%s", out)
	}
}

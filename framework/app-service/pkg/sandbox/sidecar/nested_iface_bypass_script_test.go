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
	return r.runWithPorts(ifaces, "")
}

func (r *nestedBypassScriptRun) runWithPorts(ifaces, redirectPorts string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("/bin/sh", "-c", NestedIfaceBypassScript())
	cmd.Env = append(os.Environ(),
		"PATH="+r.bin+":"+os.Getenv("PATH"),
		"IPT_STATE="+r.state,
		"MESH_BYPASS_IFACES="+ifaces,
		"MESH_IN_REDIRECT_PORTS="+redirectPorts,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

const (
	meshInOutput443  = "-d 10.233.25.10/32 -p tcp -m tcp --dport 443 -m owner ! --uid-owner 1651 -m owner ! --uid-owner 1555 -m owner ! --uid-owner 2102 -j REDIRECT --to-ports 16443"
	meshInOutput80   = "-d 10.233.25.10/32 -p tcp -m tcp --dport 80 -m owner ! --uid-owner 1651 -m owner ! --uid-owner 1555 -m owner ! --uid-owner 2102 -j REDIRECT --to-ports 16080"
	guestRedirect80  = "-d 10.233.25.10/32 -i docker -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 16080"
	guestRedirect443 = "-d 10.233.25.10/32 -i docker -p tcp -m tcp --dport 443 -j REDIRECT --to-ports 16443"
)

func seedMeshedVirtPod(r *nestedBypassScriptRun) {
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", meshInOutput443)
	r.seed("nat", "OUTPUT", meshInOutput80)
	r.seed("nat", "OUTPUT", "-p tcp -j PROXY_INIT_OUTPUT")
	r.seed("filter", "INPUT", "-p tcp -j DROP")
	r.seed("filter", "OUTPUT", "-p tcp -j DROP")
}

func countRule(rules []string, want string) int {
	n := 0
	for _, r := range rules {
		if r == want {
			n++
		}
	}
	return n
}

func TestNestedIfaceBypassScriptRedirectsGuestGatewayTraffic(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	seedMeshedVirtPod(r)

	out, err := r.runWithPorts("docker", "16080,16443")
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "mirrored 2 mesh-in redirect(s)") {
		t.Fatalf("missing mirror log:\n%s", out)
	}
	want := []string{guestRedirect80, guestRedirect443, "-i docker -j ACCEPT", "-p tcp -j PROXY_INIT_REDIRECT"}
	got := r.rules("nat", "PREROUTING")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("PREROUTING=%v\nwant %v", got, want)
	}
	if out := r.rules("nat", "OUTPUT"); out[0] != "-o docker -j ACCEPT" || countRule(out, meshInOutput80) != 1 {
		t.Fatalf("OUTPUT must keep mesh-in rules behind iface ACCEPT: %v", out)
	}
}

func TestNestedIfaceBypassScriptGuestRedirectIdempotent(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	seedMeshedVirtPod(r)

	for i := 0; i < 2; i++ {
		if out, err := r.runWithPorts("docker", "16080,16443"); err != nil {
			t.Fatalf("run %d failed: %v\n%s", i, err, out)
		}
	}
	got := r.rules("nat", "PREROUTING")
	want := []string{guestRedirect80, guestRedirect443, "-i docker -j ACCEPT", "-p tcp -j PROXY_INIT_REDIRECT"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("PREROUTING after rerun=%v\nwant %v", got, want)
	}
}

func TestNestedIfaceBypassScriptNoGuestRedirectWithoutMeshIn(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", "-p tcp -j PROXY_INIT_OUTPUT")
	r.seed("filter", "INPUT", "-p tcp -j DROP")
	r.seed("filter", "OUTPUT", "-p tcp -j DROP")

	out, err := r.runWithPorts("docker", "16080,16443")
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "mirrored 0 mesh-in redirect(s)") {
		t.Fatalf("missing mirror log:\n%s", out)
	}
	if got := r.rules("nat", "PREROUTING"); len(got) != 2 || got[0] != "-i docker -j ACCEPT" {
		t.Fatalf("PREROUTING=%v", got)
	}
}

func TestNestedIfaceBypassScriptNoGuestRedirectWithoutPortsEnv(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	seedMeshedVirtPod(r)

	if out, err := r.runWithPorts("docker", ""); err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	if got := r.rules("nat", "PREROUTING"); got[0] != "-i docker -j ACCEPT" ||
		countRule(got, guestRedirect80) != 0 || countRule(got, guestRedirect443) != 0 {
		t.Fatalf("PREROUTING=%v", got)
	}
}

func TestNestedIfaceBypassScriptIgnoresForeignRedirects(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	r.seed("nat", "PREROUTING", "-p tcp -j PROXY_INIT_REDIRECT")
	r.seed("nat", "OUTPUT", "-p tcp -m multiport --dports 80,8080 -m owner ! --uid-owner 1652 -j REDIRECT --to-ports 15001")
	r.seed("nat", "OUTPUT", "-d 10.233.0.0/16 -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 16080")
	r.seed("nat", "OUTPUT", "-d 10.233.9.9/32 -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 15001")
	r.seed("nat", "OUTPUT", meshInOutput80)
	r.seed("filter", "INPUT", "-p tcp -j DROP")
	r.seed("filter", "OUTPUT", "-p tcp -j DROP")

	out, err := r.runWithPorts("docker", "16080,16443")
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	want := []string{guestRedirect80, "-i docker -j ACCEPT", "-p tcp -j PROXY_INIT_REDIRECT"}
	if got := r.rules("nat", "PREROUTING"); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("PREROUTING=%v\nwant %v", got, want)
	}
}

func TestNestedIfaceBypassScriptGuestRedirectMultiIface(t *testing.T) {
	r := newNestedBypassScriptRun(t)
	seedMeshedVirtPod(r)

	if out, err := r.runWithPorts("docker,br0", "16080,16443"); err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	pre := r.rules("nat", "PREROUTING")
	index := func(rule string) int {
		for i, l := range pre {
			if l == rule {
				return i
			}
		}
		return -1
	}
	proxyInit := index("-p tcp -j PROXY_INIT_REDIRECT")
	for _, iface := range []string{"docker", "br0"} {
		accept := index("-i " + iface + " -j ACCEPT")
		for _, spec := range []string{
			strings.Replace(guestRedirect80, "-i docker", "-i "+iface, 1),
			strings.Replace(guestRedirect443, "-i docker", "-i "+iface, 1),
		} {
			at := index(spec)
			if at < 0 || at > accept || at > proxyInit || countRule(pre, spec) != 1 {
				t.Fatalf("%s: redirect %q at %d (accept %d, proxy-init %d): %v", iface, spec, at, accept, proxyInit, pre)
			}
		}
	}
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

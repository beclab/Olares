package lpvpndns

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockExec records every command and replies to read probes
// (`iptables -t nat -S <chain>`) from configurable dumps.
type mockExec struct {
	chainDump       string // "" => chain absent (iptables returns error)
	postroutingDump string
	postroutingErr  error // when set, `-S POSTROUTING` fails
	restoreErr      error // when set, iptables-restore fails

	cmds   [][]string // every runCmd invocation (args joined)
	stdins []string   // every runCmdStdin payload
	stdcmd [][]string // every runCmdStdin invocation args
}

func (m *mockExec) run(_ context.Context, name string, args ...string) ([]byte, error) {
	full := append([]string{name}, args...)
	m.cmds = append(m.cmds, full)
	joined := strings.Join(full, " ")

	switch {
	case strings.Contains(joined, "-S "+chainName):
		if m.chainDump == "" {
			return nil, errors.New("iptables: No chain/target/match by that name")
		}
		return []byte(m.chainDump), nil
	case strings.Contains(joined, "-S "+postRouting):
		if m.postroutingErr != nil {
			return nil, m.postroutingErr
		}
		return []byte(m.postroutingDump), nil
	}
	return nil, nil
}

func (m *mockExec) runStdin(_ context.Context, stdin, name string, args ...string) ([]byte, error) {
	m.stdcmd = append(m.stdcmd, append([]string{name}, args...))
	m.stdins = append(m.stdins, stdin)
	if m.restoreErr != nil {
		return nil, m.restoreErr
	}
	return nil, nil
}

// install swaps the package runners with the mock and returns a restore func.
func (m *mockExec) install() func() {
	origRun, origStdin := runCmd, runCmdStdin
	runCmd = m.run
	runCmdStdin = m.runStdin
	return func() { runCmd, runCmdStdin = origRun, origStdin }
}

// writeCmds returns iptables/conntrack commands that mutate state (everything
// except read-only `-S` probes).
func (m *mockExec) writeCmds() [][]string {
	var w [][]string
	for _, c := range m.cmds {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, " -S ") {
			continue
		}
		w = append(w, c)
	}
	return w
}

func (m *mockExec) hasCmdContaining(sub string) bool {
	for _, c := range m.cmds {
		if strings.Contains(strings.Join(c, " "), sub) {
			return true
		}
	}
	return false
}

const (
	testVPNIP   = "192.168.128.102"
	testPodCIDR = "10.233.64.0/18"
)

func steadyChainDump(vpnIP, podCIDR string) string {
	return strings.Join([]string{
		"-N " + chainName,
		"-A " + chainName + " -d " + podCIDR + " ! -s " + podCIDR + " -p udp -m udp --dport 53 -j SNAT --to-source " + vpnIP,
		"-A " + chainName + " -d " + podCIDR + " ! -s " + podCIDR + " -p tcp -m tcp --dport 53 -j SNAT --to-source " + vpnIP,
	}, "\n")
}

// jump on top of POSTROUTING followed by kube-proxy jump.
const steadyPostrouting = "-P POSTROUTING ACCEPT\n" +
	"-A POSTROUTING -j " + chainName + "\n" +
	"-A POSTROUTING -m comment --comment kubernetes -j KUBE-POSTROUTING"

func TestReconcile_SteadyState_NoWrites(t *testing.T) {
	m := &mockExec{
		chainDump:       steadyChainDump(testVPNIP, testPodCIDR),
		postroutingDump: steadyPostrouting,
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	if w := m.writeCmds(); len(w) != 0 {
		t.Fatalf("steady state must not write, got %v", w)
	}
	if len(m.stdcmd) != 0 {
		t.Fatalf("steady state must not run iptables-restore, got %v", m.stdcmd)
	}
}

func TestReconcile_VPNIPChange_RebuildAndNarrowConntrack(t *testing.T) {
	oldIP := "192.168.128.106"
	m := &mockExec{
		chainDump:       steadyChainDump(oldIP, testPodCIDR),
		postroutingDump: steadyPostrouting,
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	// exactly one atomic rebuild via iptables-restore --noflush
	if len(m.stdcmd) != 1 {
		t.Fatalf("want one iptables-restore, got %d (%v)", len(m.stdcmd), m.stdcmd)
	}
	restoreArgs := strings.Join(m.stdcmd[0], " ")
	if !strings.Contains(restoreArgs, "iptables-restore") || !strings.Contains(restoreArgs, "--noflush") {
		t.Fatalf("rebuild must use `iptables-restore --noflush`, got %q", restoreArgs)
	}
	assertRestoreInputRedlines(t, m.stdins[0])

	// narrow conntrack on old IP, udp + tcp, never a broad flush
	if !m.hasCmdContaining("conntrack -D --reply-dst " + oldIP + " -p udp --dport 53") {
		t.Fatalf("missing narrow udp conntrack flush; cmds=%v", m.cmds)
	}
	if !m.hasCmdContaining("conntrack -D --reply-dst " + oldIP + " -p tcp --dport 53") {
		t.Fatalf("missing narrow tcp conntrack flush; cmds=%v", m.cmds)
	}
	assertNoBroadConntrack(t, m)
}

func TestReconcile_PodCIDRDriftOnly_NoConntrack(t *testing.T) {
	m := &mockExec{
		chainDump:       steadyChainDump(testVPNIP, "10.244.0.0/16"), // wrong podCIDR, same vpnIP
		postroutingDump: steadyPostrouting,
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	if len(m.stdcmd) != 1 {
		t.Fatalf("podCIDR drift must rebuild once, got %d", len(m.stdcmd))
	}
	if m.hasCmdContaining("conntrack") {
		t.Fatalf("podCIDR-only drift must NOT touch conntrack; cmds=%v", m.cmds)
	}
}

func TestReconcile_FirstInstall_NoConntrack(t *testing.T) {
	m := &mockExec{
		chainDump:       "", // chain absent
		postroutingDump: "-P POSTROUTING ACCEPT\n-A POSTROUTING -j KUBE-POSTROUTING",
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	if len(m.stdcmd) != 1 {
		t.Fatalf("first install must rebuild once, got %d", len(m.stdcmd))
	}
	assertRestoreInputRedlines(t, m.stdins[0])
	// jump absent -> must be inserted at top
	if !m.hasCmdContaining("-I " + postRouting + " 1 -j " + chainName) {
		t.Fatalf("first install must insert top jump; cmds=%v", m.cmds)
	}
	// first install relies on new-flow SNAT + 30s expiry, not a broad flush
	if m.hasCmdContaining("conntrack") {
		t.Fatalf("first install must NOT flush conntrack; cmds=%v", m.cmds)
	}
}

func TestEnsureJump_FixesNonTopJump(t *testing.T) {
	m := &mockExec{
		// our jump is second, behind kube-proxy
		postroutingDump: "-P POSTROUTING ACCEPT\n" +
			"-A POSTROUTING -j KUBE-POSTROUTING\n" +
			"-A POSTROUTING -j " + chainName,
	}
	defer m.install()()

	if err := ensureJump(context.Background()); err != nil {
		t.Fatalf("ensureJump: %v", err)
	}

	if !m.hasCmdContaining("-I " + postRouting + " 1 -j " + chainName) {
		t.Fatalf("must insert jump on top; cmds=%v", m.cmds)
	}
	// only -C/-I/-D maintenance, never -F/-A on POSTROUTING
	for _, c := range m.cmds {
		j := strings.Join(c, " ")
		if strings.Contains(j, "-A "+postRouting) || strings.Contains(j, "-F "+postRouting) {
			t.Fatalf("jump must not be maintained via -A/-F: %q", j)
		}
	}
}

func TestEnsureJump_PostroutingReadError_NoWrites(t *testing.T) {
	m := &mockExec{
		postroutingErr: errors.New("iptables: read POSTROUTING failed"),
	}
	defer m.install()()

	if err := ensureJump(context.Background()); err == nil {
		t.Fatal("expected POSTROUTING read error")
	}
	if w := m.writeCmds(); len(w) != 0 {
		t.Fatalf("POSTROUTING read error must not mutate iptables, got %v", w)
	}
}

func TestReconcile_RebuildFailure_SkipsEnsureJump(t *testing.T) {
	m := &mockExec{
		chainDump:       "", // chain absent -> drift -> rebuild
		postroutingDump: "-P POSTROUTING ACCEPT\n-A POSTROUTING -j KUBE-POSTROUTING",
		restoreErr:      errors.New("iptables-restore: transaction failed"),
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	if len(m.stdcmd) != 1 {
		t.Fatalf("rebuild failure must attempt one restore, got %d (%v)", len(m.stdcmd), m.stdcmd)
	}
	if m.hasCmdContaining("-I " + postRouting) {
		t.Fatalf("rebuild failure must skip ensureJump; cmds=%v", m.cmds)
	}
	if m.hasCmdContaining("conntrack") {
		t.Fatalf("rebuild failure must skip conntrack; cmds=%v", m.cmds)
	}
}

func TestReconcile_PostroutingReadError_NoWrites(t *testing.T) {
	m := &mockExec{
		chainDump:      steadyChainDump(testVPNIP, testPodCIDR),
		postroutingErr: errors.New("iptables: read POSTROUTING failed"),
	}
	defer m.install()()

	reconcile(context.Background(), testVPNIP, testPodCIDR)

	if w := m.writeCmds(); len(w) != 0 {
		t.Fatalf("POSTROUTING read error must not mutate iptables, got %v", w)
	}
	if len(m.stdcmd) != 0 {
		t.Fatalf("steady chain must not rebuild on POSTROUTING read error, got %v", m.stdcmd)
	}
}

func TestTeardown_PostroutingReadError_NoChainMutation(t *testing.T) {
	m := &mockExec{
		chainDump:      steadyChainDump(testVPNIP, testPodCIDR),
		postroutingErr: errors.New("iptables: read POSTROUTING failed"),
	}
	defer m.install()()

	teardown(context.Background())

	if m.hasCmdContaining("-F "+chainName) || m.hasCmdContaining("-X "+chainName) {
		t.Fatalf("POSTROUTING read error must not flush/delete chain; cmds=%v", m.cmds)
	}
	if m.hasCmdContaining("-D " + postRouting) {
		t.Fatalf("POSTROUTING read error must not delete jumps; cmds=%v", m.cmds)
	}
	if m.hasCmdContaining("conntrack") {
		t.Fatalf("POSTROUTING read error must not flush conntrack; cmds=%v", m.cmds)
	}
}

func TestTeardown_Idempotent_NoChain(t *testing.T) {
	m := &mockExec{
		chainDump:       "",
		postroutingDump: "-P POSTROUTING ACCEPT\n-A POSTROUTING -j KUBE-POSTROUTING",
	}
	defer m.install()()

	teardown(context.Background())

	// no chain present -> must not attempt -F/-X or conntrack
	if m.hasCmdContaining("-F "+chainName) || m.hasCmdContaining("-X "+chainName) {
		t.Fatalf("teardown on absent chain must not flush/delete chain; cmds=%v", m.cmds)
	}
	if m.hasCmdContaining("conntrack") {
		t.Fatalf("teardown on absent chain must not flush conntrack; cmds=%v", m.cmds)
	}
}

func TestBuildRestoreInput_Redlines(t *testing.T) {
	assertRestoreInputRedlines(t, buildRestoreInput(testVPNIP, testPodCIDR))
}

// assertRestoreInputRedlines enforces the §4.1.1 safety red lines on the
// iptables-restore payload.
func assertRestoreInputRedlines(t *testing.T, input string) {
	t.Helper()
	if !strings.Contains(input, ":"+chainName+" - [0:0]") {
		t.Fatalf("restore input must declare own chain; got:\n%s", input)
	}
	if strings.Contains(input, ":"+postRouting) {
		t.Fatalf("restore input must NOT declare built-in POSTROUTING; got:\n%s", input)
	}
	// exactly one chain declaration line (starts with ':')
	chainDecls := 0
	tables := 0
	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(line, ":") {
			chainDecls++
		}
		if strings.HasPrefix(line, "*") {
			tables++
		}
	}
	if chainDecls != 1 {
		t.Fatalf("restore input must declare exactly one chain, got %d:\n%s", chainDecls, input)
	}
	if tables != 1 || !strings.Contains(input, "*nat") {
		t.Fatalf("restore input must be a single *nat table; got:\n%s", input)
	}
	if !strings.Contains(input, "COMMIT") {
		t.Fatalf("restore input must COMMIT; got:\n%s", input)
	}
	if !strings.Contains(input, "--to-source "+testVPNIP) {
		t.Fatalf("restore input must SNAT to vpnIP; got:\n%s", input)
	}
	if !strings.Contains(input, "! -s "+testPodCIDR) {
		t.Fatalf("restore input must exclude podCIDR source; got:\n%s", input)
	}
}

// assertNoBroadConntrack ensures no unfiltered node-wide --dport 53 flush.
func assertNoBroadConntrack(t *testing.T, m *mockExec) {
	t.Helper()
	for _, c := range m.cmds {
		j := strings.Join(c, " ")
		if !strings.Contains(j, "conntrack") {
			continue
		}
		if strings.Contains(j, "--dport 53") && !strings.Contains(j, "--reply-dst") {
			t.Fatalf("broad conntrack flush detected (no --reply-dst): %q", j)
		}
	}
}

func TestIsIPTablesChainAbsent(t *testing.T) {
	cases := []struct {
		out  string
		want bool
	}{
		{"iptables: No chain/target/match by that name", true},
		{"iptables: Permission denied", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isIPTablesChainAbsent([]byte(c.out)); got != c.want {
			t.Fatalf("isIPTablesChainAbsent(%q) = %v, want %v", c.out, got, c.want)
		}
	}
}

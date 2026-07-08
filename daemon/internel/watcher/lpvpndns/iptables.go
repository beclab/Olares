package lpvpndns

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	chainName      = "OLARES-LPVPN-DNS"
	natTable       = "nat"
	postRouting    = "POSTROUTING"
	dnsPort        = "53"
	defaultPodCIDR = "10.233.64.0/18"
)

// runCmd / runCmdStdin are package-level command runners so unit tests can
// replace them with mocks and assert exact command shape (the §4.1.1 red
// lines: --noflush, single chain declaration, narrow conntrack, etc.).
var runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

var runCmdStdin = func(ctx context.Context, stdin, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.CombinedOutput()
}

var (
	toSourceRe = regexp.MustCompile(`--to-source\s+(\S+)`)
	dstRe      = regexp.MustCompile(`(?:^|\s)-d\s+(\S+)`)
)

// readChain returns the `iptables -t nat -S OLARES-LPVPN-DNS` lines and whether
// the chain exists. A non-nil error from iptables (chain absent) yields
// exists=false.
func readChain(ctx context.Context) (lines []string, exists bool) {
	out, err := runCmd(ctx, "iptables", "-t", natTable, "-S", chainName)
	if err != nil {
		if isIPTablesChainAbsent(out) {
			klog.V(4).Infof("lpvpndns: chain %s absent", chainName)
		} else {
			klog.Warningf("lpvpndns: read chain %s failed: %v, out=%s", chainName, err, string(out))
		}
		return nil, false
	}
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, true
}

// isIPTablesChainAbsent reports whether iptables -S failed because the chain
// does not exist (expected on first install or after teardown).
func isIPTablesChainAbsent(out []byte) bool {
	return strings.Contains(string(out), "No chain/target/match")
}

// parseChain derives the SNAT target (--to-source), the destination CIDR (-d)
// and the count of `-A OLARES-LPVPN-DNS` rules from the chain dump. Semantic
// parsing avoids brittle string equality against iptables' normalized output.
func parseChain(lines []string) (vpnIP, podCIDR string, ruleCount int) {
	for _, l := range lines {
		if !strings.HasPrefix(l, "-A "+chainName) {
			continue
		}
		ruleCount++
		if vpnIP == "" {
			if m := toSourceRe.FindStringSubmatch(l); len(m) == 2 {
				vpnIP = m[1]
			}
		}
		if podCIDR == "" {
			if m := dstRe.FindStringSubmatch(l); len(m) == 2 {
				podCIDR = m[1]
			}
		}
	}
	return vpnIP, podCIDR, ruleCount
}

// reconcile is drift-only: it reads the actual state, compares against the
// desired state, and writes iptables ONLY when there is a real difference.
// In steady state it performs read-only probes and zero writes.
func reconcile(ctx context.Context, vpnIP, podCIDR string) {
	lines, exists := readChain(ctx)
	oldVPNIP, curPodCIDR, ruleCount := parseChain(lines)

	chainDrift := !exists || ruleCount != 2 || oldVPNIP != vpnIP || curPodCIDR != podCIDR
	vpnChanged := exists && oldVPNIP != "" && oldVPNIP != vpnIP

	if chainDrift {
		klog.Infof("lpvpndns: rebuild chain %s vpnIP=%s podCIDR=%s (was vpnIP=%s podCIDR=%s rules=%d exists=%v)",
			chainName, vpnIP, podCIDR, oldVPNIP, curPodCIDR, ruleCount, exists)
		if out, err := rebuildChainAtomic(ctx, vpnIP, podCIDR); err != nil {
			klog.Errorf("lpvpndns: rebuild chain failed: %v, out=%s", err, string(out))
			// Fail-closed: do not ensureJump against a chain that may be missing or
			// half-written; retry on the next tick after a successful rebuild.
			return
		}
	}

	if err := ensureJump(ctx); err != nil {
		klog.V(4).Infof("lpvpndns: ensure POSTROUTING jump failed, retry next tick: %v", err)
		return
	}

	if vpnChanged {
		klog.Infof("lpvpndns: vpnIP changed %s -> %s, narrow conntrack flush", oldVPNIP, vpnIP)
		cleanConntrack(ctx, oldVPNIP)
	}
}

// rebuildChainAtomic replaces ONLY the OLARES-LPVPN-DNS chain in a single
// iptables-restore --noflush transaction. Per iptables-restore(8), only the
// chains declared with `:` are flushed and rebuilt; all other chains and
// tables are retained. The POSTROUTING jump is never part of this input.
func rebuildChainAtomic(ctx context.Context, vpnIP, podCIDR string) ([]byte, error) {
	return runCmdStdin(ctx, buildRestoreInput(vpnIP, podCIDR), "iptables-restore", "--noflush")
}

func buildRestoreInput(vpnIP, podCIDR string) string {
	return fmt.Sprintf(`*nat
:%[1]s - [0:0]
-A %[1]s -d %[2]s ! -s %[2]s -p udp --dport %[4]s -j SNAT --to-source %[3]s
-A %[1]s -d %[2]s ! -s %[2]s -p tcp --dport %[4]s -j SNAT --to-source %[3]s
COMMIT
`, chainName, podCIDR, vpnIP, dnsPort)
}

// jumpRuleNumbers returns the 1-based rule numbers (in POSTROUTING) of every
// `-j OLARES-LPVPN-DNS` jump, ascending. A read error is distinct from an
// empty jump list so callers can fail-closed instead of treating errors as
// "no jump".
func jumpRuleNumbers(ctx context.Context) ([]int, error) {
	out, err := runCmd(ctx, "iptables", "-t", natTable, "-S", postRouting)
	if err != nil {
		return nil, err
	}
	var nums []int
	idx := 0
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "-A "+postRouting) {
			continue
		}
		idx++ // rule number among -A POSTROUTING rules (1-based)
		if strings.Contains(l, "-j "+chainName) {
			nums = append(nums, idx)
		}
	}
	sort.Ints(nums)
	return nums, nil
}

// ensureJump keeps exactly one `-j OLARES-LPVPN-DNS` jump as the first rule of
// POSTROUTING (ahead of KUBE-POSTROUTING). It is drift-only: no write when the
// jump is already unique and on top. Insert-before-delete guarantees there is
// never a window without a jump. Returns an error when POSTROUTING cannot be
// read so callers fail-closed instead of inserting duplicate jumps.
func ensureJump(ctx context.Context) error {
	nums, err := jumpRuleNumbers(ctx)
	if err != nil {
		klog.Errorf("lpvpndns: read POSTROUTING jump state failed: %v", err)
		return err
	}
	if len(nums) == 1 && nums[0] == 1 {
		return nil
	}

	if out, err := runCmd(ctx, "iptables", "-t", natTable, "-I", postRouting, "1", "-j", chainName); err != nil {
		klog.Errorf("lpvpndns: insert POSTROUTING jump failed: %v, out=%s", err, string(out))
		return err
	}

	// Delete the remaining (now-redundant) jumps, keeping the new top one
	// (number 1). Delete from highest number to lowest to keep indices valid.
	again, err := jumpRuleNumbers(ctx)
	if err != nil {
		klog.Errorf("lpvpndns: re-read POSTROUTING jump state failed: %v", err)
		return err
	}
	for i := len(again) - 1; i >= 0; i-- {
		if again[i] == 1 {
			continue
		}
		if out, err := runCmd(ctx, "iptables", "-t", natTable, "-D", postRouting, strconv.Itoa(again[i])); err != nil {
			klog.Errorf("lpvpndns: delete redundant POSTROUTING jump #%d failed: %v, out=%s", again[i], err, string(out))
		}
	}
	return nil
}

// cleanConntrack removes ONLY the reply-side conntrack entries previously SNAT
// rewritten to oldVPNIP (reply-dst == oldVPNIP) on :53. It never touches
// incluster Pod flows or external DNS flows. best-effort; a missing conntrack
// tool degrades to 30s natural expiry.
func cleanConntrack(ctx context.Context, oldVPNIP string) {
	if net.ParseIP(oldVPNIP) == nil {
		return
	}
	for _, proto := range []string{"udp", "tcp"} {
		if _, err := runCmd(ctx, "conntrack", "-D", "--reply-dst", oldVPNIP, "-p", proto, "--dport", dnsPort); err != nil {
			klog.V(4).Infof("lpvpndns: conntrack flush reply-dst=%s proto=%s: %v", oldVPNIP, proto, err)
		}
	}
}

// teardown removes the jump, flushes and deletes the chain, and narrow-cleans
// conntrack. Fully idempotent: a missing chain/jump is a no-op.
func teardown(ctx context.Context) {
	lines, exists := readChain(ctx)
	oldVPNIP, _, _ := parseChain(lines)

	jumps, err := jumpRuleNumbers(ctx)
	if err != nil {
		klog.Errorf("lpvpndns: teardown read POSTROUTING jump state failed: %v", err)
		return
	}
	for len(jumps) > 0 {
		if out, err := runCmd(ctx, "iptables", "-t", natTable, "-D", postRouting, "-j", chainName); err != nil {
			klog.Errorf("lpvpndns: teardown delete POSTROUTING jump failed: %v, out=%s", err, string(out))
			break
		}
		jumps, err = jumpRuleNumbers(ctx)
		if err != nil {
			klog.Errorf("lpvpndns: teardown re-read POSTROUTING jump state failed: %v", err)
			return
		}
	}

	if exists {
		if out, err := runCmd(ctx, "iptables", "-t", natTable, "-F", chainName); err != nil {
			klog.Errorf("lpvpndns: teardown flush chain failed: %v, out=%s", err, string(out))
		}
		if out, err := runCmd(ctx, "iptables", "-t", natTable, "-X", chainName); err != nil {
			klog.Errorf("lpvpndns: teardown delete chain failed: %v, out=%s", err, string(out))
		}
		cleanConntrack(ctx, oldVPNIP)
	}
}

package nodestatus

import (
	"context"
	"os/exec"
	"strings"
)

// CapUpgradeStages says this node can be given one stage of a cluster upgrade
// plan and run it.
//
// It is the answer to a question the control node cannot ask any other way: an
// olaresd from before staged upgrades existed serves no such route, and
// dialling it would return a 404 in the middle of the upgrade, after earlier
// stages had already changed the cluster. Declaring the capability is how a
// node that understands stages says so, and a node that never heard of the key
// says nothing — which is exactly the "not declared" the capability model
// already means.
const CapUpgradeStages = "upgrade.stages"

// CapConfigCLIVersion is the olares-cli version a node reports alongside
// CapUpgradeStages. The control node compares it with the version being rolled
// out, because a node holding the previous olares-cli would derive the
// previous plan and refuse the stage anyway — better to find that out before
// anything has run than half way through.
const CapConfigCLIVersion = "cliVersion"

func init() { MustRegisterProbe(CapUpgradeStages, probeUpgradeStages) }

// probeUpgradeStages declares that this node can run a stage of an upgrade plan
// against the machine it is on.
//
// It needs the same two things a stage needs: an olares-cli to run, and a host
// to run it against. Inside a container the second is missing — a stage
// rewrites /etc/containerd and restarts systemd units — so the capability is
// withheld rather than declared and then discovered to be a lie by a stage that
// quietly changed the container instead of the node.
func probeUpgradeStages(_ context.Context, in ProbeInput) (string, Capability, bool) {
	if !CanPowerHost(in.State) || !commandAvailable(collectCommand) {
		return CapUpgradeStages, Capability{}, false
	}
	c := Capability{Supported: true}
	if v := cliVersion(); v != "" {
		c.Config = map[string]any{CapConfigCLIVersion: v}
	}
	return CapUpgradeStages, c, true
}

// cliVersion reports the version of the olares-cli on this machine, which is
// the binary that will actually run an upgrade stage. It is a variable so the
// probe can be tested without one installed.
var cliVersion = readCLIVersion

// readCLIVersion asks the binary rather than assuming it matches this daemon.
// The two are installed by separate steps of an upgrade, and the window where
// they disagree is exactly the window this probe exists to describe.
//
// An empty answer refuses the node the upgrade, because comparing versions is
// the only thing that establishes a stage name means the same work on two
// machines. So this reads the output loosely: it looks for the version
// anywhere in the first line rather than insisting on a field count, since a
// wording change in olares-cli -v would otherwise stop upgrades on every node
// at once, and it would do it through a probe that had gone quiet rather than
// through anything that reads like a fault.
func readCLIVersion() string {
	path, err := lookPath(collectCommand)
	if err != nil {
		return ""
	}
	out, err := exec.Command(path, "-v").Output()
	if err != nil {
		return ""
	}
	// "olares-cli version 1.12.8", and the version is the only field of that
	// line that looks like one.
	for _, field := range strings.Fields(strings.SplitN(string(out), "\n", 2)[0]) {
		// A leading v is dropped rather than rejected. The control node
		// compares this with the version out of the plan, which carries none,
		// so a binary that started printing one would otherwise be refused by
		// every node for a difference in punctuation.
		if v := strings.TrimPrefix(field, "v"); looksLikeVersion(v) {
			return v
		}
	}
	return ""
}

// looksLikeVersion accepts what a release is named: digits and dots, with
// anything a prerelease or build suffix may add. It is deliberately not a
// semver parse — the control node compares this to the version it is rolling
// out and does not otherwise interpret it, so agreeing on the exact grammar
// here would be one more thing to keep in step for no gain.
func looksLikeVersion(s string) bool {
	if len(s) == 0 || !(s[0] >= '0' && s[0] <= '9') {
		return false
	}
	return strings.Contains(s, ".")
}

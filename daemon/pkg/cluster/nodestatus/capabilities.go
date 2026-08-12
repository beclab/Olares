package nodestatus

import (
	"context"
	"os/exec"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// Capability names. They are dotted strings rather than a fixed struct so a
// new capability is a new key: a client that has never heard of it ignores it
// instead of failing to parse the node.
const (
	CapPowerShutdown  = "power.shutdown"
	CapPowerReboot    = "power.reboot"
	CapLogsCollect    = "logs.collect"
	CapSetSSHPassword = "ssh.setPassword"
)

// Capability is what a node can be asked to do. Config carries whatever that
// capability needs to be described, e.g. the power modes a device supports.
type Capability struct {
	Supported bool           `json:"supported"`
	Config    map[string]any `json:"config,omitempty"`
}

// ProbeInput is what a probe may decide on. A capability that depends on who
// this node is, or on how olaresd is deployed, cannot be answered from a
// compile-time constant.
type ProbeInput struct {
	Identity Identity
	State    clistate.State
}

// Probe reports one capability of this node.
//
// Returning ok=false leaves the capability undeclared, which is the answer for
// anything this node cannot confirm — and for anything it is not the right
// node to offer. A declared capability is a promise that the operation works
// here, so nothing is declared on the strength of the binary having the code.
type Probe func(ctx context.Context, in ProbeInput) (name string, c Capability, ok bool)

// The executables the command layer hands to the process executor. A
// capability is only declared once the same name resolves here, so a node
// missing one of them reports what it can actually do.
const (
	shutdownCommand = "shutdown"
	rebootCommand   = "reboot"
	collectCommand  = "olares-cli"
	passwordCommand = "chpasswd"
)

// lookPath resolves the power commands the way the command layer will when it
// runs them, and is a variable so the probes can be tested off a real host.
var lookPath = exec.LookPath

// Detect returns the capabilities of this node.
func Detect(ctx context.Context, in ProbeInput, extra ...Probe) map[string]Capability {
	caps := make(map[string]Capability, len(defaultProbes)+len(extra))
	for _, probe := range append(append([]Probe{}, defaultProbes...), extra...) {
		name, c, ok := probe(ctx, in)
		if !ok || name == "" {
			continue
		}
		caps[name] = c
	}
	return caps
}

var defaultProbes = []Probe{probePowerShutdown, probePowerReboot, probeLogsCollect, probeSetSSHPassword}

// probePowerShutdown declares a shutdown only on a node that is both able and
// entitled to perform one. The control node is deliberately excluded: turning
// it off is the last step of a cluster shutdown, which has to sequence the
// compute nodes first, and a per-node button there would skip that.
func probePowerShutdown(_ context.Context, in ProbeInput) (string, Capability, bool) {
	if in.Identity.Role != inventory.RoleWorker {
		return CapPowerShutdown, Capability{}, false
	}
	if !CanPowerHost(in.State) || !commandAvailable(shutdownCommand) {
		return CapPowerShutdown, Capability{}, false
	}
	return CapPowerShutdown, Capability{Supported: true}, true
}

// probePowerReboot covers the control node too: rebooting it is a legitimate
// single-node operation, unlike powering it off.
func probePowerReboot(_ context.Context, in ProbeInput) (string, Capability, bool) {
	if !CanPowerHost(in.State) || !commandAvailable(rebootCommand) {
		return CapPowerReboot, Capability{}, false
	}
	return CapPowerReboot, Capability{Supported: true}, true
}

// probeLogsCollect holds in a container as well as on a host, since the
// collector runs where olaresd runs. It still needs the collector itself.
func probeLogsCollect(context.Context, ProbeInput) (string, Capability, bool) {
	if !commandAvailable(collectCommand) {
		return CapLogsCollect, Capability{}, false
	}
	return CapLogsCollect, Capability{Supported: true}, true
}

// CanPowerHost reports whether a power command issued here reaches the
// machine. Inside a container it does not: the command would act on the
// container and leave the host running, which is the worst possible answer to
// "is this node off yet".
//
// It is exported because the point that issues the command applies it too,
// rather than trusting that somebody declared the capability earlier.
func CanPowerHost(st clistate.State) bool {
	if st.ContainerMode != nil && *st.ContainerMode != "" {
		return false
	}
	return true
}

func commandAvailable(name string) bool {
	_, err := lookPath(name)
	return err == nil
}

func probeSetSSHPassword(context.Context, ProbeInput) (string, Capability, bool) {
	if !commandAvailable(passwordCommand) {
		return CapSetSSHPassword, Capability{}, false
	}
	return CapSetSSHPassword, Capability{Supported: true}, true
}

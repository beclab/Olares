package nodestatus

import (
	"context"
	"os/exec"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
)

// Capability is what a node can be asked to do. Config carries whatever that
// capability needs to be described, e.g. the power modes a device supports.
//
// Capability names are dotted strings rather than a fixed struct so a new
// capability is a new key: a client that has never heard of it ignores it
// instead of failing to parse the node. Each probe file declares its own Cap*
// constant and registers itself with MustRegisterProbe.
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

// lookPath resolves the executables the command layer will hand to the process
// executor, and is a variable so the probes can be tested off a real host. A
// capability is only declared once the same name resolves here.
var lookPath = exec.LookPath

// Detect returns the capabilities of this node.
func Detect(ctx context.Context, in ProbeInput, extra ...Probe) map[string]Capability {
	registered := defaultProbeRegistry.Probes()
	caps := make(map[string]Capability, len(registered)+len(extra))
	for _, probe := range append(append([]Probe{}, registered...), extra...) {
		name, c, ok := probe(ctx, in)
		if !ok || name == "" {
			continue
		}
		caps[name] = c
	}
	return caps
}

// CanPowerHost reports whether a power command issued here reaches the
// machine. Inside a container it does not: the command would act on the
// container and leave the host running, which is the worst possible answer to
// "is this node off yet".
//
// It is exported because the point that issues the command applies it too,
// rather than trusting that somebody declared the capability earlier. Host-
// affecting probes such as set-ssh-password reuse the same gate.
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

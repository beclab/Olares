package clusterop

import (
	"context"
	"os/exec"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// PowerError is a refusal or failure from the point that would power a
// machine. Code is stable and is what a caller maps; Message is fixed text
// safe to return over the wire; Err is this node's own detail, which stays in
// this node's log.
type PowerError struct {
	Code    string
	Message string
	Err     error
}

func (e *PowerError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *PowerError) Unwrap() error { return e.Err }

// hostPowerModule is the part of a power module that acts on the machine
// this daemon runs on. A module that does not implement it declares that it
// has nothing to run here, and PowerHost refuses rather than guessing.
type hostPowerModule interface {
	// commandName is the executable this operation needs on this node. It is
	// looked up before anything is run, because a node without it cannot
	// carry the operation out however willing the rest of the daemon is.
	commandName() string

	// newCommand builds the same node-local command the single-node power
	// endpoint runs.
	newCommand() commands.Interface
}

// hostPowerModuleFor names the module that would act on this machine. It
// reads the default registry rather than any manager's, because this is the
// execution point: it answers for the node it is running on, whether that
// node is orchestrating anything or was told to do this by one that is.
func hostPowerModuleFor(op Type) (hostPowerModule, error) {
	if module, ok := DefaultRegistry().Lookup(op); ok {
		if host, ok := module.(hostPowerModule); ok {
			return host, nil
		}
	}
	return nil, &PowerError{
		Code:    CodeUnsupportedOperation,
		Message: "this daemon does not perform that operation",
	}
}

// The seams that make the machine actually power off. A unit test replaces
// them; nothing else does.
var (
	newPowerCommand = func(op Type) commands.Interface {
		host, err := hostPowerModuleFor(op)
		if err != nil {
			return nil
		}
		return host.newCommand()
	}

	// The published snapshot rather than the live struct: the refresh loop
	// rewrites that struct field by field while requests are being served, so
	// reading it directly is a data race whose answer is a state nothing ever
	// published. This one decides whether the machine may be powered.
	validatePowerOp = func(cmd commands.Interface) error {
		st, _ := state.Snapshot()
		return state.ValidateOp(st.TerminusState, cmd)
	}

	// hostPowerSupport answers whether a power command issued on this node
	// reaches the machine it is running on.
	hostPowerSupport = func(op Type) error {
		st, _ := state.Snapshot()
		return hostPowerSupportIn(st, op)
	}

	lookPowerCommand = exec.LookPath
)

// hostPowerSupportIn is the execution point's own answer to "can I do this",
// asked of the state this daemon is in rather than of anything a caller sent.
//
// The master's precheck asks every node the same question before it dispatches
// anything, and this repeats it at the moment of execution on purpose: that
// precheck can be minutes stale, can have been answered by a different node,
// and on the single-node routes never happened at all. Getting it wrong inside
// a container means the command restarts the container and leaves the machine
// running, while the operation records that the node was powered.
func hostPowerSupportIn(st clistate.State, op Type) error {
	command, err := powerCommandName(op)
	if err != nil {
		return err
	}
	if !nodestatus.CanPowerHost(st) {
		return &PowerError{
			Code:    CodePowerUnsupported,
			Message: "olaresd runs in a container on this node, so it cannot power the machine",
		}
	}
	if _, err := lookPowerCommand(command); err != nil {
		return &PowerError{
			Code:    CodePowerUnsupported,
			Message: "this node does not have the command this operation needs",
		}
	}
	return nil
}

// LocalPowerSupport is the execution point's answer, asked without executing
// anything. The orchestrator's precheck uses it for the control node, so that
// what the plan assumes about this machine is what PowerHost will decide.
func LocalPowerSupport(op Type) error { return hostPowerSupport(op) }

func powerCommandName(op Type) (string, error) {
	host, err := hostPowerModuleFor(op)
	if err != nil {
		return "", err
	}
	return host.commandName(), nil
}

// PowerHost reboots or shuts down the machine this daemon runs on, through the
// same command and the same state check as the single-node power endpoints. A
// cluster operation is a way to sequence those, not a way around them.
//
// It returns once the command has been accepted. The machine goes down a
// moment later, which is why no caller of this ever observes what came after.
func PowerHost(ctx context.Context, op Type) error {
	// Asked first, and before the command is even built: whether this node
	// can power its own machine does not depend on the daemon happening to be
	// in a state that would otherwise have allowed it.
	if err := hostPowerSupport(op); err != nil {
		return err
	}

	cmd := newPowerCommand(op)
	if cmd == nil {
		return &PowerError{
			Code:    CodeUnsupportedOperation,
			Message: "this daemon does not perform that operation",
		}
	}
	if err := validatePowerOp(cmd); err != nil {
		return &PowerError{
			Code:    CodeHostPowerFailed,
			Message: "this node cannot be powered in its current state",
			Err:     err,
		}
	}
	if _, err := cmd.Execute(ctx, nil); err != nil {
		return &PowerError{
			Code:    CodeHostPowerFailed,
			Message: "this node could not be powered",
			Err:     err,
		}
	}
	return nil
}

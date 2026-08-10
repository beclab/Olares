package clusterop

import (
	"context"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/commands/shutdown"
)

// TypeShutdown is what a caller asks for and what goes on the wire to every
// node. It is declared here because this module is what answers for it.
const TypeShutdown Type = "shutdown"

func init() { registerBuiltInPowerOperation(shutdownModule{}) }

// shutdownModule is the cluster shutdown: every compute node is told to
// switch off, and then the control node is.
//
// It deliberately observes and confirms nothing. There is no boot to compare
// against, nothing comes back to be watched, and a machine that is on again
// later is somebody pressing the power button rather than the operation
// having succeeded — so the record ends at command_issued and no daemon that
// starts afterwards may promote it. That is also why the module implements
// no Recover at all: with no RecoverableModule to ask, the generic recovery
// framework (recovery.go) leaves a command_issued shutdown exactly where it
// was found, and applies MarkInterrupted to anything of this type it loads
// that never got that far.
type shutdownModule struct{}

var shutdownSpec = powerSpec{
	opType:     TypeShutdown,
	capability: nodestatus.CapPowerShutdown,
	awaitDown:  true,
	grace:      func(t Timeouts) time.Duration { return t.Down },

	// Shutting the control node down on its own is not something to offer:
	// it is the last step of a cluster shutdown, and on its own it leaves a
	// cluster whose compute nodes are running and whose control node is not.
	refuseControlNode: &refusal{
		code:   CodePowerUnsupported,
		reason: "the control node cannot be shut down by a node operation",
	},
}

func (shutdownModule) Type() Type { return TypeShutdown }

func (shutdownModule) Validate(req CreateRequest) error { return validatePowerScope(req) }

func (shutdownModule) Phase(Operation) (nodestatus.Phase, bool) {
	return nodestatus.PhaseShuttingDown, true
}

func (shutdownModule) Run(ctx context.Context, rt Runtime, req RunRequest) Outcome {
	return runPowerOperation(ctx, rt, req, shutdownSpec)
}

// ExecuteNode shuts down the machine this daemon runs on, through the same
// execution point and the same state check as the single-node power endpoint.
func (shutdownModule) ExecuteNode(ctx context.Context, _ NodeRequest) error {
	return PowerHost(ctx, TypeShutdown)
}

// builtInPowerOperation says this package wrote the node half above. See the
// marker's own documentation for what that permits.
func (shutdownModule) builtInPowerOperation() {}

func (shutdownModule) commandName() string { return "shutdown" }

func (shutdownModule) newCommand() commands.Interface { return shutdown.New() }

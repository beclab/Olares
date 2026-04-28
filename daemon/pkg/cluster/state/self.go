package state

import clistate "github.com/beclab/Olares/cli/pkg/daemon/state"

// TerminusDState is the lifecycle state of olaresd itself. The
// canonical definition lives in the cli module and is re-exported
// here as an alias so the daemon can use the unqualified name.
type TerminusDState = clistate.TerminusDState

const (
	Initialize = clistate.Initialize
	Running    = clistate.Running
)

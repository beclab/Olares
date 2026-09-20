package nodestatus

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// CapPowerShutdown is declared only on a worker that can reach the host power
// button. Clients that have never heard of it ignore the key.
const CapPowerShutdown = "power.shutdown"

const shutdownCommand = "shutdown"

func init() { MustRegisterProbe(CapPowerShutdown, probePowerShutdown) }

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

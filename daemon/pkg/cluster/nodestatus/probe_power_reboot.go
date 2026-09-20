package nodestatus

import "context"

// CapPowerReboot covers the control node too: rebooting it is a legitimate
// single-node operation, unlike powering it off.
const CapPowerReboot = "power.reboot"

const rebootCommand = "reboot"

func init() { MustRegisterProbe(CapPowerReboot, probePowerReboot) }

func probePowerReboot(_ context.Context, in ProbeInput) (string, Capability, bool) {
	if !CanPowerHost(in.State) || !commandAvailable(rebootCommand) {
		return CapPowerReboot, Capability{}, false
	}
	return CapPowerReboot, Capability{Supported: true}, true
}

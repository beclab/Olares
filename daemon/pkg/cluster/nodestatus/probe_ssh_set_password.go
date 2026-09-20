package nodestatus

import "context"

// CapSetSSHPassword is declared only when chpasswd can change the host login
// password. Clients that have never heard of it ignore the key.
const CapSetSSHPassword = "ssh.setPassword"

const passwordCommand = "chpasswd"

func init() { MustRegisterProbe(CapSetSSHPassword, probeSetSSHPassword) }

// probeSetSSHPassword needs the same host reachability as power: inside a
// container chpasswd would rewrite the container's users and leave the host
// SSH password untouched.
func probeSetSSHPassword(_ context.Context, in ProbeInput) (string, Capability, bool) {
	if !CanPowerHost(in.State) || !commandAvailable(passwordCommand) {
		return CapSetSSHPassword, Capability{}, false
	}
	return CapSetSSHPassword, Capability{Supported: true}, true
}

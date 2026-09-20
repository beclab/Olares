package nodestatus

import "context"

// CapLogsCollect holds in a container as well as on a host, since the
// collector runs where olaresd runs. It still needs the collector itself.
const CapLogsCollect = "logs.collect"

const collectCommand = "olares-cli"

func init() { MustRegisterProbe(CapLogsCollect, probeLogsCollect) }

func probeLogsCollect(context.Context, ProbeInput) (string, Capability, bool) {
	if !commandAvailable(collectCommand) {
		return CapLogsCollect, Capability{}, false
	}
	return CapLogsCollect, Capability{Supported: true}, true
}

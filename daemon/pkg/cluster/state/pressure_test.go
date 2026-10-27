package state

import (
	"context"
	"errors"
	"testing"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

func withThisNodeName(t *testing.T, fn func(context.Context) (string, error)) {
	t.Helper()
	prev := thisNodeName
	thisNodeName = fn
	t.Cleanup(func() { thisNodeName = prev })
}

func TestPressureForNodeKeepsWhatWasReported(t *testing.T) {
	pressures := map[string][]utils.NodePressure{
		"worker-1": {{Type: "MemoryPressure", Message: "kubelet has insufficient memory"}},
	}

	got := pressureForNode(pressures, "worker-1")
	if len(got) != 1 || got[0].Type != "MemoryPressure" {
		t.Fatalf("pressures reported by kubelet were dropped: %+v", got)
	}
}

func TestPressureForNodeUnknownNodeIsEmpty(t *testing.T) {
	pressures := map[string][]utils.NodePressure{
		"worker-1": {{Type: "DiskPressure"}},
	}

	if got := pressureForNode(pressures, "worker-2"); got != nil {
		t.Errorf("another node's pressures leaked in: %+v", got)
	}
	if got := pressureForNode(pressures, ""); got != nil {
		t.Errorf("an unresolved node name must not match anything: %+v", got)
	}
}

// The map GetNodesPressure builds is keyed by the Kubernetes node name. The OS
// hostname is a different string on any node whose kubelet was given a name of
// its own, and keying by it silently reported every node as pressure-free.
func TestRefreshNodePressureKeysOnTheKubernetesNodeName(t *testing.T) {
	restoreState(t)
	hostname := "worker-1.lan"
	setCurrentState(t, clistate.State{HostName: &hostname})

	withThisNodeName(t, func(context.Context) (string, error) { return "worker-1", nil })

	pressures := map[string][]utils.NodePressure{
		"worker-1": {{Type: "MemoryPressure", Message: "kubelet has insufficient memory"}},
	}

	got := refreshNodePressure(context.Background(), pressures)
	if len(got) != 1 {
		t.Fatalf("want this node's pressure, got %+v (hostname %q must not be the key)", got, hostname)
	}
}

func TestRefreshNodePressureWithoutANodeNameIsEmpty(t *testing.T) {
	withThisNodeName(t, func(context.Context) (string, error) { return "", errors.New("cluster unreachable") })

	got := refreshNodePressure(context.Background(), map[string][]utils.NodePressure{
		"worker-1": {{Type: "MemoryPressure"}},
	})
	if got != nil {
		t.Errorf("a node that cannot name itself must not claim another node's pressures: %+v", got)
	}
}

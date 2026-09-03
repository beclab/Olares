package upgrade

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

func master(name string, self bool) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: self}
}

func worker(name, ip string) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleWorker, IP: ip, Ready: true}
}

// Whether an upgrade is this machine's own business is not the same question
// as whether this machine can schedule other people's.
//
// Answering both with "run it here" is how a cluster ends up half upgraded:
// the control node moves to the new version and the rest stay behind, with
// nothing recording that they did. A directory listing other nodes is enough
// to know this upgrade is not a local one, whatever it says about this
// machine.
func TestWhoMayCarryOutAnUpgrade(t *testing.T) {
	for _, tc := range []struct {
		name     string
		nodes    []inventory.Node
		cluster  bool
		wantErr  string
		wantLoud bool
	}{
		{
			name:  "the only machine there is",
			nodes: []inventory.Node{master("master-1", true)},
		},
		{
			name:  "an empty directory is not a cluster",
			nodes: nil,
		},
		{
			name:     "a control node with workers needs an orchestrator",
			nodes:    []inventory.Node{master("master-1", true), worker("worker-1", "10.0.0.2")},
			wantErr:  "no orchestrator",
			wantLoud: true,
		},
		{
			name:     "workers exist and this machine is not the control node",
			nodes:    []inventory.Node{master("master-1", false), worker("worker-1", "10.0.0.2")},
			wantErr:  "does not identify this machine",
			wantLoud: true,
		},
		{
			name:     "this machine is one of the workers",
			nodes:    []inventory.Node{master("master-1", false), inventory.Node{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true, IsSelf: true}},
			wantErr:  "does not identify this machine",
			wantLoud: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, across, err := orchestratorFor(tc.nodes, nil)
			switch {
			case tc.wantLoud && err == nil:
				t.Fatalf("no error, want one mentioning %q; a silent local upgrade here leaves the"+
					" other nodes on the old version", tc.wantErr)
			case !tc.wantLoud && err != nil:
				t.Fatalf("error %v, want this machine to upgrade itself", err)
			case tc.wantLoud && !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("error = %v, want it to mention %q", err, tc.wantErr)
			}
			if across {
				t.Error("reported as scheduled across the cluster without an orchestrator")
			}
		})
	}
}

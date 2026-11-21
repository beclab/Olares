package upgrade

import (
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
)

func TestUpgradeRequestIDChangesOnlyAfterNewAuthorization(t *testing.T) {
	v := semver.MustParse("1.12.8")
	old := state.UpgradeTarget{Version: *v, RequestNonce: "first"}
	newTarget := state.UpgradeTarget{Version: *v, RequestNonce: "second"}
	if upgradeRequestID(old) != upgradeRequestID(old) {
		t.Fatal("same target changed request ID")
	}
	if upgradeRequestID(old) == upgradeRequestID(newTarget) {
		t.Fatal("new authorization reused a cancelled request ID")
	}
	if got := upgradeRequestID(state.UpgradeTarget{Version: *v}); got != "olares-upgrade-1.12.8" {
		t.Fatalf("legacy target request ID = %q", got)
	}
}

func TestClusterUpgradeProgressUsesTheWholePlan(t *testing.T) {
	op := clusterop.Operation{TotalSteps: 15, Steps: []clusterop.Step{
		{Status: clusterop.StepSucceeded}, {Status: clusterop.StepSucceeded},
	}}
	before := upgradeProgressOf(op)
	if before >= 20 {
		t.Fatalf("two of fifteen steps reported %d%%", before)
	}
	op.Steps = append(op.Steps, clusterop.Step{Status: clusterop.StepSucceeded})
	after := upgradeProgressOf(op)
	if after <= before || after >= 30 {
		t.Fatalf("progress %d -> %d, expected one-step increase", before, after)
	}
}

func master(name string, self bool) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: self}
}

func worker(name, ip string) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleWorker, IP: ip, Ready: true}
}

// Every upgrade is scheduled by the orchestrator on the control node, a single
// machine included. A machine that cannot schedule it does not fall back to
// upgrading itself: on a cluster that leaves the other nodes on the old
// version, with nothing recording that they are.
func TestWhoMayCarryOutAnUpgrade(t *testing.T) {
	for _, tc := range []struct {
		name    string
		nodes   []inventory.Node
		wantErr string
	}{
		{"the only machine there is, without an orchestrator",
			[]inventory.Node{master("master-1", true)}, "no orchestrator"},
		{"a control node with workers, without an orchestrator",
			[]inventory.Node{master("master-1", true), worker("worker-1", "10.0.0.2")}, "no orchestrator"},
		{"an empty directory", nil, "does not identify this machine"},
		{"workers exist and this machine is not the control node",
			[]inventory.Node{master("master-1", false), worker("worker-1", "10.0.0.2")}, "does not identify this machine"},
		{"this machine is one of the workers",
			[]inventory.Node{master("master-1", false), {NodeName: "worker-1", Role: inventory.RoleWorker,
				IP: "10.0.0.2", Ready: true, IsSelf: true}}, "does not identify this machine"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := orchestratorFor(tc.nodes, nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want it to mention %q", err, tc.wantErr)
			}
			if m != nil {
				t.Error("returned an orchestrator it does not have")
			}
		})
	}
	m := &clusterop.Manager{}
	if got, err := orchestratorFor([]inventory.Node{master("master-1", true)}, m); err != nil || got != m {
		t.Fatalf("a single control node with an orchestrator was refused: %v", err)
	}
}

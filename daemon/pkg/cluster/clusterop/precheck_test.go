package clusterop

import (
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// "This node does not support reboot" is a claim about the node. A node whose
// role or name the directory could not resolve has made no claim at all, and
// saying it did sends an operator looking at the wrong machine.
func TestPrecheckReportsANodeItCannotIdentify(t *testing.T) {
	for _, tc := range []struct {
		name string
		node inventory.Node
	}{
		{
			name: "no role",
			node: inventory.Node{NodeName: "worker-1", Role: inventory.RoleUnknown, IP: "10.0.0.2", Ready: true},
		},
		{
			name: "no name",
			node: inventory.Node{NodeName: "", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"), tc.node)
			m, _ := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

			if op.Status != StatusFailed || op.Code != CodePrecheckFailed {
				t.Fatalf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodePrecheckFailed)
			}
			got := nodeResult(t, op, tc.node.NodeName)
			if got.Code != CodeNodeIdentityUnknown {
				t.Errorf("code = %q, want %s", got.Code, CodeNodeIdentityUnknown)
			}
			for _, e := range c.log() {
				if e == "inspect "+tc.node.NodeName {
					t.Error("an unidentifiable node was asked what it can do")
				}
			}
		})
	}
}

// A NotReady worker will not answer, and waiting fifteen minutes to discover
// that is fifteen minutes the user spends watching a progress bar for a
// cluster reboot that was never going to work.
func TestPrecheckRefusesANotReadyWorkerImmediately(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.nodes[1].Ready = false
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodePrecheckFailed {
		t.Fatalf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodePrecheckFailed)
	}
	if got := nodeResult(t, op, "worker-1"); got.Code != CodeNodeNotReady {
		t.Errorf("code = %q, want %s", got.Code, CodeNodeNotReady)
	}
	for _, e := range c.log() {
		if e == "inspect worker-1" {
			t.Error("a NotReady node was dialled anyway")
		}
	}
	requireHostUntouched(t, c)
}

// The order of the per-node checks is the order in which the answers are
// useful. Being unaddressable explains being unreachable; not being Ready
// explains both.
func TestPrecheckReportsTheFirstThingThatIsWrong(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", ""))
	c.nodes[1].Ready = false
	c.nodes[1].Role = inventory.RoleUnknown
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if got := nodeResult(t, op, "worker-1"); got.Code != CodeNodeIdentityUnknown {
		t.Errorf("code = %q, want identity to be reported before anything derived from it", got.Code)
	}
}

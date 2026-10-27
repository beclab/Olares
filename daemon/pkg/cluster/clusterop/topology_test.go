package clusterop

import (
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// One master and its workers is the only topology this daemon can sequence
// correctly: the control node has to go last, and with two of them the second
// one is powered off by the first one's run with nobody watching.
func TestSplitClusterRefusesAnythingButOneMaster(t *testing.T) {
	for _, tc := range []struct {
		name  string
		nodes []inventory.Node
		code  string
	}{
		{
			name: "two masters",
			nodes: []inventory.Node{
				master("master-1", "10.0.0.1"),
				{NodeName: "master-2", Role: inventory.RoleMaster, IP: "10.0.0.2", Ready: true},
				worker("worker-1", "10.0.0.3"),
			},
			code: CodeUnsupportedTopology,
		},
		{
			name:  "no master at all",
			nodes: []inventory.Node{worker("worker-1", "10.0.0.2")},
			code:  CodeUnsupportedTopology,
		},
		{
			name:  "an empty directory",
			nodes: nil,
			code:  CodeUnsupportedTopology,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, code, err := splitCluster(tc.nodes)
			if err == nil {
				t.Fatal("an unsupported topology was planned anyway")
			}
			if code != tc.code {
				t.Errorf("code = %q, want %s", code, tc.code)
			}
		})
	}
}

// Falling back to "the first master in the list" is how a daemon ends up
// orchestrating a cluster it is not the control node of.
func TestSplitClusterRefusesToGuessWhichNodeItIs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		nodes []inventory.Node
	}{
		{
			name: "no node is this machine",
			nodes: []inventory.Node{
				{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true},
				worker("worker-1", "10.0.0.2"),
			},
		},
		{
			name: "this machine is a worker",
			nodes: []inventory.Node{
				{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true},
				{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true, IsSelf: true},
			},
		},
		{
			name: "two nodes claim to be this machine",
			nodes: []inventory.Node{
				master("master-1", "10.0.0.1"),
				{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true, IsSelf: true},
			},
		},
		{
			name: "the control node has no name",
			nodes: []inventory.Node{
				{NodeName: "", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
				worker("worker-1", "10.0.0.2"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, code, err := splitCluster(tc.nodes)
			if err == nil {
				t.Fatal("the plan named a control node it could not identify")
			}
			if code != CodeSelfUnresolved {
				t.Errorf("code = %q, want %s", code, CodeSelfUnresolved)
			}
		})
	}
}

func TestSplitClusterPlansOneMasterAndItsWorkers(t *testing.T) {
	p, code, err := splitCluster([]inventory.Node{
		worker("worker-1", "10.0.0.2"),
		master("master-1", "10.0.0.1"),
		worker("worker-2", "10.0.0.3"),
	})
	if err != nil {
		t.Fatalf("splitCluster: %v (%s)", err, code)
	}
	if p.master.NodeName != "master-1" {
		t.Errorf("control node = %q", p.master.NodeName)
	}
	if len(p.workers) != 2 {
		t.Fatalf("workers = %+v", p.workers)
	}
	for _, w := range p.workers {
		if w.IsSelf {
			t.Errorf("the control node was also planned as a compute node: %+v", w)
		}
	}
}

// The gate runs before anything is inspected, dispatched or powered. A
// multi-master cluster must not have a single node touched by an operation
// this daemon is going to refuse anyway.
func TestAnUnsupportedTopologyStopsBeforeTheClusterIsTouched(t *testing.T) {
	for _, ty := range []Type{TypeReboot, TypeShutdown} {
		t.Run(string(ty), func(t *testing.T) {
			c := newCluster(
				master("master-1", "10.0.0.1"),
				inventory.Node{NodeName: "master-2", Role: inventory.RoleMaster, IP: "10.0.0.2", Ready: true},
				worker("worker-1", "10.0.0.3"),
			)
			m, _ := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, ty, "client-1").ID)

			if op.Status != StatusFailed || op.Code != CodeUnsupportedTopology {
				t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodeUnsupportedTopology)
			}
			if d, p := c.counts(); d != 0 || p != 0 {
				t.Errorf("dispatch=%d power=%d on a topology this daemon refuses", d, p)
			}
			for _, e := range c.log() {
				if len(e) >= 7 && e[:7] == "inspect" {
					t.Errorf("a node was inspected before the topology was checked: %v", c.log())
					break
				}
			}
		})
	}
}

func TestAnUnidentifiableControlNodeStopsBeforeTheClusterIsTouched(t *testing.T) {
	c := newCluster(
		inventory.Node{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true},
		worker("worker-1", "10.0.0.2"),
	)
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodeSelfUnresolved {
		t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodeSelfUnresolved)
	}
	if d, p := c.counts(); d != 0 || p != 0 {
		t.Errorf("dispatch=%d power=%d without a resolvable control node", d, p)
	}
}

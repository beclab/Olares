package clusterop

import (
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

func nodeResult(t *testing.T, op Operation, name string) NodeResult {
	t.Helper()
	for _, n := range op.Nodes {
		if n.NodeName == name {
			return n
		}
	}
	t.Fatalf("no result recorded for node %q: %+v", name, op.Nodes)
	return NodeResult{}
}

func step(t *testing.T, op Operation, name string) Step {
	t.Helper()
	for _, s := range op.Steps {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("no step %q recorded: %+v", name, op.Steps)
	return Step{}
}

func hasStep(op Operation, name string) bool {
	for _, s := range op.Steps {
		if s.Name == name {
			return true
		}
	}
	return false
}

func requireHostUntouched(t *testing.T, c *cluster) {
	t.Helper()
	if _, power := c.counts(); power != 0 {
		t.Errorf("the control node was powered %d times by an operation that should have stopped first", power)
	}
}

// A shutdown that skipped a node the master could not reach leaves that node
// running while the user is told the cluster is off. There is no force.
func TestPrecheckRefusesTheWholeOperationForOneBadNode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		break_ func(*cluster)
		code   string
	}{
		{
			name:   "no address",
			break_: func(c *cluster) { c.nodes[1].IP = "" },
			code:   CodeNodeUnaddressable,
		},
		{
			name: "unreachable",
			break_: func(c *cluster) {
				c.inspectErr["worker-1"] = errors.New("dial tcp 10.0.0.2:18088: connect: no route to host")
			},
			code: CodeNodeUnreachable,
		},
		{
			name:   "cannot power itself",
			break_: func(c *cluster) { delete(c.capabilities["worker-1"], nodestatus.CapPowerReboot) },
			code:   CodePowerUnsupported,
		},
		{
			name: "declares the capability as unsupported",
			break_: func(c *cluster) {
				c.capabilities["worker-1"][nodestatus.CapPowerReboot] = nodestatus.Capability{Supported: false}
			},
			code: CodePowerUnsupported,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
			tc.break_(c)
			m, _ := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

			if op.Status != StatusFailed {
				t.Errorf("status = %q, want failed", op.Status)
			}
			if op.Code != CodePrecheckFailed {
				t.Errorf("operation code = %q, want %s", op.Code, CodePrecheckFailed)
			}
			if got := nodeResult(t, op, "worker-1"); got.Code != tc.code {
				t.Errorf("worker code = %q, want %q", got.Code, tc.code)
			}
			if got := nodeResult(t, op, "master-1"); got.Status != NodeSkipped {
				t.Errorf("control node status = %q, want skipped", got.Status)
			}
			if d, p := c.counts(); d != 0 || p != 0 {
				t.Errorf("a blocked precheck still touched the cluster: dispatch=%d power=%d", d, p)
			}
			if hasStep(op, StepWorkerCommand) {
				t.Error("the operation started dispatching after a failed precheck")
			}
			if s := step(t, op, StepPrecheck); s.Status != StepFailed || s.FinishedAt == nil {
				t.Errorf("precheck step not settled: %+v", s)
			}
		})
	}
}

func TestPrecheckAsksForTheCapabilityTheOperationNeeds(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	delete(c.capabilities["worker-1"], nodestatus.CapPowerShutdown)
	c.goesDownAndComesBack("worker-1")
	m, _ := newManager(t, c)

	rebooted := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)
	if rebooted.Status != StatusCommandIssued {
		t.Fatalf("a missing shutdown capability blocked a reboot: %+v", rebooted)
	}

	shutdown := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-2").ID)
	if shutdown.Status != StatusFailed {
		t.Fatalf("status = %q, want the shutdown refused", shutdown.Status)
	}
	if got := nodeResult(t, shutdown, "worker-1"); got.Code != CodePowerUnsupported {
		t.Errorf("worker code = %q, want %s", got.Code, CodePowerUnsupported)
	}
}

// The control node never declares a single-node shutdown of its own: turning
// it off is the last step of a cluster operation, not a button. Requiring it
// to declare one would make every cluster shutdown impossible.
func TestPrecheckDoesNotRequireTheControlNodeToDeclareShutdown(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.capabilities["master-1"] = map[string]nodestatus.Capability{}
	c.inspectErr["master-1"] = errors.New("the control node must not be asked")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Errorf("status = %q, want command_issued: %+v", op.Status, op)
	}
}

func TestPrecheckFailsWhenTheNodeDirectoryCannotBeRead(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.inventoryErr = errors.New(`Get "https://10.0.0.1:6443/api/v1/nodes": connection refused`)
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodeInventoryUnavailable {
		t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodeInventoryUnavailable)
	}
	requireHostUntouched(t, c)
}

func TestPrecheckFailsWithoutAControlNode(t *testing.T) {
	c := newCluster(worker("worker-1", "10.0.0.2"))
	c.nodes[0].IsSelf = false
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodeUnsupportedTopology {
		t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodeUnsupportedTopology)
	}
	requireHostUntouched(t, c)
}

// Compute nodes first, control node last. The control node is the machine
// running this code: anything sequenced after it does not happen.
func TestRebootPowersTheControlNodeLast(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.goesDownAndComesBack("worker-1")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued: %+v", op.Status, op)
	}
	events := strings.Join(c.log(), " | ")
	dispatch := strings.Index(events, "dispatch worker-1")
	power := strings.Index(events, "power self reboot")
	if dispatch < 0 || power < 0 || dispatch > power {
		t.Errorf("the control node was not rebooted last: %s", events)
	}
	if got := nodeResult(t, op, "worker-1"); got.Status != NodeRestarted {
		t.Errorf("worker status = %q, want restarted", got.Status)
	}
	if got := nodeResult(t, op, "master-1"); got.Status != NodeCommandIssued {
		t.Errorf("control node status = %q, want command_issued", got.Status)
	}
	for _, name := range []string{StepPrecheck, StepWorkerCommand, StepWorkerRestart, StepMasterCommand} {
		s := step(t, op, name)
		if s.StartedAt == nil || s.FinishedAt == nil {
			t.Errorf("step %q has no timing: %+v", name, s)
		}
	}
}

// A node that never stopped answering never rebooted. Treating it as restarted
// would take the control node down next, on a cluster that is not in the state
// the operation claims it is.
func TestRebootRefusesToProceedOnANodeThatNeverWentDown(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	// Nothing is scripted, so the node stays Ready on the boot it started on.
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if got := nodeResult(t, op, "worker-1"); got.Code != CodeNodeDidNotGoDown {
		t.Errorf("worker code = %q, want %s", got.Code, CodeNodeDidNotGoDown)
	}
	if op.Status != StatusFailed {
		t.Errorf("status = %q, want failed", op.Status)
	}
	requireHostUntouched(t, c)
	if got := nodeResult(t, op, "master-1"); got.Status != NodeSkipped {
		t.Errorf("control node status = %q, want skipped", got.Status)
	}
}

func TestRebootReportsANodeThatNeverCameBack(t *testing.T) {
	c := newCluster(
		master("master-1", "10.0.0.1"),
		worker("worker-1", "10.0.0.2"),
		worker("worker-2", "10.0.0.3"),
	)
	c.goesDownAndComesBack("worker-1")
	c.neverComesBack("worker-2")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusPartiallyFailed {
		t.Errorf("status = %q, want partially_failed with one node back and one gone", op.Status)
	}
	if got := nodeResult(t, op, "worker-1"); got.Status != NodeRestarted {
		t.Errorf("worker-1 status = %q, want restarted", got.Status)
	}
	got := nodeResult(t, op, "worker-2")
	if got.Status != NodeFailed || got.Code != CodeRestartTimeout {
		t.Errorf("worker-2 = %q/%q, want failed/%s", got.Status, got.Code, CodeRestartTimeout)
	}
	requireHostUntouched(t, c)
}

func TestRebootStopsWhenANodeDoesNotAcceptTheCommand(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.dispatchErr["worker-1"] = errors.New("node returned 403")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodeWorkerCommandFailed {
		t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodeWorkerCommandFailed)
	}
	if got := nodeResult(t, op, "worker-1"); got.Status != NodeFailed || got.Code != CodeDispatchFailed {
		t.Errorf("worker = %q/%q, want failed/%s", got.Status, got.Code, CodeDispatchFailed)
	}
	requireHostUntouched(t, c)
	if hasStep(op, StepWorkerRestart) {
		t.Error("the operation waited for a restart it never asked for")
	}
}

func TestSingleNodeClusterRebootsItself(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Errorf("status = %q, want command_issued", op.Status)
	}
	if d, p := c.counts(); d != 0 || p != 1 {
		t.Errorf("dispatch=%d power=%d, want nothing dispatched and one local reboot", d, p)
	}
	if len(op.Nodes) != 1 || op.Nodes[0].Status != NodeCommandIssued {
		t.Errorf("node results = %+v", op.Nodes)
	}
}

func TestControlNodeThatCannotBePoweredIsReported(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.powerSelfErr = errors.New("shutdown: command not found")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusPartiallyFailed || op.Code != CodeHostPowerFailed {
		t.Errorf("status = %q code = %q, want partially_failed/%s", op.Status, op.Code, CodeHostPowerFailed)
	}
	if got := nodeResult(t, op, "worker-1"); got.Status != NodeCommandIssued {
		t.Errorf("the compute node's result was lost: %+v", got)
	}
	if got := nodeResult(t, op, "master-1"); got.Code != CodeHostPowerFailed {
		t.Errorf("control node code = %q, want %s", got.Code, CodeHostPowerFailed)
	}
}

// Nothing waits for a node to come back from a shutdown, and the operation
// ends at command_issued: the last thing it did was tell the machine running
// it to switch off.
func TestShutdownIssuesCommandsAndClaimsNothingMore(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued", op.Status)
	}
	if op.Status == StatusSucceeded {
		t.Error("a shutdown must never report success it cannot observe")
	}
	if c.observeCount() != 0 {
		t.Errorf("a shutdown watched the cluster %d times; there is nothing to wait for", c.observeCount())
	}
	if hasStep(op, StepWorkerRestart) {
		t.Error("a shutdown recorded a restart step")
	}
	events := strings.Join(c.log(), " | ")
	if strings.Index(events, "dispatch worker-1") > strings.Index(events, "power self shutdown") {
		t.Errorf("the control node was shut down before the compute node: %s", events)
	}
	for _, name := range []string{"worker-1", "master-1"} {
		if got := nodeResult(t, op, name); got.Status != NodeCommandIssued {
			t.Errorf("%s status = %q, want command_issued", name, got.Status)
		}
	}
}

func TestShutdownStopsWhenANodeDoesNotAcceptTheCommand(t *testing.T) {
	c := newCluster(
		master("master-1", "10.0.0.1"),
		worker("worker-1", "10.0.0.2"),
		worker("worker-2", "10.0.0.3"),
	)
	c.dispatchErr["worker-2"] = errors.New("dial tcp: i/o timeout")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusPartiallyFailed {
		t.Errorf("status = %q, want partially_failed", op.Status)
	}
	requireHostUntouched(t, c)
	if got := nodeResult(t, op, "master-1"); got.Status != NodeSkipped {
		t.Errorf("control node status = %q, want skipped", got.Status)
	}
}

// The record has to be on disk before the machine is told to go down. After
// that command there may be no process left to write anything.
func TestTheRecordIsDurableBeforeTheControlNodeIsPowered(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, dir := newManager(t, c)

	var (
		seen Operation
		read bool
	)
	c.onPowerSelf = func() {
		store, err := NewStore(dir)
		if err != nil {
			t.Errorf("NewStore: %v", err)
			return
		}
		ops, err := store.List()
		if err != nil || len(ops) != 1 {
			t.Errorf("List: %v (%d records)", err, len(ops))
			return
		}
		seen, read = ops[0], true
	}

	awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if !read {
		t.Fatal("the record was not readable while the control node was being powered")
	}
	if got := nodeResult(t, seen, "worker-1"); got.Status != NodeCommandIssued {
		t.Errorf("the compute node's outcome was not on disk yet: %+v", got)
	}
	if s := step(t, seen, StepMasterCommand); s.Status != StepRunning || s.StartedAt == nil {
		t.Errorf("the control node step was not recorded before the command: %+v", s)
	}
}

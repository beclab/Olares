package clusterop

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// rebootProgress is the whole proof that a node restarted, and it is a pure
// function of what the cluster reports so it can be stated exhaustively.
func TestRebootProgressReadsTheClustersOwnView(t *testing.T) {
	const baseline = "boot-1"

	for _, tc := range []struct {
		name     string
		obs      inventory.Observation
		present  bool
		down, up bool
	}{
		{
			name:    "still Ready on the boot it was told to leave",
			obs:     inventory.Observation{Ready: true, BootID: baseline},
			present: true,
		},
		{
			name:    "NotReady on the same boot: on its way down",
			obs:     inventory.Observation{Ready: false, BootID: baseline},
			present: true,
			down:    true,
		},
		{
			name:    "gone from the directory",
			present: false,
			down:    true,
		},
		{
			name:    "back on a new boot but not Ready yet",
			obs:     inventory.Observation{Ready: false, BootID: "boot-2"},
			present: true,
			down:    true,
		},
		{
			name:    "Ready on a new boot",
			obs:     inventory.Observation{Ready: true, BootID: "boot-2"},
			present: true,
			down:    true,
			up:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			down, up := rebootProgress(tc.obs, tc.present, baseline)
			if down != tc.down || up != tc.up {
				t.Errorf("down=%v up=%v, want down=%v up=%v", down, up, tc.down, tc.up)
			}
		})
	}
}

// A node whose kubelet flapped is Ready again on the boot it started on. It
// never rebooted, and calling it restarted takes the control node down next.
func TestRebootRefusesANodeThatCameBackOnTheSameBoot(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.comesBackOnTheSameBoot("worker-1")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	got := nodeResult(t, op, "worker-1")
	if got.Status != NodeFailed || got.Code != CodeRestartTimeout {
		t.Errorf("worker = %q/%q, want failed/%s", got.Status, got.Code, CodeRestartTimeout)
	}
	requireHostUntouched(t, c)
}

// A rebooting node usually leaves the directory entirely for a while. That is
// the clearest possible evidence it went down.
func TestRebootAcceptsANodeThatLeftTheDirectoryAndCameBack(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.vanishesAndComesBack("worker-1")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if got := nodeResult(t, op, "worker-1"); got.Status != NodeRestarted {
		t.Errorf("worker = %q/%q, want restarted", got.Status, got.Code)
	}
	if op.Status != StatusCommandIssued {
		t.Errorf("status = %q, want command_issued", op.Status)
	}
}

// Waiting for a node to come back must not need the user's access token. The
// node-local status endpoint is behind one, and a cluster mid-reboot is
// exactly when the identity provider that issues them is unavailable.
func TestRebootWaitsOnTheClusterRatherThanOnEachNodesOwnEndpoint(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.vanishesAndComesBack("worker-1")
	m, _ := newManager(t, c)

	awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	events := c.log()
	dispatched := false
	for _, e := range events {
		if strings.HasPrefix(e, "dispatch ") {
			dispatched = true
			continue
		}
		if dispatched && strings.HasPrefix(e, "inspect ") {
			t.Fatalf("a restart probe called a node's own endpoint: %v", events)
		}
	}
	if c.observeCount() == 0 {
		t.Error("nothing observed the cluster while a node was restarting")
	}
}

// A reboot needs a baseline to compare against. Without one there is no proof
// available, and the operation says so instead of guessing from Ready alone.
func TestPrecheckRefusesARebootItCannotProve(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.obs["worker-1"] = inventory.Observation{Ready: true}
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if op.Status != StatusFailed || op.Code != CodePrecheckFailed {
		t.Errorf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodePrecheckFailed)
	}
	if got := nodeResult(t, op, "worker-1"); got.Code != CodeBootIDUnavailable {
		t.Errorf("worker code = %q, want %s", got.Code, CodeBootIDUnavailable)
	}
	requireHostUntouched(t, c)
}

// A shutdown proves nothing about a restart, so it needs no baseline.
func TestShutdownDoesNotNeedABootID(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.obs["worker-1"] = inventory.Observation{Ready: true}
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Errorf("status = %q, want command_issued: %+v", op.Status, op)
	}
}

// The control node's own reboot ends at command_issued because the process
// that would confirm it is the one going down. The next daemon can confirm it,
// and only by seeing that the machine is on a different boot.
func TestARebootedControlNodeIsConfirmedByTheNextDaemon(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-1"
	m, dir := newManager(t, c)

	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "client-1", Scope: ScopeNode, Target: "master-1",
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := op.ID
	op = awaitTerminal(t, m, id)
	if op.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued", op.Status)
	}

	c.hostBootID = "host-boot-2"
	c.obs["master-1"] = inventory.Observation{Ready: true, BootID: "host-boot-2"}
	restarted := newManagerAt(t, c, dir)

	got := awaitOperationStatus(t, restarted, id, StatusSucceeded)
	if got.Status != StatusSucceeded {
		t.Errorf("status = %q, want succeeded once the machine is on a new boot", got.Status)
	}
	if got := nodeResult(t, got, "master-1"); got.Status != NodeRestarted {
		t.Errorf("control node = %q, want restarted", got.Status)
	}
}

func awaitOperationStatus(t *testing.T, m *Manager, id string, want Status) Operation {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if op, ok := m.Get(id); ok && op.Status == want {
			return op
		}
		time.Sleep(time.Millisecond)
	}
	op, _ := m.Get(id)
	return op
}

func TestARebootedControlNodeIsNotConfirmedBeforeItIsReady(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-1"
	m, dir := newManager(t, c)
	id := createOp(t, m, TypeReboot, "client-1").ID
	awaitTerminal(t, m, id)

	c.hostBootID = "host-boot-2"
	c.obs["master-1"] = inventory.Observation{Ready: false, BootID: "host-boot-2"}
	restarted := newManagerAt(t, c, dir)

	got, _ := restarted.Get(id)
	if got.Status != StatusCommandIssued {
		t.Errorf("status = %q, want command_issued until the control node is Ready", got.Status)
	}
}

// olaresd restarting is not the machine restarting. Reporting success for a
// reboot that never happened is worse than reporting nothing.
func TestARestartedDaemonOnTheSameBootConfirmsNothing(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-1"
	m, dir := newManager(t, c)

	id := createOp(t, m, TypeReboot, "client-1").ID
	awaitTerminal(t, m, id)

	restarted := newManagerAt(t, c, dir)

	got, _ := restarted.Get(id)
	if got.Status != StatusCommandIssued {
		t.Errorf("status = %q, want it left at command_issued", got.Status)
	}
}

// A shutdown is not promoted by a boot change: the machine coming back on is
// somebody pressing the power button, not the operation succeeding.
func TestAShutdownIsNeverPromotedByARestart(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-1"
	m, dir := newManager(t, c)

	id := createOp(t, m, TypeShutdown, "client-1").ID
	awaitTerminal(t, m, id)

	c.hostBootID = "host-boot-2"
	restarted := newManagerAt(t, c, dir)

	got, _ := restarted.Get(id)
	if got.Status != StatusCommandIssued {
		t.Errorf("status = %q, want it left at command_issued", got.Status)
	}
}

// The record is read back by anything that can read the state directory, so
// what it holds about the boot is worth being deliberate about: an opaque id,
// and nothing that authorized the operation.
func TestTheRecordKeepsTheBootItWasToldToLeave(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-1"
	m, dir := newManager(t, c)

	id := createOp(t, m, TypeReboot, "client-1").ID
	awaitTerminal(t, m, id)

	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ops, err := store.List()
	if err != nil || len(ops) != 1 {
		t.Fatalf("List: %v (%d records)", err, len(ops))
	}
	if ops[0].HostBootID != "host-boot-1" {
		t.Errorf("recorded boot = %q, want the one the machine was on", ops[0].HostBootID)
	}
}

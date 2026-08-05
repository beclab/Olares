package clusterop

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// command_issued is where a cluster power operation ends, and the machine goes
// down a moment after it. In that gap the control node is still answering and
// its own state machine still says running — so a phase read from it alone
// bounces the page back to "running" right after the user asked for a reboot,
// which reads as the reboot having been forgotten.
func TestThePhaseSurvivesTheCommandBeingIssued(t *testing.T) {
	for _, tc := range []struct {
		opType Type
		phase  nodestatus.Phase
	}{
		{TypeReboot, nodestatus.PhaseRestarting},
		{TypeShutdown, nodestatus.PhaseShuttingDown},
	} {
		t.Run(string(tc.opType), func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"))
			m, _ := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, tc.opType, "client-1").ID)
			if op.Status != StatusCommandIssued {
				t.Fatalf("status = %q, want command_issued", op.Status)
			}

			phase, ok := m.ActivePhase()
			if !ok || phase != tc.phase {
				t.Errorf("phase = %q ok = %v, want %q while the machine is going down", phase, ok, tc.phase)
			}
		})
	}
}

// Holding the phase must not hold the lock. The operation is over as far as the
// cluster is concerned, and a machine that did not go down has to be reachable
// for another attempt.
func TestHoldingThePhaseDoesNotBlockTheNextOperation(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)
	if _, ok := m.ActivePhase(); !ok {
		t.Fatal("the phase was dropped as soon as the command was issued")
	}

	next, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "client-2", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("an operation that only holds the phase blocked the next one: %v", err)
	}
	awaitTerminal(t, m, next.ID)
}

// A machine still answering long after it was told to switch off did not switch
// off. Saying it is shutting down forever is the one answer that never becomes
// true, so the claim lapses and the node's own state is believed again.
func TestThePhaseLapsesOnAMachineThatNeverWentDown(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	deps := c.deps(t, dir)
	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)
	if _, ok := m.ActivePhase(); !ok {
		t.Fatal("the phase was dropped immediately")
	}

	// Long enough that a machine which was going to go down has done so.
	if err := deps.Sleep(context.Background(), 2*deps.Timeouts.Down); err != nil {
		t.Fatalf("sleep: %v", err)
	}

	if phase, ok := m.ActivePhase(); ok {
		t.Errorf("phase = %q, want the claim to lapse on a machine that is still answering", phase)
	}
}

// The daemon that comes up after the machine really did go down must not
// resurrect the phase out of the record. The machine is demonstrably on: a
// shutdown is never promoted to succeeded, so the record it left behind says
// command_issued forever, and reading a phase off it would pin the cluster to
// shutting_down for the rest of its life.
func TestARestartedDaemonDoesNotResurrectTheIssuedPhase(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)
	if op.Status != StatusCommandIssued {
		t.Fatalf("status = %q", op.Status)
	}

	restarted := newManagerAt(t, c, dir)

	if phase, ok := restarted.ActivePhase(); ok {
		t.Errorf("phase = %q, want none: the machine answering is proof it is on", phase)
	}
}

// A reboot this daemon can prove happened is over, and the phase goes with it.
func TestAConfirmedRebootStopsImposingAPhase(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)

	awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	// The machine came back on another boot, which is the only proof there is.
	c.hostBootID = "host-boot-2"
	restarted := newManagerAt(t, c, dir)

	if phase, ok := restarted.ActivePhase(); ok {
		t.Errorf("phase = %q, want none after the reboot was confirmed", phase)
	}
}

// An operation still in flight decides the phase. The hold is for the gap after
// one ends, and must not outrank the next one.
func TestAnOperationInFlightOutranksTheHeldPhase(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	release := make(chan struct{})
	m, _ := newManager(t, c)

	awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	c.mu.Lock()
	c.onPowerSelf = func() { <-release }
	c.mu.Unlock()

	next, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "client-2", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		phase, ok := m.ActivePhase()
		if ok && phase == nodestatus.PhaseRestarting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("phase = %q ok = %v, want the reboot in flight to decide", phase, ok)
		}
		time.Sleep(2 * time.Millisecond)
	}

	close(release)
	awaitTerminal(t, m, next.ID)
}

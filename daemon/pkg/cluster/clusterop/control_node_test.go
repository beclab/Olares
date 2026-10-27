package clusterop

import (
	"testing"
)

// indexOf is where an event appears in the run's log, or -1.
func indexOf(events []string, want string) int {
	for i, e := range events {
		if e == want {
			return i
		}
	}
	return -1
}

// The control node goes last, but whether it can go at all has to be settled
// first. Discovering that olaresd runs in a container here, after every
// compute node has already been shut down, leaves a cluster with no workers
// and a control node that is still on — and no way to reach the workers to
// bring them back.
func TestPrecheckRefusesAControlNodeThatCannotPowerItself(t *testing.T) {
	for _, opType := range []Type{TypeReboot, TypeShutdown} {
		t.Run(string(opType), func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
			c.localPowerErr = &PowerError{
				Code:    CodePowerUnsupported,
				Message: "olaresd runs in a container on this node, so it cannot power the machine",
			}
			m, _ := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, opType, "client-1").ID)

			if op.Status != StatusFailed || op.Code != CodePrecheckFailed {
				t.Fatalf("status = %q code = %q, want failed/%s", op.Status, op.Code, CodePrecheckFailed)
			}
			if got := nodeResult(t, op, "master-1"); got.Code != CodePowerUnsupported {
				t.Errorf("control node code = %q, want %s", got.Code, CodePowerUnsupported)
			}
			if got := step(t, op, StepPrecheck); got.Status != StepFailed {
				t.Errorf("precheck step = %q, want failed", got.Status)
			}

			dispatch, power := c.counts()
			if dispatch != 0 {
				t.Errorf("%d compute nodes were commanded by an operation the control node cannot finish", dispatch)
			}
			if power != 0 {
				t.Errorf("the control node was powered %d times after refusing the operation", power)
			}
			for _, e := range c.log() {
				if e == "inspect worker-1" {
					t.Error("a compute node was dialled for an operation that was already impossible")
				}
			}
		})
	}
}

// The order is the point: the local answer is cheap and decides the whole run,
// so nothing on the wire happens before it.
func TestPrecheckAsksTheControlNodeBeforeDiallingAnyWorker(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, _ := newManager(t, c)

	awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	events := c.log()
	local := indexOf(events, "local power check shutdown")
	if local < 0 {
		t.Fatalf("the control node was never asked whether it can power itself: %v", events)
	}
	for _, later := range []string{"inspect worker-1", "dispatch worker-1", "power self shutdown"} {
		at := indexOf(events, later)
		if at >= 0 && at < local {
			t.Errorf("%q happened before the control node was asked: %v", later, events)
		}
	}
}

// A control node does not declare power.shutdown: shutting it down on its own
// is not offered, because it is the last step of a cluster operation rather
// than a single-node action. Reading that declaration as "cannot shut down"
// would refuse every cluster shutdown. What decides is the same local check
// the execution point itself performs.
func TestControlNodeCapabilityComesFromItsExecutionPointNotItsDeclaration(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.capabilities["master-1"] = nil
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	if op.Status != StatusCommandIssued {
		t.Fatalf("status = %q code = %q, want %s", op.Status, op.Code, StatusCommandIssued)
	}
	for _, e := range c.log() {
		if e == "inspect master-1" {
			t.Error("the control node was asked to declare a capability it deliberately does not declare")
		}
	}
	if _, power := c.counts(); power != 1 {
		t.Errorf("the control node was powered %d times, want once", power)
	}
}

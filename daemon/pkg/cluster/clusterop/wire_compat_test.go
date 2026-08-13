package clusterop

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// What a caller reads when an operation is refused is part of the contract,
// not an implementation detail: these sentences were written to be read by
// whoever asked for the reboot, and they are what this daemon has always
// said. Each case below is the text the operation record carried before
// power operations became modules, byte for byte, on the operation and on
// the stage that refused it alike.
//
// The other half of the contract is that nothing else reaches the record.
// requireNoDetailOnDisk checks that in the same cases.

func requireNoDetailOnDisk(t *testing.T, dir string, secrets ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", entry.Name(), err)
		}
		for _, secret := range secrets {
			if strings.Contains(string(raw), secret) {
				t.Errorf("%s holds detail that must never be persisted: %q", entry.Name(), secret)
			}
		}
	}
}

func stepResult(t *testing.T, op Operation, name string) Step {
	t.Helper()
	for i := len(op.Steps) - 1; i >= 0; i-- {
		if op.Steps[i].Name == name {
			return op.Steps[i]
		}
	}
	t.Fatalf("operation has no %s step: %+v", name, op.Steps)
	return Step{}
}

// requireRefusal is the whole of the contract for one refusal: the stable
// code, the reviewed sentence, and the stage saying the same thing as the
// operation. A caller that is shown one and logged the other cannot tell
// which is true.
func requireRefusal(t *testing.T, op Operation, step, code, reason string) {
	t.Helper()
	if op.Code != code {
		t.Errorf("op.Code = %q, want %q", op.Code, code)
	}
	if op.Error != reason {
		t.Errorf("op.Error = %q, want %q", op.Error, reason)
	}
	got := stepResult(t, op, step)
	if got.Code != code {
		t.Errorf("%s step code = %q, want %q", step, got.Code, code)
	}
	if got.Error != reason {
		t.Errorf("%s step error = %q, want the same sentence as the operation %q", step, got.Error, reason)
	}
}

func TestNodeShutdownOfTheControlNodeKeepsItsReviewedRefusal(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "client-1", Scope: ScopeNode, Target: "master-1",
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := awaitTerminal(t, m, op.ID)

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	requireRefusal(t, got, StepPrecheck, CodePowerUnsupported,
		"the control node cannot be shut down by a node operation")
}

func TestNodeRebootOfANotReadyWorkerKeepsItsReviewedRefusal(t *testing.T) {
	down := inventory.Node{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2"}
	c := newCluster(master("master-1", "10.0.0.1"), down)
	m, _ := newManager(t, c)

	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "client-1", Scope: ScopeNode, Target: "worker-1",
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := awaitTerminal(t, m, op.ID)

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	requireRefusal(t, got, StepPrecheck, CodeNodeNotReady,
		"the cluster does not consider this node ready")
}

// The control node is checked first and refuses on behalf of the whole
// operation, so the sentence the operation carries is about the operation —
// not the per-node sentence, which says what the control node itself said.
func TestClusterRebootRefusedByTheControlNodeKeepsItsReviewedRefusal(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.localPowerErr = &PowerError{
		Code:    CodePowerUnsupported,
		Message: "olaresd runs in a container on this node, so it cannot power the machine",
		Err:     errors.New("token=super-secret container detail"),
	}
	m, dir := newManager(t, c)

	got := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	requireRefusal(t, got, StepPrecheck, CodePrecheckFailed,
		"the control node cannot perform this operation")
	master := nodeResult(t, got, "master-1")
	if master.Code != CodePowerUnsupported ||
		master.Error != "olaresd runs in a container on this node, so it cannot power the machine" {
		t.Errorf("control node = %q/%q, want the execution point's own refusal", master.Code, master.Error)
	}
	requireNoDetailOnDisk(t, dir, "super-secret")
}

// The message comes from the point that would have powered the machine. It
// already separates what a caller may read from the detail behind it, and
// the record keeps the first and never the second.
func TestHostPowerFailureKeepsTheExecutionPointsOwnMessage(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.powerSelfErr = &PowerError{
		Code:    CodePowerUnsupported,
		Message: "this node does not have the command this operation needs",
		Err:     errors.New("exec: \"reboot\": token=super-secret not found in $PATH"),
	}
	m, dir := newManager(t, c)

	got := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", got.Status)
	}
	requireRefusal(t, got, StepMasterCommand, CodePowerUnsupported,
		"this node does not have the command this operation needs")
	requireNoDetailOnDisk(t, dir, "super-secret")
}

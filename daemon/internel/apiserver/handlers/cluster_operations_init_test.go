package handlers

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// harmlessDeps is a complete set of side effects that reach nothing: no
// cluster, no network, and above all no power command. Setup takes the seams
// as an argument precisely so a test can say this.
func harmlessDeps(t *testing.T) clusterop.Deps {
	t.Helper()
	return clusterop.Deps{
		Inventory: func(context.Context) ([]inventory.Node, error) { return nil, nil },
		Inspect: func(context.Context, inventory.Node, clusterop.Credentials) (nodestatus.Status, error) {
			return nodestatus.Status{}, nil
		},
		Dispatch: func(context.Context, []inventory.Node, clusterop.PeerRequest, clusterop.Credentials) []clusterop.DispatchOutcome {
			return nil
		},
		Observe: func(context.Context) (map[string]inventory.Observation, error) { return nil, nil },
		PowerSelf: func(context.Context, clusterop.Type) error {
			t.Error("a test powered the machine it is running on")
			return nil
		},
		LocalPowerSupport: func(clusterop.Type) error { return nil },
		HostBootID:        func() (string, error) { return "test-boot", nil },
	}
}

// The daemon restarts more often than a cluster is powered off, so an
// operation recorded before the restart has to be there afterwards.
func TestInitClusterOperationsReadsWhatWasRecorded(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cluster-operations")
	store, err := clusterop.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	op := sampleOperation()
	op.Status = clusterop.StatusCommandIssued
	if err := store.Save(op); err != nil {
		t.Fatalf("Save: %v", err)
	}

	prev := clusterOperations
	t.Cleanup(func() { clusterOperations = prev })
	if err := InitClusterOperations(dir, harmlessDeps(t)); err != nil {
		t.Fatalf("InitClusterOperations: %v", err)
	}

	asAuthorizedUser(t)
	asMaster(t)
	resp, body := callRegistered(t, "/cluster/operations/"+op.ID, authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want the recorded operation back: %s", resp.StatusCode, body)
	}
	if got := decodeOperation(t, body); got.ID != op.ID || got.Status != clusterop.StatusCommandIssued {
		t.Errorf("operation = %+v", got)
	}
}

// What this daemon can be asked to do is settled when it starts serving. A
// module registered after that would be reachable by name from an HTTP route
// while the orchestrator that was already built, the signature reader and
// every node's own set disagree about whether it exists, so the set is
// closed before anything can be asked for.
func TestInitClusterOperationsClosesTheModuleSet(t *testing.T) {
	prev := clusterOperations
	t.Cleanup(func() { clusterOperations = prev })

	if err := InitClusterOperations(t.TempDir(), harmlessDeps(t)); err != nil {
		t.Fatalf("InitClusterOperations: %v", err)
	}

	late := &nodeModuleRecorder{typ: clusterop.Type("late-op")}
	if err := clusterop.DefaultRegistry().Register(late); err == nil {
		t.Fatal("an operation registered itself after the daemon started serving")
	}
	if _, err := clusterop.ParseType("late-op"); err == nil {
		t.Error("a route would accept an operation the orchestrator was never built with")
	}
}

// Closing the set is not a one-shot that a second startup path would trip
// over: a daemon may open its records more than once, and doing so must not
// turn a working setup into a failed one.
func TestInitClusterOperationsCanBeCalledAgain(t *testing.T) {
	prev := clusterOperations
	t.Cleanup(func() { clusterOperations = prev })

	for i := 0; i < 2; i++ {
		if err := InitClusterOperations(t.TempDir(), harmlessDeps(t)); err != nil {
			t.Fatalf("InitClusterOperations call %d: %v", i+1, err)
		}
		if clusterOperations == nil {
			t.Fatalf("call %d published no orchestrator", i+1)
		}
	}
	if err := clusterop.DefaultRegistry().Register(
		&nodeModuleRecorder{typ: clusterop.Type("late-op-again")}); err == nil {
		t.Error("the module set reopened")
	}
}

// The orchestrator and the node routes answer from one set, not two. The
// master decides an operation exists, dispatches its node half, and reads the
// owner's signature against modules; a node carries that half out against
// its own. Two sets could disagree, and the disagreement would surface as a
// node refusing work the record already says was dispatched.
func TestTheNodeRoutesAndTheOrchestratorShareOneClosedModuleSet(t *testing.T) {
	prevOps := clusterOperations
	prevNode := nodeOperations
	t.Cleanup(func() {
		clusterOperations = prevOps
		nodeOperations = prevNode
	})

	// Exactly what main wires, in the order main wires it.
	InstallNodeOperations(clusterop.DefaultRegistry())
	if err := InitClusterOperations(t.TempDir(), harmlessDeps(t)); err != nil {
		t.Fatalf("InitClusterOperations: %v", err)
	}

	if nodeOperations != clusterop.DefaultRegistry() {
		t.Fatal("the node routes answer from a different module set than the orchestrator")
	}
	if err := nodeOperations.Register(
		&nodeModuleRecorder{typ: clusterop.Type("late-node-op")}); err == nil {
		t.Error("the set the node routes act through is still open")
	}
}

func TestInitClusterOperationsRefusesAnUnusableSetup(t *testing.T) {
	prev := clusterOperations
	t.Cleanup(func() { clusterOperations = prev })
	clusterOperations = nil

	if err := InitClusterOperations(t.TempDir(), clusterop.Deps{}); err == nil {
		t.Fatal("a set of side effects with nothing in it was accepted")
	}
	if clusterOperations != nil {
		t.Error("a failed setup still published an orchestrator")
	}
}

func TestNewStoreRefusesAnEmptyDirectory(t *testing.T) {
	if _, err := clusterop.NewStore(""); err == nil {
		t.Fatal("an empty state directory was accepted")
	}
}

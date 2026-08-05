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

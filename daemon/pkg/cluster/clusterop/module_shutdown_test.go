package clusterop

import (
	"context"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

func shutdownModuleUnderTest(t *testing.T) OperationModule {
	t.Helper()
	module, ok := DefaultRegistry().Lookup(TypeShutdown)
	if !ok {
		t.Fatal("shutdown module is not registered")
	}
	return module
}

func TestShutdownModuleContract(t *testing.T) {
	module := shutdownModuleUnderTest(t)
	if module.Type() != TypeShutdown {
		t.Fatalf("Type() = %q, want %q", module.Type(), TypeShutdown)
	}
	phase, ok := module.Phase(Operation{Type: TypeShutdown, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseShuttingDown {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}

func TestShutdownModulePhaseOutlivesTheCommandBeingIssued(t *testing.T) {
	module := shutdownModuleUnderTest(t)
	phase, ok := module.Phase(Operation{Type: TypeShutdown, Status: StatusCommandIssued})
	if !ok || phase != nodestatus.PhaseShuttingDown {
		t.Fatalf("Phase() = %q, %v, want shutting_down while the machine is going down", phase, ok)
	}
}

func TestShutdownModuleAcceptsOnlyTheScopesItCanCarryOut(t *testing.T) {
	module := shutdownModuleUnderTest(t)
	for _, scope := range []string{"", ScopeCluster, ScopeNode} {
		req := CreateRequest{
			Type: TypeShutdown, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: scope, Target: "worker-1", ClusterID: "cluster-1",
		}
		if err := module.Validate(req); err != nil {
			t.Errorf("Validate(scope %q) = %v, want it accepted", scope, err)
		}
	}
	if err := module.Validate(CreateRequest{
		Type: TypeShutdown, RequestID: "client-1", Owner: "alice@olares.com", Scope: "everything",
	}); err == nil {
		t.Error("a scope this daemon has never carried out was accepted")
	}
}

// A machine that is on again is somebody pressing the power button, not a
// shutdown succeeding. The module deliberately offers no recovery, so a
// daemon that comes back leaves the command_issued record as it found it.
func TestShutdownModuleNeverConfirmsAShutdownNothingObserved(t *testing.T) {
	if _, ok := shutdownModuleUnderTest(t).(RecoverableModule); ok {
		t.Fatal("the shutdown module claims it can confirm a shutdown after the fact")
	}
}

func TestShutdownModuleShutsThisNodeDownForANodeRequest(t *testing.T) {
	node, ok := shutdownModuleUnderTest(t).(NodeOperationModule)
	if !ok {
		t.Fatal("the shutdown module cannot carry out a node-local shutdown")
	}
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
	withPowerCommands(t, rb, sd, nil)

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeShutdown, OperationID: "op-1", RequestID: "client-1"},
	})
	if err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if !sd.executed || rb.executed {
		t.Errorf("shutdown=%v reboot=%v, want only the shutdown", sd.executed, rb.executed)
	}
}

func TestShutdownModuleRefusesToRunOutsideTheManagersRuntime(t *testing.T) {
	outcome := shutdownModuleUnderTest(t).Run(context.Background(), stubRuntime{}, RunRequest{})
	if outcome.Status != StatusFailed {
		t.Fatalf("Run() on a foreign runtime = %+v, want it refused", outcome)
	}
}

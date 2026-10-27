package clusterop

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// stubRuntime is a Runtime that is not the manager's own. A power module
// needs the injected side effects only the manager's runtime carries, so
// handing it one of these must end the operation rather than have it
// improvise its way to a machine.
type stubRuntime struct{}

func (stubRuntime) Operation() (Operation, bool)                        { return Operation{}, true }
func (stubRuntime) CanContinue() bool                                   { return true }
func (stubRuntime) StartStep(string) error                              { return nil }
func (stubRuntime) FinishStep(string, StepStatus, string, string) error { return nil }
func (stubRuntime) InitNodes([]NodeResult) error                        { return nil }
func (stubRuntime) UpdateNode(string, func(*NodeResult)) error          { return nil }
func (stubRuntime) SetHostBootID(string) error                          { return nil }
func (stubRuntime) SetModuleState(json.RawMessage) error                { return nil }
func (stubRuntime) SetCommandIssuedUntil(time.Time) error               { return nil }
func (stubRuntime) Complete(Outcome) error                              { return nil }
func (stubRuntime) Now() time.Time                                      { return time.Now() }
func (stubRuntime) Context() context.Context                            { return context.Background() }

func rebootModuleUnderTest(t *testing.T) OperationModule {
	t.Helper()
	module, ok := DefaultRegistry().Lookup(TypeReboot)
	if !ok {
		t.Fatal("reboot module is not registered")
	}
	return module
}

func TestRebootModuleContract(t *testing.T) {
	module := rebootModuleUnderTest(t)
	if module.Type() != TypeReboot {
		t.Fatalf("Type() = %q, want %q", module.Type(), TypeReboot)
	}
	phase, ok := module.Phase(Operation{Type: TypeReboot, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseRestarting {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}

// command_issued is where a reboot that reached the control node ends, and
// the machine is restarting for minutes after it. The phase the module
// reports is what the cluster is doing, not how far the record has got.
func TestRebootModulePhaseOutlivesTheCommandBeingIssued(t *testing.T) {
	module := rebootModuleUnderTest(t)
	phase, ok := module.Phase(Operation{Type: TypeReboot, Status: StatusCommandIssued})
	if !ok || phase != nodestatus.PhaseRestarting {
		t.Fatalf("Phase() = %q, %v, want restarting while the machine is going down", phase, ok)
	}
}

// The scopes are the ones the daemon has always carried out. An empty scope
// is one of them: records written before scope existed carry none, and the
// run path has always read anything that is not "node" as covering the whole
// cluster.
func TestRebootModuleAcceptsOnlyTheScopesItCanCarryOut(t *testing.T) {
	module := rebootModuleUnderTest(t)
	for _, tc := range []struct{ scope, target string }{
		{"", ""},
		{ScopeCluster, ""},
		{ScopeNode, "master-1"},
	} {
		req := CreateRequest{
			Type: TypeReboot, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1",
		}
		if err := module.Validate(req); err != nil {
			t.Errorf("Validate(scope %q) = %v, want it accepted", tc.scope, err)
		}
	}
	if err := module.Validate(CreateRequest{
		Type: TypeReboot, RequestID: "client-1", Owner: "alice@olares.com", Scope: "everything",
	}); err == nil {
		t.Error("a scope this daemon has never carried out was accepted")
	}
}

// Which nodes a scope may name is the module's own question, asked before
// anything is recorded: the route checks that the owner signed this exact
// scope and target, not whether the two describe an operation this module
// could carry out.
func TestRebootModuleRefusesAScopeAndTargetThatDisagree(t *testing.T) {
	module := rebootModuleUnderTest(t)
	for _, tc := range []struct{ name, scope, target string }{
		{"a whole-cluster reboot naming one node", ScopeCluster, "master-1"},
		{"a single-node reboot naming none", ScopeNode, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := module.Validate(CreateRequest{
				Type: TypeReboot, RequestID: "client-1", Owner: "alice@olares.com",
				Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1",
			}); err == nil {
				t.Error("a request whose scope and target describe different operations was accepted")
			}
		})
	}
}

// A control node's own reboot ends at command_issued because the process
// that would confirm it goes down with the machine. Confirming it afterwards
// is the module's job, so the module has to offer one.
func TestRebootModuleCanConfirmACommandItIssued(t *testing.T) {
	if _, ok := rebootModuleUnderTest(t).(RecoverableModule); !ok {
		t.Fatal("the reboot module cannot confirm a control-node reboot after a restart")
	}
}

// A node-local reboot goes through the same execution point, and the same
// state check, as the single-node power endpoint. A cluster operation is a
// way to sequence those, not a way around them.
func TestRebootModuleRebootsThisNodeForANodeRequest(t *testing.T) {
	node, ok := rebootModuleUnderTest(t).(NodeOperationModule)
	if !ok {
		t.Fatal("the reboot module cannot carry out a node-local reboot")
	}
	rb := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}
	sd := &fakeCommand{Operation: commands.Operation{Name: commands.Shutdown}}
	withPowerCommands(t, rb, sd, nil)

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeReboot, OperationID: "op-1", RequestID: "client-1"},
	})
	if err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if !rb.executed || sd.executed {
		t.Errorf("reboot=%v shutdown=%v, want only the reboot", rb.executed, sd.executed)
	}
}

func TestRebootModuleRefusesToRunOutsideTheManagersRuntime(t *testing.T) {
	outcome := rebootModuleUnderTest(t).Run(context.Background(), stubRuntime{}, RunRequest{})
	if outcome.Status != StatusFailed {
		t.Fatalf("Run() on a foreign runtime = %+v, want it refused", outcome)
	}
}

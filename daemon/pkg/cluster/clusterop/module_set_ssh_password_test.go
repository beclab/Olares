package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/commands"
	sshpassword "github.com/beclab/Olares/daemon/pkg/commands/ssh_password"
)

func setSSHPasswordModuleUnderTest(t *testing.T) OperationModule {
	t.Helper()
	module, ok := DefaultRegistry().Lookup(TypeSetSSHPassword)
	if !ok {
		t.Fatal("set-ssh-password module is not registered")
	}
	return module
}

func TestSetSSHPasswordModuleContract(t *testing.T) {
	module := setSSHPasswordModuleUnderTest(t)
	if module.Type() != TypeSetSSHPassword {
		t.Fatalf("Type() = %q, want %q", module.Type(), TypeSetSSHPassword)
	}
	phase, ok := module.Phase(Operation{Type: TypeSetSSHPassword, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseMaintenance {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}

func TestSetSSHPasswordModuleAcceptsOnlyTheScopesItCanCarryOut(t *testing.T) {
	module := setSSHPasswordModuleUnderTest(t)
	params := json.RawMessage(`{"password":"secret"}`)
	for _, tc := range []struct{ scope, target string }{
		{"", ""},
		{ScopeCluster, ""},
		{ScopeNode, "master-1"},
	} {
		req := CreateRequest{
			Type: TypeSetSSHPassword, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1", Params: params,
		}
		if err := module.Validate(req); err != nil {
			t.Errorf("Validate(scope %q) = %v, want it accepted", tc.scope, err)
		}
	}
	if err := module.Validate(CreateRequest{
		Type: TypeSetSSHPassword, RequestID: "client-1", Owner: "alice@olares.com",
		Scope: "everything", Params: params,
	}); err == nil {
		t.Error("a scope this daemon has never carried out was accepted")
	}
}

func TestSetSSHPasswordModuleRefusesAScopeAndTargetThatDisagree(t *testing.T) {
	module := setSSHPasswordModuleUnderTest(t)
	params := json.RawMessage(`{"password":"secret"}`)
	for _, tc := range []struct{ name, scope, target string }{
		{"a whole-cluster set naming one node", ScopeCluster, "master-1"},
		{"a single-node set naming none", ScopeNode, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := module.Validate(CreateRequest{
				Type: TypeSetSSHPassword, RequestID: "client-1", Owner: "alice@olares.com",
				Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1", Params: params,
			}); err == nil {
				t.Error("a request whose scope and target describe different operations was accepted")
			}
		})
	}
}

func TestSetSSHPasswordModuleRequiresAPassword(t *testing.T) {
	module := setSSHPasswordModuleUnderTest(t)
	for _, params := range []string{``, `{}`, `{"username":"olares"}`, `{"password":""}`, `{"password":"  "}`} {
		if err := module.Validate(CreateRequest{
			Type: TypeSetSSHPassword, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: ScopeCluster, ClusterID: "cluster-1", Params: json.RawMessage(params),
		}); err == nil {
			t.Errorf("Validate(params=%s) = nil, want a refusal", params)
		}
	}
}

func TestSetSSHPasswordModuleNeverConfirmsAnythingAfterARestart(t *testing.T) {
	if _, ok := setSSHPasswordModuleUnderTest(t).(RecoverableModule); ok {
		t.Fatal("set-ssh-password finishes synchronously and has nothing to confirm after a restart")
	}
}

func TestSetSSHPasswordModuleSetsThisNodeForANodeRequest(t *testing.T) {
	node, ok := setSSHPasswordModuleUnderTest(t).(NodeOperationModule)
	if !ok {
		t.Fatal("the set-ssh-password module cannot carry out a node-local set")
	}
	cmd := &fakeSetSSHPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withSetSSHPasswordCommand(t, cmd, nil)

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeSetSSHPassword, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"username":"bob","password":"new-secret"}`),
	})
	if err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if !cmd.executed {
		t.Fatal("the ssh password command was not executed")
	}
	if cmd.username != "bob" || cmd.password != "new-secret" {
		t.Errorf("executed with user=%q password=%q, want bob/new-secret", cmd.username, cmd.password)
	}
}

func TestSetSSHPasswordModuleDefaultsTheUsernameOnThisNode(t *testing.T) {
	node := setSSHPasswordModuleUnderTest(t).(NodeOperationModule)
	cmd := &fakeSetSSHPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withSetSSHPasswordCommand(t, cmd, nil)

	if err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeSetSSHPassword, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"password":"new-secret"}`),
	}); err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if cmd.username != "olares" {
		t.Errorf("username = %q, want the same default the single-node endpoint uses", cmd.username)
	}
}

func TestSetSSHPasswordModuleHonoursTheDaemonStateOnThisNode(t *testing.T) {
	node := setSSHPasswordModuleUnderTest(t).(NodeOperationModule)
	cmd := &fakeSetSSHPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withSetSSHPasswordCommand(t, cmd, errors.New("operation is not allowed while installing"))

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeSetSSHPassword, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"password":"new-secret"}`),
	})
	if err == nil {
		t.Fatal("ExecuteNode = nil, want the state check to refuse it")
	}
	if cmd.executed {
		t.Error("the password was set after the state check refused it")
	}
}

func TestSetSSHPasswordModuleRefusesToRunOutsideTheManagersRuntime(t *testing.T) {
	outcome := setSSHPasswordModuleUnderTest(t).Run(context.Background(), stubRuntime{}, RunRequest{
		Params: json.RawMessage(`{"password":"secret"}`),
	})
	if outcome.Status != StatusFailed {
		t.Fatalf("Run() on a foreign runtime = %+v, want it refused", outcome)
	}
}

func TestSetSSHPasswordModuleSetsTheClusterThroughTheGenericFanOut(t *testing.T) {
	cmd := &fakeSetSSHPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withSetSSHPasswordCommand(t, cmd, nil)

	var dispatched []string
	prev := dispatchSetSSHPassword
	dispatchSetSSHPassword = func(_ context.Context, nodes []inventory.Node, req NodeRequest,
		_ Credentials) []DispatchOutcome {
		out := make([]DispatchOutcome, 0, len(nodes))
		for _, n := range nodes {
			dispatched = append(dispatched, n.NodeName)
			if string(req.Params) == "" || !strings.Contains(string(req.Params), "secret") {
				t.Errorf("worker params = %s, want the caller's password carried across", req.Params)
			}
			if req.Type != TypeSetSSHPassword {
				t.Errorf("worker type = %q, want %q", req.Type, TypeSetSSHPassword)
			}
			out = append(out, DispatchOutcome{NodeName: n.NodeName})
		}
		return out
	}
	t.Cleanup(func() { dispatchSetSSHPassword = prev })

	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeSetSSHPassword,
		RequestID: "client-1",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(`{"password":"secret"}`),
		Creds:     Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got := awaitTerminal(t, m, op.ID)
	if got.Status != StatusSucceeded {
		t.Fatalf("status = %q code = %q, want succeeded", got.Status, got.Code)
	}
	if !cmd.executed {
		t.Fatal("the control node did not set its own ssh password")
	}
	if len(dispatched) != 1 || dispatched[0] != "worker-1" {
		t.Fatalf("dispatched = %v, want worker-1 through the generic fan-out", dispatched)
	}
	if node := nodeResult(t, got, "worker-1"); node.Status != nodeSetSSHPasswordSucceeded {
		t.Errorf("worker status = %q, want %q", node.Status, nodeSetSSHPasswordSucceeded)
	}
	if node := nodeResult(t, got, "master-1"); node.Status != nodeSetSSHPasswordSucceeded {
		t.Errorf("master status = %q, want %q", node.Status, nodeSetSSHPasswordSucceeded)
	}
}

func TestSetSSHPasswordModuleParamsAreNeverWrittenDown(t *testing.T) {
	cmd := &fakeSetSSHPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withSetSSHPasswordCommand(t, cmd, nil)
	prev := dispatchSetSSHPassword
	dispatchSetSSHPassword = func(_ context.Context, nodes []inventory.Node, _ NodeRequest,
		_ Credentials) []DispatchOutcome {
		out := make([]DispatchOutcome, 0, len(nodes))
		for _, n := range nodes {
			out = append(out, DispatchOutcome{NodeName: n.NodeName})
		}
		return out
	}
	t.Cleanup(func() { dispatchSetSSHPassword = prev })

	const secret = "super-secret-password-value"
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeSetSSHPassword,
		RequestID: "client-1",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(`{"password":"` + secret + `"}`),
		Creds:     Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	awaitTerminal(t, m, op.ID)
	requireNoCredentialsOnDisk(t, dir, secret)
}

type fakeSetSSHPasswordCommand struct {
	commands.Operation
	executed bool
	username string
	password string
	err      error
}

func (f *fakeSetSSHPasswordCommand) Execute(_ context.Context, p any) (any, error) {
	f.executed = true
	if param, ok := p.(*sshpassword.Param); ok {
		f.username = param.Username
		f.password = param.Password
	}
	return nil, f.err
}

func (f *fakeSetSSHPasswordCommand) OperationName() commands.Operations { return f.Operation.Name }

func withSetSSHPasswordCommand(t *testing.T, cmd *fakeSetSSHPasswordCommand, validate error) {
	t.Helper()
	prevNew := newSetSSHPasswordCommand
	prevValidate := validateSetSSHPasswordOp
	newSetSSHPasswordCommand = func() commands.Interface { return cmd }
	validateSetSSHPasswordOp = func(commands.Interface) error { return validate }
	t.Cleanup(func() {
		newSetSSHPasswordCommand = prevNew
		validateSetSSHPasswordOp = prevValidate
	})
}

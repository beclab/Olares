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

func resetPasswordModuleUnderTest(t *testing.T) OperationModule {
	t.Helper()
	module, ok := DefaultRegistry().Lookup(TypeResetPassword)
	if !ok {
		t.Fatal("reset-password module is not registered")
	}
	return module
}

func TestResetPasswordModuleContract(t *testing.T) {
	module := resetPasswordModuleUnderTest(t)
	if module.Type() != TypeResetPassword {
		t.Fatalf("Type() = %q, want %q", module.Type(), TypeResetPassword)
	}
	phase, ok := module.Phase(Operation{Type: TypeResetPassword, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseMaintenance {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}

func TestResetPasswordModuleAcceptsOnlyTheScopesItCanCarryOut(t *testing.T) {
	module := resetPasswordModuleUnderTest(t)
	params := json.RawMessage(`{"password":"secret"}`)
	for _, tc := range []struct{ scope, target string }{
		{"", ""},
		{ScopeCluster, ""},
		{ScopeNode, "master-1"},
	} {
		req := CreateRequest{
			Type: TypeResetPassword, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1", Params: params,
		}
		if err := module.Validate(req); err != nil {
			t.Errorf("Validate(scope %q) = %v, want it accepted", tc.scope, err)
		}
	}
	if err := module.Validate(CreateRequest{
		Type: TypeResetPassword, RequestID: "client-1", Owner: "alice@olares.com",
		Scope: "everything", Params: params,
	}); err == nil {
		t.Error("a scope this daemon has never carried out was accepted")
	}
}

func TestResetPasswordModuleRefusesAScopeAndTargetThatDisagree(t *testing.T) {
	module := resetPasswordModuleUnderTest(t)
	params := json.RawMessage(`{"password":"secret"}`)
	for _, tc := range []struct{ name, scope, target string }{
		{"a whole-cluster reset naming one node", ScopeCluster, "master-1"},
		{"a single-node reset naming none", ScopeNode, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := module.Validate(CreateRequest{
				Type: TypeResetPassword, RequestID: "client-1", Owner: "alice@olares.com",
				Scope: tc.scope, Target: tc.target, ClusterID: "cluster-1", Params: params,
			}); err == nil {
				t.Error("a request whose scope and target describe different operations was accepted")
			}
		})
	}
}

func TestResetPasswordModuleRequiresAPassword(t *testing.T) {
	module := resetPasswordModuleUnderTest(t)
	for _, params := range []string{``, `{}`, `{"username":"olares"}`, `{"password":""}`, `{"password":"  "}`} {
		if err := module.Validate(CreateRequest{
			Type: TypeResetPassword, RequestID: "client-1", Owner: "alice@olares.com",
			Scope: ScopeCluster, ClusterID: "cluster-1", Params: json.RawMessage(params),
		}); err == nil {
			t.Errorf("Validate(params=%s) = nil, want a refusal", params)
		}
	}
}

func TestResetPasswordModuleNeverConfirmsAnythingAfterARestart(t *testing.T) {
	if _, ok := resetPasswordModuleUnderTest(t).(RecoverableModule); ok {
		t.Fatal("reset-password finishes synchronously and has nothing to confirm after a restart")
	}
}

func TestResetPasswordModuleResetsThisNodeForANodeRequest(t *testing.T) {
	node, ok := resetPasswordModuleUnderTest(t).(NodeOperationModule)
	if !ok {
		t.Fatal("the reset-password module cannot carry out a node-local reset")
	}
	cmd := &fakeResetPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withResetPasswordCommand(t, cmd, nil)

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeResetPassword, OperationID: "op-1", RequestID: "client-1"},
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

func TestResetPasswordModuleDefaultsTheUsernameOnThisNode(t *testing.T) {
	node := resetPasswordModuleUnderTest(t).(NodeOperationModule)
	cmd := &fakeResetPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withResetPasswordCommand(t, cmd, nil)

	if err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeResetPassword, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"password":"new-secret"}`),
	}); err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if cmd.username != "olares" {
		t.Errorf("username = %q, want the same default the single-node endpoint uses", cmd.username)
	}
}

func TestResetPasswordModuleHonoursTheDaemonStateOnThisNode(t *testing.T) {
	node := resetPasswordModuleUnderTest(t).(NodeOperationModule)
	cmd := &fakeResetPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withResetPasswordCommand(t, cmd, errors.New("operation is not allowed while installing"))

	err := node.ExecuteNode(context.Background(), NodeRequest{
		PeerRequest: PeerRequest{Type: TypeResetPassword, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"password":"new-secret"}`),
	})
	if err == nil {
		t.Fatal("ExecuteNode = nil, want the state check to refuse it")
	}
	if cmd.executed {
		t.Error("the password was set after the state check refused it")
	}
}

func TestResetPasswordModuleRefusesToRunOutsideTheManagersRuntime(t *testing.T) {
	outcome := resetPasswordModuleUnderTest(t).Run(context.Background(), stubRuntime{}, RunRequest{
		Params: json.RawMessage(`{"password":"secret"}`),
	})
	if outcome.Status != StatusFailed {
		t.Fatalf("Run() on a foreign runtime = %+v, want it refused", outcome)
	}
}

func TestResetPasswordModuleResetsTheClusterThroughTheGenericFanOut(t *testing.T) {
	cmd := &fakeResetPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withResetPasswordCommand(t, cmd, nil)

	var dispatched []string
	prev := dispatchResetPassword
	dispatchResetPassword = func(_ context.Context, nodes []inventory.Node, req NodeRequest,
		_ Credentials) []DispatchOutcome {
		out := make([]DispatchOutcome, 0, len(nodes))
		for _, n := range nodes {
			dispatched = append(dispatched, n.NodeName)
			if string(req.Params) == "" || !strings.Contains(string(req.Params), "secret") {
				t.Errorf("worker params = %s, want the caller's password carried across", req.Params)
			}
			if req.Type != TypeResetPassword {
				t.Errorf("worker type = %q, want %q", req.Type, TypeResetPassword)
			}
			out = append(out, DispatchOutcome{NodeName: n.NodeName})
		}
		return out
	}
	t.Cleanup(func() { dispatchResetPassword = prev })

	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeResetPassword,
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
		t.Fatal("the control node did not reset its own password")
	}
	if len(dispatched) != 1 || dispatched[0] != "worker-1" {
		t.Fatalf("dispatched = %v, want worker-1 through the generic fan-out", dispatched)
	}
	if node := nodeResult(t, got, "worker-1"); node.Status != nodeResetSucceeded {
		t.Errorf("worker status = %q, want %q", node.Status, nodeResetSucceeded)
	}
	if node := nodeResult(t, got, "master-1"); node.Status != nodeResetSucceeded {
		t.Errorf("master status = %q, want %q", node.Status, nodeResetSucceeded)
	}
}

func TestResetPasswordModuleParamsAreNeverWrittenDown(t *testing.T) {
	cmd := &fakeResetPasswordCommand{Operation: commands.Operation{Name: commands.SetSSHPassword}}
	withResetPasswordCommand(t, cmd, nil)
	prev := dispatchResetPassword
	dispatchResetPassword = func(_ context.Context, nodes []inventory.Node, _ NodeRequest,
		_ Credentials) []DispatchOutcome {
		out := make([]DispatchOutcome, 0, len(nodes))
		for _, n := range nodes {
			out = append(out, DispatchOutcome{NodeName: n.NodeName})
		}
		return out
	}
	t.Cleanup(func() { dispatchResetPassword = prev })

	const secret = "super-secret-password-value"
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeResetPassword,
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

type fakeResetPasswordCommand struct {
	commands.Operation
	executed bool
	username string
	password string
	err      error
}

func (f *fakeResetPasswordCommand) Execute(_ context.Context, p any) (any, error) {
	f.executed = true
	if param, ok := p.(*sshpassword.Param); ok {
		f.username = param.Username
		f.password = param.Password
	}
	return nil, f.err
}

func (f *fakeResetPasswordCommand) OperationName() commands.Operations { return f.Operation.Name }

func withResetPasswordCommand(t *testing.T, cmd *fakeResetPasswordCommand, validate error) {
	t.Helper()
	prevNew := newResetPasswordCommand
	prevValidate := validateResetPasswordOp
	newResetPasswordCommand = func() commands.Interface { return cmd }
	validateResetPasswordOp = func(commands.Interface) error { return validate }
	t.Cleanup(func() {
		newResetPasswordCommand = prevNew
		validateResetPasswordOp = prevValidate
	})
}

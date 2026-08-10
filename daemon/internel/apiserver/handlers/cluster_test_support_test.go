package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

const (
	testToken = "test-access-token"
	testOwner = "alice@olares.com"
)

func authHeaders() map[string]string {
	return map[string]string{AUTH_HEADER: testToken}
}

// asOwnerSignature stands in for the DID service. The fake hands back exactly
// the body the presented signature carried — the test writes that body as
// JSON — so every check the route makes on what the owner actually signed
// still runs.
func asOwnerSignature(t *testing.T) {
	t.Helper()
	prevClient := newTermipassClient
	newTermipassClient = func(_ context.Context, jws string) (ownerClient, error) {
		var body map[string]any
		if err := json.Unmarshal([]byte(jws), &body); err != nil {
			return nil, errors.New("bad signature")
		}
		return signedFakeClient{id: testOwner, body: body}, nil
	}
	prevID := olaresIDFromRelease
	olaresIDFromRelease = func() (string, error) { return testOwner, nil }
	prevClusterID := clusterIDOf
	clusterIDOf = func(context.Context) (string, error) { return "cluster-test", nil }
	t.Cleanup(func() {
		newTermipassClient = prevClient
		olaresIDFromRelease = prevID
		clusterIDOf = prevClusterID
	})
}

type signedFakeClient struct {
	id   string
	body map[string]any
}

func (f signedFakeClient) OlaresID() string { return f.id }
func (f signedFakeClient) SignedBody() any  { return f.body }

var _ client.SignedClient = signedFakeClient{}

// ownerBinding is what TermiPass signs for one cluster power operation, on top
// of whatever else it puts in the body.
func ownerBinding(ty clusterop.Type, requestID string) map[string]any {
	return map[string]any{
		"username":  "alice",
		"clusterId": "cluster-test",
		"type":      string(ty),
		"requestId": requestID,
		"scope":     clusterop.ScopeCluster,
		"expiresAt": time.Now().Add(5 * time.Minute).UnixMilli(),
	}
}

func ownerNodeBinding(ty clusterop.Type, requestID, nodeName string) map[string]any {
	binding := ownerBinding(ty, requestID)
	binding["scope"] = clusterop.ScopeNode
	binding["target"] = nodeName
	return binding
}

func signatureCarrying(t *testing.T, body map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(raw)
}

// signedHeaders carry both credentials: the operation-bound signature the
// dangerous routes require, and the access token every route requires.
func signedHeaders(t *testing.T, body map[string]any) map[string]string {
	t.Helper()
	return map[string]string{
		AUTH_HEADER:      testToken,
		SIGNATURE_HEADER: signatureCarrying(t, body),
	}
}

func signedFor(t *testing.T, ty clusterop.Type, requestID string) map[string]string {
	t.Helper()
	return signedHeaders(t, ownerBinding(ty, requestID))
}

// asAuthorizedUser satisfies the real RequireAuthorization middleware without
// an identity provider. Only token verification is faked; the middleware still
// runs, so removing it from a route still fails the unauthorized tests.
func asAuthorizedUser(t *testing.T) {
	t.Helper()
	prev := validateAccessToken
	validateAccessToken = func(token string) (bool, *utils.ValidToken, error) {
		if token != testToken {
			return false, nil, errors.New("unexpected token")
		}
		return true, &utils.ValidToken{Username: "alice", Groups: []string{utils.Owner}}, nil
	}
	t.Cleanup(func() { validateAccessToken = prev })
}

func asNode(t *testing.T, name string, role inventory.Role, err error) {
	t.Helper()
	prev := thisNodeInCluster
	thisNodeInCluster = func(context.Context) (string, inventory.Role, error) {
		return name, role, err
	}
	t.Cleanup(func() { thisNodeInCluster = prev })
}

func asMaster(t *testing.T) {
	t.Helper()
	asNode(t, "master-1", inventory.RoleMaster, nil)
	withCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning}, time.Now())
}

func asWorker(t *testing.T) {
	t.Helper()
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	withCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning}, time.Now())
}

// nodeModuleRecorder is an operation this daemon does not have. A test
// registers it into a registry of its own and installs that, so a node route
// can be driven end to end without the modules that really power a machine
// being reachable from the test binary at all.
type nodeModuleRecorder struct {
	typ            clusterop.Type
	validateEr     error
	execEr         error
	panics         bool
	panicsValidate bool

	// onExecute stands in for the work, for a test that wants to observe
	// something other than the request itself.
	onExecute func(clusterop.NodeRequest) error

	mu        sync.Mutex
	validated []clusterop.CreateRequest
	executed  []clusterop.NodeRequest
}

func (m *nodeModuleRecorder) Type() clusterop.Type { return m.typ }

func (m *nodeModuleRecorder) Validate(req clusterop.CreateRequest) error {
	m.mu.Lock()
	m.validated = append(m.validated, req)
	m.mu.Unlock()
	if m.panicsValidate {
		panic(modulePanicDetail)
	}
	return m.validateEr
}

func (m *nodeModuleRecorder) Phase(clusterop.Operation) (nodestatus.Phase, bool) { return "", false }

func (m *nodeModuleRecorder) Run(context.Context, clusterop.Runtime, clusterop.RunRequest) clusterop.Outcome {
	return clusterop.Outcome{Status: clusterop.StatusFailed, Code: clusterop.CodeModuleFailed}
}

func (m *nodeModuleRecorder) ExecuteNode(_ context.Context, req clusterop.NodeRequest) error {
	m.mu.Lock()
	m.executed = append(m.executed, req)
	onExecute := m.onExecute
	m.mu.Unlock()
	if m.panics {
		panic(modulePanicDetail)
	}
	if onExecute != nil {
		return onExecute(req)
	}
	return m.execEr
}

func (m *nodeModuleRecorder) ran() []clusterop.NodeRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]clusterop.NodeRequest(nil), m.executed...)
}

func (m *nodeModuleRecorder) checked() []clusterop.CreateRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]clusterop.CreateRequest(nil), m.validated...)
}

// modulePanicDetail is what a module panics with. It is a distinctive string
// so a test can prove none of it reached the caller.
const modulePanicDetail = "boom: module panic detail"

var (
	_ clusterop.OperationModule     = (*nodeModuleRecorder)(nil)
	_ clusterop.NodeOperationModule = (*nodeModuleRecorder)(nil)
)

// withNodeOperations gives this node a module set of its own for one test.
// The daemon-wide registry is never touched: modules cannot be unregistered
// from it, so a test that added one would change what every later test in
// this binary sees.
func withNodeOperations(t *testing.T, modules ...clusterop.OperationModule) {
	t.Helper()
	registry := clusterop.NewRegistry()
	for _, module := range modules {
		if err := registry.Register(module); err != nil {
			t.Fatalf("register %q: %v", module.Type(), err)
		}
	}
	prev := nodeOperations
	nodeOperations = registry
	t.Cleanup(func() { nodeOperations = prev })
}

// withPowerReachingTheInstalledModule lets a test drive the power route all
// the way to the module it installed.
//
// The real executor serves only the two power operations clusterop wrote
// itself, which it recognizes by a marker no type outside that package can
// carry. That is deliberate and is tested where it lives — in clusterop,
// where a test can hold such a module, and here by the one test that keeps
// the real executor. It does mean a test out here cannot both install a fake
// reboot and reach it, and the alternative is a module that really powers
// the machine running this test.
//
// So what this swaps in is the rest of that executor: reach the module the
// installed set holds for the type, and let what it said travel back
// unchanged, which is what the power endpoint does with a module it is
// willing to serve.
func withPowerReachingTheInstalledModule(t *testing.T) {
	t.Helper()
	prev := powerNodeEndpoint.execute
	powerNodeEndpoint.execute = executeInstalledNodeModule
	t.Cleanup(func() { powerNodeEndpoint.execute = prev })
}

func executeInstalledNodeModule(ctx context.Context, registry *clusterop.ModuleRegistry,
	req clusterop.NodeRequest) error {
	unsupported := &clusterop.PowerError{
		Code: clusterop.CodeUnsupportedOperation, Message: "this daemon does not perform that operation",
	}
	if registry == nil {
		return unsupported
	}
	module, ok := registry.Lookup(req.Type)
	if !ok {
		return unsupported
	}
	nodeModule, ok := module.(clusterop.NodeOperationModule)
	if !ok {
		return unsupported
	}
	return nodeModule.ExecuteNode(ctx, req)
}

// withReplayGuard gives the node routes somewhere to record the signature
// they just spent, in a directory that goes away with the test.
func withReplayGuard(t *testing.T) {
	t.Helper()
	claims, err := clusterop.NewReplayGuard(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	prev := powerClaims
	powerClaims = claims
	t.Cleanup(func() { powerClaims = prev })
}

// withCurrentState sets both what the middleware reads from the live state and
// what the handler receives as its snapshot, so the two cannot disagree.
func withCurrentState(t *testing.T, s clistate.State, observedAt time.Time) {
	t.Helper()

	state.TerminusStateMu.Lock()
	prevLive := state.CurrentState
	state.CurrentState = s
	state.TerminusStateMu.Unlock()

	prevSnapshot := currentStateSnapshot
	currentStateSnapshot = func() (clistate.State, time.Time) { return s, observedAt }

	t.Cleanup(func() {
		currentStateSnapshot = prevSnapshot
		state.TerminusStateMu.Lock()
		state.CurrentState = prevLive
		state.TerminusStateMu.Unlock()
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

func callRegisteredMethod(t *testing.T, method, path, body string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		if path == "/cluster/operations" || path == "/command/power-node" ||
			path == "/command/cluster-operation" {
			// Held as raw JSON rather than decoded values: a body is filled
			// in here and put back on the wire, and a test that asserts what
			// reached a module needs the bytes it wrote, not what a round
			// trip through Go's default types would make of them.
			var request map[string]json.RawMessage
			if json.Unmarshal([]byte(body), &request) == nil {
				if _, ok := request["scope"]; !ok {
					request["scope"] = json.RawMessage(`"` + clusterop.ScopeCluster + `"`)
				}
				if _, ok := request["clusterId"]; !ok {
					request["clusterId"] = json.RawMessage(`"cluster-test"`)
				}
				encoded, err := json.Marshal(request)
				if err != nil {
					t.Fatal(err)
				}
				body = string(encoded)
			}
		}
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := server.API.App.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, out
}

type fakeOperations struct {
	mu          sync.Mutex
	created     []clusterop.CreateRequest
	op          clusterop.Operation
	createEr    error
	stored      map[string]clusterop.Operation
	activePhase nodestatus.Phase
}

func (f *fakeOperations) Create(_ context.Context, req clusterop.CreateRequest) (clusterop.Operation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, req)
	if f.createEr != nil {
		return clusterop.Operation{}, f.createEr
	}
	return f.op, nil
}

func (f *fakeOperations) Get(id string) (clusterop.Operation, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	op, ok := f.stored[id]
	return op, ok
}

func (f *fakeOperations) GetByRequest(requestID string) (clusterop.Operation, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, op := range f.stored {
		if op.RequestID == requestID {
			return op, true
		}
	}
	return clusterop.Operation{}, false
}

func (f *fakeOperations) ActivePhase() (nodestatus.Phase, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.activePhase, f.activePhase != ""
}

func (f *fakeOperations) requests() []clusterop.CreateRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]clusterop.CreateRequest(nil), f.created...)
}

func withClusterOperations(t *testing.T, f *fakeOperations) *fakeOperations {
	t.Helper()
	prev := clusterOperations
	clusterOperations = f
	t.Cleanup(func() { clusterOperations = prev })
	return f
}

func operationsMustNotBeCreated(t *testing.T) {
	t.Helper()
	withClusterOperations(t, &fakeOperations{
		createEr: errors.New("a refused request reached the orchestrator"),
	})
}

func sampleOperation() clusterop.Operation {
	now := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)
	return clusterop.Operation{
		ID:        "op-1",
		Type:      clusterop.TypeReboot,
		RequestID: "client-1",
		Owner:     testOwner,
		Status:    clusterop.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Steps:     []clusterop.Step{},
		Nodes:     []clusterop.NodeResult{},
	}
}

func decodeOperation(t *testing.T, body []byte) clusterop.Operation {
	t.Helper()
	var env struct {
		Data clusterop.Operation `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return env.Data
}

// A login token is not enough to power off a cluster. The signature is what
// separates "somebody is signed in" from "the owner asked for this".
func TestCreateClusterOperationRequiresTheOwnerSignature(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for a token without a signature: %s", resp.StatusCode, body)
	}
}

func TestCreateClusterOperationAcceptsTheSignedScanCallbackWithoutAnAccessToken(t *testing.T) {
	f := withClusterOperations(t, &fakeOperations{op: sampleOperation()})
	asOwnerSignature(t)
	asMaster(t)

	headers := signedFor(t, clusterop.TypeReboot, "client-1")
	delete(headers, AUTH_HEADER)
	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	reqs := f.requests()
	if len(reqs) != 1 || reqs[0].Creds.Token != "" {
		t.Fatalf("requests = %+v, want one without an access token", reqs)
	}
}

func TestCreateClusterOperationRejectsAnotherCluster(t *testing.T) {
	operationsMustNotBeCreated(t)
	asOwnerSignature(t)
	asMaster(t)
	withClusterID(t, "cluster-local", nil)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1","scope":"cluster","clusterId":"cluster-remote"}`,
		signedHeaders(t, map[string]any{
			"type": "reboot", "requestId": "client-1", "scope": "cluster",
			"clusterId": "cluster-remote", "expiresAt": time.Now().Add(time.Minute).UnixMilli(),
		}))
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), clusterop.CodeSignatureMismatch) {
		t.Fatalf("status = %d, want an opaque binding mismatch: %s", resp.StatusCode, body)
	}
}

// The signature must be checked against the owner, not merely be present.
func TestCreateClusterOperationRefusesANonOwnerSignature(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	olaresIDFromRelease = func() (string, error) { return "bob@olares.com", nil }
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for somebody else's signature: %s", resp.StatusCode, body)
	}
}

// A worker holds a partial view of the cluster. Letting it orchestrate one
// would power off the nodes it happens to know about.
func TestCreateClusterOperationIsRefusedOnAWorker(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 on a worker: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "master") {
		t.Errorf("the refusal should say the route is master-only: %s", body)
	}
}

func TestCreateClusterOperationRejectsUnusableRequests(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"unknown type", `{"type":"halt","requestId":"client-1"}`},
		{"no type", `{"requestId":"client-1"}`},
		{"no request id", `{"type":"reboot"}`},
		{"blank request id", `{"type":"reboot","requestId":"   "}`},
		{"not json", `{`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			operationsMustNotBeCreated(t)
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asMaster(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations", tc.body, signedFor(t, clusterop.TypeReboot, "client-1"))

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
			}
		})
	}
}

// The response comes back before the cluster has powered anything: the caller
// is handed an id and polls it.
func TestCreateClusterOperationReturnsTheOperationImmediately(t *testing.T) {
	f := withClusterOperations(t, &fakeOperations{op: sampleOperation()})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	headers := signedFor(t, clusterop.TypeReboot, "client-1")
	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	op := decodeOperation(t, body)
	if op.ID != "op-1" || op.Status != clusterop.StatusPending {
		t.Errorf("operation on the wire = %+v", op)
	}

	reqs := f.requests()
	if len(reqs) != 1 {
		t.Fatalf("orchestrator called %d times", len(reqs))
	}
	if reqs[0].Type != clusterop.TypeReboot || reqs[0].RequestID != "client-1" {
		t.Errorf("request = %+v", reqs[0])
	}
	if reqs[0].Owner != testOwner {
		t.Errorf("owner = %q, want the identity that signed the request", reqs[0].Owner)
	}
	if reqs[0].Creds.Token != testToken {
		t.Errorf("the access token was not forwarded for the precheck: %+v", reqs[0].Creds)
	}
	if reqs[0].Creds.Signature != headers[SIGNATURE_HEADER] {
		t.Errorf("the operation-bound signature was not forwarded to the node hop: %+v", reqs[0].Creds)
	}
}

func TestCreateClusterOperationAcceptsShutdown(t *testing.T) {
	op := sampleOperation()
	op.Type = clusterop.TypeShutdown
	f := withClusterOperations(t, &fakeOperations{op: op})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"shutdown","requestId":"client-1"}`, signedFor(t, clusterop.TypeShutdown, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := f.requests()[0].Type; got != clusterop.TypeShutdown {
		t.Errorf("type = %q", got)
	}
}

// A second power operation while one is running would race over the same
// machines. The caller is told which one is in the way.
func TestCreateClusterOperationReportsAConflict(t *testing.T) {
	withClusterOperations(t, &fakeOperations{
		createEr: &clusterop.ConflictError{ActiveID: "op-running", ActiveType: clusterop.TypeShutdown},
	})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "op-running") {
		t.Errorf("the caller cannot tell which operation is in the way: %s", body)
	}
}

func TestCreateClusterOperationReportsARequestIDConflict(t *testing.T) {
	withClusterOperations(t, &fakeOperations{
		createEr: &clusterop.RequestConflictError{RequestID: "client-1", ExistingID: "op-existing"},
	})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "op-existing") {
		t.Errorf("the caller cannot identify the existing operation: %s", body)
	}
}

func TestGetClusterOperationRequiresAuthorization(t *testing.T) {
	withClusterOperations(t, &fakeOperations{stored: map[string]clusterop.Operation{"op-1": sampleOperation()}})
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/op-1", nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without an authorization header: %s", resp.StatusCode, body)
	}
}

// Polling is not a dangerous operation: a signed-in client may watch what the
// owner started, the same way it watches any other cluster read.
func TestGetClusterOperationReturnsTheRecord(t *testing.T) {
	op := sampleOperation()
	at := op.CreatedAt
	op.Status = clusterop.StatusRunning
	op.Steps = []clusterop.Step{{Name: clusterop.StepPrecheck, Status: clusterop.StepSucceeded, StartedAt: &at, FinishedAt: &at}}
	op.Nodes = []clusterop.NodeResult{{NodeName: "worker-1", Status: clusterop.NodeCommandIssued, StartedAt: &at}}
	withClusterOperations(t, &fakeOperations{stored: map[string]clusterop.Operation{"op-1": op}})
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/op-1", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}

	got := decodeOperation(t, body)
	if got.ID != "op-1" || got.Status != clusterop.StatusRunning {
		t.Errorf("operation = %+v", got)
	}
	if len(got.Steps) != 1 || len(got.Nodes) != 1 {
		t.Errorf("steps and per-node results did not survive the wire: %s", body)
	}

	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"id", "type", "requestId", "status", "createdAt", "steps", "nodes"} {
		if _, ok := env.Data[key]; !ok {
			t.Errorf("field %q missing from the wire format: %s", key, body)
		}
	}
	for _, secret := range []string{signatureCarrying(t, ownerBinding(clusterop.TypeReboot, "client-1")), testToken} {
		if strings.Contains(string(body), secret) {
			t.Errorf("the record carries a credential: %s", body)
		}
	}
}

func TestGetClusterOperationIsNotFoundForAnUnknownID(t *testing.T) {
	withClusterOperations(t, &fakeOperations{stored: map[string]clusterop.Operation{}})
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/op-missing", authHeaders())

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", resp.StatusCode, body)
	}
}

func TestGetClusterOperationIsRefusedOnAWorker(t *testing.T) {
	withClusterOperations(t, &fakeOperations{stored: map[string]clusterop.Operation{"op-1": sampleOperation()}})
	asAuthorizedUser(t)
	asWorker(t)

	resp, body := callRegistered(t, "/cluster/operations/op-1", authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 on a worker: %s", resp.StatusCode, body)
	}
}

func TestGetClusterOperationByRequestReturnsTheRecord(t *testing.T) {
	op := sampleOperation()
	op.RequestID = "client/request 1"
	withClusterOperations(t, &fakeOperations{stored: map[string]clusterop.Operation{op.ID: op}})
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/by-request/client%2Frequest%201", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := decodeOperation(t, body); got.ID != op.ID {
		t.Fatalf("operation = %+v, want %s", got, op.ID)
	}
}

func TestGetClusterOperationByRequestRequiresAuthorization(t *testing.T) {
	withClusterOperations(t, &fakeOperations{})
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/by-request/client-1", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
	}
}

func TestGetClusterOperationByRequestIsRefusedOnAWorker(t *testing.T) {
	withClusterOperations(t, &fakeOperations{})
	asAuthorizedUser(t)
	asWorker(t)

	resp, body := callRegistered(t, "/cluster/operations/by-request/client-1", authHeaders())
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 on a worker: %s", resp.StatusCode, body)
	}
}

func TestGetClusterOperationByRequestRejectsAnEmptyRequestID(t *testing.T) {
	withClusterOperations(t, &fakeOperations{})
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/by-request/", authHeaders())
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
}

func TestGetClusterOperationByRequestIsNotFound(t *testing.T) {
	withClusterOperations(t, &fakeOperations{})
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/operations/by-request/missing", authHeaders())
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", resp.StatusCode, body)
	}
}

// Params are the module's input. They travel from the caller to the module
// untouched, because a module is entitled to read them as the caller wrote
// them: a number this route re-encoded on the way through would be a
// different request than the one that was asked for.
func TestCreateClusterOperationAcceptsParams(t *testing.T) {
	f := withClusterOperations(t, &fakeOperations{op: sampleOperation()})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1","params":{"drain":true,"grace":10000000000000000001}}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	reqs := f.requests()
	if len(reqs) != 1 {
		t.Fatalf("orchestrator called %d times", len(reqs))
	}
	const want = `{"drain":true,"grace":10000000000000000001}`
	if got := strings.TrimSpace(string(reqs[0].Params)); got != want {
		t.Errorf("params = %s, want %s", got, want)
	}
}

// Params are absent from most requests, and a module that reads them has to
// be able to tell "nothing was sent" from "something empty was sent".
func TestCreateClusterOperationLeavesAbsentParamsAbsent(t *testing.T) {
	f := withClusterOperations(t, &fakeOperations{op: sampleOperation()})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := f.requests()[0].Params; len(got) != 0 {
		t.Errorf("params = %s, want nothing", got)
	}
}

// The owner signs the operation, not the params under it. Saying otherwise
// anywhere in this codebase would be a lie a reviewer could rely on, so the
// place it matters most is asserted: what a module refuses never becomes an
// operation, and the refusal is the module's to make.
func TestCreateClusterOperationReportsAModuleRefusalAsABadRequest(t *testing.T) {
	const detail = "grace exceeds what this node will wait"
	withClusterOperations(t, &fakeOperations{
		createEr: &clusterop.ModuleValidationError{
			Type: clusterop.TypeReboot, Err: errors.New(detail),
		},
	})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1","params":{"grace":1}}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
	if strings.Contains(string(body), detail) {
		t.Errorf("the reply leaked the module's own sentence: %s", body)
	}
}

// A body whose params are not JSON is not a request at all.
func TestCreateClusterOperationRejectsUnparsableParams(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1","params":{"grace":}}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
}

// The daemon may be serving before the orchestrator has been wired up. The
// caller gets an error rather than a panic.
func TestClusterOperationsWithoutAnOrchestrator(t *testing.T) {
	prev := clusterOperations
	clusterOperations = nil
	t.Cleanup(func() { clusterOperations = prev })
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("POST status = %d, want 503: %s", resp.StatusCode, body)
	}

	resp, body = callRegistered(t, "/cluster/operations/op-1", authHeaders())
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET status = %d, want 503: %s", resp.StatusCode, body)
	}
}

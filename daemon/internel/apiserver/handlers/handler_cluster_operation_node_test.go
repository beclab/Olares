package handlers

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
)

// testOperationType is an operation nothing in this daemon can carry out. A
// test registers a module for it, which is the whole point: the generic node
// route answers for whatever the module set holds, not for the two power
// operations that happen to be built in.
const testOperationType = clusterop.Type("bake-cake")

// withTestNodeOperation installs one module of a type this daemon does not
// have, plus somewhere to record spent signatures, and returns the module.
func withTestNodeOperation(t *testing.T, module *nodeModuleRecorder) *nodeModuleRecorder {
	t.Helper()
	if module.typ == "" {
		module.typ = testOperationType
	}
	withNodeOperations(t, module)
	withReplayGuard(t)
	return module
}

// nodeOperationRequest is what the master's fan-out puts on the wire for one
// node, params included.
const nodeOperationRequest = `{"type":"bake-cake","operationId":"op-1","requestId":"client-1","params":{"flavour":"almond","layers":10000000000000000001}}`

func TestClusterOperationNodeExecutesTheSignedRequest(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	ran := module.ran()
	if len(ran) != 1 {
		t.Fatalf("the module was asked to act %d times, want once", len(ran))
	}
	if ran[0].Type != testOperationType || ran[0].RequestID != "client-1" {
		t.Errorf("request = %+v", ran[0].PeerRequest)
	}
}

// The params are the module's input and reach it exactly as the caller wrote
// them: a number this daemon re-encoded on the way through would be a
// different request than the one asked for.
func TestClusterOperationNodeHandsTheModuleTheParamsUnchanged(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}

	ran := module.ran()
	if len(ran) != 1 {
		t.Fatalf("the module was asked to act %d times, want once", len(ran))
	}
	const want = `{"flavour":"almond","layers":10000000000000000001}`
	if got := strings.TrimSpace(string(ran[0].Params)); got != want {
		t.Errorf("params = %s, want %s", got, want)
	}
}

// The signature is the whole authority this hop needs, and the access token
// deliberately does not cross it.
func TestClusterOperationNodeNeedsNoAccessToken(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	headers := signedFor(t, testOperationType, "client-1")
	delete(headers, AUTH_HEADER)
	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 1 {
		t.Errorf("the module ran %d times", len(module.ran()))
	}
}

func TestClusterOperationNodeRequiresTheOwnerSignature(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without a signature: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("an unsigned request reached the module: %+v", module.ran())
	}
}

// The signature says which operation it authorizes. Presenting one signed
// for something else is how a captured owner grant would otherwise become a
// way to run any registered module.
func TestClusterOperationNodeNeedsASignatureBoundToTheRequest(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-2"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("a request the signature did not authorize reached the module: %+v", module.ran())
	}
}

// The node this request reaches is the node it acts on. A body naming
// another machine cannot redirect it.
func TestClusterOperationNodeCannotBeAimedAtAnotherMachine(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	headers := signedHeaders(t, ownerNodeBinding(testOperationType, "client-1", "worker-2"))
	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		`{"type":"bake-cake","operationId":"op-1","requestId":"client-1","scope":"node","target":"worker-2"}`,
		headers)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("a request aimed at another machine reached the module: %+v", module.ran())
	}
}

func TestClusterOperationNodeRejectsAnotherCluster(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)
	withClusterID(t, "cluster-local", nil)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		`{"type":"bake-cake","operationId":"op-1","requestId":"client-1","scope":"cluster","clusterId":"cluster-remote"}`,
		signedHeaders(t, map[string]any{
			"type": "bake-cake", "requestId": "client-1", "scope": "cluster",
			"clusterId": "cluster-remote", "expiresAt": time.Now().Add(time.Minute).UnixMilli(),
		}))

	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), clusterop.CodeSignatureMismatch) {
		t.Fatalf("status = %d, want an opaque binding mismatch: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("a request for another cluster reached the module: %+v", module.ran())
	}
}

// A type no module in this node's set holds is refused with the same stable
// code the power endpoint refuses an unknown operation with.
func TestClusterOperationNodeRefusesAnUnknownType(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		`{"type":"no-such-operation","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
	if got := reasonOf(t, body); got != clusterop.CodeUnsupportedOperation {
		t.Errorf("reason = %q, want %q: %s", got, clusterop.CodeUnsupportedOperation, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("an unknown type reached a module: %+v", module.ran())
	}
}

func TestClusterOperationNodeRejectsAnUnparsableBody(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		`{"type":"bake-cake","params":{`, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("an unparsable request reached the module: %+v", module.ran())
	}
}

// The module decides whether it can carry out what the request describes.
// The route asked the owner's signature about the operation; only the module
// can answer for the params under it.
func TestClusterOperationNodeDoesNotExecuteWhatTheModuleRefuses(t *testing.T) {
	const detail = "params name a flavour this node has never baked"
	module := withTestNodeOperation(t, &nodeModuleRecorder{validateEr: errors.New(detail)})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("a request the module refused was carried out anyway: %+v", module.ran())
	}
	if strings.Contains(string(body), detail) {
		t.Errorf("the reply leaked the module's own sentence: %s", body)
	}
}

// The module is asked about the request that actually arrived, params
// included, so a refusal can be about them.
func TestClusterOperationNodeShowsTheModuleTheRequestItValidates(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	checked := module.checked()
	if len(checked) != 1 {
		t.Fatalf("the module was asked to validate %d requests, want one", len(checked))
	}
	if checked[0].Type != testOperationType || checked[0].RequestID != "client-1" {
		t.Errorf("validated %+v", checked[0])
	}
	if !strings.Contains(string(checked[0].Params), "almond") {
		t.Errorf("the module was not shown the params it has to judge: %s", checked[0].Params)
	}
}

// A module chooses the code and the message of the error it returns, and
// both would reach whoever called if they were passed along. What it said
// belongs in this node's log; the reply carries the one stable code that
// describes what is actually known.
func TestClusterOperationNodeHidesWhatAModuleFailedWith(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{
		execEr: &clusterop.PowerError{
			Code:    clusterop.CodePowerUnsupported,
			Message: "token=super-secret; reach the oven at 10.0.0.9",
		},
	})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}
	if got := reasonOf(t, body); got != clusterop.CodeModuleFailed {
		t.Errorf("reason = %q, want %q: %s", got, clusterop.CodeModuleFailed, body)
	}
	for _, leak := range []string{"super-secret", "10.0.0.9", clusterop.CodePowerUnsupported} {
		if strings.Contains(string(body), leak) {
			t.Errorf("the reply repeated what the module said (%q): %s", leak, body)
		}
	}
	if len(module.ran()) != 1 {
		t.Errorf("the module ran %d times", len(module.ran()))
	}
}

// One signature authorizes one operation once. A replayed request is refused
// before the module is asked to act a second time.
func TestClusterOperationNodeRefusesAReplayedSignature(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)
	headers := signedFor(t, testOperationType, "client-1")

	for i := 0; i < 2; i++ {
		resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
			nodeOperationRequest, headers)
		if i == 0 && resp.StatusCode != http.StatusOK {
			t.Fatalf("first attempt status = %d: %s", resp.StatusCode, body)
		}
		if i == 1 && resp.StatusCode != http.StatusConflict {
			t.Fatalf("replay status = %d, want 409: %s", resp.StatusCode, body)
		}
	}
	if got := module.ran(); len(got) != 1 {
		t.Fatalf("the module acted %d times, want once", len(got))
	}
}

// The signature is spent before the module is asked anything. A module that
// refuses is still work this node did on the strength of that signature, and
// a caller able to present the same one again and again would have an
// unlimited way to drive somebody else's code with input nothing signed.
func TestClusterOperationNodeSpendsTheSignatureEvenWhenTheModuleRefuses(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{
		validateEr: errors.New("this node has never baked almond"),
	})
	asOwnerSignature(t)
	asWorker(t)
	headers := signedFor(t, testOperationType, "client-1")

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, headers)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("first status = %d, want 400: %s", resp.StatusCode, body)
	}

	resp, body = callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, headers)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("replay status = %d, want 409: %s", resp.StatusCode, body)
	}
	if got := module.checked(); len(got) != 1 {
		t.Errorf("the module was asked to judge %d requests, want one", len(got))
	}
}

// The same holds for a module that cannot answer at all. A panic in Validate
// is the cheapest thing to provoke and the most expensive to repeat.
func TestClusterOperationNodeSpendsTheSignatureEvenWhenValidationPanics(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{panicsValidate: true})
	asOwnerSignature(t)
	asWorker(t)
	headers := signedFor(t, testOperationType, "client-1")

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, headers)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("first status = %d, want 500: %s", resp.StatusCode, body)
	}

	resp, body = callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, headers)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("replay status = %d, want 409: %s", resp.StatusCode, body)
	}
	if got := module.checked(); len(got) != 1 {
		t.Errorf("the module was asked to judge %d requests, want one", len(got))
	}
}

// A module that panics must not take the daemon down with it, and none of
// what it panicked with may reach whoever called.
func TestClusterOperationNodeContainsAModulePanic(t *testing.T) {
	withTestNodeOperation(t, &nodeModuleRecorder{panics: true})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}
	if got := reasonOf(t, body); got != clusterop.CodeModuleFailed {
		t.Errorf("reason = %q, want %q: %s", got, clusterop.CodeModuleFailed, body)
	}
	if strings.Contains(string(body), "boom") || strings.Contains(string(body), "panic") {
		t.Errorf("the reply leaked the panic: %s", body)
	}
}

// A module is code from outside this package, and it is asked to judge a
// request before anything happens. One that panics while judging must not
// take the daemon down, must not have its operation carried out anyway, and
// must not have what it panicked with reach whoever called.
func TestClusterOperationNodeContainsAPanicWhileValidating(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{panicsValidate: true})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 0 {
		t.Errorf("a module that could not judge the request carried it out: %+v", module.ran())
	}
	if strings.Contains(string(body), "boom") || strings.Contains(string(body), "panic") {
		t.Errorf("the reply leaked the panic: %s", body)
	}
}

// The legacy power endpoint hands a request straight to a module: it
// predates module validation and never asks whether the module accepts what
// arrived. So it serves the two power operations this daemon implements
// itself and nothing else. A module registered later is reachable through
// the generic endpoint, which does ask — and is refused by the old one,
// which would otherwise be a way around that question.
func TestOnlyTheGenericEndpointServesAnOperationTheDaemonLearnedLater(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("generic endpoint status = %d: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 1 {
		t.Fatalf("the generic endpoint carried the operation out %d times, want once",
			len(module.ran()))
	}

	resp, body = callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"bake-cake","operationId":"op-2","requestId":"client-2"}`,
		signedFor(t, testOperationType, "client-2"))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("power endpoint status = %d, want 400: %s", resp.StatusCode, body)
	}
	if got := reasonOf(t, body); got != clusterop.CodeUnsupportedOperation {
		t.Errorf("reason = %q, want %q: %s", got, clusterop.CodeUnsupportedOperation, body)
	}
	if len(module.ran()) != 1 {
		t.Errorf("the power endpoint carried out an operation it does not serve: %+v", module.ran())
	}
}

// Nothing in a test binary may be able to act on the machine running it. The
// module set is installed by main, not by package initialization.
func TestClusterOperationNodeRefusesUntilTheModulesAreInstalled(t *testing.T) {
	prev := nodeOperations
	nodeOperations = nil
	t.Cleanup(func() { nodeOperations = prev })
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", resp.StatusCode, body)
	}
}

// The generic endpoint is the master's second hop for everything the power
// endpoint predates, so the control node answers it too: refusing a whole
// role here would decide, for every module, something only a module can.
func TestClusterOperationNodeAnswersOnTheControlNode(t *testing.T) {
	module := withTestNodeOperation(t, &nodeModuleRecorder{})
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/cluster-operation",
		nodeOperationRequest, signedFor(t, testOperationType, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if len(module.ran()) != 1 {
		t.Errorf("the control node did not carry out the operation: %+v", module.ran())
	}
}

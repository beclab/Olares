package handlers

import (
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
)

type powerRecorder struct {
	mu    sync.Mutex
	calls []clusterop.Type
	err   error
}

func (p *powerRecorder) record(t clusterop.Type) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, t)
	return p.err
}

// recordFor is what one module does instead of powering the machine.
func (p *powerRecorder) recordFor(t clusterop.Type) func(clusterop.NodeRequest) error {
	return func(clusterop.NodeRequest) error { return p.record(t) }
}

func (p *powerRecorder) seen() []clusterop.Type {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]clusterop.Type(nil), p.calls...)
}

// withLocalPower stands in for the two modules that really power a machine.
// The power endpoint reaches them the same way every other node-local
// operation is reached — through the installed module set — so a test can
// give this node a set of its own and watch what was asked of it without the
// real modules being reachable from the test binary at all.
func withLocalPower(t *testing.T, r *powerRecorder) *powerRecorder {
	t.Helper()
	withNodeOperations(t,
		&nodeModuleRecorder{typ: clusterop.TypeReboot, onExecute: r.recordFor(clusterop.TypeReboot)},
		&nodeModuleRecorder{typ: clusterop.TypeShutdown, onExecute: r.recordFor(clusterop.TypeShutdown)},
	)
	withReplayGuard(t)
	return r
}

func hostMustNotBePowered(t *testing.T) *powerRecorder {
	t.Helper()
	r := &powerRecorder{}
	withLocalPower(t, r)
	t.Cleanup(func() {
		if len(r.seen()) != 0 {
			t.Errorf("a refused request powered the machine: %v", r.seen())
		}
	})
	return r
}

// The peer endpoint is reached by the master over the cluster network. It is
// guarded exactly like the single-node power command it wraps, so it hands
// nobody an authority they did not already have.
func TestPowerNodeRequiresTheOwnerSignature(t *testing.T) {
	hostMustNotBePowered(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`, authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without a signature: %s", resp.StatusCode, body)
	}
}

// The worker hop carries no access token. Forwarding the caller's would hand
// every node in the cluster a credential good for every route that user can
// reach, for as long as it lives; the operation-bound signature is narrower
// than that and is the whole authority this endpoint needs.
func TestPowerNodeNeedsNoAccessToken(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asOwnerSignature(t)
	asWorker(t)

	headers := signedFor(t, clusterop.TypeReboot, "client-1")
	delete(headers, AUTH_HEADER)
	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 1 {
		t.Errorf("powered %v", got)
	}
}

func TestPowerNodeRefusesANonOwnerSignature(t *testing.T) {
	hostMustNotBePowered(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	olaresIDFromRelease = func() (string, error) { return "bob@olares.com", nil }
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`, signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for somebody else's signature: %s", resp.StatusCode, body)
	}
}

// Every node answers this, including a worker: it is how the master powers one.
func TestPowerNodePowersTheMachineItReaches(t *testing.T) {
	for _, tc := range []struct {
		body string
		want clusterop.Type
	}{
		{`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`, clusterop.TypeReboot},
		{`{"type":"shutdown","operationId":"op-1","requestId":"client-1"}`, clusterop.TypeShutdown},
	} {
		t.Run(string(tc.want), func(t *testing.T) {
			r := withLocalPower(t, &powerRecorder{})
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asWorker(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node", tc.body, signedFor(t, tc.want, "client-1"))

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d: %s", resp.StatusCode, body)
			}
			if got := r.seen(); len(got) != 1 || got[0] != tc.want {
				t.Errorf("powered %v, want one %s", got, tc.want)
			}
		})
	}
}

// The power endpoint decides nothing about what reboot or shutdown means. It
// hands the request to the module registered for the type, which is the same
// module the generic endpoint would reach and the same one the master's own
// node runs, so there is one implementation of powering a machine rather
// than one per entry point.
func TestPowerNodeCarriesTheRequestOutThroughTheRegisteredModule(t *testing.T) {
	module := &nodeModuleRecorder{typ: clusterop.TypeReboot}
	withNodeOperations(t, module)
	withReplayGuard(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	ran := module.ran()
	if len(ran) != 1 {
		t.Fatalf("the module was asked to act %d times, want once", len(ran))
	}
	if ran[0].Type != clusterop.TypeReboot || ran[0].RequestID != "client-1" ||
		ran[0].OperationID != "op-1" {
		t.Errorf("the module was not shown the request that arrived: %+v", ran[0].PeerRequest)
	}
}

// An old master reaches this endpoint, and an old master sends no params.
// Whatever a new one puts in the body, the module sees the request it was
// handed rather than something this route invented.
func TestPowerNodeSendsNoParamsItWasNotGiven(t *testing.T) {
	module := &nodeModuleRecorder{typ: clusterop.TypeReboot}
	withNodeOperations(t, module)
	withReplayGuard(t)
	asOwnerSignature(t)
	asWorker(t)

	callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	ran := module.ran()
	if len(ran) != 1 {
		t.Fatalf("the module was asked to act %d times, want once", len(ran))
	}
	if len(ran[0].Params) != 0 {
		t.Errorf("params = %s, want nothing", ran[0].Params)
	}
}

func TestPowerNodeRefusesTheControlNode(t *testing.T) {
	hostMustNotBePowered(t)
	asOwnerSignature(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", resp.StatusCode, body)
	}
}

func TestPowerNodeExecutesAClaimedRequestOnlyOnce(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asOwnerSignature(t)
	asWorker(t)
	headers := signedFor(t, clusterop.TypeReboot, "client-1")
	request := `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`

	for i := 0; i < 2; i++ {
		resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node", request, headers)
		if i == 0 && resp.StatusCode != http.StatusOK {
			t.Fatalf("attempt %d status = %d: %s", i+1, resp.StatusCode, body)
		}
		if i == 1 && resp.StatusCode != http.StatusConflict {
			t.Fatalf("attempt %d status = %d, want replay conflict: %s", i+1, resp.StatusCode, body)
		}
	}
	if got := r.seen(); len(got) != 1 {
		t.Fatalf("powered %v, want one execution", got)
	}
}

func TestFailedPowerExecutionCanBeRetried(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{err: errors.New("command was rejected")})
	asOwnerSignature(t)
	asWorker(t)
	headers := signedFor(t, clusterop.TypeReboot, "client-retry")
	request := `{"type":"reboot","operationId":"op-retry","requestId":"client-retry"}`

	resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/power-node", request, headers)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("first status = %d, want 500", resp.StatusCode)
	}
	r.mu.Lock()
	r.err = nil
	r.mu.Unlock()
	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node", request, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("retry status = %d: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 2 {
		t.Fatalf("execution attempts = %v, want failed attempt and retry", got)
	}
}

// The node this request reaches is the node it acts on. A body that names
// another machine cannot redirect it, which is what keeps the endpoint from
// becoming a way to power off any node in the cluster.
func TestPowerNodeCannotBeAimedAtAnotherMachine(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1","nodeName":"master-1","target":"10.0.0.1","ip":"10.0.0.1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 0 {
		t.Fatalf("powered %v", got)
	}

}

func TestPowerNodeRejectsAnUnusableRequest(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"unknown type", `{"type":"halt","operationId":"op-1","requestId":"client-1"}`},
		{"no type", `{"operationId":"op-1","requestId":"client-1"}`},
		{"not json", `{`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hostMustNotBePowered(t)
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asWorker(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node", tc.body, signedFor(t, clusterop.TypeReboot, "client-1"))

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", resp.StatusCode, body)
			}
		})
	}
}

func TestPowerNodeReportsAFailureToPower(t *testing.T) {
	withLocalPower(t, &powerRecorder{err: errors.New("shutdown: command not found")})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"shutdown","operationId":"op-1","requestId":"client-1"}`, signedFor(t, clusterop.TypeShutdown, "client-1"))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}
}

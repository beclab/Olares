package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
)

// internalDetail stands for whatever a failing power command says on this
// machine. It belongs in this node's log and nowhere in a reply.
const internalDetail = "unit dbus.socket not found"

// The owner signs "reboot this cluster, request client-1". Anything else the
// same signature is presented for is refused, which is what stops a captured
// twenty-minute owner token from powering the cluster off.
func TestCreateClusterOperationNeedsASignatureBoundToIt(t *testing.T) {
	for _, tc := range []struct {
		name string
		body map[string]any
	}{
		{"a body that binds nothing", map[string]any{"username": "alice"}},
		{"another operation type", ownerBinding(clusterop.TypeShutdown, "client-1")},
		{"another request id", ownerBinding(clusterop.TypeReboot, "client-2")},
		{"a single-node scope", map[string]any{
			"type": "reboot", "requestId": "client-1", "scope": "node",
			"expiresAt": time.Now().Add(5 * time.Minute).UnixMilli(),
		}},
		{"an expired binding", map[string]any{
			"type": "reboot", "requestId": "client-1", "scope": clusterop.ScopeCluster,
			"expiresAt": time.Now().Add(-time.Minute).UnixMilli(),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			operationsMustNotBeCreated(t)
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asMaster(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
				`{"type":"reboot","requestId":"client-1"}`, signedHeaders(t, tc.body))

			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
			}
		})
	}
}

func TestCreateClusterOperationAcceptsASignatureBoundToIt(t *testing.T) {
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
	reqs := f.requests()
	if len(reqs) != 1 {
		t.Fatalf("orchestrator called %d times", len(reqs))
	}
	if reqs[0].Creds.Signature != headers[SIGNATURE_HEADER] {
		t.Error("the bound signature was not handed to the run that has to present it again")
	}
}

// The worker hop presents the same operation-bound signature. It is not the
// access token: forwarding that would hand every node a credential good for
// every route the user can reach.
func TestPowerNodeNeedsASignatureBoundToTheRequestItReceived(t *testing.T) {
	for _, tc := range []struct {
		name string
		body map[string]any
		req  string
	}{
		{
			name: "a body that binds nothing",
			body: map[string]any{"username": "alice"},
			req:  `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		},
		{
			name: "another operation type",
			body: ownerBinding(clusterop.TypeShutdown, "client-1"),
			req:  `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		},
		{
			name: "another request id",
			body: ownerBinding(clusterop.TypeReboot, "client-2"),
			req:  `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		},
		{
			name: "a request that names no binding to check",
			body: ownerBinding(clusterop.TypeReboot, "client-1"),
			req:  `{"type":"reboot","operationId":"op-1"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hostMustNotBePowered(t)
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asWorker(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
				tc.req, signedHeaders(t, tc.body))

			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want 403: %s", resp.StatusCode, body)
			}
		})
	}
}

func TestPowerNodeAcceptsTheOperationItWasSignedFor(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 1 || got[0] != clusterop.TypeReboot {
		t.Errorf("powered %v", got)
	}
}

func TestDirectNodePowerEndpointIsNotRegistered(t *testing.T) {
	resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/power-this-node",
		`{"type":"reboot","requestId":"client-1"}`, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

// A refusal names a stable code and nothing else. The internal error text of a
// power command is for this node's log, not for whoever called it.
func TestPowerNodeRefusalsCarryAStableCode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(t *testing.T) *powerRecorder
		req    string
		status int
		reason string
	}{
		{
			name:   "a signature bound to something else",
			setup:  hostMustNotBePowered,
			req:    `{"type":"shutdown","operationId":"op-1","requestId":"client-1"}`,
			status: http.StatusForbidden,
			reason: clusterop.CodeSignatureMismatch,
		},
		{
			name:   "an operation this daemon cannot perform",
			setup:  hostMustNotBePowered,
			req:    `{"type":"halt","operationId":"op-1","requestId":"client-1"}`,
			status: http.StatusBadRequest,
			reason: clusterop.CodeUnsupportedOperation,
		},
		{
			name: "a node that cannot power its own machine",
			setup: func(t *testing.T) *powerRecorder {
				return withLocalPower(t, &powerRecorder{err: &clusterop.PowerError{
					Code:    clusterop.CodePowerUnsupported,
					Message: "olaresd runs in a container on this node",
				}})
			},
			req:    `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
			status: http.StatusConflict,
			reason: clusterop.CodePowerUnsupported,
		},
		{
			name: "a power command that failed",
			setup: func(t *testing.T) *powerRecorder {
				return withLocalPower(t, &powerRecorder{err: &clusterop.PowerError{
					Code:    clusterop.CodeHostPowerFailed,
					Message: "this node could not be powered",
					Err:     errors.New(internalDetail),
				}})
			},
			req:    `{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
			status: http.StatusInternalServerError,
			reason: clusterop.CodeHostPowerFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			asAuthorizedUser(t)
			asOwnerSignature(t)
			asWorker(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
				tc.req, signedFor(t, clusterop.TypeReboot, "client-1"))

			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d: %s", resp.StatusCode, tc.status, body)
			}
			if got := reasonOf(t, body); got != tc.reason {
				t.Errorf("reason = %q, want %q: %s", got, tc.reason, body)
			}
			if strings.Contains(string(body), internalDetail) {
				t.Errorf("the reply leaked this node's internal error: %s", body)
			}
		})
	}
}

func reasonOf(t *testing.T, body []byte) string {
	t.Helper()
	var env struct {
		Data struct {
			Reason string `json:"reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return env.Data.Reason
}

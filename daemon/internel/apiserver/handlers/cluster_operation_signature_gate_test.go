package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
)

// A type that registered itself as needing a signature still cannot be created
// with only an access token.
func TestClusterOperationSignatureGateStillRequiresSignatureForRegisteredTypes(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"reboot","requestId":"client-1"}`, authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for a registered type without a signature: %s", resp.StatusCode, body)
	}
}

// An empty type fails closed to the signature path rather than falling through
// to the access-token path.
func TestClusterOperationSignatureGateFailsClosedOnEmptyType(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asMaster(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"requestId":"client-1"}`, authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for an empty type without a signature: %s", resp.StatusCode, body)
	}
}

// A type that did not register a signature requirement is admitted with an
// access token past the signature middleware. It still fails later at
// ParseType for an unknown module — this test only proves RequireSignature
// was not the gate that stopped it.
func TestClusterOperationSignatureGateAdmitsUnregisteredTypeWithAccessToken(t *testing.T) {
	operationsMustNotBeCreated(t)
	asAuthorizedUser(t)
	asMaster(t)

	typ := "no-signature-required-example"
	if clusterop.RequiresSignature(clusterop.Type(typ)) {
		t.Fatalf("%q unexpectedly requires a signature", typ)
	}

	resp, body := callRegisteredMethod(t, http.MethodPost, "/cluster/operations",
		`{"type":"`+typ+`","requestId":"client-1"}`, authHeaders())

	if resp.StatusCode == http.StatusForbidden && strings.Contains(string(body), "request is forbidden") {
		t.Fatalf("signature middleware rejected an unregistered type that presented an access token: %s", body)
	}
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("orchestrator accepted %q; the request should not create an operation", typ)
	}
}

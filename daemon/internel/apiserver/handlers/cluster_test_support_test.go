package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
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
	t.Cleanup(func() {
		newTermipassClient = prevClient
		olaresIDFromRelease = prevID
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

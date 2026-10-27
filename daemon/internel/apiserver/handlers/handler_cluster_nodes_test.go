package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// callRegistered drives the routes the daemon actually serves, middleware
// included. A test that mounts a bare handler on a throwaway app keeps passing
// after somebody deletes the guard from the registration, which is the one
// mistake worth catching here.
func callRegistered(t *testing.T, path string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := server.API.App.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, body
}

func withNodeDirectory(t *testing.T, fn func(context.Context) ([]inventory.Node, error)) {
	t.Helper()
	prev := listClusterNodes
	listClusterNodes = fn
	t.Cleanup(func() { listClusterNodes = prev })
}

func directoryMustNotBeRead(t *testing.T) {
	t.Helper()
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		t.Error("the node directory was read by a request that should have been refused")
		return nil, nil
	})
}

func TestClusterNodesRefusesAnUnauthorizedRequest(t *testing.T) {
	directoryMustNotBeRead(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster/nodes", nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without an authorization header: %s", resp.StatusCode, body)
	}
}

func TestClusterNodesAcceptsAnOwnerSignatureForNodeOperationRouting(t *testing.T) {
	asOwnerSignature(t)
	asMaster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{{NodeName: "worker-1", IP: "10.0.0.2"}}, nil
	})
	headers := map[string]string{
		SIGNATURE_HEADER: signatureCarrying(t,
			ownerNodeBinding("reboot", "client-1", "worker-1")),
	}

	resp, body := callRegistered(t, "/cluster/nodes", headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}
}

// The directory is the master's answer. A worker holds a partial view, and
// answering with it would let a caller act on a cluster it cannot see.
func TestClusterNodesIsRefusedOnAWorker(t *testing.T) {
	directoryMustNotBeRead(t)
	asAuthorizedUser(t)
	asWorker(t)

	resp, body := callRegistered(t, "/cluster/nodes", authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 on a worker: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "master") {
		t.Errorf("the refusal should say the route is master-only: %s", body)
	}
}

func TestClusterNodesReturnsEveryNodeOnTheMaster(t *testing.T) {
	asAuthorizedUser(t)
	asMaster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
			{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: false},
			{NodeName: "worker-noip", Role: inventory.RoleWorker, Ready: false},
		}, nil
	})

	resp, body := callRegistered(t, "/cluster/nodes", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}

	var env struct {
		Data []inventory.Node `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if len(env.Data) != 3 {
		t.Fatalf("want 3 nodes on the wire, got %d: %s", len(env.Data), body)
	}
	if env.Data[1].Ready || env.Data[2].IP != "" {
		t.Errorf("NotReady / unaddressable nodes altered on the wire: %s", body)
	}
	if env.Data[0].Role != inventory.RoleMaster || !env.Data[0].IsSelf {
		t.Errorf("master entry mangled: %s", body)
	}
}

func TestClusterNodesWireFieldNames(t *testing.T) {
	asAuthorizedUser(t)
	asMaster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true}}, nil
	})

	_, body := callRegistered(t, "/cluster/nodes", authHeaders())

	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if len(env.Data) != 1 {
		t.Fatalf("want one node, got %s", body)
	}
	for _, key := range []string{"nodeName", "role", "ip", "ready", "isSelf"} {
		if _, ok := env.Data[0][key]; !ok {
			t.Errorf("field %q missing from the node directory wire format: %s", key, body)
		}
	}
}

func TestClusterNodesEmptyClusterIsAList(t *testing.T) {
	asAuthorizedUser(t)
	asMaster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{}, nil
	})

	_, body := callRegistered(t, "/cluster/nodes", authHeaders())

	var env struct {
		Data []inventory.Node `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if env.Data == nil {
		t.Errorf("want [] rather than null: %s", body)
	}
}

// Whatever went wrong reading the cluster, the caller is a Settings page. It
// gets one stable sentence; the endpoint, the certificate and the address stay
// in the daemon log.
func TestClusterNodesFailureDoesNotLeakClusterInternals(t *testing.T) {
	asAuthorizedUser(t)
	asMaster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return nil, errors.New(`Get "https://10.0.0.1:6443/api/v1/nodes": x509: certificate signed by unknown authority`)
	})

	resp, body := callRegistered(t, "/cluster/nodes", authHeaders())
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}

	text := string(body)
	for _, leak := range []string{"x509", "10.0.0.1", "6443", "certificate", "https://"} {
		if strings.Contains(text, leak) {
			t.Errorf("response leaks %q: %s", leak, text)
		}
	}
	if !strings.Contains(text, nodeDirectoryUnavailable) {
		t.Errorf("want the stable message %q: %s", nodeDirectoryUnavailable, text)
	}
}

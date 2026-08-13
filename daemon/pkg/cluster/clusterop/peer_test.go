package clusterop

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// A node that did not answer is not a node that refused: the two are told
// apart so the operation record says which happened.
func TestPeerOutcomesClassifyEveryFanOutResult(t *testing.T) {
	results := []fanout.NodeResult{
		{Node: fanout.NodeTarget{Name: "worker-1"}, Status: fanout.StatusOK},
		{Node: fanout.NodeTarget{Name: "worker-2"}, Status: fanout.StatusUnreachable, Err: "no route to host"},
		{Node: fanout.NodeTarget{Name: "worker-3"}, Status: fanout.StatusTimeout, Err: "deadline exceeded"},
		{Node: fanout.NodeTarget{Name: "worker-4"}, Status: fanout.StatusError, Err: "node returned 403"},
	}

	got := peerOutcomes(results)
	if len(got) != 4 {
		t.Fatalf("got %d outcomes", len(got))
	}
	want := map[string]string{
		"worker-1": "",
		"worker-2": CodeNodeUnreachable,
		"worker-3": CodeNodeUnreachable,
		"worker-4": CodeDispatchFailed,
	}
	for _, o := range got {
		if want[o.NodeName] != o.Code {
			t.Errorf("%s code = %q, want %q", o.NodeName, o.Code, want[o.NodeName])
		}
		if o.Code != "" && o.Err == "" {
			t.Errorf("%s failed without saying why", o.NodeName)
		}
	}
}

func TestPeerTargetsCarryWhatTheFanOutNeeds(t *testing.T) {
	targets := peerTargets([]inventory.Node{
		{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2"},
		{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", IsSelf: true},
	})

	if len(targets) != 2 {
		t.Fatalf("got %d targets", len(targets))
	}
	if targets[0].Name != "worker-1" || targets[0].IP != "10.0.0.2" || targets[0].IsMaster {
		t.Errorf("worker target = %+v", targets[0])
	}
	if !targets[1].IsSelf || !targets[1].IsMaster {
		t.Errorf("master target = %+v", targets[1])
	}
}

func TestNodeStatusPrecheckUsesTheLeastCredentialAvailable(t *testing.T) {
	if got := nodeStatusHeaders(Credentials{Token: "token", Signature: "signature"}); got["X-Authorization"] != "token" {
		t.Errorf("headers = %v, want the local access token when available", got)
	}
	got := nodeStatusHeaders(Credentials{Signature: "signature"})
	if len(got) != 1 || got[signatureHeaderName] != "signature" {
		t.Errorf("headers = %v, want only the operation-bound signature", got)
	}
}

// peerCall is one request a node's olaresd received.
type peerCall struct {
	path   string
	body   []byte
	header http.Header
}

// peerRecorder stands in for another node's olaresd. It records what the
// fan-out actually put on the wire, so a test can assert the endpoint a
// dispatch reached rather than the constant it was built from — which is
// what rolling compatibility with an older worker depends on.
type peerRecorder struct {
	mu    sync.Mutex
	calls []peerCall
}

func newPeerRecorder(t *testing.T) *peerRecorder {
	t.Helper()
	r := &peerRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		r.mu.Lock()
		r.calls = append(r.calls, peerCall{path: req.URL.Path, body: body, header: req.Header.Clone()})
		r.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse %s: %v", server.URL, err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("port of %s: %v", server.URL, err)
	}
	prev := peerPort
	peerPort = port
	t.Cleanup(func() { peerPort = prev })
	return r
}

func (r *peerRecorder) seen() []peerCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]peerCall(nil), r.calls...)
}

// oneCall fails unless exactly one request arrived, and returns it.
func (r *peerRecorder) oneCall(t *testing.T) peerCall {
	t.Helper()
	calls := r.seen()
	if len(calls) != 1 {
		t.Fatalf("the fan-out made %d calls, want 1", len(calls))
	}
	return calls[0]
}

func recordingNode() []inventory.Node {
	return []inventory.Node{{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "127.0.0.1", Ready: true}}
}

// A new master must still be able to power an older worker, which serves the
// power endpoint and nothing else. The built-in power operations therefore
// keep dispatching to the path that endpoint has always been on.
func TestBuiltInPowerDispatchUsesLegacyPath(t *testing.T) {
	peer := newPeerRecorder(t)

	for _, opType := range []Type{TypeReboot, TypeShutdown} {
		outcomes := dispatchPower(context.Background(), recordingNode(), PeerRequest{
			Type: opType, OperationID: "op-1", RequestID: "client-1",
			Scope: ScopeCluster, ClusterID: "cluster-1",
		}, Credentials{Signature: "signature"})
		if len(outcomes) != 1 || outcomes[0].Code != "" {
			t.Fatalf("%s dispatch = %+v, want the node to have accepted it", opType, outcomes)
		}
	}

	calls := peer.seen()
	if len(calls) != 2 {
		t.Fatalf("the fan-out made %d calls, want one per operation", len(calls))
	}
	for _, call := range calls {
		if call.path != PeerPath {
			t.Errorf("power dispatch reached %q, want %q", call.path, PeerPath)
		}
	}
}

// Everything else goes to the generic endpoint, which is the one that
// carries module params.
func TestGenericNodeDispatchUsesTheClusterOperationPath(t *testing.T) {
	peer := newPeerRecorder(t)

	outcomes := DispatchNodeOperation(context.Background(), recordingNode(), NodeRequest{
		PeerRequest: PeerRequest{
			Type: Type("bake-cake"), OperationID: "op-1", RequestID: "client-1",
			Scope: ScopeNode, Target: "worker-1", ClusterID: "cluster-1",
		},
	}, Credentials{Signature: "signature"})

	if len(outcomes) != 1 || outcomes[0].Code != "" {
		t.Fatalf("dispatch = %+v, want the node to have accepted it", outcomes)
	}
	if got := peer.oneCall(t).path; got != ClusterOperationPath {
		t.Errorf("generic dispatch reached %q, want %q", got, ClusterOperationPath)
	}
}

// The node receives the caller's params exactly as they were written. They
// are the module's input, and a number the daemon re-encoded on the way
// through would reach the node as a different request than the one asked
// for.
func TestGenericNodeDispatchCarriesTheParamsUnchanged(t *testing.T) {
	peer := newPeerRecorder(t)
	params := json.RawMessage(`{"value":10000000000000000001}`)

	DispatchNodeOperation(context.Background(), recordingNode(), NodeRequest{
		PeerRequest: PeerRequest{
			Type: Type("bake-cake"), OperationID: "op-1", RequestID: "client-1",
			Scope: ScopeNode, Target: "worker-1", ClusterID: "cluster-1",
		},
		Params: params,
	}, Credentials{Signature: "signature"})

	var sent struct {
		Params json.RawMessage `json:"params"`
	}
	body := peer.oneCall(t).body
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if string(sent.Params) != string(params) {
		t.Errorf("params on the wire = %s, want %s", sent.Params, params)
	}
}

// The generic hop for a signature-bound type is no wider than the power hop:
// the operation-bound signature crosses it and the caller's access token does
// not, because a token opens every route that user can reach for as long as
// it lives.
func TestGenericNodeDispatchPresentsOnlyTheOperationSignature(t *testing.T) {
	peer := newPeerRecorder(t)

	DispatchNodeOperation(context.Background(), recordingNode(), NodeRequest{
		PeerRequest: PeerRequest{
			Type: TypeReboot, OperationID: "op-1", RequestID: "client-1",
			Scope: ScopeNode, Target: "worker-1", ClusterID: "cluster-1",
		},
	}, Credentials{Token: "access-token", Signature: "signature"})

	header := peer.oneCall(t).header
	if got := header.Get(signatureHeaderName); got != "signature" {
		t.Errorf("%s = %q, want the operation-bound signature", signatureHeaderName, got)
	}
	if got := header.Get(authorizationHeaderName); got != "" {
		t.Errorf("%s = %q, want the access token left behind", authorizationHeaderName, got)
	}
}

// A type that did not register a signature requirement is admitted on the
// master with an access token, so that same token is what the fan-out presents
// to the node — otherwise the node hop would have no credential at all.
func TestGenericNodeDispatchPresentsTokenWhenSignatureNotRequired(t *testing.T) {
	peer := newPeerRecorder(t)
	typ := Type("fold-laundry")
	if RequiresSignature(typ) {
		t.Fatalf("%q unexpectedly requires a signature", typ)
	}

	DispatchNodeOperation(context.Background(), recordingNode(), NodeRequest{
		PeerRequest: PeerRequest{
			Type: typ, OperationID: "op-1", RequestID: "client-1",
			Scope: ScopeNode, Target: "worker-1", ClusterID: "cluster-1",
		},
	}, Credentials{Token: "access-token", Signature: "signature"})

	header := peer.oneCall(t).header
	if got := header.Get(authorizationHeaderName); got != "access-token" {
		t.Errorf("%s = %q, want the access token", authorizationHeaderName, got)
	}
	if got := header.Get(signatureHeaderName); got != "" {
		t.Errorf("%s = %q, want the signature left behind for a token-admitted type", signatureHeaderName, got)
	}
}

func TestPeerRequestCarriesTheFullBinding(t *testing.T) {
	body, err := json.Marshal(PeerRequest{
		Type: TypeShutdown, OperationID: "op-1", RequestID: "client-1",
		Scope: ScopeNode, Target: "worker-1", ClusterID: "cluster-1",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fields["scope"] != ScopeNode || fields["target"] != "worker-1" || fields["clusterId"] != "cluster-1" {
		t.Errorf("peer request omitted its authorization binding: %s", body)
	}
}

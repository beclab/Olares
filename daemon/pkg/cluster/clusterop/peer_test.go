package clusterop

import (
	"encoding/json"
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

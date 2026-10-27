package inventory

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"k8s.io/client-go/kubernetes/fake"
)

func node(name string, master, ready bool, ip string) *corev1.Node {
	n := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{}},
	}
	if master {
		n.Labels["node-role.kubernetes.io/control-plane"] = "true"
	}
	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}
	n.Status.Conditions = []corev1.NodeCondition{{Type: corev1.NodeReady, Status: readyStatus}}
	if ip != "" {
		n.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: ip}}
	}
	return n
}

func listerFor(hostIPs []string, objs ...runtime.Object) *Lister {
	return &Lister{
		Client:       fake.NewSimpleClientset(objs...),
		HostIPs:      func() ([]string, error) { return hostIPs, nil },
		SelfNodeName: func() (string, error) { return "not-this-machine", nil },
	}
}

func byName(nodes []Node) map[string]Node {
	m := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		m[n.NodeName] = n
	}
	return m
}

func TestListReturnsEveryNodeIncludingNotReady(t *testing.T) {
	l := listerFor([]string{"10.0.0.1"},
		node("master-1", true, true, "10.0.0.1"),
		node("worker-1", false, true, "10.0.0.2"),
		node("worker-2", false, false, "10.0.0.3"),
	)

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("want 3 nodes, got %d: %+v", len(nodes), nodes)
	}

	got := byName(nodes)
	if got["worker-2"].Ready {
		t.Errorf("worker-2 is NotReady, got ready=true: %+v", got["worker-2"])
	}
	if !got["worker-1"].Ready || !got["master-1"].Ready {
		t.Errorf("ready nodes reported as not ready: %+v", got)
	}
}

func TestListReportsRoles(t *testing.T) {
	l := listerFor(nil,
		node("master-1", true, true, "10.0.0.1"),
		node("worker-1", false, true, "10.0.0.2"),
	)

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byName(nodes)
	if got["master-1"].Role != RoleMaster {
		t.Errorf("master-1 role = %q, want %q", got["master-1"].Role, RoleMaster)
	}
	if got["worker-1"].Role != RoleWorker {
		t.Errorf("worker-1 role = %q, want %q", got["worker-1"].Role, RoleWorker)
	}
}

func TestListLegacyMasterLabelIsMaster(t *testing.T) {
	legacy := node("master-legacy", false, true, "10.0.0.9")
	legacy.Labels["node-role.kubernetes.io/master"] = "true"

	l := listerFor(nil, legacy)
	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nodes[0].Role != RoleMaster {
		t.Errorf("legacy master label not honoured: %+v", nodes[0])
	}
}

func TestListKeepsNodeWithoutInternalIP(t *testing.T) {
	l := listerFor([]string{"10.0.0.1"},
		node("master-1", true, true, "10.0.0.1"),
		node("worker-noip", false, false, ""),
	)

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := byName(nodes)["worker-noip"]
	if !ok {
		t.Fatalf("node without an internal IP was dropped: %+v", nodes)
	}
	if got.IP != "" {
		t.Errorf("want empty IP for a node with no internal address, got %q", got.IP)
	}
	if got.IsSelf {
		t.Errorf("node without an IP must not be claimed as self: %+v", got)
	}
}

func TestListMarksSelfByHostIP(t *testing.T) {
	l := listerFor([]string{"127.0.0.1", "10.0.0.2"},
		node("master-1", true, true, "10.0.0.1"),
		node("worker-1", false, true, "10.0.0.2"),
	)

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byName(nodes)
	if !got["worker-1"].IsSelf {
		t.Errorf("worker-1 matches a host IP, want isSelf=true: %+v", got["worker-1"])
	}
	if got["master-1"].IsSelf {
		t.Errorf("master-1 does not match a host IP, want isSelf=false: %+v", got["master-1"])
	}
}

func TestListStillReturnsNodesWhenHostIPsFail(t *testing.T) {
	l := &Lister{
		Client:       fake.NewSimpleClientset(node("master-1", true, true, "10.0.0.1")),
		HostIPs:      func() ([]string, error) { return nil, errors.New("no interfaces") },
		SelfNodeName: func() (string, error) { return "not-this-machine", nil },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("host IP lookup failure must not fail the directory: %v", err)
	}
	if len(nodes) != 1 || nodes[0].IsSelf {
		t.Errorf("want the node listed with isSelf=false, got %+v", nodes)
	}
}

// A node whose address is gone — kubelet down, IP not yet assigned — is
// exactly the node an operator is looking at, and address matching cannot
// find it. Its Kubernetes name is still known locally.
func TestListMarksSelfByNodeNameWhenNoAddressMatches(t *testing.T) {
	l := &Lister{
		Client: fake.NewSimpleClientset(
			node("master-1", true, true, "10.0.0.1"),
			node("worker-1", false, false, ""),
		),
		HostIPs:      func() ([]string, error) { return []string{"10.0.0.2"}, nil },
		SelfNodeName: func() (string, error) { return "worker-1", nil },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byName(nodes)
	if !got["worker-1"].IsSelf {
		t.Errorf("this machine did not recognize itself by node name: %+v", got["worker-1"])
	}
	if got["master-1"].IsSelf {
		t.Errorf("another node claimed as self: %+v", got["master-1"])
	}
}

func TestListMatchesSelfNodeNameCaseInsensitively(t *testing.T) {
	l := &Lister{
		Client:       fake.NewSimpleClientset(node("worker-1", false, false, "")),
		HostIPs:      func() ([]string, error) { return nil, nil },
		SelfNodeName: func() (string, error) { return "Worker-1", nil },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nodes[0].IsSelf {
		t.Errorf("hostname casing must not hide the local node: %+v", nodes[0])
	}
}

func TestListMarksSelfByHostnameLabel(t *testing.T) {
	labelled := node("node-abcdef", false, true, "")
	labelled.Labels["kubernetes.io/hostname"] = "worker-1"

	l := &Lister{
		Client:       fake.NewSimpleClientset(labelled),
		HostIPs:      func() ([]string, error) { return nil, nil },
		SelfNodeName: func() (string, error) { return "worker-1", nil },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nodes[0].IsSelf {
		t.Errorf("a node registered under a name of its own was not recognized: %+v", nodes[0])
	}
}

func TestListAddressMatchWinsOverName(t *testing.T) {
	l := &Lister{
		Client: fake.NewSimpleClientset(
			node("worker-1", false, true, "10.0.0.2"),
			node("stale-name", false, false, ""),
		),
		HostIPs:      func() ([]string, error) { return []string{"10.0.0.2"}, nil },
		SelfNodeName: func() (string, error) { return "stale-name", nil },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byName(nodes)
	if !got["worker-1"].IsSelf {
		t.Errorf("the address match should identify this machine: %+v", got["worker-1"])
	}
	if got["stale-name"].IsSelf {
		t.Errorf("two nodes cannot both be this machine: %+v", nodes)
	}
}

func TestListStillReturnsNodesWhenSelfNodeNameFails(t *testing.T) {
	l := &Lister{
		Client:       fake.NewSimpleClientset(node("worker-1", false, true, "")),
		HostIPs:      func() ([]string, error) { return nil, nil },
		SelfNodeName: func() (string, error) { return "", errors.New("no hostname") },
	}

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("self lookup failure must not fail the directory: %v", err)
	}
	if len(nodes) != 1 || nodes[0].IsSelf {
		t.Errorf("want the node listed with isSelf=false, got %+v", nodes)
	}
}

func TestListPropagatesClusterError(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "nodes", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("apiserver down")
	})
	l := &Lister{Client: client, HostIPs: func() ([]string, error) { return nil, nil }}

	if _, err := l.List(context.Background()); err == nil {
		t.Fatal("want an error when the node list cannot be read")
	}
}

func TestListEmptyClusterReturnsEmptySlice(t *testing.T) {
	l := listerFor(nil)

	nodes, err := l.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nodes == nil {
		t.Fatal("want a non-nil empty slice so the wire format is [] and not null")
	}
	if len(nodes) != 0 {
		t.Errorf("want no nodes, got %+v", nodes)
	}
}

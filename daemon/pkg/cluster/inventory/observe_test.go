package inventory

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func booted(n *corev1.Node, bootID string) *corev1.Node {
	n.Status.NodeInfo.BootID = bootID
	return n
}

// Watching a node restart is the master's own job: it reads the cluster it
// already talks to, rather than dialling each node's user-facing endpoint.
func TestObserveReportsReadinessAndTheBootEachNodeIsOn(t *testing.T) {
	l := &Lister{Client: fake.NewSimpleClientset(
		booted(node("master-1", true, true, "10.0.0.1"), "boot-m"),
		booted(node("worker-1", false, true, "10.0.0.2"), "boot-w1"),
		booted(node("worker-2", false, false, "10.0.0.3"), "boot-w2"),
	)}

	got, err := l.Observe(context.Background())
	if err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("observed %d nodes, want 3: %+v", len(got), got)
	}
	if o := got["worker-1"]; !o.Ready || o.BootID != "boot-w1" {
		t.Errorf("worker-1 = %+v", o)
	}
	if o := got["worker-2"]; o.Ready || o.BootID != "boot-w2" {
		t.Errorf("worker-2 = %+v", o)
	}
}

// A node that has left the cluster is absent rather than present-and-false.
// The difference is what tells a reboot the node really went away.
func TestObserveOmitsANodeThatIsNoLongerThere(t *testing.T) {
	l := &Lister{Client: fake.NewSimpleClientset(booted(node("master-1", true, true, "10.0.0.1"), "boot-m"))}

	got, err := l.Observe(context.Background())
	if err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if _, ok := got["worker-1"]; ok {
		t.Error("a node that is not in the cluster was observed")
	}
}

// kubelet reports no boot id on a node it has not fully registered. Reporting
// an empty one is honest; inventing one would prove a restart that never
// happened.
func TestObserveReportsAnEmptyBootIDAsEmpty(t *testing.T) {
	l := &Lister{Client: fake.NewSimpleClientset(node("worker-1", false, true, "10.0.0.2"))}

	got, err := l.Observe(context.Background())
	if err != nil {
		t.Fatalf("Observe: %v", err)
	}
	if o := got["worker-1"]; o.BootID != "" {
		t.Errorf("boot id = %q, want empty", o.BootID)
	}
}

func TestObserveReportsAnUnreadableCluster(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "nodes", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection refused")
	})
	l := &Lister{Client: client}

	if _, err := l.Observe(context.Background()); err == nil {
		t.Fatal("an unreadable cluster was reported as an empty one")
	}
}

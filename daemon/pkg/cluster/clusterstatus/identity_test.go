package clusterstatus

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func kubeSystem(uid string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: kubeSystemNamespace, UID: types.UID(uid)},
	}
}

// The identifier is the kube-system namespace's UID: it is created once when
// the cluster is, survives upgrades, restarts and nodes joining, and is
// replaced when the cluster is. Nothing about it is a credential and reading
// it needs no write access anywhere.
func TestClusterIDIsTheKubeSystemNamespaceUID(t *testing.T) {
	client := fake.NewSimpleClientset(kubeSystem("f8c1a0de-1f1e-4c1a-9a52-2c9a0f1c0b11"))

	got, err := ClusterID(context.Background(), client)
	if err != nil {
		t.Fatalf("ClusterID: %v", err)
	}
	if got != "f8c1a0de-1f1e-4c1a-9a52-2c9a0f1c0b11" {
		t.Errorf("clusterId = %q", got)
	}
}

// A cluster that cannot be read has no identifier to report. Making one up
// would hand every reinstall the same name, or every read a different one.
func TestClusterIDFailsRatherThanGuessing(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "namespaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("apiserver unreachable")
	})

	if got, err := ClusterID(context.Background(), client); err == nil {
		t.Fatalf("ClusterID = %q, want an error", got)
	}
}

func TestClusterIDIsEmptyWhenTheNamespaceHasNoUID(t *testing.T) {
	client := fake.NewSimpleClientset(kubeSystem(""))

	if got, err := ClusterID(context.Background(), client); err == nil {
		t.Fatalf("ClusterID = %q, want an error rather than an empty identifier", got)
	}
}

// Reading it must not need more than a read: the whole point of using an
// object the cluster already owns is that olaresd writes nothing to hold it.
func TestClusterIDOnlyReads(t *testing.T) {
	client := fake.NewSimpleClientset(kubeSystem("ns-uid"))

	if _, err := ClusterID(context.Background(), client); err != nil {
		t.Fatalf("ClusterID: %v", err)
	}
	for _, action := range client.Actions() {
		if action.GetVerb() != "get" {
			t.Errorf("clusterId resolution performed a %q on %q", action.GetVerb(), action.GetResource().Resource)
		}
	}
}

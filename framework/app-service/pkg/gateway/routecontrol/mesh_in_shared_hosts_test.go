package routecontrol

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMeshInSharedHostsReconcileCreatesCM(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "litellm-alice",
			Labels: map[string]string{security.NamespaceInClusterCallerLabel: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	r := &MeshInSharedHostsReconciler{Client: c, platformDomain: "olares.com"}
	targets := []SharedHostsTarget{{
		CallerNamespace: "litellm-alice",
		Viewer:          "alice",
		Hosts:           []string{"abcd1234.alice.olares.com"},
	}}
	if err := r.ReconcileNamespace(context.Background(), "litellm-alice", targets); err != nil {
		t.Fatal(err)
	}
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "litellm-alice", Name: constants.MeshInSharedHostsCMName,
	}, cm); err != nil {
		t.Fatal(err)
	}
	body := cm.Data[constants.MeshInSharedHostsFileName]
	if !strings.Contains(body, "abcd1234.alice.olares.com") {
		t.Fatalf("hosts body = %q", body)
	}
	if cm.Labels[constants.MeshInSharedHostsManagedByLabel] != sharedHostsManagedByValue {
		t.Fatalf("managed-by label = %q", cm.Labels[constants.MeshInSharedHostsManagedByLabel])
	}
}

func TestMaterializeHostLogicalPattern(t *testing.T) {
	h, reason := materializeHost("abcd1234.*.olares.com", "alice", "olares.com")
	if reason != "" || h != "abcd1234.alice.olares.com" {
		t.Fatalf("got host=%q reason=%q", h, reason)
	}
	_, reason = materializeHost("x.shared.olares.com", "alice", "olares.com")
	if reason == "" {
		t.Fatal("expected v2 guard drop reason")
	}
}

func TestMeshInTLSSecretName(t *testing.T) {
	if got := meshInTLSSecretName("Alice"); got != "olares-mesh-in-tls-alice" {
		t.Fatalf("got %q", got)
	}
}

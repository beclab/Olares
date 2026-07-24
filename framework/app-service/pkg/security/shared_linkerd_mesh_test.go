package security

import "testing"

func TestNewSharedLinkerdControlPlaneIngressNetworkPolicy(t *testing.T) {
	np := NewSharedLinkerdControlPlaneIngressNetworkPolicy("ollama-shared", map[string]string{"app": "ollama"})
	if np.Name != SharedLinkerdMeshIngressNPName {
		t.Fatalf("name = %q, want %q", np.Name, SharedLinkerdMeshIngressNPName)
	}
	if np.Namespace != "ollama-shared" {
		t.Fatalf("namespace = %q", np.Namespace)
	}
	if got := np.Spec.PodSelector.MatchLabels["app"]; got != "ollama" {
		t.Fatalf("podSelector app = %q", got)
	}
	if len(np.Spec.Ingress) != 1 || len(np.Spec.Ingress[0].From) != 1 {
		t.Fatalf("ingress peers = %#v", np.Spec.Ingress)
	}
	peer := np.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]
	if peer != "os-mesh" {
		t.Fatalf("peer namespace = %q, want os-mesh", peer)
	}
}

func TestSharedLinkerdMeshIngressPeerNamespaces(t *testing.T) {
	if len(SharedLinkerdMeshIngressPeerNamespaces) == 0 {
		t.Fatal("peer list empty")
	}
	if SharedLinkerdMeshIngressPeerNamespaces[0] != "os-mesh" {
		t.Fatalf("first peer = %q, want os-mesh", SharedLinkerdMeshIngressPeerNamespaces[0])
	}
}

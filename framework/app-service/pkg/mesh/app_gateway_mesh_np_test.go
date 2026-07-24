package mesh

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

func TestEnsureAppGatewayMeshNetworkPolicies(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = networkingv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	if err := EnsureAppGatewayMeshNetworkPolicies(context.Background(), c); err != nil {
		t.Fatal(err)
	}

	meshNP := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: security.MeshControlPlaneNamespace,
		Name:      security.AppGatewayMeshNPName,
	}, meshNP); err != nil {
		t.Fatalf("os-mesh NP: %v", err)
	}
	if len(meshNP.Spec.Ingress) != 1 || len(meshNP.Spec.Ingress[0].From) != 4 {
		t.Fatalf("os-mesh from peers = %+v", meshNP.Spec.Ingress)
	}
	foundCaller := false
	for _, peer := range meshNP.Spec.Ingress[0].From {
		if peer.NamespaceSelector == nil {
			continue
		}
		if peer.NamespaceSelector.MatchLabels[security.NamespaceInClusterCallerLabel] == "true" {
			foundCaller = true
		}
	}
	if !foundCaller {
		t.Fatal("expected in-cluster-caller peer on os-mesh NP")
	}
	if len(meshNP.Spec.Ingress[0].Ports) != 5 {
		t.Fatalf("ports = %+v", meshNP.Spec.Ingress[0].Ports)
	}

	gwNP := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: security.AppGatewayNamespace,
		Name:      security.AppGatewayMeshNPName,
	}, gwNP); err != nil {
		t.Fatalf("os-gateway NP: %v", err)
	}
	if len(gwNP.Spec.Ingress) != 1 || len(gwNP.Spec.Ingress[0].From) != 1 {
		t.Fatalf("os-gateway from = %+v", gwNP.Spec.Ingress)
	}
	peerNS := gwNP.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]
	if peerNS != security.MeshControlPlaneNamespace {
		t.Fatalf("os-gateway peer ns = %q", peerNS)
	}

	// Idempotent update path.
	if err := EnsureAppGatewayMeshNetworkPolicies(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}

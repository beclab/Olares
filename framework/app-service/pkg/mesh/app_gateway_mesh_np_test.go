package mesh

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	if len(meshNP.Spec.Ingress) != 2 || len(meshNP.Spec.Ingress[0].From) != 6 {
		t.Fatalf("os-mesh from peers = %+v", meshNP.Spec.Ingress)
	}
	foundCaller := false
	platformPeers := 0
	for _, peer := range meshNP.Spec.Ingress[0].From {
		if peer.NamespaceSelector == nil {
			continue
		}
		if peer.PodSelector != nil && peer.PodSelector.MatchLabels["traffic.olares.io/platform-mesh"] == "true" &&
			len(peer.NamespaceSelector.MatchExpressions) == 1 {
			platformPeers++
		}
		if peer.NamespaceSelector.MatchLabels[security.NamespaceInClusterCallerLabel] == "true" {
			foundCaller = true
		}
	}
	if platformPeers != 2 {
		t.Fatal("platform mesh peers must be constrained by both namespace and pod")
	}
	if !foundCaller {
		t.Fatal("expected in-cluster-caller peer on os-mesh NP")
	}
	if len(meshNP.Spec.Ingress[0].Ports) != 5 {
		t.Fatalf("ports = %+v", meshNP.Spec.Ingress[0].Ports)
	}
	appIngress := meshNP.Spec.Ingress[1]
	if len(appIngress.From) != 1 || appIngress.From[0].PodSelector == nil ||
		appIngress.From[0].PodSelector.MatchLabels["linkerd.io/control-plane-ns"] != security.MeshControlPlaneNamespace {
		t.Fatalf("app ingress requires an enrolled proxy: %+v", appIngress.From)
	}
	selector, err := metav1.LabelSelectorAsSelector(appIngress.From[0].NamespaceSelector)
	if err != nil {
		t.Fatal(err)
	}
	if !selector.Matches(labels.Set{constants.ApplicationNameLabel: "photos", constants.ApplicationInstallUserLabel: "alice"}) ||
		selector.Matches(labels.Set{constants.ApplicationNameLabel: "photos"}) ||
		selector.Matches(labels.Set{constants.ApplicationInstallUserLabel: "alice"}) {
		t.Fatal("app ingress must require both installer ownership labels")
	}
	ports := map[int32]bool{8080: true, 8086: true, 8090: true}
	for _, port := range appIngress.Ports {
		if port.Port == nil || port.Protocol == nil || *port.Protocol != "TCP" || !ports[port.Port.IntVal] || port.EndPort != nil {
			t.Fatalf("unexpected app control-plane port: %+v", port)
		}
		delete(ports, port.Port.IntVal)
	}
	if len(ports) != 0 {
		t.Fatalf("missing app control-plane ports: %v", ports)
	}
	if !security.IsAppGatewayMeshManagedNP(meshNP) {
		t.Fatalf("os-mesh NP labels must be security-sweep exempt: %+v", meshNP.Labels)
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

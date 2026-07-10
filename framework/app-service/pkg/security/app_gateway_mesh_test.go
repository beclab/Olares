package security

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAppGatewayMeshNetworkPolicy_osMesh_includesPolicyPort8090(t *testing.T) {
	np := NewAppGatewayMeshNetworkPolicy("os-mesh", "os-gateway")
	require.Equal(t, AppGatewayMeshNPName, np.Name)
	require.Equal(t, "os-mesh", np.Namespace)
	require.Len(t, np.Spec.Ingress, 1)
	require.Len(t, np.Spec.Ingress[0].From, len(LinkerdControlPlaneIngressPeerNamespaces)+2)
	sharedPeer := np.Spec.Ingress[0].From[len(np.Spec.Ingress[0].From)-2]
	require.Equal(t, "true", sharedPeer.NamespaceSelector.MatchLabels[NamespaceSharedLabel])
	callerPeer := np.Spec.Ingress[0].From[len(np.Spec.Ingress[0].From)-1]
	require.Equal(t, "true", callerPeer.NamespaceSelector.MatchLabels[NamespaceInClusterCallerLabel])
	require.Len(t, np.Spec.Ingress[0].Ports, len(LinkerdMeshIngressPortsFromAppGateway))

	got := make([]int32, 0, len(np.Spec.Ingress[0].Ports))
	for _, p := range np.Spec.Ingress[0].Ports {
		require.NotNil(t, p.Port)
		got = append(got, p.Port.IntVal)
	}
	require.Equal(t, LinkerdMeshIngressPortsFromAppGateway, got)
	require.Contains(t, got, int32(8090))
}

func TestNewAppGatewayMeshNetworkPolicy_appGateway_noPortList(t *testing.T) {
	np := NewAppGatewayMeshNetworkPolicy("os-gateway", "os-mesh")
	require.Empty(t, np.Spec.Ingress[0].Ports, "app-gateway side allows all TCP from os-mesh")
}

func TestNewLinkerdMeshPrometheusScrapeNetworkPolicy(t *testing.T) {
	np := NewLinkerdMeshPrometheusScrapeNetworkPolicy("linkerd-viz")
	require.Equal(t, LinkerdMeshPrometheusScrapeNPName, np.Name)
	require.Equal(t, PlatformPrometheusNamespace,
		np.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"])
	require.Contains(t, LinkerdMeshPrometheusScrapePorts, int32(4191))
}

func TestAppGatewayMeshNamespaces_includeLinkerdViz(t *testing.T) {
	require.True(t, IsAppGatewayMeshNamespace("linkerd-viz"))
	require.Equal(t, "os-mesh", AppGatewayMeshPeerNamespace("linkerd-viz"))
}

func TestNewSharedLinkerdControlPlaneIngressNetworkPolicy(t *testing.T) {
	np := NewSharedLinkerdControlPlaneIngressNetworkPolicy("ollamaserver-shared", map[string]string{"app": "ollama"})
	require.Equal(t, SharedLinkerdMeshIngressNPName, np.Name)
	require.Equal(t, "ollamaserver-shared", np.Namespace)
	require.Equal(t, "ollama", np.Spec.PodSelector.MatchLabels["app"])

	require.Len(t, np.Spec.Ingress, 1)
	require.Len(t, np.Spec.Ingress[0].From, len(SharedLinkerdMeshIngressPeerNamespaces))
	gotPeers := make([]string, 0, len(np.Spec.Ingress[0].From))
	for _, peer := range np.Spec.Ingress[0].From {
		require.NotNil(t, peer.NamespaceSelector)
		gotPeers = append(gotPeers, peer.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"])
	}
	require.Equal(t, SharedLinkerdMeshIngressPeerNamespaces, gotPeers,
		"peers must include os-mesh (control-plane) and linkerd-viz (tap/metrics)")
}

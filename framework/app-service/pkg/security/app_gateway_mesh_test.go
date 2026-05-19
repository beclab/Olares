package security

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAppGatewayMeshNetworkPolicy_linkerd_includesPolicyPort8090(t *testing.T) {
	np := NewAppGatewayMeshNetworkPolicy("linkerd", "app-gateway")
	require.Equal(t, AppGatewayMeshNPName, np.Name)
	require.Equal(t, "linkerd", np.Namespace)
	require.Len(t, np.Spec.Ingress, 1)
	require.Len(t, np.Spec.Ingress[0].From, len(LinkerdControlPlaneIngressPeerNamespaces))
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
	np := NewAppGatewayMeshNetworkPolicy("app-gateway", "linkerd")
	require.Empty(t, np.Spec.Ingress[0].Ports, "app-gateway side allows all TCP from linkerd")
}

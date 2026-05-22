package routecontrol

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

func TestIsManagedNetworkPolicy(t *testing.T) {
	managedLabels := map[string]string{ManagedByLabel: ManagedByValue}

	cases := []struct {
		name string
		np   *networkingv1.NetworkPolicy
		want bool
	}{
		{
			name: "nil",
			np:   nil,
			want: false,
		},
		{
			name: "managed ingress NP",
			np: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: NetworkPolicyName, Labels: managedLabels},
			},
			want: true,
		},
		{
			name: "managed shared linkerd mesh NP",
			np: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: security.SharedLinkerdMeshIngressNPName, Labels: managedLabels},
			},
			want: true,
		},
		{
			name: "ingress NP missing managed-by label",
			np: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: NetworkPolicyName},
			},
			want: false,
		},
		{
			name: "other NP with managed-by label",
			np: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "user-defined-np", Labels: managedLabels},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := IsManagedNetworkPolicy(tc.np); got != tc.want {
				t.Fatalf("IsManagedNetworkPolicy(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

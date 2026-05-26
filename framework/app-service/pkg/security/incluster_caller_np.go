package security

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// Legacy caller-side egress NetworkPolicy constants.
//
// NP-minimal v1.0 (archdoc Shared集群内访问v1 §3.2) drops these managed egress
// policies entirely; CallerReconciler.cleanupCallerResources still references
// the names to GC them from clusters upgrading from earlier versions. The
// builder helpers below are retained only for unit-test / migration paths and
// MUST NOT be invoked by production reconcile.
const (
	CallerToAppGatewayEgressNPName = "caller-to-app-gateway-egress-np"
	CallerMeshEgressNPName         = "caller-mesh-egress-np"
	CallerDNSEgressNPName          = "caller-dns-egress-np"
	CallerMiddlewareEgressNPName   = "caller-middleware-egress-np"

	callerNPComponentLabel = "route-control"
	callerNPManagedBy      = "app-service"

	egOwningGatewayLabel     = "gateway.envoyproxy.io/owning-gateway-name"
	defaultOwningGatewayName = "app-gateway"
)

// NewCallerToAppGatewayEgressNP builds egress from every pod in ns to app-gateway
// data-plane pods on TCP 80 and 443.
func NewCallerToAppGatewayEgressNP(ns, gatewayNS string) *netv1.NetworkPolicy {
	if gatewayNS == "" {
		gatewayNS = "app-gateway"
	}
	tcp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolTCP)))
	ports := []netv1.NetworkPolicyPort{
		{Protocol: tcp, Port: intstrPtr(80)},
		{Protocol: tcp, Port: intstrPtr(443)},
	}
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CallerToAppGatewayEgressNPName,
			Namespace: ns,
			Labels:    callerNPLabels(),
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": gatewayNS,
								},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									egOwningGatewayLabel: defaultOwningGatewayName,
								},
							},
						},
					},
					Ports: ports,
				},
			},
		},
	}
}

// NewCallerDNSEgressNP builds egress from caller pods to kube-system DNS (CoreDNS).
func NewCallerDNSEgressNP(ns string) *netv1.NetworkPolicy {
	tcp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolTCP)))
	udp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolUDP)))
	ports := []netv1.NetworkPolicyPort{
		{Protocol: udp, Port: intstrPtr(53)},
		{Protocol: tcp, Port: intstrPtr(53)},
	}
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CallerDNSEgressNPName,
			Namespace: ns,
			Labels:    callerNPLabels(),
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "kube-system",
								},
							},
						},
					},
					Ports: ports,
				},
			},
		},
	}
}

// NewCallerMiddlewareEgressNP allows caller workloads to reach user-system middleware
// (Postgres/Citus, Authelia) required for app startup before in-cluster HTTP probes.
func NewCallerMiddlewareEgressNP(ns, userSystemNS string) *netv1.NetworkPolicy {
	if userSystemNS == "" {
		userSystemNS = "user-system"
	}
	tcp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolTCP)))
	// Match user-space-<viewer> and user-system-<viewer> via prefix on metadata.name.
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CallerMiddlewareEgressNPName,
			Namespace: ns,
			Labels:    callerNPLabels(),
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": userSystemNS,
								},
							},
						},
					},
					Ports: []netv1.NetworkPolicyPort{
						{Protocol: tcp, Port: intstrPtr(5432)},
						{Protocol: tcp, Port: intstrPtr(9091)},
					},
				},
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "os-platform",
								},
							},
						},
					},
					Ports: []netv1.NetworkPolicyPort{
						{Protocol: tcp, Port: intstrPtr(5432)},
					},
				},
			},
		},
	}
}

// NewCallerMeshEgressNP builds egress from caller pods to linkerd control-plane ports.
func NewCallerMeshEgressNP(ns string) *netv1.NetworkPolicy {
	tcp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolTCP)))
	meshPorts := []int32{8080, 8086, 8090}
	ports := make([]netv1.NetworkPolicyPort, 0, len(meshPorts))
	for _, p := range meshPorts {
		ports = append(ports, netv1.NetworkPolicyPort{Protocol: tcp, Port: intstrPtr(p)})
	}
	peers := []netv1.NetworkPolicyPeer{
		{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "linkerd",
				},
			},
		},
		{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "linkerd-viz",
				},
			},
		},
	}
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CallerMeshEgressNPName,
			Namespace: ns,
			Labels:    callerNPLabels(),
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{
				{To: peers, Ports: ports},
			},
		},
	}
}

func callerNPLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": callerNPManagedBy,
		"app.kubernetes.io/component":  callerNPComponentLabel,
	}
}

package security

import (
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NamespaceTypeLabel       = "bytetrade.io/ns-type"
	NamespaceOwnerLabel      = "bytetrade.io/ns-owner"
	NamespaceSharedLabel     = "bytetrade.io/ns-shared"
	NamespaceMiddlewareLabel = "bytetrade.io/ns-middleware"
	System                   = "system"
	Network                  = "network"
	Internal                 = "user-internal"
	Protected                = "protected"
	UserSpace                = "user-space"
)

var (
	// NPDenyAll is a network policy template deny all ingress.
	NPDenyAll = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress:     []netv1.NetworkPolicyIngressRule{},
		},
	} // end NPDenyAll

	// NPAllowAll is a network policy template allow all ingress.
	NPAllowAll = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
		},
	} // end NPAllowAll

	// NPUnderLayerSystem is a network policy template for under layer system.
	NPUnderLayerSystem = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "k8s-app",
						Operator: metav1.LabelSelectorOpNotIn,
						Values: []string{
							"kube-dns",
						},
					},
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Protected,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubesphere.io/namespace": "kubesphere-system",
								},
							},
						},
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"tier": "bfl",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: UserSpace,
								},
							},
						},
					},
				},
			},
		},
	} // end NPUnderLayerSystem

	// NPOSSystem is a network policy template for os-framework and os-platform.
	NPOSSystem = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Protected,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubesphere.io/namespace": "kube-system",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubesphere.io/namespace": "kubesphere-system",
								},
							},
						},
					},
				},
			},
		},
	} // end NPOSSystem

	// NPOSNetwork is a network policy template for os-network.
	NPOSNetwork = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubesphere.io/namespace": "kube-system",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubesphere.io/namespace": "kubesphere-system",
								},
							},
						},
					},
				},
			},
		},
	} // end NPOSNetwork

	// NPOSProtected is a network policy template for os-protected.
	NPOSProtected = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "tapr-sysevent",
								},
							},
						},
					},
				},
			},
		},
	} // end NPOSProtected

	// NPUserSystem is a network policy template for user-system namespace.
	NPUserSystem = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceOwnerLabel: "",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceSharedLabel: "true",
								},
							},
						},
					},
				},
			},
		},
	} // end NPUserSystem

	// NPUserSpace is a network policy template for user space namespace.
	NPUserSpace = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceOwnerLabel: "",
									NamespaceTypeLabel:  Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Network,
								},
							},
						},
					},
				},
				{
					From: []netv1.NetworkPolicyPeer{
						{
							// allow external traffic via nodeport
							IPBlock: &netv1.IPBlock{
								CIDR:   "0.0.0.0/0",
								Except: []string{"10.233.0.0/16"},
							},
						},
					},
				},
			},
		},
	} // end NPUserSpace

	// NPAppSpace is a network policy template for application.
	NPAppSpace = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceOwnerLabel: "",
									NamespaceTypeLabel:  Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"tier": "bfl",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceOwnerLabel: "",
									NamespaceTypeLabel:  UserSpace,
								},
							},
						},
					},
				},
			},
		},
	} // end NPAppSpace

	// NPSharedSpace is a network policy template for shared app namespace.
	NPSharedSpace = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: Internal,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
					},
				},
			},
		},
	} // end NPSharedSpace

	NSFilesPolicy = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "files-np",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "files",
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"tier": "bfl",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: UserSpace,
								},
							},
						},
					},
				},
			},
		},
	} // end NSFilesPolicy

	NPSystemProvider = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-provider-np",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"provider": "true",
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceSharedLabel: "true",
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceTypeLabel: System,
								},
							},
						},
					},
				},
			},
		},
	} // end NPOSSystemProvider

	// allow cluster scoped applications in shared namespace to access system middleware services
	NPSystemMiddleware = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-middleware-np",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.AppMiddlewareLabel: "true",
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									NamespaceSharedLabel: "true",
								},
							},
						},
					},
				},
			},
		},
	} // end NPSystemMiddleware

	NPSharedEntrance = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared-entrance-np",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.AppSharedEntrancesLabel: "true",
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{},
				// {
				// 	From: []netv1.NetworkPolicyPeer{
				// 		{
				// 			NamespaceSelector: &metav1.LabelSelector{
				// 				MatchLabels: map[string]string{
				// 					NamespaceSharedLabel: "true",
				// 				},
				// 			},
				// 		},
				// 	},
				// },
			},
		},
	} // end NPSharedEntrance

	NPIngress = netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-np",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"tier": "bfl",
				},
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "10.233.0.0/16",
							},
						},
					},
				},
			},
		},
	} // end NPIngress
)

// NodeTunnelRule returns node tunnel rule.
func NodeTunnelRule() []netv1.NetworkPolicyPeer {
	var rules []netv1.NetworkPolicyPeer

	for _, cidr := range utils.GetAllNodesTunnelIPCIDRs() {
		rules = append(rules, netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: cidr,
			},
		})
	}

	return rules
}

type NetworkPolicies []*netv1.NetworkPolicy

func (n NetworkPolicies) Main() *netv1.NetworkPolicy {
	if len(n) > 0 {
		return n[0]
	}
	return nil
}

func (n NetworkPolicies) Additional() []*netv1.NetworkPolicy {
	if len(n) > 1 {
		return n[1:]
	}
	return nil
}

func (n NetworkPolicies) Contains(np *netv1.NetworkPolicy) bool {
	for _, item := range n {
		if item.Name == np.Name && item.Namespace == np.Namespace {
			return true
		}
	}

	return false
}

func (n NetworkPolicies) Name() string {
	if main := n.Main(); main != nil {
		return main.Name
	}
	return ""
}

func (n NetworkPolicies) SetName(name string) {
	if main := n.Main(); main != nil {
		main.Name = name
	}
}

func (n NetworkPolicies) SetNamespace(namespace string) {
	for _, np := range n {
		np.Namespace = namespace
	}
}

func (n NetworkPolicies) Namespace() string {
	if main := n.Main(); main != nil {
		return main.Namespace
	}
	return ""
}

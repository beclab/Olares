package access

import (
	"context"
	"fmt"
	"maps"
	"net/netip"

	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type accessTransportPeer struct {
	namespace string
	selector  map[string]string
}

func accessTransportSelector(selector map[string]string) metav1.LabelSelector {
	match := maps.Clone(selector)
	match["linkerd.io/control-plane-ns"] = mesh.LinkerdNamespace
	return metav1.LabelSelector{MatchLabels: match}
}

func (peer accessTransportPeer) policyPeer() netv1.NetworkPolicyPeer {
	return netv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": peer.namespace}},
		PodSelector:       ptr.To(accessTransportSelector(peer.selector)),
	}
}

type accessTransportLeg struct {
	name        string
	from, to    accessTransportPeer
	service     string
	servicePort int32
	backendPort int32
}

func (b boundPolicy) transportLegs() []accessTransportLeg {
	source := accessTransportPeer{b.policy.SourceSANamespace, b.sourceSelector}
	target := accessTransportPeer{b.policy.Namespace, b.policy.TargetSelector}
	port, _, _ := unstructured.NestedInt64(b.target, "service", "servicePort")
	backend := accessTransportLeg{
		name: "target", from: source, to: target, service: bindingRef(b.target, "service").Name,
		servicePort: int32(port), backendPort: b.policy.ServerPort,
	}
	if b.policy.RouteMode == RouteModeDirect {
		return []accessTransportLeg{backend}
	}
	gateway := accessTransportPeer{internalNamespace, internalSelector()}
	backend.from = gateway
	return []accessTransportLeg{
		{name: "gateway", from: source, to: gateway, service: internalGateway, servicePort: internalPort, backendPort: internalPort},
		backend,
	}
}

// The access coordinator writes these transport permissions only after the
// Linkerd deny Servers and HTTPRoute bindings are ready.
func (b boundPolicy) reconcileTransport(ctx context.Context, c client.Client, withdraw bool) error {
	var objects []DesiredObject
	for _, leg := range b.transportLegs() {
		if !withdraw && (len(leg.from.selector) == 0 || len(leg.to.selector) == 0 ||
			leg.backendPort < 1 || leg.backendPort > 65535 || leg.servicePort < 1 || leg.servicePort > 65535) {
			return fmt.Errorf("invalid mesh transport binding")
		}
		out := b.transportPolicy(leg.from.namespace, leg.name+"-egress")
		in := b.transportPolicy(leg.to.namespace, leg.name+"-ingress")
		if !withdraw {
			ports := []netv1.NetworkPolicyPort{
				{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt32(leg.backendPort))},
				{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt32(4143))},
			}
			in.Spec = netv1.NetworkPolicySpec{
				PodSelector: accessTransportSelector(leg.to.selector),
				PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
				Ingress:     []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{leg.from.policyPeer()}, Ports: ports}},
			}
			out.Spec = netv1.NetworkPolicySpec{
				PodSelector: accessTransportSelector(leg.from.selector),
				PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
				Egress:      []netv1.NetworkPolicyEgressRule{{To: []netv1.NetworkPolicyPeer{leg.to.policyPeer()}, Ports: ports}},
			}
			var service corev1.Service
			if err := c.Get(ctx, client.ObjectKey{Namespace: leg.to.namespace, Name: leg.service}, &service); err != nil {
				return err
			}
			expected := maps.Clone(leg.to.selector)
			if leg.name == "gateway" && service.Spec.Selector["app.kubernetes.io/name"] == "envoy" {
				expected["app.kubernetes.io/name"] = "envoy"
			}
			if !maps.Equal(service.Spec.Selector, expected) || !service.DeletionTimestamp.IsZero() ||
				service.Spec.Type == corev1.ServiceTypeExternalName ||
				(leg.name == "target" && string(service.UID) != bindingRef(b.target, "service").UID) {
				return fmt.Errorf("mesh transport Service %s/%s changed since binding", service.Namespace, service.Name)
			}
			if leg.name == "target" {
				workload, err := boundWorkload(ctx, c, b.target)
				if err != nil {
					return err
				}
				obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&service)
				if err != nil {
					return err
				}
				if _, err := boundPort(b.target, obj, workload.Object, b.targetPortName); err != nil {
					return err
				}
			}
			addresses := service.Spec.ClusterIPs
			if len(addresses) == 0 {
				addresses = []string{service.Spec.ClusterIP}
			}
			for _, address := range addresses {
				if ip, err := netip.ParseAddr(address); err == nil && ip.IsGlobalUnicast() {
					ip = ip.Unmap()
					out.Spec.Egress = append(out.Spec.Egress, netv1.NetworkPolicyEgressRule{
						To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: netip.PrefixFrom(ip, ip.BitLen()).String()}}},
						Ports: []netv1.NetworkPolicyPort{{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt32(leg.servicePort))}},
					})
				}
			}
		}
		for _, policy := range []*netv1.NetworkPolicy{&in, &out} {
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(policy)
			if err != nil {
				return err
			}
			u := &unstructured.Unstructured{Object: obj}
			objects = append(objects, DesiredObject{GVK: u.GroupVersionKind(), Name: policy.Name, Namespace: policy.Namespace, Object: u, Delete: withdraw})
		}
	}
	return applyProjections(ctx, c, objects)
}

func (b boundPolicy) transportPolicy(namespace, suffix string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{APIVersion: netv1.SchemeGroupVersion.String(), Kind: "NetworkPolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name: b.policy.ProjectionID + "-transport-" + suffix, Namespace: namespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByAccessPolicy, PolicyNameLabel: projectionOwner(b.policy),
				security.RouteControlComponentLabel: security.AccessTransportComponentValue,
			},
		},
	}
}

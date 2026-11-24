package routecontrol

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const routeTransportLabel = "gateway.olares.io/transport-route"

func transportLabels(key client.ObjectKey) map[string]string {
	digest := sha256.Sum256([]byte(key.Namespace + "/" + key.Name))
	return map[string]string{
		ManagedByLabel: ManagedByValue, RouteControlComponentLabel: RouteControlComponentValue,
		routeTransportLabel: fmt.Sprintf("%x", digest[:16]),
	}
}

type meshTransportPeer struct {
	namespace string
	selector  map[string]string
}

func (p meshTransportPeer) networkPolicyPeer() netv1.NetworkPolicyPeer {
	return netv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": p.namespace}},
		PodSelector:       &metav1.LabelSelector{MatchLabels: p.selector},
	}
}

type meshTransportBackend struct {
	peer        meshTransportPeer
	targetPorts []netv1.NetworkPolicyPort
	serviceIPs  []netv1.NetworkPolicyPeer
	servicePort netv1.NetworkPolicyPort
}

func meshTransportPort(port int32) netv1.NetworkPolicyPort {
	return netv1.NetworkPolicyPort{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt32(port))}
}

func transportBackend(ctx context.Context, c client.Client, svc *corev1.Service, port int32) (meshTransportBackend, error) {
	backend := meshTransportBackend{
		peer:        meshTransportPeer{namespace: svc.Namespace, selector: maps.Clone(svc.Spec.Selector)},
		servicePort: meshTransportPort(port),
	}
	if len(backend.peer.selector) == 0 {
		return backend, fmt.Errorf("transport backend %s/%s has no Pod selector", svc.Namespace, svc.Name)
	}
	if ns := backend.peer.selector["linkerd.io/control-plane-ns"]; ns != "" && ns != mesh.LinkerdNamespace {
		return backend, fmt.Errorf("transport backend %s/%s selects a different mesh", svc.Namespace, svc.Name)
	}
	backend.peer.selector["linkerd.io/control-plane-ns"] = mesh.LinkerdNamespace
	var portName string
	var portFound bool
	for _, servicePort := range svc.Spec.Ports {
		if servicePort.Port == port && (servicePort.Protocol == "" || servicePort.Protocol == corev1.ProtocolTCP) {
			portName = servicePort.Name
			portFound = true
			break
		}
	}
	if !portFound {
		return backend, fmt.Errorf("transport backend %s/%s has no bound TCP Service port", svc.Namespace, svc.Name)
	}
	var endpoints discoveryv1.EndpointSliceList
	if err := c.List(ctx, &endpoints, client.InNamespace(svc.Namespace),
		client.MatchingLabels{discoveryv1.LabelServiceName: svc.Name}); err != nil {
		return backend, err
	}
	var ports []int32
	for _, slice := range endpoints.Items {
		owner := metav1.GetControllerOf(&slice)
		if owner == nil || owner.Kind != "Service" || owner.Name != svc.Name || owner.UID != svc.UID {
			continue
		}
		if !hasInterceptedHTTPPort(slice.Ports, portName, 0) ||
			!slices.ContainsFunc(slice.Endpoints, func(endpoint discoveryv1.Endpoint) bool {
				return ptr.Deref(endpoint.Conditions.Ready, false) && !ptr.Deref(endpoint.Conditions.Terminating, false)
			}) {
			continue
		}
		for _, endpointPort := range slice.Ports {
			if ptr.Deref(endpointPort.Name, "") == portName &&
				ptr.Deref(endpointPort.Protocol, corev1.ProtocolTCP) == corev1.ProtocolTCP &&
				endpointPort.Port != nil && *endpointPort.Port > 0 {
				ports = append(ports, *endpointPort.Port)
			}
		}
	}
	if len(ports) == 0 {
		return backend, fmt.Errorf("transport backend %s/%s has no intercepted target port", svc.Namespace, svc.Name)
	}
	// Depending on CNI hook ordering, the peer's inbound proxy port may be
	// visible instead of the intercepted backend port.
	ports = append(ports, 4143)
	slices.Sort(ports)
	for _, target := range slices.Compact(ports) {
		backend.targetPorts = append(backend.targetPorts, meshTransportPort(target))
	}
	addresses := svc.Spec.ClusterIPs
	if len(addresses) == 0 {
		addresses = []string{svc.Spec.ClusterIP}
	}
	for _, address := range addresses {
		if ip, err := netip.ParseAddr(address); err == nil && ip.IsGlobalUnicast() && ip.Zone() == "" {
			ip = ip.Unmap()
			backend.serviceIPs = append(backend.serviceIPs, netv1.NetworkPolicyPeer{
				IPBlock: &netv1.IPBlock{CIDR: netip.PrefixFrom(ip, ip.BitLen()).String()},
			})
		}
	}
	return backend, nil
}

func transportPair(labels map[string]string, leg string, source meshTransportPeer, backend meshTransportBackend) []netv1.NetworkPolicy {
	base := "olares-route-" + labels[routeTransportLabel] + "-" + leg
	out := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: base + "-egress", Namespace: source.namespace, Labels: labels},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: source.selector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress:      []netv1.NetworkPolicyEgressRule{{To: []netv1.NetworkPolicyPeer{backend.peer.networkPolicyPeer()}, Ports: backend.targetPorts}},
		},
	}
	if len(backend.serviceIPs) > 0 {
		out.Spec.Egress = append(out.Spec.Egress, netv1.NetworkPolicyEgressRule{To: backend.serviceIPs, Ports: []netv1.NetworkPolicyPort{backend.servicePort}})
	}
	if tenant, ok := strings.CutPrefix(source.namespace, "user-system-"); ok && tenant != "" &&
		source.selector["app"] == "systemserver" && source.selector["linkerd.io/control-plane-ns"] == mesh.LinkerdNamespace {
		// Any Egress policy isolates the Pod, even before traffic-group activation.
		// Keep the coordinator reachable in the same write that enables route egress.
		out.Spec.Egress = append(out.Spec.Egress, security.SystemServerCitusEgressRule())
	}
	in := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: base + "-ingress", Namespace: backend.peer.namespace, Labels: labels},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: backend.peer.selector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress:     []netv1.NetworkPolicyIngressRule{{From: []netv1.NetworkPolicyPeer{source.networkPolicyPeer()}, Ports: backend.targetPorts}},
		},
	}
	return []netv1.NetworkPolicy{out, in}
}

// systemLoginTransport runs before the frontend readiness check. The login
// frontend needs this hop to finish its own HTTP health check.
func systemLoginTransport(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) ([]netv1.NetworkPolicy, error) {
	if !isSystemLoginRoute(srr) {
		return nil, nil
	}
	authNS := autheliaNSForSRR(srr)
	services := []struct {
		namespace string
		selector  map[string]string
	}{
		{authNS, map[string]string{"app": "systemserver"}},
		{"os-framework", map[string]string{"app": "authelia-backend"}},
	}
	var backends []meshTransportBackend
	for _, target := range services {
		ref := srrv1alpha1.UpstreamRef{ServiceName: AutheliaBackendServiceName, Port: AutheliaBackendPort}
		var svc corev1.Service
		if err := c.Get(ctx, client.ObjectKey{Namespace: target.namespace, Name: ref.ServiceName}, &svc); err != nil {
			return nil, err
		}
		if svc.Spec.Type != corev1.ServiceTypeClusterIP || !maps.Equal(svc.Spec.Selector, target.selector) {
			return nil, fmt.Errorf("login dependency %s/%s has an unexpected Service configuration", svc.Namespace, svc.Name)
		}
		if err := meshedReadyEndpoints(ctx, c, target.namespace, ref); err != nil {
			return nil, err
		}
		backend, err := transportBackend(ctx, c, &svc, ref.Port)
		if err != nil {
			return nil, err
		}
		if !slices.EqualFunc(backend.targetPorts, []netv1.NetworkPolicyPort{meshTransportPort(4143), meshTransportPort(AutheliaBackendPort)}, func(a, b netv1.NetworkPolicyPort) bool {
			return *a.Port == *b.Port
		}) {
			return nil, fmt.Errorf("login dependency %s/%s has an unexpected target port", svc.Namespace, svc.Name)
		}
		backends = append(backends, backend)
	}
	return transportPair(transportLabels(client.ObjectKeyFromObject(srr)), "login-backend", backends[0].peer, backends[1]), nil
}

func applyMeshTransport(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, service *corev1.Service, port int32, bootstrap map[client.ObjectKey]bool) error {
	key := client.ObjectKeyFromObject(srr)
	labels := transportLabels(key)
	gatewayPeer := meshTransportPeer{namespace: gw.gatewayNamespace(), selector: map[string]string{
		"gateway.envoyproxy.io/owning-gateway-name": gw.gatewayName(), "gateway.envoyproxy.io/owning-gateway-namespace": gw.gatewayNamespace(),
		"app.kubernetes.io/component": "proxy", "app.kubernetes.io/managed-by": "envoy-gateway", "app.kubernetes.io/name": "envoy",
		"linkerd.io/control-plane-ns": mesh.LinkerdNamespace,
	}}
	backend, err := transportBackend(ctx, c, service, port)
	if err != nil {
		return err
	}
	policies := transportPair(labels, "http", gatewayPeer, backend)
	if isApplicationSRR(srr) && srrAuthMode(srr) != gateway.AuthModeNone {
		auth := make([]meshTransportBackend, 0, 2)
		for _, namespace := range []string{autheliaNSForSRR(srr), "os-framework"} {
			var svc corev1.Service
			if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: AutheliaBackendServiceName}, &svc); err != nil {
				return err
			}
			b, err := transportBackend(ctx, c, &svc, AutheliaBackendPort)
			if err != nil {
				return err
			}
			auth = append(auth, b)
		}
		policies = append(policies, transportPair(labels, "auth-proxy", gatewayPeer, auth[0])...)
		policies = append(policies, transportPair(labels, "auth-server", auth[0].peer, auth[1])...)
	}

	keep, err := upsertMeshTransports(ctx, c, key, policies)
	if err != nil {
		return err
	}
	maps.Copy(keep, bootstrap)
	return pruneMeshTransport(ctx, c, key, keep)
}

func upsertMeshTransports(ctx context.Context, c client.Client, key client.ObjectKey, policies []netv1.NetworkPolicy) (map[client.ObjectKey]bool, error) {
	labels := transportLabels(key)
	keep := make(map[client.ObjectKey]bool, len(policies))
	for _, policy := range policies {
		policyKey := client.ObjectKeyFromObject(&policy)
		var current netv1.NetworkPolicy
		err := c.Get(ctx, policyKey, &current)
		switch {
		case apierrors.IsNotFound(err):
			err = c.Create(ctx, &policy)
		case err == nil:
			if !ownsTransportPolicy(&current, labels) {
				return nil, fmt.Errorf("transport NetworkPolicy %s has another owner", policyKey)
			}
			if !apiequality.Semantic.DeepEqual(current.Spec, policy.Spec) {
				current.Spec = policy.Spec
				err = c.Update(ctx, &current)
			}
		}
		if err != nil {
			return nil, err
		}
		keep[policyKey] = true
	}
	return keep, nil
}

func ownsTransportPolicy(policy *netv1.NetworkPolicy, labels map[string]string) bool {
	for key, value := range labels {
		if policy.Labels[key] != value {
			return false
		}
	}
	return true
}

func pruneMeshTransport(ctx context.Context, c client.Client, key client.ObjectKey, keep map[client.ObjectKey]bool) error {
	var policies netv1.NetworkPolicyList
	if err := c.List(ctx, &policies, client.MatchingLabels(transportLabels(key))); err != nil {
		return err
	}
	for _, policy := range policies.Items {
		if !keep[client.ObjectKeyFromObject(&policy)] {
			if err := c.Delete(ctx, &policy, client.Preconditions{UID: &policy.UID, ResourceVersion: &policy.ResourceVersion}); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}
	return nil
}

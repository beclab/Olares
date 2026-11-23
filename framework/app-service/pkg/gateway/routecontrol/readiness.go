package routecontrol

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

const ReasonAwaitingDataPlane = "AwaitingDataPlane"

var entryMeshService = srrv1alpha1.UpstreamRef{ServiceName: "app-gateway-l4-mtls", Port: 8443}

// servingReadiness distinguishes resources accepted by Kubernetes from an
// observed, programmed listener with meshed backend endpoints.
func servingReadiness(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, routeName string) error {
	gateway := &unstructured.Unstructured{}
	gateway.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(ctx, client.ObjectKey{Namespace: gw.gatewayNamespace(), Name: gw.gatewayName()}, gateway); err != nil {
		return err
	}
	conditions, _, _ := unstructured.NestedSlice(gateway.Object, "status", "conditions")
	if !conditionReady(conditions, gateway.GetGeneration(), "Programmed") {
		return fmt.Errorf("gateway is not programmed for its current generation")
	}
	if !gatewayListenerReady(gateway, l4MTLSGatewaySectN) {
		return fmt.Errorf("entry mTLS listener is not programmed")
	}
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(ctx, client.ObjectKey{Namespace: srr.Namespace, Name: routeName}, route); err != nil {
		return err
	}
	if !entryParentReady(route, gw) {
		return fmt.Errorf("HTTPRoute not accepted on the entry mTLS listener for its current generation")
	}

	if srrAuthMode(srr) != "none" {
		if err := gatewayPolicyReady(ctx, c, securityPolicyGVK, srr.Namespace, securityPolicyName(srr), gw); err != nil {
			return err
		}
	}
	validation := schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"}
	if err := gatewayPolicyReady(ctx, c, validation, gw.gatewayNamespace(), "app-gateway-l4-entry-mtls", gw); err != nil {
		return err
	}

	if needsL4Authz(srr) {
		if !gatewayListenerReady(gateway, l4AuthzSection) {
			return fmt.Errorf("L4 authorization listener is not programmed")
		}
		if err := gatewayPolicyReadyForSection(ctx, c, validation, gw.gatewayNamespace(), "app-gateway-l4-authz-mtls", gw, l4AuthzSection); err != nil {
			return err
		}
		check := l4AuthzRoute(srr)
		if err := c.Get(ctx, client.ObjectKeyFromObject(check), check); err != nil {
			return err
		}
		if !entryParentReadyForSection(check, gw, l4AuthzSection) {
			return fmt.Errorf("authorization route is not accepted for its current generation")
		}
		service := entryMeshService
		service.Port = l4AuthzPort
		if err := meshedReadyEndpoints(ctx, c, gw.gatewayNamespace(), service); err != nil {
			return err
		}
	}

	backendNS := srr.Spec.Upstream.ServiceNamespace
	if backendNS == "" {
		backendNS = srr.Namespace
	}
	if err := meshedReadyEndpoints(ctx, c, gw.gatewayNamespace(), entryMeshService); err != nil {
		return err
	}
	return httpRouteMeshDependenciesReady(ctx, c, backendNS, srr)
}

func gatewayListenerReady(gateway *unstructured.Unstructured, section string) bool {
	listeners, _, _ := unstructured.NestedSlice(gateway.Object, "status", "listeners")
	for _, raw := range listeners {
		listener, _ := raw.(map[string]any)
		conditions, _, _ := unstructured.NestedSlice(listener, "conditions")
		if listener["name"] == section &&
			conditionReady(conditions, gateway.GetGeneration(), "Programmed") &&
			conditionReady(conditions, gateway.GetGeneration(), "ResolvedRefs") {
			return true
		}
	}
	return false
}

// Gate the entire ExtAuth chain before publishing a protected route. The
// per-user Service selects the system-server proxy, which forwards to Authelia
// in os-framework. Both legs must have enforced mesh endpoints.
func httpRouteMeshDependenciesReady(ctx context.Context, c client.Client, backendNS string, srr *srrv1alpha1.SharedRouteRegistry) error {
	if err := meshedReadyEndpoints(ctx, c, backendNS, srr.Spec.Upstream); err != nil {
		return err
	}
	if !isSystemLoginRoute(srr) && (!isApplicationSRR(srr) || srrAuthMode(srr) == gateway.AuthModeNone) {
		return nil
	}
	authNS := autheliaNSForSRR(srr)
	if authNS == "" {
		return fmt.Errorf("Authelia ExtAuth has no bound user namespace")
	}
	ref := srrv1alpha1.UpstreamRef{ServiceName: AutheliaBackendServiceName, Port: AutheliaBackendPort}
	if err := meshedReadyEndpoints(ctx, c, authNS, ref); err != nil {
		return fmt.Errorf("Authelia proxy: %w", err)
	}
	if err := meshedReadyEndpoints(ctx, c, "os-framework", ref); err != nil {
		return fmt.Errorf("Authelia server: %w", err)
	}
	return nil
}

func gatewayPolicyReady(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ns, name string, gw GatewayRef) error {
	return gatewayPolicyReadyForSection(ctx, c, gvk, ns, name, gw, l4MTLSGatewaySectN)
}

func gatewayPolicyReadyForSection(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ns, name string, gw GatewayRef, section string) error {
	policy := &unstructured.Unstructured{}
	policy.SetGroupVersionKind(gvk)
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, policy); err != nil {
		return err
	}
	if gvk.Kind == "ClientTrafficPolicy" && !validEntryClientPolicy(policy, gw, section) {
		return fmt.Errorf("ClientTrafficPolicy %s does not enforce the expected Entry client identity", name)
	}
	ancestors, _, _ := unstructured.NestedSlice(policy.Object, "status", "ancestors")
	for _, raw := range ancestors {
		ancestor, _ := raw.(map[string]any)
		ref, _, _ := unstructured.NestedMap(ancestor, "ancestorRef")
		if ancestor["controllerName"] != "gateway.envoyproxy.io/gatewayclass-controller" || ref["name"] != gw.gatewayName() ||
			(ref["sectionName"] != nil && ref["sectionName"] != section) {
			continue
		}
		n, _, _ := unstructured.NestedString(ref, "namespace")
		if n == "" {
			n = ns
		}
		if n != gw.gatewayNamespace() {
			continue
		}
		conditions, _, _ := unstructured.NestedSlice(ancestor, "conditions")
		if conditionReady(conditions, policy.GetGeneration(), "Accepted") {
			return nil
		}
	}
	return fmt.Errorf("%s %s has no accepted status for this gateway", gvk.Kind, name)
}

func validEntryClientPolicy(policy *unstructured.Unstructured, gw GatewayRef, section string) bool {
	ref, _, _ := unstructured.NestedMap(policy.Object, "spec", "targetRef")
	if ref["group"] != "gateway.networking.k8s.io" || ref["kind"] != "Gateway" ||
		ref["name"] != gw.gatewayName() || ref["sectionName"] != section {
		return false
	}
	validation, _, _ := unstructured.NestedMap(policy.Object, "spec", "tls", "clientValidation")
	if validation["mode"] != "RequireAndVerify" {
		return false
	}
	cas, _, _ := unstructured.NestedSlice(validation, "caCertificateRefs")
	names, _, _ := unstructured.NestedMap(validation, "subjectAltNames")
	dnsNames, _, _ := unstructured.NestedSlice(names, "dnsNames")
	if len(cas) != 1 || len(names) != 1 || len(dnsNames) != 1 {
		return false
	}
	ca, _ := cas[0].(map[string]any)
	dns, _ := dnsNames[0].(map[string]any)
	return ca["kind"] == "Secret" && ca["name"] == "olares-l4-agw-mtls" &&
		(ca["group"] == nil || ca["group"] == "") &&
		(ca["namespace"] == nil || ca["namespace"] == gw.gatewayNamespace()) &&
		dns["type"] == "Exact" && dns["value"] == "l4-bfl-proxy.entry.client"
}

func entryParentReady(route *unstructured.Unstructured, gw GatewayRef) bool {
	return entryParentReadyForSection(route, gw, l4MTLSGatewaySectN)
}

func entryParentReadyForSection(route *unstructured.Unstructured, gw GatewayRef, section string) bool {
	parents, _, _ := unstructured.NestedSlice(route.Object, "status", "parents")
	for _, raw := range parents {
		parent, _ := raw.(map[string]any)
		ref, _, _ := unstructured.NestedMap(parent, "parentRef")
		if parent["controllerName"] != "gateway.envoyproxy.io/gatewayclass-controller" ||
			ref["name"] != gw.gatewayName() || ref["sectionName"] != section {
			continue
		}
		ns, _, _ := unstructured.NestedString(ref, "namespace")
		if ns == "" {
			ns = route.GetNamespace()
		}
		if ns != gw.gatewayNamespace() {
			continue
		}
		conditions, _, _ := unstructured.NestedSlice(parent, "conditions")
		if conditionReady(conditions, route.GetGeneration(), "Accepted") &&
			conditionReady(conditions, route.GetGeneration(), "ResolvedRefs") {
			return true
		}
	}
	return false
}

func conditionReady(conditions []any, generation int64, kind string) bool {
	for _, raw := range conditions {
		cond, _ := raw.(map[string]any)
		observed, present, _ := unstructured.NestedInt64(cond, "observedGeneration")
		if present && observed == generation && cond["type"] == kind && cond["status"] == "True" {
			return true
		}
	}
	return false
}

func meshedReadyEndpoints(ctx context.Context, c client.Client, ns string, upstream srrv1alpha1.UpstreamRef) error {
	serviceName := upstream.ServiceName
	var service corev1.Service
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: serviceName}, &service); err != nil {
		return err
	}
	if service.UID == "" || len(service.Spec.Selector) == 0 {
		return fmt.Errorf("%s/%s lacks an owned, selector-backed Service", ns, serviceName)
	}
	port, err := resolveServicePort(&service, upstream)
	if err != nil {
		return err
	}
	var portName string
	for _, servicePort := range service.Spec.Ports {
		if servicePort.Port == port && (servicePort.Protocol == "" || servicePort.Protocol == corev1.ProtocolTCP) {
			portName = servicePort.Name
			break
		}
	}
	var entryPort int32
	if ns == defaultGatewayNS && serviceName == entryMeshService.ServiceName {
		entryPort = upstream.Port
	}
	selector := labels.SelectorFromSet(service.Spec.Selector)
	var endpointSlices discoveryv1.EndpointSliceList
	if err := c.List(ctx, &endpointSlices, client.InNamespace(ns), client.MatchingLabels{discoveryv1.LabelServiceName: serviceName}); err != nil {
		return err
	}
	found := false
	for _, sl := range endpointSlices.Items {
		if owner := metav1.GetControllerOf(&sl); owner == nil || owner.Kind != "Service" ||
			owner.Name != service.Name || owner.UID != service.UID {
			return fmt.Errorf("%s/%s has an EndpointSlice with an unexpected owner", ns, serviceName)
		}
		for _, endpoint := range sl.Endpoints {
			if endpoint.Conditions.Ready != nil && !*endpoint.Conditions.Ready {
				continue
			}
			if !hasInterceptedHTTPPort(sl.Ports, portName, entryPort) {
				return fmt.Errorf("%s/%s has a serving endpoint without an intercepted HTTP target port", ns, serviceName)
			}
			ref := endpoint.TargetRef
			if endpoint.Conditions.Ready == nil || ref == nil || ref.Kind != "Pod" || ref.UID == "" || (ref.Namespace != "" && ref.Namespace != ns) {
				return fmt.Errorf("%s/%s has an endpoint without a bound Pod", ns, serviceName)
			}
			var pod corev1.Pod
			if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: ref.Name}, &pod); err != nil {
				return err
			}
			if pod.UID != ref.UID || !selector.Matches(labels.Set(pod.Labels)) || !pod.DeletionTimestamp.IsZero() ||
				mesh.CheckHTTPMeshInterception(&pod) != nil || !mesh.LinkerdProxyReady(&pod) ||
				pod.Annotations[mesh.LinkerdDefaultInboundPolicyAnnotation] != mesh.InboundPolicyClusterAuthenticated ||
				!mesh.HasHTTPMeshInit(&pod) ||
				!mesh.ProxyHasInboundPolicy(&pod, mesh.InboundPolicyClusterAuthenticated) {
				return fmt.Errorf("%s/%s has a serving endpoint without an enforced Linkerd mesh posture", ns, serviceName)
			}
			if len(endpoint.Addresses) == 0 {
				return fmt.Errorf("%s/%s has a serving endpoint without an address", ns, serviceName)
			}
			for _, address := range endpoint.Addresses {
				if address != pod.Status.PodIP && !slices.Contains(pod.Status.PodIPs, corev1.PodIP{IP: address}) {
					return fmt.Errorf("%s/%s has a serving endpoint address outside its bound Pod", ns, serviceName)
				}
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("%s/%s has no ready mesh endpoints", ns, serviceName)
	}
	return nil
}

func hasInterceptedHTTPPort(ports []discoveryv1.EndpointPort, portName string, entryPort int32) bool {
	for _, port := range ports {
		if ptr.Deref(port.Name, "") != portName || ptr.Deref(port.Protocol, corev1.ProtocolTCP) != corev1.ProtocolTCP || port.Port == nil {
			continue
		}
		if entryPort != 0 {
			return *port.Port == entryPort
		}
		switch *port.Port {
		case 3389, 4140, 4143, 4190, 4191:
			return false
		default:
			return true
		}
	}
	return false
}

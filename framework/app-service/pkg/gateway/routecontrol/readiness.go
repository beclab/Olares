package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

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
	gatewayObject := &unstructured.Unstructured{}
	gatewayObject.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(ctx, client.ObjectKey{Namespace: gw.gatewayNamespace(), Name: gw.gatewayName()}, gatewayObject); err != nil {
		return err
	}
	conditions, _, _ := unstructured.NestedSlice(gatewayObject.Object, "status", "conditions")
	if !conditionReady(conditions, gatewayObject.GetGeneration(), "Programmed") {
		return fmt.Errorf("gateway is not programmed for its current generation")
	}
	if !gatewayListenerReady(gatewayObject, l4MTLSGatewaySectN) {
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
	if needsL4Authz(srr) {
		parents, _, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
		if err != nil || !reflect.DeepEqual(parents, parentRefsForSRR(gw, srr)) {
			return fmt.Errorf("protected HTTPRoute must only attach to Entry mTLS")
		}
	}

	if isSystemDesktopRoute(srr) && srrAuthMode(srr) != gateway.AuthModeExtAuth {
		return fmt.Errorf("Desktop Entry requires external authorization")
	}
	if srrAuthMode(srr) != gateway.AuthModeNone {
		policy, err := acceptedGatewayPolicy(ctx, c, securityPolicyGVK, srr.Namespace, securityPolicyName(srr), gw, l4MTLSGatewaySectN)
		if err != nil {
			return err
		}
		if needsL4Authz(srr) {
			desired := desiredApplicationExtAuthSecurityPolicy(srr)
			if desired == nil || !reflect.DeepEqual(policy.Object["spec"], desired.Object["spec"]) {
				return fmt.Errorf("Entry does not enforce the expected external authorization")
			}
		}
	}
	validation := schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"}
	if _, err := acceptedGatewayPolicy(ctx, c, validation, gw.gatewayNamespace(), "app-gateway-l4-entry-mtls", gw, l4MTLSGatewaySectN); err != nil {
		return err
	}

	if needsL4Authz(srr) {
		if !gatewayListenerReady(gatewayObject, l4AuthzSection) {
			return fmt.Errorf("L4 authorization listener is not programmed")
		}
		if _, err := acceptedGatewayPolicy(ctx, c, validation, gw.gatewayNamespace(), "app-gateway-l4-authz-mtls", gw, l4AuthzSection); err != nil {
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
	wizard := gateway.IsSystemWizardRoute(srr)
	if !wizard && !isSystemLoginRoute(srr) && (!isApplicationSRR(srr) || srrAuthMode(srr) == gateway.AuthModeNone) {
		return nil
	}
	authNS := autheliaNSForSRR(srr)
	if authNS == "" {
		return fmt.Errorf("Authelia ExtAuth has no bound user namespace")
	}
	ref := srrv1alpha1.UpstreamRef{ServiceName: AutheliaBackendServiceName, Port: AutheliaBackendPort}
	if err := checkMeshEndpoints(ctx, c, authNS, ref, wizard); err != nil {
		return fmt.Errorf("Authelia proxy: %w", err)
	}
	if err := meshedReadyEndpoints(ctx, c, "os-framework", ref); err != nil {
		return fmt.Errorf("Authelia server: %w", err)
	}
	return nil
}

func acceptedGatewayPolicy(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ns, name string, gw GatewayRef, section string) (*unstructured.Unstructured, error) {
	policy := &unstructured.Unstructured{}
	policy.SetGroupVersionKind(gvk)
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, policy); err != nil {
		return nil, err
	}
	if !policy.GetDeletionTimestamp().IsZero() {
		return nil, fmt.Errorf("%s %s is being deleted", gvk.Kind, name)
	}
	if gvk.Kind == "ClientTrafficPolicy" && !validEntryClientPolicy(policy, gw, section) {
		return nil, fmt.Errorf("ClientTrafficPolicy %s does not enforce the expected Entry client identity", name)
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
			return policy, nil
		}
	}
	return nil, fmt.Errorf("%s %s has no accepted status for this gateway", gvk.Kind, name)
}

func validEntryClientPolicy(policy *unstructured.Unstructured, gw GatewayRef, section string) bool {
	ref, _, _ := unstructured.NestedMap(policy.Object, "spec", "targetRef")
	if ref["group"] != "gateway.networking.k8s.io" || ref["kind"] != "Gateway" ||
		ref["name"] != gw.gatewayName() || ref["sectionName"] != section {
		return false
	}
	headers, _, _ := unstructured.NestedMap(policy.Object, "spec", "headers", "earlyRequestHeaders")
	if !validEntryIdentityHeaders(headers) {
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
	return checkMeshEndpoints(ctx, c, ns, upstream, false)
}

func checkMeshEndpoints(ctx context.Context, c client.Client, ns string, upstream srrv1alpha1.UpstreamRef, requireAuthContext bool) error {
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
			targetPort := interceptedHTTPPort(sl.Ports, portName, entryPort)
			if targetPort == 0 {
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
				!mesh.HasHTTPMeshInit(&pod) {
				return fmt.Errorf("%s/%s has a serving endpoint without an enforced Linkerd mesh posture", ns, serviceName)
			}
			if err := endpointInboundPolicyReady(ctx, c, &pod, serviceName, targetPort); err != nil {
				return fmt.Errorf("%s/%s has a serving endpoint without an enforced Linkerd mesh posture: %w", ns, serviceName, err)
			}
			if requireAuthContext && pod.Annotations["traffic.olares.io/auth-proxy-context-version"] != "1" {
				return fmt.Errorf("%s/%s has a serving endpoint without the tenant Auth proxy configuration", ns, serviceName)
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

// A default-deny system-server exposes the tenant Auth proxy through a
// dedicated Server. Evaluate that exception only on the resolved endpoint
// port, never on the Service port or the nonce API listener.
func endpointInboundPolicyReady(ctx context.Context, c client.Client, pod *corev1.Pod, serviceName string, targetPort int32) error {
	policy := pod.Annotations[mesh.LinkerdDefaultInboundPolicyAnnotation]
	if !mesh.ProxyHasInboundPolicy(pod, policy) {
		return fmt.Errorf("proxy policy does not match the Pod annotation")
	}
	if policy == mesh.InboundPolicyClusterAuthenticated {
		return nil
	}
	if policy != "deny" || serviceName != AutheliaBackendServiceName || targetPort != AutheliaBackendPort ||
		!strings.HasPrefix(pod.Namespace, "user-system-") || pod.Labels["app"] != "systemserver" {
		return fmt.Errorf("unsupported inbound policy %q for target port %d", policy, targetPort)
	}
	server := &unstructured.Unstructured{}
	server.SetGroupVersionKind(schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "Server"})
	key := client.ObjectKey{Namespace: pod.Namespace, Name: fmt.Sprintf("systemserver-existing-%d", targetPort)}
	if err := c.Get(ctx, key, server); err != nil {
		return fmt.Errorf("get Linkerd Server %s: %w", key, err)
	}
	access, _, _ := unstructured.NestedString(server.Object, "spec", "accessPolicy")
	selector, _, _ := unstructured.NestedMap(server.Object, "spec", "podSelector")
	port, _, _ := unstructured.NestedInt64(server.Object, "spec", "port")
	if !server.GetDeletionTimestamp().IsZero() || access != mesh.InboundPolicyClusterAuthenticated ||
		port != int64(targetPort) || !reflect.DeepEqual(selector, map[string]any{"matchLabels": map[string]any{"app": "systemserver"}}) {
		return fmt.Errorf("Linkerd Server %s does not enforce the expected system-server policy on port %d", key, targetPort)
	}
	for _, containers := range [][]corev1.Container{pod.Spec.Containers, pod.Spec.InitContainers} {
		for _, container := range containers {
			for _, p := range container.Ports {
				if p.ContainerPort == targetPort && (p.Protocol == "" || p.Protocol == corev1.ProtocolTCP) {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("Linkerd Server %s targets an undeclared TCP port %d", key, targetPort)
}

func hasInterceptedHTTPPort(ports []discoveryv1.EndpointPort, portName string, entryPort int32) bool {
	return interceptedHTTPPort(ports, portName, entryPort) != 0
}

func interceptedHTTPPort(ports []discoveryv1.EndpointPort, portName string, entryPort int32) int32 {
	for _, port := range ports {
		if ptr.Deref(port.Name, "") != portName || ptr.Deref(port.Protocol, corev1.ProtocolTCP) != corev1.ProtocolTCP || port.Port == nil {
			continue
		}
		if entryPort != 0 {
			if *port.Port == entryPort {
				return entryPort
			}
			return 0
		}
		switch *port.Port {
		case 3389, 4140, 4143, 4190, 4191:
			return 0
		default:
			if *port.Port > 0 && *port.Port <= 65535 {
				return *port.Port
			}
			return 0
		}
	}
	return 0
}

// Forwarding headers are trusted only from the Entry mTLS peer. Do not erase
// that verified edge context and replace the browser IP with a mesh Pod IP.
func entryIdentityHeaderSettings() map[string]any {
	return map[string]any{
		"remove":        []any{"x-bfl-user", "x-caller-user", "x-caller-appid", "x-olares-caller-jwt", "remote-user", "remote-groups", "remote-name", "remote-email", "remote-access-token", "authelia-nonce"},
		"removeOnMatch": []any{map[string]any{"type": "Prefix", "value": "remote-"}, map[string]any{"type": "Prefix", "value": "authelia-"}},
	}
}

// The CRD declares removal names as a set. Compare semantics rather than list
// order, while rejecting additions that could recreate an unverified identity.
func validEntryIdentityHeaders(headers map[string]any) bool {
	if len(headers) != 2 {
		return false
	}
	want := entryIdentityHeaderSettings()
	removed, ok := headers["remove"].([]any)
	if !ok || len(removed) != len(want["remove"].([]any)) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range removed {
		name, ok := value.(string)
		if !ok {
			return false
		}
		seen[strings.ToLower(name)] = true
	}
	for _, value := range want["remove"].([]any) {
		if !seen[value.(string)] {
			return false
		}
	}
	patterns, ok := headers["removeOnMatch"].([]any)
	if !ok || len(patterns) != 2 {
		return false
	}
	matched := map[string]bool{}
	for _, value := range patterns {
		p, ok := value.(map[string]any)
		if !ok || len(p) != 2 || p["type"] != "Prefix" {
			return false
		}
		prefix, ok := p["value"].(string)
		if !ok {
			return false
		}
		matched[prefix] = true
	}
	return matched["remote-"] && matched["authelia-"]
}

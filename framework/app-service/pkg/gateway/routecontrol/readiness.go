package routecontrol

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const ReasonAwaitingDataPlane = "AwaitingDataPlane"

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
	listeners, _, _ := unstructured.NestedSlice(gateway.Object, "status", "listeners")
	listenerReady := false
	for _, raw := range listeners {
		listener, _ := raw.(map[string]any)
		conditions, _, _ := unstructured.NestedSlice(listener, "conditions")
		if listener["name"] == l4MTLSGatewaySectN &&
			conditionReady(conditions, gateway.GetGeneration(), "Programmed") &&
			conditionReady(conditions, gateway.GetGeneration(), "ResolvedRefs") {
			listenerReady = true
		}
	}
	if !listenerReady {
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

	backendNS := srr.Spec.Upstream.ServiceNamespace
	if backendNS == "" {
		backendNS = srr.Namespace
	}
	backend := srr.Spec.Upstream.ServiceName
	if err := meshedReadyEndpoints(ctx, c, gw.gatewayNamespace(), "app-gateway-l4-mtls"); err != nil {
		return err
	}
	return meshedReadyEndpoints(ctx, c, backendNS, backend)
}

func gatewayPolicyReady(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ns, name string, gw GatewayRef) error {
	policy := &unstructured.Unstructured{}
	policy.SetGroupVersionKind(gvk)
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, policy); err != nil {
		return err
	}
	ancestors, _, _ := unstructured.NestedSlice(policy.Object, "status", "ancestors")
	for _, raw := range ancestors {
		ancestor, _ := raw.(map[string]any)
		ref, _, _ := unstructured.NestedMap(ancestor, "ancestorRef")
		if ancestor["controllerName"] != "gateway.envoyproxy.io/gatewayclass-controller" || ref["name"] != gw.gatewayName() ||
			(ref["sectionName"] != nil && ref["sectionName"] != l4MTLSGatewaySectN) {
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

func entryParentReady(route *unstructured.Unstructured, gw GatewayRef) bool {
	parents, _, _ := unstructured.NestedSlice(route.Object, "status", "parents")
	for _, raw := range parents {
		parent, _ := raw.(map[string]any)
		ref, _, _ := unstructured.NestedMap(parent, "parentRef")
		if parent["controllerName"] != "gateway.envoyproxy.io/gatewayclass-controller" ||
			ref["name"] != gw.gatewayName() || ref["sectionName"] != l4MTLSGatewaySectN {
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

func meshedReadyEndpoints(ctx context.Context, c client.Client, ns, serviceName string) error {
	var slices discoveryv1.EndpointSliceList
	if err := c.List(ctx, &slices, client.InNamespace(ns), client.MatchingLabels{discoveryv1.LabelServiceName: serviceName}); err != nil {
		return err
	}
	found := false
	for _, sl := range slices.Items {
		for _, endpoint := range sl.Endpoints {
			if endpoint.Conditions.Ready != nil && !*endpoint.Conditions.Ready {
				continue
			}
			ref := endpoint.TargetRef
			if endpoint.Conditions.Ready == nil || ref == nil || ref.Kind != "Pod" || ref.UID == "" || (ref.Namespace != "" && ref.Namespace != ns) {
				return fmt.Errorf("%s/%s has an endpoint without a bound Pod", ns, serviceName)
			}
			var pod corev1.Pod
			if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: ref.Name}, &pod); err != nil {
				return err
			}
			proxyReady := false
			for _, s := range pod.Status.ContainerStatuses {
				if s.Name == "linkerd-proxy" && s.Ready {
					proxyReady = true
					break
				}
			}
			if pod.UID != ref.UID || pod.Spec.HostNetwork || !pod.DeletionTimestamp.IsZero() || !proxyReady ||
				pod.Annotations["config.linkerd.io/skip-inbound-ports"] != "" || pod.Annotations["config.linkerd.io/skip-outbound-ports"] != "" {
				return fmt.Errorf("%s/%s has a serving endpoint without a ready Linkerd proxy", ns, serviceName)
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("%s/%s has no ready mesh endpoints", ns, serviceName)
	}
	return nil
}

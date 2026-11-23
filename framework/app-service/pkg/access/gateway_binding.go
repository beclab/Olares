package access

import (
	"context"
	"errors"
	"fmt"
	"maps"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

const (
	internalNamespace = "os-gateway"
	internalGateway   = "app-access-internal"
	internalListener  = "access"
	internalAccount   = "app-access-envoy"
	internalPort      = 8080
)

var (
	routeGVK   = schema.GroupVersionKind{Group: policyRouteGroup, Version: "v1", Kind: "HTTPRoute"}
	grantGVK   = schema.GroupVersionKind{Group: policyRouteGroup, Version: "v1beta1", Kind: "ReferenceGrant"}
	gatewayGVK = schema.GroupVersionKind{Group: policyRouteGroup, Version: "v1", Kind: "Gateway"}
)

func (b boundPolicy) gatewayPolicy() PolicySpec {
	p := b.policy
	p.Namespace = internalNamespace
	p.RouteMode = RouteModeDirect
	p.ServerPort = internalPort
	p.TargetSelector = internalSelector()
	return p
}

func internalSelector() map[string]string {
	return map[string]string{
		"gateway.envoyproxy.io/owning-gateway-name":      internalGateway,
		"gateway.envoyproxy.io/owning-gateway-namespace": internalNamespace,
		"app.kubernetes.io/component":                    "proxy",
		"app.kubernetes.io/managed-by":                   "envoy-gateway",
	}
}

func (b boundPolicy) host() string { return b.policy.ProjectionID + ".access.olares.internal" }

func (b boundPolicy) gatewayRoute() DesiredObject {
	host := b.host()
	svc := bindingRef(b.target, "service").Name
	port, _, _ := unstructured.NestedInt64(b.target, "service", "servicePort")
	return b.gatewayObject(routeGVK, PolicyNamespace, "edge", map[string]any{
		"parentRefs": []any{map[string]any{"group": policyRouteGroup, "kind": "Gateway", "namespace": internalNamespace, "name": internalGateway, "sectionName": internalListener}},
		"hostnames":  []any{host},
		"rules": []any{map[string]any{
			"matches": httpMatches(b.policy.Rules),
			"filters": []any{map[string]any{
				"type":                  "RequestHeaderModifier",
				"requestHeaderModifier": map[string]any{"remove": []any{"X-Bfl-User", "X-Caller-User", "X-Caller-AppId", "X-Caller-Jwt"}},
			}},
			"backendRefs": []any{map[string]any{"group": "", "kind": "Service", "namespace": b.policy.Namespace, "name": svc, "port": port, "weight": int64(1)}},
		}},
	})
}

func (b boundPolicy) referenceGrant() DesiredObject {
	return b.gatewayObject(grantGVK, b.policy.Namespace, "ref", map[string]any{
		"from": []any{map[string]any{"group": policyRouteGroup, "kind": "HTTPRoute", "namespace": PolicyNamespace}},
		"to":   []any{map[string]any{"group": "", "kind": "Service", "name": bindingRef(b.target, "service").Name}},
	})
}

func (b boundPolicy) gatewayObject(gvk schema.GroupVersionKind, ns, suffix string, spec map[string]any) DesiredObject {
	u := &unstructured.Unstructured{Object: map[string]any{"spec": spec}}
	u.SetGroupVersionKind(gvk)
	u.SetNamespace(ns)
	u.SetName(b.policy.ProjectionID + "-" + suffix)
	u.SetLabels(map[string]string{ManagedByLabel: ManagedByAccessPolicy, PolicyNameLabel: projectionOwner(b.policy)})
	return DesiredObject{GVK: gvk, Namespace: ns, Name: u.GetName(), Object: u}
}

func (b boundPolicy) reconcileGateway(ctx context.Context, c client.Client) error {
	// The pilot only permits exact paths across both independent HTTP parsers.
	for _, r := range b.policy.Rules {
		if r.PathType != "" && r.PathType != PathTypeExact {
			return errors.Join(fmt.Errorf("internal gateway grants require exact paths"), b.withdrawGateway(ctx, c))
		}
	}
	readyErr := b.checkGateway(ctx, c)
	if readyErr != nil {
		return errors.Join(readyErr, b.withdrawGateway(ctx, c))
	}

	source := b.gatewayPolicy()
	// Share a deny Server across grants on this one reserved HTTP listener.
	source.ServerID = internalGateway
	target := b.policy
	target.GatewaySAName = internalAccount

	in, err := buildHostProjections(source, b.host())
	if err != nil {
		return err
	}
	out, err := buildHostProjections(target, b.host())
	if err != nil {
		return err
	}
	// Apply default-deny and routes first. If readiness is lost, remove
	// both allows while retaining routes so controllers can reconverge.
	route, grant := b.gatewayRoute(), b.referenceGrant()
	if err := applyProjections(ctx, c, append(append(append([]DesiredObject{grant}, out[:3]...), in[1:3]...), route)); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	for _, ref := range []struct {
		route                          DesiredObject
		group, kind, ns, name, section string
	}{
		{route, policyRouteGroup, "Gateway", internalNamespace, internalGateway, internalListener},
		{in[2], "policy.linkerd.io", "Server", internalNamespace, internalGateway + "-server", ""},
		{out[2], "policy.linkerd.io", "Server", b.policy.Namespace, out[0].Name, ""},
	} {
		ok, err := routeReady(ctx, c, ref.route, ref.group, ref.kind, ref.ns, ref.name, ref.section)
		if err != nil || !ok {
			return errors.Join(err, fmt.Errorf("route %s is not accepted for its current generation", ref.route.Name), b.withdrawAllows(ctx, c))
		}
	}
	if err := b.stillActive(ctx, c); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	if err := applyProjections(ctx, c, []DesiredObject{out[3]}); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	// A's gateway ingress permission is the final write.
	if err := b.stillActive(ctx, c); err != nil {
		return errors.Join(err, b.withdrawAllows(ctx, c))
	}
	return applyProjections(ctx, c, []DesiredObject{in[3]})
}

func buildHostProjections(p PolicySpec, host string) ([]DesiredObject, error) {
	plan, err := CompilePolicy(p, true)
	if err != nil {
		return nil, err
	}
	objs, err := BuildLinkerdObjects(plan, p)
	if err != nil {
		return nil, err
	}
	if err := unstructured.SetNestedStringSlice(objs[2].Object.Object, []string{host}, "spec", "hostnames"); err != nil {
		return nil, err
	}
	return objs, nil
}

func (b boundPolicy) withdrawAllows(ctx context.Context, c client.Client) error {
	var result error
	for _, p := range []PolicySpec{b.gatewayPolicy(), b.policy} {
		p.DesiredState = DesiredRevoked
		plan, err := CompilePolicy(p, true)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		for _, d := range deleteMarkers(plan, p) {
			if d.GVK.Kind == "AuthorizationPolicy" {
				result = errors.Join(result, applyProjections(ctx, c, []DesiredObject{d}))
			}
		}
	}
	return result
}

func (b boundPolicy) withdrawGateway(ctx context.Context, c client.Client) error {
	if err := b.withdrawAllows(ctx, c); err != nil {
		return err
	}
	route, grant := b.gatewayRoute(), b.referenceGrant()
	route.Delete, grant.Delete = true, true
	return errors.Join(applyProjections(ctx, c, []DesiredObject{route, grant}),
		reconcileProjectedPolicy(ctx, c, revoked(b.gatewayPolicy())), reconcileProjectedPolicy(ctx, c, revoked(b.policy)))
}

func revoked(p PolicySpec) PolicySpec { p.DesiredState = DesiredRevoked; return p }

func (b boundPolicy) checkGateway(ctx context.Context, c client.Client) error {
	ref := objectMap(b.scope, "gatewayRef")
	if value(ref, "name") != internalGateway || value(ref, "listenerName") != internalListener {
		return fmt.Errorf("grant does not use the reserved internal gateway")
	}
	gw, err := boundObject(ctx, c, internalNamespace, gatewayGVK, bindingRef(b.scope, "gatewayRef"))
	if err != nil {
		return err
	}
	spec, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	if len(spec) != 1 {
		return fmt.Errorf("internal gateway must have exactly one listener")
	}
	config, _ := spec[0].(map[string]any)
	portNumber, _, _ := unstructured.NestedInt64(config, "port")
	if value(config, "name") != internalListener || value(config, "protocol") != "HTTP" ||
		value(config, "hostname") != "*.access.olares.internal" || portNumber != 8080 ||
		value(config, "allowedRoutes", "namespaces", "from") != "Selector" ||
		value(config, "allowedRoutes", "namespaces", "selector", "matchLabels", "kubernetes.io/metadata.name") != PolicyNamespace ||
		value(gw.Object, "spec", "infrastructure", "parametersRef", "name") != internalGateway {
		return fmt.Errorf("internal gateway listener is not isolated")
	}
	conds, _, _ := unstructured.NestedSlice(gw.Object, "status", "conditions")
	if !currentCondition(conds, gw.GetGeneration(), "Programmed") {
		return fmt.Errorf("internal gateway is not programmed")
	}
	listeners, _, _ := unstructured.NestedSlice(gw.Object, "status", "listeners")
	var listenerReady bool
	for _, item := range listeners {
		l, _ := item.(map[string]any)
		conds, _, _ := unstructured.NestedSlice(l, "conditions")
		if value(l, "name") == internalListener && currentCondition(conds, gw.GetGeneration(), "Programmed") && currentCondition(conds, gw.GetGeneration(), "ResolvedRefs") {
			listenerReady = true
		}
	}
	if !listenerReady {
		return fmt.Errorf("internal gateway listener is not ready")
	}
	server := &unstructured.Unstructured{}
	server.SetGroupVersionKind(schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "Server"})
	if err := c.Get(ctx, client.ObjectKey{Namespace: internalNamespace, Name: internalGateway + "-server"}, server); err != nil {
		return err
	}
	port, _, _ := unstructured.NestedInt64(server.Object, "spec", "port")
	selector, _, _ := unstructured.NestedStringMap(server.Object, "spec", "podSelector", "matchLabels")
	if port != internalPort || !maps.Equal(selector, internalSelector()) || value(server.Object, "spec", "proxyProtocol") != "HTTP/1" ||
		value(server.Object, "spec", "accessPolicy") != "deny" {
		return fmt.Errorf("internal gateway requires a default-deny HTTP Server")
	}
	if err := checkMeshService(ctx, c, internalNamespace, internalGateway, internalAccount, internalPort); err != nil {
		return err
	}
	return b.checkTarget(ctx, c)
}

func (b boundPolicy) checkTarget(ctx context.Context, c client.Client) error {
	return checkMeshService(ctx, c, b.policy.Namespace, bindingRef(b.target, "service").Name, bindingRef(b.target, "serviceAccount").Name, b.policy.ServerPort)
}

func checkMeshService(ctx context.Context, c client.Client, ns, svc, sa string, port int32) error {
	var slices discoveryv1.EndpointSliceList
	if err := c.List(ctx, &slices, client.InNamespace(ns), client.MatchingLabels{discoveryv1.LabelServiceName: svc}); err != nil {
		return err
	}
	var service corev1.Service
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: svc}, &service); err != nil {
		return err
	}
	found := false
	for _, slice := range slices.Items {
		owned := false
		for _, ref := range slice.OwnerReferences {
			if ref.Kind == "Service" && ref.UID == service.UID {
				owned = true
			}
		}
		if !owned {
			return fmt.Errorf("unexpected EndpointSlice owner for %s/%s", ns, svc)
		}
		portFound := false
		for _, p := range slice.Ports {
			if p.Port != nil && *p.Port == port {
				portFound = true
			}
		}
		for _, e := range slice.Endpoints {
			if e.Conditions.Ready != nil && !*e.Conditions.Ready {
				continue
			}
			ref := e.TargetRef
			if e.Conditions.Ready == nil || !portFound || ref == nil || ref.Kind != "Pod" || (ref.Namespace != "" && ref.Namespace != ns) || ref.UID == "" {
				return fmt.Errorf("invalid ready endpoint for %s/%s", ns, svc)
			}
			var pod corev1.Pod
			if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: ref.Name}, &pod); err != nil {
				return err
			}
			if pod.UID != ref.UID || pod.Spec.ServiceAccountName != sa || !pod.DeletionTimestamp.IsZero() ||
				mesh.CheckHTTPMeshInterception(&pod) != nil {
				return fmt.Errorf("endpoint identity or interception is invalid for %s/%s", ns, svc)
			}
			proxyReady := false
			for _, s := range pod.Status.ContainerStatuses {
				proxyReady = proxyReady || (s.Name == "linkerd-proxy" && s.Ready)
			}
			if !proxyReady {
				return fmt.Errorf("endpoint Linkerd proxy is not ready for %s/%s", ns, svc)
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no ready mesh endpoints for %s/%s", ns, svc)
	}
	return nil
}

func currentCondition(values []any, generation int64, kind string) bool {
	for _, item := range values {
		v, _ := item.(map[string]any)
		g, present, _ := unstructured.NestedInt64(v, "observedGeneration")
		if value(v, "type") == kind && value(v, "status") == "True" && present && g == generation {
			return true
		}
	}
	return false
}

func routeReady(ctx context.Context, c client.Client, d DesiredObject, group, kind, ns, name, section string) (bool, error) {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(routeGVK)
	if err := c.Get(ctx, client.ObjectKey{Namespace: d.Namespace, Name: d.Name}, u); err != nil {
		return false, err
	}
	parents, _, _ := unstructured.NestedSlice(u.Object, "status", "parents")
	for _, item := range parents {
		p, _ := item.(map[string]any)
		ref := objectMap(p, "parentRef")
		if (kind == "Gateway" && value(p, "controllerName") != "gateway.envoyproxy.io/gatewayclass-controller") ||
			value(ref, "group") != group || value(ref, "kind") != kind || value(ref, "name") != name || value(ref, "sectionName") != section {
			continue
		}
		parentNS := value(ref, "namespace")
		if parentNS == "" {
			parentNS = u.GetNamespace()
		}
		if parentNS != ns {
			continue
		}
		conditions, _, _ := unstructured.NestedSlice(p, "conditions")
		if currentCondition(conditions, u.GetGeneration(), "Accepted") &&
			(kind == "Server" || currentCondition(conditions, u.GetGeneration(), "ResolvedRefs")) {
			return true, nil
		}
	}
	return false, nil
}

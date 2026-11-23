package access

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ManagedByLabel      = "app.kubernetes.io/managed-by"
	ManagedByAccessPolicy = "access-policy"
	PolicyNameLabel     = "access.olares.io/policy"
)

// DesiredObject is one projected API object the reconciler should apply or delete.
type DesiredObject struct {
	GVK       schema.GroupVersionKind
	Namespace string
	Name      string
	Object    *unstructured.Unstructured
	Delete    bool
}

// BuildLinkerdObjects turns a compile plan into minimal Linkerd policy objects.
// When SkipCompile or Revoked, returns delete markers for previously managed names.
func BuildLinkerdObjects(plan CompiledObjects, p PolicySpec) ([]DesiredObject, error) {
	if plan.SkipCompile {
		if p.DesiredState == DesiredRevoked {
			return deleteMarkers(plan, p), nil
		}
		return nil, nil
	}
	if plan.ServerName == "" || plan.AuthorizationPolicyName == "" {
		return nil, fmt.Errorf("incomplete compile plan")
	}
	ns := p.Namespace
	labels := map[string]string{
		ManagedByLabel:  ManagedByAccessPolicy,
		PolicyNameLabel: p.Name,
	}

	server := &unstructured.Unstructured{}
	server.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "policy.linkerd.io", Version: "v1beta1", Kind: "Server",
	})
	server.SetName(plan.ServerName)
	server.SetNamespace(ns)
	server.SetLabels(labels)
	_ = unstructured.SetNestedField(server.Object, int64(p.ServerPort), "spec", "port")
	_ = unstructured.SetNestedStringMap(server.Object, map[string]string{"app": p.Name}, "spec", "podSelector", "matchLabels")

	mtls := &unstructured.Unstructured{}
	mtls.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "MeshTLSAuthentication",
	})
	mtls.SetName(plan.MeshTLSAuthName)
	mtls.SetNamespace(ns)
	mtls.SetLabels(labels)
	_ = unstructured.SetNestedStringSlice(mtls.Object, []string{plan.IdentityPrincipal}, "spec", "identities")

	authz := &unstructured.Unstructured{}
	authz.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "AuthorizationPolicy",
	})
	authz.SetName(plan.AuthorizationPolicyName)
	authz.SetNamespace(ns)
	authz.SetLabels(labels)
	_ = unstructured.SetNestedField(authz.Object, map[string]any{
		"group": "policy.linkerd.io",
		"kind":  "Server",
		"name":  plan.ServerName,
	}, "spec", "targetRef")
	_ = unstructured.SetNestedSlice(authz.Object, []any{
		map[string]any{
			"group": "policy.linkerd.io",
			"kind":  "MeshTLSAuthentication",
			"name":  plan.MeshTLSAuthName,
		},
	}, "spec", "requiredAuthenticationRefs")

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "policy.linkerd.io", Version: "v1beta3", Kind: "HTTPRoute",
	})
	route.SetName(plan.HTTPRouteName)
	route.SetNamespace(ns)
	route.SetLabels(labels)
	_ = unstructured.SetNestedSlice(route.Object, []any{
		map[string]any{
			"kind":  "Server",
			"group": "policy.linkerd.io",
			"name":  plan.ServerName,
		},
	}, "spec", "parentRefs")

	out := []DesiredObject{
		{GVK: server.GroupVersionKind(), Namespace: ns, Name: plan.ServerName, Object: server},
		{GVK: mtls.GroupVersionKind(), Namespace: ns, Name: plan.MeshTLSAuthName, Object: mtls},
		{GVK: authz.GroupVersionKind(), Namespace: ns, Name: plan.AuthorizationPolicyName, Object: authz},
		{GVK: route.GroupVersionKind(), Namespace: ns, Name: plan.HTTPRouteName, Object: route},
	}
	return out, nil
}

func deleteMarkers(plan CompiledObjects, p PolicySpec) []DesiredObject {
	ns := p.Namespace
	names := []struct {
		gvk  schema.GroupVersionKind
		name string
	}{
		{schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta1", Kind: "Server"}, plan.ServerName},
		{schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "MeshTLSAuthentication"}, plan.MeshTLSAuthName},
		{schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "AuthorizationPolicy"}, plan.AuthorizationPolicyName},
		{schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1beta3", Kind: "HTTPRoute"}, plan.HTTPRouteName},
	}
	var out []DesiredObject
	for _, n := range names {
		if n.name == "" {
			continue
		}
		out = append(out, DesiredObject{
			GVK: n.gvk, Namespace: ns, Name: n.name, Delete: true,
			Object: &unstructured.Unstructured{Object: map[string]any{
				"metadata": map[string]any{
					"name":      n.name,
					"namespace": ns,
					"labels": map[string]any{
						ManagedByLabel:  ManagedByAccessPolicy,
						PolicyNameLabel: p.Name,
					},
				},
			}},
		})
	}
	return out
}

// PolicyFromUnstructured extracts PolicySpec fields from an AppAccessPolicy-like object.
func PolicyFromUnstructured(u *unstructured.Unstructured) (PolicySpec, error) {
	if u == nil {
		return PolicySpec{}, fmt.Errorf("nil policy object")
	}
	p := PolicySpec{Name: u.GetName()}
	if ns, ok, _ := unstructured.NestedString(u.Object, "spec", "targetNamespace"); ok && ns != "" {
		p.Namespace = ns
	} else {
		p.Namespace = u.GetNamespace()
	}
	p.DesiredState, _, _ = unstructured.NestedString(u.Object, "spec", "desiredState")
	p.RouteMode, _, _ = unstructured.NestedString(u.Object, "spec", "routeMode")
	if port, ok, _ := unstructured.NestedInt64(u.Object, "spec", "serverPort"); ok {
		p.ServerPort = int32(port)
	}
	p.SourceSANamespace, _, _ = unstructured.NestedString(u.Object, "spec", "sourceServiceAccount", "namespace")
	p.SourceSAName, _, _ = unstructured.NestedString(u.Object, "spec", "sourceServiceAccount", "name")
	p.GatewaySAName, _, _ = unstructured.NestedString(u.Object, "spec", "gatewayServiceAccount", "name")
	rulesRaw, _, _ := unstructured.NestedSlice(u.Object, "spec", "rules")
	for _, r := range rulesRaw {
		m, ok := r.(map[string]any)
		if !ok {
			continue
		}
		rule := HTTPRule{}
		if v, ok := m["method"].(string); ok {
			rule.Method = v
		}
		if v, ok := m["path"].(string); ok {
			rule.Path = v
		}
		if v, ok := m["pathType"].(string); ok {
			rule.PathType = v
		}
		p.Rules = append(p.Rules, rule)
	}
	return p, nil
}

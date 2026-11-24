package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const (
	l4AuthzSection = "l4-authz"
	l4AuthzPort    = 9444
)

func needsL4Authz(srr *srrv1alpha1.SharedRouteRegistry) bool {
	return isApplicationSRR(srr) && srrAuthMode(srr) == gateway.AuthModeExtAuth
}

func l4AuthzRoute(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	route.SetName(srr.Name + ".l4-authz")
	route.SetNamespace(srr.Namespace)
	return route
}

func validateL4AuthzAuthorities(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if !needsL4Authz(srr) {
		return nil
	}
	if ns := autheliaNSForSRR(srr); ns == "" || len(validation.IsDNS1123Label(ns)) != 0 {
		return fmt.Errorf("L4 authorization requires a valid entrance owner")
	}
	if len(srr.Spec.HostPatterns) == 0 {
		return fmt.Errorf("L4 authorization requires an entrance authority")
	}
	authorities := make(map[string]bool, len(srr.Spec.HostPatterns))
	for _, raw := range srr.Spec.HostPatterns {
		host := strings.ToLower(strings.TrimSpace(raw))
		if len(validation.IsDNS1123Subdomain(host)) != 0 {
			return fmt.Errorf("entrance has no exact authority binding for %q", host)
		}
		authorities[host] = true
	}
	var registries srrv1alpha1.SharedRouteRegistryList
	if err := c.List(ctx, &registries); err != nil {
		return err
	}
	for _, other := range registries.Items {
		if (other.Namespace == srr.Namespace && other.Name == srr.Name) || !other.DeletionTimestamp.IsZero() ||
			other.Spec.RouteMode == srrv1alpha1.RouteModeDirect {
			continue
		}
		for _, host := range other.Spec.HostPatterns {
			if authorities[strings.ToLower(strings.TrimSpace(host))] {
				return fmt.Errorf("entrance authority %q is also registered by %s/%s", host, other.Namespace, other.Name)
			}
		}
	}
	return nil
}

func desiredL4AuthzRoute(gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry) (*unstructured.Unstructured, error) {
	authNS := autheliaNSForSRR(srr)
	owner := strings.TrimPrefix(authNS, "user-system-")
	if authNS == "" || len(validation.IsDNS1123Label(authNS)) != 0 {
		return nil, fmt.Errorf("L4 authorization requires a valid entrance owner")
	}
	hosts := HTTPRouteHostnames(srr.Spec.HostPatterns)
	if len(hosts) == 0 {
		return nil, fmt.Errorf("L4 authorization requires a bound entrance authority")
	}
	for _, host := range hosts {
		if strings.Contains(host.(string), "*") {
			return nil, fmt.Errorf("L4 authorization requires exact entrance authorities")
		}
	}
	rule := map[string]any{
		"matches": []any{map[string]any{
			"path": map[string]any{"type": "PathPrefix", "value": AutheliaVerifyPath},
		}},
		"backendRefs": []any{map[string]any{
			"name": AutheliaBackendServiceName, "namespace": authNS, "port": int64(AutheliaBackendPort),
		}},
		"filters": []any{
			map[string]any{"type": "URLRewrite", "urlRewrite": map[string]any{
				"path": map[string]any{"type": "ReplaceFullPath", "replaceFullPath": AutheliaVerifyPath},
			}},
			map[string]any{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
				"set": []any{map[string]any{"name": "X-BFL-USER", "value": owner}},
			}},
		},
	}
	route := l4AuthzRoute(srr)
	route.SetLabels(map[string]string{ManagedByLabel: ManagedByValue, InstanceLabel: srr.Name})
	setOwnerSRR(route, srr)
	route.Object["spec"] = map[string]any{
		"parentRefs": []any{map[string]any{
			"group": "gateway.networking.k8s.io", "kind": "Gateway", "namespace": gw.gatewayNamespace(),
			"name": gw.gatewayName(), "sectionName": l4AuthzSection,
		}},
		"hostnames": hosts,
		"rules":     []any{rule},
	}
	return route, nil
}

func reconcileL4AuthzRoute(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, enabled bool) error {
	current := l4AuthzRoute(srr)
	err := c.Get(ctx, client.ObjectKeyFromObject(current), current)
	if apierrors.IsNotFound(err) {
		if !enabled {
			return nil
		}
		desired, err := desiredL4AuthzRoute(gw, srr)
		if err != nil {
			return err
		}
		return c.Create(ctx, desired)
	}
	if err != nil {
		return err
	}
	owner := metav1.GetControllerOf(current)
	if owner == nil || owner.Kind != "SharedRouteRegistry" || owner.UID != srr.UID {
		return fmt.Errorf("authorization route %s/%s is not owned by this entrance", current.GetNamespace(), current.GetName())
	}
	if !enabled {
		uid, version := current.GetUID(), current.GetResourceVersion()
		return client.IgnoreNotFound(c.Delete(ctx, current, client.Preconditions{UID: &uid, ResourceVersion: &version}))
	}
	desired, err := desiredL4AuthzRoute(gw, srr)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		return c.Update(ctx, current)
	}
	return nil
}

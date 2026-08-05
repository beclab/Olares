package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
)

const (
	probePathsAnnotation = mesh.ProbePathsAnnotation
	probeUARequiredLabel = mesh.ProbeUARequiredLabel
)

// Probe UA must match webhook.getProbeUA shape: {uuid}/{md5hex}.
// Full crypto verify stays with Authelia/probe_secret; EG rejects missing/malformed UA
// so probe paths never become an unsigned auth hole.
const probeUAGuardLua = `
function envoy_on_request(request_handle)
  local path = request_handle:headers():get(":path")
  if path == nil then
    return
  end
  -- strip query
  local q = string.find(path, "?", 1, true)
  if q then
    path = string.sub(path, 1, q - 1)
  end
  local ua = request_handle:headers():get("user-agent")
  if ua == nil or ua == "" then
    request_handle:respond({[":status"] = "403"}, "probe ua required")
    return
  end
  -- uuid/md5hex (webhook getProbeUA)
  if not string.match(ua, "^[%x%-]+/[%x]+$") then
    request_handle:respond({[":status"] = "403"}, "probe ua invalid")
    return
  end
end
`

func entranceProbeRouteName(srr *srrv1alpha1.SharedRouteRegistry) string {
	if srr == nil {
		return ""
	}
	return mesh.EntranceProbeBypassRouteName(srr.Name)
}

func entranceProbePolicyName(srr *srrv1alpha1.SharedRouteRegistry) string {
	if srr == nil {
		return ""
	}
	return mesh.EntranceProbeBypassPolicyName(srr.Name)
}

func needsEntranceProbeBypass(srr *srrv1alpha1.SharedRouteRegistry) bool {
	return needsEntranceExtAuth(srr)
}

// uniqueProbePaths dedupes and sorts probe paths (empty paths dropped).
func uniqueProbePaths(paths []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" || p == "/" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// listProbePathsForUpstream collects HTTPGet probe paths from Pods selected by the upstream Service.
// Data source: Service.Spec.Selector → Pods in the Service namespace (1.12.6 getHTTProbePath parity).
func listProbePathsForUpstream(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) ([]string, error) {
	if c == nil || srr == nil || svc == nil {
		return nil, fmt.Errorf("client, srr, and service required")
	}
	if len(svc.Spec.Selector) == 0 {
		return nil, nil
	}
	var pods corev1.PodList
	if err := c.List(ctx, &pods, &client.ListOptions{
		Namespace:     svc.Namespace,
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
	}); err != nil {
		return nil, fmt.Errorf("list pods for probe paths: %w", err)
	}
	var raw []string
	for i := range pods.Items {
		raw = append(raw, sidecar.GetHTTPProbePaths(&pods.Items[i])...)
	}
	return uniqueProbePaths(raw), nil
}

func desiredEntranceProbeHTTPRoute(gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, port int32, paths []string) *unstructured.Unstructured {
	hosts := HTTPRouteHostnames(srr.Spec.HostPatterns)
	parentRef := map[string]any{
		"group":     "gateway.networking.k8s.io",
		"kind":      "Gateway",
		"namespace": gw.gatewayNamespace(),
		"name":      gw.gatewayName(),
	}
	if section, ok := gatewaySectionForSRR(gw, srr); ok {
		parentRef["sectionName"] = section
	}
	backendRef := map[string]any{
		"group":     "",
		"kind":      "Service",
		"name":      srr.Spec.Upstream.ServiceName,
		"namespace": srr.Namespace,
		"port":      int64(port),
		"weight":    int64(1),
	}
	if ns := srr.Spec.Upstream.ServiceNamespace; ns != "" && ns != srr.Namespace {
		backendRef["namespace"] = ns
	}
	var rules []any
	for _, p := range paths {
		rules = append(rules, map[string]any{
			"matches": []any{
				map[string]any{
					"path": map[string]any{"type": "Exact", "value": p},
				},
			},
			"backendRefs": []any{backendRef},
		})
	}
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      entranceProbeRouteName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel:                   ManagedByValue,
				InstanceLabel:                    srr.Name,
				"gateway.olares.io/auth-kind":    mesh.AuthKindEntranceProbeBypass,
			},
			"annotations": map[string]any{
				probePathsAnnotation: strings.Join(paths, ","),
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{parentRef},
			"hostnames":  hosts,
			"rules":      rules,
		},
	}}
	setOwnerSRR(obj, srr)
	return obj
}

func desiredEntranceProbeUAPolicy(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "EnvoyExtensionPolicy",
		"metadata": map[string]any{
			"name":      entranceProbePolicyName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel:       ManagedByValue,
				InstanceLabel:        srr.Name,
				"gateway.olares.io/auth-kind": mesh.AuthKindEntranceProbeBypass,
				probeUARequiredLabel: "true",
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  entranceProbeRouteName(srr),
			},
			"lua": []any{
				map[string]any{
					"type":   "Inline",
					"inline": strings.TrimSpace(probeUAGuardLua),
				},
			},
		},
	}}
}

func applyOrDeleteEntranceProbeBypass(ctx context.Context, c client.Client, gw GatewayRef, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service, port int32) error {
	if !needsEntranceProbeBypass(srr) {
		return deleteEntranceProbeBypass(ctx, c, srr)
	}
	paths, err := listProbePathsForUpstream(ctx, c, srr, svc)
	if err != nil {
		klog.Errorf("deenvy: list probe paths for %s/%s failed: %v", srr.Namespace, srr.Name, err)
		return err
	}
	if len(paths) == 0 {
		return deleteEntranceProbeBypass(ctx, c, srr)
	}
	if err := applyUnstructured(ctx, c, desiredEntranceProbeHTTPRoute(gw, srr, port, paths), schema.GroupVersionKind{
		Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute",
	}); err != nil {
		return fmt.Errorf("apply probe bypass HTTPRoute: %w", err)
	}
	if err := applyUnstructured(ctx, c, desiredEntranceProbeUAPolicy(srr), envoyExtensionPolicyGVK); err != nil {
		return fmt.Errorf("apply probe UA EnvoyExtensionPolicy: %w", err)
	}
	return nil
}

func applyUnstructured(ctx context.Context, c client.Client, desired *unstructured.Unstructured, gvk schema.GroupVersionKind) error {
	desired.SetGroupVersionKind(gvk)
	if desired.GetKind() == "HTTPRoute" {
		// Owner set by caller when SRR available; optional.
	}
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(gvk)
	key := types.NamespacedName{Namespace: desired.GetNamespace(), Name: desired.GetName()}
	err := c.Get(ctx, key, current)
	if apierrors.IsNotFound(err) {
		if err := c.Create(ctx, desired); err != nil {
			klog.Errorf("deenvy: create %s %s/%s failed: %v", gvk.Kind, key.Namespace, key.Name, err)
			return err
		}
		klog.Infof("deenvy: created %s %s/%s", gvk.Kind, key.Namespace, key.Name)
		return nil
	}
	if err != nil {
		return err
	}
	desiredSpec, _, _ := unstructured.NestedMap(desired.Object, "spec")
	currentSpec, _, _ := unstructured.NestedMap(current.Object, "spec")
	labelsEqual := reflect.DeepEqual(desired.GetLabels(), current.GetLabels())
	annsEqual := reflect.DeepEqual(desired.GetAnnotations(), current.GetAnnotations())
	if reflect.DeepEqual(desiredSpec, currentSpec) && labelsEqual && annsEqual {
		return nil
	}
	updated := current.DeepCopy()
	if err := unstructured.SetNestedMap(updated.Object, desiredSpec, "spec"); err != nil {
		return err
	}
	updated.SetLabels(desired.GetLabels())
	updated.SetAnnotations(desired.GetAnnotations())
	if err := c.Update(ctx, updated); err != nil {
		klog.Errorf("deenvy: update %s %s/%s failed: %v", gvk.Kind, key.Namespace, key.Name, err)
		return err
	}
	klog.Infof("deenvy: updated %s %s/%s", gvk.Kind, key.Namespace, key.Name)
	return nil
}

func deleteEntranceProbeBypass(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return nil
	}
	for _, pair := range []struct {
		name string
		gvk  schema.GroupVersionKind
	}{
		{entranceProbePolicyName(srr), envoyExtensionPolicyGVK},
		{entranceProbeRouteName(srr), schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}},
	} {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(pair.gvk)
		obj.SetNamespace(srr.Namespace)
		obj.SetName(pair.name)
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("deenvy: delete %s %s/%s failed: %v", pair.gvk.Kind, srr.Namespace, pair.name, err)
			return err
		}
	}
	return nil
}

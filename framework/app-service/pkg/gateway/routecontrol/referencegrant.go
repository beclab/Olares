package routecontrol

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const referenceGrantAPIVersion = "gateway.networking.k8s.io/v1beta1"

var referenceGrantGVK = schema.GroupVersionKind{
	Group:   "gateway.networking.k8s.io",
	Version: "v1beta1",
	Kind:    "ReferenceGrant",
}

// referenceGrantName derives a stable ReferenceGrant name for HTTPRoute in
// srr.Namespace referencing a Service in upstreamNS.
func referenceGrantName(srr *srrv1alpha1.SharedRouteRegistry) string {
	const prefix = "allow-httproute-"
	raw := prefix + srr.Namespace
	if len(raw) <= 63 {
		return raw
	}
	trim := raw[:63]
	return strings.TrimRight(trim, "-")
}

func applyReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, upstreamNS, serviceName string) error {
	if upstreamNS == "" || upstreamNS == srr.Namespace || serviceName == "" {
		return deleteReferenceGrant(ctx, c, srr)
	}
	name := referenceGrantName(srr)
	desired := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": referenceGrantAPIVersion,
		"kind":       "ReferenceGrant",
		"metadata": map[string]any{
			"name":      name,
			"namespace": upstreamNS,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		"spec": map[string]any{
			"from": []any{
				map[string]any{
					"group":     "gateway.networking.k8s.io",
					"kind":      "HTTPRoute",
					"namespace": srr.Namespace,
				},
			},
			"to": []any{
				map[string]any{
					"group": "",
					"kind":  "Service",
					"name":  serviceName,
				},
			},
		},
	}}
	desired.SetGroupVersionKind(referenceGrantGVK)

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(referenceGrantGVK)
	err := c.Get(ctx, types.NamespacedName{Namespace: upstreamNS, Name: name}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desired)
	case err != nil:
		return err
	}
	if !unstructuredSpecEqual(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		labels := current.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabel] = ManagedByValue
		labels[InstanceLabel] = srr.Name
		current.SetLabels(labels)
		return c.Update(ctx, current)
	}
	return nil
}

func deleteReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	upstream := srr.Spec.Upstream.ServiceNamespace
	if upstream == "" {
		upstream = srr.Namespace
	}
	if upstream == srr.Namespace {
		return nil
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(referenceGrantGVK)
	obj.SetName(referenceGrantName(srr))
	obj.SetNamespace(upstream)
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete ReferenceGrant %s/%s: %w", upstream, obj.GetName(), err)
	}
	return nil
}

// jwksReferenceGrantName derives a stable ReferenceGrant name in
// CallerJWTJWKSServiceNamespace allowing SecurityPolicy in srr.Namespace to
// reference the caller-jwt-jwks Service (WI-OC-JWKS-RG-01).
func jwksReferenceGrantName(srr *srrv1alpha1.SharedRouteRegistry) string {
	const prefix = "allow-securitypolicy-jwks-"
	raw := prefix + srr.Namespace
	if len(raw) <= 63 {
		return raw
	}
	return strings.TrimRight(raw[:63], "-")
}

func applyJWKSReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	name := jwksReferenceGrantName(srr)
	ns := CallerJWTJWKSServiceNamespace
	desired := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": referenceGrantAPIVersion,
		"kind":       "ReferenceGrant",
		"metadata": map[string]any{
			"name":      name,
			"namespace": ns,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
			},
		},
		"spec": map[string]any{
			"from": []any{
				map[string]any{
					"group":     "gateway.envoyproxy.io",
					"kind":      "SecurityPolicy",
					"namespace": srr.Namespace,
				},
			},
			"to": []any{
				map[string]any{
					"group": "",
					"kind":  "Service",
					"name":  CallerJWTJWKSServiceName,
				},
			},
		},
	}}
	desired.SetGroupVersionKind(referenceGrantGVK)

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(referenceGrantGVK)
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desired)
	case err != nil:
		return err
	}
	if !unstructuredSpecEqual(current.Object["spec"], desired.Object["spec"]) {
		current.Object["spec"] = desired.Object["spec"]
		labels := current.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[ManagedByLabel] = ManagedByValue
		labels[InstanceLabel] = srr.Name
		current.SetLabels(labels)
		return c.Update(ctx, current)
	}
	return nil
}

func deleteJWKSReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return nil
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(referenceGrantGVK)
	obj.SetName(jwksReferenceGrantName(srr))
	obj.SetNamespace(CallerJWTJWKSServiceNamespace)
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete JWKS ReferenceGrant %s/%s: %w", obj.GetNamespace(), obj.GetName(), err)
	}
	return nil
}

func unstructuredSpecEqual(a, b any) bool {
	am, okA := a.(map[string]any)
	bm, okB := b.(map[string]any)
	if !okA || !okB {
		return false
	}
	return fmt.Sprintf("%v", am) == fmt.Sprintf("%v", bm)
}

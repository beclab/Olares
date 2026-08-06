package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

const (
	autheliaBackendSvcFmt = "authelia-backend.user-system-%s"
	// autheliaVerifyPathOverride is the full Authelia verify path (no trailing
	// slash). Must use pathOverride so EG replaces the request path instead of
	// appending via path_prefix (oes parity; not Shared ext-authz).
	autheliaVerifyPathOverride = "/api/verify"
	autheliaBackendPort        = int64(9091)
)

// entranceExtAuthHeaders are forwarded to Authelia on ExtAuth requests
// (aligned with l4 shared ExtAuth request headers + Cookie).
var entranceExtAuthHeaders = []any{
	"Cookie",
	"X-BFL-USER",
	"X-Authorization",
	"X-Forwarded-For",
	"X-Original-URL",
	"X-Original-Method",
	"Authorization",
}

// entranceExtAuthPolicyName returns the ExtAuth SecurityPolicy name for an entrance SRR.
func entranceExtAuthPolicyName(srr *srrv1alpha1.SharedRouteRegistry) string {
	if srr == nil {
		return ""
	}
	return mesh.EntranceExtAuthPolicyName(srr.Name)
}

// desiredEntranceExtAuthPolicy builds Authelia HTTP extAuth for ordinary entrances.
func desiredEntranceExtAuthPolicy(srr *srrv1alpha1.SharedRouteRegistry, owner string) *unstructured.Unstructured {
	routeName := httpRouteName(srr)
	owner = strings.TrimSpace(owner)
	if owner == "" {
		owner = "unknown"
	}
	backendHost := fmt.Sprintf(autheliaBackendSvcFmt, owner)
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "SecurityPolicy",
		"metadata": map[string]any{
			"name":      entranceExtAuthPolicyName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
				"gateway.olares.io/auth-kind": "entrance-ext-auth",
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  routeName,
			},
			"extAuth": map[string]any{
				"failOpen":         false,
				"headersToExtAuth": entranceExtAuthHeaders,
				"http": map[string]any{
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      strings.Split(backendHost, ".")[0],
							"namespace": fmt.Sprintf("user-system-%s", owner),
							"port":      autheliaBackendPort,
						},
					},
					"pathOverride": autheliaVerifyPathOverride,
				},
			},
		},
	}}
}

func ownerFromSRRNamespace(ns string) string {
	const prefix = "user-system-"
	if strings.HasPrefix(ns, prefix) {
		return strings.TrimPrefix(ns, prefix)
	}
	// Application NS is often {app}-{owner}; prefer label when present later.
	parts := strings.Split(ns, "-")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

// applyEntranceExtAuthPolicy creates/updates the entrance Authelia ExtAuth policy
// and the cross-namespace ReferenceGrant to authelia-backend.
func applyEntranceExtAuthPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	owner := ownerFromSRRNamespace(srr.Namespace)
	if srr.Labels != nil {
		if o := strings.TrimSpace(srr.Labels["applications.app.bytetrade.io/owner"]); o != "" {
			owner = o
		}
	}
	desired := desiredEntranceExtAuthPolicy(srr, owner)
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(securityPolicyGVK)
	key := types.NamespacedName{Namespace: srr.Namespace, Name: entranceExtAuthPolicyName(srr)}
	err := c.Get(ctx, key, current)
	if apierrors.IsNotFound(err) {
		desired.SetGroupVersionKind(securityPolicyGVK)
		if err := c.Create(ctx, desired); err != nil {
			klog.Errorf("deenvy: create entrance ExtAuth SecurityPolicy %s/%s failed: %v", key.Namespace, key.Name, err)
			return fmt.Errorf("create entrance ExtAuth SecurityPolicy: %w", err)
		}
		klog.Infof("deenvy: created entrance ExtAuth policy %s/%s", key.Namespace, key.Name)
	} else if err != nil {
		klog.Errorf("deenvy: get entrance ExtAuth SecurityPolicy %s/%s failed: %v", key.Namespace, key.Name, err)
		return fmt.Errorf("get entrance ExtAuth SecurityPolicy: %w", err)
	} else {
		desiredSpec, _, _ := unstructured.NestedMap(desired.Object, "spec")
		currentSpec, _, _ := unstructured.NestedMap(current.Object, "spec")
		if !reflect.DeepEqual(desiredSpec, currentSpec) {
			updated := current.DeepCopy()
			if err := unstructured.SetNestedMap(updated.Object, desiredSpec, "spec"); err != nil {
				return err
			}
			if labels := desired.GetLabels(); labels != nil {
				updated.SetLabels(labels)
			}
			if err := c.Update(ctx, updated); err != nil {
				klog.Errorf("deenvy: update entrance ExtAuth SecurityPolicy %s/%s failed: %v", key.Namespace, key.Name, err)
				return fmt.Errorf("update entrance ExtAuth SecurityPolicy: %w", err)
			}
			klog.Infof("deenvy: updated entrance ExtAuth policy %s/%s", key.Namespace, key.Name)
		}
	}
	if err := applyAutheliaExtAuthReferenceGrant(ctx, c, srr, owner); err != nil {
		klog.Errorf("deenvy: apply Authelia ReferenceGrant for srr=%s/%s owner=%s failed: %v",
			srr.Namespace, srr.Name, owner, err)
		return fmt.Errorf("apply Authelia ReferenceGrant: %w", err)
	}
	return nil
}

func deleteEntranceExtAuthPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return nil
	}
	owner := ownerFromSRRNamespace(srr.Namespace)
	if srr.Labels != nil {
		if o := strings.TrimSpace(srr.Labels["applications.app.bytetrade.io/owner"]); o != "" {
			owner = o
		}
	}
	if err := deleteAutheliaExtAuthReferenceGrant(ctx, c, srr, owner); err != nil {
		return err
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(securityPolicyGVK)
	obj.SetNamespace(srr.Namespace)
	obj.SetName(entranceExtAuthPolicyName(srr))
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// autheliaExtAuthReferenceGrantName is stable in user-system-{owner}.
func autheliaExtAuthReferenceGrantName(fromNS string) string {
	raw := "allow-securitypolicy-authelia-" + fromNS
	if len(raw) <= 63 {
		return raw
	}
	return strings.TrimRight(raw[:63], "-")
}

func applyAutheliaExtAuthReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, owner string) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("authelia owner is empty for srr %s/%s", srr.Namespace, srr.Name)
	}
	ns := fmt.Sprintf("user-system-%s", owner)
	name := autheliaExtAuthReferenceGrantName(srr.Namespace)
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
					"name":  "authelia-backend",
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
		if err := c.Create(ctx, desired); err != nil {
			return fmt.Errorf("create Authelia ReferenceGrant %s/%s: %w", ns, name, err)
		}
		klog.Infof("deenvy: created Authelia ReferenceGrant %s/%s", ns, name)
		return nil
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
		if err := c.Update(ctx, current); err != nil {
			return fmt.Errorf("update Authelia ReferenceGrant %s/%s: %w", ns, name, err)
		}
		klog.Infof("deenvy: updated Authelia ReferenceGrant %s/%s", ns, name)
	}
	return nil
}

func deleteAutheliaExtAuthReferenceGrant(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, owner string) error {
	if srr == nil || strings.TrimSpace(owner) == "" {
		return nil
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(referenceGrantGVK)
	obj.SetName(autheliaExtAuthReferenceGrantName(srr.Namespace))
	obj.SetNamespace(fmt.Sprintf("user-system-%s", owner))
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete Authelia ReferenceGrant %s/%s: %w", obj.GetNamespace(), obj.GetName(), err)
	}
	return nil
}

// needsEntranceExtAuth reports whether this SRR should get Authelia ExtAuth
// (ordinary application entrances) instead of Shared JWT authn.
func needsEntranceExtAuth(srr *srrv1alpha1.SharedRouteRegistry) bool {
	if srr == nil {
		return false
	}
	return srr.Spec.EntranceClass == srrv1alpha1.EntranceClassApplication
}

package routecontrol

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
)

var envoyExtensionPolicyGVK = schema.GroupVersionKind{
	Group:   "gateway.envoyproxy.io",
	Version: "v1alpha1",
	Kind:    "EnvoyExtensionPolicy",
}

func entranceCookiePolicyName(srr *srrv1alpha1.SharedRouteRegistry) string {
	if srr == nil {
		return ""
	}
	return mesh.EntranceCookiePolicyName(srr.Name)
}

// needsEntranceCookie reports whether this SRR needs Cookie Domain EEP (same set as ExtAuth).
func needsEntranceCookie(srr *srrv1alpha1.SharedRouteRegistry) bool {
	return needsEntranceExtAuth(srr)
}

// desiredEntranceCookiePolicy builds EnvoyExtensionPolicy Lua that clears Set-Cookie Domain=.
func desiredEntranceCookiePolicy(srr *srrv1alpha1.SharedRouteRegistry) *unstructured.Unstructured {
	routeName := httpRouteName(srr)
	lua := strings.TrimSpace(sidecar.EnvoySetCookieLua())
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "EnvoyExtensionPolicy",
		"metadata": map[string]any{
			"name":      entranceCookiePolicyName(srr),
			"namespace": srr.Namespace,
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  srr.Name,
				"gateway.olares.io/auth-kind": mesh.AuthKindEntranceCookie,
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  routeName,
			},
			"lua": []any{
				map[string]any{
					"type":   "Inline",
					"inline": lua,
				},
			},
		},
	}}
}

// applyEntranceCookiePolicy creates/updates the entrance Cookie EnvoyExtensionPolicy.
func applyEntranceCookiePolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return fmt.Errorf("srr is nil")
	}
	desired := desiredEntranceCookiePolicy(srr)
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(envoyExtensionPolicyGVK)
	key := types.NamespacedName{Namespace: srr.Namespace, Name: entranceCookiePolicyName(srr)}
	err := c.Get(ctx, key, current)
	if apierrors.IsNotFound(err) {
		desired.SetGroupVersionKind(envoyExtensionPolicyGVK)
		if err := c.Create(ctx, desired); err != nil {
			klog.Errorf("deenvy: create entrance Cookie EEP %s/%s failed: %v", key.Namespace, key.Name, err)
			return fmt.Errorf("create entrance Cookie EnvoyExtensionPolicy: %w", err)
		}
		klog.Infof("deenvy: created entrance Cookie EEP %s/%s", key.Namespace, key.Name)
		return nil
	}
	if err != nil {
		klog.Errorf("deenvy: get entrance Cookie EEP %s/%s failed: %v", key.Namespace, key.Name, err)
		return fmt.Errorf("get entrance Cookie EnvoyExtensionPolicy: %w", err)
	}
	desiredSpec, _, _ := unstructured.NestedMap(desired.Object, "spec")
	currentSpec, _, _ := unstructured.NestedMap(current.Object, "spec")
	if reflect.DeepEqual(desiredSpec, currentSpec) {
		return nil
	}
	updated := current.DeepCopy()
	if err := unstructured.SetNestedMap(updated.Object, desiredSpec, "spec"); err != nil {
		return err
	}
	if labels := desired.GetLabels(); labels != nil {
		updated.SetLabels(labels)
	}
	if err := c.Update(ctx, updated); err != nil {
		klog.Errorf("deenvy: update entrance Cookie EEP %s/%s failed: %v", key.Namespace, key.Name, err)
		return fmt.Errorf("update entrance Cookie EnvoyExtensionPolicy: %w", err)
	}
	klog.Infof("deenvy: updated entrance Cookie EEP %s/%s", key.Namespace, key.Name)
	return nil
}

func deleteEntranceCookiePolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	if srr == nil {
		return nil
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(envoyExtensionPolicyGVK)
	obj.SetNamespace(srr.Namespace)
	obj.SetName(entranceCookiePolicyName(srr))
	if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("deenvy: delete entrance Cookie EEP %s/%s failed: %v", srr.Namespace, entranceCookiePolicyName(srr), err)
		return err
	}
	return nil
}

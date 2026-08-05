package mesh

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

var (
	sharedRouteRegistryGVR = schema.GroupVersionResource{
		Group: "gateway.olares.io", Version: "v1alpha1", Resource: "sharedrouteregistries",
	}
)

const (
	entranceClassApplication = "application"
	authKindEntranceExtAuth  = "entrance-ext-auth"
)

// EntranceExtAuthTarget is one ordinary entrance that must have ExtAuth SecurityPolicy.
type EntranceExtAuthTarget struct {
	Namespace string
	SRRName   string
	// HTTPRoute is the expected SecurityPolicy targetRef.name (usually SRR name).
	HTTPRoute string
}

// EvaluateEntranceExtAuthCovered is the pure ExtAuth coverage predicate.
// Empty targets → true (vacuous). EG Deployment readiness must never feed this.
func EvaluateEntranceExtAuthCovered(targets []EntranceExtAuthTarget, hasMatchingPolicy func(t EntranceExtAuthTarget) bool) bool {
	if hasMatchingPolicy == nil {
		return false
	}
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.SRRName) == "" {
			return false
		}
		if !hasMatchingPolicy(t) {
			return false
		}
	}
	return true
}

// SecurityPolicyMatchesEntranceExtAuth reports whether an object is a valid
// entrance ExtAuth SecurityPolicy for the expected HTTPRoute name.
func SecurityPolicyMatchesEntranceExtAuth(obj *unstructured.Unstructured, expectedHTTPRoute string) bool {
	if obj == nil {
		return false
	}
	if kind, _, _ := unstructured.NestedString(obj.Object, "kind"); kind != "" && kind != "SecurityPolicy" {
		return false
	}
	labels := obj.GetLabels()
	if labels != nil {
		if ak := labels["gateway.olares.io/auth-kind"]; ak != "" && ak != authKindEntranceExtAuth {
			return false
		}
	}
	expectedHTTPRoute = strings.TrimSpace(expectedHTTPRoute)
	if expectedHTTPRoute == "" {
		return true
	}
	name, _, _ := unstructured.NestedString(obj.Object, "spec", "targetRef", "name")
	kind, _, _ := unstructured.NestedString(obj.Object, "spec", "targetRef", "kind")
	if kind != "" && kind != "HTTPRoute" {
		return false
	}
	return name == expectedHTTPRoute
}

// ListEntranceExtAuthTargets lists application-class SRRs that require ExtAuth.
func ListEntranceExtAuthTargets(ctx context.Context, dc dynamic.Interface) ([]EntranceExtAuthTarget, error) {
	if dc == nil {
		return nil, fmt.Errorf("dynamic client required")
	}
	list, err := dc.Resource(sharedRouteRegistryGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list SharedRouteRegistry: %w", err)
	}
	var out []EntranceExtAuthTarget
	for i := range list.Items {
		item := &list.Items[i]
		class, _, _ := unstructured.NestedString(item.Object, "spec", "entranceClass")
		if class != entranceClassApplication {
			continue
		}
		route := item.GetName()
		if statusRoute, ok, _ := unstructured.NestedString(item.Object, "status", "httpRouteName"); ok && strings.TrimSpace(statusRoute) != "" {
			route = statusRoute
		}
		out = append(out, EntranceExtAuthTarget{
			Namespace: item.GetNamespace(),
			SRRName:   item.GetName(),
			HTTPRoute: route,
		})
	}
	return out, nil
}

// GetEntranceExtAuthPolicy fetches the entrance ExtAuth SecurityPolicy for an SRR.
func GetEntranceExtAuthPolicy(ctx context.Context, dc dynamic.Interface, ns, srrName string) (*unstructured.Unstructured, error) {
	if dc == nil || ns == "" || srrName == "" {
		return nil, fmt.Errorf("dynamic client, namespace, and srr name required")
	}
	obj, err := dc.Resource(securityPolicyGVR).Namespace(ns).Get(ctx, EntranceExtAuthPolicyName(srrName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ProbeEntranceExtAuthCovered lists application entrances and verifies each has
// a matching ExtAuth SecurityPolicy (existence + targetRef). Fail-closed on API errors.
func ProbeEntranceExtAuthCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := ListEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		klog.Errorf("mesh: list entrance ExtAuth targets failed: %v", err)
		return false, err
	}
	covered := EvaluateEntranceExtAuthCovered(targets, func(t EntranceExtAuthTarget) bool {
		obj, err := GetEntranceExtAuthPolicy(ctx, dc, t.Namespace, t.SRRName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("mesh: get entrance ExtAuth SecurityPolicy %s/%s failed: %v", t.Namespace, EntranceExtAuthPolicyName(t.SRRName), err)
			} else {
				klog.V(2).Infof("mesh: entrance ExtAuth SecurityPolicy missing %s/%s", t.Namespace, EntranceExtAuthPolicyName(t.SRRName))
			}
			return false
		}
		if !SecurityPolicyMatchesEntranceExtAuth(obj, t.HTTPRoute) {
			klog.Errorf("mesh: entrance ExtAuth SecurityPolicy %s/%s targetRef mismatch want HTTPRoute=%s", t.Namespace, obj.GetName(), t.HTTPRoute)
			return false
		}
		return true
	})
	return covered, nil
}

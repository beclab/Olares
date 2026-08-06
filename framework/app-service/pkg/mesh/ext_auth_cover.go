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
	referenceGrantGVR = schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "referencegrants",
	}
)

const (
	entranceClassApplication     = "application"
	authKindEntranceExtAuth      = "entrance-ext-auth"
	autheliaVerifyPathNormalized = "/api/verify"
)

// EntranceExtAuthTarget is one ordinary entrance that must have ExtAuth SecurityPolicy.
type EntranceExtAuthTarget struct {
	Namespace string
	SRRName   string
	// HTTPRoute is the expected SecurityPolicy targetRef.name (usually SRR name).
	HTTPRoute string
	// Owner is the install owner used to locate user-system-{owner} Authelia Grant.
	Owner string
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

// NormalizeEntranceExtAuthPath returns the canonical Authelia verify path for
// comparison (trim space and trailing slash). Empty input yields "".
func NormalizeEntranceExtAuthPath(path string) string {
	p := strings.TrimSpace(path)
	p = strings.TrimSuffix(p, "/")
	return p
}

// EntranceExtAuthPathOK reports whether the SecurityPolicy CR declares verify
// via pathOverride (preferred) or path, normalized to /api/verify.
func EntranceExtAuthPathOK(obj *unstructured.Unstructured) bool {
	if obj == nil {
		return false
	}
	pathOverride, _, _ := unstructured.NestedString(obj.Object, "spec", "extAuth", "http", "pathOverride")
	path, _, _ := unstructured.NestedString(obj.Object, "spec", "extAuth", "http", "path")
	p := NormalizeEntranceExtAuthPath(pathOverride)
	if p == "" {
		p = NormalizeEntranceExtAuthPath(path)
	}
	return p == autheliaVerifyPathNormalized
}

// EntranceExtAuthAcceptedOK is fail-closed when status says not Accepted.
// Missing status is allowed (policy still reconciling).
func EntranceExtAuthAcceptedOK(obj *unstructured.Unstructured) bool {
	if obj == nil {
		return false
	}
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !found || len(conditions) == 0 {
		return true
	}
	for _, item := range conditions {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		typ, _ := m["type"].(string)
		status, _ := m["status"].(string)
		if typ == "Accepted" && status != "True" {
			return false
		}
	}
	return true
}

// SecurityPolicyMatchesEntranceExtAuth reports whether an object is a valid
// entrance ExtAuth SecurityPolicy for the expected HTTPRoute name, with verify path.
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
	if !EntranceExtAuthPathOK(obj) {
		return false
	}
	if !EntranceExtAuthAcceptedOK(obj) {
		return false
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

// AutheliaExtAuthReferenceGrantName matches routecontrol grant naming.
func AutheliaExtAuthReferenceGrantName(fromNS string) string {
	raw := "allow-securitypolicy-authelia-" + fromNS
	if len(raw) <= 63 {
		return raw
	}
	return strings.TrimRight(raw[:63], "-")
}

// ownerFromAppNamespace extracts owner suffix from app or user-system namespaces.
func ownerFromAppNamespace(ns string) string {
	const prefix = "user-system-"
	if strings.HasPrefix(ns, prefix) {
		return strings.TrimPrefix(ns, prefix)
	}
	parts := strings.Split(ns, "-")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

// HasAutheliaExtAuthReferenceGrant reports whether the Authelia Grant exists.
func HasAutheliaExtAuthReferenceGrant(ctx context.Context, dc dynamic.Interface, fromNS, owner string) bool {
	if dc == nil || fromNS == "" || owner == "" {
		return false
	}
	ns := "user-system-" + owner
	name := AutheliaExtAuthReferenceGrantName(fromNS)
	_, err := dc.Resource(referenceGrantGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("mesh: get Authelia ReferenceGrant %s/%s failed: %v", ns, name, err)
		}
		return false
	}
	return true
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
		authLevel, _, _ := unstructured.NestedString(item.Object, "spec", "authLevel")
		if strings.EqualFold(strings.TrimSpace(authLevel), "public") {
			// public entrances have no ExtAuth PEP; treat as covered elsewhere.
			continue
		}
		route := item.GetName()
		if statusRoute, ok, _ := unstructured.NestedString(item.Object, "status", "httpRouteName"); ok && strings.TrimSpace(statusRoute) != "" {
			route = statusRoute
		}
		owner := ""
		if labels := item.GetLabels(); labels != nil {
			owner = strings.TrimSpace(labels["applications.app.bytetrade.io/owner"])
		}
		if owner == "" {
			owner = ownerFromAppNamespace(item.GetNamespace())
		}
		out = append(out, EntranceExtAuthTarget{
			Namespace: item.GetNamespace(),
			SRRName:   item.GetName(),
			HTTPRoute: route,
			Owner:     owner,
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
// a matching ExtAuth SecurityPolicy (existence + targetRef + path + Accepted)
// and Authelia ReferenceGrant. Fail-closed on API errors.
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
			klog.Errorf("mesh: entrance ExtAuth SecurityPolicy %s/%s mismatch want HTTPRoute=%s path=/api/verify", t.Namespace, obj.GetName(), t.HTTPRoute)
			return false
		}
		if t.Owner == "" || !HasAutheliaExtAuthReferenceGrant(ctx, dc, t.Namespace, t.Owner) {
			klog.Errorf("mesh: Authelia ReferenceGrant missing for entrance ExtAuth ns=%s owner=%s", t.Namespace, t.Owner)
			return false
		}
		return true
	})
	return covered, nil
}

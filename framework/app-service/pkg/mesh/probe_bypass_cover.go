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

const (
	EntranceProbeBypassRouteSuffix  = "-probe-bypass"
	EntranceProbeBypassPolicySuffix = "-entrance-probe"
	AuthKindEntranceProbeBypass     = "entrance-probe-bypass"
	ProbeUARequiredLabel            = "gateway.olares.io/probe-ua-required"
	ProbePathsAnnotation            = "gateway.olares.io/probe-paths"
)

var httpRouteGVR = schema.GroupVersionResource{
	Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
}

// EntranceProbeBypassRouteName is the companion HTTPRoute without ExtAuth.
func EntranceProbeBypassRouteName(srrName string) string {
	return srrName + EntranceProbeBypassRouteSuffix
}

// EntranceProbeBypassPolicyName is the UA-guard EnvoyExtensionPolicy for probe routes.
func EntranceProbeBypassPolicyName(srrName string) string {
	return srrName + EntranceProbeBypassPolicySuffix
}

// EntranceProbeBypassTarget is one ordinary entrance that may need probe bypass.
type EntranceProbeBypassTarget struct {
	Namespace string
	SRRName   string
	// ExpectedPaths empty → vacuous Ready for this entrance.
	ExpectedPaths []string
}

// EvaluateEntranceProbeBypassReady is the pure probe-bypass predicate.
func EvaluateEntranceProbeBypassReady(targets []EntranceProbeBypassTarget, ready func(t EntranceProbeBypassTarget) bool) bool {
	if ready == nil {
		return false
	}
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.SRRName) == "" {
			return false
		}
		if len(t.ExpectedPaths) == 0 {
			continue
		}
		if !ready(t) {
			return false
		}
	}
	return true
}

// ProbeBypassObjectsReady reports whether route+EEP exist with UA guard wired.
func ProbeBypassObjectsReady(route, policy *unstructured.Unstructured, expectedPaths []string) bool {
	if route == nil || policy == nil {
		return false
	}
	if labels := policy.GetLabels(); labels == nil || labels[ProbeUARequiredLabel] != "true" {
		return false
	}
	lua, found, _ := unstructured.NestedSlice(policy.Object, "spec", "lua")
	if !found || len(lua) == 0 {
		return false
	}
	ann := route.GetAnnotations()
	if ann == nil {
		return false
	}
	got := strings.Split(ann[ProbePathsAnnotation], ",")
	want := map[string]struct{}{}
	for _, p := range expectedPaths {
		want[p] = struct{}{}
	}
	for _, p := range got {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		delete(want, p)
	}
	return len(want) == 0
}

// ListEntranceProbeBypassTargets lists application-class SRRs (paths filled by caller via ExpectedPaths).
func ListEntranceProbeBypassTargets(ctx context.Context, dc dynamic.Interface) ([]EntranceProbeBypassTarget, error) {
	ext, err := ListEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		return nil, err
	}
	out := make([]EntranceProbeBypassTarget, 0, len(ext))
	for _, t := range ext {
		out = append(out, EntranceProbeBypassTarget{
			Namespace: t.Namespace,
			SRRName:   t.SRRName,
		})
	}
	return out, nil
}

// ProbeEntranceProbeBypassReady verifies probe bypass route+UA EEP for targets with ExpectedPaths.
func ProbeEntranceProbeBypassReady(ctx context.Context, dc dynamic.Interface, targets []EntranceProbeBypassTarget) (bool, error) {
	if dc == nil {
		return false, fmt.Errorf("dynamic client required")
	}
	return EvaluateEntranceProbeBypassReady(targets, func(t EntranceProbeBypassTarget) bool {
		route, err := dc.Resource(httpRouteGVR).Namespace(t.Namespace).Get(ctx, EntranceProbeBypassRouteName(t.SRRName), metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("mesh: get probe HTTPRoute %s/%s failed: %v", t.Namespace, EntranceProbeBypassRouteName(t.SRRName), err)
			}
			return false
		}
		pol, err := dc.Resource(envoyExtensionPolicyGVR).Namespace(t.Namespace).Get(ctx, EntranceProbeBypassPolicyName(t.SRRName), metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("mesh: get probe EEP %s/%s failed: %v", t.Namespace, EntranceProbeBypassPolicyName(t.SRRName), err)
			}
			return false
		}
		return ProbeBypassObjectsReady(route, pol, t.ExpectedPaths)
	}), nil
}

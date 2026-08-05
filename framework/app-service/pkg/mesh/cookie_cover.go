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
	EntranceCookiePolicySuffix = "-entrance-cookie"
	AuthKindEntranceCookie     = "entrance-cookie"
)

var envoyExtensionPolicyGVR = schema.GroupVersionResource{
	Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "envoyextensionpolicies",
}

// EntranceCookiePolicyName returns the Cookie EnvoyExtensionPolicy object name.
func EntranceCookiePolicyName(srrName string) string {
	return srrName + EntranceCookiePolicySuffix
}

// EntranceCookieTarget is one ordinary entrance that must have Cookie EEP.
type EntranceCookieTarget struct {
	Namespace string
	SRRName   string
	HTTPRoute string
}

// EvaluateEntranceCookieCovered is the pure Cookie coverage predicate.
func EvaluateEntranceCookieCovered(targets []EntranceCookieTarget, hasMatchingPolicy func(t EntranceCookieTarget) bool) bool {
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

// EnvoyExtensionPolicyMatchesEntranceCookie reports a valid Cookie EEP for HTTPRoute.
func EnvoyExtensionPolicyMatchesEntranceCookie(obj *unstructured.Unstructured, expectedHTTPRoute string) bool {
	if obj == nil {
		return false
	}
	if kind, _, _ := unstructured.NestedString(obj.Object, "kind"); kind != "" && kind != "EnvoyExtensionPolicy" {
		return false
	}
	if labels := obj.GetLabels(); labels != nil {
		if ak := labels["gateway.olares.io/auth-kind"]; ak != "" && ak != AuthKindEntranceCookie {
			return false
		}
	}
	lua, found, _ := unstructured.NestedSlice(obj.Object, "spec", "lua")
	if !found || len(lua) == 0 {
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

// ListEntranceCookieTargets lists application-class SRRs that require Cookie EEP.
func ListEntranceCookieTargets(ctx context.Context, dc dynamic.Interface) ([]EntranceCookieTarget, error) {
	extTargets, err := ListEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		return nil, err
	}
	out := make([]EntranceCookieTarget, 0, len(extTargets))
	for _, t := range extTargets {
		out = append(out, EntranceCookieTarget{
			Namespace: t.Namespace,
			SRRName:   t.SRRName,
			HTTPRoute: t.HTTPRoute,
		})
	}
	return out, nil
}

// GetEntranceCookiePolicy fetches the Cookie EnvoyExtensionPolicy for an SRR.
func GetEntranceCookiePolicy(ctx context.Context, dc dynamic.Interface, ns, srrName string) (*unstructured.Unstructured, error) {
	if dc == nil || ns == "" || srrName == "" {
		return nil, fmt.Errorf("dynamic client, namespace, and srr name required")
	}
	return dc.Resource(envoyExtensionPolicyGVR).Namespace(ns).Get(ctx, EntranceCookiePolicyName(srrName), metav1.GetOptions{})
}

// ProbeEntranceCookieCovered verifies each application entrance has Cookie EEP.
func ProbeEntranceCookieCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := ListEntranceCookieTargets(ctx, dc)
	if err != nil {
		klog.Errorf("mesh: list entrance Cookie targets failed: %v", err)
		return false, err
	}
	covered := EvaluateEntranceCookieCovered(targets, func(t EntranceCookieTarget) bool {
		obj, err := GetEntranceCookiePolicy(ctx, dc, t.Namespace, t.SRRName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("mesh: get entrance Cookie EEP %s/%s failed: %v", t.Namespace, EntranceCookiePolicyName(t.SRRName), err)
			} else {
				klog.V(2).Infof("mesh: entrance Cookie EEP missing %s/%s", t.Namespace, EntranceCookiePolicyName(t.SRRName))
			}
			return false
		}
		if !EnvoyExtensionPolicyMatchesEntranceCookie(obj, t.HTTPRoute) {
			klog.Errorf("mesh: entrance Cookie EEP %s/%s targetRef/lua mismatch want HTTPRoute=%s", t.Namespace, obj.GetName(), t.HTTPRoute)
			return false
		}
		return true
	})
	return covered, nil
}

package mesh

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

const (
	AuthKindLabel               = "gateway.olares.io/auth-kind"
	AuthKindEntranceExtAuth     = "entrance-ext-auth"
	AuthKindEntranceCookie      = "entrance-cookie"
	AuthKindEntranceProbeBypass = "entrance-probe-bypass"

	EntranceCookiePolicySuffix      = "-entrance-cookie"
	EntranceProbeBypassRouteSuffix  = "-probe-bypass"
	EntranceProbeBypassPolicySuffix = "-entrance-probe"
)

var (
	envoySecurityPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "securitypolicies",
	}
	envoyExtensionPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "envoyextensionpolicies",
	}
	httpRouteGVRForGC = schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}
)

// IsEntranceEGPEPObject reports whether an object is a north-south EG login PEP
// leftover (ExtAuth / Cookie EEP / probe-bypass route) that EDGE must garbage-collect.
func IsEntranceEGPEPObject(name string, labels map[string]string) bool {
	if labels != nil {
		switch labels[AuthKindLabel] {
		case AuthKindEntranceExtAuth, AuthKindEntranceCookie, AuthKindEntranceProbeBypass:
			return true
		}
	}
	n := strings.TrimSpace(name)
	if n == "" {
		return false
	}
	return strings.HasSuffix(n, EntranceExtAuthPolicySuffix) ||
		strings.HasSuffix(n, EntranceCookiePolicySuffix) ||
		strings.HasSuffix(n, EntranceProbeBypassRouteSuffix) ||
		strings.HasSuffix(n, EntranceProbeBypassPolicySuffix)
}

// EvaluateNoEntranceEGExtAuth is the pure M-EG predicate: no leftover objects.
func EvaluateNoEntranceEGExtAuth(leftoverCount int) bool {
	return leftoverCount == 0
}

// GarbageCollectEntranceEGPEP deletes cluster-wide EG entrance login PEP leftovers.
// Shared JWT SecurityPolicy (-jwt-authn) is intentionally retained.
func GarbageCollectEntranceEGPEP(ctx context.Context, dc dynamic.Interface) (deleted int, err error) {
	if dc == nil {
		return 0, fmt.Errorf("dynamic client required")
	}
	for _, gvr := range []schema.GroupVersionResource{
		envoySecurityPolicyGVR,
		envoyExtensionPolicyGVR,
		httpRouteGVRForGC,
	} {
		n, e := gcEntranceEGPEPResource(ctx, dc, gvr)
		deleted += n
		if e != nil {
			// CRD may be absent on clusters without EG; treat NotFound as empty.
			if apierrors.IsNotFound(e) {
				klog.V(2).Infof("mesh: EG PEP GC skip missing CRD %s: %v", gvr.String(), e)
				continue
			}
			return deleted, e
		}
	}
	return deleted, nil
}

func gcEntranceEGPEPResource(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource) (int, error) {
	list, err := dc.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	deleted := 0
	for i := range list.Items {
		obj := &list.Items[i]
		if !IsEntranceEGPEPObject(obj.GetName(), obj.GetLabels()) {
			continue
		}
		// Never delete Shared JWT authn SecurityPolicy.
		if gvr == envoySecurityPolicyGVR && strings.HasSuffix(obj.GetName(), SecurityPolicyJWTSuffix()) {
			continue
		}
		ns := obj.GetNamespace()
		name := obj.GetName()
		if err := dc.Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			klog.Errorf("mesh: EG PEP GC delete %s %s/%s failed: %v", gvr.Resource, ns, name, err)
			return deleted, err
		}
		deleted++
		klog.Infof("mesh: EG PEP GC deleted %s %s/%s", gvr.Resource, ns, name)
	}
	return deleted, nil
}

// SecurityPolicyJWTSuffix is the Shared JWT SecurityPolicy name suffix (retained).
func SecurityPolicyJWTSuffix() string {
	return "-jwt-authn"
}

// CountEntranceEGPEPLeftovers counts EG north-south login PEP objects (read-only).
func CountEntranceEGPEPLeftovers(ctx context.Context, dc dynamic.Interface) (int, error) {
	if dc == nil {
		return 0, fmt.Errorf("dynamic client required")
	}
	total := 0
	for _, gvr := range []schema.GroupVersionResource{
		envoySecurityPolicyGVR,
		envoyExtensionPolicyGVR,
		httpRouteGVRForGC,
	} {
		list, err := dc.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return total, err
		}
		for i := range list.Items {
			obj := &list.Items[i]
			if gvr == envoySecurityPolicyGVR && strings.HasSuffix(obj.GetName(), SecurityPolicyJWTSuffix()) {
				continue
			}
			if IsEntranceEGPEPObject(obj.GetName(), obj.GetLabels()) {
				total++
			}
		}
	}
	return total, nil
}

// EnsureNoEntranceEGPEP runs GC then reports whether M-EG is satisfied.
func EnsureNoEntranceEGPEP(ctx context.Context, dc dynamic.Interface) (bool, string, error) {
	if _, err := GarbageCollectEntranceEGPEP(ctx, dc); err != nil {
		return false, "", err
	}
	n, err := CountEntranceEGPEPLeftovers(ctx, dc)
	if err != nil {
		return false, "", err
	}
	ok := EvaluateNoEntranceEGExtAuth(n)
	if !ok {
		return false, fmt.Sprintf("EG entrance PEP leftovers=%d", n), nil
	}
	return true, "no entrance EG PEP leftovers", nil
}

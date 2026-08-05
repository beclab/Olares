package upgrade

import (
	"context"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

var deenvyEnvoyExtensionPolicyGVR = schema.GroupVersionResource{
	Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "envoyextensionpolicies",
}

const (
	deenvyEntranceCookieSuffix = "-entrance-cookie"
	deenvyAuthKindCookie       = "entrance-cookie"
)

type entranceCookieTarget struct {
	Namespace string
	SRRName   string
	HTTPRoute string
}

func evaluateEntranceCookieCovered(targets []entranceCookieTarget, hasMatchingPolicy func(t entranceCookieTarget) bool) bool {
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

func envoyExtensionPolicyMatchesEntranceCookie(obj *unstructured.Unstructured, expectedHTTPRoute string) bool {
	if obj == nil {
		return false
	}
	if labels := obj.GetLabels(); labels != nil {
		if ak := labels["gateway.olares.io/auth-kind"]; ak != "" && ak != deenvyAuthKindCookie {
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

func listEntranceCookieTargets(ctx context.Context, dc dynamic.Interface) ([]entranceCookieTarget, error) {
	ext, err := listEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		return nil, err
	}
	out := make([]entranceCookieTarget, 0, len(ext))
	for _, t := range ext {
		out = append(out, entranceCookieTarget{
			Namespace: t.Namespace,
			SRRName:   t.SRRName,
			HTTPRoute: t.HTTPRoute,
		})
	}
	return out, nil
}

func probeEntranceCookieCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := listEntranceCookieTargets(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: list entrance Cookie targets failed: %v", err)
		return false, err
	}
	covered := evaluateEntranceCookieCovered(targets, func(t entranceCookieTarget) bool {
		name := t.SRRName + deenvyEntranceCookieSuffix
		obj, err := dc.Resource(deenvyEnvoyExtensionPolicyGVR).Namespace(t.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Errorf("deenvy: get Cookie EEP %s/%s failed: %v", t.Namespace, name, err)
			}
			return false
		}
		if !envoyExtensionPolicyMatchesEntranceCookie(obj, t.HTTPRoute) {
			logger.Errorf("deenvy: Cookie EEP %s/%s mismatch want HTTPRoute=%s", t.Namespace, name, t.HTTPRoute)
			return false
		}
		return true
	})
	return covered, nil
}

// assignCookieDepConditions records Cookie coverage independently of EG readiness.
func assignCookieDepConditions(conds map[string]bool, cookieCovered bool) {
	if conds == nil {
		return
	}
	conds["EntranceCookieCovered"] = cookieCovered
}

func ensureCookieProbeOK(ctx context.Context, dc dynamic.Interface, conds map[string]bool) bool {
	if dc == nil {
		assignCookieDepConditions(conds, false)
		return false
	}
	ok, err := probeEntranceCookieCovered(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: EntranceCookieCovered probe failed: %v", err)
		ok = false
	}
	assignCookieDepConditions(conds, ok)
	return ok
}
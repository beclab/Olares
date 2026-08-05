package upgrade

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

var (
	deenvySRR_GVR = schema.GroupVersionResource{
		Group: "gateway.olares.io", Version: "v1alpha1", Resource: "sharedrouteregistries",
	}
	deenvySecurityPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "securitypolicies",
	}
)

const (
	deenvyEntranceExtAuthSuffix = "-entrance-ext-auth"
	deenvyEntranceClassApp      = "application"
	deenvyAuthKindExtAuth       = "entrance-ext-auth"
)

type entranceExtAuthTarget struct {
	Namespace string
	SRRName   string
	HTTPRoute string
}

// evaluateEntranceExtAuthCovered is the pure ExtAuth coverage predicate.
// Empty targets → true. EG Deployment readiness must never feed this.
func evaluateEntranceExtAuthCovered(targets []entranceExtAuthTarget, hasMatchingPolicy func(t entranceExtAuthTarget) bool) bool {
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

func securityPolicyMatchesEntranceExtAuth(obj *unstructured.Unstructured, expectedHTTPRoute string) bool {
	if obj == nil {
		return false
	}
	if labels := obj.GetLabels(); labels != nil {
		if ak := labels["gateway.olares.io/auth-kind"]; ak != "" && ak != deenvyAuthKindExtAuth {
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

func listEntranceExtAuthTargets(ctx context.Context, dc dynamic.Interface) ([]entranceExtAuthTarget, error) {
	if dc == nil {
		return nil, fmt.Errorf("dynamic client required")
	}
	list, err := dc.Resource(deenvySRR_GVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list SharedRouteRegistry: %w", err)
	}
	var out []entranceExtAuthTarget
	for i := range list.Items {
		item := &list.Items[i]
		class, _, _ := unstructured.NestedString(item.Object, "spec", "entranceClass")
		if class != deenvyEntranceClassApp {
			continue
		}
		route := item.GetName()
		if statusRoute, ok, _ := unstructured.NestedString(item.Object, "status", "httpRouteName"); ok && strings.TrimSpace(statusRoute) != "" {
			route = statusRoute
		}
		out = append(out, entranceExtAuthTarget{
			Namespace: item.GetNamespace(),
			SRRName:   item.GetName(),
			HTTPRoute: route,
		})
	}
	return out, nil
}

// probeEntranceExtAuthCovered verifies each application entrance has ExtAuth SecurityPolicy.
func probeEntranceExtAuthCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := listEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: list entrance ExtAuth targets failed: %v", err)
		return false, err
	}
	covered := evaluateEntranceExtAuthCovered(targets, func(t entranceExtAuthTarget) bool {
		name := t.SRRName + deenvyEntranceExtAuthSuffix
		obj, err := dc.Resource(deenvySecurityPolicyGVR).Namespace(t.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Errorf("deenvy: get ExtAuth SecurityPolicy %s/%s failed: %v", t.Namespace, name, err)
			}
			return false
		}
		if !securityPolicyMatchesEntranceExtAuth(obj, t.HTTPRoute) {
			logger.Errorf("deenvy: ExtAuth SecurityPolicy %s/%s targetRef mismatch want HTTPRoute=%s", t.Namespace, name, t.HTTPRoute)
			return false
		}
		return true
	})
	return covered, nil
}

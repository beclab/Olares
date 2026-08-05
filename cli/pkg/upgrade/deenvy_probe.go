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

var (
	deenvyHTTPRouteGVR = schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}
)

const (
	deenvyProbeRouteSuffix  = "-probe-bypass"
	deenvyProbePolicySuffix = "-entrance-probe"
	deenvyProbeUALabel      = "gateway.olares.io/probe-ua-required"
	deenvyProbePathsAnn     = "gateway.olares.io/probe-paths"
)

func assignProbeBypassDepConditions(conds map[string]bool, ready bool) {
	if conds == nil {
		return
	}
	conds["EntranceProbeBypassReady"] = ready
}

func probeBypassObjectsReady(route, policy *unstructured.Unstructured) bool {
	if route == nil || policy == nil {
		return false
	}
	if labels := policy.GetLabels(); labels == nil || labels[deenvyProbeUALabel] != "true" {
		return false
	}
	lua, found, _ := unstructured.NestedSlice(policy.Object, "spec", "lua")
	if !found || len(lua) == 0 {
		return false
	}
	ann := route.GetAnnotations()
	if ann == nil || strings.TrimSpace(ann[deenvyProbePathsAnn]) == "" {
		return false
	}
	return true
}

// probeEntranceProbeBypassReady: application SRRs without a probe-bypass route are OK
// only when no probe-paths annotation is expected yet; if a probe-bypass route exists it
// must have UA EEP. Vacuous true when no application SRRs or none have probe routes.
func probeEntranceProbeBypassReady(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := listEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: list SRR for probe bypass failed: %v", err)
		return false, err
	}
	for _, t := range targets {
		routeName := t.SRRName + deenvyProbeRouteSuffix
		route, err := dc.Resource(deenvyHTTPRouteGVR).Namespace(t.Namespace).Get(ctx, routeName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue // no HTTP probes discovered → vacuous OK for this entrance
		}
		if err != nil {
			logger.Errorf("deenvy: get probe HTTPRoute %s/%s failed: %v", t.Namespace, routeName, err)
			return false, err
		}
		polName := t.SRRName + deenvyProbePolicySuffix
		pol, err := dc.Resource(deenvyEnvoyExtensionPolicyGVR).Namespace(t.Namespace).Get(ctx, polName, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Errorf("deenvy: get probe EEP %s/%s failed: %v", t.Namespace, polName, err)
			}
			return false, nil
		}
		if !probeBypassObjectsReady(route, pol) {
			return false, nil
		}
	}
	return true, nil
}

func ensureProbeBypassProbeOK(ctx context.Context, dc dynamic.Interface, conds map[string]bool) bool {
	if dc == nil {
		assignProbeBypassDepConditions(conds, false)
		return false
	}
	ok, err := probeEntranceProbeBypassReady(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: EntranceProbeBypassReady probe failed: %v", err)
		ok = false
	}
	assignProbeBypassDepConditions(conds, ok)
	return ok
}

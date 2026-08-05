package upgrade

import (
	"context"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

const (
	deenvyImagesRouteSuffix     = "-images-upload"
	deenvyAuxCapsAnn            = "gateway.olares.io/aux-capabilities"
	deenvyAuxExtAuthBypassAnn   = "gateway.olares.io/aux-extauth-bypass"
	deenvyAuthKindImages        = "entrance-images"
	deenvyAuxImagesSvc          = "tapr-images-svc"
	deenvyAuxHelperSuffix       = "-deenvy-aux"
)

type entranceAuxNeeds struct {
	WS     bool
	Upload bool
	Images bool
}

func (n entranceAuxNeeds) vacuous() bool {
	return !n.WS && !n.Upload && !n.Images
}

type entranceAuxTarget struct {
	Namespace string
	SRRName   string
	HTTPRoute string
	Needs     entranceAuxNeeds
}

func assignAuxDepConditions(conds map[string]bool, covered bool) {
	if conds == nil {
		return
	}
	conds["EntranceAuxCovered"] = covered
}

func parseAuxCaps(v string) entranceAuxNeeds {
	var n entranceAuxNeeds
	for _, p := range strings.Split(v, ",") {
		switch strings.TrimSpace(p) {
		case "ws":
			n.WS = true
		case "upload":
			n.Upload = true
		case "images":
			n.Images = true
		}
	}
	return n
}

func evaluateEntranceAuxCovered(targets []entranceAuxTarget, ready func(t entranceAuxTarget) bool) bool {
	if ready == nil {
		return false
	}
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.SRRName) == "" {
			return false
		}
		if t.Needs.vacuous() {
			continue
		}
		if !ready(t) {
			return false
		}
	}
	return true
}

func httpRouteHasAuxPath(route *unstructured.Unstructured, pathPrefix, svcSuffixOrName string, port int64) bool {
	if route == nil {
		return false
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	for _, r := range rules {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		matches, _, _ := unstructured.NestedSlice(rm, "matches")
		pathOK := false
		for _, m := range matches {
			mm, ok := m.(map[string]any)
			if !ok {
				continue
			}
			pval, _, _ := unstructured.NestedString(mm, "path", "value")
			ptype, _, _ := unstructured.NestedString(mm, "path", "type")
			if pval == pathPrefix && (ptype == "PathPrefix" || ptype == "") {
				pathOK = true
				break
			}
		}
		if !pathOK {
			continue
		}
		backends, _, _ := unstructured.NestedSlice(rm, "backendRefs")
		for _, b := range backends {
			bm, ok := b.(map[string]any)
			if !ok {
				continue
			}
			name, _ := bm["name"].(string)
			p, _, _ := unstructured.NestedInt64(bm, "port")
			if port > 0 && p != port {
				continue
			}
			if svcSuffixOrName == deenvyAuxImagesSvc {
				if name == deenvyAuxImagesSvc {
					return true
				}
				continue
			}
			if strings.HasSuffix(name, svcSuffixOrName) {
				return true
			}
		}
	}
	return false
}

func imagesRouteBypassesExtAuth(route *unstructured.Unstructured) bool {
	if route == nil {
		return false
	}
	labels := route.GetLabels()
	if labels == nil || labels["gateway.olares.io/auth-kind"] != deenvyAuthKindImages {
		return false
	}
	ann := route.GetAnnotations()
	if ann == nil || ann[deenvyAuxExtAuthBypassAnn] != "true" {
		return false
	}
	return httpRouteHasAuxPath(route, "/images/upload", deenvyAuxImagesSvc, 8080)
}

func auxTargetReady(t entranceAuxTarget, main, images *unstructured.Unstructured) bool {
	if t.Needs.WS && !httpRouteHasAuxPath(main, "/ws", deenvyAuxHelperSuffix, 40010) {
		return false
	}
	if t.Needs.Upload && !httpRouteHasAuxPath(main, "/upload", deenvyAuxHelperSuffix, 40030) {
		return false
	}
	if t.Needs.Images && !imagesRouteBypassesExtAuth(images) {
		return false
	}
	return true
}

func probeEntranceAuxCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := listEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: list SRR for aux failed: %v", err)
		return false, err
	}
	auxTargets := make([]entranceAuxTarget, 0, len(targets))
	for _, t := range targets {
		needs := entranceAuxNeeds{Images: true}
		route, err := dc.Resource(deenvyHTTPRouteGVR).Namespace(t.Namespace).Get(ctx, t.HTTPRoute, metav1.GetOptions{})
		if err == nil && route != nil {
			if ann := route.GetAnnotations(); ann != nil {
				parsed := parseAuxCaps(ann[deenvyAuxCapsAnn])
				needs.WS = parsed.WS
				needs.Upload = parsed.Upload
				if parsed.Images {
					needs.Images = true
				}
			}
		}
		auxTargets = append(auxTargets, entranceAuxTarget{
			Namespace: t.Namespace,
			SRRName:   t.SRRName,
			HTTPRoute: t.HTTPRoute,
			Needs:     needs,
		})
	}
	return evaluateEntranceAuxCovered(auxTargets, func(t entranceAuxTarget) bool {
		main, err := dc.Resource(deenvyHTTPRouteGVR).Namespace(t.Namespace).Get(ctx, t.HTTPRoute, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Errorf("deenvy: get main HTTPRoute %s/%s for aux: %v", t.Namespace, t.HTTPRoute, err)
			}
			return false
		}
		var images *unstructured.Unstructured
		if t.Needs.Images {
			images, err = dc.Resource(deenvyHTTPRouteGVR).Namespace(t.Namespace).Get(ctx, t.SRRName+deenvyImagesRouteSuffix, metav1.GetOptions{})
			if err != nil {
				return false
			}
		}
		return auxTargetReady(t, main, images)
	}), nil
}

func ensureAuxProbeOK(ctx context.Context, dc dynamic.Interface, conds map[string]bool) bool {
	if dc == nil {
		assignAuxDepConditions(conds, false)
		return false
	}
	ok, err := probeEntranceAuxCovered(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: EntranceAuxCovered probe failed: %v", err)
		ok = false
	}
	assignAuxDepConditions(conds, ok)
	return ok
}

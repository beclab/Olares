package mesh

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

const (
	EntranceImagesRouteSuffix  = "-images-upload"
	AuthKindEntranceAux        = "entrance-aux"
	AuthKindEntranceImages     = "entrance-images"
	AuxCapabilitiesAnnotation  = "gateway.olares.io/aux-capabilities"
	AuxExtAuthBypassAnnotation = "gateway.olares.io/aux-extauth-bypass"
	AuxWSPathPrefix            = "/ws"
	AuxUploadPathPrefix        = "/upload"
	AuxImagesPathPrefix        = "/images/upload"
	AuxImagesServiceName       = "tapr-images-svc"
	AuxHelperServiceNameSuffix = "-deenvy-aux"
)

// EntranceImagesRouteName is the companion HTTPRoute for /images/upload (no ExtAuth).
func EntranceImagesRouteName(srrName string) string {
	return srrName + EntranceImagesRouteSuffix
}

// EntranceAuxNeeds declares which Aux paths an entrance requires.
type EntranceAuxNeeds struct {
	WS     bool
	Upload bool
	Images bool
}

// Vacuous reports whether no Aux capability is required.
func (n EntranceAuxNeeds) Vacuous() bool {
	return !n.WS && !n.Upload && !n.Images
}

// ParseAuxCapabilitiesAnnotation parses gateway.olares.io/aux-capabilities.
func ParseAuxCapabilitiesAnnotation(v string) EntranceAuxNeeds {
	var n EntranceAuxNeeds
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

// EntranceAuxTarget is one ordinary entrance evaluated for AuxCovered.
type EntranceAuxTarget struct {
	Namespace       string
	SRRName         string
	HTTPRoute       string
	UpstreamSvc     string // for helper Service name; empty → skip Service check
	Needs           EntranceAuxNeeds
	ImagesNamespace string // user-system-{owner}; empty → skip ns check on images backend
}

// EvaluateEntranceAuxCovered is the pure Aux coverage predicate.
// Vacuous needs → that entrance is Ready; missing callback → false.
func EvaluateEntranceAuxCovered(targets []EntranceAuxTarget, ready func(t EntranceAuxTarget) bool) bool {
	if ready == nil {
		return false
	}
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.SRRName) == "" {
			return false
		}
		if t.Needs.Vacuous() {
			continue
		}
		if !ready(t) {
			return false
		}
	}
	return true
}

// HTTPRouteHasAuxPathRule reports PathPrefix match to a *-deenvy-aux Service (ws/upload)
// or to tapr-images-svc (images).
func HTTPRouteHasAuxPathRule(route *unstructured.Unstructured, pathPrefix, wantSvcSuffixOrName string, port int64) bool {
	if route == nil || pathPrefix == "" {
		return false
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	for _, r := range rules {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		if !ruleMatchesPathPrefix(rm, pathPrefix) {
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
			if wantSvcSuffixOrName == AuxImagesServiceName {
				if name == AuxImagesServiceName {
					return true
				}
				continue
			}
			if strings.HasSuffix(name, wantSvcSuffixOrName) {
				return true
			}
		}
	}
	return false
}

func ruleMatchesPathPrefix(rule map[string]any, pathPrefix string) bool {
	matches, _, _ := unstructured.NestedSlice(rule, "matches")
	for _, m := range matches {
		mm, ok := m.(map[string]any)
		if !ok {
			continue
		}
		ptype, _, _ := unstructured.NestedString(mm, "path", "type")
		pvalue, _, _ := unstructured.NestedString(mm, "path", "value")
		if (ptype == "PathPrefix" || ptype == "") && pvalue == pathPrefix {
			return true
		}
	}
	return false
}

// ImagesRouteBypassesExtAuth reports the /images/upload companion route is labeled
// and annotated so it does not require ExtAuth (matches legacy edge behavior).
func ImagesRouteBypassesExtAuth(route *unstructured.Unstructured) bool {
	if route == nil {
		return false
	}
	labels := route.GetLabels()
	if labels == nil || labels["gateway.olares.io/auth-kind"] != AuthKindEntranceImages {
		return false
	}
	ann := route.GetAnnotations()
	if ann == nil || ann[AuxExtAuthBypassAnnotation] != "true" {
		return false
	}
	return HTTPRouteHasAuxPathRule(route, AuxImagesPathPrefix, AuxImagesServiceName, 8080)
}

// AuxTargetReadyFromRoutes evaluates one entrance using main + images HTTPRoutes.
func AuxTargetReadyFromRoutes(t EntranceAuxTarget, mainRoute, imagesRoute *unstructured.Unstructured) bool {
	if t.Needs.WS {
		if !HTTPRouteHasAuxPathRule(mainRoute, AuxWSPathPrefix, AuxHelperServiceNameSuffix, 40010) {
			return false
		}
	}
	if t.Needs.Upload {
		if !HTTPRouteHasAuxPathRule(mainRoute, AuxUploadPathPrefix, AuxHelperServiceNameSuffix, 40030) {
			return false
		}
	}
	if t.Needs.Images {
		if !ImagesRouteBypassesExtAuth(imagesRoute) {
			return false
		}
		if t.ImagesNamespace != "" {
			if !imagesBackendNamespaceOK(imagesRoute, t.ImagesNamespace) {
				return false
			}
		}
	}
	return true
}

func imagesBackendNamespaceOK(route *unstructured.Unstructured, wantNS string) bool {
	if route == nil || wantNS == "" {
		return true
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	for _, r := range rules {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		if !ruleMatchesPathPrefix(rm, AuxImagesPathPrefix) {
			continue
		}
		backends, _, _ := unstructured.NestedSlice(rm, "backendRefs")
		for _, b := range backends {
			bm, ok := b.(map[string]any)
			if !ok {
				continue
			}
			name, _ := bm["name"].(string)
			if name != AuxImagesServiceName {
				continue
			}
			ns, _ := bm["namespace"].(string)
			if ns == "" || ns == wantNS {
				return true
			}
			return false
		}
	}
	return false
}

// ListEntranceAuxTargets lists application-class SRRs; Needs default to Images=true.
// WS/Upload are filled from the main-route annotation when present.
func ListEntranceAuxTargets(ctx context.Context, dc dynamic.Interface) ([]EntranceAuxTarget, error) {
	ext, err := ListEntranceExtAuthTargets(ctx, dc)
	if err != nil {
		return nil, err
	}
	out := make([]EntranceAuxTarget, 0, len(ext))
	for _, t := range ext {
		needs := EntranceAuxNeeds{Images: true}
		route, err := dc.Resource(httpRouteGVR).Namespace(t.Namespace).Get(ctx, t.HTTPRoute, metav1.GetOptions{})
		if err == nil && route != nil {
			ann := route.GetAnnotations()
			if ann != nil {
				parsed := ParseAuxCapabilitiesAnnotation(ann[AuxCapabilitiesAnnotation])
				needs.WS = parsed.WS
				needs.Upload = parsed.Upload
				if parsed.Images {
					needs.Images = true
				}
			}
		}
		out = append(out, EntranceAuxTarget{
			Namespace: t.Namespace,
			SRRName:   t.SRRName,
			HTTPRoute: t.HTTPRoute,
			Needs:     needs,
		})
	}
	return out, nil
}

// ProbeEntranceAuxCovered verifies Aux north-south routes for application entrances.
func ProbeEntranceAuxCovered(ctx context.Context, dc dynamic.Interface) (bool, error) {
	if dc == nil {
		return false, fmt.Errorf("dynamic client required")
	}
	targets, err := ListEntranceAuxTargets(ctx, dc)
	if err != nil {
		return false, err
	}
	return EvaluateEntranceAuxCovered(targets, func(t EntranceAuxTarget) bool {
		main, err := dc.Resource(httpRouteGVR).Namespace(t.Namespace).Get(ctx, t.HTTPRoute, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("mesh: get main HTTPRoute %s/%s for aux failed: %v", t.Namespace, t.HTTPRoute, err)
			}
			return false
		}
		var images *unstructured.Unstructured
		if t.Needs.Images {
			images, err = dc.Resource(httpRouteGVR).Namespace(t.Namespace).Get(ctx, EntranceImagesRouteName(t.SRRName), metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					klog.Errorf("mesh: get images HTTPRoute %s/%s failed: %v", t.Namespace, EntranceImagesRouteName(t.SRRName), err)
				}
				return false
			}
		}
		return AuxTargetReadyFromRoutes(t, main, images)
	}), nil
}

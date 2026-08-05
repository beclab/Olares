package mesh

import "strings"

const (
	AnnotationRouteMode        = "gateway.olares.io/route-mode"
	AnnotationRouteModeGateway = "gateway"
)

// EntranceRouteModeTarget is one Application that must use route-mode=gateway
// before de-envoy cutover (ordinary or shared with entrances).
type EntranceRouteModeTarget struct {
	Namespace string
	AppName   string
	Mode      string // current annotation value (may be empty)
}

// EvaluateEntranceRouteModeGatewayReady requires every target to be gateway.
// Empty targets → true. Any non-gateway (including empty/direct) → false.
// Partial clusters (some gateway, some not) therefore never Ready.
func EvaluateEntranceRouteModeGatewayReady(targets []EntranceRouteModeTarget) bool {
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.AppName) == "" {
			return false
		}
		if strings.ToLower(strings.TrimSpace(t.Mode)) != AnnotationRouteModeGateway {
			return false
		}
	}
	return true
}

// HasPartialEntranceRouteModeGateway reports mixed gateway vs non-gateway among targets.
func HasPartialEntranceRouteModeGateway(targets []EntranceRouteModeTarget) bool {
	var sawGateway, sawOther bool
	for _, t := range targets {
		if strings.ToLower(strings.TrimSpace(t.Mode)) == AnnotationRouteModeGateway {
			sawGateway = true
		} else {
			sawOther = true
		}
	}
	return sawGateway && sawOther
}

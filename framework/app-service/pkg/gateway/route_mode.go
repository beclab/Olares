package gateway

import (
	"context"
	"strings"
	"sync"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

const (
	// AnnotationRouteModeDirect keeps the legacy l4-bfl-proxy direct upstream path.
	AnnotationRouteModeDirect = "direct"

	// SettingGatewayRouteMode is the optional manifest settings override
	// copied into Application.spec.settings (P1, below explicit annotations).
	SettingGatewayRouteMode = "gatewayRouteMode"

	defaultGatewayNS   = "os-gateway"
	defaultGatewayName = "app-gateway"
)

var gatewayReadyCache struct {
	mu        sync.RWMutex
	ready     bool
	expiresAt time.Time
}

const gatewayReadyCacheTTL = 30 * time.Second

// ComputeRouteModePatch decides whether the Application needs a route-mode
// annotation patch and what value to set.
//
// Priority: explicit annotation (P0, never overwritten) > manifest settings
// gatewayRouteMode (P1) > automatic gateway for shared apps when the cluster
// gate is on, the platform domain resolves and the app-gateway Gateway is
// ready. Everything else is left untouched (direct semantics).
func ComputeRouteModePatch(ctx context.Context, c client.Client, app *appv1alpha1.Application) (needsPatch bool, mode string, err error) {
	if app == nil {
		return false, "", nil
	}
	if v, ok := explicitRouteModeAnnotation(app); ok {
		return false, v, nil
	}
	if s := settingsRouteMode(app); s == AnnotationRouteModeGateway || s == AnnotationRouteModeDirect {
		return true, s, nil
	}
	if !appcfg.IsGatewaySharedApp(app) {
		return false, "", nil
	}
	if !cluster.GetInClusterGatewayEnabled(ctx) {
		klog.V(2).Infof("route-mode: skip auto gateway for app=%s: inClusterGatewayEnabled=false", app.Spec.Name)
		return false, "", nil
	}
	if cluster.GetPlatformDomain(ctx) == "" {
		klog.V(2).Infof("route-mode: skip auto gateway for app=%s: platformDomain empty", app.Spec.Name)
		return false, "", nil
	}
	if c != nil && !appGatewayReady(ctx, c) {
		klog.V(2).Infof("route-mode: skip auto gateway for app=%s: app-gateway not ready", app.Spec.Name)
		return false, "", nil
	}
	return true, AnnotationRouteModeGateway, nil
}

// ApplyRouteModeAnnotation sets app.metadata.annotations[route-mode] in memory
// when ComputeRouteModePatch says a patch is needed. Use before Application
// Create or before a Patch that already persists metadata.
func ApplyRouteModeAnnotation(ctx context.Context, c client.Client, app *appv1alpha1.Application) error {
	need, mode, err := ComputeRouteModePatch(ctx, c, app)
	if err != nil || !need {
		return err
	}
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	app.Annotations[AnnotationRouteMode] = mode
	return nil
}

// EnsureRouteModeAnnotation persists the route-mode annotation when the
// automation policy requires it. It is the convergence path for Applications
// whose annotation was removed or that were created before app-gateway became
// ready; explicit operator annotations are never overwritten.
func EnsureRouteModeAnnotation(ctx context.Context, c client.Client, app *appv1alpha1.Application) error {
	need, mode, err := ComputeRouteModePatch(ctx, c, app)
	if err != nil || !need {
		return err
	}
	patched := app.DeepCopy()
	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}
	patched.Annotations[AnnotationRouteMode] = mode
	if err := c.Patch(ctx, patched, client.MergeFrom(app)); err != nil {
		return err
	}
	app.Annotations = patched.Annotations
	return nil
}

func explicitRouteModeAnnotation(app *appv1alpha1.Application) (string, bool) {
	if app == nil || app.Annotations == nil {
		return "", false
	}
	v := strings.ToLower(strings.TrimSpace(app.Annotations[AnnotationRouteMode]))
	if v != AnnotationRouteModeGateway && v != AnnotationRouteModeDirect {
		return "", false
	}
	return v, true
}

func settingsRouteMode(app *appv1alpha1.Application) string {
	if app == nil || app.Spec.Settings == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(app.Spec.Settings[SettingGatewayRouteMode]))
}

func appGatewayReady(ctx context.Context, c client.Client) bool {
	gatewayReadyCache.mu.RLock()
	if time.Now().Before(gatewayReadyCache.expiresAt) {
		ready := gatewayReadyCache.ready
		gatewayReadyCache.mu.RUnlock()
		return ready
	}
	gatewayReadyCache.mu.RUnlock()

	ready := checkAppGatewayReady(ctx, c)

	gatewayReadyCache.mu.Lock()
	gatewayReadyCache.ready = ready
	gatewayReadyCache.expiresAt = time.Now().Add(gatewayReadyCacheTTL)
	gatewayReadyCache.mu.Unlock()
	return ready
}

func checkAppGatewayReady(ctx context.Context, c client.Client) bool {
	gw := &unstructured.Unstructured{}
	gw.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "gateway.networking.k8s.io",
		Version: "v1",
		Kind:    "Gateway",
	})
	if err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, gw); err != nil {
		return false
	}
	conditions, found, err := unstructured.NestedSlice(gw.Object, "status", "conditions")
	if err != nil || !found {
		// Gateway exists; treat as ready when status is not yet populated.
		return true
	}
	for _, item := range conditions {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		typ, _ := m["type"].(string)
		status, _ := m["status"].(string)
		if typ == "Accepted" && status == "True" {
			return true
		}
	}
	return false
}

// resetGatewayReadyCacheForTest clears the gateway readiness memoization cache.
func resetGatewayReadyCacheForTest() {
	gatewayReadyCache.mu.Lock()
	gatewayReadyCache.ready = false
	gatewayReadyCache.expiresAt = time.Time{}
	gatewayReadyCache.mu.Unlock()
}

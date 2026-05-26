package appcfg

import (
	"fmt"
	"strings"
)

const (
	manifestRouteModeDirect  = "direct"
	manifestRouteModeGateway = "gateway"
	manifestInClusterDirect  = "direct"
)

// ValidateCallerInClusterManifest rejects unsafe combinations for apps that
// declare appScope.appRef (cluster-internal Shared callers).
func ValidateCallerInClusterManifest(cfg *ApplicationConfig) error {
	if cfg == nil || len(cfg.AppScope.AppRef) == 0 {
		return nil
	}
	routeMode := strings.ToLower(strings.TrimSpace(cfg.GatewayRouteMode))
	inCluster := strings.ToLower(strings.TrimSpace(cfg.InClusterMode))
	if routeMode == manifestRouteModeDirect && inCluster == manifestRouteModeGateway {
		return fmt.Errorf("manifest: gatewayRouteMode=direct is incompatible with inCluster=gateway")
	}
	if routeMode == manifestRouteModeDirect {
		return fmt.Errorf("manifest: gatewayRouteMode=direct is incompatible with appScope.appRef; use gateway or omit gatewayRouteMode")
	}
	if mode := inCluster; mode == manifestInClusterDirect {
		if !cfg.OnlyAdmin && !cfg.AppScope.ClusterScoped {
			return fmt.Errorf("manifest: in-cluster=direct requires onlyAdmin or clusterScoped appScope for apps with appRef")
		}
	}
	return nil
}

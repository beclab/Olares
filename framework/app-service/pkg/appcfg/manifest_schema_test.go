package appcfg

import (
	"strings"
	"testing"

	"github.com/beclab/api/manifest"
)

func TestValidateCallerInClusterManifest_noAppRef(t *testing.T) {
	if err := ValidateCallerInClusterManifest(&ApplicationConfig{}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateCallerInClusterManifest_rejectsDirectRouteMode(t *testing.T) {
	cfg := &ApplicationConfig{
		AppScope: manifest.AppScope{
			AppRef:        []string{"ollamav2"},
			ClusterScoped: true,
		},
		GatewayRouteMode: manifestRouteModeDirect,
	}
	err := ValidateCallerInClusterManifest(cfg)
	if err == nil || !strings.Contains(err.Error(), "gatewayRouteMode=direct") {
		t.Fatalf("want gatewayRouteMode error, got %v", err)
	}
}

func TestValidateCallerInClusterManifest_allowsGateway(t *testing.T) {
	cfg := &ApplicationConfig{
		AppScope: manifest.AppScope{
			AppRef:        []string{"ollamav2"},
			ClusterScoped: true,
		},
		GatewayRouteMode: manifestRouteModeGateway,
	}
	if err := ValidateCallerInClusterManifest(cfg); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateCallerInClusterManifest_rejectsDirectRouteWithInClusterGateway(t *testing.T) {
	cfg := &ApplicationConfig{
		AppScope: manifest.AppScope{
			AppRef:        []string{"ollamav2"},
			ClusterScoped: true,
		},
		GatewayRouteMode: manifestRouteModeDirect,
		InClusterMode:    manifestRouteModeGateway,
	}
	err := ValidateCallerInClusterManifest(cfg)
	if err == nil || !strings.Contains(err.Error(), "inCluster=gateway") {
		t.Fatalf("want inCluster=gateway conflict, got %v", err)
	}
}

func TestValidateCallerInClusterManifest_rejectsInClusterDirectWithoutOwner(t *testing.T) {
	cfg := &ApplicationConfig{
		AppScope: manifest.AppScope{AppRef: []string{"ollamav2"}},
		InClusterMode: manifestInClusterDirect,
	}
	if err := ValidateCallerInClusterManifest(cfg); err == nil {
		t.Fatal("expected error for in-cluster=direct without owner scope")
	}
}

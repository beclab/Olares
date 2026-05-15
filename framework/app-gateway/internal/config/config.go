package config

import (
	"sync"

	appgateway "github.com/beclab/Olares/framework/app-gateway"
	"gopkg.in/yaml.v3"
)

// Defaults is loaded from config/defaults.yaml (single source of truth).
type Defaults struct {
	Namespace string `yaml:"namespace"`
	Gateway   struct {
		Name             string `yaml:"name"`
		GatewayClassName string `yaml:"gatewayClassName"`
	} `yaml:"gateway"`
	Demo struct {
		Enabled   bool   `yaml:"enabled"`
		Host      string `yaml:"host"`
		MeshDebug bool   `yaml:"meshDebug"`
	} `yaml:"demo"`
	EnvoyProxy struct {
		Enabled   bool `yaml:"enabled"`
		Name      string `yaml:"name"`
		AccessLog struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"accessLog"`
	} `yaml:"envoyProxy"`
	Mesh struct {
		Linkerd struct {
			Enabled     bool   `yaml:"enabled"`
			OpaquePorts string `yaml:"opaquePorts"`
		} `yaml:"linkerd"`
	} `yaml:"mesh"`
	Vendor struct {
		LinkerdNamespace string `yaml:"linkerdNamespace"`
	} `yaml:"vendor"`
}

// MeshLinkerdEnabled reports whether EG data-plane pods should get linkerd-proxy via EnvoyProxy CR.
func (d Defaults) MeshLinkerdEnabled() bool {
	return d.Mesh.Linkerd.Enabled
}

// DemoMeshDebugEnabled reports whether install should rollout/wait demo workloads for mesh (dev/debug only).
func (d Defaults) DemoMeshDebugEnabled() bool {
	return d.Demo.MeshDebug && d.Demo.Enabled
}

var (
	loadOnce sync.Once
	cached   Defaults
	loadErr  error
)

// Load returns parsed defaults (cached).
func Load() (Defaults, error) {
	loadOnce.Do(func() {
		loadErr = yaml.Unmarshal(appgateway.DefaultsYAML, &cached)
	})
	return cached, loadErr
}

// Namespace returns the app-gateway namespace (EG control plane + Gateway API objects).
// Override: APP_GATEWAY_NAMESPACE.
func Namespace() string {
	d, err := Load()
	if err != nil || d.Namespace == "" {
		return "app-gateway"
	}
	return d.Namespace
}

// LinkerdNamespace returns the Linkerd control plane namespace.
func LinkerdNamespace() string {
	d, err := Load()
	if err != nil || d.Vendor.LinkerdNamespace == "" {
		return "linkerd"
	}
	return d.Vendor.LinkerdNamespace
}

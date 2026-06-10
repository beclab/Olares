package terminus

import (
	"os"
	"path/filepath"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/core/connector"
)

const (
	appGatewaySystemReleaseName = "app-gateway-system"
	appGatewaySystemDirName     = "app-gateway-system"
)

// appGatewayStackEnabled reports whether the unified ingress stack is part of this run.
// Default on for Olares install/upgrade; set APP_GATEWAY_STACK_ENABLED=0 only for exceptional dev skips.
func appGatewayStackEnabled() bool {
	v := os.Getenv("APP_GATEWAY_STACK_ENABLED")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

func resolveInstallerDir(runtime connector.Runtime) string {
	if d := os.Getenv("OLARES_INSTALLER_DIR"); d != "" {
		return d
	}
	return runtime.GetInstallerDir()
}

func resolveAppGatewayNamespace() string {
	if ns := os.Getenv("APP_GATEWAY_NAMESPACE"); ns != "" {
		return ns
	}
	return agwconfig.Namespace()
}

func appGatewaySystemPath(installerDir string) string {
	return filepath.Join(installerDir, "wizard", "config", appGatewaySystemDirName)
}

func appGatewaySystemCRDsPath(installerDir string) string {
	return filepath.Join(appGatewaySystemPath(installerDir), "crds")
}

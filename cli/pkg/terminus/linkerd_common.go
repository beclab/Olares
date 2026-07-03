package terminus

import (
	"os"
	"path/filepath"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/core/connector"
)

// linkerdPostInstallEnabled reports whether Linkerd post-os-system tasks run (default on).
// Set OLARES_LINKERD_POST_INSTALL_ENABLED=0 to skip (vendor debug only).
func linkerdPostInstallEnabled() bool {
	v := os.Getenv("OLARES_LINKERD_POST_INSTALL_ENABLED")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

func resolveInstallerDir(runtime connector.Runtime) string {
	if d := os.Getenv("OLARES_INSTALLER_DIR"); d != "" {
		return d
	}
	return runtime.GetInstallerDir()
}

func osFrameworkDeployPath(installerDir string) string {
	return filepath.Join(installerDir, "wizard", "config", "os-framework", "templates", "deploy")
}

func agwconfigLinkerdNamespace() string {
	return agwconfig.LinkerdNamespace()
}

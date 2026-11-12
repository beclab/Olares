package app

import (
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

// ResolveInstallNamespace returns the Kubernetes namespace used for a new
// install. It never honors ApplicationConfig.Namespace as a manifest-provided
// override (that field is only meaningful after persistence on an
// ApplicationManager). Order: middleware → shared → <app>-<owner>.
func ResolveInstallNamespace(app, owner string, cfg *appcfg.ApplicationConfig) string {
	if cfg != nil && cfg.IsMiddleware() {
		name := cfg.AppName
		if name == "" {
			name = app
		}
		return fmt.Sprintf("%s-%s", name, "middleware")
	}
	if cfg != nil && cfg.IsShared() {
		return SharedAppNamespace(app)
	}
	return fmt.Sprintf("%s-%s", app, owner)
}

package appcfg

import (
	"fmt"
	manifestsysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"path/filepath"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"
)

// Type aliases to the shared manifest types in github.com/beclab/api so that
// the structs describing application manifests live in a single place.
type (
	AppMetaData             = manifest.AppMetaData
	AppConfiguration        = manifest.AppConfiguration
	AppSpec                 = manifest.AppSpec
	Hardware                = manifest.Hardware
	CpuConfig               = manifest.CpuConfig
	GpuConfig               = manifest.GpuConfig
	SupportClient           = manifest.SupportClient
	Permission              = manifest.Permission
	ProviderPermission      = manifest.ProviderPermission
	Policy                  = manifest.Policy
	Dependency              = manifest.Dependency
	Conflict                = manifest.Conflict
	Options                 = manifest.Options
	ResetCookie             = manifest.ResetCookie
	AppScope                = manifest.AppScope
	WsConfig                = manifest.WsConfig
	Upload                  = manifest.Upload
	OIDC                    = manifest.OIDC
	Chart                   = manifest.Chart
	Provider                = manifest.Provider
	SpecialResource         = manifest.SpecialResource
	ResourceRequirement     = manifest.ResourceRequirement
	ResourceMode            = manifest.ResourceMode
	Entrance                = appv1alpha1.Entrance
	ServicePort             = appv1alpha1.ServicePort
	TailScale               = appv1alpha1.TailScale
	Middleware              = manifest.Middleware
	Application             = appv1alpha1.Application
	ApplicationSpec         = appv1alpha1.ApplicationSpec
	AppEnvVar               = manifestsysv1alpha1.AppEnvVar
	ACL                     = appv1alpha1.ACL
	ApplicationManager      = appv1alpha1.ApplicationManager
	ApplicationManagerState = appv1alpha1.ApplicationManagerState
	OpRecord                = appv1alpha1.OpRecord
)

func ChartNamespace(c *Chart, owner string) string {
	if c.Shared {
		return fmt.Sprintf("%s-%s", c.Name, "shared")
	}
	return fmt.Sprintf("%s-%s", c.Name, owner)
}

// ChartPath returns the on-disk path to a sub-chart's files.
func ChartPath(appName, chartName string) string {
	return AppChartPath(filepath.Join(appName, chartName))
}

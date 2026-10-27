package common

import "github.com/beclab/Olares/cli/pkg/systemcomponents"

// The namespaces Olares deploys into are defined in pkg/systemcomponents, which
// is shared with olaresd, and re-exported here so that existing callers keep
// working and the two can never drift apart.
const (
	NamespaceKubeSystem                 = systemcomponents.NamespaceKubeSystem
	NamespaceKubesphereMonitoringSystem = systemcomponents.NamespaceKubesphereMonitoringSystem
	NamespaceKubesphereSystem           = systemcomponents.NamespaceKubesphereSystem
	NamespaceOsFramework                = systemcomponents.NamespaceOsFramework
	NamespaceOsPlatform                 = systemcomponents.NamespaceOsPlatform
	NamespaceOsProtected                = systemcomponents.NamespaceOsProtected
	NamespaceOsGateway                  = systemcomponents.NamespaceOsGateway
	NamespaceOsMesh                     = systemcomponents.NamespaceOsMesh
	NamespaceOsGpu                      = systemcomponents.NamespaceOsGpu
	NamespaceOsNetwork                  = systemcomponents.NamespaceOsNetwork
	NamespaceUserSpacePrefix            = systemcomponents.NamespaceUserSpacePrefix
	NamespaceUserSystemPrefix           = systemcomponents.NamespaceUserSystemPrefix
)

const (
	NamespaceDefault                  = "default"
	NamespaceKubeNodeLease            = "kube-node-lease"
	NamespaceKubePublic               = "kube-public"
	NamespaceKubekeySystem            = "kubekey-system"
	NamespaceKubesphereControlsSystem = "kubesphere-controls-system"

	ChartNameRedis               = "redis"
	ChartNameSnapshotController  = "snapshot-controller"
	ChartNameKsCore              = "ks-core"
	ChartNameKsCoreConfig        = "ks-core-config"
	ChartNameKsConfig            = "ks-config"
	ChartNameMonitorNotification = "monitor-notification"
	ChartNameAccount             = "account"
	ChartNameOSFramework         = "os-framework"
	ChartNameOSPlatform          = "os-platform"
	ChartNameSettings            = "settings"
)

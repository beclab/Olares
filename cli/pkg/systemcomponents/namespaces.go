package systemcomponents

// The namespaces an Olares installation puts its own workloads in.
//
// They are defined here rather than in pkg/common so that this package stays a
// leaf with no dependency on the rest of the CLI: olaresd imports it, and
// pulling the CLI's runtime packages into the daemon would drag in their whole
// dependency tree. pkg/common re-exports these, so there is still a single
// definition of each name.
const (
	NamespaceKubeSystem                 = "kube-system"
	NamespaceKubesphereSystem           = "kubesphere-system"
	NamespaceKubesphereMonitoringSystem = "kubesphere-monitoring-system"

	NamespaceOsFramework = "os-framework"
	NamespaceOsPlatform  = "os-platform"
	NamespaceOsProtected = "os-protected"
	NamespaceOsGateway   = "os-gateway"
	NamespaceOsMesh      = "os-mesh"
	NamespaceOsGpu       = "os-gpu"
	NamespaceOsNetwork   = "os-network"

	// NamespaceUserSpacePrefix and NamespaceUserSystemPrefix are prefixed to an
	// Olares user name to form that user's two namespaces.
	NamespaceUserSpacePrefix  = "user-space-"
	NamespaceUserSystemPrefix = "user-system-"
)

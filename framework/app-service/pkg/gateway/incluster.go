package gateway

// Cluster-internal Shared access annotations and settings.
const (
	// AnnotationInCluster selects gateway vs direct cluster-internal routing.
	AnnotationInCluster = "gateway.olares.io/in-cluster"
	// InClusterGateway enables mesh + egress NP + app-gateway path.
	InClusterGateway = "gateway"
	// InClusterDirect keeps legacy direct Service access.
	InClusterDirect = "direct"

	// SettingInClusterMode is copied from the install manifest into
	// Application.spec.settings (P1 override), same as gatewayRouteMode.
	SettingInClusterMode = "inCluster"
)

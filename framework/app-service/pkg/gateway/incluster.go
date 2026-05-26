package gateway

// Cluster-internal Shared access annotations and settings.
const (
	// AnnotationInCluster selects gateway vs direct cluster-internal routing.
	AnnotationInCluster = "gateway.olares.io/in-cluster"
	// InClusterGateway enables mesh + egress NP + app-gateway path.
	InClusterGateway = "gateway"
	// InClusterDirect keeps legacy direct Service access.
	InClusterDirect = "direct"
)

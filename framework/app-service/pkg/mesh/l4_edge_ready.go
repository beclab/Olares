package mesh

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// L4ProxyNamespace hosts the edge north-south PEP (l4-bfl-proxy).
	L4ProxyNamespace = "os-network"
	// L4ProxyDeploymentName is the Track-A data-plane Deployment.
	L4ProxyDeploymentName = "l4-bfl-proxy"
)

// EvaluateL4ProxyReady is the pure Track-A M-L4 predicate:
// verify/302 paths and Cookie Domain rewrite and probe bypass.
func EvaluateL4ProxyReady(verifyReady, cookieReady, probeReady bool) bool {
	return verifyReady && cookieReady && probeReady
}

// IsL4ProxyDeploymentReady reports whether the l4 north-south PEP Deployment
// has at least one Ready replica. Used for direct oes stop-inject without
// waiting for Accept to write SteadyGate L4ProxyReady (avoids chicken-egg).
func IsL4ProxyDeploymentReady(ctx context.Context, kube kubernetes.Interface) bool {
	if kube == nil {
		return false
	}
	dep, err := kube.AppsV1().Deployments(L4ProxyNamespace).Get(ctx, L4ProxyDeploymentName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return dep.Status.ReadyReplicas >= 1
}

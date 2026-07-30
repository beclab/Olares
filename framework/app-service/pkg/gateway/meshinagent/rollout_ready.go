package meshinagent

import (
	"context"
	"sync"

	"k8s.io/client-go/kubernetes"
)

var (
	meshReadyMu   sync.RWMutex
	meshReadyCheck func(context.Context, kubernetes.Interface) bool
)

// SetMeshControlPlaneReadyCheck installs the cluster probe used before inject
// rollouts. main wires this at process start.
func SetMeshControlPlaneReadyCheck(fn func(context.Context, kubernetes.Interface) bool) {
	meshReadyMu.Lock()
	defer meshReadyMu.Unlock()
	meshReadyCheck = fn
}

// IsMeshControlPlaneReady reports whether the mesh control plane can accept
// injected workloads.
func IsMeshControlPlaneReady(ctx context.Context, kube kubernetes.Interface) bool {
	meshReadyMu.RLock()
	fn := meshReadyCheck
	meshReadyMu.RUnlock()
	if fn == nil {
		return false
	}
	return fn(ctx, kube)
}

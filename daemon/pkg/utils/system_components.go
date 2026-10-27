package utils

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/systemcomponents"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// informerSource serves the system component readiness check from olaresd's
// shared informer caches, so that the status loop does not list the whole
// cluster on every tick.
type informerSource struct{}

func (informerSource) Namespaces(ctx context.Context) ([]string, error) {
	return ListNamespaces(ctx)
}

func (informerSource) Deployments(ctx context.Context) ([]*appsv1.Deployment, error) {
	return ListDeployments(ctx)
}

func (informerSource) StatefulSets(ctx context.Context) ([]*appsv1.StatefulSet, error) {
	return ListStatefulSets(ctx)
}

func (informerSource) DaemonSets(ctx context.Context) ([]*appsv1.DaemonSet, error) {
	return ListDaemonSets(ctx)
}

func (informerSource) Pods(ctx context.Context) ([]*corev1.Pod, error) {
	return ListPods(ctx)
}

// CheckSystemComponents reports why the Olares system components are not ready,
// or nil when they all are.
func CheckSystemComponents(ctx context.Context) error {
	return systemcomponents.NewChecker(informerSource{}).Check(ctx, systemcomponents.Default())
}

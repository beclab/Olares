package systemcomponents

import (
	"context"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Source supplies the cluster objects a readiness check reads.
//
// The CLI backs it with live List calls against the API server. The daemon
// backs it with its shared informer caches so that its status loop, which runs
// every few seconds, does not list the whole cluster on every tick.
type Source interface {
	Namespaces(ctx context.Context) ([]string, error)
	Deployments(ctx context.Context) ([]*appsv1.Deployment, error)
	StatefulSets(ctx context.Context) ([]*appsv1.StatefulSet, error)
	DaemonSets(ctx context.Context) ([]*appsv1.DaemonSet, error)
	Pods(ctx context.Context) ([]*corev1.Pod, error)
}

// NewClientSource returns a Source that lists from the API server. Results are
// memoised for the lifetime of the source, and pods are only fetched if a
// component turns out to be unready, so a passing check costs four List calls.
// Use one source per check so it never serves stale objects.
func NewClientSource(client ctrlclient.Client) Source {
	return &clientSource{client: client}
}

type clientSource struct {
	client ctrlclient.Client

	namespacesOnce sync.Once
	namespaces     []string
	namespacesErr  error

	deploymentsOnce sync.Once
	deployments     []*appsv1.Deployment
	deploymentsErr  error

	statefulSetsOnce sync.Once
	statefulSets     []*appsv1.StatefulSet
	statefulSetsErr  error

	daemonSetsOnce sync.Once
	daemonSets     []*appsv1.DaemonSet
	daemonSetsErr  error

	podsOnce sync.Once
	pods     []*corev1.Pod
	podsErr  error
}

func (s *clientSource) Namespaces(ctx context.Context) ([]string, error) {
	s.namespacesOnce.Do(func() {
		var list corev1.NamespaceList
		if err := s.client.List(ctx, &list); err != nil {
			s.namespacesErr = err
			return
		}
		s.namespaces = make([]string, 0, len(list.Items))
		for i := range list.Items {
			s.namespaces = append(s.namespaces, list.Items[i].Name)
		}
	})
	return s.namespaces, s.namespacesErr
}

func (s *clientSource) Deployments(ctx context.Context) ([]*appsv1.Deployment, error) {
	s.deploymentsOnce.Do(func() {
		var list appsv1.DeploymentList
		if err := s.client.List(ctx, &list); err != nil {
			s.deploymentsErr = err
			return
		}
		s.deployments = make([]*appsv1.Deployment, 0, len(list.Items))
		for i := range list.Items {
			s.deployments = append(s.deployments, &list.Items[i])
		}
	})
	return s.deployments, s.deploymentsErr
}

func (s *clientSource) StatefulSets(ctx context.Context) ([]*appsv1.StatefulSet, error) {
	s.statefulSetsOnce.Do(func() {
		var list appsv1.StatefulSetList
		if err := s.client.List(ctx, &list); err != nil {
			s.statefulSetsErr = err
			return
		}
		s.statefulSets = make([]*appsv1.StatefulSet, 0, len(list.Items))
		for i := range list.Items {
			s.statefulSets = append(s.statefulSets, &list.Items[i])
		}
	})
	return s.statefulSets, s.statefulSetsErr
}

func (s *clientSource) DaemonSets(ctx context.Context) ([]*appsv1.DaemonSet, error) {
	s.daemonSetsOnce.Do(func() {
		var list appsv1.DaemonSetList
		if err := s.client.List(ctx, &list); err != nil {
			s.daemonSetsErr = err
			return
		}
		s.daemonSets = make([]*appsv1.DaemonSet, 0, len(list.Items))
		for i := range list.Items {
			s.daemonSets = append(s.daemonSets, &list.Items[i])
		}
	})
	return s.daemonSets, s.daemonSetsErr
}

func (s *clientSource) Pods(ctx context.Context) ([]*corev1.Pod, error) {
	s.podsOnce.Do(func() {
		var list corev1.PodList
		if err := s.client.List(ctx, &list); err != nil {
			s.podsErr = err
			return
		}
		s.pods = make([]*corev1.Pod, 0, len(list.Items))
		for i := range list.Items {
			s.pods = append(s.pods, &list.Items[i])
		}
	})
	return s.pods, s.podsErr
}

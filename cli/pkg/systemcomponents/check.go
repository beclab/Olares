package systemcomponents

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// maxReportedFailures caps how many components a single error message names, so
// that a cluster that is still coming up does not print a wall of text on every
// retry.
const maxReportedFailures = 5

// NotReadyError reports components that are missing or not ready. It is
// distinct from the errors Check returns when it cannot read the cluster at
// all, so that a caller can tell a genuinely broken system apart from an
// unreachable API server.
type NotReadyError struct {
	// Failures holds one entry per component, each already prefixed with the
	// component's identifier.
	Failures []string
}

func (e *NotReadyError) Error() string {
	reported := e.Failures
	if len(reported) > maxReportedFailures {
		reported = reported[:maxReportedFailures]
	}
	msg := fmt.Sprintf("%d Olares system component(s) not ready: %s",
		len(e.Failures), strings.Join(reported, "; "))
	if len(e.Failures) > len(reported) {
		msg = fmt.Sprintf("%s; and %d more", msg, len(e.Failures)-len(reported))
	}
	return msg
}

// Checker verifies that a set of components is ready on a cluster.
type Checker struct {
	source Source
	node   string
}

// NewChecker returns a Checker reading through the given source.
func NewChecker(source Source) *Checker {
	return &Checker{source: source}
}

// OnNode restricts the check to a single node: components still have to exist,
// but instead of comparing cluster wide replica counts the checker asserts that
// the pods scheduled on that node are ready. Use it after restarting or
// re-addressing one node, where the rest of the cluster is not the concern.
func (c *Checker) OnNode(node string) *Checker {
	c.node = node
	return c
}

// Check reports why the given components are not ready, or nil when they all
// are. Components the cluster is not expected to have (see Optional) are
// skipped when their workload is absent.
func (c *Checker) Check(ctx context.Context, components []Component) error {
	users, err := c.resolveUsers(ctx, components)
	if err != nil {
		return err
	}

	idx, err := c.buildIndex(ctx)
	if err != nil {
		return err
	}

	// in node mode every assertion is made against pods, so a failure to read
	// them is reported as such instead of being folded into NotReadyError
	if c.node != "" {
		if _, err := c.source.Pods(ctx); err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}
	}

	var failures []string
	for _, component := range Resolve(components, users) {
		if err := c.checkOne(ctx, component, idx); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", component.ID(), err))
		}
	}
	if len(failures) == 0 {
		return nil
	}
	return &NotReadyError{Failures: failures}
}

// Resolve expands the per user components in the list once for every user, and
// drops them entirely when there is no user yet.
func Resolve(components []Component, users []string) []Component {
	resolved := make([]Component, 0, len(components))
	for _, component := range components {
		if !component.isPerUser() {
			resolved = append(resolved, component)
			continue
		}
		for _, user := range users {
			resolved = append(resolved, component.forUser(user))
		}
	}
	return resolved
}

// resolveUsers returns the Olares users whose workloads have to be checked. It
// only looks at the cluster when the component list actually contains per user
// entries, so targeted checks stay to a single List call.
func (c *Checker) resolveUsers(ctx context.Context, components []Component) ([]string, error) {
	perUser := false
	for _, component := range components {
		if component.isPerUser() {
			perUser = true
			break
		}
	}
	if !perUser {
		return nil, nil
	}

	namespaces, err := c.source.Namespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var users []string
	for _, namespace := range namespaces {
		if !strings.HasPrefix(namespace, NamespaceUserSpacePrefix) {
			continue
		}
		if user := strings.TrimPrefix(namespace, NamespaceUserSpacePrefix); user != "" {
			users = append(users, user)
		}
	}
	sort.Strings(users)
	return users, nil
}

// workload is the part of a Deployment, StatefulSet or DaemonSet the checker
// cares about, so that the three kinds can share the same code path.
type workload struct {
	namespace string
	name      string
	selector  *metav1.LabelSelector
	assert    func() error
}

type clusterIndex struct {
	deployments  map[string]*appsv1.Deployment
	statefulSets map[string]*appsv1.StatefulSet
	daemonSets   map[string]*appsv1.DaemonSet

	deploymentsByNamespace  map[string][]*appsv1.Deployment
	statefulSetsByNamespace map[string][]*appsv1.StatefulSet
	daemonSetsByNamespace   map[string][]*appsv1.DaemonSet
}

func (c *Checker) buildIndex(ctx context.Context) (*clusterIndex, error) {
	deployments, err := c.source.Deployments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	statefulSets, err := c.source.StatefulSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}
	daemonSets, err := c.source.DaemonSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	idx := &clusterIndex{
		deployments:             make(map[string]*appsv1.Deployment, len(deployments)),
		statefulSets:            make(map[string]*appsv1.StatefulSet, len(statefulSets)),
		daemonSets:              make(map[string]*appsv1.DaemonSet, len(daemonSets)),
		deploymentsByNamespace:  make(map[string][]*appsv1.Deployment),
		statefulSetsByNamespace: make(map[string][]*appsv1.StatefulSet),
		daemonSetsByNamespace:   make(map[string][]*appsv1.DaemonSet),
	}
	for _, d := range deployments {
		idx.deployments[objectKey(d.Namespace, d.Name)] = d
		idx.deploymentsByNamespace[d.Namespace] = append(idx.deploymentsByNamespace[d.Namespace], d)
	}
	for _, s := range statefulSets {
		idx.statefulSets[objectKey(s.Namespace, s.Name)] = s
		idx.statefulSetsByNamespace[s.Namespace] = append(idx.statefulSetsByNamespace[s.Namespace], s)
	}
	for _, d := range daemonSets {
		idx.daemonSets[objectKey(d.Namespace, d.Name)] = d
		idx.daemonSetsByNamespace[d.Namespace] = append(idx.daemonSetsByNamespace[d.Namespace], d)
	}
	return idx, nil
}

func objectKey(namespace, name string) string {
	return namespace + "/" + name
}

func (c *Checker) checkOne(ctx context.Context, component Component, idx *clusterIndex) error {
	workloads, err := locate(component, idx)
	if err != nil {
		return err
	}

	if len(workloads) == 0 {
		if component.Presence == Optional {
			return nil
		}
		return fmt.Errorf("not found")
	}

	for _, w := range workloads {
		err := c.assertWorkload(ctx, w)
		if err == nil {
			continue
		}
		// an optional component that was scaled away is no different from one
		// that was never deployed here
		if component.Presence == Optional && errors.Is(err, ErrScaledToZero) {
			continue
		}
		return err
	}
	return nil
}

func (c *Checker) assertWorkload(ctx context.Context, w workload) error {
	if c.node != "" {
		return c.assertPodsOnNode(ctx, w)
	}

	err := w.assert()
	if err == nil {
		return nil
	}
	// a workload with no desired replicas has no pods to explain itself with,
	// and the sentinel has to survive for the caller to recognise it
	if errors.Is(err, ErrScaledToZero) {
		return err
	}
	if reason := c.firstUnreadyPod(ctx, w, ""); reason != "" {
		return fmt.Errorf("%s: %s", err, reason)
	}
	return err
}

// assertPodsOnNode asserts the pods of a workload that live on the checker's
// node. While a live pod is there its own state decides, so a replica running
// on some other node stays none of this check's business.
//
// Finding no live pod is the ambiguous case, and it falls back to the workload's
// own state to tell the two readings apart. A workload that is ready cluster
// wide simply keeps its replicas elsewhere, and passes. A workload that is not
// ready has yet to recreate the pods it lost, which is the state right after
// they were deleted, and must not pass: nothing has been proven about this node
// at all. Without that fallback the check would report success the moment its
// pods went away.
func (c *Checker) assertPodsOnNode(ctx context.Context, w workload) error {
	pods, err := c.podsOf(ctx, w, c.node)
	if err != nil {
		return err
	}

	live := 0
	for _, pod := range pods {
		// a finished pod is an execution record rather than a replica, and
		// counting it as one would hide that nothing is actually serving here
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		live++
		if err := AssertPodReady(pod); err != nil {
			return err
		}
	}

	if live == 0 {
		if err := w.assert(); err != nil {
			return fmt.Errorf("no pod running on node %s: %w", c.node, err)
		}
	}
	return nil
}

// firstUnreadyPod returns the reason the first unready pod of a workload gives,
// so that a bare replica count is accompanied by something actionable. It
// returns an empty string when the pods cannot be read or all of them look fine,
// in which case the replica count stands on its own.
func (c *Checker) firstUnreadyPod(ctx context.Context, w workload, node string) string {
	pods, err := c.podsOf(ctx, w, node)
	if err != nil {
		return ""
	}
	if len(pods) == 0 {
		return "no pod has been created yet"
	}
	for _, pod := range pods {
		if err := AssertPodReady(pod); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (c *Checker) podsOf(ctx context.Context, w workload, node string) ([]*corev1.Pod, error) {
	if w.selector == nil {
		return nil, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(w.selector)
	if err != nil {
		return nil, err
	}

	pods, err := c.source.Pods(ctx)
	if err != nil {
		return nil, err
	}

	var matched []*corev1.Pod
	for _, pod := range pods {
		if pod.Namespace != w.namespace {
			continue
		}
		if node != "" && pod.Spec.NodeName != node {
			continue
		}
		if !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		matched = append(matched, pod)
	}
	return matched, nil
}

// locate finds the workloads a component refers to, either by name or, for
// workloads whose names are generated per cluster, by label selector.
func locate(component Component, idx *clusterIndex) ([]workload, error) {
	if component.Selector != "" {
		return locateBySelector(component, idx)
	}
	return locateByName(component, idx), nil
}

func locateByName(component Component, idx *clusterIndex) []workload {
	key := objectKey(component.Namespace, component.Name)
	switch component.Kind {
	case Deployment:
		if d, ok := idx.deployments[key]; ok {
			return []workload{deploymentWorkload(d)}
		}
	case StatefulSet:
		if s, ok := idx.statefulSets[key]; ok {
			return []workload{statefulSetWorkload(s)}
		}
	case DaemonSet:
		if d, ok := idx.daemonSets[key]; ok {
			return []workload{daemonSetWorkload(d)}
		}
	}
	return nil
}

func locateBySelector(component Component, idx *clusterIndex) ([]workload, error) {
	selector, err := labels.Parse(component.Selector)
	if err != nil {
		return nil, fmt.Errorf("invalid selector %q: %w", component.Selector, err)
	}

	var workloads []workload
	switch component.Kind {
	case Deployment:
		for _, d := range idx.deploymentsByNamespace[component.Namespace] {
			if selector.Matches(labels.Set(d.Labels)) {
				workloads = append(workloads, deploymentWorkload(d))
			}
		}
	case StatefulSet:
		for _, s := range idx.statefulSetsByNamespace[component.Namespace] {
			if selector.Matches(labels.Set(s.Labels)) {
				workloads = append(workloads, statefulSetWorkload(s))
			}
		}
	case DaemonSet:
		for _, d := range idx.daemonSetsByNamespace[component.Namespace] {
			if selector.Matches(labels.Set(d.Labels)) {
				workloads = append(workloads, daemonSetWorkload(d))
			}
		}
	}
	return workloads, nil
}

func deploymentWorkload(d *appsv1.Deployment) workload {
	return workload{
		namespace: d.Namespace,
		name:      d.Name,
		selector:  d.Spec.Selector,
		assert:    func() error { return AssertDeploymentReady(d) },
	}
}

func statefulSetWorkload(s *appsv1.StatefulSet) workload {
	return workload{
		namespace: s.Namespace,
		name:      s.Name,
		selector:  s.Spec.Selector,
		assert:    func() error { return AssertStatefulSetReady(s) },
	}
}

func daemonSetWorkload(d *appsv1.DaemonSet) workload {
	return workload{
		namespace: d.Namespace,
		name:      d.Name,
		selector:  d.Spec.Selector,
		assert:    func() error { return AssertDaemonSetReady(d) },
	}
}

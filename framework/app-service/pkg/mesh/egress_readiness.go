package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

// EgressReadinessConfig lives in a platform namespace so a tenant cannot attest
// to its own traffic and implicitly turn on namespace-wide egress isolation.
var EgressReadinessConfig = client.ObjectKey{Namespace: "os-framework", Name: "olares-egress-readiness"}

type egressReadiness struct {
	NamespaceUID   types.UID             `json:"namespaceUID"`
	Pods           map[string]types.UID  `json:"pods"`
	NativePods     []string              `json:"nativePods"`
	Authorizations []egressAuthorization `json:"authorizations"`
	Dependencies   []string              `json:"dependencies"`
}

type egressAuthorization struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Namespace  string    `json:"namespace"`
	Name       string    `json:"name"`
	UID        types.UID `json:"uid"`
	Generation int64     `json:"generation"`
}

// CheckEgressReadiness validates a traffic-tested snapshot before first
// isolation. Validation augments the test evidence: readiness and Kubernetes
// object status alone cannot prove CNI reachability or every app dependency.
func CheckEgressReadiness(ctx context.Context, r client.Reader, ns *corev1.Namespace) error {
	var record corev1.ConfigMap
	if err := r.Get(ctx, EgressReadinessConfig, &record); err != nil {
		return fmt.Errorf("read egress readiness: %w", err)
	}
	var plan egressReadiness
	if err := json.Unmarshal([]byte(record.Data[ns.Name]), &plan); err != nil {
		return fmt.Errorf("invalid or missing egress readiness for %s: %w", ns.Name, err)
	}
	if ns.UID == "" || plan.NamespaceUID != ns.UID || len(plan.Pods) == 0 {
		return fmt.Errorf("egress readiness does not identify the current namespace and workloads")
	}
	if err := checkEgressRollouts(ctx, r, ns.Name); err != nil {
		return err
	}
	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(ns.Name)); err != nil {
		return err
	}
	meshCount := 0
	active := make([]corev1.Pod, 0, len(pods.Items))
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		active = append(active, *pod)
		ready := slices.ContainsFunc(pod.Status.Conditions, func(c corev1.PodCondition) bool {
			return c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue
		})
		if pod.UID == "" || plan.Pods[pod.Name] != pod.UID || !pod.DeletionTimestamp.IsZero() || !ready {
			return fmt.Errorf("Pod %s is unverified or not ready", pod.Name)
		}
		if slices.Contains(plan.NativePods, pod.Name) && LinkerdProxy(pod) == nil &&
			pod.Annotations[LinkerdServerTLSModeAnnotation] != MTLSModeRequired {
			continue
		}
		inbound := pod.Annotations[LinkerdDefaultInboundPolicyAnnotation]
		if (inbound != InboundPolicyClusterAuthenticated && inbound != "deny") ||
			CheckHTTPMeshInterception(pod) != nil || !HasHTTPMeshInit(pod) ||
			!LinkerdProxyReady(pod) || !ProxyHasInboundPolicy(pod, inbound) {
			return fmt.Errorf("Pod %s does not have a ready and enforcing mesh proxy", pod.Name)
		}
		meshCount++
	}
	if len(active) != len(plan.Pods) || (meshCount > 0 && len(plan.Authorizations) == 0) {
		return fmt.Errorf("egress readiness lacks current workload or authorization evidence")
	}
	for _, pod := range active {
		systemRoute := ""
		if security.IsUserSpaceNamespaces(ns.Name) {
			switch pod.Labels["app"] {
			case "authelia":
				systemRoute = "sys-auth"
			case "olares-app":
				systemRoute = "sys-desktop"
			}
		}
		if systemRoute != "" && !slices.ContainsFunc(plan.Authorizations, func(auth egressAuthorization) bool {
			return auth.APIVersion == gatewayv1alpha1.GroupVersion.String() && auth.Kind == "SharedRouteRegistry" &&
				auth.Namespace == ns.Name && auth.Name == systemRoute
		}) {
			return fmt.Errorf("missing authorization evidence for %s/%s", ns.Name, systemRoute)
		}
	}
	for _, auth := range plan.Authorizations {
		if err := checkEgressAuthorization(ctx, r, auth); err != nil {
			return err
		}
	}

	dependencies := []string{"default/kubernetes", "kube-system/coredns"}
	if meshCount > 0 {
		dependencies = append(dependencies, "os-mesh/linkerd-identity", "os-mesh/linkerd-dst", "os-mesh/linkerd-policy")
	}
	for _, dep := range security.ActivePlatformTrafficServices(ns.Name, active) {
		dependencies = append(dependencies, dep.String())
	}
	for _, dep := range append(dependencies, plan.Dependencies...) {
		namespace, name, ok := strings.Cut(dep, "/")
		if !ok || namespace == "" || name == "" || strings.Contains(name, "/") {
			return fmt.Errorf("invalid egress dependency %q", dep)
		}
		if err := checkEgressService(ctx, r, client.ObjectKey{Namespace: namespace, Name: name}); err != nil {
			return err
		}
	}
	return nil
}

func checkEgressRollouts(ctx context.Context, r client.Reader, namespace string) error {
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, d := range deployments.Items {
		want := int32(1)
		if d.Spec.Replicas != nil {
			want = *d.Spec.Replicas
		}
		if !d.DeletionTimestamp.IsZero() || d.Status.ObservedGeneration < d.Generation ||
			d.Status.UpdatedReplicas != want || d.Status.AvailableReplicas < want || d.Status.Replicas != want {
			return fmt.Errorf("Deployment %s has not completed its rollout", d.Name)
		}
	}
	var sets appsv1.StatefulSetList
	if err := r.List(ctx, &sets, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, s := range sets.Items {
		want := int32(1)
		if s.Spec.Replicas != nil {
			want = *s.Spec.Replicas
		}
		if !s.DeletionTimestamp.IsZero() || s.Status.ObservedGeneration < s.Generation ||
			s.Status.UpdatedReplicas != want || s.Status.ReadyReplicas != want || s.Status.Replicas != want {
			return fmt.Errorf("StatefulSet %s has not completed its rollout", s.Name)
		}
	}
	var daemons appsv1.DaemonSetList
	if err := r.List(ctx, &daemons, client.InNamespace(namespace)); err != nil {
		return err
	}
	for _, d := range daemons.Items {
		want := d.Status.DesiredNumberScheduled
		if !d.DeletionTimestamp.IsZero() || d.Status.ObservedGeneration < d.Generation ||
			d.Status.UpdatedNumberScheduled != want || d.Status.NumberAvailable != want || d.Status.NumberMisscheduled != 0 {
			return fmt.Errorf("DaemonSet %s has not completed its rollout", d.Name)
		}
	}
	return nil
}

func checkEgressAuthorization(ctx context.Context, r client.Reader, auth egressAuthorization) error {
	gvk := schema.FromAPIVersionAndKind(auth.APIVersion, auth.Kind)
	key := client.ObjectKey{Namespace: auth.Namespace, Name: auth.Name}
	if key.Namespace == "" || key.Name == "" || auth.UID == "" || auth.Generation < 1 {
		return fmt.Errorf("incomplete authorization evidence for %s", key)
	}
	if gvk == gatewayv1alpha1.GroupVersion.WithKind("SharedRouteRegistry") {
		var route gatewayv1alpha1.SharedRouteRegistry
		if err := r.Get(ctx, key, &route); err != nil {
			return err
		}
		ready := meta.FindStatusCondition(route.Status.Conditions, "Ready")
		if route.UID != auth.UID || !route.DeletionTimestamp.IsZero() || auth.Generation != route.Generation ||
			route.Status.ObservedGeneration != route.Generation || ready == nil ||
			ready.ObservedGeneration != route.Generation || ready.Status != metav1.ConditionTrue ||
			ready.Reason != "Reconciled" || route.Spec.RouteMode != gatewayv1alpha1.RouteModeGateway {
			return fmt.Errorf("route %s is not ready for the reviewed generation", key)
		}
		return nil
	}
	if gvk != (schema.GroupVersionKind{Group: "policy.linkerd.io", Version: "v1alpha1", Kind: "AuthorizationPolicy"}) {
		return fmt.Errorf("unsupported authorization evidence %s", gvk)
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, key, obj); err != nil {
		return err
	}
	if obj.GetUID() != auth.UID || obj.GetGeneration() != auth.Generation || !obj.GetDeletionTimestamp().IsZero() {
		return fmt.Errorf("authorization %s changed after verification", key)
	}
	return nil
}

func checkEgressService(ctx context.Context, r client.Reader, key client.ObjectKey) error {
	var service corev1.Service
	if err := r.Get(ctx, key, &service); err != nil {
		return err
	}
	if service.UID == "" || !service.DeletionTimestamp.IsZero() {
		return fmt.Errorf("dependency Service %s is not active", key)
	}
	var endpoints discoveryv1.EndpointSliceList
	if err := r.List(ctx, &endpoints, client.InNamespace(key.Namespace),
		client.MatchingLabels{discoveryv1.LabelServiceName: key.Name}); err != nil {
		return err
	}
	isKubernetesAPI := key.Namespace == security.KubernetesAPINamespace && key.Name == security.KubernetesAPIName
	var apiEndpoints corev1.Endpoints
	if isKubernetesAPI {
		if err := r.Get(ctx, key, &apiEndpoints); err != nil {
			return err
		}
		if apiEndpoints.UID == "" || !apiEndpoints.DeletionTimestamp.IsZero() {
			return fmt.Errorf("Kubernetes API Endpoints are not active")
		}
	}
	for _, slice := range endpoints.Items {
		if !slice.DeletionTimestamp.IsZero() || len(slice.Ports) == 0 {
			continue
		}
		owner := metav1.GetControllerOf(&slice)
		if !isKubernetesAPI && (owner == nil || owner.Kind != "Service" || owner.Name != key.Name || owner.UID != service.UID) {
			continue
		}
		for _, e := range slice.Endpoints {
			if e.Conditions.Ready != nil && *e.Conditions.Ready && len(e.Addresses) > 0 &&
				(e.Conditions.Terminating == nil || !*e.Conditions.Terminating) &&
				(!isKubernetesAPI || matchesAPIDestination(slice.Ports, e.Addresses, apiEndpoints.Subsets)) {
				return nil
			}
		}
	}
	return fmt.Errorf("dependency Service %s has no ready endpoints", key)
}

// The apiserver can publish a selectorless EndpointSlice without an owner
// reference. Cross-check the actual HTTPS destination against the Endpoints
// used by the egress policy generator.
func matchesAPIDestination(ports []discoveryv1.EndpointPort, addresses []string, subsets []corev1.EndpointSubset) bool {
	for _, subset := range subsets {
		for _, p := range subset.Ports {
			if p.Name != "https" || p.Protocol != corev1.ProtocolTCP {
				continue
			}
			if !slices.ContainsFunc(ports, func(port discoveryv1.EndpointPort) bool {
				return port.Name != nil && *port.Name == p.Name && port.Port != nil && *port.Port == p.Port &&
					(port.Protocol == nil || *port.Protocol == corev1.ProtocolTCP)
			}) {
				continue
			}
			for _, address := range subset.Addresses {
				if address.IP != "" && slices.Contains(addresses, address.IP) {
					return true
				}
			}
		}
	}
	return false
}

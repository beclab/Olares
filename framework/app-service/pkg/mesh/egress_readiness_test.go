package mesh

import (
	"encoding/json"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gatewayv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func TestCheckEgressReadiness(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*egressReadiness, *corev1.Pod, *appsv1.Deployment, *gatewayv1alpha1.SharedRouteRegistry, *discoveryv1.EndpointSlice)
		want   string
	}{
		{name: "current workload and verified authorization"},
		{name: "rolled pod invalidates its inspection", mutate: func(_ *egressReadiness, p *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			p.UID = "replacement"
		}, want: "unverified"},
		{name: "sidecar is not ready", mutate: func(_ *egressReadiness, p *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			p.Status.InitContainerStatuses[0].Ready = false
		}, want: "enforcing mesh"},
		{name: "mesh interception cannot be exempted", mutate: func(e *egressReadiness, p *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			e.NativePods = []string{p.Name}
			p.Annotations[LinkerdSkipInboundPortsAnnotation] = "80"
		}, want: "enforcing mesh"},
		{name: "remaining replicas not rolled", mutate: func(_ *egressReadiness, _ *corev1.Pod, d *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			d.Spec.Replicas = ptr.To(int32(2))
		}, want: "rollout"},
		{name: "authorization evidence missing", mutate: func(e *egressReadiness, _ *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			e.Authorizations = nil
		}, want: "authorization evidence"},
		{name: "stale route readiness", mutate: func(_ *egressReadiness, _ *corev1.Pod, _ *appsv1.Deployment, r *gatewayv1alpha1.SharedRouteRegistry, _ *discoveryv1.EndpointSlice) {
			r.Status.Conditions[0].ObservedGeneration = 1
		}, want: "reviewed generation"},
		{name: "dependency unavailable", mutate: func(_ *egressReadiness, _ *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, e *discoveryv1.EndpointSlice) {
			e.Endpoints[0].Conditions.Ready = ptr.To(false)
		}, want: "no ready endpoints"},
		{name: "dependency slice belongs to old Service", mutate: func(_ *egressReadiness, _ *corev1.Pod, _ *appsv1.Deployment, _ *gatewayv1alpha1.SharedRouteRegistry, e *discoveryv1.EndpointSlice) {
			e.OwnerReferences[0].UID = "old-service"
		}, want: "no ready endpoints"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: "ns"}}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: ns.Name, UID: "pod",
				Annotations: map[string]string{LinkerdDefaultInboundPolicyAnnotation: InboundPolicyClusterAuthenticated}},
				Spec: corev1.PodSpec{InitContainers: []corev1.Container{
					{Name: "linkerd-init"},
					{Name: LinkerdProxyContainer, RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways),
						Env: []corev1.EnvVar{{Name: LinkerdProxyInboundPolicyEnv, Value: InboundPolicyClusterAuthenticated}}},
				}},
				Status: corev1.PodStatus{
					Conditions:            []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
					InitContainerStatuses: []corev1.ContainerStatus{{Name: LinkerdProxyContainer, Ready: true}},
				},
			}
			dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: ns.Name, Generation: 2},
				Spec:   appsv1.DeploymentSpec{Replicas: ptr.To(int32(1))},
				Status: appsv1.DeploymentStatus{ObservedGeneration: 2, UpdatedReplicas: 1, AvailableReplicas: 1, Replicas: 1}}
			var route gatewayv1alpha1.SharedRouteRegistry
			route.ObjectMeta = metav1.ObjectMeta{Name: "sys-desktop", Namespace: ns.Name, UID: "route", Generation: 2}
			route.Spec.RouteMode = gatewayv1alpha1.RouteModeGateway
			route.Status.ObservedGeneration = 2
			route.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Reconciled", ObservedGeneration: 2}}
			plan := egressReadiness{NamespaceUID: ns.UID, Pods: map[string]types.UID{pod.Name: pod.UID},
				Authorizations: []egressAuthorization{{APIVersion: gatewayv1alpha1.GroupVersion.String(),
					Kind: "SharedRouteRegistry", Namespace: ns.Name, Name: route.Name, UID: route.UID, Generation: route.Generation}},
				Dependencies: []string{"os-platform/cache"}}
			objs := []client.Object{ns, pod, dep, &route}
			for _, key := range []client.ObjectKey{
				{Namespace: "default", Name: "kubernetes"}, {Namespace: "kube-system", Name: "coredns"},
				{Namespace: "os-mesh", Name: "linkerd-dst"}, {Namespace: "os-mesh", Name: "linkerd-identity"},
				{Namespace: "os-mesh", Name: "linkerd-policy"}, {Namespace: "os-platform", Name: "cache"},
			} {
				uid := types.UID(key.Name)
				service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, UID: uid}}
				slice := &discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name,
					Labels: map[string]string{discoveryv1.LabelServiceName: key.Name}, OwnerReferences: []metav1.OwnerReference{
						{Kind: "Service", Name: key.Name, UID: uid, Controller: ptr.To(true)},
					}}, Ports: []discoveryv1.EndpointPort{{Port: ptr.To(int32(1234))}},
					Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"10.0.0.1"},
						Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}}}}
				if key.Name == "cache" && tc.mutate != nil {
					tc.mutate(&plan, pod, dep, &route, slice)
				}
				if key.Name == "kubernetes" {
					// A real apiserver slice can be ownerless.
					slice.OwnerReferences = nil
					slice.Ports[0].Name = ptr.To("https")
					objs = append(objs, &corev1.Endpoints{ObjectMeta: service.ObjectMeta, Subsets: []corev1.EndpointSubset{{
						Addresses: []corev1.EndpointAddress{{IP: slice.Endpoints[0].Addresses[0]}},
						Ports:     []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: *slice.Ports[0].Port}},
					}}})
				}
				objs = append(objs, service, slice)
			}
			record, err := json.Marshal(plan)
			if err != nil {
				t.Fatal(err)
			}
			objs = append(objs, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Namespace: EgressReadinessConfig.Namespace, Name: EgressReadinessConfig.Name},
				Data: map[string]string{ns.Name: string(record)}})
			s := runtime.NewScheme()
			for _, register := range []func(*runtime.Scheme) error{scheme.AddToScheme, gatewayv1alpha1.AddToScheme} {
				if err := register(s); err != nil {
					t.Fatal(err)
				}
			}
			c := fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
			got := CheckEgressReadiness(t.Context(), c, ns)
			if tc.want == "" && got != nil || tc.want != "" && (got == nil || !strings.Contains(got.Error(), tc.want)) {
				t.Fatalf("readiness = %v, want %q", got, tc.want)
			}
		})
	}
}

func TestEgressAuthorizationPolicyMustMatchInspectedRevision(t *testing.T) {
	ref := egressAuthorization{APIVersion: "policy.linkerd.io/v1alpha1", Kind: "AuthorizationPolicy",
		Namespace: "apps", Name: "allow-frontend", UID: "approved-policy", Generation: 2}
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": ref.APIVersion, "kind": ref.Kind,
		"metadata": map[string]any{"namespace": ref.Namespace, "name": ref.Name,
			"uid": string(ref.UID), "generation": ref.Generation},
	}}
	c := fake.NewClientBuilder().WithObjects(obj).Build()
	if err := checkEgressAuthorization(t.Context(), c, ref); err != nil {
		t.Fatal(err)
	}
	ref.Generation--
	if err := checkEgressAuthorization(t.Context(), c, ref); err == nil {
		t.Fatal("a different policy generation must invalidate the readiness record")
	}
}

func TestEgressAPIDestinationMustMatchPortAndAddress(t *testing.T) {
	ports := []discoveryv1.EndpointPort{{Name: ptr.To("https"), Port: ptr.To(int32(6443))}}
	subsets := []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "192.0.2.1"}},
		Ports:     []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 6443}},
	}}
	if !matchesAPIDestination(ports, []string{"192.0.2.1"}, subsets) {
		t.Fatal("expected the API destination in both objects to match")
	}
	if matchesAPIDestination(ports, []string{"192.0.2.2"}, subsets) {
		t.Fatal("slice claiming a different address cannot satisfy readiness")
	}
	ports[0].Port = ptr.To(int32(443))
	if matchesAPIDestination(ports, []string{"192.0.2.1"}, subsets) {
		t.Fatal("slice claiming a different HTTPS port cannot satisfy readiness")
	}
}

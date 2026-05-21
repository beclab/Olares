package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func newFixture(t *testing.T, svc *corev1.Service, srr *srrv1alpha1.SharedRouteRegistry) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("srr scheme: %v", err)
	}
	httpRouteGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
	httpRouteListGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRouteList"}
	scheme.AddKnownTypeWithName(httpRouteGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(httpRouteListGVK, &unstructured.UnstructuredList{})
	refGrantGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1beta1", Kind: "ReferenceGrant"}
	scheme.AddKnownTypeWithName(refGrantGVK, &unstructured.Unstructured{})

	objs := []client.Object{srr}
	if svc != nil {
		objs = append(objs, svc)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).WithObjects(objs...).Build()
}

func backendService(ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: ns},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ollama"},
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP},
			},
		},
	}
}

func logicalSRR(ns, name string) *srrv1alpha1.SharedRouteRegistry {
	return &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  ns,
			Generation: 7,
			UID:        types.UID("srr-uid"),
		},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{"01234567.*.olares.com"},
			Upstream: srrv1alpha1.UpstreamRef{
				ServiceName: "ollama",
				Port:        11434,
			},
		},
	}
}

func TestReconcileSharedRoute_GatewayMode_HappyPath(t *testing.T) {
	srr := logicalSRR("ollama-shared", "shared-ollama-api")
	svc := backendService("ollama-shared")
	c := newFixture(t, svc, srr)

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.Reason != ReasonReconciled {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.HTTPRouteName != "shared-ollama-api" {
		t.Fatalf("httpRouteName=%q", res.HTTPRouteName)
	}

	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama-api"}, hr); err != nil {
		t.Fatalf("get HTTPRoute: %v", err)
	}
	spec := hr.Object["spec"].(map[string]any)
	hosts := spec["hostnames"].([]any)
	if len(hosts) != 1 || hosts[0].(string) != "*.olares.com" {
		t.Fatalf("hostnames: %v", hosts)
	}
	rules := spec["rules"].([]any)
	matches := rules[0].(map[string]any)["matches"].([]any)
	headers := matches[0].(map[string]any)["headers"].([]any)
	if hdr := headers[0].(map[string]any); hdr["type"] != "RegularExpression" || hdr["name"] != "Host" {
		t.Fatalf("host header match: %+v", hdr)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: NetworkPolicyName}, np); err != nil {
		t.Fatalf("get NetworkPolicy: %v", err)
	}
	if got := np.Labels[ManagedByLabel]; got != ManagedByValue {
		t.Fatalf("NP managed-by=%q", got)
	}
	if len(np.OwnerReferences) != 1 || np.OwnerReferences[0].UID != srr.UID {
		t.Fatalf("NP ownerRefs: %+v", np.OwnerReferences)
	}
}

func TestReconcileSharedRoute_DirectMode_Cleans(t *testing.T) {
	srr := logicalSRR("ollama-shared", "shared-ollama-api")
	srr.Spec.RouteMode = srrv1alpha1.RouteModeDirect
	svc := backendService("ollama-shared")
	c := newFixture(t, svc, srr)

	hr := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata":   map[string]any{"name": "shared-ollama-api", "namespace": "ollama-shared"},
		"spec":       map[string]any{},
	}}
	if err := c.Create(context.Background(), hr); err != nil {
		t.Fatalf("seed HTTPRoute: %v", err)
	}
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ollama-shared", Name: NetworkPolicyName},
	}
	if err := c.Create(context.Background(), np); err != nil {
		t.Fatalf("seed NP: %v", err)
	}

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.Reason != ReasonDirectMode {
		t.Fatalf("unexpected result: %+v", res)
	}
	err = c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama-api"}, hr)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("HTTPRoute should be gone: err=%v", err)
	}
	err = c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: NetworkPolicyName}, np)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("NP should be gone: err=%v", err)
	}
}

func TestReconcileSharedRoute_BackendMissing(t *testing.T) {
	srr := logicalSRR("ollama-shared", "shared-ollama-api")
	c := newFixture(t, nil, srr)

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionFalse || res.Reason != ReasonBackendMissing {
		t.Fatalf("expected backend-missing result; got %+v", res)
	}
}

func TestReconcileSharedRoute_CrossNamespaceNPInUpstreamNS(t *testing.T) {
	srr := logicalSRR("ollamav2-brucedai", "shared-a5be2268-ollamav2")
	srr.Spec.Upstream = srrv1alpha1.UpstreamRef{
		ServiceName:      "sharedentrances-ollama",
		ServiceNamespace: "ollamaserver-shared",
		Port:             80,
	}
	svc := backendService("ollamaserver-shared")
	svc.Name = "sharedentrances-ollama"
	svc.Spec.Ports[0].Port = 80
	c := newFixture(t, svc, srr)

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollamaserver-shared", Name: NetworkPolicyName}, np); err != nil {
		t.Fatalf("get NetworkPolicy in upstream NS: %v", err)
	}
	if np.Spec.PodSelector.MatchLabels["app"] != "ollama" {
		t.Fatalf("podSelector: %+v", np.Spec.PodSelector.MatchLabels)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollamav2-brucedai", Name: NetworkPolicyName}, &networkingv1.NetworkPolicy{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("NP should not be in SRR namespace: err=%v", err)
	}

	rg := &unstructured.Unstructured{}
	rg.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1beta1", Kind: "ReferenceGrant"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollamaserver-shared", Name: referenceGrantName(srr)}, rg); err != nil {
		t.Fatalf("get ReferenceGrant: %v", err)
	}
}

func TestReconcileSharedRoute_InvalidSpec(t *testing.T) {
	srr := logicalSRR("ollama-shared", "shared-ollama-api")
	srr.Spec.HostPatterns = nil
	c := newFixture(t, backendService("ollama-shared"), srr)

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionFalse || res.Reason != ReasonInvalidSpec {
		t.Fatalf("expected invalid-spec; got %+v", res)
	}
}

func TestUpdateSRRStatus_Idempotent(t *testing.T) {
	srr := logicalSRR("ollama-shared", "shared-ollama-api")
	c := newFixture(t, backendService("ollama-shared"), srr)

	res := ReconcileResult{Status: metav1.ConditionTrue, Reason: ReasonReconciled, Message: "ok", HTTPRouteName: "shared-ollama-api"}
	if err := UpdateSRRStatus(context.Background(), c, srr, res); err != nil {
		t.Fatalf("UpdateSRRStatus: %v", err)
	}

	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama-api"}, got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status.HTTPRouteName != "shared-ollama-api" || got.Status.ObservedGeneration != srr.Generation {
		t.Fatalf("status not patched: %+v", got.Status)
	}
	if len(got.Status.Conditions) != 1 || got.Status.Conditions[0].Reason != ReasonReconciled {
		t.Fatalf("conditions: %+v", got.Status.Conditions)
	}

	if err := UpdateSRRStatus(context.Background(), c, got, res); err != nil {
		t.Fatalf("UpdateSRRStatus second pass: %v", err)
	}
}

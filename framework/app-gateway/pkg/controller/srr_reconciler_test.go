package controller

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gwapi "github.com/beclab/Olares/framework/app-gateway/pkg/api/v1alpha1"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatalf("add clientgo: %v", err)
	}
	if err := networkingv1.AddToScheme(s); err != nil {
		t.Fatalf("add networking: %v", err)
	}
	if err := gwapi.AddToScheme(s); err != nil {
		t.Fatalf("add gwapi: %v", err)
	}
	return s
}

func newReconciler(t *testing.T, initial ...client.Object) (*SRRReconciler, client.Client) {
	t.Helper()
	s := newScheme(t)
	c := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(initial...).
		WithStatusSubresource(&gwapi.SharedRouteRegistry{}).
		Build()
	return &SRRReconciler{
		Client:            c,
		Scheme:            s,
		GatewayNamespace:  "app-gateway",
		GatewayName:       "app-gateway",
		GatewaySectionRef: "http",
	}, c
}

func makeSRR(mode gwapi.RouteMode, port int32) *gwapi.SharedRouteRegistry {
	return &gwapi.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-ollama",
			Namespace: "ollama-shared",
			UID:       types.UID("srr-uid"),
		},
		Spec: gwapi.SharedRouteRegistrySpec{
			RouteMode:    mode,
			HostPatterns: []string{"abc.shared.example.com"},
			Upstream: gwapi.UpstreamRef{
				ServiceName: "ollama",
				Port:        port,
			},
		},
	}
}

func makeService(port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ollama"},
			Ports:    []corev1.ServicePort{{Name: "http", Port: port, Protocol: corev1.ProtocolTCP}},
		},
	}
}

func TestReconcile_GatewayMode_CreatesHTTPRouteAndNP(t *testing.T) {
	srr := makeSRR(gwapi.RouteModeGateway, 11434)
	svc := makeService(11434)
	r, c := newReconciler(t, srr, svc)

	ctx := context.Background()
	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: srr.Name, Namespace: srr.Namespace}}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, hr); err != nil {
		t.Fatalf("HTTPRoute missing: %v", err)
	}
	parents, _, _ := unstructuredFieldList(hr.Object, "spec", "parentRefs")
	if len(parents) != 1 {
		t.Fatalf("want 1 parentRef, got %d", len(parents))
	}
	p, _ := parents[0].(map[string]any)
	if p["namespace"] != "app-gateway" || p["name"] != "app-gateway" || p["sectionName"] != "http" {
		t.Fatalf("unexpected parentRef: %+v", p)
	}
	hostnames, _, _ := unstructuredFieldList(hr.Object, "spec", "hostnames")
	if len(hostnames) != 1 || hostnames[0] != "abc.shared.example.com" {
		t.Fatalf("hostnames mismatch: %v", hostnames)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: NetworkPolicyName}, np); err != nil {
		t.Fatalf("NetworkPolicy missing: %v", err)
	}
	if got := np.Spec.PodSelector.MatchLabels["app"]; got != "ollama" {
		t.Fatalf("podSelector mismatch: %+v", np.Spec.PodSelector)
	}
	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("want 1 ingress rule, got %d", len(np.Spec.Ingress))
	}
	from := np.Spec.Ingress[0].From
	if len(from) != 1 || from[0].NamespaceSelector == nil || from[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "app-gateway" {
		t.Fatalf("ingress.from mismatch: %+v", from)
	}

	got := &gwapi.SharedRouteRegistry{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, got); err != nil {
		t.Fatalf("get SRR: %v", err)
	}
	if got.Status.HTTPRouteName != srr.Name {
		t.Fatalf("status.httpRouteName = %q want %q", got.Status.HTTPRouteName, srr.Name)
	}
	if !readyTrue(got) {
		t.Fatalf("Ready condition not True: %+v", got.Status.Conditions)
	}
}

func TestReconcile_DirectMode_RemovesArtifacts(t *testing.T) {
	srr := makeSRR(gwapi.RouteModeGateway, 11434)
	svc := makeService(11434)
	r, c := newReconciler(t, srr, svc)
	ctx := context.Background()
	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: srr.Name, Namespace: srr.Namespace}}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got := &gwapi.SharedRouteRegistry{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, got); err != nil {
		t.Fatalf("get srr: %v", err)
	}
	patched := got.DeepCopy()
	patched.Spec.RouteMode = gwapi.RouteModeDirect
	if err := c.Patch(ctx, patched, client.MergeFrom(got)); err != nil {
		t.Fatalf("flip routeMode: %v", err)
	}

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: srr.Name, Namespace: srr.Namespace}}); err != nil {
		t.Fatalf("reconcile direct: %v", err)
	}

	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, hr); !apierrors.IsNotFound(err) {
		t.Fatalf("HTTPRoute should be gone, err=%v", err)
	}
	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: NetworkPolicyName}, np); !apierrors.IsNotFound(err) {
		t.Fatalf("NP should be gone, err=%v", err)
	}
}

func TestReconcile_GatewayMode_BackendMissing(t *testing.T) {
	srr := makeSRR(gwapi.RouteModeGateway, 11434)
	r, c := newReconciler(t, srr)
	ctx := context.Background()

	res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: srr.Name, Namespace: srr.Namespace}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !res.Requeue {
		t.Fatalf("expected requeue when backend missing")
	}
	got := &gwapi.SharedRouteRegistry{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, got); err != nil {
		t.Fatalf("get srr: %v", err)
	}
	if readyTrue(got) {
		t.Fatalf("Ready should be False; got %+v", got.Status.Conditions)
	}
	found := false
	for _, cond := range got.Status.Conditions {
		if cond.Type == ConditionReady && cond.Reason == ReasonBackendMissing {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected BackendServiceMissing reason; got %+v", got.Status.Conditions)
	}
}

func readyTrue(srr *gwapi.SharedRouteRegistry) bool {
	for _, c := range srr.Status.Conditions {
		if c.Type == ConditionReady && c.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func unstructuredFieldList(obj map[string]any, path ...string) ([]any, bool, error) {
	v, found, err := unstructured.NestedSlice(obj, path...)
	return v, found, err
}

// PR-7: logical hostPattern <hash8>.*.<domain> must produce
// spec.hostnames=["*.<domain>"] AND a Host RegularExpression header match.
func TestReconcile_GatewayMode_LogicalPattern(t *testing.T) {
	srr := &gwapi.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-a5be2268-ollamav2",
			Namespace: "ollama-shared",
			UID:       types.UID("srr-v2-uid"),
		},
		Spec: gwapi.SharedRouteRegistrySpec{
			RouteMode:    gwapi.RouteModeGateway,
			HostPatterns: []string{"a5be2268.*.olares.com"},
			Upstream: gwapi.UpstreamRef{
				ServiceName: "ollama",
				Port:        11434,
			},
		},
	}
	svc := makeService(11434)
	r, c := newReconciler(t, srr, svc)
	ctx := context.Background()

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: srr.Name, Namespace: srr.Namespace}}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(ctx, types.NamespacedName{Namespace: srr.Namespace, Name: srr.Name}, hr); err != nil {
		t.Fatalf("HTTPRoute missing: %v", err)
	}

	hostnames, _, _ := unstructuredFieldList(hr.Object, "spec", "hostnames")
	if len(hostnames) != 1 || hostnames[0] != "*.olares.com" {
		t.Fatalf("hostnames mismatch: got %v, want [*.olares.com]", hostnames)
	}

	rules, _, _ := unstructuredFieldList(hr.Object, "spec", "rules")
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	rule, _ := rules[0].(map[string]any)
	matches, _ := rule["matches"].([]any)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m, _ := matches[0].(map[string]any)
	hdrs, _ := m["headers"].([]any)
	if len(hdrs) != 1 {
		t.Fatalf("expected 1 header match, got %v", hdrs)
	}
	h, _ := hdrs[0].(map[string]any)
	if h["name"] != "Host" || h["type"] != "RegularExpression" {
		t.Fatalf("header match wrong: %+v", h)
	}
	val, _ := h["value"].(string)
	want := HostRegexValue(LogicalPattern{Hash8: "a5be2268", PlatformDomain: "olares.com"})
	if val != want {
		t.Fatalf("Host regex value mismatch: got %q want %q", val, want)
	}
}

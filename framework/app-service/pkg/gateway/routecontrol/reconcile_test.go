package routecontrol

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func mustHTTPRouteSectionName(t *testing.T, route *unstructured.Unstructured) string {
	t.Helper()
	parentRefs, found, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
	if err != nil || !found || len(parentRefs) == 0 {
		t.Fatalf("spec.parentRefs missing: found=%v err=%v", found, err)
	}
	parentRef, ok := parentRefs[0].(map[string]any)
	if !ok {
		t.Fatalf("parentRefs[0] type = %T, want map[string]any", parentRefs[0])
	}
	section, ok := parentRef["sectionName"].(string)
	if !ok {
		t.Fatalf("parentRefs[0].sectionName type = %T, want string", parentRef["sectionName"])
	}
	return section
}

func mustHTTPRouteParentRef(t *testing.T, route *unstructured.Unstructured) map[string]any {
	t.Helper()
	parentRefs, found, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
	if err != nil || !found || len(parentRefs) == 0 {
		t.Fatalf("spec.parentRefs missing: found=%v err=%v", found, err)
	}
	parentRef, ok := parentRefs[0].(map[string]any)
	if !ok {
		t.Fatalf("parentRefs[0] type = %T, want map[string]any", parentRefs[0])
	}
	return parentRef
}

func mustHTTPRouteSectionNameAbsent(t *testing.T, route *unstructured.Unstructured) {
	t.Helper()
	parentRef := mustHTTPRouteParentRef(t, route)
	if _, exists := parentRef["sectionName"]; exists {
		t.Fatalf("parentRefs[0].sectionName exists, want absent for application")
	}
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := srrv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := networkingv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	gw := schema.GroupVersion{Group: "gateway.networking.k8s.io", Version: "v1"}
	s.AddKnownTypeWithName(gw.WithKind("HTTPRoute"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(gw.WithKind("HTTPRouteList"), &unstructured.UnstructuredList{})
	eg := schema.GroupVersion{Group: "gateway.envoyproxy.io", Version: "v1alpha1"}
	s.AddKnownTypeWithName(eg.WithKind("BackendTrafficPolicy"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(eg.WithKind("BackendTrafficPolicyList"), &unstructured.UnstructuredList{})
	return s
}

func mustHTTPRouteFirstRule(t *testing.T, route *unstructured.Unstructured) map[string]any {
	t.Helper()
	rules, found, err := unstructured.NestedSlice(route.Object, "spec", "rules")
	if err != nil || !found || len(rules) == 0 {
		t.Fatalf("spec.rules missing: found=%v err=%v", found, err)
	}
	rule, ok := rules[0].(map[string]any)
	if !ok {
		t.Fatalf("rules[0] type=%T, want map[string]any", rules[0])
	}
	return rule
}

func TestResolveServicePort(t *testing.T) {
	svc := &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{
		{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
		{Name: "metrics", Port: 9090, Protocol: corev1.ProtocolTCP},
	}}}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{Port: 9090}); err != nil || p != 9090 {
		t.Errorf("by port = %d, %v", p, err)
	}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{PortName: "http"}); err != nil || p != 80 {
		t.Errorf("by name = %d, %v", p, err)
	}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{}); err != nil || p != 80 {
		t.Errorf("default = %d, %v", p, err)
	}
	if _, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{Port: 1234}); err == nil {
		t.Error("missing port should error")
	}
}

func TestReconcileSharedRouteGatewayMode(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-web", Namespace: "demo-shared", UID: "uid-1"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"ab12cd34.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.HTTPRouteName != "shared-demo-web" {
		t.Fatalf("result = %+v", res)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-web"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	if got := mustHTTPRouteSectionName(t, route); got != "http" {
		t.Fatalf("HTTPRoute sectionName = %q, want http", got)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: NetworkPolicyName}, np); err != nil {
		t.Fatalf("NetworkPolicy not created: %v", err)
	}
	// The NP must carry the route-control component label so the
	// security-controller namespace sweep does not delete it.
	if np.Labels[ManagedByLabel] != ManagedByValue {
		t.Fatalf("NetworkPolicy managed-by = %q, want %q", np.Labels[ManagedByLabel], ManagedByValue)
	}
	if np.Labels[RouteControlComponentLabel] != RouteControlComponentValue {
		t.Fatalf("NetworkPolicy component = %q, want %q", np.Labels[RouteControlComponentLabel], RouteControlComponentValue)
	}
}

func TestReconcileSharedRouteGatewayModeApplicationSection(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-app", Namespace: "demo-shared", UID: "uid-app"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.viewer.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue {
		t.Fatalf("result status = %s, want True", res.Status)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-app"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	parentRef := mustHTTPRouteParentRef(t, route)
	if got := parentRef["namespace"]; got != "os-gateway" {
		t.Fatalf("parentRefs[0].namespace = %v, want os-gateway", got)
	}
	if got := parentRef["name"]; got != "app-gateway" {
		t.Fatalf("parentRefs[0].name = %v, want app-gateway", got)
	}
	mustHTTPRouteSectionNameAbsent(t, route)
}

func TestReconcileSharedRouteGatewayModeEmptyEntranceClassDefaultsToShared(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-empty", Namespace: "demo-shared", UID: "uid-empty"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{"empty.shared.olares.com"},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-empty"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	if got := mustHTTPRouteSectionName(t, route); got != "http" {
		t.Fatalf("HTTPRoute sectionName = %q, want http for empty EntranceClass", got)
	}
}

func TestReconcileSharedRouteGatewayModeApplicationHostContracts(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	const pattern = "ab12cd34.*.olares.com"
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-app-host", Namespace: "demo-shared", UID: "uid-app-host"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{pattern},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-app-host"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}

	hostnames, found, err := unstructured.NestedSlice(route.Object, "spec", "hostnames")
	if err != nil || !found || len(hostnames) != 1 {
		t.Fatalf("spec.hostnames invalid: found=%v len=%d err=%v", found, len(hostnames), err)
	}
	if got := hostnames[0]; got != "*.olares.com" {
		t.Fatalf("hostname = %v, want *.olares.com", got)
	}

	rules, found, err := unstructured.NestedSlice(route.Object, "spec", "rules")
	if err != nil || !found || len(rules) == 0 {
		t.Fatalf("spec.rules invalid: found=%v len=%d err=%v", found, len(rules), err)
	}
	firstRule, ok := rules[0].(map[string]any)
	if !ok {
		t.Fatalf("spec.rules[0] type = %T, want map[string]any", rules[0])
	}
	matches, ok := firstRule["matches"].([]any)
	if !ok || len(matches) == 0 {
		t.Fatalf("spec.rules[0].matches invalid: %v", firstRule["matches"])
	}
	firstMatch, ok := matches[0].(map[string]any)
	if !ok {
		t.Fatalf("spec.rules[0].matches[0] type = %T, want map[string]any", matches[0])
	}
	headers, ok := firstMatch["headers"].([]any)
	if !ok || len(headers) == 0 {
		t.Fatalf("headers missing in first match: %v", firstMatch["headers"])
	}
	header, ok := headers[0].(map[string]any)
	if !ok {
		t.Fatalf("first header type = %T, want map[string]any", headers[0])
	}
	wantHeader, ok := HostHeaderMatch(pattern)
	if !ok {
		t.Fatalf("HostHeaderMatch(%q) returned !ok", pattern)
	}
	if got := header["value"]; got != wantHeader["value"] {
		t.Fatalf("host header regex = %v, want %v", got, wantHeader["value"])
	}
}

func TestReconcileSharedRouteGatewayModeApplicationMultiEntranceHostPattern(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	const pattern = "e31111940.*.olares.cn"
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-e3111194-terminal", Namespace: "demo-shared", UID: "uid-app-multi"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{pattern},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue {
		t.Fatalf("result status = %s, want True (%s)", res.Status, res.Message)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "app-e3111194-terminal"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	hostnames, found, err := unstructured.NestedSlice(route.Object, "spec", "hostnames")
	if err != nil || !found || len(hostnames) != 1 {
		t.Fatalf("spec.hostnames invalid: found=%v len=%d err=%v", found, len(hostnames), err)
	}
	if got := hostnames[0]; got != "*.olares.cn" {
		t.Fatalf("hostname = %v, want *.olares.cn", got)
	}
	wantHeader, ok := HostHeaderMatch(pattern)
	if !ok {
		t.Fatalf("HostHeaderMatch(%q) returned !ok", pattern)
	}
	rules, _, err := unstructured.NestedSlice(route.Object, "spec", "rules")
	if err != nil || len(rules) == 0 {
		t.Fatalf("spec.rules: %v err=%v", rules, err)
	}
	firstRule := rules[0].(map[string]any)
	matches := firstRule["matches"].([]any)
	header := matches[0].(map[string]any)["headers"].([]any)[0].(map[string]any)
	if got := header["value"]; got != wantHeader["value"] {
		t.Fatalf("host header regex = %v, want %v", got, wantHeader["value"])
	}
}

func TestReconcileSharedRouteGatewayModeApplicationBootstrapReady(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-app-bootstrap", Namespace: "demo-shared", UID: "uid-app-bootstrap"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"ab12cd34.*.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.Reason != ReasonReconciled {
		t.Fatalf("result = %+v, want Ready=True/Reconciled", res)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-app-bootstrap"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	mustHTTPRouteSectionNameAbsent(t, route)
}

func TestReconcileSharedRouteDirectMode(t *testing.T) {
	s := testScheme(t)
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-web", Namespace: "demo-shared", UID: "uid-2"},
		Spec:       srrv1alpha1.SharedRouteRegistrySpec{RouteMode: srrv1alpha1.RouteModeDirect},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(srr).Build()
	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute direct: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.Reason != ReasonDirectMode {
		t.Fatalf("direct result = %+v", res)
	}
}

func TestApplyHTTPRouteMaterializesTimeouts(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-timeout", Namespace: "demo-shared", UID: "uid-timeout"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"timeout.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-timeout"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}

	rule := mustHTTPRouteFirstRule(t, route)
	timeouts, ok := rule["timeouts"].(map[string]any)
	if !ok {
		t.Fatalf("timeouts type=%T, want map[string]any", rule["timeouts"])
	}
	if got := timeouts["backendRequest"]; got != "600s" {
		t.Fatalf("timeouts.backendRequest=%v, want 600s", got)
	}
	if got := timeouts["request"]; got != "600s" {
		t.Fatalf("timeouts.request=%v, want 600s", got)
	}
}

func TestApplyHTTPRouteProbeFailureStillWritesTimeouts(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-probe-fail", Namespace: "demo-shared", UID: "uid-probe-fail"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"probe-fail.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-demo-probe-fail",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"rules": []any{
				map[string]any{
					"timeouts": map[string]any{
						"request": int64(12345), // malformed on purpose: triggers probe failure
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-probe-fail"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}
	rule := mustHTTPRouteFirstRule(t, route)
	timeouts, ok := rule["timeouts"].(map[string]any)
	if !ok {
		t.Fatalf("timeouts type=%T, want map[string]any", rule["timeouts"])
	}
	if got := timeouts["request"]; got != "600s" {
		t.Fatalf("timeouts.request=%v, want 600s", got)
	}
}

func TestApplyHTTPRouteDiffOnlyTimeouts(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-diff-timeout", Namespace: "demo-shared", UID: "uid-diff-timeout"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"diff-timeout.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-demo-diff-timeout",
			"namespace": "demo-shared",
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  "shared-demo-diff-timeout",
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"group":       "gateway.networking.k8s.io",
					"kind":        "Gateway",
					"namespace":   "os-gateway",
					"name":        "app-gateway",
					"sectionName": "http",
				},
			},
			"hostnames": []any{"diff-timeout.shared.olares.com"},
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{
							"path": map[string]any{"type": "PathPrefix", "value": "/"},
						},
					},
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      "demo-svc",
							"namespace": "demo-shared",
							"port":      int64(8080),
							"weight":    int64(1),
						},
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute).Build()

	before := &unstructured.Unstructured{}
	before.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-diff-timeout"}, before); err != nil {
		t.Fatalf("get before route: %v", err)
	}
	beforeRule := mustHTTPRouteFirstRule(t, before)
	beforeMatches := beforeRule["matches"]
	beforeBackendRefs := beforeRule["backendRefs"]

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	after := &unstructured.Unstructured{}
	after.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-diff-timeout"}, after); err != nil {
		t.Fatalf("get after route: %v", err)
	}
	afterRule := mustHTTPRouteFirstRule(t, after)
	if !reflect.DeepEqual(beforeMatches, afterRule["matches"]) {
		t.Fatalf("matches changed after timeout materialization")
	}
	if !reflect.DeepEqual(beforeBackendRefs, afterRule["backendRefs"]) {
		t.Fatalf("backendRefs changed after timeout materialization")
	}
	timeouts, ok := afterRule["timeouts"].(map[string]any)
	if !ok {
		t.Fatalf("timeouts type=%T, want map[string]any", afterRule["timeouts"])
	}
	if got := timeouts["request"]; got != "600s" {
		t.Fatalf("timeouts.request=%v, want 600s", got)
	}
}

func mustRouteRequestTimeout(t *testing.T, route *unstructured.Unstructured) string {
	t.Helper()
	rule := mustHTTPRouteFirstRule(t, route)
	timeouts, ok := rule["timeouts"].(map[string]any)
	if !ok {
		t.Fatalf("timeouts type=%T, want map[string]any", rule["timeouts"])
	}
	raw, ok := timeouts["request"].(string)
	if !ok {
		t.Fatalf("timeouts.request type=%T, want string", timeouts["request"])
	}
	return raw
}

func TestProbePreservesExternalFloorAcrossReconcile(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-probe-floor", Namespace: "demo-shared", UID: "uid-probe-floor"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"floor.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-probe-floor",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"group":       "gateway.networking.k8s.io",
					"kind":        "Gateway",
					"namespace":   "os-gateway",
					"name":        "app-gateway",
					"sectionName": "http",
				},
			},
			"hostnames": []any{"floor.shared.olares.com"},
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{
							"path": map[string]any{"type": "PathPrefix", "value": "/"},
						},
					},
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      "demo-svc",
							"namespace": "demo-shared",
							"port":      int64(8080),
							"weight":    int64(1),
						},
					},
					"timeouts": map[string]any{
						"backendRequest": "1800s",
						"request":        "1800s",
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute).Build()

	for i := 0; i < 2; i++ {
		if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
			t.Fatalf("ReconcileSharedRoute run=%d: %v", i+1, err)
		}
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-probe-floor"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}
	if got := mustRouteRequestTimeout(t, route); got != "1800s" {
		t.Fatalf("timeouts.request=%s, want 1800s", got)
	}
	if got := route.GetAnnotations()[TimeoutFloorAnnotation]; got != "1800s" {
		t.Fatalf("annotation %s=%q, want 1800s", TimeoutFloorAnnotation, got)
	}
}

func TestProbeBTPParseErrorKeepsRouteFloor(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-btp-parse", Namespace: "demo-shared", UID: "uid-btp-parse"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"btp.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-btp-parse",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"group":       "gateway.networking.k8s.io",
					"kind":        "Gateway",
					"namespace":   "os-gateway",
					"name":        "app-gateway",
					"sectionName": "http",
				},
			},
			"hostnames": []any{"btp.shared.olares.com"},
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{
							"path": map[string]any{"type": "PathPrefix", "value": "/"},
						},
					},
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      "demo-svc",
							"namespace": "demo-shared",
							"port":      int64(8080),
							"weight":    int64(1),
						},
					},
					"timeouts": map[string]any{
						"backendRequest": "1800s",
						"request":        "1800s",
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	btp := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "BackendTrafficPolicy",
		"metadata": map[string]any{
			"name":      "btp-shared-btp-parse",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  "shared-btp-parse",
			},
			"timeout": map[string]any{
				"http": map[string]any{
					"requestTimeout": "bad",
				},
			},
		},
	}}
	btp.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "BackendTrafficPolicy"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute, btp).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-btp-parse"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}
	if got := mustRouteRequestTimeout(t, route); got != "1800s" {
		t.Fatalf("timeouts.request=%s, want 1800s", got)
	}
}

func TestEffectiveRespectsAnnotationWhenManaged(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-annotation-floor", Namespace: "demo-shared", UID: "uid-annotation-floor"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"annotation.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-annotation-floor",
			"namespace": "demo-shared",
			"labels": map[string]any{
				ManagedByLabel: ManagedByValue,
				InstanceLabel:  "shared-annotation-floor",
			},
			"annotations": map[string]any{
				TimeoutFloorAnnotation: "1800s",
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"group":       "gateway.networking.k8s.io",
					"kind":        "Gateway",
					"namespace":   "os-gateway",
					"name":        "app-gateway",
					"sectionName": "http",
				},
			},
			"hostnames": []any{"annotation.shared.olares.com"},
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{
							"path": map[string]any{"type": "PathPrefix", "value": "/"},
						},
					},
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      "demo-svc",
							"namespace": "demo-shared",
							"port":      int64(8080),
							"weight":    int64(1),
						},
					},
					"timeouts": map[string]any{
						"backendRequest": "600s",
						"request":        "600s",
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-annotation-floor"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}
	if got := mustRouteRequestTimeout(t, route); got != "1800s" {
		t.Fatalf("timeouts.request=%s, want 1800s", got)
	}
}

func TestApplyHTTPRoutePersistsFloorAnnotationWhenSpecUnchanged(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-spec-unchanged", Namespace: "demo-shared", UID: "uid-spec-unchanged"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"spec-unchanged.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	existingRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      "shared-spec-unchanged",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"group":       "gateway.networking.k8s.io",
					"kind":        "Gateway",
					"namespace":   "os-gateway",
					"name":        "app-gateway",
					"sectionName": "http",
				},
			},
			"hostnames": []any{"spec-unchanged.shared.olares.com"},
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{
							"path": map[string]any{"type": "PathPrefix", "value": "/"},
						},
					},
					"backendRefs": []any{
						map[string]any{
							"group":     "",
							"kind":      "Service",
							"name":      "demo-svc",
							"namespace": "demo-shared",
							"port":      int64(8080),
							"weight":    int64(1),
						},
					},
					"timeouts": map[string]any{
						"backendRequest": "1800s",
						"request":        "1800s",
					},
				},
			},
		},
	}}
	existingRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, existingRoute).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-spec-unchanged"}, route); err != nil {
		t.Fatalf("get route: %v", err)
	}
	if got := route.GetLabels()[ManagedByLabel]; got != ManagedByValue {
		t.Fatalf("managed-by label=%q, want %q", got, ManagedByValue)
	}
	if got := route.GetLabels()[InstanceLabel]; got != "shared-spec-unchanged" {
		t.Fatalf("instance label=%q, want shared-spec-unchanged", got)
	}
	if got := route.GetAnnotations()[TimeoutFloorAnnotation]; got != "1800s" {
		t.Fatalf("annotation %s=%q, want 1800s", TimeoutFloorAnnotation, got)
	}
	if got := mustRouteRequestTimeout(t, route); got != "1800s" {
		t.Fatalf("timeouts.request=%s, want 1800s", got)
	}
}

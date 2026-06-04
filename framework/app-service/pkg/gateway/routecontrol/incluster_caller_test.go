package routecontrol

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

func TestIsMeshMandatoryCallerNamespace(t *testing.T) {
	cases := []struct {
		ns   string
		want bool
	}{
		{"user-space-alice", true},
		{"user-system-bob", true},
		{"litellm-alice", true},
		{"os-network", false},
		{"app-gateway", false},
		{"linkerd", false},
		{"kube-system", false},
		{"os-framework", false},
	}
	for _, tc := range cases {
		if got := isMeshMandatoryCallerNamespace(tc.ns); got != tc.want {
			t.Fatalf("isMeshMandatoryCallerNamespace(%q) = %v, want %v", tc.ns, got, tc.want)
		}
	}
}

// TestCallerReconciler_optInInjectsAndWritesGatewayIngress is the NP-minimal
// v1.0 happy path: opt-in caller NS gets linkerd.io/inject=enabled and the
// gateway-side singleton ingress NP appears (NO managed caller egress).
func TestCallerReconciler_optInInjectsAndWritesGatewayIngress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-gateway"}},
			&appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name: "litellm",
					Annotations: map[string]string{
						gateway.AnnotationInCluster: gateway.InClusterGateway,
					},
				},
				Spec: appv1alpha1.ApplicationSpec{
					Name:      "litellm",
					Namespace: ns,
					Settings:  map[string]string{"clusterAppRef": "ollamav2"},
				},
			},
		).Build()

	r := &CallerReconciler{Client: c, GatewayNS: "app-gateway"}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("NP-minimal v1.0: caller egress %q must not be created", name)
		}
	}

	var gwNP networkingv1.NetworkPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &gwNP); err != nil {
		t.Fatalf("gateway caller ingress NP missing: %v", err)
	}
	if len(gwNP.Spec.PodSelector.MatchLabels) != 0 || len(gwNP.Spec.PodSelector.MatchExpressions) != 0 {
		t.Fatalf("gateway caller ingress podSelector must be empty, got %#v", gwNP.Spec.PodSelector)
	}
	if len(gwNP.Spec.Ingress) != 1 || len(gwNP.Spec.Ingress[0].Ports) != 0 {
		t.Fatalf("gateway caller ingress must omit Ports, got %#v", gwNP.Spec.Ingress)
	}

	var nsObj corev1.Namespace
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, &nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if nsObj.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("inject = %q", nsObj.Annotations[LinkerdInjectAnnotation])
	}
	if nsObj.Labels[security.NamespaceInClusterCallerLabel] != "true" {
		t.Fatalf("caller label = %q", nsObj.Labels[security.NamespaceInClusterCallerLabel])
	}
	if nsObj.Annotations[LinkerdSkipInboundPortsAnnotation] != PureCallerInboundSkipPorts {
		t.Fatalf("skip-inbound-ports = %q", nsObj.Annotations[LinkerdSkipInboundPortsAnnotation])
	}
	if nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation] != "1-8080,8083-65535" {
		t.Fatalf("skip-outbound-ports = %q", nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation])
	}
}

func TestCallerReconciler_osNetworkNoOp(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-network"}}).
		Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), "os-network"); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	var np networkingv1.NetworkPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &np); err == nil {
		t.Fatal("os-network reconcile must not create gateway caller ingress NP")
	}
}

// TestCallerReconciler_optInAlsoGCsLegacyEgress is the upgrade-path safety net:
// clusters that ran pre-v1.0 still carry caller egress NPs in opted-in caller
// namespaces. opt-out cleanup never fires while the caller remains opted-in,
// so the opt-in reconcile branch must GC the legacy NPs every loop.
func TestCallerReconciler_optInAlsoGCsLegacyEgress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-gateway"}},
			security.NewCallerMeshEgressNP(ns),
			security.NewCallerToAppGatewayEgressNP(ns, "app-gateway"),
			security.NewCallerDNSEgressNP(ns),
			security.NewCallerMiddlewareEgressNP(ns, "user-system-alice"),
			&appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name: "litellm",
					Annotations: map[string]string{
						gateway.AnnotationInCluster: gateway.InClusterGateway,
					},
				},
				Spec: appv1alpha1.ApplicationSpec{
					Name:      "litellm",
					Namespace: ns,
					Settings:  map[string]string{"clusterAppRef": "ollamav2"},
				},
			},
		).Build()

	r := &CallerReconciler{Client: c, GatewayNS: "app-gateway"}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("legacy caller egress %q must be GCed on opt-in reconcile", name)
		}
	}

	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &networkingv1.NetworkPolicy{}); err != nil {
		t.Fatalf("gateway ingress NP must still be present: %v", err)
	}
}

// TestCallerReconciler_optOutGCsLegacyEgress validates the upgrade-path cleanup:
// any pre-v1.0 caller egress NPs still in the namespace are GCed when the app
// opts out (or is deleted).
func TestCallerReconciler_optOutGCsLegacyEgress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        ns,
				Annotations: map[string]string{LinkerdInjectAnnotation: LinkerdInjectEnabled},
			}},
			security.NewCallerMeshEgressNP(ns),
			security.NewCallerToAppGatewayEgressNP(ns, "app-gateway"),
			security.NewCallerDNSEgressNP(ns),
			security.NewCallerMiddlewareEgressNP(ns, "user-system-alice"),
		).Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("legacy NP %q should be GCed on opt-out", name)
		}
	}
}

func TestComputeSkipOutboundPorts_TC201(t *testing.T) {
	got, err := ComputeSkipOutboundPorts([]int32{8081, 8082})
	if err != nil {
		t.Fatalf("ComputeSkipOutboundPorts: %v", err)
	}
	if got != "1-8080,8083-65535" {
		t.Fatalf("ComputeSkipOutboundPorts([8081,8082]) = %q", got)
	}
}

func TestComputeSkipOutboundPorts_TC201b(t *testing.T) {
	cases := [][]int32{nil, {}}
	for _, in := range cases {
		got, err := ComputeSkipOutboundPorts(in)
		if err == nil {
			t.Fatalf("ComputeSkipOutboundPorts(%v) expected error", in)
		}
		if got != "" {
			t.Fatalf("ComputeSkipOutboundPorts(%v) = %q, want empty", in, got)
		}
	}
}

func TestComputeSkipOutboundPorts_TC202(t *testing.T) {
	got, err := ComputeSkipOutboundPorts([]int32{8081, 8082})
	if err != nil {
		t.Fatalf("ComputeSkipOutboundPorts: %v", err)
	}
	tokens := strings.Split(got, ",")
	if containsExact(tokens, "80") {
		t.Fatalf("skip-outbound tokens must not contain standalone 80: %v", tokens)
	}
	if containsExact(tokens, "8081") {
		t.Fatalf("skip-outbound tokens must not contain standalone 8081: %v", tokens)
	}
	if containsExact(tokens, "8082") {
		t.Fatalf("skip-outbound tokens must not contain standalone 8082: %v", tokens)
	}
}

func TestComputeSkipOutboundPorts_TCT1701(t *testing.T) {
	got, err := ComputeSkipOutboundPorts([]int32{8081, 8082})
	if err != nil {
		t.Fatalf("ComputeSkipOutboundPorts: %v", err)
	}
	if !portInSkipRange(80, got) {
		t.Fatalf("port 80 must stay in skip-outbound range, got %q", got)
	}
}

func TestEnsureCallerNamespaceLinkerdSkipPorts_TC203(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc203"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}).
		Build()
	if err := ensureCallerNamespaceLinkerdSkipPorts(context.Background(), c, ns); err != nil {
		t.Fatalf("ensureCallerNamespaceLinkerdSkipPorts: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if nsObj.Annotations[LinkerdSkipInboundPortsAnnotation] != PureCallerInboundSkipPorts {
		t.Fatalf("skip-inbound-ports = %q", nsObj.Annotations[LinkerdSkipInboundPortsAnnotation])
	}
	if nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation] != "1-8080,8083-65535" {
		t.Fatalf("skip-outbound-ports = %q", nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation])
	}
}

func TestEnsureCallerNamespaceLinkerdSkipPorts_TC204(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc204"
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "srr-ready", Namespace: ns},
		Status: srrv1alpha1.SharedRouteRegistryStatus{
			Conditions: []metav1.Condition{{
				Type:   ConditionReady,
				Status: metav1.ConditionTrue,
				Reason: ReasonReconciled,
			}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			srr,
		).Build()
	if err := ensureCallerNamespaceLinkerdSkipPorts(context.Background(), c, ns); err != nil {
		t.Fatalf("ensureCallerNamespaceLinkerdSkipPorts: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if nsObj.Annotations[LinkerdSkipInboundPortsAnnotation] != OlaresEnvoyInboundSkipPorts {
		t.Fatalf("skip-inbound-ports = %q", nsObj.Annotations[LinkerdSkipInboundPortsAnnotation])
	}
	if nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation] != "1-8080,8083-65535" {
		t.Fatalf("skip-outbound-ports = %q", nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation])
	}
}

func TestEnsureCallerNamespaceLinkerdSkipPorts_TC204b(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc204b"
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "srr-not-ready", Namespace: ns},
		Status: srrv1alpha1.SharedRouteRegistryStatus{
			Conditions: []metav1.Condition{{
				Type:   ConditionReady,
				Status: metav1.ConditionFalse,
				Reason: ReasonBackendMissing,
			}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			srr,
		).Build()
	if err := ensureCallerNamespaceLinkerdSkipPorts(context.Background(), c, ns); err != nil {
		t.Fatalf("ensureCallerNamespaceLinkerdSkipPorts: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if nsObj.Annotations[LinkerdSkipInboundPortsAnnotation] != PureCallerInboundSkipPorts {
		t.Fatalf("skip-inbound-ports = %q", nsObj.Annotations[LinkerdSkipInboundPortsAnnotation])
	}
}

func TestIsCalleeNamespace_TC204c(t *testing.T) {
	scheme := buildCallerScheme(t)
	baseClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	failingClient := &alwaysFailListClient{
		Client: baseClient,
		err:    errors.New("api unavailable"),
	}
	got, err := isCalleeNamespace(context.Background(), failingClient, "user-space-tc204c")
	if err != nil {
		t.Fatalf("isCalleeNamespace: %v", err)
	}
	if !got {
		t.Fatalf("isCalleeNamespace fail-safe expected true")
	}
}

func TestCallerReconciler_TC205OptOutRemovesSkipAnnotations(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc205"
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
				Annotations: map[string]string{
					AnnotationLinkerdInject:                LinkerdInjectDisabled,
					LinkerdSkipInboundPortsAnnotation:      PureCallerInboundSkipPorts,
					LinkerdSkipOutboundPortsAnnotation:     "1-8080,8083-65535",
					LinkerdInjectAnnotation:                LinkerdInjectEnabled,
					gateway.AnnotationInCluster:            gateway.InClusterGateway,
					security.NamespaceInClusterCallerLabel: "true",
				},
				Labels: map[string]string{
					security.NamespaceInClusterCallerLabel: "true",
				},
			},
		},
	).Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if _, ok := nsObj.Annotations[LinkerdSkipInboundPortsAnnotation]; ok {
		t.Fatalf("skip-inbound-ports annotation must be removed")
	}
	if _, ok := nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation]; ok {
		t.Fatalf("skip-outbound-ports annotation must be removed")
	}
}

func TestCallerReconciler_TC206NoOptInNoSkipWrites(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc206"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}).
		Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if _, ok := nsObj.Annotations[LinkerdSkipInboundPortsAnnotation]; ok {
		t.Fatalf("skip-inbound-ports must not be written")
	}
	if _, ok := nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation]; ok {
		t.Fatalf("skip-outbound-ports must not be written")
	}
}

func TestCallerReconciler_TC207OperatorOptOutNoSkipWrites(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc207"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name:        ns,
			Annotations: map[string]string{AnnotationLinkerdInject: LinkerdInjectDisabled},
		}}).Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	nsObj := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if _, ok := nsObj.Annotations[LinkerdSkipInboundPortsAnnotation]; ok {
		t.Fatalf("skip-inbound-ports must not be present")
	}
	if _, ok := nsObj.Annotations[LinkerdSkipOutboundPortsAnnotation]; ok {
		t.Fatalf("skip-outbound-ports must not be present")
	}
}

func TestCallerReconciler_TC208IdempotentSkipAnnotations(t *testing.T) {
	scheme := buildCallerScheme(t)
	ns := "user-space-tc208"
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-gateway"}},
		&appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name: "litellm-idempotent",
				Annotations: map[string]string{
					gateway.AnnotationInCluster: gateway.InClusterGateway,
				},
			},
			Spec: appv1alpha1.ApplicationSpec{
				Name:      "litellm-idempotent",
				Namespace: ns,
				Settings:  map[string]string{"clusterAppRef": "ollamav2"},
			},
		},
	).Build()
	r := &CallerReconciler{Client: c, GatewayNS: "app-gateway"}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	first := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, first); err != nil {
		t.Fatalf("get first ns: %v", err)
	}
	firstRV := first.ResourceVersion
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	second := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, second); err != nil {
		t.Fatalf("get second ns: %v", err)
	}
	if second.ResourceVersion != firstRV {
		t.Fatalf("resourceVersion changed on idempotent reconcile: %q -> %q", firstRV, second.ResourceVersion)
	}
}

func TestChartRender_TC209(t *testing.T) {
	if os.Getenv("RUN_HELM_TEMPLATE_TESTS") != "1" {
		return
	}
	if _, err := exec.LookPath("helm"); err != nil {
		t.Fatalf("RUN_HELM_TEMPLATE_TESTS=1 but helm CLI absent: %v", err)
	}
	chartPath := filepath.Clean(filepath.Join("..", "..", "..", "..", "app-gateway", ".olares", "config", "user", "helm-charts", "app-gateway"))

	defaultOut := runHelmTemplateDataPlane(t, chartPath)
	mustContain(t, defaultOut, "name: incluster-strong")
	mustContain(t, defaultOut, "port: 8081")
	mustContain(t, defaultOut, "targetPort: 10080")
	mustContain(t, defaultOut, "protocol: TCP")
	mustContain(t, defaultOut, "name: incluster-http-strong")
	mustContain(t, defaultOut, "port: 8082")
	mustContain(t, defaultOut, "targetPort: 10080")

	overrideOut := runHelmTemplateDataPlane(t, chartPath, "--set", "inCluster.strongIdentityServicePort=8082")
	mustContain(t, overrideOut, "name: incluster-strong")
	mustContain(t, overrideOut, "port: 8082")
	overrideHTTPStrongOut := runHelmTemplateDataPlane(t, chartPath, "--set", "inCluster.httpStrongServicePort=18082")
	mustContain(t, overrideHTTPStrongOut, "name: incluster-http-strong")
	mustContain(t, overrideHTTPStrongOut, "port: 18082")
}

func TestDefaultsOpaquePorts_TC210(t *testing.T) {
	defaultsPath := filepath.Clean(filepath.Join("..", "..", "..", "..", "app-gateway", "config", "defaults.yaml"))
	content, err := os.ReadFile(defaultsPath)
	if err != nil {
		t.Fatalf("read defaults.yaml: %v", err)
	}
	var opaque string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "opaquePorts:") {
			opaque = strings.TrimSpace(strings.TrimPrefix(trimmed, "opaquePorts:"))
			opaque = strings.Trim(opaque, "\"")
			break
		}
	}
	if opaque == "" {
		t.Fatalf("opaquePorts not found in defaults.yaml")
	}
	expanded, err := expandPortSet(opaque)
	if err != nil {
		t.Fatalf("expand opaquePorts: %v", err)
	}
	if expanded[8081] {
		t.Fatalf("opaquePorts must not include 8081, got %q", opaque)
	}
	if expanded[8082] {
		t.Fatalf("opaquePorts must not include 8082, got %q", opaque)
	}
}

func portInSkipRange(port int, skip string) bool {
	expanded, err := expandPortSet(skip)
	if err != nil {
		return false
	}
	return expanded[port]
}

type alwaysFailListClient struct {
	client.Client
	err error
}

func (c *alwaysFailListClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.err
}

func buildCallerScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return scheme
}

func containsExact(tokens []string, target string) bool {
	for _, token := range tokens {
		if strings.TrimSpace(token) == target {
			return true
		}
	}
	return false
}

func runHelmTemplateDataPlane(t *testing.T, chartPath string, extraArgs ...string) string {
	t.Helper()
	args := []string{
		"template", "app-gateway", chartPath,
		"--show-only", "templates/data-plane-svc.yaml",
	}
	args = append(args, extraArgs...)
	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, string(out))
	}
	return string(out)
}

func mustContain(t *testing.T, output, needle string) {
	t.Helper()
	if !strings.Contains(output, needle) {
		t.Fatalf("helm output missing %q\n%s", needle, output)
	}
}

func expandPortSet(raw string) (map[int]bool, error) {
	result := make(map[int]bool)
	for _, token := range strings.Split(raw, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.Contains(token, "-") {
			parts := strings.SplitN(token, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}
			for p := start; p <= end; p++ {
				result[p] = true
			}
			continue
		}
		port, err := strconv.Atoi(token)
		if err != nil {
			return nil, err
		}
		result[port] = true
	}
	return result, nil
}

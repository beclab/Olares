package routecontrol

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func testSchemeWithSecurityPolicy(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := testScheme(t)
	eg := schema.GroupVersion{Group: "gateway.envoyproxy.io", Version: "v1alpha1"}
	s.AddKnownTypeWithName(eg.WithKind("SecurityPolicy"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(eg.WithKind("SecurityPolicyList"), &unstructured.UnstructuredList{})
	rg := schema.GroupVersion{Group: "gateway.networking.k8s.io", Version: "v1beta1"}
	s.AddKnownTypeWithName(rg.WithKind("ReferenceGrant"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(rg.WithKind("ReferenceGrantList"), &unstructured.UnstructuredList{})
	return s
}

func TestDesiredSharedRouteSecurityPolicyJWTAuthn(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-shared", Namespace: "demo-shared"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			HostPatterns: []string{"*.shared.olares.com"},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "api"},
		},
	}
	spec := SecurityPolicySpecForTest(srr)
	meta := SecurityPolicyObjectMetaForTest(srr)

	if meta.Name != "demo-shared-jwt-authn" {
		t.Fatalf("name = %q, want demo-shared-jwt-authn", meta.Name)
	}
	if meta.Namespace != "demo-shared" {
		t.Fatalf("namespace = %q, want demo-shared", meta.Namespace)
	}

	targetRef, ok := spec["targetRef"].(map[string]any)
	if !ok {
		t.Fatalf("targetRef type = %T", spec["targetRef"])
	}
	if targetRef["kind"] != "HTTPRoute" || targetRef["name"] != "demo-shared" {
		t.Fatalf("targetRef = %#v", targetRef)
	}

	jwt, ok := spec["jwt"].(map[string]any)
	if !ok {
		t.Fatalf("jwt type = %T", spec["jwt"])
	}
	providers, ok := jwt["providers"].([]any)
	if !ok || len(providers) != 1 {
		t.Fatalf("providers = %#v", jwt["providers"])
	}
	provider, ok := providers[0].(map[string]any)
	if !ok {
		t.Fatalf("provider type = %T", providers[0])
	}
	if provider["issuer"] != CallerJWTIssuer {
		t.Fatalf("issuer = %v", provider["issuer"])
	}
	audiences, ok := provider["audiences"].([]any)
	if !ok || len(audiences) != 1 || audiences[0] != CallerJWTAudience {
		t.Fatalf("audiences = %#v", provider["audiences"])
	}

	claimToHeaders, ok := provider["claimToHeaders"].([]any)
	if !ok || len(claimToHeaders) != 3 {
		t.Fatalf("claimToHeaders = %#v", provider["claimToHeaders"])
	}
	claimMap := claimToHeaders[0].(map[string]any)
	if claimMap["claim"] != CallerJWTViewerClaim || claimMap["header"] != CallerJWTViewerHeader {
		t.Fatalf("claimToHeaders[0] = %#v", claimMap)
	}
	// Spoofed client X-BFL-USER is overwritten after JWT authn; keep this mapping.
	if claimMap["header"] != "X-BFL-USER" {
		t.Fatalf("viewer header must remain X-BFL-USER for gateway overwrite, got %#v", claimMap["header"])
	}
	appidMap := claimToHeaders[1].(map[string]any)
	if appidMap["claim"] != CallerJWTAppidClaim || appidMap["header"] != CallerJWTAppidHeader {
		t.Fatalf("claimToHeaders[1] = %#v", appidMap)
	}
	if appidMap["header"] != "X-Shared-Appid" {
		t.Fatalf("shared appid header must remain X-Shared-Appid, got %#v", appidMap["header"])
	}
	clientMap := claimToHeaders[2].(map[string]any)
	if clientMap["claim"] != CallerJWTClientAppidClaim || clientMap["header"] != CallerJWTClientAppidHeader {
		t.Fatalf("claimToHeaders[2] = %#v", clientMap)
	}
	if clientMap["header"] != "X-Caller-Appid" {
		t.Fatalf("ordinary client appid header must remain X-Caller-Appid, got %#v", clientMap["header"])
	}
	// Three mappings stay configured; a given token carries the claims for its caller type.
	if CallerJWTViewerClaim != callerjwt.ClaimViewer ||
		CallerJWTAppidClaim != callerjwt.ClaimAppid ||
		CallerJWTClientAppidClaim != callerjwt.ClaimClientAppid {
		t.Fatalf("SecurityPolicy claim paths must alias callerjwt nested claim constants")
	}

	extractFrom, ok := provider["extractFrom"].(map[string]any)
	if !ok {
		t.Fatalf("extractFrom type = %T", provider["extractFrom"])
	}
	headers, ok := extractFrom["headers"].([]any)
	if !ok || len(headers) != 1 {
		t.Fatalf("extractFrom.headers = %#v", extractFrom["headers"])
	}
	hdr := headers[0].(map[string]any)
	if hdr["name"] != callerjwt.CallerJWTHeaderName {
		t.Fatalf("extractFrom header = %#v, want name %q", hdr, callerjwt.CallerJWTHeaderName)
	}
	if _, hasPrefix := hdr["valuePrefix"]; hasPrefix {
		t.Fatalf("caller JWT header must be bare JWT without valuePrefix, got %#v", hdr)
	}

	remoteJWKS, ok := provider["remoteJWKS"].(map[string]any)
	if !ok {
		t.Fatalf("remoteJWKS type = %T", provider["remoteJWKS"])
	}
	if remoteJWKS["uri"] != CallerJWTJWKSURI {
		t.Fatalf("remoteJWKS.uri = %v", remoteJWKS["uri"])
	}
	backendRefs, ok := remoteJWKS["backendRefs"].([]any)
	if !ok || len(backendRefs) != 1 {
		t.Fatalf("remoteJWKS.backendRefs = %#v", remoteJWKS["backendRefs"])
	}
	backend := backendRefs[0].(map[string]any)
	if backend["name"] != CallerJWTJWKSServiceName || backend["namespace"] != CallerJWTJWKSServiceNamespace {
		t.Fatalf("backendRef = %#v", backend)
	}
}

func TestIssuedJWTSatisfiesClaimToHeaderPaths(t *testing.T) {
	ring, err := callerjwt.NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := callerjwt.NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	cases := []struct {
		name   string
		req    callerjwt.IssueRequest
		header string
		claim  string
		want   string
	}{
		{
			name:   "X-Shared-Appid",
			req:    callerjwt.IssueRequest{Namespace: "router-shared", ServiceAccountName: "router", AppRef: "router", Appid: "f3395cd5"},
			header: CallerJWTAppidHeader,
			claim:  CallerJWTAppidClaim,
			want:   "f3395cd5",
		},
		{
			name:   "X-BFL-USER",
			req:    callerjwt.IssueRequest{Namespace: "ns", ServiceAccountName: "sa", AppRef: "demo", Viewer: "alice", ClientAppid: "6bf98da2"},
			header: CallerJWTViewerHeader,
			claim:  CallerJWTViewerClaim,
			want:   "alice",
		},
		{
			name:   "X-Caller-Appid",
			req:    callerjwt.IssueRequest{Namespace: "ns", ServiceAccountName: "sa", AppRef: "demo", Viewer: "alice", ClientAppid: "6bf98da2"},
			header: CallerJWTClientAppidHeader,
			claim:  CallerJWTClientAppidClaim,
			want:   "6bf98da2",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := issuer.Issue(tc.req)
			if err != nil {
				t.Fatalf("Issue: %v", err)
			}
			got := walkJWTClaimPath(t, token, tc.claim)
			if got != tc.want {
				t.Fatalf("header %s claim %s = %q, want %q", tc.header, tc.claim, got, tc.want)
			}
		})
	}
}

func walkJWTClaimPath(t *testing.T, token, claimPath string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		t.Fatalf("token parts = %d", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cur := any(payload)
	for _, part := range strings.Split(claimPath, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("path %q: not object before %q", claimPath, part)
		}
		next, ok := obj[part]
		if !ok {
			t.Fatalf("path %q: missing %q", claimPath, part)
		}
		cur = next
	}
	got, ok := cur.(string)
	if !ok {
		t.Fatalf("path %q: got %#v", claimPath, cur)
	}
	return got
}

func TestReconcileSharedRouteGatewayModeCreatesSecurityPolicy(t *testing.T) {
	ctx := context.Background()
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-shared", Namespace: "demo-shared"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{"*.shared.olares.com"},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "api", Port: 80},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "demo-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolTCP}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(srr, svc).Build()

	res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != ReasonReconciled {
		t.Fatalf("reason = %q, want %q", res.Reason, ReasonReconciled)
	}

	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(securityPolicyGVK)
	if err := c.Get(ctx, types.NamespacedName{Namespace: "demo-shared", Name: "demo-shared-jwt-authn"}, got); err != nil {
		t.Fatalf("get SecurityPolicy: %v", err)
	}

	grant := &unstructured.Unstructured{}
	grant.SetGroupVersionKind(referenceGrantGVK)
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: CallerJWTJWKSServiceNamespace,
		Name:      jwksReferenceGrantName(srr),
	}, grant); err != nil {
		t.Fatalf("get JWKS ReferenceGrant: %v", err)
	}
	from, found, err := unstructured.NestedSlice(grant.Object, "spec", "from")
	if err != nil || !found || len(from) != 1 {
		t.Fatalf("spec.from = %#v err=%v", from, err)
	}
	from0, ok := from[0].(map[string]any)
	if !ok || from0["kind"] != "SecurityPolicy" || from0["namespace"] != "demo-shared" {
		t.Fatalf("spec.from[0] = %#v", from0)
	}
	to, found, err := unstructured.NestedSlice(grant.Object, "spec", "to")
	if err != nil || !found || len(to) != 1 {
		t.Fatalf("spec.to = %#v err=%v", to, err)
	}
	to0, ok := to[0].(map[string]any)
	if !ok || to0["kind"] != "Service" || to0["name"] != CallerJWTJWKSServiceName {
		t.Fatalf("spec.to[0] = %#v", to0)
	}
}

func TestReconcileSharedRouteDirectModeDeletesSecurityPolicy(t *testing.T) {
	ctx := context.Background()
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-shared", Namespace: "demo-shared"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeDirect,
			HostPatterns: []string{"*.shared.olares.com"},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "api"},
		},
	}
	policy := desiredSharedRouteSecurityPolicy(srr)
	grant := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": referenceGrantAPIVersion,
		"kind":       "ReferenceGrant",
		"metadata": map[string]any{
			"name":      jwksReferenceGrantName(srr),
			"namespace": CallerJWTJWKSServiceNamespace,
		},
	}}
	grant.SetGroupVersionKind(referenceGrantGVK)
	c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(srr, policy, grant).Build()

	res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != ReasonDirectMode {
		t.Fatalf("reason = %q, want %q", res.Reason, ReasonDirectMode)
	}

	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(securityPolicyGVK)
	if err := c.Get(ctx, types.NamespacedName{Namespace: "demo-shared", Name: "demo-shared-jwt-authn"}, got); err == nil {
		t.Fatal("expected SecurityPolicy to be deleted")
	}
	gotGrant := &unstructured.Unstructured{}
	gotGrant.SetGroupVersionKind(referenceGrantGVK)
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: CallerJWTJWKSServiceNamespace,
		Name:      jwksReferenceGrantName(srr),
	}, gotGrant); err == nil {
		t.Fatal("expected JWKS ReferenceGrant to be deleted")
	}
}

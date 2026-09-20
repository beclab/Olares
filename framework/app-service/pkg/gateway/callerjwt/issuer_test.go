package callerjwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add application scheme: %v", err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}
	return s
}

func TestIssueCallerJWTClaimsAndVerify(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	token, err := issuer.Issue(IssueRequest{
		Namespace:          "user-space-alice-demo",
		ServiceAccountName: "demo",
		AppRef:             "demo",
		Entrance:           "web",
		Viewer:             "alice",
		ClientAppid:        "6bf98da2",
		TTL:                30 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	claims, err := issuer.ParseClaims(token)
	if err != nil {
		t.Fatalf("ParseClaims: %v", err)
	}
	if claims.Issuer != IssuerURL {
		t.Fatalf("iss = %q, want %q", claims.Issuer, IssuerURL)
	}
	wantSub := SPIFFESubject("user-space-alice-demo", "demo")
	if claims.Subject != wantSub {
		t.Fatalf("sub = %q, want %q", claims.Subject, wantSub)
	}
	if !audienceContains(claims.Audience, Audience) {
		t.Fatalf("aud = %v, want %q", claims.Audience, Audience)
	}
	if claims.AppRef() != "demo" {
		t.Fatalf("appRef = %q", claims.AppRef())
	}
	if claims.Entrance() != "web" {
		t.Fatalf("entrance = %q", claims.Entrance())
	}
	if claims.Viewer() != "alice" {
		t.Fatalf("viewer = %q", claims.Viewer())
	}
	if claims.Appid() != "" {
		t.Fatalf("ordinary caller must omit shared appid, got %q", claims.Appid())
	}
	if claims.ClientAppid() != "6bf98da2" {
		t.Fatalf("clientAppid = %q", claims.ClientAppid())
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatalf("exp missing or in the past")
	}
}

func TestIssueRejectsViewerAndAppidTogether(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	_, err = issuer.Issue(IssueRequest{
		Namespace:          "ns",
		ServiceAccountName: "sa",
		AppRef:             "demo",
		Viewer:             "alice",
		Appid:              "deadbeef",
	})
	if err == nil {
		t.Fatal("expected error when viewer and shared appid are both set")
	}
}

func TestIssueRejectsAppidAndClientAppidTogether(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	_, err = issuer.Issue(IssueRequest{
		Namespace:          "ns",
		ServiceAccountName: "sa",
		AppRef:             "demo",
		Appid:              "deadbeef",
		ClientAppid:        "6bf98da2",
	})
	if err == nil {
		t.Fatal("expected error when shared appid and clientAppid are both set")
	}
}

func TestIssueSharedCallerAppidOmitsViewer(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	token, err := issuer.Issue(IssueRequest{
		Namespace:          "shared-agg",
		ServiceAccountName: "agg",
		AppRef:             "agg",
		Appid:              "a1b2c3d4",
	})
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := issuer.ParseClaims(token)
	if err != nil {
		t.Fatalf("ParseClaims: %v", err)
	}
	if claims.Appid() != "a1b2c3d4" {
		t.Fatalf("appid = %q", claims.Appid())
	}
	if claims.Viewer() != "" {
		t.Fatalf("shared caller must omit viewer, got %q", claims.Viewer())
	}
	if claims.ClientAppid() != "" {
		t.Fatalf("shared caller must omit clientAppid, got %q", claims.ClientAppid())
	}
}

func TestIssuePayloadNestedForEnvoyClaimPaths(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	sharedToken, err := issuer.Issue(IssueRequest{
		Namespace:          "router-shared",
		ServiceAccountName: "router",
		AppRef:             "router",
		Appid:              "f3395cd5",
	})
	if err != nil {
		t.Fatalf("Issue shared: %v", err)
	}
	assertEnvoyClaimPath(t, sharedToken, ClaimAppid, "f3395cd5")
	assertEnvoyClaimPathMissing(t, sharedToken, ClaimViewer)
	assertEnvoyClaimPathMissing(t, sharedToken, ClaimClientAppid)
	assertNoFlatDottedIdentityKeys(t, sharedToken)

	ordinaryToken, err := issuer.Issue(IssueRequest{
		Namespace:          "user-space-alice-demo",
		ServiceAccountName: "demo",
		AppRef:             "demo",
		Entrance:           "web",
		Viewer:             "alice",
		ClientAppid:        "6bf98da2",
	})
	if err != nil {
		t.Fatalf("Issue ordinary: %v", err)
	}
	// All three post-authn identity headers' claim paths must resolve for the
	// claims that token type carries (viewer + clientAppid; not shared appid).
	assertEnvoyClaimPath(t, ordinaryToken, ClaimViewer, "alice")
	assertEnvoyClaimPath(t, ordinaryToken, ClaimClientAppid, "6bf98da2")
	assertEnvoyClaimPathMissing(t, ordinaryToken, ClaimAppid)
	assertEnvoyClaimPath(t, ordinaryToken, ClaimEntrance, "web")
	assertEnvoyClaimPath(t, ordinaryToken, ClaimAppRef, "demo")
	assertNoFlatDottedIdentityKeys(t, ordinaryToken)
}

func decodeJWTPayload(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		t.Fatalf("token parts = %d", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return payload
}

func assertNoFlatDottedIdentityKeys(t *testing.T, token string) {
	t.Helper()
	payload := decodeJWTPayload(t, token)
	for _, k := range []string{
		ClaimAppid, ClaimClientAppid, ClaimViewer, ClaimAppRef, ClaimEntrance,
	} {
		if _, ok := payload[k]; ok {
			t.Fatalf("payload must not use flat key %q", k)
		}
	}
}

func assertEnvoyClaimPath(t *testing.T, token, claimPath, want string) {
	t.Helper()
	payload := decodeJWTPayload(t, token)
	cur := any(payload)
	for _, part := range strings.Split(claimPath, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("claim path %q: not an object before %q", claimPath, part)
		}
		next, ok := obj[part]
		if !ok {
			t.Fatalf("claim path %q: missing %q in %#v", claimPath, part, obj)
		}
		cur = next
	}
	got, ok := cur.(string)
	if !ok {
		t.Fatalf("claim path %q: got %#v, want string %q", claimPath, cur, want)
	}
	if got != want {
		t.Fatalf("claim path %q = %q, want %q", claimPath, got, want)
	}
}

func assertEnvoyClaimPathMissing(t *testing.T, token, claimPath string) {
	t.Helper()
	payload := decodeJWTPayload(t, token)
	cur := any(payload)
	parts := strings.Split(claimPath, ".")
	for i, part := range parts {
		obj, ok := cur.(map[string]any)
		if !ok {
			return
		}
		next, ok := obj[part]
		if !ok {
			return
		}
		if i == len(parts)-1 {
			t.Fatalf("claim path %q must be absent for this token type, got %#v", claimPath, next)
		}
		cur = next
	}
}

func TestShouldIssueCallerJWTSharedAppRefAndOptOut(t *testing.T) {
	sharedLabels := map[string]string{"app.bytetrade.io/app-shared": "true"}
	appRefOnly := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "agg", Labels: sharedLabels},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "agg",
			Appid:     "deadbeef",
			Namespace: "shared-agg",
			Settings:  map[string]string{settingAppRef: "ollama"},
		},
	}
	if !shouldIssueCallerJWT(appRefOnly) {
		t.Fatal("shared app with only appRef must issue JWT")
	}
	optOut := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "agg", Labels: sharedLabels},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "agg",
			Appid:     "deadbeef",
			Namespace: "shared-agg",
			Settings: map[string]string{
				settingSharedAppDeps: "ollama",
				settingOptOutMesh:    "disabled",
			},
		},
	}
	if !shouldIssueCallerJWT(optOut) {
		t.Fatal("shared app always issues JWT even with mesh-inject opt-out")
	}
	intentOnly := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "agg", Labels: sharedLabels},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "agg",
			Appid:     "deadbeef",
			Namespace: "shared-agg",
			Settings:  map[string]string{settingNeedsSharedAccess: "true"},
		},
	}
	if !shouldIssueCallerJWT(intentOnly) {
		t.Fatal("shared app always issues JWT even with needsSharedAccess alone")
	}
	pureCallee := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "engine", Labels: sharedLabels},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "engine",
			Appid:     "82f26852",
			Namespace: "engine-shared",
			Settings:  map[string]string{settingSharedCallerDecide: "false"},
		},
	}
	if !shouldIssueCallerJWT(pureCallee) {
		t.Fatal("shared pure callee (decide=false) must still issue JWT")
	}
}

func TestBuildJWKSIncludesRotationKey(t *testing.T) {
	ring, err := NewKeyRingForTest(true)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	set, err := BuildJWKS(ring)
	if err != nil {
		t.Fatalf("BuildJWKS: %v", err)
	}
	if len(set.Keys) != 2 {
		t.Fatalf("jwks keys = %d, want 2", len(set.Keys))
	}
	data, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	if _, err := VerifyJWKSResponse(data); err != nil {
		t.Fatalf("VerifyJWKSResponse: %v", err)
	}
}

func TestJWKSHandlerReturns200(t *testing.T) {
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, JWKSPath, nil)
	rec := httptest.NewRecorder()
	JWKSHandler(issuer).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if _, err := VerifyJWKSResponse(rec.Body.Bytes()); err != nil {
		t.Fatalf("VerifyJWKSResponse: %v", err)
	}
}

func TestIssuerReconcilerCreatesJWTSecretForCaller(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "user-space-alice",
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Appid:     "6bf98da2",
			Namespace: "user-space-alice-demo",
			Owner:     "alice",
			Settings: map[string]string{
				settingNeedsSharedAccess: "true",
				settingSharedAppDeps:     "web",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err != nil {
		t.Fatalf("reconcileApplication: %v", err)
	}

	secret := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: app.Spec.Namespace,
		Name:      AppJWTSecretName,
	}, secret); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	token := string(secret.Data[AppJWTSecretDataKey])
	if token == "" {
		t.Fatalf("secret token is empty")
	}
	claims, err := issuer.ParseClaims(token)
	if err != nil {
		t.Fatalf("ParseClaims: %v", err)
	}
	if claims.AppRef() != "demo" {
		t.Fatalf("appRef = %q", claims.AppRef())
	}
	if claims.Viewer() != "alice" {
		t.Fatalf("viewer = %q", claims.Viewer())
	}
	if claims.ClientAppid() != "6bf98da2" {
		t.Fatalf("clientAppid = %q", claims.ClientAppid())
	}
	if claims.Appid() != "" {
		t.Fatalf("ordinary caller must omit shared appid, got %q", claims.Appid())
	}
}

func TestIssuerReconcilerCreatesJWTSecretWhenDecideTrue(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "litellm",
			Namespace: "user-space-alice",
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "litellm",
			Appid:     "6aead52a",
			Namespace: "litellm-alice",
			Owner:     "alice",
			Settings: map[string]string{
				settingSharedCallerDecide: "true",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err != nil {
		t.Fatalf("reconcileApplication: %v", err)
	}

	secret := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: app.Spec.Namespace,
		Name:      AppJWTSecretName,
	}, secret); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if len(secret.Data[AppJWTSecretDataKey]) == 0 {
		t.Fatalf("secret token is empty")
	}
}

func TestIssuerReconcileRequeuesForJWTRefresh(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	keys := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      IssuerKeysSecretName,
			Namespace: JWKSServiceNamespace,
		},
		Data: map[string][]byte{
			SigningKeyPEM:   encodePrivateKeyPEM(ring.Active),
			SigningKeyIDKey: []byte(ring.Active.KID),
		},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "user-space-alice"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Appid:     "6bf98da2",
			Namespace: "user-space-alice-demo",
			Owner:     "alice",
			Settings: map[string]string{
				settingSharedAppDeps: "web",
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(keys, app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	res, err := r.Reconcile(context.Background(), reconcileRequest(app))
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if res.RequeueAfter != JWTRefreshInterval {
		t.Fatalf("RequeueAfter = %v, want %v", res.RequeueAfter, JWTRefreshInterval)
	}
}

func reconcileRequest(app *appv1alpha1.Application) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: app.Name, Namespace: app.Namespace}}
}

func TestIssueRequestFromApplicationOmitsCalleeAsEntrance(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "user-space-alice"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Appid:     "6bf98da2",
			Namespace: "user-space-alice-demo",
			Owner:     "alice",
			Settings: map[string]string{
				settingSharedAppDeps: "shared-llm,shared-rag",
				"serviceAccountName": "demo-sa",
			},
		},
	}
	req := issueRequestFromApplication(app)
	if req.Entrance != "" {
		t.Fatalf("Entrance = %q, want empty (sharedAppDeps must not map to olares.entrance)", req.Entrance)
	}
	if req.AppRef != "demo" || req.Viewer != "alice" || req.ServiceAccountName != "demo-sa" {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.Appid != "" {
		t.Fatalf("ordinary caller Appid = %q, want empty", req.Appid)
	}
	if req.ClientAppid != "6bf98da2" {
		t.Fatalf("ClientAppid = %q", req.ClientAppid)
	}
	if req.Namespace != "user-space-alice-demo" {
		t.Fatalf("Namespace = %q", req.Namespace)
	}
}

func TestIssueRequestFromSharedApplicationUsesAppid(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agg",
			Namespace: "user-system-alice",
			Labels:    map[string]string{"app.bytetrade.io/app-shared": "true"},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "agg",
			Appid:     "deadbeef",
			Namespace: "shared-agg",
			Owner:     "alice",
			Settings: map[string]string{
				settingSharedAppDeps: "ollama",
			},
		},
	}
	req := issueRequestFromApplication(app)
	if req.Appid != "deadbeef" {
		t.Fatalf("Appid = %q", req.Appid)
	}
	if req.Viewer != "" {
		t.Fatalf("Viewer = %q, want empty for shared caller", req.Viewer)
	}
	if req.ClientAppid != "" {
		t.Fatalf("ClientAppid = %q, want empty for shared caller", req.ClientAppid)
	}
}

func TestIssuerReconcilerOrdinaryMissingAppidErrors(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "user-space-alice"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "user-space-alice-demo",
			Owner:     "alice",
			Settings: map[string]string{
				settingSharedCallerDecide: "true",
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err == nil {
		t.Fatal("expected error when ordinary caller has empty spec.appid")
	}
}

func TestIssuerReconcilerIssuesJWTForSharedPureCallee(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "engine",
			Namespace: "user-system",
			Labels:    map[string]string{"app.bytetrade.io/app-shared": "true"},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "engine",
			Appid:     "82f26852",
			Namespace: "engine-shared",
			Settings:  map[string]string{settingSharedCallerDecide: "false"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err != nil {
		t.Fatalf("reconcileApplication: %v", err)
	}
	secret := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: app.Spec.Namespace,
		Name:      AppJWTSecretName,
	}, secret); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if len(secret.Data[AppJWTSecretDataKey]) == 0 {
		t.Fatal("expected non-empty caller-jwt token")
	}
}

func TestIssuerReconcilerSharedMissingAppidErrors(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "engine",
			Namespace: "user-system",
			Labels:    map[string]string{"app.bytetrade.io/app-shared": "true"},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "engine",
			Namespace: "engine-shared",
			Settings:  map[string]string{settingSharedCallerDecide: "false"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err == nil {
		t.Fatal("expected error when shared app has empty spec.appid")
	}
}

func TestIssuerReconcilerDeletesSecretWithoutDependency(t *testing.T) {
	scheme := testScheme(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "user-space-alice",
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "user-space-alice-demo",
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppJWTSecretName,
			Namespace: app.Spec.Namespace,
		},
		Data: map[string][]byte{AppJWTSecretDataKey: []byte("stale")},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, secret).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileApplication(context.Background(), app); err != nil {
		t.Fatalf("reconcileApplication: %v", err)
	}
	err = c.Get(context.Background(), types.NamespacedName{
		Namespace: app.Spec.Namespace,
		Name:      AppJWTSecretName,
	}, &corev1.Secret{})
	if err == nil {
		t.Fatalf("expected secret to be deleted")
	}
}

package authz

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestParseL5dClientID(t *testing.T) {
	const suffix = ".serviceaccount.identity.linkerd.cluster.local"
	valid := "default.user-space-alice" + suffix
	sa, ns, err := ParseL5dClientID(valid)
	if err != nil || sa != "default" || ns != "user-space-alice" {
		t.Fatalf("valid: sa=%q ns=%q err=%v", sa, ns, err)
	}
	if _, _, err := ParseL5dClientID(""); err == nil {
		t.Fatal("empty expected error")
	}
	if _, _, err := ParseL5dClientID("not-spiffe"); err == nil {
		t.Fatal("malformed expected error")
	}
	if _, _, err := ParseL5dClientID("onlyns" + suffix); err == nil {
		t.Fatal("short core expected error")
	}
}

func TestDeriveViewer(t *testing.T) {
	cases := []struct {
		ns   string
		want string
		ok   bool
	}{
		{"user-space-alice", "alice", true},
		{"USER-SPACE-BOB", "bob", true},
		{"user-system-svc", "svc", true},
		{"kube-system", "", false},
		{"linkerd", "", false},
	}
	for _, tc := range cases {
		got, ok := DeriveViewer(tc.ns)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("DeriveViewer(%q) = %q,%v want %q,%v", tc.ns, got, ok, tc.want, tc.ok)
		}
	}
}

func TestIsSharedInclusterHost(t *testing.T) {
	if !IsSharedInclusterHost("a1b2c3d4.alice.olares.com") {
		t.Fatal("expected shared host")
	}
	if IsSharedInclusterHost("a1b2c3d.alice.olares.com") {
		t.Fatal("7-char prefix must not match")
	}
	if IsSharedInclusterHost("notshared.example.com") {
		t.Fatal("non-hash8 must not match")
	}
	if IsSharedInclusterHost("") {
		t.Fatal("empty must not match")
	}
}

func TestHostViewerLabel(t *testing.T) {
	if got := HostViewerLabel("a1b2c3d4.alice.olares.com:443"); got != "alice" {
		t.Fatalf("viewer = %q", got)
	}
	if got := HostViewerLabel("short"); got != "" {
		t.Fatalf("short host viewer = %q", got)
	}
}

// newAuthzFakeClient builds a controller-runtime fake client preloaded with
// the given Namespaces. Shared by TC-031/034/035/036 to exercise the
// loadKnownUsers helper before invoking the pure DeriveViewerWithMeta function.
func newAuthzFakeClient(t *testing.T, namespaces ...*corev1.Namespace) ctrlclient.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("client-go scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("corev1 scheme: %v", err)
	}
	objs := make([]ctrlclient.Object, 0, len(namespaces))
	for _, ns := range namespaces {
		objs = append(objs, ns)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

// TC-031 verifies DeriveViewerWithMeta path 1 (ns_label) wins when the caller
// namespace carries bytetrade.io/ns-owner. The fake client also seeds the
// label-derivation branch of loadKnownUsers so we additionally assert the
// helper accepts ns-owner as a known-user signal.
func TestDeriveViewerWithMeta_NsLabel_TC031(t *testing.T) {
	cli := newAuthzFakeClient(t, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "litellm-brucedai",
			Labels: map[string]string{nsOwnerLabel: "brucedai"},
		},
	})
	h := &authzHandler{audit: discardAuditor(), k8sClient: cli}
	known := h.loadKnownUsers(context.Background())
	if _, ok := known["brucedai"]; !ok {
		t.Fatalf("TC-031 loadKnownUsers should include brucedai via ns-owner label, got %v", known)
	}
	viewer, source, ok := DeriveViewerWithMeta(
		"litellm-brucedai",
		map[string]string{nsOwnerLabel: "brucedai"},
		known,
	)
	if !ok || viewer != "brucedai" || source != SourceNsLabel {
		t.Fatalf("TC-031: viewer=%q source=%q ok=%v want brucedai/ns_label/true", viewer, source, ok)
	}
}

// TC-032 verifies DeriveViewerWithMeta keeps DeriveViewer-equivalent semantics
// on the user-space- prefix path (no label, no knownUsers required).
func TestDeriveViewerWithMeta_PrefixUserSpace_TC032(t *testing.T) {
	viewer, source, ok := DeriveViewerWithMeta("user-space-alice", nil, nil)
	if !ok || viewer != "alice" || source != SourcePrefixUserSpace {
		t.Fatalf("TC-032: viewer=%q source=%q ok=%v want alice/prefix_user_space/true", viewer, source, ok)
	}
}

// TC-033 verifies DeriveViewerWithMeta keeps DeriveViewer-equivalent semantics
// on the user-system- prefix path.
func TestDeriveViewerWithMeta_PrefixUserSystem_TC033(t *testing.T) {
	viewer, source, ok := DeriveViewerWithMeta("user-system-bob", nil, nil)
	if !ok || viewer != "bob" || source != SourcePrefixUserSystem {
		t.Fatalf("TC-033: viewer=%q source=%q ok=%v want bob/prefix_user_system/true", viewer, source, ok)
	}
}

// TC-034 closes G-B: loadKnownUsers derives {brucedai} from a user-space-*
// namespace and DeriveViewerWithMeta then resolves the <app>-<user> caller
// namespace (litellm-brucedai) via path 4 app_user_fallback.
func TestDeriveViewerWithMeta_AppUserFallback_TC034(t *testing.T) {
	cli := newAuthzFakeClient(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-brucedai"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "litellm-brucedai"}},
	)
	h := &authzHandler{audit: discardAuditor(), k8sClient: cli}
	known := h.loadKnownUsers(context.Background())
	if _, ok := known["brucedai"]; !ok {
		t.Fatalf("TC-034 loadKnownUsers should include brucedai via user-space-* prefix, got %v", known)
	}
	viewer, source, ok := DeriveViewerWithMeta("litellm-brucedai", nil, known)
	if !ok || viewer != "brucedai" || source != SourceAppUserFallback {
		t.Fatalf("TC-034: viewer=%q source=%q ok=%v want brucedai/app_user_fallback/true", viewer, source, ok)
	}
}

// TC-035 verifies that a <app>-<user> caller namespace whose trailing segment
// is not a known user falls through to source=none, and that InClusterIdentity
// then denies the request with INVALID_CALLER_IDENTITY.
func TestDeriveViewerWithMeta_AppUserUnknown_Deny_TC035(t *testing.T) {
	cli := newAuthzFakeClient(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-brucedai"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "litellm-unknown"}},
	)
	h := &authzHandler{audit: discardAuditor(), k8sClient: cli}
	known := h.loadKnownUsers(context.Background())
	viewer, source, ok := DeriveViewerWithMeta("litellm-unknown", nil, known)
	if ok || source != SourceNone {
		t.Fatalf("TC-035 DeriveViewerWithMeta: viewer=%q source=%q ok=%v want none/false", viewer, source, ok)
	}
	dec := InClusterIdentity(
		"a1b2c3d4.brucedai.olares.com",
		map[string]string{"l5d-client-id": "default.litellm-unknown.serviceaccount.identity.linkerd.cluster.local"},
		known,
	)
	if dec.Action != ActionDeny || dec.Code != CodeInvalidCallerIdentity {
		t.Fatalf("TC-035 InClusterIdentity: %+v want Deny INVALID_CALLER_IDENTITY", dec)
	}
}

// TC-036 verifies platform namespaces (kube-system) are not resolvable to a
// viewer. The natural fall-through aligns with the routecontrol
// isMeshMandatoryCallerNamespace exclusion set without requiring an explicit
// blocklist coupling.
func TestDeriveViewerWithMeta_PlatformNS_TC036(t *testing.T) {
	cli := newAuthzFakeClient(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-brucedai"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
	)
	h := &authzHandler{audit: discardAuditor(), k8sClient: cli}
	known := h.loadKnownUsers(context.Background())
	viewer, source, ok := DeriveViewerWithMeta("kube-system", nil, known)
	if ok || source != SourceNone {
		t.Fatalf("TC-036: viewer=%q source=%q ok=%v want none/false", viewer, source, ok)
	}
}

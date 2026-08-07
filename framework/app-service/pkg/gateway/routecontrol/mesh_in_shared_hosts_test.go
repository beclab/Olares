package routecontrol

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestMeshInSharedHostsReconcileCreatesCM(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "litellm-alice",
			Labels: map[string]string{security.NamespaceInClusterCallerLabel: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	r := &MeshInSharedHostsReconciler{Client: c, platformDomain: "olares.com"}
	targets := []SharedHostsTarget{{
		CallerNamespace: "litellm-alice",
		Viewer:          "alice",
		Hosts:           []string{"abcd1234.alice.olares.com"},
	}}
	if err := r.ReconcileNamespace(context.Background(), "litellm-alice", targets); err != nil {
		t.Fatal(err)
	}
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "litellm-alice", Name: constants.MeshInSharedHostsCMName,
	}, cm); err != nil {
		t.Fatal(err)
	}
	body := cm.Data[constants.MeshInSharedHostsFileName]
	if !strings.Contains(body, "abcd1234.alice.olares.com") {
		t.Fatalf("hosts body = %q", body)
	}
	if cm.Labels[constants.MeshInSharedHostsManagedByLabel] != sharedHostsManagedByValue {
		t.Fatalf("managed-by label = %q", cm.Labels[constants.MeshInSharedHostsManagedByLabel])
	}
}

func TestMaterializeHostLogicalPattern(t *testing.T) {
	h, reason := materializeHost("abcd1234.*.olares.com", "alice", "olares.com")
	if reason != "" || h != "abcd1234.alice.olares.com" {
		t.Fatalf("got host=%q reason=%q", h, reason)
	}
	h, reason = materializeHost("x.shared.olares.com", "alice", "olares.com")
	if reason != "" || h != "x.shared.olares.com" {
		t.Fatalf("shared exact host=%q reason=%q, want kept", h, reason)
	}
}

func TestEnumerateHostsSplitsAuthAndTLSByEntranceClass(t *testing.T) {
	srrs := []srrv1alpha1.SharedRouteRegistry{
		{
			Spec: srrv1alpha1.SharedRouteRegistrySpec{
				EntranceClass: srrv1alpha1.EntranceClassApplication,
				HostPatterns:  []string{"abcd1234.*.olares.com"},
				RouteMode:     srrv1alpha1.RouteModeGateway,
			},
		},
		{
			Spec: srrv1alpha1.SharedRouteRegistrySpec{
				EntranceClass: srrv1alpha1.EntranceClassShared,
				HostPatterns:  []string{"deadbeef.shared.olares.com"},
				RouteMode:     srrv1alpha1.RouteModeGateway,
			},
		},
		{
			Spec: srrv1alpha1.SharedRouteRegistrySpec{
				HostPatterns: []string{"cafebabe.shared.olares.com"},
				RouteMode:    srrv1alpha1.RouteModeGateway,
			},
		},
	}
	auth, tlsHosts := enumerateHostsForViewer("alice", srrs, "olares.com")
	wantAuth := map[string]bool{
		"abcd1234.alice.olares.com":  true,
		"deadbeef.shared.olares.com": true,
		"cafebabe.shared.olares.com": true,
	}
	if len(auth) != 3 {
		t.Fatalf("auth=%v", auth)
	}
	for _, h := range auth {
		if !wantAuth[h] {
			t.Fatalf("unexpected auth host %q", h)
		}
	}
	if len(tlsHosts) != 1 || tlsHosts[0] != "abcd1234.alice.olares.com" {
		t.Fatalf("tls=%v, want only application viewer host", tlsHosts)
	}
}

func TestBuildSharedHostsConfigMapDataWritesTLSKey(t *testing.T) {
	data := buildSharedHostsConfigMapData([]SharedHostsTarget{{
		CallerNamespace: "chat-alice",
		Viewer:          "alice",
		Hosts:           []string{"abcd1234.alice.olares.com", "deadbeef.shared.olares.com"},
		TLSHosts:        []string{"abcd1234.alice.olares.com"},
	}})
	authBody := data[constants.MeshInSharedHostsFileName]
	tlsBody := data[constants.MeshInTLSHostsFileName]
	if !strings.Contains(authBody, "deadbeef.shared.olares.com") || !strings.Contains(authBody, "abcd1234.alice.olares.com") {
		t.Fatalf("auth body=%q", authBody)
	}
	if !strings.Contains(tlsBody, "abcd1234.alice.olares.com") {
		t.Fatalf("tls body missing application host: %q", tlsBody)
	}
	if strings.Contains(tlsBody, "deadbeef.shared.olares.com") {
		t.Fatalf("tls body must not include shared host: %q", tlsBody)
	}
}

func TestMeshInTLSSecretName(t *testing.T) {
	if got := meshInTLSSecretName("Alice"); got != "olares-mesh-in-tls-alice" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildDemandIndexTargetsCallerNS(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	callerNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "litellm-vodevall",
			Labels: map[string]string{
				security.NamespaceInClusterCallerLabel: "true",
				"bytetrade.io/ns-owner":                "vodevall",
			},
		},
	}
	sharedNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollamallmbasev316f7cd-shared",
			Labels: map[string]string{
				"bytetrade.io/ns-shared": "true",
				"bytetrade.io/ns-owner":  "vodevall",
			},
		},
	}
	sharedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "llminit",
			Namespace: "ollamallmbasev316f7cd-shared",
			Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(callerNS, sharedNS, sharedPod).Build()
	got, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("demand len=%d want 1: %+v", len(got), got)
	}
	if got[0].CallerNamespace != "litellm-vodevall" || got[0].CertViewer != "vodevall" {
		t.Fatalf("got %+v", got[0])
	}
}

func TestCallerSharedAppRefsFromDecideEdges(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "litellm-vodevall-litellm",
			Annotations: map[string]string{
				"gateway.olares.io/shared-caller-edges": "ollamallmbasev316f7cd",
			},
		},
		Spec: appv1alpha1.ApplicationSpec{Namespace: "litellm-vodevall"},
	}
	refs := callerSharedAppRefs(app)
	if len(refs) != 1 || refs[0] != "ollamallmbasev316f7cd" {
		t.Fatalf("refs=%v", refs)
	}
}

// sharedHostsScheme registers the types BuildSharedHostsDemand lists.
func sharedHostsScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)
	return scheme
}

func callerNamespace(ns string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
			Labels: map[string]string{security.NamespaceInClusterCallerLabel: "true"},
		},
	}
}

// sharedServerApp is an installed shared app. owner is the installer, which is
// deliberately not the caller viewer in these tests.
func sharedServerApp(name, ns, owner string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns + "-" + name,
			Labels: map[string]string{constants.AppSharedLabel: constants.AppSharedTrue},
		},
		Spec: appv1alpha1.ApplicationSpec{Name: name, Namespace: ns, Owner: owner},
	}
}

func callerApp(name, ns, owner string, settings map[string]string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: ns + "-" + name},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      name,
			Namespace: ns,
			Owner:     owner,
			Settings:  settings,
		},
	}
}

func eligibilityCallerApp(name, ns, owner string) *appv1alpha1.Application {
	return callerApp(name, ns, owner, map[string]string{
		"gateway.olares.io/shared-caller-decide":        "true",
		"gateway.olares.io/shared-caller-decide-source": "eligibility",
	})
}

func gatewaySRR(name, ns, pattern string) *srrv1alpha1.SharedRouteRegistry {
	return &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{pattern},
		},
	}
}

func demandFor(t *testing.T, got []SharedHostsTarget, ns, viewer string) SharedHostsTarget {
	t.Helper()
	for _, target := range got {
		if target.CallerNamespace == ns && target.Viewer == viewer {
			return target
		}
	}
	t.Fatalf("no demand for ns=%s viewer=%s: %+v", ns, viewer, got)
	return SharedHostsTarget{}
}

// TestBuildSharedHostsDemandEligibilityWithoutRefs pins AC-1/AC-2: the host is
// materialized under the caller's own owner even though the shared app was
// installed by someone else, and one caller's demand does not depend on the
// other's (DEFECT-SH-N6-OWNER-01).
func TestBuildSharedHostsDemandEligibilityWithoutRefs(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(
		callerNamespace("chat-alice"),
		callerNamespace("chat-bob"),
		eligibilityCallerApp("chat", "chat-alice", "alice"),
		eligibilityCallerApp("chat", "chat-bob", "bob"),
		sharedServerApp("ollama", "ollama-shared", "admin"),
		gatewaySRR("shared-ollama", "ollama-shared", "abcd1234.*.olares.com"),
	).Build()

	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("demand len=%d want 2: %+v", len(got), got)
	}
	alice := demandFor(t, got, "chat-alice", "alice")
	if len(alice.Hosts) != 1 || alice.Hosts[0] != "abcd1234.alice.olares.com" {
		t.Fatalf("alice hosts=%v", alice.Hosts)
	}
	bob := demandFor(t, got, "chat-bob", "bob")
	if len(bob.Hosts) != 1 || bob.Hosts[0] != "abcd1234.bob.olares.com" {
		t.Fatalf("bob hosts=%v", bob.Hosts)
	}
}

// TestBuildSharedHostsDemandCoversEverySharedApp pins AC-3: an eligibility
// caller may reach any installed shared app, so every shared gateway SRR is
// materialized, not only those of the app its owner installed.
func TestBuildSharedHostsDemandCoversEverySharedApp(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(
		callerNamespace("chat-alice"),
		eligibilityCallerApp("chat", "chat-alice", "alice"),
		sharedServerApp("ollama", "ollama-shared", "admin"),
		gatewaySRR("shared-ollama", "ollama-shared", "abcd1234.*.olares.com"),
		sharedServerApp("whisper", "whisper-shared", "bob"),
		gatewaySRR("shared-whisper", "whisper-shared", "deadbeef.*.olares.com"),
	).Build()

	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	alice := demandFor(t, got, "chat-alice", "alice")
	want := []string{"abcd1234.alice.olares.com", "deadbeef.alice.olares.com"}
	if len(alice.Hosts) != len(want) {
		t.Fatalf("hosts=%v want %v", alice.Hosts, want)
	}
	for i, h := range want {
		if alice.Hosts[i] != h {
			t.Fatalf("hosts=%v want %v", alice.Hosts, want)
		}
	}
}

// TestBuildSharedHostsDemandNamedDepsUseCallerViewer pins AC-4: a caller with
// named shared deps gets only that callee's routes, still materialized under
// its own viewer rather than the callee installer's.
func TestBuildSharedHostsDemandNamedDepsUseCallerViewer(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(
		callerNamespace("litellm-alice"),
		callerApp("litellm", "litellm-alice", "alice", map[string]string{
			"gateway.olares.io/shared-caller-decide": "true",
			"clusterAppRef":                          "ollama",
		}),
		sharedServerApp("ollama", "ollama-shared", "admin"),
		gatewaySRR("shared-ollama", "ollama-shared", "abcd1234.*.olares.com"),
		sharedServerApp("whisper", "whisper-shared", "admin"),
		gatewaySRR("shared-whisper", "whisper-shared", "deadbeef.*.olares.com"),
	).Build()

	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("demand len=%d want 1: %+v", len(got), got)
	}
	target := demandFor(t, got, "litellm-alice", "alice")
	if len(target.Hosts) != 1 || target.Hosts[0] != "abcd1234.alice.olares.com" {
		t.Fatalf("hosts=%v want only the named callee under the caller viewer", target.Hosts)
	}
}

func TestBuildSharedHostsDemandDropsUnresolvedRef(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(
		callerNamespace("litellm-alice"),
		callerApp("litellm", "litellm-alice", "alice", map[string]string{
			"gateway.olares.io/shared-caller-decide": "true",
			"clusterAppRef":                          "gone",
		}),
		sharedServerApp("ollama", "ollama-shared", "admin"),
		gatewaySRR("shared-ollama", "ollama-shared", "abcd1234.*.olares.com"),
	).Build()

	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("demand=%+v want none for an unresolved dep", got)
	}
}

// TestBuildSharedHostsDemandSkipsNonSharedSRR pins AC-5: gateway SRRs outside a
// shared server namespace are private per-user routes and must stay out of the
// eligibility allowlist.
func TestBuildSharedHostsDemandSkipsNonSharedSRR(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(
		callerNamespace("chat-alice"),
		eligibilityCallerApp("chat", "chat-alice", "alice"),
		callerApp("private", "private-bob", "bob", nil),
		gatewaySRR("private-route", "private-bob", "cafebabe.*.olares.com"),
	).Build()

	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	target := demandFor(t, got, "chat-alice", "alice")
	if len(target.Hosts) != 0 {
		t.Fatalf("hosts=%v want none: non-shared SRR leaked into the allowlist", target.Hosts)
	}
}

// TestNoteEmptySharedHostsTargets pins AC-6: an opted-in caller that resolves no
// host is reported instead of silently falling back to passthrough.
func TestNoteEmptySharedHostsTargets(t *testing.T) {
	got := noteEmptySharedHostsTargets("chat-alice", []SharedHostsTarget{
		{CallerNamespace: "chat-alice", Viewer: "alice"},
		{CallerNamespace: "chat-alice", Viewer: "bob", Hosts: []string{"abcd1234.bob.olares.com"}},
	})
	if got != 1 {
		t.Fatalf("empty targets reported = %d want 1", got)
	}
}

func sharedHostsRequest(ns string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: ns, Name: constants.MeshInSharedHostsCMName,
	}}
}

// sharedHostsCallerFixture builds an opted-in caller NS owned by alice plus a
// shared app installed by admin, so the single logical SRR pattern resolves to
// abcd1234.alice.olares.com.
func sharedHostsCallerFixture(extra ...client.Object) client.Client {
	objs := []client.Object{
		callerNamespace("chat-alice"),
		callerApp("chat", "chat-alice", "alice", map[string]string{
			"gateway.olares.io/shared-caller-decide": "true",
		}),
		sharedServerApp("ollama", "ollama-shared", "admin"),
		gatewaySRR("shared-ollama", "ollama-shared", "abcd1234.*.olares.com"),
	}
	return fake.NewClientBuilder().WithScheme(sharedHostsScheme()).WithObjects(append(objs, extra...)...).Build()
}

func TestMeshInSharedHostsKeepsAllowlistWhenPlatformDomainUnavailable(t *testing.T) {
	cluster.PrimePlatformDomainForTest("")
	defer cluster.ResetPlatformDomainForTest()

	c := sharedHostsCallerFixture(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MeshInSharedHostsCMName,
			Namespace: "chat-alice",
			Labels: map[string]string{
				constants.MeshInSharedHostsManagedByLabel: sharedHostsManagedByValue,
			},
		},
		Data: map[string]string{
			constants.MeshInSharedHostsFileName: sharedHostsFileText([]string{"abcd1234.alice.olares.com"}),
		},
	})
	r := &MeshInSharedHostsReconciler{Client: c}

	res, err := r.Reconcile(context.Background(), sharedHostsRequest("chat-alice"))
	if err != nil {
		t.Fatal(err)
	}
	if res.RequeueAfter != sharedHostsPlatformDomainRequeue {
		t.Fatalf("RequeueAfter=%s want %s", res.RequeueAfter, sharedHostsPlatformDomainRequeue)
	}
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "chat-alice", Name: constants.MeshInSharedHostsCMName,
	}, cm); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cm.Data[constants.MeshInSharedHostsFileName], "abcd1234.alice.olares.com") {
		t.Fatalf("allowlist blanked while platform domain unavailable: %q",
			cm.Data[constants.MeshInSharedHostsFileName])
	}
}

// TestMeshInSharedHostsResolvesPlatformDomainAfterStartupRace pins the fix for
// the one-shot capture: the domain was read once in SetupWithManager, so a
// lookup that lost the race against cluster owner provisioning left the
// reconciler with "" — and an empty allowlist — for the whole process lifetime.
func TestMeshInSharedHostsResolvesPlatformDomainAfterStartupRace(t *testing.T) {
	cluster.PrimePlatformDomainForTest("")
	defer cluster.ResetPlatformDomainForTest()

	c := sharedHostsCallerFixture()
	r := &MeshInSharedHostsReconciler{Client: c}
	ctx := context.Background()

	res, err := r.Reconcile(ctx, sharedHostsRequest("chat-alice"))
	if err != nil {
		t.Fatal(err)
	}
	if res.RequeueAfter != sharedHostsPlatformDomainRequeue {
		t.Fatalf("first reconcile RequeueAfter=%s want %s", res.RequeueAfter, sharedHostsPlatformDomainRequeue)
	}
	cm := &corev1.ConfigMap{}
	err = c.Get(ctx, types.NamespacedName{
		Namespace: "chat-alice", Name: constants.MeshInSharedHostsCMName,
	}, cm)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected no ConfigMap while domain unavailable, err=%v data=%v", err, cm.Data)
	}

	cluster.PrimePlatformDomainForTest("olares.com")
	res, err = r.Reconcile(ctx, sharedHostsRequest("chat-alice"))
	if err != nil {
		t.Fatal(err)
	}
	if res.RequeueAfter != 0 {
		t.Fatalf("second reconcile requeued: %s", res.RequeueAfter)
	}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "chat-alice", Name: constants.MeshInSharedHostsCMName,
	}, cm); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cm.Data[constants.MeshInSharedHostsFileName], "abcd1234.alice.olares.com") {
		t.Fatalf("allowlist body = %q", cm.Data[constants.MeshInSharedHostsFileName])
	}
}

func TestResolvePlatformDomainPrefersConfiguredValue(t *testing.T) {
	cluster.PrimePlatformDomainForTest("")
	defer cluster.ResetPlatformDomainForTest()

	r := &MeshInSharedHostsReconciler{platformDomain: " Olares.COM "}
	if got := r.resolvePlatformDomain(context.Background()); got != "olares.com" {
		t.Fatalf("configured domain = %q", got)
	}
	empty := &MeshInSharedHostsReconciler{}
	if got := empty.resolvePlatformDomain(context.Background()); got != "" {
		t.Fatalf("unresolvable domain = %q", got)
	}
}

func TestIsClusterScopedOrCallerAppEligibility(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Settings: map[string]string{
				"gateway.olares.io/shared-caller-decide": "true",
			},
		},
	}
	if !isClusterScopedOrCallerApp(app) {
		t.Fatal("decide=true without refs must watch as caller app")
	}
}

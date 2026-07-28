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
	_, reason = materializeHost("x.shared.olares.com", "alice", "olares.com")
	if reason == "" {
		t.Fatal("expected v2 guard drop reason")
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

func TestBuildSharedHostsDemandEligibilityWithoutRefs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)

	callerNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "chat-alice",
			Labels: map[string]string{security.NamespaceInClusterCallerLabel: "true"},
		},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "chat-alice-chat"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "chat",
			Namespace: "chat-alice",
			Owner:     "alice",
			Settings: map[string]string{
				"gateway.olares.io/shared-caller-decide":        "true",
				"gateway.olares.io/shared-caller-decide-source": "eligibility",
			},
		},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo", Namespace: "user-space-alice"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{"abcd1234.*.olares.com"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(callerNS, app, srr).Build()
	got, err := BuildSharedHostsDemand(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("demand len=%d want 1: %+v", len(got), got)
	}
	if got[0].CallerNamespace != "chat-alice" || got[0].Viewer != "alice" {
		t.Fatalf("got %+v", got[0])
	}
	if len(got[0].Hosts) != 1 || got[0].Hosts[0] != "abcd1234.alice.olares.com" {
		t.Fatalf("hosts=%v", got[0].Hosts)
	}
}

func sharedHostsRequest(ns string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: ns, Name: constants.MeshInSharedHostsCMName,
	}}
}

// sharedHostsCallerFixture builds an opted-in caller NS whose viewer resolves a
// single logical SRR pattern to abcd1234.alice.olares.com.
func sharedHostsCallerFixture(extra ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)

	objs := []client.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "chat-alice",
				Labels: map[string]string{security.NamespaceInClusterCallerLabel: "true"},
			},
		},
		&appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: "chat-alice-chat"},
			Spec: appv1alpha1.ApplicationSpec{
				Name:      "chat",
				Namespace: "chat-alice",
				Owner:     "alice",
				Settings: map[string]string{
					"gateway.olares.io/shared-caller-decide": "true",
				},
			},
		},
		&srrv1alpha1.SharedRouteRegistry{
			ObjectMeta: metav1.ObjectMeta{Name: "shared-demo", Namespace: "user-space-alice"},
			Spec: srrv1alpha1.SharedRouteRegistrySpec{
				RouteMode:    srrv1alpha1.RouteModeGateway,
				HostPatterns: []string{"abcd1234.*.olares.com"},
			},
		},
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(append(objs, extra...)...).Build()
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

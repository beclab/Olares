package routecontrol

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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


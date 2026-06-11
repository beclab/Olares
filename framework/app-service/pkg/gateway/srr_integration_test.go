package gateway

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func newReconcileFixture(t *testing.T) (client.Client, *appv1alpha1.Application, *corev1.Service) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add app.bytetrade.io scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add gateway scheme: %v", err)
	}

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama-shared-ollama",
			UID:  types.UID("ollama-uid"),
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434, URL: "abc.shared.example.com"},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
	return c, app, svc
}

func TestReconcile_CreateThenUpdate(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	ctx := context.Background()
	spec, err := BuildSpec(app, svc)
	if err != nil {
		t.Fatalf("BuildSpec: %v", err)
	}
	created, err := Reconcile(ctx, c, app, spec)
	if err != nil {
		t.Fatalf("Reconcile create: %v", err)
	}
	if created.Name != "shared-ollama" || created.Namespace != "ollama-shared" {
		t.Fatalf("unexpected SRR identity: %+v", created.ObjectMeta)
	}
	if len(created.OwnerReferences) != 1 || created.OwnerReferences[0].UID != app.UID {
		t.Fatalf("missing/incorrect ownerReference: %+v", created.OwnerReferences)
	}

	app.Spec.SharedEntrances[0].URL = "DEF.shared.example.com"
	spec2, err := BuildSpec(app, svc)
	if err != nil {
		t.Fatalf("BuildSpec 2: %v", err)
	}
	if _, err := Reconcile(ctx, c, app, spec2); err != nil {
		t.Fatalf("Reconcile update: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama"}, got); err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if len(got.Spec.HostPatterns) != 1 || got.Spec.HostPatterns[0] != "def.shared.example.com" {
		t.Fatalf("hostPatterns not updated: %v", got.Spec.HostPatterns)
	}
}

func TestDelete_IdempotentAndCleansUp(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	ctx := context.Background()
	spec, _ := BuildSpec(app, svc)
	if _, err := Reconcile(ctx, c, app, spec); err != nil {
		t.Fatalf("seed Reconcile: %v", err)
	}

	if err := Delete(ctx, c, app); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama"}, got)
	if err == nil {
		t.Fatalf("SRR still present after Delete")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected error after Delete: %v", err)
	}

	if err := Delete(ctx, c, app); err != nil {
		t.Fatalf("second Delete should be idempotent: %v", err)
	}
}

// Per-entrance SRRs and logical hostPattern uniqueness.

func TestReconcileForEntrance_CreateAndUpdate(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	app.Spec.Appid = "a5be2268"
	ctx := context.Background()
	ent := app.Spec.SharedEntrances[0]
	ent.Name = "ollamav2"
	app.Spec.SharedEntrances[0] = ent

	spec, err := BuildSpecForEntrance(app, ent, 0, svc, "olares.com")
	if err != nil {
		t.Fatalf("BuildSpecForEntrance: %v", err)
	}
	if _, err := ReconcileForEntrance(ctx, c, app, ent, spec); err != nil {
		t.Fatalf("ReconcileForEntrance create: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	appid := appcfg.AppName(app.Spec.Name).GetAppID()
	name := ResourceNameForEntrance(appid, "ollamav2")
	if err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: name}, got); err != nil {
		t.Fatalf("get %s: %v", name, err)
	}
	if got.Labels["gateway.olares.io/appid"] != appid {
		t.Fatalf("missing appid label: %v", got.Labels)
	}
	if got.Labels["gateway.olares.io/entrance"] != "ollamav2" {
		t.Fatalf("missing entrance label: %v", got.Labels)
	}
	if len(got.Spec.HostPatterns) != 1 || !IsLogicalHostPattern(got.Spec.HostPatterns[0]) {
		t.Fatalf("HostPatterns not logical: %v", got.Spec.HostPatterns)
	}
	// Update path: re-reconcile with same spec should be a no-op (object stable).
	if _, err := ReconcileForEntrance(ctx, c, app, ent, spec); err != nil {
		t.Fatalf("re-reconcile: %v", err)
	}
}

func TestCheckLogicalPatternUniqueness(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	app.Spec.Appid = "a5be2268"
	ctx := context.Background()
	ent := app.Spec.SharedEntrances[0]
	ent.Name = "ollamav2"
	spec, err := BuildSpecForEntrance(app, ent, 0, svc, "olares.com")
	if err != nil {
		t.Fatalf("BuildSpecForEntrance: %v", err)
	}
	if _, err := ReconcileForEntrance(ctx, c, app, ent, spec); err != nil {
		t.Fatalf("seed: %v", err)
	}
	pat := spec.HostPatterns[0]

	// Self-namespaced lookup should not flag collision.
	appid := appcfg.AppName(app.Spec.Name).GetAppID()
	if err := CheckLogicalPatternUniqueness(ctx, c, pat, "ollama-shared", ResourceNameForEntrance(appid, "ollamav2")); err != nil {
		t.Fatalf("self-uniqueness false positive: %v", err)
	}
	// A different SRR carrying the same pattern in another namespace must fail.
	other := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-x-y", Namespace: "x-shared"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{pat},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "svc", Port: 80},
		},
	}
	if err := c.Create(ctx, other); err != nil {
		t.Fatalf("seed other: %v", err)
	}
	err = CheckLogicalPatternUniqueness(ctx, c, pat, "ollama-shared", ResourceNameForEntrance(appid, "ollamav2"))
	if err == nil {
		t.Fatal("expected HASH8_COLLISION error")
	}
}

func TestPruneEntranceSRRs(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	app.Spec.Appid = "a5be2268"
	ctx := context.Background()

	// Seed two per-entrance SRRs labeled with the same instance.
	for _, name := range []string{"ollamav2", "ollamav3"} {
		ent := app.Spec.SharedEntrances[0]
		ent.Name = name
		spec, _ := BuildSpecForEntrance(app, ent, 0, svc, "olares.com")
		if _, err := ReconcileForEntrance(ctx, c, app, ent, spec); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	keep := map[string]struct{}{
		ResourceNameForEntrance(appcfg.AppName(app.Spec.Name).GetAppID(), "ollamav2"): {},
	}
	if err := PruneEntranceSRRs(ctx, c, app, keep); err != nil {
		t.Fatalf("PruneEntranceSRRs: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	appid := appcfg.AppName(app.Spec.Name).GetAppID()
	err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: ResourceNameForEntrance(appid, "ollamav3")}, got)
	if err == nil {
		t.Fatal("stale SRR not pruned")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: ResourceNameForEntrance(appid, "ollamav2")}, got); err != nil {
		t.Fatalf("kept SRR missing: %v", err)
	}
}

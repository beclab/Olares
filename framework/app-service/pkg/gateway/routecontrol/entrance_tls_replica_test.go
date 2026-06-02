package routecontrol

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	dto "github.com/prometheus/client_model/go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildDemandIndex_TC401_HomeViewerDemand(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "user-space-alice",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
	})
}

func TestBuildDemandIndex_TC402_CrossViewerDemand(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "user-space-alice",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
		newCallerApp("caller-a", "user-space-alice", "wise,calendar"),
		newClusterApp("wise", "bob"),
		newClusterApp("calendar", "charlie"),
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
		{CallerNamespace: "user-space-alice", CertViewer: "bob"},
		{CallerNamespace: "user-space-alice", CertViewer: "charlie"},
	})
}

func TestBuildDemandIndex_TC402b_UnresolvedAppRef(t *testing.T) {
	before := counterValue(t, replicaErrorsTotal.WithLabelValues("app_ref_unresolved"))
	c := newReplicaFixture(t,
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "user-space-alice",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
		newCallerApp("caller-a", "user-space-alice", "ghost,wise"),
		newClusterApp("wise", "bob"),
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
		{CallerNamespace: "user-space-alice", CertViewer: "bob"},
	})
	after := counterValue(t, replicaErrorsTotal.WithLabelValues("app_ref_unresolved"))
	if after-before != 1 {
		t.Fatalf("app_ref_unresolved delta=%v want=1", after-before)
	}
}

func TestBuildDemandIndex_TC402c_MultiOwnerWarnOnly(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "user-space-alice",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
		newCallerApp("caller-a", "user-space-alice", "dup"),
		newClusterApp("dup", "bob"),
		newClusterApp("dup", "charlie"),
	)

	before := counterValue(t, replicaErrorsTotal.WithLabelValues("app_ref_multi_owner"))
	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
		{CallerNamespace: "user-space-alice", CertViewer: "bob"},
		{CallerNamespace: "user-space-alice", CertViewer: "charlie"},
	})
	after := counterValue(t, replicaErrorsTotal.WithLabelValues("app_ref_multi_owner"))
	if after != before {
		t.Fatalf("multi owner must not increment error counter: before=%v after=%v", before, after)
	}
}

func TestBuildDemandIndex_TC402d_PerfAndCallCounts(t *testing.T) {
	objs := []client.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "user-space-alice",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
		newCallerApp("caller-a", "user-space-alice", "app-0,app-1,app-2,app-3,app-4,app-5,app-6,app-7,app-8,app-9"),
	}
	for i := 0; i < 1000; i++ {
		objs = append(objs, newClusterApp(
			"app-"+fmt.Sprintf("%d", i),
			"owner-"+fmt.Sprintf("%d", i),
		))
	}
	c := newReplicaFixture(t, objs...)

	buildCalls := 0
	resolveCalls := 0
	testBuildClusterAppOwnerIndexHook = func() { buildCalls++ }
	testResolveClusterAppOwnerHook = func() { resolveCalls++ }
	defer func() {
		testBuildClusterAppOwnerIndexHook = nil
		testResolveClusterAppOwnerHook = nil
	}()

	start := time.Now()
	_, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Fatalf("BuildDemandIndex too slow: %v", elapsed)
	}
	if buildCalls != 1 {
		t.Fatalf("buildClusterAppOwnerIndex calls=%d want=1", buildCalls)
	}
	if resolveCalls != 10 {
		t.Fatalf("resolveClusterAppOwner calls=%d want=10", resolveCalls)
	}
}

func TestBuildDemandIndex_TC403_SharedNamespaceOwnerDemand(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "ollamav3-shared",
				Labels: map[string]string{nsOwnerLabel: "brucedai"},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "ollamav3-shared",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "ollamav3-shared", CertViewer: "brucedai"},
	})
}

func TestBuildDemandIndex_TC404_SharedNamespaceWithoutOwnerSkipped(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ollamav3-shared",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "ollamav3-shared",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, nil)
}

func TestBuildDemandIndex_TC405_SharedNamespaceSkipsCrossViewerBranch(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "ollamav3-shared",
				Labels: map[string]string{nsOwnerLabel: "brucedai"},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "ollamav3-shared",
				Labels:    map[string]string{constants.AppSharedEntrancesLabel: "true"},
			},
		},
		newCallerApp("caller-a", "ollamav3-shared", "wise"),
		newClusterApp("wise", "alice"),
	)

	index, err := BuildDemandIndex(context.Background(), c, "olares.com")
	if err != nil {
		t.Fatalf("BuildDemandIndex: %v", err)
	}
	assertTargets(t, index, []ReplicaTarget{
		{CallerNamespace: "ollamav3-shared", CertViewer: "brucedai"},
	})
}

func TestSyncReplicasForViewer_TC403_NoDemandNoCreate(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice"}},
		newSourceSecret("alice", "cert-a", "key-a"),
	)
	if err := SyncReplicasForViewer(context.Background(), c, "alice", nil); err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice",
		Name:      entranceTLSSecretName("alice"),
	}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("replica should not exist, got err=%v", err)
	}
}

func TestReconcileReplica_TC404_HashNoop(t *testing.T) {
	before := counterValue(t, replicaSyncTotal.WithLabelValues("noop"))
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		newSourceSecret("alice", "cert-a", "key-a"),
		newReplicaSecret("user-space-alice", "alice", "cert-a", "key-a"),
	)
	wrote, err := ReconcileReplica(context.Background(), c, ReplicaTarget{
		CallerNamespace: "user-space-alice",
		CertViewer:      "alice",
	})
	if err != nil {
		t.Fatalf("ReconcileReplica: %v", err)
	}
	if wrote {
		t.Fatal("expected noop")
	}
	after := counterValue(t, replicaSyncTotal.WithLabelValues("noop"))
	if after-before != 1 {
		t.Fatalf("noop delta=%v want=1", after-before)
	}
}

func TestReconcileReplica_TC405_HashUpdate(t *testing.T) {
	before := counterValue(t, replicaSyncTotal.WithLabelValues("updated"))
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		newSourceSecret("alice", "cert-b", "key-b"),
		newReplicaSecret("user-space-alice", "alice", "cert-a", "key-a"),
	)
	wrote, err := ReconcileReplica(context.Background(), c, ReplicaTarget{
		CallerNamespace: "user-space-alice",
		CertViewer:      "alice",
	})
	if err != nil {
		t.Fatalf("ReconcileReplica: %v", err)
	}
	if !wrote {
		t.Fatal("expected update write")
	}
	replica := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}, replica); err != nil {
		t.Fatalf("get replica: %v", err)
	}
	if replica.Annotations[annotationTLSContentHash] != tlsMaterialHash("cert-b", "key-b") {
		t.Fatalf("hash=%q", replica.Annotations[annotationTLSContentHash])
	}
	after := counterValue(t, replicaSyncTotal.WithLabelValues("updated"))
	if after-before != 1 {
		t.Fatalf("updated delta=%v want=1", after-before)
	}
}

func TestSyncReplicasForViewer_TCA61_BumpSharedWorkloadWhenReplicaChanged(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ollamav3-shared", UID: types.UID("ns-shared")}},
		newSourceSecretWithRV("brucedai", "cert-a", "key-a", "rv-1"),
		newDeployment("ollamav3-shared", "app-deploy", nil),
		newReplicaSet("ollamav3-shared", "app-rs", "app-deploy"),
		newSharedEntrancePod("ollamav3-shared", "entrance-0", "app-rs", false),
	)

	err := SyncReplicasForViewer(context.Background(), c, "brucedai", []ReplicaTarget{
		{CallerNamespace: "ollamav3-shared", CertViewer: "brucedai"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}

	deploy := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamav3-shared",
		Name:      "app-deploy",
	}, deploy); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if got := deploy.Spec.Template.Annotations[annotationD2ReplicaRevision]; got != "rv-1" {
		t.Fatalf("annotation %s=%q want rv-1", annotationD2ReplicaRevision, got)
	}
}

func TestSyncReplicasForViewer_TCA62_NoBumpWhenD2AlreadyPresent(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ollamav3-shared", UID: types.UID("ns-shared")}},
		newSourceSecretWithRV("brucedai", "cert-a", "key-a", "rv-2"),
		newDeployment("ollamav3-shared", "app-deploy", nil),
		newReplicaSet("ollamav3-shared", "app-rs", "app-deploy"),
		newSharedEntrancePod("ollamav3-shared", "entrance-0", "app-rs", true),
	)

	err := SyncReplicasForViewer(context.Background(), c, "brucedai", []ReplicaTarget{
		{CallerNamespace: "ollamav3-shared", CertViewer: "brucedai"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}

	deploy := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamav3-shared",
		Name:      "app-deploy",
	}, deploy); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if got := deploy.Spec.Template.Annotations[annotationD2ReplicaRevision]; got != "" {
		t.Fatalf("annotation should stay empty, got=%q", got)
	}
}

func TestSyncReplicasForViewer_TCA63_NoBumpWhenRevisionAlreadyCurrent(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ollamav3-shared", UID: types.UID("ns-shared")}},
		newSourceSecretWithRV("brucedai", "cert-a", "key-a", "rv-3"),
		newDeployment("ollamav3-shared", "app-deploy", map[string]string{annotationD2ReplicaRevision: "rv-3"}),
		newReplicaSet("ollamav3-shared", "app-rs", "app-deploy"),
		newSharedEntrancePod("ollamav3-shared", "entrance-0", "app-rs", false),
	)

	err := SyncReplicasForViewer(context.Background(), c, "brucedai", []ReplicaTarget{
		{CallerNamespace: "ollamav3-shared", CertViewer: "brucedai"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}

	deploy := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamav3-shared",
		Name:      "app-deploy",
	}, deploy); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if got := deploy.Spec.Template.Annotations[annotationD2ReplicaRevision]; got != "rv-3" {
		t.Fatalf("annotation should stay rv-3, got=%q", got)
	}
}

func TestSyncReplicasForViewer_TCA64_UserSpaceReplicaDoesNotBump(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-brucedai", UID: types.UID("ns-user")}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ollamav3-shared", UID: types.UID("ns-shared")}},
		newSourceSecretWithRV("brucedai", "cert-a", "key-a", "rv-4"),
		newDeployment("ollamav3-shared", "app-deploy", nil),
		newReplicaSet("ollamav3-shared", "app-rs", "app-deploy"),
		newSharedEntrancePod("ollamav3-shared", "entrance-0", "app-rs", false),
	)

	err := SyncReplicasForViewer(context.Background(), c, "brucedai", []ReplicaTarget{
		{CallerNamespace: "user-space-brucedai", CertViewer: "brucedai"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}

	deploy := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamav3-shared",
		Name:      "app-deploy",
	}, deploy); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if got := deploy.Spec.Template.Annotations[annotationD2ReplicaRevision]; got != "" {
		t.Fatalf("annotation should stay empty, got=%q", got)
	}
}

func TestSyncReplicasForViewer_TC406_DemandGoneDeletesReplica(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-bob", UID: types.UID("ns-bob")}},
		newSourceSecret("alice", "cert-a", "key-a"),
		newReplicaSecret("user-space-bob", "alice", "cert-a", "key-a"),
	)
	err := SyncReplicasForViewer(context.Background(), c, "alice", []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}
	err = c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-bob", Name: entranceTLSSecretName("alice"),
	}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("stale replica should be deleted, got err=%v", err)
	}
}

func TestSyncReplicasForViewer_TC407_SourceDeleteGCAllReplicas(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-bob", UID: types.UID("ns-bob")}},
		newReplicaSecret("user-space-alice", "alice", "cert-a", "key-a"),
		newReplicaSecret("user-space-bob", "alice", "cert-a", "key-a"),
	)
	err := SyncReplicasForViewer(context.Background(), c, "alice", []ReplicaTarget{
		{CallerNamespace: "user-space-alice", CertViewer: "alice"},
		{CallerNamespace: "user-space-bob", CertViewer: "alice"},
	})
	if err != nil {
		t.Fatalf("SyncReplicasForViewer: %v", err)
	}
	for _, ns := range []string{"user-space-alice", "user-space-bob"} {
		err = c.Get(context.Background(), types.NamespacedName{
			Namespace: ns, Name: entranceTLSSecretName("alice"),
		}, &corev1.Secret{})
		if !apierrors.IsNotFound(err) {
			t.Fatalf("replica %s should be deleted, err=%v", ns, err)
		}
	}
}

func TestStaticGuards_TC411s(t *testing.T) {
	root := findModuleRoot(t)

	if n := countPatternInGoFiles(t, root, `ListPodsForReload|PushReloadToPods`, true); n != 0 {
		t.Fatalf("forbidden pod push symbols=%d", n)
	}
	if n := countPatternInGoFiles(t, root, `olares_d2_replica_|d2\.replica\.|D2_REPLICA_`, true); n == 0 {
		t.Fatal("expected D2 replica naming markers")
	}
	if n := countPatternInGoFiles(t, root, `entrance_tls_replica_|tls_offloader_|CERT_RELOAD_|(^|[^A-Z0-9_])REPLICA_[A-Z0-9_]*`, true); n != 0 {
		t.Fatalf("forbidden legacy naming matches=%d", n)
	}
	if matches, _ := filepath.Glob(filepath.Join(root, "pkg", "sandbox", "sidecar", "cert_reload*.go")); len(matches) != 0 {
		t.Fatalf("unexpected cert_reload files: %v", matches)
	}
	if n := countPatternInGoFiles(t, root, `CertReloadWatcher|inotify|nginx -t|nginx -s reload`, true); n != 0 {
		t.Fatalf("forbidden reload symbols=%d", n)
	}
}

func TestApplyReplicaPatch_TC412_SchemaAndLabels(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
	)
	src := newSourceSecret("alice", "cert-a", "key-a")
	dst := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "user-space-alice",
			Name:      entranceTLSSecretName("alice"),
		},
	}
	ns := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "user-space-alice"}, ns); err != nil {
		t.Fatalf("get namespace: %v", err)
	}
	wrote, err := applyReplicaPatch(context.Background(), c, src, dst, ns)
	if err != nil {
		t.Fatalf("applyReplicaPatch: %v", err)
	}
	if !wrote {
		t.Fatal("expected write")
	}
	got := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}, got); err != nil {
		t.Fatalf("get replica: %v", err)
	}
	if got.Type != corev1.SecretTypeTLS {
		t.Fatalf("type=%q", got.Type)
	}
	if got.Labels[ManagedByLabel] != ManagedByValue ||
		got.Labels[labelGatewayRouteControl] != gatewayRouteControlValue ||
		got.Labels[labelTLSViewer] != "alice" ||
		got.Labels[labelTLSReplica] != "true" {
		t.Fatalf("labels=%v", got.Labels)
	}
	if got.Annotations[annotationTLSContentHash] == "" {
		t.Fatal("missing content-hash annotation")
	}
	if len(got.Data) != 2 || got.Data[corev1.TLSCertKey] == nil || got.Data[corev1.TLSPrivateKeyKey] == nil {
		t.Fatalf("invalid data keys=%v", mapKeys(got.Data))
	}
}

func TestFanOutReplica_TC415_IndexFailureKeepsLastDemand(t *testing.T) {
	before := counterValue(t, replicaErrorsTotal.WithLabelValues("index_failed"))
	base := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		newSourceSecret("alice", "cert-a", "key-a"),
	)
	c := &selectiveFailListClient{
		Client:    base,
		failKinds: map[string]bool{"PodList": true, "ApplicationList": true},
	}
	r := &EntranceTLSReconciler{
		Client:     c,
		lastDemand: []ReplicaTarget{{CallerNamespace: "user-space-alice", CertViewer: "alice"}},
	}
	if err := r.fanOutReplica(context.Background(), "alice"); err == nil {
		t.Fatal("fanOutReplica should return index rebuild error")
	}
	err := base.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}, &corev1.Secret{})
	if err != nil {
		t.Fatalf("replica should still be created from last demand: %v", err)
	}
	after := counterValue(t, replicaErrorsTotal.WithLabelValues("index_failed"))
	if after-before != 1 {
		t.Fatalf("index_failed delta=%v want=1", after-before)
	}
}

func TestReplicaHashGauge_TC416_AgeAndReset(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
	)
	ns := &corev1.Namespace{}
	_ = c.Get(context.Background(), types.NamespacedName{Name: "user-space-alice"}, ns)
	dst := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}}
	src1 := newSourceSecret("alice", "cert-a", "key-a")

	_, err := applyReplicaPatch(context.Background(), c, src1, dst, ns)
	if err != nil {
		t.Fatalf("apply 1: %v", err)
	}
	v0 := gaugeValue(t, replicaContentHashAgeSeconds.WithLabelValues("alice", "user-space-alice"))
	time.Sleep(20 * time.Millisecond)
	_, err = applyReplicaPatch(context.Background(), c, src1, dst, ns)
	if err != nil {
		t.Fatalf("apply 2: %v", err)
	}
	v1 := gaugeValue(t, replicaContentHashAgeSeconds.WithLabelValues("alice", "user-space-alice"))
	if v1 <= v0 {
		t.Fatalf("hash age should increase: before=%v after=%v", v0, v1)
	}
	src2 := newSourceSecret("alice", "cert-b", "key-b")
	_, err = applyReplicaPatch(context.Background(), c, src2, dst, ns)
	if err != nil {
		t.Fatalf("apply 3: %v", err)
	}
	v2 := gaugeValue(t, replicaContentHashAgeSeconds.WithLabelValues("alice", "user-space-alice"))
	if v2 > 1 {
		t.Fatalf("hash age should reset close to zero, got=%v", v2)
	}
}

func TestApplyReplicaPatch_TC417_OwnerReferenceFields(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
	)
	src := newSourceSecret("alice", "cert-a", "key-a")
	dst := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}}
	ns := &corev1.Namespace{}
	_ = c.Get(context.Background(), types.NamespacedName{Name: "user-space-alice"}, ns)
	if _, err := applyReplicaPatch(context.Background(), c, src, dst, ns); err != nil {
		t.Fatalf("applyReplicaPatch: %v", err)
	}
	got := &corev1.Secret{}
	_ = c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}, got)
	if len(got.OwnerReferences) != 1 {
		t.Fatalf("owner refs=%v", got.OwnerReferences)
	}
	ref := got.OwnerReferences[0]
	if ref.Kind != "Namespace" || ref.Name != "user-space-alice" || ref.UID != types.UID("ns-alice") {
		t.Fatalf("owner ref=%+v", ref)
	}
	if ref.Controller == nil || *ref.Controller {
		t.Fatalf("controller must be false: %+v", ref)
	}
	if ref.BlockOwnerDeletion == nil || *ref.BlockOwnerDeletion {
		t.Fatalf("blockOwnerDeletion must be false: %+v", ref)
	}
}

func TestSweepOrphanReplicas_TC418_SingleThresholdGC(t *testing.T) {
	before := counterValue(t, replicaSyncTotal.WithLabelValues("gc_periodic"))
	c := newReplicaFixture(t,
		newReplicaSecret("user-space-a", "alice", "cert-a", "key-a"),
		newReplicaSecret("user-space-b", "alice", "cert-a", "key-a"),
	)
	err := sweepOrphanReplicas(context.Background(), c, []ReplicaTarget{
		{CallerNamespace: "user-space-a", CertViewer: "alice"},
	})
	if err != nil {
		t.Fatalf("sweepOrphanReplicas: %v", err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-a", Name: entranceTLSSecretName("alice"),
	}, &corev1.Secret{}); err != nil {
		t.Fatalf("replica A should stay: %v", err)
	}
	err = c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-b", Name: entranceTLSSecretName("alice"),
	}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("replica B should be deleted: %v", err)
	}
	after := counterValue(t, replicaSyncTotal.WithLabelValues("gc_periodic"))
	if after-before != 1 {
		t.Fatalf("gc_periodic delta=%v want=1", after-before)
	}
}

func TestSyncReplicasForViewer_TC419_NoDeleteOnToggleThrash(t *testing.T) {
	c := newReplicaFixture(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: defaultGatewayNS}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice", UID: types.UID("ns-alice")}},
		newSourceSecret("alice", "cert-a", "key-a"),
	)
	index := []ReplicaTarget{{CallerNamespace: "user-space-alice", CertViewer: "alice"}}
	for i := 0; i < 3; i++ {
		if err := SyncReplicasForViewer(context.Background(), c, "alice", index); err != nil {
			t.Fatalf("sync loop %d: %v", i, err)
		}
	}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-space-alice", Name: entranceTLSSecretName("alice"),
	}, &corev1.Secret{}); err != nil {
		t.Fatalf("replica should remain during toggle thrash: %v", err)
	}
}

func newReplicaFixture(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}

func newCallerApp(name, ns, refs string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      name,
			Namespace: ns,
			Settings: map[string]string{
				"clusterAppRef": refs,
			},
		},
	}
}

func newClusterApp(name, owner string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-" + owner},
		Spec: appv1alpha1.ApplicationSpec{
			Name:  name,
			Owner: owner,
			Settings: map[string]string{
				"clusterScoped": "true",
			},
		},
	}
}

func newSourceSecret(viewer, cert, key string) *corev1.Secret {
	hash := tlsMaterialHash(cert, key)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultGatewayNS,
			Name:      entranceTLSSecretName(viewer),
			Labels: map[string]string{
				ManagedByLabel:           ManagedByValue,
				labelGatewayRouteControl: gatewayRouteControlValue,
				labelTLSViewer:           viewer,
			},
			Annotations: map[string]string{
				annotationTLSContentHash: hash,
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(cert),
			corev1.TLSPrivateKeyKey: []byte(key),
		},
	}
}

func newReplicaSecret(ns, viewer, cert, key string) *corev1.Secret {
	hash := tlsMaterialHash(cert, key)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      entranceTLSSecretName(viewer),
			Labels: map[string]string{
				ManagedByLabel:           ManagedByValue,
				labelGatewayRouteControl: gatewayRouteControlValue,
				labelTLSViewer:           viewer,
				labelTLSReplica:          "true",
			},
			Annotations: map[string]string{
				annotationTLSContentHash: hash,
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(cert),
			corev1.TLSPrivateKeyKey: []byte(key),
		},
	}
}

func newSourceSecretWithRV(viewer, cert, key, rv string) *corev1.Secret {
	sec := newSourceSecret(viewer, cert, key)
	sec.ResourceVersion = rv
	return sec
}

func newDeployment(namespace, name string, annotations map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
			},
		},
	}
}

func newReplicaSet(namespace, name, deployment string) *appsv1.ReplicaSet {
	controller := true
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deployment,
					Controller: &controller,
				},
			},
		},
	}
}

func newSharedEntrancePod(namespace, name, ownerReplicaSet string, withD2 bool) *corev1.Pod {
	controller := true
	containers := []corev1.Container{{Name: "main"}}
	if withD2 {
		containers = append(containers, corev1.Container{Name: constants.D2SidecarContainerName})
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				constants.AppSharedEntrancesLabel: "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       ownerReplicaSet,
					Controller: &controller,
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}
}

func assertTargets(t *testing.T, got, want []ReplicaTarget) {
	t.Helper()
	toSet := func(in []ReplicaTarget) map[string]struct{} {
		out := make(map[string]struct{}, len(in))
		for _, item := range in {
			out[item.CallerNamespace+"|"+item.CertViewer] = struct{}{}
		}
		return out
	}
	gotSet := toSet(got)
	wantSet := toSet(want)
	if len(gotSet) != len(wantSet) {
		t.Fatalf("target size got=%d want=%d got=%v", len(gotSet), len(wantSet), got)
	}
	for key := range wantSet {
		if _, ok := gotSet[key]; !ok {
			t.Fatalf("missing target %q in %v", key, got)
		}
	}
}

func mapKeys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func counterValue(t *testing.T, c interface{ Write(*dto.Metric) error }) float64 {
	t.Helper()
	m := &dto.Metric{}
	if err := c.Write(m); err != nil {
		t.Fatalf("read counter metric: %v", err)
	}
	return m.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, g interface{ Write(*dto.Metric) error }) float64 {
	t.Helper()
	m := &dto.Metric{}
	if err := g.Write(m); err != nil {
		t.Fatalf("read gauge metric: %v", err)
	}
	return m.GetGauge().GetValue()
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			t.Fatalf("go.mod not found from %s", wd)
		}
		cur = parent
	}
}

func countPatternInGoFiles(t *testing.T, root, pattern string, skipTests bool) int {
	t.Helper()
	re := regexp.MustCompile(pattern)
	hits := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if strings.HasPrefix(base, ".git") || base == "vendor" || base == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if skipTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hits += len(re.FindAll(data, -1))
		return nil
	})
	if err != nil {
		t.Fatalf("walk files: %v", err)
	}
	return hits
}

type selectiveFailListClient struct {
	client.Client
	failKinds map[string]bool
}

func (s *selectiveFailListClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	switch list.(type) {
	case *corev1.PodList:
		if s.failKinds["PodList"] {
			return fmt.Errorf("mock pod list failure")
		}
	case *appv1alpha1.ApplicationList:
		if s.failKinds["ApplicationList"] {
			return fmt.Errorf("mock application list failure")
		}
	}
	return s.Client.List(ctx, list, opts...)
}

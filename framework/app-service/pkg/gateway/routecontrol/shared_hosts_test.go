package routecontrol

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/prometheus/client_golang/prometheus/testutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testPlatformDomain = "example.com"
	testCallerNS       = "litellm-userA"
)

func resetSharedHostsMetrics() {
	sharedHostsReconcileTotal.Reset()
	sharedHostsDropTotal.Reset()
	sharedHostsGCTotal.Reset()
	sharedHostsTargetCount.Reset()
	sharedHostsCount.Reset()
	sharedHostsContentHashAgeSeconds.Reset()
	sharedHostsHashStateMu.Lock()
	sharedHostsHashState = map[string]hashSnapshot{}
	sharedHostsHashStateMu.Unlock()
}

func newSharedHostsClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("srr scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func newOptInNS(name string, optIn bool) *corev1.Namespace {
	labels := map[string]string{}
	if optIn {
		labels[security.NamespaceInClusterCallerLabel] = "true"
	}
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
}

func newPlaceholderCM(ns string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: constants.D2SharedHostsVolumeName, Namespace: ns},
		Data:       map[string]string{constants.D2SharedHostsFileName: ""},
	}
}

func newSRR(ns, name string, patterns ...string) *srrv1alpha1.SharedRouteRegistry {
	return &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: patterns,
		},
	}
}

func runReconcile(t *testing.T, c client.Client, ns string) {
	t.Helper()
	r := &SharedHostsReconciler{Client: c, platformDomain: testPlatformDomain}
	if _, err := r.Reconcile(context.Background(), requestForNS(ns)); err != nil {
		t.Fatalf("reconcile %s: %v", ns, err)
	}
}

func getCM(t *testing.T, c client.Client, ns string) *corev1.ConfigMap {
	t.Helper()
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: ns, Name: constants.D2SharedHostsVolumeName,
	}, cm); err != nil {
		t.Fatalf("get cm %s: %v", ns, err)
	}
	return cm
}

// TC-N6-01: happy single ref / viewer / SRR / host.
func TestSharedHosts_TC01_HappySingleRefViewer(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	cm := getCM(t, c, testCallerNS)
	want := "a1b2c3d4.usera.example.com"
	if !strings.Contains(cm.Data[constants.D2SharedHostsFileName], want) {
		t.Fatalf("shared-hosts.txt missing %s; got=%q", want, cm.Data[constants.D2SharedHostsFileName])
	}
	if !strings.Contains(cm.Data["usera"], want) {
		t.Fatalf("per-viewer usera missing %s; got=%q", want, cm.Data["usera"])
	}
	if cm.Labels[constants.D2SharedHostsManagedByLabel] != sharedHostsManagedByValue {
		t.Fatalf("managed-by label not stamped: %v", cm.Labels)
	}
	if cm.Annotations[sharedHostsContentHashAnnotation] == "" {
		t.Fatalf("content-hash annotation not stamped")
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResUpdated)); v != 1 {
		t.Fatalf("reconcile_total{updated} want 1 got %v", v)
	}
}

// TC-N6-02: multi viewer aggregation across two callers in same NS.
func TestSharedHosts_TC02_MultiViewerAggregation(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm-a", testCallerNS, "ollama"),
		newCallerApp("litellm-b", testCallerNS, "redis"),
		newClusterApp("ollama", "userA"),
		newClusterApp("redis", "userB"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
		newSRR("user-space-userB", "redis-srr", "deadbeef.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	cm := getCM(t, c, testCallerNS)
	all := cm.Data[constants.D2SharedHostsFileName]
	for _, want := range []string{"a1b2c3d4.usera.example.com", "deadbeef.userb.example.com"} {
		if !strings.Contains(all, want) {
			t.Fatalf("shared-hosts.txt missing %s; got=%q", want, all)
		}
	}
	if _, ok := cm.Data["usera"]; !ok {
		t.Fatalf("per-viewer usera missing")
	}
	if _, ok := cm.Data["userb"]; !ok {
		t.Fatalf("per-viewer userb missing")
	}
}

// TC-N6-03: exact host (path A — no wildcard, not replaced by viewer).
func TestSharedHosts_TC03_ExactHostPath(t *testing.T) {
	got, reason := materializeHost("ollama-public.example.com", "userA", testPlatformDomain)
	if got != "ollama-public.example.com" || reason != "" {
		t.Fatalf("want exact-host pass; got=%q reason=%q", got, reason)
	}
}

// TC-N6-04: host guard — non_platform_host.
func TestSharedHosts_TC04_NonPlatformDrop(t *testing.T) {
	resetSharedHostsMetrics()
	got, reason := materializeHost("foo.other-domain.com", "userA", testPlatformDomain)
	if got != "" || reason != rDropNonPlatformHost {
		t.Fatalf("want non_platform_host; got=%q reason=%q", got, reason)
	}
}

// TC-N6-05: host guard — v2_pattern_guarded.
func TestSharedHosts_TC05_V2GuardDrop(t *testing.T) {
	got, reason := materializeHost("abc12345.shared.example.com", "userA", testPlatformDomain)
	if got != "" || reason != rDropV2PatternGuarded {
		t.Fatalf("want v2_pattern_guarded; got=%q reason=%q", got, reason)
	}
}

// TC-N6-06: host guard — multi_wildcard.
func TestSharedHosts_TC06_MultiWildcardDrop(t *testing.T) {
	got, reason := materializeHost("*.x.*.example.com", "userA", testPlatformDomain)
	if got != "" || reason != rDropMultiWildcard {
		t.Fatalf("want multi_wildcard; got=%q reason=%q", got, reason)
	}
}

// TC-N6-07: idempotent — second reconcile is skipped, no extra Update.
func TestSharedHosts_TC07_IdempotentSkip(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	rvAfterFirst := getCM(t, c, testCallerNS).ResourceVersion
	runReconcile(t, c, testCallerNS)
	rvAfterSecond := getCM(t, c, testCallerNS).ResourceVersion
	if rvAfterFirst != rvAfterSecond {
		t.Fatalf("ResourceVersion changed across idempotent reconciles: %s -> %s", rvAfterFirst, rvAfterSecond)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResUpdated)); v != 1 {
		t.Fatalf("reconcile_total{updated} want 1 got %v", v)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResSkipped)); v != 1 {
		t.Fatalf("reconcile_total{skipped} want 1 got %v", v)
	}
}

// TC-N6-08: result/reason switch must be exhaustive — every const reachable.
func TestSharedHosts_TC08_ReasonSwitchExhaustive(t *testing.T) {
	results := []string{
		rResUpdated, rResSkipped, rResDeleted, rResNotFound, rResListFailed,
		rResGetFailed, rResUpdateFailed, rResUpdateConflict, rResSkippedUnmanaged,
	}
	dropReasons := []string{
		rDropOwnerUnresolved, rDropMultiRef, rDropNonPlatformHost, rDropV2PatternGuarded,
		rDropMultiWildcard, rDropInvalidChars, rDropEmptyPatterns,
	}
	gcReasons := []string{rGCNSOptOut, rGCNSDeleted, rGCViewer}
	for _, r := range results {
		if r == "" || r == "other" {
			t.Fatalf("result value invalid: %q", r)
		}
	}
	for _, r := range dropReasons {
		if r == "" || r == "other" {
			t.Fatalf("drop reason value invalid: %q", r)
		}
	}
	for _, r := range gcReasons {
		if r == "" || r == "other" {
			t.Fatalf("gc reason value invalid: %q", r)
		}
	}
	// Literal interlock with WI-T1-3 / WI-T1-5 d2_inject_skipped_total{reason}.
	if rDropMultiRef != "multi_ref_unsupported" {
		t.Fatalf("multi_ref_unsupported drift; got %q", rDropMultiRef)
	}
}

// TC-N6-09: NS opt-out → managed CM deleted; gc_total{ns_opt_out} +1.
func TestSharedHosts_TC09_NSOptOutGC(t *testing.T) {
	resetSharedHostsMetrics()
	cm := newPlaceholderCM(testCallerNS)
	cm.Labels = map[string]string{constants.D2SharedHostsManagedByLabel: sharedHostsManagedByValue}
	cm.Data["userA"] = sharedHostsFileText([]string{"a1b2c3d4.userA.example.com"})
	c := newSharedHostsClient(t, newOptInNS(testCallerNS, false), cm)
	runReconcile(t, c, testCallerNS)
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: testCallerNS, Name: constants.D2SharedHostsVolumeName,
	}, &corev1.ConfigMap{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("CM should be deleted on NS opt-out; err=%v", err)
	}
	if v := testutil.ToFloat64(sharedHostsGCTotal.WithLabelValues(rGCNSOptOut)); v != 1 {
		t.Fatalf("gc_total{ns_opt_out} want 1 got %v", v)
	}
}

// TC-N6-10: NS deleted → gc_total{ns_deleted} +1 (no CM in fixture).
func TestSharedHosts_TC10_NSDeletedGC(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t) // no NS, no CM
	runReconcile(t, c, testCallerNS)
	if v := testutil.ToFloat64(sharedHostsGCTotal.WithLabelValues(rGCNSDeleted)); v != 1 {
		t.Fatalf("gc_total{ns_deleted} want 1 got %v", v)
	}
}

// TC-N6-11: viewer_gone — per-viewer key disappears, gc_total{viewer_gone} +1.
func TestSharedHosts_TC11_ViewerGoneGC(t *testing.T) {
	resetSharedHostsMetrics()
	cm := newPlaceholderCM(testCallerNS)
	cm.Labels = map[string]string{constants.D2SharedHostsManagedByLabel: sharedHostsManagedByValue}
	cm.Data["userA"] = sharedHostsFileText([]string{"old.userA.example.com"})
	cm.Data["userB"] = sharedHostsFileText([]string{"stale.userB.example.com"})
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		cm,
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	after := getCM(t, c, testCallerNS)
	if _, ok := after.Data["userB"]; ok {
		t.Fatalf("per-viewer userB should be GC'd; data=%v", after.Data)
	}
	if v := testutil.ToFloat64(sharedHostsGCTotal.WithLabelValues(rGCViewer)); v < 1 {
		t.Fatalf("gc_total{viewer_gone} want >=1 got %v", v)
	}
}

// TC-N6-12: empty placeholder -> non-empty after reconciler (handoff).
func TestSharedHosts_TC12_PlaceholderToNonEmpty(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	before := getCM(t, c, testCallerNS)
	if before.Data[constants.D2SharedHostsFileName] != "" {
		t.Fatalf("placeholder must start empty; got=%q", before.Data[constants.D2SharedHostsFileName])
	}
	runReconcile(t, c, testCallerNS)
	after := getCM(t, c, testCallerNS)
	if !strings.Contains(after.Data[constants.D2SharedHostsFileName], "a1b2c3d4.usera.example.com") {
		t.Fatalf("after reconcile shared-hosts.txt missing expected host; got=%q",
			after.Data[constants.D2SharedHostsFileName])
	}
	if after.Annotations[sharedHostsContentHashAnnotation] == "" {
		t.Fatalf("content-hash annotation not stamped on handoff")
	}
}

type failingClient struct {
	client.Client
	failSRRList    bool
	conflictOnce   bool
	conflictsFired int
}

func (f *failingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if f.failSRRList {
		if _, ok := list.(*srrv1alpha1.SharedRouteRegistryList); ok {
			return fmt.Errorf("mock srr list failure")
		}
	}
	return f.Client.List(ctx, list, opts...)
}
func (f *failingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if f.conflictOnce && f.conflictsFired == 0 {
		f.conflictsFired++
		gr := schema.GroupResource{Group: "", Resource: "configmaps"}
		return apierrors.NewConflict(gr, obj.GetName(), fmt.Errorf("mock conflict"))
	}
	return f.Client.Update(ctx, obj, opts...)
}

// TC-N6-13: List SRR failure leaves existing CM intact; reconcile_total{list_failed} +1.
func TestSharedHosts_TC13_ListSRRFailure(t *testing.T) {
	resetSharedHostsMetrics()
	cm := newPlaceholderCM(testCallerNS)
	cm.Labels = map[string]string{constants.D2SharedHostsManagedByLabel: sharedHostsManagedByValue}
	cm.Data["userA"] = "previous"
	base := newSharedHostsClient(t, newOptInNS(testCallerNS, true), cm)
	c := &failingClient{Client: base, failSRRList: true}
	r := &SharedHostsReconciler{Client: c, platformDomain: testPlatformDomain}
	if _, err := r.Reconcile(context.Background(), requestForNS(testCallerNS)); err == nil {
		t.Fatalf("expected list err to propagate")
	}
	after := getCM(t, base, testCallerNS)
	if after.Data["userA"] != "previous" {
		t.Fatalf("existing CM clobbered on list failure: %v", after.Data)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResListFailed)); v != 1 {
		t.Fatalf("reconcile_total{list_failed} want 1 got %v", v)
	}
}

// TC-N6-14: CM NotFound -> reconciler does NOT create, waits for webhook.
func TestSharedHosts_TC14_NotFoundWaitPlaceholder(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t, newOptInNS(testCallerNS, true)) // no CM in fixture
	runReconcile(t, c, testCallerNS)
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: testCallerNS, Name: constants.D2SharedHostsVolumeName,
	}, &corev1.ConfigMap{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("reconciler must not Create CM; got err=%v", err)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResNotFound)); v != 1 {
		t.Fatalf("reconcile_total{not_found_wait_placeholder} want 1 got %v", v)
	}
}

// TC-N6-15: Update conflict reported then a fresh reconcile succeeds.
func TestSharedHosts_TC15_UpdateConflictRetried(t *testing.T) {
	resetSharedHostsMetrics()
	base := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	c := &failingClient{Client: base, conflictOnce: true}
	r := &SharedHostsReconciler{Client: c, platformDomain: testPlatformDomain}
	if _, err := r.Reconcile(context.Background(), requestForNS(testCallerNS)); err == nil {
		t.Fatalf("first reconcile should surface conflict err for requeue")
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResUpdateConflict)); v != 1 {
		t.Fatalf("reconcile_total{update_conflict_retried} want 1 got %v", v)
	}
	if _, err := r.Reconcile(context.Background(), requestForNS(testCallerNS)); err != nil {
		t.Fatalf("retry reconcile: %v", err)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResUpdated)); v != 1 {
		t.Fatalf("reconcile_total{updated} want 1 got %v", v)
	}
}

// TC-N6-16: owner_unresolved — clusterAppRef not in ownerIndex.
func TestSharedHosts_TC16_OwnerUnresolved(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ghost"),
	)
	runReconcile(t, c, testCallerNS)
	cm := getCM(t, c, testCallerNS)
	if cm.Data[constants.D2SharedHostsFileName] != sharedHostsFileText(nil) {
		t.Fatalf("unresolved owner must not contribute hosts; got=%q", cm.Data[constants.D2SharedHostsFileName])
	}
	if v := testutil.ToFloat64(sharedHostsDropTotal.WithLabelValues(rDropOwnerUnresolved)); v < 1 {
		t.Fatalf("drop_total{owner_unresolved} want >=1 got %v", v)
	}
}

// TC-N6-17: multi-ref MVP — only primary projected, multi_ref_unsupported emitted.
func TestSharedHosts_TC17_MultiRefMVP(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama,redis"),
		newClusterApp("ollama", "userA"),
		newClusterApp("redis", "userB"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
		newSRR("user-space-userB", "redis-srr", "deadbeef.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	cm := getCM(t, c, testCallerNS)
	if !strings.Contains(cm.Data[constants.D2SharedHostsFileName], "a1b2c3d4.usera.example.com") {
		t.Fatalf("primary ref not projected; got=%q", cm.Data[constants.D2SharedHostsFileName])
	}
	if strings.Contains(cm.Data[constants.D2SharedHostsFileName], "deadbeef.userb.example.com") {
		t.Fatalf("secondary ref must NOT project in MVP; got=%q", cm.Data[constants.D2SharedHostsFileName])
	}
	if v := testutil.ToFloat64(sharedHostsDropTotal.WithLabelValues(rDropMultiRef)); v < 1 {
		t.Fatalf("drop_total{multi_ref_unsupported} want >=1 got %v", v)
	}
}

// TC-N6-18: normalization — case + trim — lowercase host, no leading/trailing ws.
func TestSharedHosts_TC18_NormalizeCaseTrim(t *testing.T) {
	got, reason := materializeHost("  Ollama-Public.Example.COM  ", "USERa", testPlatformDomain)
	if reason != "" {
		t.Fatalf("normalize must not drop; reason=%q", reason)
	}
	if got != "ollama-public.example.com" {
		t.Fatalf("normalize: want lowercase trimmed; got=%q", got)
	}
}

// TC-N6-19: file format — header comments + trailing newlines.
func TestSharedHosts_TC19_FileFormat(t *testing.T) {
	txt := sharedHostsFileText([]string{"a1.userA.example.com", "b2.userA.example.com"})
	if !strings.HasPrefix(txt, "# managed by app-service SharedHostsReconciler") {
		t.Fatalf("missing managed-by header; got=%q", txt)
	}
	for _, h := range []string{"a1.userA.example.com\n", "b2.userA.example.com\n"} {
		if !strings.Contains(txt, h) {
			t.Fatalf("missing line %q in %q", h, txt)
		}
	}
	if !strings.HasSuffix(txt, "\n") {
		t.Fatalf("last line must end with newline")
	}
}

// TC-N6-20: fan-out — single SRR change enqueues every opt-in NS exactly once.
func TestSharedHosts_TC20_FanOut(t *testing.T) {
	c := newSharedHostsClient(t,
		newOptInNS("ns-a", true),
		newOptInNS("ns-b", true),
		newOptInNS("ns-c", false), // opt-out: must NOT be enqueued
	)
	r := &SharedHostsReconciler{Client: c}
	reqs := r.fanOutOnSRR(context.Background(), nil)
	if len(reqs) != 2 {
		t.Fatalf("fan-out want 2 opt-in NS reqs; got=%d (%v)", len(reqs), reqs)
	}
	seen := map[string]int{}
	for _, req := range reqs {
		if req.Name != constants.D2SharedHostsVolumeName {
			t.Fatalf("req.Name must be %s; got %s", constants.D2SharedHostsVolumeName, req.Name)
		}
		seen[req.Namespace]++
	}
	for _, ns := range []string{"ns-a", "ns-b"} {
		if seen[ns] != 1 {
			t.Fatalf("dedupe failed for %s: count=%d", ns, seen[ns])
		}
	}
	if seen["ns-c"] != 0 {
		t.Fatalf("opt-out ns-c must not be enqueued: count=%d", seen["ns-c"])
	}
}

// TC-N6-21: webhook + reconciler don't clobber each other — re-running webhook
// after a reconciler Update must short-circuit on AlreadyExists; the reconciler
// must never remove the managed-by label on subsequent reconciles.
func TestSharedHosts_TC21_WebhookReconcilerCoexist(t *testing.T) {
	resetSharedHostsMetrics()
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		newPlaceholderCM(testCallerNS),
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	// Simulate webhook re-issuing Create (AlreadyExists -> short-circuit, no data overwrite).
	err := c.Create(context.Background(), newPlaceholderCM(testCallerNS))
	if !apierrors.IsAlreadyExists(err) {
		t.Fatalf("webhook re-create must hit AlreadyExists; got=%v", err)
	}
	runReconcile(t, c, testCallerNS) // second reconcile after webhook re-attempt
	cm := getCM(t, c, testCallerNS)
	if cm.Labels[constants.D2SharedHostsManagedByLabel] != sharedHostsManagedByValue {
		t.Fatalf("managed-by label dropped after coexist cycle: %v", cm.Labels)
	}
	if !strings.Contains(cm.Data[constants.D2SharedHostsFileName], "a1b2c3d4.usera.example.com") {
		t.Fatalf("data clobbered after coexist cycle: %q", cm.Data[constants.D2SharedHostsFileName])
	}
}

// Extras: cover the remaining wrapper-shaped mappers + predicates so the
// reconciler file clears the ≥80% per-file coverage gate.
func TestSharedHosts_ExtrasMappersAndPredicates(t *testing.T) {
	c := newSharedHostsClient(t, newOptInNS("ns-a", true), newOptInNS("ns-b", true))
	r := &SharedHostsReconciler{Client: c}
	if reqs := r.fanOutOnPod(context.Background(), nil); len(reqs) != 2 {
		t.Fatalf("fanOutOnPod want 2 opt-in ns reqs; got %d", len(reqs))
	}
	if reqs := r.requeueNamespace(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-x"}}); len(reqs) != 1 || reqs[0].Namespace != "ns-x" {
		t.Fatalf("requeueNamespace want 1 req for ns-x; got %v", reqs)
	}
	if reqs := r.requeueNamespace(context.Background(), nil); reqs != nil {
		t.Fatalf("requeueNamespace nil obj must return nil; got %v", reqs)
	}
	clusterApp := newClusterApp("ollama", "userA")
	if reqs := r.fanOutOnApplication(context.Background(), clusterApp); len(reqs) != 2 {
		t.Fatalf("fanOutOnApplication clusterScoped want 2 reqs; got %d", len(reqs))
	}
	nsApp := newCallerApp("litellm", testCallerNS, "ollama")
	reqs := r.fanOutOnApplication(context.Background(), nsApp)
	if len(reqs) != 1 || reqs[0].Namespace != testCallerNS {
		t.Fatalf("fanOutOnApplication namespaced want 1 req for %s; got %v", testCallerNS, reqs)
	}
	if reqs := r.fanOutOnApplication(context.Background(), nil); reqs != nil {
		t.Fatalf("fanOutOnApplication nil obj must return nil; got %v", reqs)
	}
	v3App := newV3SharedClusterApp("ollamav3", "userA")
	if reqs := r.fanOutOnApplication(context.Background(), v3App); len(reqs) != 2 {
		t.Fatalf("fanOutOnApplication v3 shared (no clusterScoped) want 2 reqs; got %d", len(reqs))
	}
	if !isSharedHostsConfigMap(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: constants.D2SharedHostsVolumeName}}) {
		t.Fatal("isSharedHostsConfigMap should match by name")
	}
	if isSharedHostsConfigMap(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "other"}}) {
		t.Fatal("isSharedHostsConfigMap must reject unrelated CMs")
	}
	if !isGatewayModeSRR(newSRR("user-space-userA", "srr", "x")) {
		t.Fatal("isGatewayModeSRR gateway mode must pass")
	}
	directSRR := newSRR("user-space-userA", "srr", "x")
	directSRR.Spec.RouteMode = srrv1alpha1.RouteModeDirect
	if isGatewayModeSRR(directSRR) {
		t.Fatal("isGatewayModeSRR direct mode must fail")
	}
	if !isClusterScopedOrCallerApp(clusterApp) {
		t.Fatal("isClusterScopedOrCallerApp clusterScoped must pass")
	}
	if !isClusterScopedOrCallerApp(nsApp) {
		t.Fatal("isClusterScopedOrCallerApp caller (clusterAppRef!=\"\") must pass")
	}
	noRef := newCallerApp("noref", testCallerNS, "")
	if isClusterScopedOrCallerApp(noRef) {
		t.Fatal("isClusterScopedOrCallerApp no-ref non-cluster must fail")
	}
	if !isClusterScopedOrCallerApp(v3App) {
		t.Fatal("isClusterScopedOrCallerApp v3 shared (no clusterScoped) must pass")
	}
	if hasInClusterCallerLabel(nil) {
		t.Fatal("hasInClusterCallerLabel nil must be false")
	}
	if !hasInClusterCallerLabel(newOptInNS("x", true)) {
		t.Fatal("hasInClusterCallerLabel opted-in must be true")
	}
	// Cover materializeHost's invalid-chars and empty-pattern branches.
	if _, reason := materializeHost("", "userA", testPlatformDomain); reason != rDropEmptyPatterns {
		t.Fatalf("empty pattern reason want %q got %q", rDropEmptyPatterns, reason)
	}
	if _, reason := materializeHost("bad_under.example.com", "userA", testPlatformDomain); reason != rDropInvalidChars {
		t.Fatalf("invalid chars reason want %q got %q", rDropInvalidChars, reason)
	}
	if _, reason := materializeHost("x.example.com", "", testPlatformDomain); reason != rDropOwnerUnresolved {
		t.Fatalf("blank viewer reason want %q got %q", rDropOwnerUnresolved, reason)
	}
}

// TC-N6-22: third-party same-name CM (no managed-by label set) is skipped.
func TestSharedHosts_TC22_SkippedUnmanaged(t *testing.T) {
	resetSharedHostsMetrics()
	cm := newPlaceholderCM(testCallerNS)
	cm.Labels = map[string]string{constants.D2SharedHostsManagedByLabel: "third-party"} // not us
	cm.Data["custom"] = "do not touch"
	c := newSharedHostsClient(t,
		newOptInNS(testCallerNS, true),
		cm,
		newCallerApp("litellm", testCallerNS, "ollama"),
		newClusterApp("ollama", "userA"),
		newSRR("user-space-userA", "ollama-srr", "a1b2c3d4.*.example.com"),
	)
	runReconcile(t, c, testCallerNS)
	after := getCM(t, c, testCallerNS)
	if after.Data["custom"] != "do not touch" {
		t.Fatalf("third-party CM clobbered: %v", after.Data)
	}
	if v := testutil.ToFloat64(sharedHostsReconcileTotal.WithLabelValues(rResSkippedUnmanaged)); v != 1 {
		t.Fatalf("reconcile_total{skipped_unmanaged} want 1 got %v", v)
	}
}

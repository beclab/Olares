package appstate

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newCleanupManagerFixture(t *testing.T, ns string) (*appcfg.ApplicationConfig, []byte) {
	t.Helper()
	cfg := &appcfg.ApplicationConfig{
		AppName:   "demoapp",
		Namespace: ns,
		OwnerName: "alice",
	}
	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal app config: %v", err)
	}
	return cfg, cfgBytes
}

// Happy path: helm Uninstall succeeds, namespace is gone before cleanup runs,
// so the NS poll short-circuits on its first-shot check.
func TestCleanupAfterInstallFailure_NamespaceAlreadyGone(t *testing.T) {
	_, cfgJSON := newCleanupManagerFixture(t, "demoapp-ns")
	am := testutil.NewAppManager("demoapp",
		testutil.WithNamespace("demoapp-ns"),
		testutil.WithConfigJSON(string(cfgJSON)),
	)
	c := testutil.NewFakeClient(am)

	fake := testutil.NewFakeHelmOps()
	injectHelmOps(t, fake)

	if err := cleanupAfterInstallFailure(context.Background(), c, am); err != nil {
		t.Fatalf("cleanupAfterInstallFailure returned %v, want nil", err)
	}
	if fake.CallCount("Uninstall") != 1 {
		t.Errorf("Uninstall calls=%d, want 1", fake.CallCount("Uninstall"))
	}
}

// Cleanup is idempotent: a helm Uninstall that returns ErrReleaseNotFound
// must NOT prevent the cleanup from declaring success. We surface this by
// having FakeHelmOps return a sentinel error other than ErrReleaseNotFound to
// confirm the helper swallows it as a warning and still proceeds to the NS
// poll. The fake namespace is absent, so the helper still returns nil.
func TestCleanupAfterInstallFailure_HelmErrorIsNonFatal(t *testing.T) {
	_, cfgJSON := newCleanupManagerFixture(t, "demoapp-ns")
	am := testutil.NewAppManager("demoapp",
		testutil.WithNamespace("demoapp-ns"),
		testutil.WithConfigJSON(string(cfgJSON)),
	)
	c := testutil.NewFakeClient(am)

	fake := testutil.NewFakeHelmOps()
	fake.UninstallErr = errors.New("transient helm error")
	injectHelmOps(t, fake)

	if err := cleanupAfterInstallFailure(context.Background(), c, am); err != nil {
		t.Fatalf("cleanupAfterInstallFailure should swallow helm errors, got %v", err)
	}
	if fake.CallCount("Uninstall") != 1 {
		t.Errorf("Uninstall calls=%d, want 1", fake.CallCount("Uninstall"))
	}
}

// Empty Spec.Config (failure before unmarshal) must not crash: we skip the
// helm path entirely and the NS poll short-circuits if AppNamespace is also
// missing.
func TestCleanupAfterInstallFailure_EmptyConfigSkipsHelm(t *testing.T) {
	am := testutil.NewAppManager("demoapp", testutil.WithNamespace(""))
	c := testutil.NewFakeClient(am)

	fake := testutil.NewFakeHelmOps()
	injectHelmOps(t, fake)

	if err := cleanupAfterInstallFailure(context.Background(), c, am); err != nil {
		t.Fatalf("cleanupAfterInstallFailure returned %v, want nil", err)
	}
	if fake.CallCount("Uninstall") != 0 {
		t.Errorf("Uninstall must not be called when Spec.Config is empty, got %d", fake.CallCount("Uninstall"))
	}
}

// Protected namespaces (user-space, os-system, ...) are never deleted by the
// app lifecycle, so the helper must NOT block on them — the poll has to
// short-circuit on IsProtectedNamespace, not wait for the timeout.
func TestCleanupAfterInstallFailure_ProtectedNamespaceShortCircuits(t *testing.T) {
	_, cfgJSON := newCleanupManagerFixture(t, "user-space-alice")
	am := testutil.NewAppManager("demoapp",
		testutil.WithNamespace("user-space-alice"),
		testutil.WithConfigJSON(string(cfgJSON)),
	)
	// Even though the namespace very much exists, the helper must return
	// nil immediately.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-space-alice"}}
	c := testutil.NewFakeClient(am, ns)
	fake := testutil.NewFakeHelmOps()
	injectHelmOps(t, fake)

	done := make(chan error, 1)
	go func() { done <- cleanupAfterInstallFailure(context.Background(), c, am) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("cleanupAfterInstallFailure returned %v, want nil for protected NS", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("cleanupAfterInstallFailure did not short-circuit for protected NS")
	}
}

// When the namespace is still present and never gets deleted, the helper must
// return context.DeadlineExceeded (or Canceled) instead of blocking forever.
// We use a short ctx deadline to keep the test fast; the real
// installFailureNSDeletionTimeout (5min) is enforced internally with
// context.WithTimeout, but the OUTER ctx deadline can cut it short.
//
// The fake namespace carries the standard "kubernetes" finalizer so that
// cleanupAfterInstallFailure's c.Delete call only marks it for deletion
// (DeletionTimestamp gets set) but does not physically remove the object
// from the fake store — mirroring real K8s NS finalizer behavior and
// keeping the poll loop spinning until the outer ctx deadline fires.
func TestCleanupAfterInstallFailure_NamespaceStillPresentBlocksUntilCtx(t *testing.T) {
	_, cfgJSON := newCleanupManagerFixture(t, "demoapp-ns")
	am := testutil.NewAppManager("demoapp",
		testutil.WithNamespace("demoapp-ns"),
		testutil.WithConfigJSON(string(cfgJSON)),
	)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "demoapp-ns",
			Finalizers: []string{"kubernetes"},
		},
	}
	c := testutil.NewFakeClient(am, ns)
	fake := testutil.NewFakeHelmOps()
	injectHelmOps(t, fake)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	err := cleanupAfterInstallFailure(ctx, c, am)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("cleanupAfterInstallFailure returned nil, want context deadline error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("cleanupAfterInstallFailure returned %v, want DeadlineExceeded/Canceled", err)
	}
	if elapsed > 4*time.Second {
		t.Fatalf("cleanupAfterInstallFailure blocked for %s, expected ~2s ctx deadline", elapsed)
	}
}

// waitForNamespaceGone must short-circuit on its first-shot Get before the
// 1-second ticker fires, so a test that asserts the helper returns in
// well under 1 second proves we don't pay an unnecessary tick on the fast path.
func TestWaitForNamespaceGone_FastPathNoTick(t *testing.T) {
	c := testutil.NewFakeClient()

	start := time.Now()
	if err := waitForNamespaceGone(context.Background(), c, "missing-ns"); err != nil {
		t.Fatalf("waitForNamespaceGone returned %v, want nil", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("waitForNamespaceGone first-shot path took %s, expected < 500ms", elapsed)
	}
}

// waitForNamespaceGone returns when the namespace is deleted mid-poll. Drive
// this by deleting the NS from another goroutine after the helper has had a
// chance to take at least one tick.
func TestWaitForNamespaceGone_ResolvesAfterDelete(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demoapp-ns"}}
	c := testutil.NewFakeClient(ns)

	go func() {
		time.Sleep(1500 * time.Millisecond)
		_ = c.Delete(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demoapp-ns"}})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := waitForNamespaceGone(ctx, c, "demoapp-ns"); err != nil {
		t.Fatalf("waitForNamespaceGone returned %v, want nil after delete", err)
	}
	// Sanity check: the NS really is gone in the fake store.
	var got corev1.Namespace
	if err := c.Get(context.Background(), types.NamespacedName{Name: "demoapp-ns"}, &got); !apierrors.IsNotFound(err) {
		t.Fatalf("expected IsNotFound, got %v", err)
	}
}

// Compile-time sanity that the helper file uses the right client type. This
// keeps the test from drifting if someone changes the helper signature.
var _ = func(ctx context.Context, c client.Client) {
	_ = waitForNamespaceGone(ctx, c, "")
}

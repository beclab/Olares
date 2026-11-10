package helmops_test

import (
	"context"
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// -----------------------------------------------------------------
// WaitForNamespaceDeleted — the only method with non-trivial
// scheduling logic (poll loop with deadline). Every branch of the
// state machine is exercised.
// -----------------------------------------------------------------

func TestWaitForNamespaceDeleted_EmptyRequest_ReturnsImmediately(t *testing.T) {
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{})
	if res.Err != nil {
		t.Fatalf("empty request should succeed: %v", res.Err)
	}
}

func TestWaitForNamespaceDeleted_ProtectedNamespace_Skipped(t *testing.T) {
	// A protected namespace (user-space- prefix) must NOT block —
	// the adapter filters it out even if the caller passes it.
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary: "user-space-bfl-admin",
	})
	if res.Err != nil {
		t.Fatalf("protected namespace should be skipped: %v", res.Err)
	}
}

func TestWaitForNamespaceDeleted_DeletesExistingThenSucceeds(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	c := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ns).
		Build()
	o := helmops.New(c, &rest.Config{})

	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary: "user-x-app-x",
		Timeout: time.Second,
	})
	if res.Err != nil {
		t.Fatalf("existing namespace should be deleted then succeed: %v", res.Err)
	}

	got := &corev1.Namespace{}
	err := c.Get(context.Background(), client.ObjectKey{Name: "user-x-app-x"}, got)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("namespace still present after wait: err=%v", err)
	}
}

func TestWaitForNamespaceDeleted_DeleteErrorPropagates(t *testing.T) {
	wantErr := errors.New("apiserver refuse delete")
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	base := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ns).
		Build()
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return wantErr
		},
	})
	o := helmops.New(c, &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary: "user-x-app-x",
		Timeout: 50 * time.Millisecond,
	})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestWaitForNamespaceDeleted_DoesNotDeleteProtected(t *testing.T) {
	var deletes []string
	base := fakeclient.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.DeleteOption) error {
			deletes = append(deletes, obj.GetName())
			return nil
		},
	})
	o := helmops.New(c, &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary:  "user-space-bfl-admin",
		SharedNS: []string{"kube-system"},
	})
	if res.Err != nil {
		t.Fatalf("protected namespace should be skipped: %v", res.Err)
	}
	if len(deletes) != 0 {
		t.Fatalf("Delete called for protected namespaces: %v", deletes)
	}
}

func TestWaitForNamespaceDeleted_AlreadyGone_FastPath(t *testing.T) {
	// No preseeded namespace → immediate probe returns NotFound
	// → no tick channel wait, no clock advance needed.
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary: "user-x-app-x",
		Timeout: time.Second,
	})
	if res.Err != nil {
		t.Fatalf("gone namespace should succeed immediately: %v", res.Err)
	}
}

func TestWaitForNamespaceDeleted_PresentThenGone_TickAndSucceed(t *testing.T) {
	// Seed the namespace; interceptor removes it after the first
	// Get so the second probe sees NotFound.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	base := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ns).
		Build()

	var getCount int
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return nil
		},
		Get: func(ctx context.Context, wc client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			getCount++
			if getCount == 1 {
				// First probe: namespace still exists.
				return wc.Get(ctx, key, obj, opts...)
			}
			// Subsequent probes: namespace gone.
			return apierrors.NewNotFound(corev1.Resource("namespaces"), key.Name)
		},
	})

	fc := newFakeClock()
	o := helmops.New(c, &rest.Config{},
		helmops.WithClock(fc),
		helmops.WithPollInterval(50*time.Millisecond),
		helmops.WithNamespaceWaitTimeout(time.Second),
	)

	done := make(chan ports.WorkloadResult, 1)
	go func() {
		done <- o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
			Primary: "user-x-app-x",
		})
	}()

	// The first probe (before entering the loop) sees "still
	// exists". Fire the tick to release the After wait, which
	// runs the second probe → NotFound → success.
	fc.Fire(50 * time.Millisecond)
	select {
	case res := <-done:
		if res.Err != nil {
			t.Fatalf("expected success after second probe: %v", res.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForNamespaceDeleted did not return within 2s")
	}
	if getCount < 2 {
		t.Errorf("expected at least 2 Get calls, got %d", getCount)
	}
}

func TestWaitForNamespaceDeleted_NonNotFoundGetError_Propagates(t *testing.T) {
	wantErr := errors.New("apiserver timeout")
	base := fakeclient.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()
	c := interceptor.NewClient(base, interceptor.Funcs{
		Get: func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
			return wantErr
		},
	})
	o := helmops.New(c, &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary: "user-x-app-x",
	})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestWaitForNamespaceDeleted_Timeout_Errors(t *testing.T) {
	// Uses REAL time (no WithClock override): the fake clock
	// requires precise synchronisation between the deadline
	// capture and the tick advance which is inherently racy
	// against a background goroutine. A short real timeout
	// (50ms) with a short poll (10ms) tests the timeout path
	// deterministically without introducing test-side barriers
	// that would bleed into the adapter's public surface.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	base := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ns).
		Build()
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return nil
		},
	})

	o := helmops.New(c, &rest.Config{},
		helmops.WithPollInterval(10*time.Millisecond),
	)

	done := make(chan ports.WorkloadResult, 1)
	go func() {
		done <- o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
			Primary: "user-x-app-x",
			Timeout: 50 * time.Millisecond,
		})
	}()

	select {
	case res := <-done:
		if res.Err == nil {
			t.Fatal("expected timeout error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForNamespaceDeleted did not return within 2s")
	}
}

func TestWaitForNamespaceDeleted_ContextCanceled(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	base := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ns).
		Build()
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return nil
		},
	})

	fc := newFakeClock()
	o := helmops.New(c, &rest.Config{},
		helmops.WithClock(fc),
		helmops.WithPollInterval(50*time.Millisecond),
		helmops.WithNamespaceWaitTimeout(10*time.Second),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan ports.WorkloadResult, 1)
	go func() {
		done <- o.WaitForNamespaceDeleted(ctx, ports.NamespaceWaitRequest{
			Primary: "user-x-app-x",
		})
	}()

	// Give the goroutine a moment to enter the select, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case res := <-done:
		if !errors.Is(res.Err, context.Canceled) {
			t.Errorf("err: %v want context.Canceled", res.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForNamespaceDeleted did not return after context cancel")
	}
}

func TestWaitForNamespaceDeleted_SharedNS_AllMustGo(t *testing.T) {
	// Two namespaces preseeded: primary and one shared. Primary
	// is removed on second probe; shared removed on third probe.
	// Result: three tick cycles before success.
	nsPrimary := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "user-x-app-x"}}
	nsShared := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "shared-app-x"}}
	base := fakeclient.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(nsPrimary, nsShared).
		Build()

	// getCounts[name] tracks per-target probes. probeNamespaces
	// short-circuits on the FIRST target that still exists, so
	// the per-target getCount does not increment uniformly:
	//
	//   Immediate probe: get(primary) = 1 → exists → stops.
	//   Tick 1: get(primary) = 2 → NotFound. get(shared) = 1 → exists → stops.
	//   Tick 2: get(primary) = 3 → NotFound. get(shared) = 2 → NotFound → success.
	//
	// So primary disappears when its getCount >= 2, shared
	// when its getCount >= 2. Two Fires suffice.
	getCounts := map[string]int{}
	c := interceptor.NewClient(base, interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return nil
		},
		Get: func(ctx context.Context, wc client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			getCounts[key.Name]++
			if key.Name == "user-x-app-x" && getCounts[key.Name] >= 2 {
				return apierrors.NewNotFound(corev1.Resource("namespaces"), key.Name)
			}
			if key.Name == "shared-app-x" && getCounts[key.Name] >= 2 {
				return apierrors.NewNotFound(corev1.Resource("namespaces"), key.Name)
			}
			return wc.Get(ctx, key, obj, opts...)
		},
	})

	fc := newFakeClock()
	o := helmops.New(c, &rest.Config{},
		helmops.WithClock(fc),
		helmops.WithPollInterval(10*time.Millisecond),
		helmops.WithNamespaceWaitTimeout(5*time.Second),
	)

	done := make(chan ports.WorkloadResult, 1)
	go func() {
		done <- o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
			Primary:  "user-x-app-x",
			SharedNS: []string{"shared-app-x"},
		})
	}()

	// Immediate probe: primary exists (get=1) → into tick loop.
	// Tick 1: primary NotFound (get=2), shared exists (get=1) → loop again.
	fc.Fire(10 * time.Millisecond)
	// Tick 2: primary NotFound (get=3), shared NotFound (get=2) → success.
	fc.Fire(10 * time.Millisecond)

	select {
	case res := <-done:
		if res.Err != nil {
			t.Fatalf("expected success once all namespaces gone: %v", res.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForNamespaceDeleted did not return within 2s")
	}
}

func TestWaitForNamespaceDeleted_SharedNS_ProtectedSkipped(t *testing.T) {
	// Every shared namespace is protected → adapter should treat
	// the list as effectively empty and short-circuit to success.
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{})
	res := o.WaitForNamespaceDeleted(context.Background(), ports.NamespaceWaitRequest{
		Primary:  "user-space-bfl-admin",
		SharedNS: []string{"kube-system", "kubesphere-system"},
	})
	if res.Err != nil {
		t.Fatalf("all-protected targets should succeed: %v", res.Err)
	}
}

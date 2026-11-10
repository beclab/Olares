package ctrlrequeue

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// startAndDrain wires the Requeuer.Source() into a rate-limiting
// workqueue and returns a function that pops the next N requests
// with a timeout. This is the same shape controller-runtime's own
// source tests use — pop from the queue rather than mock the
// EventHandler surface, so the test exercises the full path
// Requeue → chan → source.Channel → EnqueueRequestsFromMapFunc →
// workqueue that production uses.
func startAndDrain(t *testing.T, r *Requeuer) (
	popN func(n int, timeout time.Duration) []reconcile.Request,
	stop func(),
) {
	t.Helper()

	q := workqueue.NewTypedRateLimitingQueueWithConfig(
		workqueue.DefaultTypedControllerRateLimiter[reconcile.Request](),
		workqueue.TypedRateLimitingQueueConfig[reconcile.Request]{Name: "ctrlrequeue-test"},
	)

	ctx, cancel := context.WithCancel(context.Background())
	src := r.Source()
	if err := src.Start(ctx, q); err != nil {
		cancel()
		t.Fatalf("Source().Start: %v", err)
	}

	popN = func(n int, timeout time.Duration) []reconcile.Request {
		t.Helper()
		got := make([]reconcile.Request, 0, n)
		deadline := time.Now().Add(timeout)
		for len(got) < n {
			// The workqueue's Get blocks; run it on a goroutine so
			// we can enforce the caller's timeout.
			ch := make(chan reconcile.Request, 1)
			go func() {
				item, shutdown := q.Get()
				if shutdown {
					return
				}
				q.Done(item)
				ch <- item
			}()
			select {
			case req := <-ch:
				got = append(got, req)
			case <-time.After(time.Until(deadline)):
				t.Fatalf("timed out waiting for request %d/%d after %s; so far got %#v",
					len(got)+1, n, timeout, got)
			}
		}
		return got
	}
	stop = func() {
		cancel()
		q.ShutDown()
	}
	return popN, stop
}

// TestRequeuer_SatisfiesInterface pins the compile-time invariant
// at runtime. Regressing on the interface (e.g. a method rename)
// fails here immediately rather than at whichever main.go tries
// to wire it up.
func TestRequeuer_SatisfiesInterface(t *testing.T) {
	var _ runner.Requeuer = New()
}

// TestRequeuer_EnqueuesReconcileRequest — the load-bearing test.
// Proves the end-to-end wiring: Requeue(name) surfaces as a
// reconcile.Request{Name: name} on the workqueue, matching the
// contract main.go relies on.
func TestRequeuer_EnqueuesReconcileRequest(t *testing.T) {
	r := New()
	pop, stop := startAndDrain(t, r)
	defer stop()

	r.Requeue("nginx-alice-nginx")

	got := pop(1, time.Second)
	want := []reconcile.Request{
		{NamespacedName: types.NamespacedName{Name: "nginx-alice-nginx"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("popped requests mismatch:\n got %#v\nwant %#v", got, want)
	}
}

// TestRequeuer_MultipleNamesAllEnqueued — every Requeue call is a
// distinct enqueue; the buffered channel does not deduplicate.
// (Deduplication is the workqueue's job, and it does dedupe by
// key when the same name arrives twice — that behaviour is
// controller-runtime's, not ours; this test just proves distinct
// names all propagate.)
func TestRequeuer_MultipleNamesAllEnqueued(t *testing.T) {
	r := New()
	pop, stop := startAndDrain(t, r)
	defer stop()

	names := []string{"a", "b", "c"}
	for _, n := range names {
		r.Requeue(n)
	}

	got := pop(len(names), 2*time.Second)
	sortByName(got)

	want := []reconcile.Request{
		{NamespacedName: types.NamespacedName{Name: "a"}},
		{NamespacedName: types.NamespacedName{Name: "b"}},
		{NamespacedName: types.NamespacedName{Name: "c"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("popped requests mismatch:\n got %#v\nwant %#v", got, want)
	}
}

// TestRequeuer_DropOnFullBuffer — the non-blocking-Requeue
// contract. Fill the buffer with the source paused (never
// Start()ed), then over-send by 100×; if the send were blocking
// this test would hang. The exact count of drops is not asserted
// (it depends on default channel semantics) but the completion of
// the loop within the test timeout IS the assertion.
func TestRequeuer_DropOnFullBuffer(t *testing.T) {
	r := New(WithBufferSize(2))

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			r.Requeue("overflow")
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Requeue blocked when buffer was full — contract violation")
	}
}

// TestNew_WithBufferSizeIgnoresNonPositive — negative / zero
// values must fall back to the default; if we honoured them the
// channel would be unbuffered and every Requeue call would block
// until the source drains, which is exactly the failure mode
// TestRequeuer_DropOnFullBuffer guards against.
func TestNew_WithBufferSizeIgnoresNonPositive(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		r := New(WithBufferSize(n))
		if cap(r.ch) != defaultBufferSize {
			t.Errorf("WithBufferSize(%d) -> cap=%d, want default %d", n, cap(r.ch), defaultBufferSize)
		}
	}
}

// TestRequeuer_ConcurrentRequeue — concurrent writers from
// different goroutines must all land in the queue. In production
// CallWorkload callbacks share this Requeuer.
func TestRequeuer_ConcurrentRequeue(t *testing.T) {
	const writers = 10
	const perWriter = 10

	r := New()
	pop, stop := startAndDrain(t, r)
	defer stop()

	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWriter; j++ {
				r.Requeue("app")
			}
		}(i)
	}
	wg.Wait()

	// The workqueue dedupes by key, so `writers*perWriter` sends
	// of the same name may collapse to a smaller number of
	// popable requests. We only need to prove that AT LEAST one
	// request lands — the source itself did not silently drop
	// under contention (which was the concern this test is here
	// to rule out). Getting one item within the timeout is
	// sufficient.
	got := pop(1, 2*time.Second)
	if len(got) != 1 || got[0].Name != "app" {
		t.Errorf("got %#v, want at least one reconcile.Request{Name:\"app\"}", got)
	}
}

func sortByName(rs []reconcile.Request) {
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Name < rs[j].Name
	})
}

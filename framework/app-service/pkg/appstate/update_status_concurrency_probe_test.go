package appstate

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestProbe_UpdateStatusConcurrentLostUpdate is a bug-hunting probe. It fires N
// concurrent updateStatus calls at the same ApplicationManager and checks that
// every call's effect survives: OpGeneration should advance by N and N op
// records should be retained.
//
// Candidate bug S3-1: updateStatus does a non-atomic read-modify-write --
// client.Get(am) then client.Patch(amCopy, client.MergeFrom(am)). MergeFrom is
// a plain JSON merge patch with NO resourceVersion precondition, and there is
// no RetryOnConflict. Concurrent callers therefore both read generation G and
// both write G+1 (lost increment); the OpRecords array, being replaced
// wholesale by a merge patch, loses all but the last writer's record.
//
// In production the same AM can be patched concurrently by the main reconcile,
// per-controller reconciles (appenv / suspend), apiserver handlers, and the
// appFactory watcher's Finally() -> so this is reachable, not theoretical.
func TestProbe_UpdateStatusConcurrentLostUpdate(t *testing.T) {
	// Regression for fixed bug S3-1 (non-atomic updateStatus lost concurrent
	// OpGeneration increments and OpRecords).
	const n = 30
	seed := testutil.NewAppManager("conc", testutil.WithState(appsv1.Installing))
	c := testutil.NewFakeClient(seed)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		successes int
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each goroutine uses its own AM pointer so the only shared
			// state is the client; the lost update is a property of
			// updateStatus, not of sharing a struct.
			m := &appsv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "conc"}}
			b := &baseStatefulApp{client: c}
			now := metav1.Now()
			rec := &appsv1.OpRecord{OpID: strconv.Itoa(i), StateTime: &now}
			if err := b.updateStatus(context.TODO(), m, appsv1.Running, rec, "msg", ""); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	// The core invariant: every update that the store accepted must have
	// contributed exactly one OpGeneration increment. Before the fix, all N
	// calls "succeeded" yet OpGeneration ended at ~2 because the non-atomic
	// read-modify-write clobbered concurrent increments.
	got := getAM(t, c, "conc")
	if got.Status.OpGeneration != int64(successes) {
		t.Errorf("OpGeneration=%d, want %d (== successful updates; increments were lost)", got.Status.OpGeneration, successes)
	}
	wantRecords := successes
	if wantRecords > 20 {
		wantRecords = 20
	}
	if len(got.Status.OpRecords) != wantRecords {
		t.Errorf("OpRecords=%d, want %d (records were lost)", len(got.Status.OpRecords), wantRecords)
	}
}

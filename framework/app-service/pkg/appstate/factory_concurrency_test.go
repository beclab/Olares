package appstate

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// fakeInProgress is a minimal StatefulInProgressApp used to exercise the
// factory's concurrency control. Cleanup closes done exactly once, mirroring the
// real apps where Cleanup cancels the context backing Done().
type fakeInProgress struct {
	*baseStatefulApp
	done     chan struct{}
	closeOne sync.Once
	cleaned  int32
}

func (f *fakeInProgress) IsTimeout() bool { return false }

func (f *fakeInProgress) Exec(ctx context.Context) (StatefulInProgressApp, error) { return f, nil }

func (f *fakeInProgress) Cancel(ctx context.Context) error { return nil }

func (f *fakeInProgress) Cleanup(ctx context.Context) {
	atomic.AddInt32(&f.cleaned, 1)
	f.closeOne.Do(func() { close(f.done) })
}

func (f *fakeInProgress) Done() <-chan struct{} { return f.done }

func newFakeInProgress(name string, state appsv1.ApplicationManagerState) *fakeInProgress {
	am := testutil.NewAppManager(name, testutil.WithState(state))
	c := testutil.NewFakeClient(am)
	return &fakeInProgress{
		baseStatefulApp: &baseStatefulApp{manager: am, client: c},
		done:            make(chan struct{}),
	}
}

// Concurrent execAndWatch calls for the same app must execute the operation only
// once; the rest must observe the already in-progress app.
func TestExecAndWatchDedupsConcurrent(t *testing.T) {
	app := newFakeInProgress("dedup-app", appsv1.Installing)

	var execCount int32
	exec := func(ctx context.Context) (StatefulInProgressApp, error) {
		atomic.AddInt32(&execCount, 1)
		return app, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := appFactory.execAndWatch(context.Background(), app, exec); err != nil {
				t.Errorf("execAndWatch: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&execCount); got != 1 {
		t.Fatalf("exec called %d times, want exactly 1", got)
	}
	if got := appFactory.countInProgressApp(appsv1.Installing.String()); got != 1 {
		t.Fatalf("in-progress count=%d, want 1", got)
	}

	// Releasing the app removes it from the registry.
	if !appFactory.cancelOperation("dedup-app") {
		t.Fatal("cancelOperation returned false for registered app")
	}
	testutil.Eventually(t, time.Second, 10*time.Millisecond, func() bool {
		return appFactory.countInProgressApp(appsv1.Installing.String()) == 0
	})
}

// Concurrent cancelOperation calls for the same app must report success exactly
// once and clean up the app exactly once.
func TestCancelOperationConcurrent(t *testing.T) {
	app := newFakeInProgress("cancel-app", appsv1.Upgrading)

	if _, err := appFactory.execAndWatch(context.Background(), app, func(ctx context.Context) (StatefulInProgressApp, error) {
		return app, nil
	}); err != nil {
		t.Fatalf("register app: %v", err)
	}

	var trueCount int32
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if appFactory.cancelOperation("cancel-app") {
				atomic.AddInt32(&trueCount, 1)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&trueCount); got != 1 {
		t.Fatalf("cancelOperation returned true %d times, want exactly 1", got)
	}
	if got := atomic.LoadInt32(&app.cleaned); got != 1 {
		t.Fatalf("Cleanup called %d times, want exactly 1", got)
	}
	if got := appFactory.countInProgressApp(appsv1.Upgrading.String()); got != 0 {
		t.Fatalf("in-progress count=%d after cancel, want 0", got)
	}
}

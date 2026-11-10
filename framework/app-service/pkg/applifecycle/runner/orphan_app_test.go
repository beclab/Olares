package runner_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestReconcile_OrphanApplication_ForceDeletesNamespaceAndEmitsEvents(t *testing.T) {
	h := newHarness(t)
	h.amStore.AddApp(&appv1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "demo",
			Annotations: map[string]string{constants.AppMarketSourceKey: "market"},
		},
		Spec: appv1.ApplicationSpec{
			Name:       "demo",
			RawAppName: "demo-raw",
			Namespace:  "demo-alice",
			Owner:      "alice",
			Settings:   map[string]string{"title": "Demo"},
		},
	})

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "demo"})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if out.Requeue || out.RequeueAfter != 0 {
		t.Fatalf("outcome = %+v, want zero", out)
	}

	waitUntil(t, func() bool {
		return len(h.recorder.Events()) >= 2 && len(namespaceWaits(h)) >= 1
	})

	calls := namespaceWaits(h)
	if calls[0].Primary != "demo-alice" {
		t.Fatalf("namespace wait primary = %q, want demo-alice", calls[0].Primary)
	}
	if calls[0].Timeout != 30*time.Minute {
		t.Fatalf("namespace wait timeout = %s, want 30m", calls[0].Timeout)
	}

	events := h.recorder.Events()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	opID := strconv.FormatInt(h.clock.Now().Unix(), 10)
	want := []ports.AppLifecycleEvent{
		{
			Name: "demo", RawAppName: "demo-raw", Type: "app",
			OpType: string(appv1.UninstallOp), OpID: opID,
			State: string(appv1.Uninstalling), User: "alice", Title: "Demo",
			Reason:       constants.AppForceUninstall,
			Message:      "force delete application demo successfully",
			MarketSource: "market",
		},
		{
			Name: "demo", RawAppName: "demo-raw", Type: "app",
			OpType: string(appv1.UninstallOp), OpID: opID,
			State: string(appv1.Uninstalled), User: "alice", Title: "Demo",
			Reason:       constants.AppForceUninstalled,
			Message:      "app demo was force uninstalled",
			MarketSource: "market",
		},
	}
	for i, ev := range want {
		got := events[i]
		if got.Name != ev.Name || got.RawAppName != ev.RawAppName || got.Type != ev.Type ||
			got.OpType != ev.OpType || got.OpID != ev.OpID || got.State != ev.State ||
			got.User != ev.User || got.Title != ev.Title || got.Reason != ev.Reason ||
			got.Message != ev.Message || got.MarketSource != ev.MarketSource {
			t.Fatalf("event[%d] = %+v, want %+v", i, got, ev)
		}
		if got.EventID == "" || !got.CreateTime.Equal(h.clock.Now()) {
			t.Fatalf("event[%d] id/time = %q/%s", i, got.EventID, got.CreateTime)
		}
	}
}

func TestReconcile_OrphanMiddleware_DeletesNamespaceWithoutApplication(t *testing.T) {
	h := newHarness(t)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "citus-middleware-citus"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	waitUntil(t, func() bool { return len(namespaceWaits(h)) >= 1 })

	calls := namespaceWaits(h)
	if calls[0].Primary != "citus-middleware" {
		t.Fatalf("namespace wait primary = %q, want citus-middleware", calls[0].Primary)
	}
	if events := h.recorder.Events(); len(events) != 0 {
		t.Fatalf("events = %d, want 0 without an Application", len(events))
	}
}

func TestReconcile_OrphanWithoutApplication_DoesNotDeleteNamespace(t *testing.T) {
	h := newHarness(t)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "demo"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if calls := namespaceWaits(h); len(calls) != 0 {
		t.Fatalf("namespace waits = %d, want 0", len(calls))
	}
}

func TestReconcile_OrphanApplication_ProtectedNamespaceSkipped(t *testing.T) {
	h := newHarness(t)
	h.amStore.AddApp(&appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "sys"},
		Spec: appv1.ApplicationSpec{
			Name: "sys", Namespace: "user-space-alice", Owner: "alice",
		},
	})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "sys"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if calls := namespaceWaits(h); len(calls) != 0 {
		t.Fatalf("namespace waits = %d, want 0", len(calls))
	}
	if events := h.recorder.Events(); len(events) != 0 {
		t.Fatalf("events = %d, want 0", len(events))
	}
}

func TestReconcile_OrphanApplication_DeleteFailureDoesNotEmitUninstalled(t *testing.T) {
	h := newHarness(t)
	h.amStore.AddApp(&appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
		Spec: appv1.ApplicationSpec{
			Name: "demo", Namespace: "demo-alice", Owner: "alice",
		},
	})
	h.workloadOps.WaitForNamespaceDeletedFunc = func(context.Context, ports.NamespaceWaitRequest) ports.WorkloadResult {
		return ports.WorkloadResult{Err: errors.New("namespace busy")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "demo"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	waitUntil(t, func() bool { return len(h.recorder.Events()) >= 2 })

	events := h.recorder.Events()
	for _, ev := range events {
		if ev.State == string(appv1.Uninstalled) {
			t.Fatalf("emitted Uninstalled after delete failure: %+v", ev)
		}
	}
	last := events[len(events)-1]
	if last.State != string(appv1.Uninstalling) || last.Reason != constants.AppForceUninstall {
		t.Fatalf("last event = %+v, want Uninstalling/ForceUninstall", last)
	}
	if last.Message != "force delete application demo failed: namespace busy" {
		t.Fatalf("failure message = %q", last.Message)
	}
}

func TestReconcile_OrphanApplication_DoesNotStartTwice(t *testing.T) {
	h := newHarness(t)
	h.amStore.AddApp(&appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
		Spec: appv1.ApplicationSpec{
			Name: "demo", Namespace: "demo-alice", Owner: "alice",
		},
	})
	started := make(chan struct{})
	var once sync.Once
	h.workloadOps.WaitForNamespaceDeletedFunc = func(ctx context.Context, _ ports.NamespaceWaitRequest) ports.WorkloadResult {
		once.Do(func() { close(started) })
		<-ctx.Done()
		return ports.WorkloadResult{Err: ctx.Err()}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "demo"}); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("namespace wait did not start")
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "demo"}); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	if calls := namespaceWaits(h); len(calls) != 1 {
		t.Fatalf("namespace waits = %d, want 1", len(calls))
	}
}

func namespaceWaits(h *harness) []ports.NamespaceWaitRequest {
	var out []ports.NamespaceWaitRequest
	for _, call := range h.workloadOps.Calls() {
		if call.Method == "WaitForNamespaceDeleted" && call.NSReq != nil {
			out = append(out, *call.NSReq)
		}
	}
	return out
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}

package runner_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

const durableRequestID = "request-123"

func seedDurableRequest(h *harness, name string, state appv1.ApplicationManagerState, op appv1.OpType, requestID string) {
	h.t.Helper()
	h.seedAM(name, string(state))
	am, err := h.amStore.Get(context.Background(), name)
	if err != nil {
		h.t.Fatalf("get seeded AM: %v", err)
	}
	am.Spec.OpType = op
	if am.Annotations == nil {
		am.Annotations = map[string]string{}
	}
	am.Annotations[domain.AnnotationOperationRequestID] = requestID
	if err := h.amStore.UpdateFull(context.Background(), am); err != nil {
		h.t.Fatalf("persist durable request: %v", err)
	}
}

func TestReconcile_RecoversPersistedOperationWithFreshQueue(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Running, appv1.Stopping)
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.TriggerOp{Op: appv1.StopOp},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Next: appv1.Stopping,
				Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Stopping,
				}}},
			}
		},
	})
	seedDurableRequest(h, "app1", appv1.Running, appv1.StopOp, durableRequestID)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("state = %q, want %q", got.Status.State, appv1.Stopping)
	}
	if got.Status.OpType != appv1.StopOp || got.Status.OpID != durableRequestID {
		t.Fatalf("operation identity = (%q, %q), want (%q, %q)",
			got.Status.OpType, got.Status.OpID, appv1.StopOp, durableRequestID)
	}
}

func TestReconcile_DoesNotReplayConsumedDurableRequest(t *testing.T) {
	h := newHarness(t)
	calls := 0
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.TriggerOp{Op: appv1.StopOp},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			calls++
			return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				Message: "triggered",
			}}}}
		},
	})
	seedDurableRequest(h, "app1", appv1.Running, appv1.StopOp, durableRequestID)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	if calls != 1 {
		t.Fatalf("TriggerOp handler calls = %d, want 1", calls)
	}
}

func TestReconcile_DisallowedDurableRequestFallsBackToTick(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				Reason: "TickWon",
			}}}}
		},
	})
	seedDurableRequest(h, "app1", appv1.Running, appv1.StopOp, durableRequestID)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.Reason != "TickWon" {
		t.Fatalf("reason = %q, want TickWon", got.Status.Reason)
	}
	if got.Status.OpID != "" {
		t.Fatalf("Status.OpID = %q, want empty", got.Status.OpID)
	}
}

func TestReconcile_QueuedActionWinsOverDurableRequest(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				Reason: "QueuedTickWon",
			}}}}
		},
	})
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.TriggerOp{Op: appv1.StopOp},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				Reason: "DurableTriggerWon",
			}}}}
		},
	})
	seedDurableRequest(h, "app1", appv1.Running, appv1.StopOp, durableRequestID)
	h.runner.Pending().Deliver("app1", domain.Tick{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.Reason != "QueuedTickWon" {
		t.Fatalf("reason = %q, want QueuedTickWon", got.Status.Reason)
	}
	if got.Status.OpID != "" {
		t.Fatalf("Status.OpID = %q, want empty", got.Status.OpID)
	}
}

func TestReconcile_TriggerOpStampsSynthesizedPatch(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Running, appv1.Stopping)
	h.table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.TriggerOp{Op: appv1.StopOp},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Stopping, Message: "stopping"}
		},
	})
	seedDurableRequest(h, "app1", appv1.Running, appv1.StopOp, durableRequestID)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.OpType != appv1.StopOp || got.Status.OpID != durableRequestID {
		t.Fatalf("operation identity = (%q, %q), want (%q, %q)",
			got.Status.OpType, got.Status.OpID, appv1.StopOp, durableRequestID)
	}
}

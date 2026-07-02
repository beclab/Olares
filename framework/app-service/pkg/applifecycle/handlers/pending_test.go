package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newPendingAM(name string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.InstallOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:  appv1.Pending,
			OpType: appv1.InstallOp,
			OpID:   "op-42",
		},
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnInitialInstall_PublishesWaitingForInstall pins the fresh
// install bootstrap. The pre-refactor HTTP handler wrote Reason,
// Message onto Pending, and the state-change informer
// published them. The placeholder only set State, so bfl never saw
// "waiting for install".
func TestOnInitialInstall_PublishesWaitingForInstall(t *testing.T) {
	am := newPendingAM("app1")
	am.Status = appv1.ApplicationManagerStatus{}
	res := handlers.OnInitialInstall(context.Background(), domain.World{AM: am})
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != appv1.Pending {
		t.Fatalf("Next = %q, want Pending", res.Next)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if patch.Status.State != appv1.Pending || patch.Status.Reason != appv1.Pending.String() ||
		patch.Status.Message != "waiting for install" {
		t.Fatalf("PatchAM.Status = %#v", patch.Status)
	}
	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok {
		t.Fatal("no EmitEvent effect")
	}
	if ev.Reason != appv1.Pending.String() || ev.Message != "waiting for install" {
		t.Fatalf("EmitEvent = %#v", ev)
	}
}

// TestReconcile_FreshInstall_EmitsWaitingForInstall runs the same
// edge through the Runner so the recorded event carries the
// operation id the status patch just wrote.
func TestReconcile_FreshInstall_EmitsWaitingForInstall(t *testing.T) {
	h := newE2EHarness(t)
	am := newPendingAM("app1")
	am.Status = appv1.ApplicationManagerStatus{}
	am.Annotations = map[string]string{domain.AnnotationOperationRequestID: "req-1"}
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, err := h.amStore.Get(context.Background(), "app1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status.State != appv1.Pending || got.Status.OpType != appv1.InstallOp || got.Status.OpID != "req-1" ||
		got.Status.Reason != appv1.Pending.String() || got.Status.Message != "waiting for install" {
		t.Fatalf("Status = %#v", got.Status)
	}
	events := h.recorder.Events()
	if len(events) != 1 {
		t.Fatalf("events = %#v, want one", events)
	}
	ev := events[0]
	if ev.State != string(appv1.Pending) || ev.OpType != string(appv1.InstallOp) || ev.OpID != "req-1" ||
		ev.Reason != appv1.Pending.String() || ev.Message != "waiting for install" {
		t.Fatalf("event = %#v", ev)
	}
}

// TestOnPendingTick_SlotFree_Advances pins the happy path: with
// zero other AMs Downloading, the handler emits a PatchAM and
// EmitEvent that flip the state to Downloading.
func TestOnPendingTick_SlotFree_Advances(t *testing.T) {
	am := newPendingAM("app1")
	w := domain.World{
		AM:        am,
		Now:       time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		AppExists: false,
		ClusterStats: domain.ClusterStats{
			DownloadingCount: 0,
			NextDownloading:  "app1",
		},
	}

	res := handlers.OnPendingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != appv1.Downloading {
		t.Fatalf("Next = %q, want Downloading", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if patch.Status.State != appv1.Downloading {
		t.Fatalf("PatchAM.State = %q, want Downloading", patch.Status.State)
	}
	if patch.Status.Message != "start to download" {
		t.Fatalf("PatchAM.Message = %q, want legacy string 'start to download'", patch.Status.Message)
	}
	if patch.Status.Reason != "Downloading" {
		t.Fatalf("PatchAM.Reason = %q, want 'Downloading'", patch.Status.Reason)
	}

	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok {
		t.Fatal("no EmitEvent effect")
	}
	if ev.Type != "Normal" {
		t.Fatalf("EmitEvent.Type = %q, want Normal", ev.Type)
	}

	// No ScheduleTick on the happy path — the AM is advancing;
	// controller-runtime will re-reconcile as the informer cache
	// picks up the state change.
	if _, ok := findEffect[domain.ScheduleTick](res.Effects); ok {
		t.Fatal("unexpected ScheduleTick on the advance branch")
	}
}

// TestOnPendingTick_SlotFree_NotHeadOfQueue_Waits pins FIFO: the
// download slot is empty, but an older Pending AM is the head of
// the queue, so this AM must not jump to Downloading.
func TestOnPendingTick_SlotFree_NotHeadOfQueue_Waits(t *testing.T) {
	am := newPendingAM("app-b")
	w := domain.World{
		AM:        am,
		AppExists: false,
		ClusterStats: domain.ClusterStats{
			DownloadingCount: 0,
			NextDownloading:  "app-a",
		},
	}

	res := handlers.OnPendingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (must not preempt queue head)", res.Next)
	}
	if _, ok := findEffect[domain.ScheduleTick](res.Effects); !ok {
		t.Fatal("no ScheduleTick on FIFO-wait branch")
	}
	if _, ok := findEffect[domain.EmitEvent](res.Effects); ok {
		t.Fatal("EmitEvent on FIFO-wait branch; expected none")
	}
}

// TestOnPendingTick_SlotOccupied_WaitsWithBackoff pins the "bed
// full" branch: another AM is currently Downloading, so we must
// NOT advance — instead emit a ScheduleTick with the documented
// backoff and a PatchAM noting the wait reason.
func TestOnPendingTick_SlotOccupied_WaitsWithBackoff(t *testing.T) {
	am := newPendingAM("app1")
	w := domain.World{
		AM:           am,
		Now:          time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		AppExists:    false,
		ClusterStats: domain.ClusterStats{DownloadingCount: 1},
	}

	res := handlers.OnPendingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in Pending)", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if patch.Status.State != appv1.Pending {
		t.Fatalf("PatchAM.State = %q, want Pending (staying in place)", patch.Status.State)
	}
	if patch.Status.Reason != "WaitingForDownloadingSlot" {
		t.Fatalf("PatchAM.Reason = %q, want WaitingForDownloadingSlot", patch.Status.Reason)
	}

	tick, ok := findEffect[domain.ScheduleTick](res.Effects)
	if !ok {
		t.Fatal("no ScheduleTick effect on bed-full branch")
	}
	if tick.After <= 0 {
		t.Fatalf("ScheduleTick.After = %v, want > 0", tick.After)
	}

	// No EmitEvent — see the waitForSlot comment for why we
	// deliberately don't spam events per waited tick.
	if _, ok := findEffect[domain.EmitEvent](res.Effects); ok {
		t.Fatal("EmitEvent emitted on wait branch; expected none to avoid event storm")
	}
}

// TestOnPendingTick_SlotJustFull pins the boundary: exactly
// pendingDownloadingSlotSize other AMs already downloading is still
// full (the check is `>= size`, not `> size`).
func TestOnPendingTick_SlotJustFull(t *testing.T) {
	am := newPendingAM("app1")
	w := domain.World{
		AM:           am,
		ClusterStats: domain.ClusterStats{DownloadingCount: 1}, // == slot size
	}
	res := handlers.OnPendingTick(context.Background(), w)
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (bed full boundary)", res.Next)
	}
	if _, ok := findEffect[domain.ScheduleTick](res.Effects); !ok {
		t.Fatal("no ScheduleTick emitted at boundary")
	}
}

// TestOnPendingTick_AppExists_SurfacesInvariant pins the torn-state
// invariant: a Pending AM whose Application CR exists is a leak
// (helm ran but the state machine never left Pending). Handler
// must surface it as an Err so tests and operators catch it.
func TestOnPendingTick_AppExists_SurfacesInvariant(t *testing.T) {
	am := newPendingAM("app1")
	w := domain.World{
		AM:        am,
		AppExists: true,
	}
	res := handlers.OnPendingTick(context.Background(), w)
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for AppExists-while-Pending invariant violation")
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty on invariant failure", res.Next)
	}
	if len(res.Effects) != 0 {
		t.Fatalf("Effects = %+v, want none on invariant failure (cleanup Effect is a follow-up)", res.Effects)
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// TestPending_EndToEnd_SoloAdvances pins the end-to-end wiring: a
// single Pending AM (nothing else in the cluster) reconciles to
// Downloading in a single tick. Also verifies the AMStore actually
// received the state-change PatchAM.
func TestPending_EndToEnd_SoloAdvances(t *testing.T) {
	h := newE2EHarness(t)
	// Seed a Pending AM. Do NOT seed a paired Application — the
	// happy path requires AppExists=false.
	h.amStore.Add(newPendingAM("myapp"))

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if out.Requeue {
		t.Fatalf("Requeue = true, want false on happy path; outcome = %+v", out)
	}

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("state after tick = %q, want Downloading", got.Status.State)
	}
	if got.Status.Message != "start to download" {
		t.Fatalf("Message = %q, want 'start to download'", got.Status.Message)
	}
}

// TestPending_EndToEnd_TwoWayContention pins the cluster-wide gate:
// two Pending AMs where one has already advanced to Downloading
// (seeded state == Downloading). The still-Pending one MUST wait
// (Requeue-after set from ScheduleTick) and NOT flip to Downloading.
//
// This is the load-bearing test for replacing the legacy
// process-wide mutex with a ClusterStats-driven declarative guard.
func TestPending_EndToEnd_TwoWayContention(t *testing.T) {
	h := newE2EHarness(t)

	// AM A: already Downloading (simulates a peer holding the slot).
	holder := newPendingAM("holder")
	holder.Status.State = appv1.Downloading
	h.amStore.Add(holder)

	// AM B: Pending; must observe the busy slot and wait.
	h.amStore.Add(newPendingAM("waiter"))

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if out.RequeueAfter <= 0 {
		t.Fatalf("RequeueAfter = %v, want > 0 (from bed-full ScheduleTick)", out.RequeueAfter)
	}

	got, _ := h.amStore.Get(context.Background(), "waiter")
	if got.Status.State != appv1.Pending {
		t.Fatalf("state = %q, want Pending (waiter must not preempt holder)", got.Status.State)
	}
	if got.Status.Reason != "WaitingForDownloadingSlot" {
		t.Fatalf("Reason = %q, want WaitingForDownloadingSlot", got.Status.Reason)
	}

	// Holder must be unchanged — the waiter's reconcile must not
	// mutate other AMs.
	holderAfter, _ := h.amStore.Get(context.Background(), "holder")
	if holderAfter.Status.State != appv1.Downloading {
		t.Fatalf("holder state = %q, want Downloading unchanged", holderAfter.Status.State)
	}
}

// TestPending_EndToEnd_SlotDrained_ThenAdvances pins the follow-up
// tick after the peer completes: once the peer leaves Downloading,
// the next Reconcile on the waiter MUST advance. This is the
// "eventual consistency is enough" property that makes dropping
// the mutex safe in practice.
func TestPending_EndToEnd_SlotDrained_ThenAdvances(t *testing.T) {
	h := newE2EHarness(t)

	holder := newPendingAM("holder")
	holder.Status.State = appv1.Downloading
	h.amStore.Add(holder)
	h.amStore.Add(newPendingAM("waiter"))

	// R1: waiter must wait.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "waiter")
	if got.Status.State != appv1.Pending {
		t.Fatalf("R1 state = %q, want Pending", got.Status.State)
	}

	// Peer completes its download — drop it out of Downloading.
	// Use PatchStatus directly (as the completing peer's
	// Reconcile would).
	current, _ := h.amStore.Get(context.Background(), "holder")
	if err := h.amStore.PatchStatus(context.Background(), "holder", current.ResourceVersion,
		ports.StatusPatch{State: appv1.Installing}); err != nil {
		t.Fatalf("simulate holder advance: %v", err)
	}

	// R2 on waiter: slot is free, must advance.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "waiter")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("R2 state = %q, want Downloading after slot drained", got.Status.State)
	}
}

// TestPending_EndToEnd_FIFO_OlderPendingWins pins download-slot
// FIFO: two Pending AMs, empty Downloading slot. The later-queued
// AM is reconciled first and must stay Pending; the earlier-queued
// AM then takes the slot. Name order is the opposite of queue
// order so a name-sort would fail this test.
func TestPending_EndToEnd_FIFO_OlderPendingWins(t *testing.T) {
	h := newE2EHarness(t)

	later := newPendingAM("app-a") // would win a name-sort
	later.Status.Payload = map[string]string{
		handlers.PayloadPendingQueuedAt: "2026-01-01T12:00:01Z",
	}
	h.amStore.Add(later)

	earlier := newPendingAM("app-z")
	earlier.Status.Payload = map[string]string{
		handlers.PayloadPendingQueuedAt: "2026-01-01T12:00:00Z",
	}
	h.amStore.Add(earlier)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app-a"}); err != nil {
		t.Fatalf("reconcile later-queued: %v", err)
	}
	gotLater, _ := h.amStore.Get(context.Background(), "app-a")
	if gotLater.Status.State != appv1.Pending {
		t.Fatalf("later-queued state = %q, want Pending (must not jump the FIFO queue)", gotLater.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app-z"}); err != nil {
		t.Fatalf("reconcile earlier-queued: %v", err)
	}
	gotEarlier, _ := h.amStore.Get(context.Background(), "app-z")
	if gotEarlier.Status.State != appv1.Downloading {
		t.Fatalf("earlier-queued state = %q, want Downloading", gotEarlier.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app-a"}); err != nil {
		t.Fatalf("reconcile later-queued after head took slot: %v", err)
	}
	gotLater, _ = h.amStore.Get(context.Background(), "app-a")
	if gotLater.Status.State != appv1.Pending {
		t.Fatalf("later-queued state after head advanced = %q, want Pending", gotLater.Status.State)
	}
}

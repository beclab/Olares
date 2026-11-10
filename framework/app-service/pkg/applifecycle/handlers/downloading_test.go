package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newDownloadingAM(name string, payload map[string]string) *appv1.ApplicationManager {
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
			State:   appv1.Downloading,
			OpType:  appv1.InstallOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func downloadingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnDownloadingTick_EmptyPhase_StartsPull pins the fresh-entry
// branch: an AM whose Payload has no downloading.phase key must
// dispatch CallWorkload{WorkloadPullImage} and stamp phase=pull.
func TestOnDownloadingTick_EmptyPhase_StartsPull(t *testing.T) {
	am := newDownloadingAM("app1", nil)
	res := handlers.OnDownloadingTick(context.Background(), downloadingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in Downloading)", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if patch.Status.State != appv1.Downloading {
		t.Fatalf("PatchAM.State = %q, want Downloading", patch.Status.State)
	}
	if got := patch.Status.Payload[handlers.PayloadDownloadingPhase]; got != handlers.PhaseDownloadingPull {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadDownloadingPhase, got, handlers.PhaseDownloadingPull)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadPullImage {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadPullImage", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("CallWorkload OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}

	if _, ok := findEffect[domain.EmitEvent](res.Effects); !ok {
		t.Fatal("no EmitEvent on entry")
	}
}

func TestOnDownloadingTick_PullPhaseRecoversWithoutStatusEffects(t *testing.T) {
	am := newDownloadingAM("app1", map[string]string{
		handlers.PayloadDownloadingPhase: handlers.PhaseDownloadingPull,
	})
	assertTickRecovery(t, handlers.OnDownloadingTick, downloadingWorldFor(am), domain.CallWorkload{
		Kind:      domain.WorkloadPullImage,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "ImagePullFailed"},
	})
}

// TestOnDownloadingTick_UnknownPhase_HardFails pins the safety net
// for a bogus phase value — treated as a bug, not silent no-op.
func TestOnDownloadingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newDownloadingAM("app1", map[string]string{
		handlers.PayloadDownloadingPhase: "grabbed-by-alien",
	})
	res := handlers.OnDownloadingTick(context.Background(), downloadingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnDownloadingWorkloadReady_PullPhase_Advances pins the
// terminal branch: pull callback → hand off to Installing.
func TestOnDownloadingWorkloadReady_PullPhase_Advances(t *testing.T) {
	am := newDownloadingAM("app1", map[string]string{
		handlers.PayloadDownloadingPhase: handlers.PhaseDownloadingPull,
	})
	res := handlers.OnDownloadingWorkloadReady(context.Background(), downloadingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Installing {
		t.Fatalf("Next = %q, want Installing", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Installing {
		t.Fatalf("PatchAM.State = %q, want Installing", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadDownloadingPhase]; ok {
		t.Fatal("Payload[downloading.phase] still set after leaving Downloading; must be cleared")
	}

	// Downloading→Installing is a plain handoff — no OpRecord;
	// legacy also does not write one here (only at terminal
	// Success/Failed states).
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on Downloading→Installing handoff (legacy writes none)")
	}
}

// TestOnDownloadingWorkloadReady_EmptyPhase_HardFails — same
// invariant as installing.OnWorkloadReady: every CallWorkload must
// stamp the phase first.
func TestOnDownloadingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newDownloadingAM("app1", nil)
	res := handlers.OnDownloadingWorkloadReady(context.Background(), downloadingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

// TestOnDownloadingWorkloadFailed_UsesActionData pins the failure
// branch: WorkloadFailed's Reason/Message (populated by the
// Runner from ImagePuller error data) land on the AM.
func TestOnDownloadingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newDownloadingAM("app1", map[string]string{
		handlers.PayloadDownloadingPhase: handlers.PhaseDownloadingPull,
	})
	w := downloadingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "ImagePullStartFailed",
		Message: "registry unreachable",
	}

	res := handlers.OnDownloadingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.DownloadFailed {
		t.Fatalf("Next = %q, want DownloadFailed", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "ImagePullStartFailed" {
		t.Fatalf("Reason = %q, want ImagePullStartFailed (from Action)", patch.Status.Reason)
	}
	if patch.Status.Message != "registry unreachable" {
		t.Fatalf("Message = %q, want the Action's Message", patch.Status.Message)
	}

	// Effect order: OpRecord MUST precede PatchAM (see
	// installing.go for the RV-conflict rationale).
	if len(res.Effects) < 2 {
		t.Fatalf("Effects len = %d, want >= 2", len(res.Effects))
	}
	if _, ok := res.Effects[0].(domain.AppendOpRecord); !ok {
		t.Fatalf("Effects[0] = %T, want AppendOpRecord", res.Effects[0])
	}
	if _, ok := res.Effects[1].(domain.PatchAM); !ok {
		t.Fatalf("Effects[1] = %T, want PatchAM", res.Effects[1])
	}

	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok {
		t.Fatal("no EmitEvent")
	}
	if ev.Type != "Warning" {
		t.Fatalf("EmitEvent.Type = %q, want Warning", ev.Type)
	}
}

// TestOnDownloadingWorkloadFailed_FallsBackWhenActionEmpty pins
// the safety net: a WorkloadFailed with no Reason/Message still
// produces sensible defaults so InstallFailed doesn't land empty.
func TestOnDownloadingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newDownloadingAM("app1", map[string]string{
		handlers.PayloadDownloadingPhase: handlers.PhaseDownloadingPull,
	})
	w := downloadingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnDownloadingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason == "" {
		t.Fatal("fallback Reason empty")
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// seedDownloading seeds an AM in Downloading state with empty phase.
func (h *e2eHarness) seedDownloading(name string) string {
	am := newDownloadingAM(name, nil)
	// Downloading state: paired Application does NOT exist yet
	// (helm hasn't run). Don't seed AddApp.
	return h.amStore.Add(am)
}

// TestDownload_EndToEnd_HappyPath drives Downloading → Installing
// through two Reconciles:
//
//	R1: Downloading(phase="")    → dispatch WorkloadPullImage
//	R2: Downloading(phase=pull)  → pull callback → Installing
func TestDownload_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedDownloading("myapp")

	// R1: dispatches pull.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("after R1, state = %q, want Downloading", got.Status.State)
	}
	if p := got.Status.Payload[handlers.PayloadDownloadingPhase]; p != handlers.PhaseDownloadingPull {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseDownloadingPull)
	}

	// Runner's dispatchImagePull should have called CreateJob.
	calls := h.imagePuller.Calls()
	if len(calls) == 0 || calls[len(calls)-1].Method != "CreateJob" {
		t.Fatalf("image puller calls = %+v, want last method CreateJob", calls)
	}

	// R2: consumes WorkloadReady, advances to Installing.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Installing {
		t.Fatalf("after R2, state = %q, want Installing", got.Status.State)
	}
	if _, ok := got.Status.Payload[handlers.PayloadDownloadingPhase]; ok {
		t.Fatal("Payload[downloading.phase] still set after leaving Downloading")
	}
}

// TestDownload_EndToEnd_ResolveFailure pins the failure branch:
// ResolveRefs returns err → Runner packs it into WorkloadFailed →
// handler lands DownloadFailed with the Runner-provided Reason.
func TestDownload_EndToEnd_ResolveFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedDownloading("myapp")

	h.imagePuller.ResolveRefsFunc = func(_ context.Context, _ *appv1.ApplicationManager, _ any) ([]appv1.Ref, error) {
		return nil, errors.New("appcfg parse error")
	}

	// R1: dispatches pull (fails inside goroutine).
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	// R2: consumes WorkloadFailed, lands DownloadFailed.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.DownloadFailed {
		t.Fatalf("state = %q, want DownloadFailed", got.Status.State)
	}
	if got.Status.Reason != "ImageResolveFailed" {
		t.Fatalf("Reason = %q, want ImageResolveFailed (Runner-classified)", got.Status.Reason)
	}
	// OpRecord landed.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecord after failure")
	}
	if got.Status.OpRecords[0].Status != appv1.DownloadFailed {
		t.Fatalf("head OpRecord.Status = %q, want DownloadFailed", got.Status.OpRecords[0].Status)
	}
}

// TestDownload_EndToEnd_CreateJobFailure pins the second failure
// mode: ResolveRefs succeeds but CreateJob errors.
func TestDownload_EndToEnd_CreateJobFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedDownloading("myapp")

	h.imagePuller.CreateJobFunc = func(_ context.Context, _ *appv1.ApplicationManager, _ []appv1.Ref) error {
		return errors.New("registry down")
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.DownloadFailed {
		t.Fatalf("state = %q, want DownloadFailed", got.Status.State)
	}
	if got.Status.Reason != "ImagePullStartFailed" {
		t.Fatalf("Reason = %q, want ImagePullStartFailed", got.Status.Reason)
	}
}

// TestInstallLifecycle_EndToEnd_PendingThroughInitializing is the
// load-bearing chained-driver test: exercises the full Pending →
// Downloading → Installing → Initializing sequence with THREE
// handlers wired (pending, downloading, installing). If any
// handler's Payload-clearing or state-flip semantics regress, this
// test breaks.
func TestInstallLifecycle_EndToEnd_PendingThroughInitializing(t *testing.T) {
	h := newE2EHarness(t)
	// Seed Pending; happy path requires AppExists=false.
	h.amStore.Add(newPendingAM("myapp"))

	// R1: Pending → Downloading (handler emits PatchAM directly,
	// no CallWorkload, so no drain needed).
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("R1 state = %q, want Downloading", got.Status.State)
	}

	// R2: Downloading (phase="") → dispatch pull.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)

	// R3: Downloading (phase="pull") + WorkloadReady → Installing.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Installing {
		t.Fatalf("R3 state = %q, want Installing", got.Status.State)
	}

	// R4-R8: the Installing pipeline — validate, helm-install,
	// scale-up, wait-startup, then the hand-off to Initializing.
	driveInstallToInitializing(t, h, "myapp")

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after install pipeline, state = %q, want Initializing", got.Status.State)
	}
	// Final OpRecord must be the Initializing (success) one.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after full lifecycle")
	}
	if got.Status.OpRecords[0].Status != appv1.Initializing {
		t.Fatalf("head OpRecord.Status = %q, want Initializing (from install success)", got.Status.OpRecords[0].Status)
	}
}

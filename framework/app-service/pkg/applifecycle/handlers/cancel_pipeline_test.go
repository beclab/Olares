package handlers_test

import (
	"context"
	"errors"
	"reflect"
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
// Fixtures — build AMs already seated in *Canceling states so the
// Tick / WorkloadReady / WorkloadFailed handlers can be exercised
// directly without going through their source-state CancelOp.
// -----------------------------------------------------------------

func newCancelingAM(name string, state appv1.ApplicationManagerState, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.CancelOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   state,
			OpType:  appv1.CancelOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func cancelingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// worldWithAction attaches an Action to the World — used to
// simulate the Runner threading a WorkloadFailed{Reason, Message}
// into the failure branch tests.
func worldWithAction(am *appv1.ApplicationManager, a domain.Action) domain.World {
	w := cancelingWorldFor(am)
	w.Action = a
	return w
}

// -----------------------------------------------------------------
// Layer 1 — SOURCE → *Canceling transitions
// -----------------------------------------------------------------

type sourceCancelCase struct {
	name    string
	source  appv1.ApplicationManagerState
	handler func(context.Context, domain.World) domain.Result
	target  appv1.ApplicationManagerState
}

func sourceCancelMatrix() []sourceCancelCase {
	return []sourceCancelCase{
		{"pending", appv1.Pending, handlers.OnPendingCancel, appv1.PendingCanceling},
		{"downloading", appv1.Downloading, handlers.OnDownloadingCancel, appv1.DownloadingCanceling},
		{"installing", appv1.Installing, handlers.OnInstallingCancel, appv1.InstallingCanceling},
	}
}

// TestSourceCancel_Transitions pins the CancelOp flip for all
// three Pattern A/B source states.
func TestSourceCancel_Transitions(t *testing.T) {
	for _, tc := range sourceCancelMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newCancelingAM("app1", tc.source, map[string]string{
				// Simulate a source-state phase already stamped —
				// verifies the CancelOp handler clears it.
				handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
			})
			res := tc.handler(context.Background(), cancelingWorldFor(am))
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != tc.target {
				t.Fatalf("Next = %q, want %q", res.Next, tc.target)
			}
			patch, ok := findEffect[domain.PatchAM](res.Effects)
			if !ok {
				t.Fatal("no PatchAM")
			}
			if patch.Status.Reason != "CancelRequested" {
				t.Fatalf("Reason = %q, want CancelRequested", patch.Status.Reason)
			}
			// Payload cleared — source phase key gone.
			if patch.Status.Payload == nil || len(patch.Status.Payload) != 0 {
				t.Fatalf("Payload = %+v, want empty non-nil map", patch.Status.Payload)
			}
			if tc.source == appv1.Installing {
				if _, ok := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; ok {
					t.Fatal("user CancelOp unexpectedly stamped an InstallFailed terminal intent")
				}
			}
			// No CallWorkload at this edge — the *Canceling Tick
			// dispatches its own phase kickoff.
			if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
				t.Fatal("unexpected CallWorkload on source→*Canceling")
			}
		})
	}
}

// -----------------------------------------------------------------
// Layer 1 — *Canceling Tick empty-phase → dispatch first CallWorkload
// -----------------------------------------------------------------
//
// Each Pattern's first-phase entry effect is uniquely identifiable
// by the (Kind, phase key, phase value) triple stamped in the
// PatchAM. Test them individually since the shape differs
// (Pattern A: one Kind, Pattern B: image cancel, Pattern C:
// helm uninstall).

func TestPendingCancelingTick_EmptyPhase_DispatchesWaitNS(t *testing.T) {
	am := newCancelingAM("app1", appv1.PendingCanceling, nil)
	res := handlers.OnPendingCancelingTick(context.Background(), cancelingWorldFor(am))
	assertPhaseAndCall(t, res, handlers.PayloadPendingCancelingPhase,
		handlers.PhaseCancelWaitNS, domain.CallWorkload{
			Kind:      domain.WorkloadWaitNamespaceGone,
			OnSuccess: domain.WorkloadReady{},
			OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
		})
}

func TestDownloadingCancelingTick_EmptyPhase_DispatchesCancelPull(t *testing.T) {
	am := newCancelingAM("app1", appv1.DownloadingCanceling, nil)
	res := handlers.OnDownloadingCancelingTick(context.Background(), cancelingWorldFor(am))
	assertPhaseAndCall(t, res, handlers.PayloadDownloadingCancelingPhase,
		handlers.PhaseCancelImagePull, domain.CallWorkload{
			Kind:      domain.WorkloadCancelImagePull,
			OnSuccess: domain.WorkloadReady{},
			OnFailure: domain.WorkloadFailed{Reason: "ImageCancelFailed"},
			// The pull being canceled normally still holds the
			// AM's workload claim.
			Preempts: true,
		})
}

func TestInstallingCancelingTick_EmptyPhase_DispatchesHelmUninstall(t *testing.T) {
	am := newCancelingAM("app1", appv1.InstallingCanceling, nil)
	res := handlers.OnInstallingCancelingTick(context.Background(), cancelingWorldFor(am))
	assertPhaseAndCall(t, res, handlers.PayloadInstallingCancelingPhase,
		handlers.PhaseCancelHelmUninstall, domain.CallWorkload{
			Kind:      domain.WorkloadUninstall,
			OnSuccess: domain.WorkloadReady{},
			OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
			Preempts:  true,
		})
}

// assertPhaseAndCall is the shared Tick-empty-phase assertion —
// verifies a PatchAM stamps the expected phase and a CallWorkload
// with the expected callback contract fires.
func assertPhaseAndCall(t *testing.T, res domain.Result, payloadKey, phase string, call domain.CallWorkload) {
	t.Helper()
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in *Canceling)", res.Next)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[payloadKey]; got != phase {
		t.Fatalf("Payload[%q] = %q, want %q", payloadKey, got, phase)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if !reflect.DeepEqual(cw, call) {
		t.Fatalf("CallWorkload = %#v, want exactly %#v", cw, call)
	}
}

// -----------------------------------------------------------------
// Layer 1 — In-flight Tick is a no-op (stray periodic Tick guard)
// -----------------------------------------------------------------

type inflightTickCase struct {
	name         string
	state        appv1.ApplicationManagerState
	payloadKey   string
	phase        string
	inFlightKind domain.WorkloadKind
	handler      func(context.Context, domain.World) domain.Result
}

func inflightTickMatrix() []inflightTickCase {
	return []inflightTickCase{
		{"pendingCanceling/empty", appv1.PendingCanceling,
			handlers.PayloadPendingCancelingPhase, "",
			domain.WorkloadWaitNamespaceGone,
			handlers.OnPendingCancelingTick},
		{"pendingCanceling/waitns", appv1.PendingCanceling,
			handlers.PayloadPendingCancelingPhase, handlers.PhaseCancelWaitNS,
			domain.WorkloadWaitNamespaceGone,
			handlers.OnPendingCancelingTick},
		{"downloadingCanceling/empty", appv1.DownloadingCanceling,
			handlers.PayloadDownloadingCancelingPhase, "",
			domain.WorkloadCancelImagePull,
			handlers.OnDownloadingCancelingTick},
		{"downloadingCanceling/cancelpull", appv1.DownloadingCanceling,
			handlers.PayloadDownloadingCancelingPhase, handlers.PhaseCancelImagePull,
			domain.WorkloadCancelImagePull,
			handlers.OnDownloadingCancelingTick},
		{"downloadingCanceling/waitns", appv1.DownloadingCanceling,
			handlers.PayloadDownloadingCancelingPhase, handlers.PhaseCancelWaitNS,
			domain.WorkloadWaitNamespaceGone,
			handlers.OnDownloadingCancelingTick},
		{"installingCanceling/empty", appv1.InstallingCanceling,
			handlers.PayloadInstallingCancelingPhase, "",
			domain.WorkloadUninstall,
			handlers.OnInstallingCancelingTick},
		{"installingCanceling/helm", appv1.InstallingCanceling,
			handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelHelmUninstall,
			domain.WorkloadUninstall,
			handlers.OnInstallingCancelingTick},
		{"installingCanceling/waitns", appv1.InstallingCanceling,
			handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelWaitNS,
			domain.WorkloadWaitNamespaceGone,
			handlers.OnInstallingCancelingTick},
	}
}

func TestCancelingTick_InFlightPhase_IsNoop(t *testing.T) {
	for _, tc := range inflightTickMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newCancelingAM("app1", tc.state, map[string]string{
				tc.payloadKey: tc.phase,
			})
			w := cancelingWorldFor(am)
			w.WorkloadInFlight = true
			w.InFlightKind = tc.inFlightKind
			res := tc.handler(context.Background(), w)
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if len(res.Effects) != 0 {
				t.Fatalf("Effects = %+v, want none (in-flight Tick must not re-dispatch)", res.Effects)
			}
		})
	}
}

func TestDownloadingCancelingTick_EmptyPhase_StillDispatchesWhenPullInFlight(t *testing.T) {
	am := newCancelingAM("app1", appv1.DownloadingCanceling, nil)
	w := cancelingWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadPullImage
	res := handlers.OnDownloadingCancelingTick(context.Background(), w)
	assertPhaseAndCall(t, res, handlers.PayloadDownloadingCancelingPhase,
		handlers.PhaseCancelImagePull, domain.CallWorkload{
			Kind:      domain.WorkloadCancelImagePull,
			OnSuccess: domain.WorkloadReady{},
			OnFailure: domain.WorkloadFailed{Reason: "ImageCancelFailed"},
			Preempts:  true,
		})
}

func TestInstallingCancelingTick_EmptyPhase_StillDispatchesWhenInstallInFlight(t *testing.T) {
	am := newCancelingAM("app1", appv1.InstallingCanceling, nil)
	w := cancelingWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadInstall
	res := handlers.OnInstallingCancelingTick(context.Background(), w)
	assertPhaseAndCall(t, res, handlers.PayloadInstallingCancelingPhase,
		handlers.PhaseCancelHelmUninstall, domain.CallWorkload{
			Kind:      domain.WorkloadUninstall,
			OnSuccess: domain.WorkloadReady{},
			OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
			Preempts:  true,
		})
}

func TestCancelingTick_PersistedPhase_RecoversOutstandingWorkload(t *testing.T) {
	cases := []struct {
		name       string
		state      appv1.ApplicationManagerState
		payloadKey string
		phase      string
		handler    func(context.Context, domain.World) domain.Result
		call       domain.CallWorkload
	}{
		{
			name:       "pending/wait-ns",
			state:      appv1.PendingCanceling,
			payloadKey: handlers.PayloadPendingCancelingPhase,
			phase:      handlers.PhaseCancelWaitNS,
			handler:    handlers.OnPendingCancelingTick,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitNamespaceGone,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
			},
		},
		{
			name:       "downloading/cancel-pull",
			state:      appv1.DownloadingCanceling,
			payloadKey: handlers.PayloadDownloadingCancelingPhase,
			phase:      handlers.PhaseCancelImagePull,
			handler:    handlers.OnDownloadingCancelingTick,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadCancelImagePull,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "ImageCancelFailed"},
				Preempts:  true,
			},
		},
		{
			name:       "downloading/wait-ns",
			state:      appv1.DownloadingCanceling,
			payloadKey: handlers.PayloadDownloadingCancelingPhase,
			phase:      handlers.PhaseCancelWaitNS,
			handler:    handlers.OnDownloadingCancelingTick,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitNamespaceGone,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
			},
		},
		{
			name:       "installing/helm-uninstall",
			state:      appv1.InstallingCanceling,
			payloadKey: handlers.PayloadInstallingCancelingPhase,
			phase:      handlers.PhaseCancelHelmUninstall,
			handler:    handlers.OnInstallingCancelingTick,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadUninstall,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
				Preempts:  true,
			},
		},
		{
			name:       "installing/wait-ns",
			state:      appv1.InstallingCanceling,
			payloadKey: handlers.PayloadInstallingCancelingPhase,
			phase:      handlers.PhaseCancelWaitNS,
			handler:    handlers.OnInstallingCancelingTick,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitNamespaceGone,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newCancelingAM("app1", tc.state, map[string]string{tc.payloadKey: tc.phase})
			w := cancelingWorldFor(am)

			recovered := tc.handler(context.Background(), w)
			want := domain.Result{Effects: []domain.Effect{tc.call}}
			if !reflect.DeepEqual(recovered, want) {
				t.Fatalf("recovered Result = %#v, want exactly %#v", recovered, want)
			}

			w.WorkloadInFlight = true
			inFlight := tc.handler(context.Background(), w)
			if !reflect.DeepEqual(inFlight, domain.Result{}) {
				t.Fatalf("in-flight Result = %#v, want empty", inFlight)
			}
		})
	}
}

// TestCancelingTick_UnknownPhase_HardFails is the safety-net: an
// unrecognised phase key must fail loudly, not silently no-op.
func TestCancelingTick_UnknownPhase_HardFails(t *testing.T) {
	cases := []struct {
		name       string
		state      appv1.ApplicationManagerState
		payloadKey string
		handler    func(context.Context, domain.World) domain.Result
	}{
		{"pending", appv1.PendingCanceling, handlers.PayloadPendingCancelingPhase, handlers.OnPendingCancelingTick},
		{"downloading", appv1.DownloadingCanceling, handlers.PayloadDownloadingCancelingPhase, handlers.OnDownloadingCancelingTick},
		{"installing", appv1.InstallingCanceling, handlers.PayloadInstallingCancelingPhase, handlers.OnInstallingCancelingTick},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newCancelingAM("app1", tc.state, map[string]string{tc.payloadKey: "bogus"})
			res := tc.handler(context.Background(), cancelingWorldFor(am))
			if res.Err == nil {
				t.Fatal("Err = nil, want failure for unknown phase")
			}
		})
	}
}

// -----------------------------------------------------------------
// Layer 1 — WorkloadReady advances phase→phase→terminal
// -----------------------------------------------------------------

// TestPendingCancelingWorkloadReady_WaitNS_FinishesToCanceled pins
// Pattern A's single-phase terminal: WorkloadReady at wait-ns →
// PendingCanceled (no OpRecord).
func TestPendingCancelingWorkloadReady_WaitNS_FinishesToCanceled(t *testing.T) {
	am := newCancelingAM("app1", appv1.PendingCanceling, map[string]string{
		handlers.PayloadPendingCancelingPhase: handlers.PhaseCancelWaitNS,
	})
	res := handlers.OnPendingCancelingWorkloadReady(context.Background(), cancelingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.PendingCanceled {
		t.Fatalf("Next = %q, want PendingCanceled", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.PendingCanceled {
		t.Fatalf("PatchAM.State = %q, want PendingCanceled", patch.Status.State)
	}
	if patch.Status.Payload == nil || len(patch.Status.Payload) != 0 {
		t.Fatalf("Payload = %+v, want empty non-nil (clear phase)", patch.Status.Payload)
	}
	// Pattern A does NOT emit OpRecord (matches legacy withOpRecord=false).
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord for Pattern A (Pending)")
	}
}

// TestDownloadingCancelingWorkloadReady_CancelPull_AdvancesToWaitNS
// pins Pattern B's phase-1 → phase-2 advance.
func TestDownloadingCancelingWorkloadReady_CancelPull_AdvancesToWaitNS(t *testing.T) {
	am := newCancelingAM("app1", appv1.DownloadingCanceling, map[string]string{
		handlers.PayloadDownloadingCancelingPhase: handlers.PhaseCancelImagePull,
	})
	res := handlers.OnDownloadingCancelingWorkloadReady(context.Background(), cancelingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in DownloadingCanceling)", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadDownloadingCancelingPhase]; got != handlers.PhaseCancelWaitNS {
		t.Fatalf("phase after advance = %q, want %q", got, handlers.PhaseCancelWaitNS)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadWaitNamespaceGone {
		t.Fatalf("Kind = %q, want WaitNamespaceGone", cw.Kind)
	}
}

// TestInstallingCancelingWorkloadReady_WaitNS_FinishesWithOpRecord
// pins Pattern C's key legacy quirk: unlike Pattern A/B, the
// terminal WorkloadReady emits an OpRecord (withOpRecord=true).
func TestInstallingCancelingWorkloadReady_WaitNS_FinishesWithOpRecord(t *testing.T) {
	am := newCancelingAM("app1", appv1.InstallingCanceling, map[string]string{
		handlers.PayloadInstallingCancelingPhase: handlers.PhaseCancelWaitNS,
	})
	res := handlers.OnInstallingCancelingWorkloadReady(context.Background(), cancelingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCanceled {
		t.Fatalf("Next = %q, want InstallingCanceled", res.Next)
	}
	op, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord — Pattern C must emit one")
	}
	if op.Record.Status != appv1.InstallingCanceled {
		t.Fatalf("OpRecord.Status = %q, want InstallingCanceled", op.Record.Status)
	}
	// Effect ordering: AppendOpRecord before PatchAM (same
	// invariant as installing.finishInstall). Verify by index.
	var opIdx, patchIdx = -1, -1
	for i, e := range res.Effects {
		switch e.(type) {
		case domain.AppendOpRecord:
			opIdx = i
		case domain.PatchAM:
			patchIdx = i
		}
	}
	if opIdx == -1 || patchIdx == -1 || opIdx > patchIdx {
		t.Fatalf("Effect ordering broken: opIdx=%d patchIdx=%d — AppendOpRecord must precede PatchAM", opIdx, patchIdx)
	}
}

// -----------------------------------------------------------------
// Layer 1 — InstallingCanceling terminal intent
// -----------------------------------------------------------------
//
// The same pipeline serves two callers: a user cancel (no intent
// payload, lands InstallingCanceled) and an install failure (intent
// = InstallFailed, lands InstallFailed with the stashed diagnostic).

// TestInstallingCancelingWaitNS_WithInstallFailedIntent_LandsInstallFailed
// drives the install-failure leg: helm-uninstall Ready must carry the
// intent into the wait-ns phase, and wait-ns Ready must land
// InstallFailed rather than InstallingCanceled.
func TestInstallingCancelingWaitNS_WithInstallFailedIntent_LandsInstallFailed(t *testing.T) {
	intent := map[string]string{
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	}

	uninstalled := newCancelingAM("app1", appv1.InstallingCanceling, withPhase(intent,
		handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelHelmUninstall))
	advance := handlers.OnInstallingCancelingWorkloadReady(context.Background(), cancelingWorldFor(uninstalled))
	if advance.Err != nil {
		t.Fatalf("helm-uninstall Ready: Err = %v", advance.Err)
	}
	advancePatch, _ := findEffect[domain.PatchAM](advance.Effects)
	if got := advancePatch.Status.Payload[handlers.PayloadInstallingCancelingPhase]; got != handlers.PhaseCancelWaitNS {
		t.Fatalf("phase after advance = %q, want %q", got, handlers.PhaseCancelWaitNS)
	}
	for k, want := range intent {
		if got := advancePatch.Status.Payload[k]; got != want {
			t.Fatalf("wait-ns patch dropped Payload[%q] = %q, want %q", k, got, want)
		}
	}

	waiting := newCancelingAM("app1", appv1.InstallingCanceling, withPhase(intent,
		handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelWaitNS))
	res := handlers.OnInstallingCancelingWorkloadReady(context.Background(), cancelingWorldFor(waiting))
	if res.Err != nil {
		t.Fatalf("wait-ns Ready: Err = %v", res.Err)
	}
	if res.Next != appv1.InstallFailed {
		t.Fatalf("Next = %q, want InstallFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.InstallFailed {
		t.Fatalf("PatchAM.State = %q, want InstallFailed", patch.Status.State)
	}
	// The original failure, not the cancel pipeline's own wording.
	if patch.Status.Message != "chart apply failed" {
		t.Fatalf("PatchAM.Message = %q, want the stashed install failure", patch.Status.Message)
	}
	if patch.Status.Reason != "HelmError" {
		t.Fatalf("PatchAM.Reason = %q, want HelmError", patch.Status.Reason)
	}
	// Cleanup already ran here; InstallFailed's Tick must not re-run it.
	if patch.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanupDone {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallFailedCleanup,
			patch.Status.Payload[handlers.PayloadInstallFailedCleanup], handlers.PhaseInstallFailedCleanupDone)
	}
	for _, staleKey := range []string{
		handlers.PayloadInstallingCancelingPhase,
		handlers.PayloadInstallingCancelingTerminal,
		handlers.PayloadInstallingFailMessage,
		handlers.PayloadInstallingFailReason,
	} {
		if _, ok := patch.Status.Payload[staleKey]; ok {
			t.Fatalf("terminal PatchAM retained stale Payload[%q]", staleKey)
		}
	}
	op, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord")
	}
	if op.Record.Status != appv1.InstallFailed {
		t.Fatalf("OpRecord.Status = %q, want InstallFailed", op.Record.Status)
	}
}

// TestInstallingCancelingWaitNS_UserCancelStillLandsInstallingCanceled
// is the other half of the branch: without the terminal-intent key the
// pipeline keeps its user-cancel semantics.
func TestInstallingCancelingWaitNS_UserCancelStillLandsInstallingCanceled(t *testing.T) {
	am := newCancelingAM("app1", appv1.InstallingCanceling, map[string]string{
		handlers.PayloadInstallingCancelingPhase: handlers.PhaseCancelWaitNS,
	})
	res := handlers.OnInstallingCancelingWorkloadReady(context.Background(), cancelingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCanceled {
		t.Fatalf("Next = %q, want InstallingCanceled", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if _, ok := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; ok {
		t.Fatal("user cancellation terminal patch unexpectedly carries InstallFailed intent")
	}
	op, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord")
	}
	if op.Record.Status != appv1.InstallingCanceled {
		t.Fatalf("OpRecord.Status = %q, want InstallingCanceled", op.Record.Status)
	}
}

// TestInstallingCancelingCleanupFailureStillLandsInstallingCancelFailed
// pins that the terminal intent only redirects the SUCCESS branch: a
// cleanup workload that fails is still a failed cancel, not an
// InstallFailed.
func TestInstallingCancelingCleanupFailureStillLandsInstallingCancelFailed(t *testing.T) {
	am := newCancelingAM("app1", appv1.InstallingCanceling, map[string]string{
		handlers.PayloadInstallingCancelingPhase:    handlers.PhaseCancelWaitNS,
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	})
	w := worldWithAction(am, domain.WorkloadFailed{
		Reason:  "NSPollTimeout",
		Message: "namespace still terminating",
	})
	res := handlers.OnInstallingCancelingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCancelFailed {
		t.Fatalf("Next = %q, want InstallingCancelFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "NSPollTimeout" {
		t.Fatalf("Reason = %q, want NSPollTimeout", patch.Status.Reason)
	}
	for key, want := range map[string]string{
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	} {
		if got := patch.Status.Payload[key]; got != want {
			t.Fatalf("InstallingCancelFailed patch Payload[%q] = %q, want %q", key, got, want)
		}
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingPhase]; got != "" {
		t.Fatalf("InstallingCancelFailed patch phase = %q, want cleared", got)
	}
}

// withPhase returns a copy of base with the phase key set, so the
// intent fixture can be reused across phases without aliasing.
func withPhase(base map[string]string, phaseKey, phase string) map[string]string {
	out := make(map[string]string, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	out[phaseKey] = phase
	return out
}

// -----------------------------------------------------------------
// Layer 1 — WorkloadFailed → *CancelFailed
// -----------------------------------------------------------------

type failedCase struct {
	name         string
	state        appv1.ApplicationManagerState
	phaseKey     string
	phase        string
	handler      func(context.Context, domain.World) domain.Result
	expectedFail appv1.ApplicationManagerState
	expectedRsn  string // Runner-supplied Reason via the WorkloadFailed action
}

func failedMatrix() []failedCase {
	return []failedCase{
		{"pending/waitns", appv1.PendingCanceling,
			handlers.PayloadPendingCancelingPhase, handlers.PhaseCancelWaitNS,
			handlers.OnPendingCancelingWorkloadFailed,
			appv1.PendingCancelFailed, "NSPollTimeout"},
		{"downloading/cancelpull", appv1.DownloadingCanceling,
			handlers.PayloadDownloadingCancelingPhase, handlers.PhaseCancelImagePull,
			handlers.OnDownloadingCancelingWorkloadFailed,
			appv1.DownloadingCancelFailed, "ImageCancelFailed"},
		{"installing/helm", appv1.InstallingCanceling,
			handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelHelmUninstall,
			handlers.OnInstallingCancelingWorkloadFailed,
			appv1.InstallingCancelFailed, "HelmUninstallFailed"},
	}
}

// TestCancelingWorkloadFailed_Transitions pins the shared
// fail→*CancelFailed pattern including OpRecord + Warning event.
func TestCancelingWorkloadFailed_Transitions(t *testing.T) {
	for _, tc := range failedMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newCancelingAM("app1", tc.state, map[string]string{tc.phaseKey: tc.phase})
			// Simulate Runner-supplied failure reason.
			w := worldWithAction(am, domain.WorkloadFailed{
				Reason:  tc.expectedRsn,
				Message: "workload adapter error",
			})
			res := tc.handler(context.Background(), w)
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != tc.expectedFail {
				t.Fatalf("Next = %q, want %q", res.Next, tc.expectedFail)
			}
			patch, _ := findEffect[domain.PatchAM](res.Effects)
			if patch.Status.Reason != tc.expectedRsn {
				t.Fatalf("Reason = %q, want %q", patch.Status.Reason, tc.expectedRsn)
			}
			// Every fail branch writes an OpRecord.
			op, ok := findEffect[domain.AppendOpRecord](res.Effects)
			if !ok {
				t.Fatal("no AppendOpRecord")
			}
			if op.Record.Status != tc.expectedFail {
				t.Fatalf("OpRecord.Status = %q, want %q", op.Record.Status, tc.expectedFail)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok {
				t.Fatal("no EmitEvent")
			}
			if ev.Type != "Warning" {
				t.Fatalf("EmitEvent.Type = %q, want Warning", ev.Type)
			}
		})
	}
}

// -----------------------------------------------------------------
// Layer 2 — End-to-end happy path for each pattern
// -----------------------------------------------------------------

// e2eCancelSetup seeds a *Canceling AM with the specified initial
// phase (nil = fresh entry). Used by the three e2e happy-path
// tests below.
func (h *e2eHarness) seedCanceling(name string, state appv1.ApplicationManagerState) string {
	am := newCancelingAM(name, state, nil)
	return h.amStore.Add(am)
}

// TestPendingCanceling_EndToEnd_HappyPath drives PendingCanceling
// → PendingCanceled through 2 Reconciles:
//
//	R1: phase=""      → dispatch WaitNS
//	R2: phase=wait-ns → WorkloadReady → PendingCanceled (no OpRecord)
func TestPendingCanceling_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedCanceling("myapp", appv1.PendingCanceling)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.PendingCanceling {
		t.Fatalf("after R1, state = %q, want PendingCanceling", got.Status.State)
	}
	if p := got.Status.Payload[handlers.PayloadPendingCancelingPhase]; p != handlers.PhaseCancelWaitNS {
		t.Fatalf("after R1, phase = %q, want wait-ns", p)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.PendingCanceled {
		t.Fatalf("final state = %q, want PendingCanceled", got.Status.State)
	}
	// Pattern A does NOT emit OpRecord.
	if len(got.Status.OpRecords) != 0 {
		t.Fatalf("OpRecords = %+v, want none (Pattern A)", got.Status.OpRecords)
	}
}

// TestDownloadingCanceling_EndToEnd_HappyPath drives
// DownloadingCanceling → DownloadingCanceled through 3 Reconciles:
//
//	R1: phase=""            → dispatch CancelImagePull
//	R2: phase=cancel-pull   → WorkloadReady → dispatch WaitNS
//	R3: phase=wait-ns       → WorkloadReady → DownloadingCanceled
func TestDownloadingCanceling_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedCanceling("myapp", appv1.DownloadingCanceling)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.DownloadingCanceled {
		t.Fatalf("final state = %q, want DownloadingCanceled", got.Status.State)
	}
	// Verify the ImagePuller.CancelJob was called during phase 1.
	var cancelCalls int
	for _, c := range h.imagePuller.Calls() {
		if c.Method == "CancelJob" {
			cancelCalls++
		}
	}
	if cancelCalls == 0 {
		t.Fatal("ImagePuller.CancelJob was not called")
	}
	if len(got.Status.OpRecords) != 0 {
		t.Fatalf("OpRecords = %+v, want none (Pattern B)", got.Status.OpRecords)
	}
}

// TestInstallingCanceling_EndToEnd_HappyPath drives
// InstallingCanceling → InstallingCanceled through 3 Reconciles:
//
//	R1: phase=""              → dispatch Uninstall
//	R2: phase=helm-uninstall  → WorkloadReady → dispatch WaitNS
//	R3: phase=wait-ns         → WorkloadReady → InstallingCanceled (+ OpRecord)
func TestInstallingCanceling_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedCanceling("myapp", appv1.InstallingCanceling)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCanceled {
		t.Fatalf("final state = %q, want InstallingCanceled", got.Status.State)
	}
	// Pattern C DOES emit OpRecord.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after Pattern C — legacy withOpRecord=true violated")
	}
	if got.Status.OpRecords[0].Status != appv1.InstallingCanceled {
		t.Fatalf("head OpRecord.Status = %q, want InstallingCanceled", got.Status.OpRecords[0].Status)
	}
}

// TestInstallingCanceling_EndToEnd_HelmFailure pins the Pattern C
// phase-1 failure path: WorkloadUninstall returns err → transition
// to InstallingCancelFailed with the workload's Reason.
func TestInstallingCanceling_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedCanceling("myapp", appv1.InstallingCanceling)
	h.workloadOps.UninstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "PartialReleaseCorrupt", Err: errors.New("chart delete failed")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCancelFailed {
		t.Fatalf("state = %q, want InstallingCancelFailed", got.Status.State)
	}
	if got.Status.Reason != "PartialReleaseCorrupt" {
		t.Fatalf("Reason = %q, want PartialReleaseCorrupt (Runner-classified)", got.Status.Reason)
	}
}

// TestInstallingCanceling_EndToEnd_InstallFailureCleanupRetry proves
// that an ordinary install failure keeps its terminal intent across a
// cleanup failure and a later CancelOp retry. The successful retry
// must land InstallFailed; the same retry without intent remains the
// user-cancel path covered by TestInstallingCanceling_EndToEnd_HappyPath.
func TestInstallingCanceling_EndToEnd_InstallFailureCleanupRetry(t *testing.T) {
	h := newE2EHarness(t)
	am := newCancelingAM("myapp", appv1.InstallingCanceling, map[string]string{
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	})
	am.Spec.OpType = appv1.InstallOp
	am.Status.OpType = appv1.InstallOp
	h.amStore.Add(am)

	uninstallCalls := 0
	h.workloadOps.UninstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		uninstallCalls++
		if uninstallCalls == 1 {
			return ports.WorkloadResult{
				LandedReason: "PartialReleaseCorrupt",
				Err:          errors.New("chart delete failed"),
			}
		}
		return ports.WorkloadResult{}
	}

	// R1 dispatch uninstall; R2 consumes its failure.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCancelFailed {
		t.Fatalf("after R2, state = %q, want InstallingCancelFailed", got.Status.State)
	}
	assertInstallFailureIntent(t, got.Status.Payload)
	if _, ok := got.Status.Payload[handlers.PayloadInstallingCancelingPhase]; ok {
		t.Fatal("InstallingCancelFailed retained stale canceling phase")
	}

	// R3 retries via CancelOp and must keep the original install
	// terminal intent while resetting the phase.
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCanceling {
		t.Fatalf("after R3, state = %q, want InstallingCanceling", got.Status.State)
	}
	assertInstallFailureIntent(t, got.Status.Payload)
	if _, ok := got.Status.Payload[handlers.PayloadInstallingCancelingPhase]; ok {
		t.Fatal("retry retained stale canceling phase")
	}

	// R4 uninstall succeeds; R5 advances to wait-ns; R6 lands.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallFailed {
		t.Fatalf("final state = %q, want InstallFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" || got.Status.Message != "chart apply failed" {
		t.Fatalf("final diagnostic = (%q, %q), want original install failure", got.Status.Reason, got.Status.Message)
	}
	if got.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanupDone {
		t.Fatalf("cleanup marker = %q, want done", got.Status.Payload[handlers.PayloadInstallFailedCleanup])
	}
	for _, staleKey := range []string{
		handlers.PayloadInstallingCancelingPhase,
		handlers.PayloadInstallingCancelingTerminal,
		handlers.PayloadInstallingFailMessage,
		handlers.PayloadInstallingFailReason,
	} {
		if _, ok := got.Status.Payload[staleKey]; ok {
			t.Fatalf("final payload retained stale key %q: %+v", staleKey, got.Status.Payload)
		}
	}
}

func assertInstallFailureIntent(t *testing.T, payload map[string]string) {
	t.Helper()
	for key, want := range map[string]string{
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	} {
		if got := payload[key]; got != want {
			t.Fatalf("Payload[%q] = %q, want %q", key, got, want)
		}
	}
}

// -----------------------------------------------------------------
// Layer 2 — Full install-cancel-reinstall lifecycle
// -----------------------------------------------------------------

// TestFullLifecycle_InstallCancel is the marquee load-bearing test
// for Pattern C: drives Pending → Downloading → Installing, cancels
// mid-installing (after helm-install kicked off but before
// completion), verifies the InstallingCanceling pipeline runs
// helm-uninstall + wait-ns, lands in InstallingCanceled, then
// re-installs back through the full pipeline to Running.
//
// This is the "InstallCancel" full lifecycle — bookend to
// TestFullLifecycle_UpgradeCancel (Pattern C handoff variant).
//
// Sequence (14+ Reconciles):
//
//	R1:  Pending → Downloading
//	R2:  Downloading(phase="") → dispatch pull
//	R3:  Downloading(phase=pull) + WorkloadReady → Installing
//	R4:  Installing(phase="") → dispatch expose-ports
//	R5:  Installing(phase=expose-ports) + WorkloadReady → dispatch validate
//	R6:  Installing(phase=validate) + WorkloadReady → dispatch helm-install
//	[cancel injected]
//	R7:  Installing + TriggerOp{CancelOp} → InstallingCanceling
//	R8:  InstallingCanceling(phase="") → dispatch Uninstall
//	R9:  InstallingCanceling(phase=helm-uninstall) + WorkloadReady → dispatch WaitNS
//	R10: InstallingCanceling(phase=wait-ns) + WorkloadReady → InstallingCanceled (+ OpRecord)
//	[re-install injected]
//	R11: InstallingCanceled + TriggerOp{InstallOp} → Pending
//	R12+: full install pipeline back to Running (the Installing leg
//	      is driven by driveInstallToInitializing)
func TestFullLifecycle_InstallCancel(t *testing.T) {
	h := newE2EHarness(t)
	h.amStore.Add(newPendingAM("myapp"))

	installStarted, releaseInstall, installFn := parkedWorkload()
	h.workloadOps.InstallFunc = installFn

	// R1-R6: Pending → Downloading → Installing(expose-ports) →
	// validate → helm-install.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Installing {
		t.Fatalf("after R6, state = %q, want Installing", got.Status.State)
	}
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingHelmInstall {
		t.Fatalf("after R6, phase = %q, want helm-install", p)
	}

	awaitChan(t, installStarted, "helm install to hold the claim")
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})

	// R7: Installing + CancelOp → InstallingCanceling
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCanceling {
		t.Fatalf("after R7, state = %q, want InstallingCanceling", got.Status.State)
	}
	// Source phase was cleared.
	if got.Status.Payload[handlers.PayloadInstallingPhase] != "" {
		t.Fatalf("source Installing phase not cleared: %q", got.Status.Payload[handlers.PayloadInstallingPhase])
	}

	// R8: InstallingCanceling(phase="") → dispatch Uninstall, which
	// must preempt the still-running helm install.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R8: %v", err)
	}
	releaseInstall()
	h.workloadOps.InstallFunc = nil
	drainAndRebuild(t, h)
	var uninstallCalls int
	for _, c := range h.workloadOps.Calls() {
		if c.Method == "Uninstall" {
			uninstallCalls++
		}
	}
	if uninstallCalls == 0 {
		t.Fatal("WorkloadOps.Uninstall was never called; cancel did not preempt helm install")
	}
	// R9: helm-uninstall done → dispatch wait-ns
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R9: %v", err)
	}
	drainAndRebuild(t, h)
	// R10: wait-ns done → InstallingCanceled
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R10: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallingCanceled {
		t.Fatalf("after R10, state = %q, want InstallingCanceled", got.Status.State)
	}
	if len(got.Status.OpRecords) == 0 || got.Status.OpRecords[0].Status != appv1.InstallingCanceled {
		t.Fatalf("no InstallingCanceled OpRecord at head after R10: %+v", got.Status.OpRecords)
	}

	// Re-install: single-op flip to Pending, then full install
	// pipeline back to Running.
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.InstallOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R10: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Pending {
		t.Fatalf("after R10, state = %q, want Pending", got.Status.State)
	}

	// Pending → Downloading → Installing → Initializing → Running.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R11: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R12: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R13: %v", err)
	}
	driveInstallToInitializing(t, h, "myapp")
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("initializing dispatch: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("initializing finish: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}
}

// TestFullLifecycle_DownloadCancel_MidPull drives a cancel that
// arrives while the image pull is STILL RUNNING — the only way a
// user can cancel a download in practice.
//
// Regression: launchCallWorkload refuses a second claim for an AM
// that already has a workload goroutine, so the DownloadingCanceling
// Tick's CallWorkload{CancelImagePull} was silently dropped while the
// pull polled on. ImagePuller.CancelJob never ran, the ImageManager
// CR kept downloading, and the AM sat in DownloadingCanceling for
// ever.
//
// Sequence:
//
//	R1: Pending → Downloading
//	R2: Downloading(phase="") → dispatch pull; goroutine parks at 40%
//	[cancel injected while the pull still holds the claim]
//	R3: Downloading + TriggerOp{CancelOp} → DownloadingCanceling
//	R4: DownloadingCanceling(phase="") → dispatch CancelImagePull
//	R5: WorkloadReady @ cancel-pull → dispatch wait-ns
//	R6: WorkloadReady @ wait-ns → DownloadingCanceled
func TestFullLifecycle_DownloadCancel_MidPull(t *testing.T) {
	h := newE2EHarness(t)
	h.amStore.Add(newPendingAM("myapp"))

	// The pull parks at 40%: the poll loop only wakes when the test
	// advances the fake clock, so the claim stays held across R3/R4.
	h.imagePuller.SetProgress("myapp", ports.ImageProgress{Percent: 40})
	// CancelJob flips the IM to DownloadingCanceled, which Progress
	// reports as Failed — see imgpull.computeProgress. That is what
	// lets the abandoned pull goroutine exit.
	h.imagePuller.CancelJobFunc = func(_ context.Context, name string) error {
		h.imagePuller.SetProgress(name, ports.ImageProgress{
			Percent: 40,
			Failed:  true,
			Reason:  appv1.DownloadingCanceled.String(),
		})
		return nil
	}

	// R1 + R2: Pending → Downloading → pull dispatched.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	awaitPullParked(t, h)

	// The user cancels mid-download.
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})

	// R3: Downloading + CancelOp → DownloadingCanceling.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.DownloadingCanceling {
		t.Fatalf("after R3, state = %q, want DownloadingCanceling", got.Status.State)
	}

	// R4: DownloadingCanceling(phase="") → dispatch CancelImagePull.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)

	var cancelCalls int
	for _, c := range h.imagePuller.Calls() {
		if c.Method == "CancelJob" {
			cancelCalls++
		}
	}
	if cancelCalls == 0 {
		t.Fatal("ImagePuller.CancelJob was never called; the cancel never reached the image puller")
	}

	// R5 + R6: cancel-pull → wait-ns → DownloadingCanceled.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.DownloadingCanceled {
		t.Fatalf("final state = %q, want DownloadingCanceled", got.Status.State)
	}
}

// awaitPullParked blocks until the image-pull goroutine has
// registered its poll ticker with the fake clock, i.e. the claim is
// held and the dispatcher is parked waiting for a tick that only the
// test can deliver.
func awaitPullParked(t *testing.T, h *e2eHarness) {
	t.Helper()
	for i := 0; i < 2000; i++ {
		if h.runner.InFlight() > 0 && h.clock.Pending() > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("image pull goroutine never parked on its poll ticker")
}

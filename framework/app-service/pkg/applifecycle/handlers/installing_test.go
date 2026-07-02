package handlers_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// The tests below are split into two layers. This split mirrors the
// long-term testing strategy for the handlers package:
//
//   Layer 1 — DIRECT HANDLER TESTS  (Test_OnInstalling*_Direct)
//       Call the handler function directly with a hand-crafted World.
//       Assert on the returned Result (Next + Effects slice).
//       Fast, no goroutine, no ports involved. Best for enumerating
//       every branch of the handler's decision tree.
//
//   Layer 2 — END-TO-END TESTS      (TestInstall_EndToEnd_*)
//       Wire the full harness (Engine + Runner + fake ports) and
//       drive a real Reconcile. Assert on AMStore mutations,
//       EventRecorder emissions, and WorkloadOps calls. Best for
//       proving that the handler + Runner + table registration all
//       hang together, and for regression-testing the wiring.

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newInstallingAM(name string, payload map[string]string) *appv1.ApplicationManager {
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
			State:   appv1.Installing,
			OpType:  appv1.InstallOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

// worldFor is the Layer-1 shortcut for building a World with just an
// AM and Now.
func worldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// findEffect returns the first Effect in effs matching predicate.
// Helper to write concise assertions like
//
//	pa, ok := findEffect[domain.PatchAM](res.Effects)
//
// which returns the first PatchAM and true, or the zero PatchAM and
// false.
func findEffect[T domain.Effect](effs []domain.Effect) (T, bool) {
	var zero T
	for _, e := range effs {
		if v, ok := e.(T); ok {
			return v, true
		}
	}
	return zero, false
}

func assertTickRecovery(
	t *testing.T,
	handler func(context.Context, domain.World) domain.Result,
	world domain.World,
	want domain.CallWorkload,
) {
	t.Helper()

	world.WorkloadInFlight = true
	inFlight := handler(context.Background(), world)
	if inFlight.Err != nil {
		t.Fatalf("in-flight Err = %v, want nil", inFlight.Err)
	}
	if inFlight.Next != "" {
		t.Fatalf("in-flight Next = %q, want empty", inFlight.Next)
	}
	if len(inFlight.Effects) != 0 {
		t.Fatalf("in-flight Effects = %+v, want none", inFlight.Effects)
	}

	world.WorkloadInFlight = false
	recovered := handler(context.Background(), world)
	if recovered.Err != nil {
		t.Fatalf("recovery Err = %v, want nil", recovered.Err)
	}
	if recovered.Next != "" {
		t.Fatalf("recovery Next = %q, want empty", recovered.Next)
	}
	if len(recovered.Effects) != 1 {
		t.Fatalf("recovery Effects = %+v, want exactly one CallWorkload", recovered.Effects)
	}
	got, ok := recovered.Effects[0].(domain.CallWorkload)
	if !ok {
		t.Fatalf("recovery Effects[0] = %T, want CallWorkload", recovered.Effects[0])
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("recovery CallWorkload = %#v, want %#v", got, want)
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnInstallingTick_EmptyPhase_StartsExposePorts pins the fresh-
// entry branch: an AM whose Status.Payload has no installing.phase
// key must dispatch SetExposePorts first, matching installing_app.go.
func TestOnInstallingTick_EmptyPhase_StartsExposePorts(t *testing.T) {
	am := newInstallingAM("app1", nil)
	res := handlers.OnInstallingTick(context.Background(), worldFor(am))

	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in Installing while phase advances)", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect emitted")
	}
	if patch.Status.State != appv1.Installing {
		t.Fatalf("PatchAM.State = %q, want Installing", patch.Status.State)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingExposePorts {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingExposePorts)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect emitted")
	}
	if cw.Kind != domain.WorkloadSetExposePorts {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadSetExposePorts", cw.Kind)
	}
	if cw.OnSuccess == nil {
		t.Fatal("CallWorkload.OnSuccess is nil (violates CONTRACT.md rule 5)")
	}
	if cw.OnFailure == nil {
		t.Fatal("CallWorkload.OnFailure is nil (violates CONTRACT.md rule 5)")
	}

	// EmitEvent must accompany the phase flip so bfl users see the
	// state fan out.
	if _, ok := findEffect[domain.EmitEvent](res.Effects); !ok {
		t.Fatal("no EmitEvent effect emitted alongside CallWorkload dispatch")
	}
}

func TestOnInstallingTick_RecoversEveryPhaseWithoutStatusEffects(t *testing.T) {
	tests := []struct {
		phase string
		call  domain.CallWorkload
	}{
		{
			phase: handlers.PhaseInstallingExposePorts,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadSetExposePorts,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "SetExposePortsFailed"},
			},
		},
		{
			phase: handlers.PhaseInstallingValidate,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadValidateInstall,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "InstallValidationRejected"},
			},
		},
		{
			phase: handlers.PhaseInstallingHelmInstall,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadInstall,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "HelmInstallFailed"},
			},
		},
		{
			phase: handlers.PhaseInstallingPostValidate,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadValidateInstall,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "InstallValidationRejected"},
			},
		},
		{
			phase: handlers.PhaseInstallingScaleUp,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadResume,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "InstallScaleUpFailed"},
			},
		},
		{
			phase: handlers.PhaseInstallingWaitStartUp,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitStartUp,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: constants.AppStopDueToStartUpFailed},
			},
		},
		{
			phase: handlers.PhaseInstallingWaitLaunch,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitLaunch,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: constants.AppStopDueToStartUpFailed},
			},
		},
		{
			phase: handlers.PhaseInstallingCleanup,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadCleanupInstallFailure,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "InstallCleanupFailed"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			am := newInstallingAM("app1", map[string]string{
				handlers.PayloadInstallingPhase: tc.phase,
			})
			assertTickRecovery(t, handlers.OnInstallingTick, worldFor(am), tc.call)
		})
	}
}

// TestOnInstallingTick_UnknownPhase_HardFails pins the safety-net
// invariant: a bogus phase value in Payload is a bug — surface it as
// an Err so operators + test suites see it immediately instead of
// silently no-op'ing forever.
func TestOnInstallingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: "grabbed-by-alien",
	})
	res := handlers.OnInstallingTick(context.Background(), worldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnInstallingWorkloadReady_ValidatePhase_StartsHelm pins the
// validate → helm-install phase advance.
func TestOnInstallingWorkloadReady_ValidatePhase_StartsHelm(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingValidate,
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingHelmInstall {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingHelmInstall)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadInstall {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadInstall", cw.Kind)
	}
}

// TestOnInstallingWorkloadReady_HelmInstallPhase_StartsScaleUp pins
// the edge whose absence left two-phase apps deployed at replicas=0:
// a successful helm install MUST be followed by the Scale(-1) that
// restores the manifest-declared replica counts, not by a jump
// straight to Initializing.
//
// This is the single-phase case (World.HasWorkloadReplicas false),
// where the scale-up is a helm-side no-op but still dispatched — see
// startScaleUp's doc-comment for why the handler does not branch.
func TestOnInstallingWorkloadReady_HelmInstallPhase_StartsScaleUp(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingScaleUp {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingScaleUp)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadResume {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadResume (helm Scale(-1))", cw.Kind)
	}
}

// TestOnInstallingWorkloadReady_TwoPhase_HelmInstallStartsPostValidate
// pins the routing difference for an app that declares
// workloadReplicas: helm has deployed it at replicas=0, so the
// resource check runs NOW (it needs a real release to measure)
// rather than before the install.
func TestOnInstallingWorkloadReady_TwoPhase_HelmInstallStartsPostValidate(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
	})
	w := worldFor(am)
	w.HasWorkloadReplicas = true

	res := handlers.OnInstallingWorkloadReady(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingPostValidate {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingPostValidate)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadValidateInstall {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadValidateInstall", cw.Kind)
	}
}

// TestOnInstallingTick_TwoPhase_StillStartsExposePorts: two-phase
// apps also assign expose ports first; they skip pre-helm
// validation only after that step succeeds.
func TestOnInstallingTick_TwoPhase_StillStartsExposePorts(t *testing.T) {
	am := newInstallingAM("app1", nil)
	w := worldFor(am)
	w.HasWorkloadReplicas = true

	res := handlers.OnInstallingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingExposePorts {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingExposePorts)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadSetExposePorts {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadSetExposePorts", cw.Kind)
	}
}

func TestOnInstallingWorkloadReady_ExposePorts_StartsValidate(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingExposePorts,
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingValidate {
		t.Fatalf("phase = %q, want validate", got)
	}
	cw, _ := findEffect[domain.CallWorkload](res.Effects)
	if cw.Kind != domain.WorkloadValidateInstall {
		t.Fatalf("Kind = %q, want WorkloadValidateInstall", cw.Kind)
	}
}

func TestOnInstallingWorkloadReady_ExposePorts_TwoPhase_StartsHelm(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingExposePorts,
	})
	w := worldFor(am)
	w.HasWorkloadReplicas = true
	res := handlers.OnInstallingWorkloadReady(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingHelmInstall {
		t.Fatalf("phase = %q, want helm-install", got)
	}
	cw, _ := findEffect[domain.CallWorkload](res.Effects)
	if cw.Kind != domain.WorkloadInstall {
		t.Fatalf("Kind = %q, want WorkloadInstall", cw.Kind)
	}
}

// TestOnInstallingWorkloadReady_ScaleUpPhase_StartsWaitStartUp pins
// that a non-middleware app waits for its freshly-scaled pods to
// become ready before handing off.
func TestOnInstallingWorkloadReady_ScaleUpPhase_StartsWaitStartUp(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingScaleUp,
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingWaitStartUp {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingWaitStartUp)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadWaitStartUp {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadWaitStartUp", cw.Kind)
	}
}

// TestOnInstallingWorkloadReady_Middleware_ScaleUpStartsWaitLaunch
// pins the middleware branch: it never passes through Initializing,
// so the launch wait happens here and Running follows directly.
func TestOnInstallingWorkloadReady_Middleware_ScaleUpStartsWaitLaunch(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingScaleUp,
	})
	am.Spec.Type = appv1.Middleware

	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != handlers.PhaseInstallingWaitLaunch {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingPhase, got, handlers.PhaseInstallingWaitLaunch)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadWaitLaunch {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadWaitLaunch", cw.Kind)
	}
}

// TestOnInstallingWorkloadReady_Middleware_WaitLaunchLandsRunning
// pins the middleware terminal: Running, with the terminal OpRecord,
// no Initializing hop.
func TestOnInstallingWorkloadReady_Middleware_WaitLaunchLandsRunning(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitLaunch,
	})
	am.Spec.Type = appv1.Middleware

	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Running {
		t.Fatalf("Next = %q, want Running", res.Next)
	}
	rec, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord effect")
	}
	if rec.Record.Status != appv1.Running {
		t.Fatalf("OpRecord.Status = %q, want Running", rec.Record.Status)
	}
}

// TestOnInstallingWorkloadReady_WaitStartUpPhase_FinishesInstall pins
// the terminal branch: once pods are ready, the state MUST advance to
// Initializing with an OpRecord written.
func TestOnInstallingWorkloadReady_WaitStartUpPhase_FinishesInstall(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitStartUp,
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Initializing {
		t.Fatalf("Next = %q, want Initializing", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Initializing {
		t.Fatalf("PatchAM.State = %q, want Initializing", patch.Status.State)
	}
	if got, ok := patch.Status.Payload[handlers.PayloadInstallingPhase]; ok {
		t.Fatalf("Payload[%q] still %q after leaving Installing; must be cleared", handlers.PayloadInstallingPhase, got)
	}

	rec, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord effect")
	}
	if rec.Record.OpID != "op-42" {
		t.Fatalf("OpRecord.OpID = %q, want op-42", rec.Record.OpID)
	}
	if rec.Record.Status != appv1.Initializing {
		t.Fatalf("OpRecord.Status = %q, want Initializing", rec.Record.Status)
	}
}

// TestOnInstallingWorkloadReady_EmptyPhase_HardFails pins that a
// WorkloadReady with no phase context is treated as a bug (either
// a Runner mis-route or a race against a pre-phase-tracking binary).
func TestOnInstallingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newInstallingAM("app1", nil)
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

// TestOnInstallingWorkloadFailed_UsesActionData pins the primary
// error path: the WorkloadFailed action's Reason/Message are threaded
// through World.Action and stashed on the InstallingCanceling handoff,
// which is what the eventual InstallFailed PatchAM + OpRecord +
// EmitEvent are written from.
func TestOnInstallingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingValidate,
	})
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "AppUnschedulable",
		Message: "no GPU available",
	}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %q, want InstallingCanceling", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	// The Installing sub-phase is blanked so it cannot survive the
	// key-by-key Payload merge into InstallingCanceling.
	if got := patch.Status.Payload[handlers.PayloadInstallingPhase]; got != "" {
		t.Fatalf("phase = %q, want cleared on leaving Installing", got)
	}
	if patch.Status.Payload[handlers.PayloadInstallingFailMessage] != "no GPU available" {
		t.Fatalf("stashed message = %q", patch.Status.Payload[handlers.PayloadInstallingFailMessage])
	}
	if patch.Status.Payload[handlers.PayloadInstallingFailReason] != "AppUnschedulable" {
		t.Fatalf("stashed reason = %q", patch.Status.Payload[handlers.PayloadInstallingFailReason])
	}
}

// TestInstallingFailure_EntersInstallingCancelingWithInstallFailedIntent
// pins the shared cleanup pipeline: an ordinary install failure no
// longer tears the release down in-state, it hands off to
// InstallingCanceling with a payload marker saying the pipeline must
// land in InstallFailed rather than the user-cancel terminal.
func TestInstallingFailure_EntersInstallingCancelingWithInstallFailedIntent(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
	})
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "HelmError",
		Message: "chart apply failed",
	}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %q, want InstallingCanceling", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if patch.Status.State != appv1.InstallingCanceling {
		t.Fatalf("PatchAM.State = %q, want InstallingCanceling", patch.Status.State)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; got != string(appv1.InstallFailed) {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInstallingCancelingTerminal, got, appv1.InstallFailed)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingFailReason]; got != "HelmError" {
		t.Fatalf("stashed reason = %q, want HelmError", got)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingFailMessage]; got != "chart apply failed" {
		t.Fatalf("stashed message = %q, want %q", got, "chart apply failed")
	}
	// The InstallingCanceling Tick owns the uninstall dispatch; the
	// handoff must not race it with a second claim.
	if cw, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatalf("unexpected CallWorkload %+v on the InstallingCanceling handoff", cw)
	}
}

func TestOnInstallingWorkloadReady_Cleanup_LandsInstallFailed(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase:       handlers.PhaseInstallingCleanup,
		handlers.PayloadInstallingFailMessage: "no GPU available",
		handlers.PayloadInstallingFailReason:  "AppUnschedulable",
	})
	res := handlers.OnInstallingWorkloadReady(context.Background(), worldFor(am))
	if res.Next != appv1.InstallFailed {
		t.Fatalf("Next = %q, want InstallFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.InstallFailed {
		t.Fatalf("PatchAM.State = %q, want InstallFailed", patch.Status.State)
	}
	if patch.Status.Message != "no GPU available" {
		t.Fatalf("PatchAM.Message = %q", patch.Status.Message)
	}
	if patch.Status.Reason != "AppUnschedulable" {
		t.Fatalf("PatchAM.Reason = %q", patch.Status.Reason)
	}
	if patch.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanupDone {
		t.Fatal("cleanup not marked done on InstallFailed")
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on InstallFailed")
	}
}

// TestOnInstallingWorkloadFailed_UnschedulablePods_LandsStopping
// pins the routing hint the helmops adapter sets when helm failed
// because pods could not be scheduled. This must NOT become an
// InstallFailed: the release is deployed, and tearing it down would
// force the user to reinstall instead of freeing capacity and
// resuming.
//
// stopAll widens the eventual stop to the shared server the app
// owns, so it has to reach the AM's annotations before the status
// flip — which means a PatchSpec ahead of the PatchAM.
func TestOnInstallingWorkloadFailed_UnschedulablePods_LandsStopping(t *testing.T) {
	for _, tc := range []struct {
		name       string
		stopAll    bool
		wantAnnots bool
	}{
		{name: "app-side pods", stopAll: false, wantAnnots: false},
		{name: "server-side pods", stopAll: true, wantAnnots: true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newInstallingAM("app1", map[string]string{
				handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
			})
			w := worldFor(am)
			w.Action = domain.WorkloadFailed{
				Reason:      constants.AppUnschedulable,
				Message:     "pod is pending",
				LandedState: appv1.Stopping,
				StopAll:     tc.stopAll,
			}

			res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != appv1.Stopping {
				t.Fatalf("Next = %q, want Stopping", res.Next)
			}

			patch, _ := findEffect[domain.PatchAM](res.Effects)
			if patch.Status.Reason != constants.AppUnschedulable {
				t.Fatalf("PatchAM.Reason = %q, want %q", patch.Status.Reason, constants.AppUnschedulable)
			}
			// This path does not convert the operation — legacy
			// leaves Spec.OpType as the install.
			if patch.Status.OpType != "" {
				t.Fatalf("PatchAM.OpType = %q, want unset (install is not converted here)", patch.Status.OpType)
			}

			spec, ok := findEffect[domain.PatchSpec](res.Effects)
			if ok != tc.wantAnnots {
				t.Fatalf("PatchSpec present = %v, want %v", ok, tc.wantAnnots)
			}
			if !tc.wantAnnots {
				return
			}
			if spec.SetAnnotations[handlers.AnnotStopAll] != "true" {
				t.Fatalf("PatchSpec.SetAnnotations = %v, want %s=true", spec.SetAnnotations, handlers.AnnotStopAll)
			}
			if spec.OpType != "" {
				t.Fatalf("PatchSpec.OpType = %q, want unset", spec.OpType)
			}
			if _, isSpec := res.Effects[0].(domain.PatchSpec); !isSpec {
				t.Fatalf("Effects[0] = %T, want PatchSpec (must precede PatchAM; see domain.PatchSpec ordering note)", res.Effects[0])
			}
		})
	}
}

// TestOnInstallingWorkloadFailed_PostValidateRejection_ConvertsToStop
// pins the post-helm rejection path. On top of landing in Stopping,
// this one rewrites the operation itself: an install that is not
// going to finish should stop advertising itself as in progress, so
// both Spec.OpType and Status.OpType become Stop.
func TestOnInstallingWorkloadFailed_PostValidateRejection_ConvertsToStop(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingPostValidate,
	})
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  constants.AppUnschedulable,
		Message: "insufficient GPU after helm render",
	}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}

	spec, ok := findEffect[domain.PatchSpec](res.Effects)
	if !ok {
		t.Fatal("no PatchSpec effect; the operation must be converted to a stop")
	}
	if spec.OpType != appv1.StopOp {
		t.Fatalf("PatchSpec.OpType = %q, want StopOp", spec.OpType)
	}
	if _, isSpec := res.Effects[0].(domain.PatchSpec); !isSpec {
		t.Fatalf("Effects[0] = %T, want PatchSpec (must precede PatchAM)", res.Effects[0])
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.OpType != appv1.StopOp {
		t.Fatalf("PatchAM.OpType = %q, want StopOp", patch.Status.OpType)
	}
	if patch.Status.Message != "insufficient GPU after helm render" {
		t.Fatalf("PatchAM.Message = %q, want the rejection message", patch.Status.Message)
	}
}

// TestOnInstallingWorkloadFailed_WaitStartUpTimeout_LandsStopping
// pins the readiness-timeout path: pods exist but never became
// ready, which legacy reports as StartUpFailed and drives to
// Stopping without converting the operation.
func TestOnInstallingWorkloadFailed_WaitStartUpTimeout_LandsStopping(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitStartUp,
	})
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{Message: "timed out waiting for pods"}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != constants.AppStopDueToStartUpFailed {
		t.Fatalf("PatchAM.Reason = %q, want %q", patch.Status.Reason, constants.AppStopDueToStartUpFailed)
	}
	if _, ok := findEffect[domain.PatchSpec](res.Effects); ok {
		t.Fatal("unexpected PatchSpec; a startup timeout does not convert the operation")
	}
}

// TestOnInstallingWorkloadFailed_MiddlewareLaunchTimeout pins the
// one failure that unwinds rather than stops: a middleware that
// never launched has to have its release rolled back, and
// InstallingCanceling is the handler that owns that.
func TestOnInstallingWorkloadFailed_MiddlewareLaunchTimeout(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitLaunch,
	})
	am.Spec.Type = appv1.Middleware
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{Message: "launch timed out"}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %q, want InstallingCanceling", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != constants.AppStopDueToStartUpFailed {
		t.Fatalf("PatchAM.Reason = %q, want %q", patch.Status.Reason, constants.AppStopDueToStartUpFailed)
	}
}

// TestOnInstallingWorkloadFailed_ScaleUpFailure_CleansUpToInstallFailed
// pins that a failed Scale(-1) is a plain install failure: unlike
// the unschedulable-pod and rejection paths, there is no partial
// success worth preserving for a resume, so the release is torn down
// on the way to InstallFailed.
func TestOnInstallingWorkloadFailed_ScaleUpFailure_CleansUpToInstallFailed(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingScaleUp,
	})
	w := worldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %q, want InstallingCanceling", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; got != string(appv1.InstallFailed) {
		t.Fatalf("terminal intent = %q, want InstallFailed", got)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingFailReason]; got != "InstallScaleUpFailed" {
		t.Fatalf("stashed reason = %q, want InstallScaleUpFailed", got)
	}
}

// TestOnInstallingWorkloadFailed_FallsBackWhenActionEmpty pins the
// fallback: if the Action carries no failure detail (empty Reason /
// Message, or a non-WorkloadFailed action), the handler synthesizes
// a phase-derived pair so we always write SOMETHING to InstallFailed.
func TestOnInstallingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newInstallingAM("app1", map[string]string{
		handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall,
	})
	w := worldFor(am)
	// Empty WorkloadFailed — Runner would produce this if the
	// underlying WorkloadResult had no LandedReason / err.
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnInstallingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; got != string(appv1.InstallFailed) {
		t.Fatalf("terminal intent = %q, want InstallFailed", got)
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingFailReason]; got != "HelmInstallFailed" {
		t.Fatalf("stashed reason = %q, want HelmInstallFailed", got)
	}
	if patch.Status.Payload[handlers.PayloadInstallingFailMessage] == "" {
		t.Fatal("stashed message empty; expected phase-derived string")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// e2eHarness bundles the same wiring as runner_test.go but with the
// real installing handler registered on the table. Kept as a package-
// private helper here so runner_test.go stays untouched.
type e2eHarness struct {
	t           *testing.T
	amStore     *fake.AMStore
	workloadOps *fake.WorkloadOps
	validator   *fake.InstallValidator
	imagePuller *fake.ImagePuller
	recorder    *fake.EventRecorder
	clock       *fake.Clock
	requeuer    *runner.RecordingRequeuer

	table  *domain.TransitionTable
	engine *domain.Engine
	runner *runner.Runner
}

func newE2EHarness(t *testing.T) *e2eHarness {
	t.Helper()
	table := domain.NewDefaultTable()
	handlers.Register(table)

	engine := domain.NewEngine(table)
	h := &e2eHarness{
		t:           t,
		amStore:     fake.NewAMStore(),
		workloadOps: &fake.WorkloadOps{},
		validator:   &fake.InstallValidator{},
		imagePuller: fake.NewImagePuller(),
		recorder:    &fake.EventRecorder{},
		clock:       fake.NewClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)),
		requeuer:    &runner.RecordingRequeuer{},
		table:       table,
		engine:      engine,
	}
	r, err := runner.New(runner.Config{
		Engine:      engine,
		AMStore:     h.amStore,
		WorkloadOps: h.workloadOps,
		Validator:   h.validator,
		ImagePuller: h.imagePuller,
		Recorder:    h.recorder,
		Clock:       h.clock,
		Requeuer:    h.requeuer,
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	h.runner = r
	t.Cleanup(r.Stop)
	return h
}

// seedInstalling seeds an AM in Installing state (with empty phase).
func (h *e2eHarness) seedInstalling(name string) string {
	am := newInstallingAM(name, nil)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestInstall_EndToEnd_HappyPath drives the full install lifecycle:
//
//	R1: Installing (phase="")             — dispatches SetExposePorts
//	R2: Installing (phase="expose-ports") — ports ready → dispatches ValidateInstall
//	R3: Installing (phase="validate")     — validate ready → dispatches Install
//	R4: Installing (phase="helm-install") — install ready → dispatches Resume (Scale(-1))
//	R5: Installing (phase="scale-up")     — scale ready → dispatches WaitForStartUp
//	R6: Installing (phase="wait-startup") — pods ready → advances to Initializing
//
// It's the "load-bearing" test for the whole ports+runner+handler
// stack — if it goes green, the pieces are wired.
func TestInstall_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInstalling("myapp")

	// R1 — dispatches SetExposePorts.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingExposePorts {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseInstallingExposePorts)
	}

	// R2 — expose-ports ready → ValidateInstall.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingValidate {
		t.Fatalf("after R2, phase = %q, want %q", p, handlers.PhaseInstallingValidate)
	}

	// R3 — validate ready → Install.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingHelmInstall {
		t.Fatalf("after R3, phase = %q, want %q", p, handlers.PhaseInstallingHelmInstall)
	}
	if wc := h.workloadOps.Calls(); len(wc) == 0 || wc[len(wc)-1].Method != "Install" {
		t.Fatalf("after R3, workload calls = %+v, want last method Install", wc)
	}

	// R4 — install ready → scale-up.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingScaleUp {
		t.Fatalf("after R4, phase = %q, want %q", p, handlers.PhaseInstallingScaleUp)
	}
	if wc := h.workloadOps.Calls(); len(wc) == 0 || wc[len(wc)-1].Method != "Resume" {
		t.Fatalf("after R4, workload calls = %+v, want last method Resume (helm Scale(-1))", wc)
	}

	// R5 — scale ready → wait-startup.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != handlers.PhaseInstallingWaitStartUp {
		t.Fatalf("after R5, phase = %q, want %q", p, handlers.PhaseInstallingWaitStartUp)
	}
	if wc := h.workloadOps.Calls(); len(wc) == 0 || wc[len(wc)-1].Method != "WaitForStartUp" {
		t.Fatalf("after R5, workload calls = %+v, want last method WaitForStartUp", wc)
	}

	// R6 — pods ready → Initializing.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after R6, state = %q, want Initializing", got.Status.State)
	}

	// A successful OpRecord landed.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecord after happy path finish")
	}
	if got.Status.OpRecords[0].Status != appv1.Initializing {
		t.Fatalf("head OpRecord.Status = %q, want Initializing", got.Status.OpRecords[0].Status)
	}
}

// TestInstall_EndToEnd_ValidatorRejection pins the validator-reject
// path: the validator returns Decision{OK:false, Reason,...}, the
// Runner unifies it into WorkloadFailed, and the handler lands in
// InstallFailed with the rejection Reason preserved.
func TestInstall_EndToEnd_ValidatorRejection(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInstalling("myapp")

	h.validator.ValidateInstallFunc = func(_ context.Context, _ ports.ValidationInput) (ports.ValidationDecision, error) {
		return ports.ValidationDecision{
			OK:        false,
			Reason:    "AppUnschedulable",
			Message:   "insufficient CPU",
			Validator: "cluster-capacity",
		}, nil
	}

	// R1 expose-ports, R2 validate (rejects), R3 hand off to
	// InstallingCanceling, R4 helm uninstall, R5 wait-ns, R6 land
	// InstallFailed.
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
	if got.Status.State != appv1.InstallingCanceling {
		t.Fatalf("after R3, state = %q, want InstallingCanceling", got.Status.State)
	}
	if p := got.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; p != string(appv1.InstallFailed) {
		t.Fatalf("after R3, terminal intent = %q, want InstallFailed", p)
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
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallFailed {
		t.Fatalf("state = %q, want InstallFailed", got.Status.State)
	}
	// The release was torn down by the shared cancel pipeline.
	var uninstallCalls int
	for _, c := range h.workloadOps.Calls() {
		if c.Method == "Uninstall" {
			uninstallCalls++
		}
	}
	if uninstallCalls == 0 {
		t.Fatal("WorkloadOps.Uninstall was never called; the failed install was not torn down")
	}
	if got.Status.Reason != "AppUnschedulable" {
		t.Fatalf("Reason = %q, want AppUnschedulable (preserved from validator decision)", got.Status.Reason)
	}
	for _, staleKey := range []string{
		handlers.PayloadInstallingCancelingPhase,
		handlers.PayloadInstallingCancelingTerminal,
		handlers.PayloadInstallingFailMessage,
		handlers.PayloadInstallingFailReason,
	} {
		if _, ok := got.Status.Payload[staleKey]; ok {
			t.Fatalf("InstallFailed retained stale Payload[%q]: %+v", staleKey, got.Status.Payload)
		}
	}

	// Warning event was emitted. Events are stamped with the AM
	// state captured at Reconcile START, so the failure event
	// bears State=Installing (the "we entered failure DURING the
	// Installing state" reading) — the UI keys off Reason to know
	// it's a failure. Assert by Reason, not by State.
	var found *ports.AppLifecycleEvent
	for _, ev := range h.recorder.Events() {
		ev := ev // shadow for pointer capture
		if ev.Reason == "AppUnschedulable" {
			found = &ev
			break
		}
	}
	if found == nil {
		t.Fatalf("no event with Reason=AppUnschedulable; got: %+v", h.recorder.Events())
	}
	if found.Message == "" {
		t.Fatalf("event Message empty; expected validator's rejection message; got: %+v", *found)
	}
}

// TestInstall_EndToEnd_HelmFailure pins the helm-install failure
// path: WorkloadOps.Install returns an error, the Runner packs it
// into WorkloadFailed, and the handler transitions to InstallFailed.
func TestInstall_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInstalling("myapp")
	h.workloadOps.InstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("chart apply failed")}
	}

	// R1 expose-ports, R2 validate, R3 helm (fails), R4 hand off to
	// InstallingCanceling, R5 helm uninstall, R6 wait-ns, R7
	// InstallFailed.
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
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InstallFailed {
		t.Fatalf("state = %q, want InstallFailed", got.Status.State)
	}
	// The original helm diagnostic survived the cleanup round-trip.
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError (stashed across the cleanup)", got.Status.Reason)
	}
}

// drainAndRebuild waits for any Runner-owned goroutine to complete
// (so the delivered Action is in the pending queue), then rebuilds
// the Runner in place so ctx cancellation doesn't affect the next
// Reconcile. Same trick as runner_test.go's callback-consumption
// case — factored into a helper because every e2e test uses it.
func drainAndRebuild(t *testing.T, h *e2eHarness) {
	t.Helper()
	awaitDispatchers(t, h)
	h.runner.Stop()
	r, err := runner.New(runner.Config{
		Engine: h.engine, AMStore: h.amStore, WorkloadOps: h.workloadOps,
		Validator: h.validator, ImagePuller: h.imagePuller,
		Recorder: h.recorder, Clock: h.clock,
		Requeuer: h.requeuer, Pending: h.runner.Pending(),
	})
	if err != nil {
		t.Fatalf("rebuild runner: %v", err)
	}
	h.runner = r
	t.Cleanup(r.Stop)
}

// driveInstallToInitializing reconciles an AM that has just entered
// Installing (no phase stamped) until it lands in Initializing,
// walking the whole pipeline:
//
//	validate → helm-install → scale-up → wait-startup
//
// Each phase dispatches exactly one CallWorkload, so each Reconcile
// is followed by a drain that lets the callback reach the pending
// queue before the next one consumes it. The final Reconcile
// consumes the readiness callback and leaves Installing.
//
// Shared by the lifecycle tests that only pass THROUGH the install
// pipeline on their way to the state they actually exercise; tests
// asserting on the pipeline itself (see
// TestInstall_EndToEnd_HappyPath) still step through it explicitly.
func driveInstallToInitializing(t *testing.T, h *e2eHarness, name string) {
	t.Helper()
	for i, phase := range installingDispatchPhases {
		if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: name}); err != nil {
			t.Fatalf("installing round %d (dispatching %s): %v", i+1, phase, err)
		}
		drainAndRebuild(t, h)

		got, err := h.amStore.Get(context.Background(), name)
		if err != nil {
			t.Fatalf("installing round %d: get %s: %v", i+1, name, err)
		}
		if p := got.Status.Payload[handlers.PayloadInstallingPhase]; p != phase {
			t.Fatalf("installing round %d: phase = %q, want %q", i+1, p, phase)
		}
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: name}); err != nil {
		t.Fatalf("installing finish: %v", err)
	}
}

// installingDispatchPhases is the phase an AM carries after each of
// the Reconciles driveInstallToInitializing performs, in order. Only
// covers the single-phase (no workloadReplicas) route, which is what
// the fake harness produces since its AMs carry no Spec.Config.
var installingDispatchPhases = []string{
	handlers.PhaseInstallingExposePorts,
	handlers.PhaseInstallingValidate,
	handlers.PhaseInstallingHelmInstall,
	handlers.PhaseInstallingScaleUp,
	handlers.PhaseInstallingWaitStartUp,
}

// awaitDispatchers blocks until every launched CallWorkload goroutine
// has finished, driving the fake clock so timer-based dispatchers can
// make progress.
//
// The image pull polls ImageManager on a Clock ticker until it
// reports a terminal state, so a Stop() issued while it waits would
// cancel the pull rather than complete it. InFlight is incremented
// synchronously by the Runner, so it is already non-zero here for a
// Reconcile that dispatched — no startup race to sleep around.
func awaitDispatchers(t *testing.T, h *e2eHarness) {
	t.Helper()
	for i := 0; i < 2000; i++ {
		if h.runner.InFlight() == 0 {
			return
		}
		if h.clock.Pending() > 0 {
			h.clock.Advance(2 * imagePullPollBeat)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("dispatchers never drained; %d still in flight", h.runner.InFlight())
}

// parkedWorkload returns a WorkloadOps hook that blocks until
// release() is called. started closes once the goroutine has
// entered the hook, i.e. the AM's workload claim is held.
func parkedWorkload() (started <-chan struct{}, release func(), fn func(context.Context, ports.WorkloadRequest) ports.WorkloadResult) {
	st := make(chan struct{})
	rel := make(chan struct{})
	var once sync.Once
	fn = func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
		once.Do(func() { close(st) })
		<-rel
		return ports.WorkloadResult{}
	}
	return st, func() { close(rel) }, fn
}

func awaitChan(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// imagePullPollBeat mirrors the Runner's unexported poll interval.
const imagePullPollBeat = time.Second

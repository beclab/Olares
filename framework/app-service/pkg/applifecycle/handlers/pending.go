package handlers

import (
	"context"
	"fmt"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// Package-local mirrors of the legacy pending-state messages. See
// installing.go for the rationale behind holding these locally
// rather than importing pkg/constants — they're short strings and
// keeping them inline makes review-diffing against
// pkg/appstate/pending_app.go one-shot.
const (
	pendingMsgStartDownload     = "start to download"
	pendingMsgBedFull           = "waiting: another app is currently downloading"
	pendingMsgWaitingForInstall = "waiting for install"
	pendingReasonAdvance        = "Downloading"
	pendingReasonWaiting        = "WaitingForDownloadingSlot"
)

// pendingBedFullBackoff is the requeue delay handed to ScheduleTick
// when the cluster-wide Downloading slot is occupied. Deliberately
// short (1s) so the wait is user-imperceptible; the informer cache
// backing ClusterStats.DownloadingCount typically ticks every few
// hundred milliseconds anyway.
//
// A longer backoff would trade latency for a smaller reconcile
// storm, but the "at most 1 Downloading" gate means we're already
// looking at a fleet of at most ~a handful of Pending AMs at once —
// the reconcile rate isn't a real bottleneck.
const pendingBedFullBackoff = 1 * time.Second

// pendingDownloadingSlotSize is the maximum number of AMs allowed to
// be in the Downloading state cluster-wide. Legacy
// pkg/appstate/pending_app.go hard-codes 1 via the `count < 1`
// check. Kept as a named constant here so a future PR that widens
// the parallelism (or wires it to a Config knob) has exactly one
// place to touch.
const pendingDownloadingSlotSize = 1

// registerPending wires the Pending-state handlers into the
// TransitionTable. Same pattern as registerInstalling — see that
// function's comment for the SetOrAddHandler vs Add rationale.
//
// Wiring semantics:
//
//   - (InitialState, TriggerOp{InstallOp}) is the fresh-install
//     bootstrap. The placeholder only advanced State; OnInitialInstall
//     restores the pre-refactor Pending status and its event.
//   - (Pending, Tick{}) already has a placeholder targeting
//     Downloading. actionForEdge classifies that edge as Tick
//     because Pending only waits for the download slot. We replace
//     the placeholder handler; the reachable edges are unchanged.
//   - AddReachable(Pending, Pending) declares the self-loop we
//     take on "bed full" (handler returns Result{} with no Next).
//     Currently Result{Next:""} is treated as "stay in place" by
//     Engine.Step, so this AddReachable is defensive documentation
//     rather than a hard requirement — but it matches
//     registerInstalling's pattern and keeps the two handlers
//     symmetric.
//
// The cancel path is active, but is registered with the rest of its
// pipeline in cancel_pipeline.go: registerPendingCanceling replaces
// (Pending, TriggerOp{Op:CancelOp}) with OnPendingCancel, then drives
// PendingCanceling through its namespace-wait workload to
// PendingCanceled or PendingCancelFailed.
func registerPending(t *domain.TransitionTable) {
	t.AddReachable(appv1.Pending, appv1.Pending)
	t.SetOrAddHandler(domain.InitialState, domain.TriggerOp{Op: appv1.InstallOp}, OnInitialInstall)
	t.SetOrAddHandler(appv1.Pending, domain.Tick{}, OnPendingTick)
}

// OnInitialInstall is the fresh-install bootstrap: "" → Pending.
//
// Legacy correspondence: handler_installer_install.go wrote this
// status itself (Reason=pending, Message="waiting for install").
// The retired ApplicationManager controller then published it
// because Status.State changed. The Runner only publishes
// EmitEvent, so the placeholder that only set State dropped the
// event bfl waits on.
func OnInitialInstall(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("initial: install handler invoked with nil AM"),
		}
	}
	reason := appv1.Pending.String()
	return domain.Result{
		Next: appv1.Pending,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Pending,
				Reason:  reason,
				Message: pendingMsgWaitingForInstall,
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: pendingMsgWaitingForInstall,
			},
		},
	}
}

// OnPendingTick is the reconcile-driven handler for the Pending
// state. It advances Pending → Downloading whenever the cluster-wide
// single-slot download gate is free, and otherwise returns a
// ScheduleTick so the next reconcile will re-check.
//
// Legacy correspondence: pkg/appstate/pending_app.go:63-112 —
// specifically the `addLimitedStatefulApp` call, its check-func
// (list AMs, count Downloading, return count < 1) and its add-func
// (bump State + StatusTime + Reason + Message + OpGeneration).
//
// CONTRACT NOTES
//
//   - The legacy code enforces "at most 1 concurrent Downloading"
//     via `factory.addLimitedStatefulApp`, which acquires a
//     process-wide mutex around the (list, check, add) sequence.
//     This handler replaces that with a declarative read of
//     World.ClusterStats.DownloadingCount populated by
//     Runner.computeClusterStats. The mutex is GONE — coordination
//     now flows through the AMStore + informer cache.
//
//   - Slot occupancy is DownloadingCount; slot *order* is
//     ClusterStats.NextDownloading (oldest pending.queuedAt).
//     A later-queued AM whose Tick fires first still waits.
//
//   - Two replicas with a stale informer can still both observe
//     an empty slot; that TOCTOU matches the previous design.
//     In-process MaxConcurrentReconciles=1 plus FIFO head
//     selection is enough for a single app-service.
//
// SKELETON SCOPE — deferred to follow-up PRs:
//
//   - AppExists guard: legacy code detects a leaked Application CR
//     paired with a Pending AM as an "unknown state" and schedules
//     a namespace-delete cleanup. This handler surfaces that
//     invariant violation as an Err so tests catch it, but does
//     NOT emit the cleanup Effect — that needs a new
//     WorkloadOps.RemoveUnknownApplication port method.
func OnPendingTick(_ context.Context, w domain.World) domain.Result {
	if w.AppExists {
		// A Pending AM whose Application CR already exists means
		// helm has RAN for this app but the AM never advanced past
		// Pending — a torn state. Legacy code cleans it up. We
		// surface it as an error so operators see it; the cleanup
		// Effect is a follow-up PR (see scope note above).
		return domain.Result{
			Err: fmt.Errorf("pending: Application CR already exists for %q; the paired helm release was not cleaned up (skeleton scope does not emit cleanup Effect yet)",
				w.AM.Name),
		}
	}

	// Cluster-wide Downloading slot check. The current AM is
	// Pending (not Downloading), so ClusterStats.DownloadingCount
	// counts only OTHER AMs and the guard collapses to a bare
	// slot-size comparison.
	if w.ClusterStats.DownloadingCount >= pendingDownloadingSlotSize {
		return waitForSlot(w)
	}

	// FIFO: even when the slot is free, only the oldest Pending AM
	// (ClusterStats.NextDownloading) may enter Downloading. An empty
	// NextDownloading is treated as "no ordering info" so hand-crafted
	// unit Worlds that only set DownloadingCount keep working.
	if head := w.ClusterStats.NextDownloading; head != "" && w.AM.Name != head {
		return waitForSlot(w)
	}

	return advanceToDownloading(w)
}

// advanceToDownloading is the effect bundle for the "slot free"
// branch. Kept as a helper so tests can hand-construct the exact
// same Result to compare against.
//
// The PatchAM must set the AM's Status.State to Downloading; the
// Reason/Message match the legacy strings from pending_app.go so
// operator playbooks + UI translations keep working. OpGeneration
// is intentionally NOT bumped here — the Runner is the right owner
// of that field (it needs to bump on every reconcile that emits an
// operation-scoped PatchAM, not only in this handler). Follow-up
// PR when the Runner grows the OpGeneration-bump helper.
func advanceToDownloading(w domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Downloading,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Downloading,
				Reason:  pendingReasonAdvance,
				Message: pendingMsgStartDownload,
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  pendingReasonAdvance,
				Message: pendingMsgStartDownload,
			},
		},
	}
}

// waitForSlot is the effect bundle for the "bed full" branch.
// Returns Result{Next:""} so the state remains Pending, plus a
// ScheduleTick that re-invokes Reconcile after
// pendingBedFullBackoff. Also emits a small PatchAM updating
// Reason/Message so operators inspecting the AM see WHY it's
// stuck at Pending.
//
// Note: we deliberately do NOT emit an EmitEvent here. The
// AppLifecycleEvent stream is high-signal; emitting one per waited
// tick would generate one event per second per waiting AM. The
// Status.Reason field is the correct place to surface "we're
// waiting" — operators tail AM status; they don't tail events for
// this signal.
func waitForSlot(w domain.World) domain.Result {
	return domain.Result{
		// Next: "" means "stay in current state". Engine.Step
		// treats that as a valid no-op — it does NOT check the
		// AddReachable table for "stay-in-place", so we don't
		// need to worry about (Pending, Pending) missing from
		// the reachable set.
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Pending,
				Reason:  pendingReasonWaiting,
				Message: pendingMsgBedFull,
			}},
			domain.ScheduleTick{After: pendingBedFullBackoff},
		},
	}
}

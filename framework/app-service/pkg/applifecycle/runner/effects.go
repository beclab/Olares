package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/google/uuid"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// dispatchResult applies every Effect in res.Effects, in order, and
// returns the aggregated ReconcileOutcome describing what the caller
// (Runner.Reconcile) should return to controller-runtime.
//
// Two Effects need special handling:
//
//   - AppendOpRecord is buffered and merged into the immediately-
//     following PatchAM so the OpRecord write is atomic with the
//     status update, matching legacy makeRecord + updateStatus
//     behaviour.
//   - CallWorkload is spawned as a Runner-owned goroutine; the call
//     itself is non-blocking. Only ONE CallWorkload per AM should
//     be in flight at a time — the handler contract enforces this
//     (rule 4: at most one long-running effect per Result).
//
// dispatchResult also synthesises a PatchAM effect from res.Next +
// res.Message + res.Reason if no explicit PatchAM was returned. This
// is the "convenience path" table_default.go's placeholder handler
// relies on.
func (r *Runner) dispatchResult(ctx context.Context, world domain.World, res domain.Result) (ReconcileOutcome, error) {
	effects := ensureStatePatch(world, res)
	effects = stampTriggerOperation(effects, world)

	var out ReconcileOutcome
	var pendingOpRecord *appv1.OpRecord

	// Running view of the AM status as the effect list mutates it.
	// Handlers bundle PatchAM followed by EmitEvent to announce the
	// transition they just made, so the event has to read the patched
	// values — World.AM is the pre-Reconcile snapshot and would
	// report the state and progress of the *previous* operation.
	var status *appv1.ApplicationManagerStatus
	if world.AM != nil {
		status = world.AM.Status.DeepCopy()
	}

	// Both AM writes carry an optimistic-lock precondition on the
	// ResourceVersion, and a spec write bumps it. Track the version
	// across the effect list so a PatchSpec followed by a PatchAM
	// (the post-helm-rejection bundle) does not make the status write
	// fail a precondition the Runner itself invalidated.
	rv := ""
	if world.AM != nil {
		rv = world.AM.ResourceVersion
	}

	for _, eff := range effects {
		switch e := eff.(type) {

		case domain.AppendOpRecord:
			// Buffer; will be merged into the next PatchAM (or
			// flushed as its own PatchAM at the end if no PatchAM
			// follows).
			rec := e.Record
			pendingOpRecord = &rec

		case domain.PatchSpec:
			newRV, err := r.applyPatchSpec(ctx, world, rv, e)
			if err != nil {
				return out, err
			}
			rv = newRV

		case domain.PatchAM:
			if err := r.applyPatchAM(ctx, world, rv, e, pendingOpRecord); err != nil {
				return out, err
			}
			pendingOpRecord = nil
			advanceStatusView(status, e)

		case domain.EmitEvent:
			if err := r.applyEmitEvent(ctx, world, status, e); err != nil {
				return out, err
			}

		case domain.ScheduleTick:
			out.RequeueAfter = minPositiveDuration(out.RequeueAfter, e.After)

		case domain.CallWorkload:
			if err := r.launchCallWorkload(ctx, world, e); err != nil {
				return out, err
			}

		case domain.CancelInFlight:
			if world.AM != nil {
				r.workloads.cancelInFlight(world.AM.Name)
			}

		default:
			return out, fmt.Errorf("runner: unknown Effect type %T", eff)
		}
	}

	// If a handler emitted AppendOpRecord without a following
	// PatchAM (rare, but legal), flush it as a standalone patch so
	// the OpRecord reaches etcd.
	if pendingOpRecord != nil {
		if err := r.applyPatchAM(ctx, world, rv, domain.PatchAM{}, pendingOpRecord); err != nil {
			return out, err
		}
	}

	return out, nil
}

func stampTriggerOperation(effects []domain.Effect, world domain.World) []domain.Effect {
	trigger, ok := world.Action.(domain.TriggerOp)
	if !ok || world.AM == nil {
		return effects
	}
	requestID := world.AM.Annotations[domain.AnnotationOperationRequestID]
	if requestID == "" {
		return effects
	}

	out := append([]domain.Effect(nil), effects...)
	for i, effect := range out {
		patch, ok := effect.(domain.PatchAM)
		if !ok {
			continue
		}
		if patch.Status.OpType == "" {
			patch.Status.OpType = trigger.Op
		}
		if patch.Status.OpID == "" {
			patch.Status.OpID = requestID
		}
		out[i] = patch
		break
	}
	return out
}

// ensureStatePatch inserts a synthesised PatchAM at the head of
// res.Effects when the handler returned a Next state (or a
// Message / Reason) without an explicit PatchAM. This keeps the
// convenience path documented on domain.Result.Message working.
//
// If res.Next is empty AND res.Message/Reason are empty AND an
// explicit PatchAM is already present, ensureStatePatch is a no-op.
func ensureStatePatch(world domain.World, res domain.Result) []domain.Effect {
	if hasExplicitPatchAM(res.Effects) {
		return res.Effects
	}
	// If nothing about the AM would change, don't synthesise.
	if res.Next == "" && res.Message == "" && res.Reason == "" {
		return res.Effects
	}
	target := res.Next
	if target == "" {
		target = world.Current()
	}
	synth := domain.PatchAM{
		Status: appv1.ApplicationManagerStatus{
			State:   target,
			Message: res.Message,
			Reason:  res.Reason,
		},
	}
	out := make([]domain.Effect, 0, len(res.Effects)+1)
	out = append(out, synth)
	out = append(out, res.Effects...)
	return out
}

// advanceStatusView folds an applied PatchAM into the running status
// view dispatchResult hands to applyEmitEvent. Only the fields that
// appear on an emitted event are tracked, and only when the patch
// actually sets them — StatusPatch treats empty as "leave alone", so
// the view has to do the same.
func advanceStatusView(status *appv1.ApplicationManagerStatus, patch domain.PatchAM) {
	if status == nil {
		return
	}
	if patch.Status.State != "" {
		status.State = patch.Status.State
	}
	if patch.Status.Progress != "" {
		status.Progress = patch.Status.Progress
	}
	if patch.Status.OpType != "" {
		status.OpType = patch.Status.OpType
	}
	if patch.Status.OpID != "" {
		status.OpID = patch.Status.OpID
	}
}

func hasExplicitPatchAM(effects []domain.Effect) bool {
	for _, e := range effects {
		if _, ok := e.(domain.PatchAM); ok {
			return true
		}
	}
	return false
}

// applyPatchAM merges the PatchAM (and any buffered OpRecord) into
// a ports.StatusPatch and calls AMStore.PatchStatus under the
// world's captured ResourceVersion.
//
// Conflict handling:
//
//   - IsNotFound   → benign; the AM was deleted between Reconcile
//     start and this write. Skip.
//   - IsConflict   → not surfaced as an error; instead the Runner
//     sets ReconcileOutcome.Requeue=true so controller-runtime
//     re-enqueues immediately (next Reconcile re-reads RV).
//   - other errors → surfaced verbatim; controller-runtime backs off.
func (r *Runner) applyPatchAM(ctx context.Context, world domain.World, rv string, patch domain.PatchAM, opRecord *appv1.OpRecord) error {
	if world.AM == nil {
		return fmt.Errorf("runner: applyPatchAM invoked without World.AM")
	}
	sp := ports.StatusPatch{
		State:             patch.Status.State,
		Message:           patch.Status.Message,
		Reason:            patch.Status.Reason,
		OpType:            patch.Status.OpType,
		OpID:              patch.Status.OpID,
		Progress:          patch.Status.Progress,
		Payload:           patch.Status.Payload,
		DeletePayloadKeys: patch.DeletePayloadKeys,
	}
	if opRecord != nil {
		sp.OpRecord = opRecord
	}
	r.stampPendingQueuedAt(world, patch, &sp)
	// StatusTime is when the AM entered its current state. Refresh it
	// only on a real transition, or when a legacy object has no stamp
	// yet. Same-state progress/phase writes must not extend the TTL.
	now := metav1.NewTime(r.clock.Now())
	target := patch.Status.State
	if target == "" {
		target = world.Current()
	}
	if target != world.Current() || world.AM.Status.StatusTime == nil || world.AM.Status.StatusTime.IsZero() {
		sp.StatusTime = &now
	}
	sp.UpdateTime = &now

	name := world.AM.Name
	acquired := false
	if holdsDownloadSlot(target) {
		if !r.downloadSlot.tryAcquire(name) {
			return errDownloadSlotBusy
		}
		acquired = !holdsDownloadSlot(world.Current())
	}

	err := r.amStore.PatchStatus(ctx, name, rv, sp)
	if err != nil && acquired {
		r.downloadSlot.release(name)
	}
	if err == nil && holdsDownloadSlot(world.Current()) && target != "" && !holdsDownloadSlot(target) {
		r.downloadSlot.release(name)
	}
	switch {
	case err == nil:
		return nil
	case apierrors.IsNotFound(err):
		r.downloadSlot.release(name)
		return nil
	case apierrors.IsConflict(err):
		return errConflict
	default:
		return fmt.Errorf("runner: PatchStatus %s: %w", name, err)
	}
}

// applyPatchSpec writes the PatchSpec through
// AMStore.PatchSpecAndAnnotations and returns the ResourceVersion
// subsequent writes in the same effect list must use.
//
// Because PatchSpecAndAnnotations does not hand back the patched
// object, the new version comes from a re-read. That read is the
// price of letting a handler bundle a spec write with a status write;
// it is paid only on the rare paths that emit PatchSpec (today: the
// post-helm rejection that converts an install into a stop).
//
// Conflict handling matches applyPatchAM: IsNotFound is benign (the
// AM was deleted mid-Reconcile), IsConflict folds into errConflict so
// Reconcile requeues and restores the pending action, everything else
// surfaces verbatim.
func (r *Runner) applyPatchSpec(ctx context.Context, world domain.World, rv string, patch domain.PatchSpec) (string, error) {
	if world.AM == nil {
		return rv, fmt.Errorf("runner: applyPatchSpec invoked without World.AM")
	}
	name := world.AM.Name

	spec := &ports.SpecPatch{
		Config: patch.Config,
		OpType: patch.OpType,
	}
	var ann *ports.AnnotationPatch
	if len(patch.SetAnnotations) > 0 || len(patch.DeleteAnnotations) > 0 {
		ann = &ports.AnnotationPatch{
			Set:    patch.SetAnnotations,
			Delete: patch.DeleteAnnotations,
		}
	}

	err := r.amStore.PatchSpecAndAnnotations(ctx, name, rv, spec, ann)
	switch {
	case err == nil:
	case apierrors.IsNotFound(err):
		r.downloadSlot.release(name)
		return rv, nil
	case apierrors.IsConflict(err):
		return rv, errConflict
	default:
		return rv, fmt.Errorf("runner: PatchSpecAndAnnotations %s: %w", name, err)
	}

	// Re-read for the bumped ResourceVersion. A NotFound here is the
	// same benign race as above; any other read error would strand
	// the following PatchAM on a stale version, so it is surfaced.
	am, err := r.amStore.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return rv, nil
		}
		return rv, fmt.Errorf("runner: AMStore.Get %s after spec patch: %w", name, err)
	}
	return am.ResourceVersion, nil
}

// stampPendingQueuedAt writes pending.queuedAt once when the AM is
// landing in (or remaining in) Pending and the key is not already
// set. waitForSlot patches omit Payload so a subsequent wait does
// not refresh the FIFO key. An explicit empty Payload map means
// "clear" and is left alone.
func (r *Runner) stampPendingQueuedAt(world domain.World, patch domain.PatchAM, sp *ports.StatusPatch) {
	next := patch.Status.State
	if next == "" {
		next = world.Current()
	}
	if next != appv1.Pending && next != appv1.UpgradePending {
		return
	}
	if patch.Status.Payload != nil && len(patch.Status.Payload) == 0 {
		return
	}
	if world.AM != nil && world.AM.Status.Payload[domain.PayloadPendingQueuedAt] != "" {
		return
	}
	if sp.Payload != nil && sp.Payload[domain.PayloadPendingQueuedAt] != "" {
		return
	}
	if sp.Payload == nil {
		sp.Payload = make(map[string]string, 1)
	}
	sp.Payload[domain.PayloadPendingQueuedAt] = r.clock.Now().UTC().Format(time.RFC3339Nano)
}

// applyEmitEvent translates domain.EmitEvent into a
// ports.AppLifecycleEvent and hands it to the recorder. The Runner
// backfills identity fields (Name, RawAppName, User, OpType, OpID,
// State) so handlers do not need to remember to.
//
// Identity comes from World.AM's Spec, which no effect can change.
// RawAppName, Title, Icon, MarketSource and ChartOwner match the
// pre-refactor state-change event (ApplicationManagerController.
// publishStateChangeEvent). The status fields come from `status`,
// the view dispatchResult advances past every applied PatchAM, so
// an event announcing a transition reports the state it moved to.
// A nil status falls back to the World snapshot for callers that
// emit without patching.
//
// Same-state phase updates (helm steps, queue waits) still return
// EmitEvent, but only a real status.state change is published.
// Image-pull percent changes are the exception: they stay in
// Downloading / UpgradeDownloading and are published by
// publishImagePullProgress on every observed percent.
func (r *Runner) applyEmitEvent(ctx context.Context, world domain.World, status *appv1.ApplicationManagerStatus, e domain.EmitEvent) error {
	if world.AM == nil {
		return fmt.Errorf("runner: applyEmitEvent invoked without World.AM")
	}
	am := world.AM
	if status == nil {
		status = &am.Status
	}
	from := am.Status.State
	to := status.State
	if to == "" || to == from {
		return nil
	}
	klog.Infof("state change %s %s -> %s reason=%s", am.Name, from, to, e.Reason)
	ev := ports.AppLifecycleEvent{
		EventID:      fmt.Sprintf("%s-%s-%d", am.Spec.AppOwner, am.Name, r.clock.Now().UnixNano()),
		Name:         am.Spec.AppName,
		RawAppName:   legacyEventRawAppName(am),
		Type:         am.Spec.Type.String(),
		OpType:       string(status.OpType),
		OpID:         status.OpID,
		State:        string(status.State),
		Progress:     status.Progress,
		User:         am.Spec.AppOwner,
		Title:        apputils.AppTitle(am.Spec.Config),
		Icon:         apputils.AppIcon(am.Spec.Config),
		Reason:       e.Reason,
		Message:      e.Message,
		MarketSource: amAnnotation(am, constants.AppMarketSourceKey),
		ChartOwner:   appcfg.GetChartOwner(am),
		CreateTime:   r.clock.Now(),
	}
	if err := r.recorder.RecordAppEvent(ctx, ev); err != nil {
		return fmt.Errorf("runner: RecordAppEvent %s: %w", am.Name, err)
	}
	return nil
}

// launchCallWorkload spawns a Runner-owned goroutine that routes the
// CallWorkload to the correct port and, on completion, delivers
// OnSuccess or OnFailure into PendingActionQueue and pokes the
// Requeuer. Non-blocking.
//
// Skeleton scope: the routing table covers only the WorkloadKind
// values the Install / Uninstall / Upgrade handlers will need first.
// Other kinds return an "unrouted" error to the pending queue as
// WorkloadFailed so the state machine can still make forward progress
// and log the miss. Fill-in per handler as they land.
func (r *Runner) launchCallWorkload(_ context.Context, world domain.World, cw domain.CallWorkload) error {
	amName := ""
	if world.AM != nil {
		amName = world.AM.Name
	}
	if amName == "" {
		return fmt.Errorf("runner: launchCallWorkload invoked without World.AM.Name")
	}
	if cw.Preempts {
		if r.workloads.preemptClaim(amName) {
			klog.Infof("runner: %s preempts the workload in flight for am=%s", cw.Kind, amName)
		}
	}
	// Snapshot the AM pointer for the goroutine. World is immutable
	// and the AM pointer is captured; the goroutine must not mutate
	// it (handler-contract-style discipline extends to Runner
	// callbacks).
	am := world.AM
	req := ports.WorkloadRequest{
		AM:        am,
		Namespace: am.Spec.AppNamespace,
		Token:     amToken(am),
		Opt: ports.WorkloadOpt{
			Source:             am.Spec.Source,
			MarketSource:       appcfg.GetMarketSource(am),
			SkipWaitForStartUp: amPreUpgradeStopped(am),
		},
	}
	// Each workload gets its own cancellable context so a preempting
	// CallWorkload can stop the one it interrupts. See workloadClaim.
	// A timeout-enabled state also bounds that context by the same
	// persisted deadline Reconcile uses.
	ctx, cancel := context.WithCancel(r.ctx)
	if remaining, ok := workloadTimeout(r.engine.Table(), am, r.clock.Now()); ok {
		req.Timeout = remaining
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, remaining)
		parentCancel := cancel
		ctx = timeoutCtx
		cancel = func() {
			timeoutCancel()
			parentCancel()
		}
	}
	token, claimed := r.workloads.claim(amName, cw.Kind, cancel)
	if !claimed {
		cancel()
		return nil
	}

	r.wg.Add(1)
	r.inFlight.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.inFlight.Add(-1)

		// ctx is bounded by Runner shutdown and by this claim being
		// preempted — not by the invoking Reconcile returning, so a
		// helm install is never cut in half by a finished reconcile.
		// The Runner's Stop() waits for wg to drain.
		defer cancel()

		// AppCfg is materialised here rather than in buildWorld so
		// the JSON parse stays off the hot Reconcile path for AMs
		// that emit no CallWorkload. A parse failure folds into the
		// normal WorkloadFailed route instead of aborting the
		// dispatch, which would strand the AM in the phase the
		// preceding PatchAM just stamped.
		var result ports.WorkloadResult
		if cfg, err := appCfgFromAM(am); err != nil {
			result = ports.WorkloadResult{LandedReason: "AppConfigInvalid", Err: err}
		} else {
			req.AppCfg = cfg
			result = r.dispatchWorkloadKind(ctx, cw.Kind, req)
		}

		// Priority for failure-action population:
		//
		//   1. Workload-provided data (result.LandedReason /
		//      result.Err) wins whenever present. The workload
		//      knows the actual failure — the validator's decision
		//      code, the helm error, etc. — and its
		//      diagnostic is what operators need in the AM Status.
		//
		//   2. Handler-supplied OnFailure fields are the FALLBACK
		//      for when the workload returned no data (rare but
		//      possible for canceled or short-circuited ops).
		//      They also give the handler a way to tag the
		//      failure with sub-phase context that the workload
		//      itself doesn't know about — but only when the
		//      workload didn't provide its own.
		//
		// This inverts the naive "handler's value wins" rule,
		// which would suppress a rich validator decision behind a
		// generic phase-name reason.
		var action domain.Action
		if result.Err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// The persisted state deadline elapsed and the workload
			// failed. Deliver the same action Reconcile would, so both
			// observers land on one transition. A successful result
			// still queues OnSuccess: Reconcile consumes completed
			// callbacks before applying a state deadline.
			action = domain.TimeoutFired{}
		} else if result.Err != nil {
			wf := domain.WorkloadFailed{}
			if handlerWF, ok := cw.OnFailure.(domain.WorkloadFailed); ok {
				wf = handlerWF
			}
			if result.LandedReason != "" {
				wf.Reason = result.LandedReason
			}
			if result.Err != nil {
				wf.Message = result.Err.Error()
			}
			// Routing hints are workload-only: a handler cannot
			// know that helm failed on an unschedulable pod rather
			// than a bad chart, so there is nothing to fall back to
			// and no handler value to preserve.
			wf.LandedState = domain.State(result.LandedState)
			wf.StopAll = result.StopAll
			// For non-WorkloadFailed OnFailure (e.g. a bare Tick),
			// the caller intentionally routed the failure through
			// a different Action shape — respect that verbatim
			// rather than fabricating a WorkloadFailed.
			if _, isWF := cw.OnFailure.(domain.WorkloadFailed); cw.OnFailure != nil && !isWF {
				action = cw.OnFailure
			} else {
				action = wf
			}
		} else {
			if cw.OnSuccess != nil {
				action = cw.OnSuccess
			} else {
				action = domain.WorkloadReady{}
			}
		}

		if !r.workloads.completeClaim(amName, token) {
			// A preempting CallWorkload took the slot while this
			// one ran — typically a cancel stopping the very pull
			// that is reporting here. Delivering now would feed the
			// cancel pipeline a callback belonging to the work it
			// just interrupted.
			klog.Infof("runner: dropping %s result for am=%s; the claim was preempted", cw.Kind, amName)
			return
		}
		r.pending.Deliver(amName, action)
		r.requeuer.Requeue(amName)
	}()
	return nil
}

// workloadTimeout is the time left before the AM's state deadline.
// The bool is false for states that do not time out. A non-positive
// duration means the deadline has already elapsed; context.WithTimeout
// treats that as immediately done.
func workloadTimeout(table *domain.TransitionTable, am *appv1.ApplicationManager, now time.Time) (time.Duration, bool) {
	if table == nil || am == nil {
		return 0, false
	}
	ttl, ok := table.Timeout(am.Status.State)
	if !ok {
		return 0, false
	}
	return stateEntryTime(am, now).Add(ttl).Sub(now), true
}

// appCfgFromAM materialises the *appcfg.ApplicationConfig that every
// workload port needs from the AM's Spec.Config blob.
//
// Returns an untyped nil — not a typed nil pointer — when Spec.Config
// is absent, so adapters see WorkloadRequest.AppCfg == nil and take
// their own fallback path (helmops re-unmarshals Spec.Config; imgpull
// reports a nil-config error naming itself). A typed nil would defeat
// those nil checks and surface as a confusing type error instead.
func appCfgFromAM(am *appv1.ApplicationManager) (any, error) {
	if am == nil || am.Spec.Config == "" {
		return nil, nil
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(am.Spec.Config), &cfg); err != nil {
		return nil, fmt.Errorf("runner: unmarshal AM.Spec.Config for %s: %w", am.Name, err)
	}
	return &cfg, nil
}

// dispatchWorkloadKind is the port-routing switch for CallWorkload.
// Kept small on purpose: additional kinds arrive as handlers land.
// An unrouted Kind returns a WorkloadResult with a descriptive Err
// so the corresponding WorkloadFailed action carries a real reason.
func (r *Runner) dispatchWorkloadKind(ctx context.Context, kind domain.WorkloadKind, req ports.WorkloadRequest) ports.WorkloadResult {
	switch kind {
	case domain.WorkloadInstall:
		res := r.workloadOps.Install(ctx, req)
		if res.Err != nil {
			r.deleteAllocationsForApp(ctx, req)
		}
		return res
	case domain.WorkloadUpgrade:
		return r.workloadOps.Upgrade(ctx, req)
	case domain.WorkloadUninstall:
		return r.dispatchUninstall(ctx, req)
	case domain.WorkloadStop:
		return r.dispatchStop(ctx, req)
	case domain.WorkloadResume:
		return r.dispatchResume(ctx, req)
	case domain.WorkloadApplyEnv:
		return r.workloadOps.ApplyEnv(ctx, req)
	case domain.WorkloadWaitStartUp:
		return r.workloadOps.WaitForStartUp(ctx, req)
	case domain.WorkloadWaitLaunch:
		return r.workloadOps.WaitForLaunch(ctx, req)
	case domain.WorkloadWaitNamespaceGone:
		nsReq := ports.NamespaceWaitRequest{
			Primary:  req.Namespace,
			SharedNS: sharedNamespaces(req),
			Timeout:  req.Timeout,
		}
		return r.workloadOps.WaitForNamespaceDeleted(ctx, nsReq)
	case domain.WorkloadValidateInstall:
		return r.dispatchValidateInstall(ctx, req)
	case domain.WorkloadSetExposePorts:
		return r.dispatchSetExposePorts(ctx, req)
	case domain.WorkloadCleanupInstallFailure:
		return r.dispatchCleanupInstallFailure(ctx, req)
	case domain.WorkloadPrepareUpgrade:
		return r.dispatchPrepareUpgrade(ctx, req)
	case domain.WorkloadDeleteAllocations:
		return r.dispatchDeleteAllocations(ctx, req)
	case domain.WorkloadPullImage:
		return r.dispatchImagePull(ctx, req)
	case domain.WorkloadCancelImagePull:
		return r.dispatchCancelImagePull(ctx, req)
	default:
		return ports.WorkloadResult{
			LandedReason: "UnroutedWorkloadKind",
			Err:          fmt.Errorf("runner: no port routing for CallWorkload{Kind=%q}", kind),
		}
	}
}

// dispatchValidateInstall bridges the WorkloadRequest → ValidationInput
// impedance mismatch and converts a Rejection Decision into a
// WorkloadResult with Err set + LandedReason populated. This is the
// core reason we DO NOT need to introduce a new "ValidationRejected"
// Action variant: the state machine treats rejection identically to
// any other workload failure, and the WorkloadFailed handler branches
// on Reason.
//
// Design contract for validator adapters: a genuine transport /
// programming error uses err != nil directly; a validation rejection
// returns Decision{OK: false, Reason, Message, ...} with err == nil.
// The Runner unifies both: rejection folds into a synthetic
// validationRejectedError with the Decision's Reason/Message, so
// downstream handlers see a single failure shape.
func (r *Runner) dispatchValidateInstall(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	in := ports.ValidationInput{
		AppCfg: req.AppCfg,
		Op:     ports.ValidateOpInstall,
		Token:  req.Token,
	}
	decision, err := r.validator.ValidateInstall(ctx, in)
	if err != nil {
		return ports.WorkloadResult{
			LandedReason: decision.Reason,
			Err:          fmt.Errorf("runner: ValidateInstall: %w", err),
		}
	}
	if !decision.OK {
		if decision.Validator == validation.NameComputeAllocation {
			if nerr := r.validator.NotifyInsufficientCompute(ctx, req.AppCfg, errors.New(decision.Message)); nerr != nil {
				klog.Warningf("runner: NotifyInsufficientCompute: %v", nerr)
			}
		}
		return ports.WorkloadResult{
			LandedReason: decision.Reason,
			Err:          &validationRejectedError{decision: decision},
			StopAll:      r.managesSharedServer(ctx, req.AppCfg),
		}
	}
	return ports.WorkloadResult{}
}

// managesSharedServer answers "if this app has to be stopped, does
// the shared server it owns have to come down with it?" — the bit
// that decides whether the AM gets the stop-all flag.
//
// Consulted only when validation rejects, so the compute-target
// lookup stays off the reconcile hot path. A lookup error is logged
// and reported as false, matching legacy installing_app.go: a
// misread here should scope the stop too narrowly (recoverable by a
// retry) rather than tear down a shared server other apps depend on.
func (r *Runner) managesSharedServer(ctx context.Context, appCfg any) bool {
	if r.computeOps == nil || appCfg == nil {
		return false
	}
	manages, err := r.computeOps.ManagesSharedServer(ctx, appCfg)
	if err != nil {
		klog.Warningf("runner: ManagesSharedServer: %v; scoping stop to the app's own workloads", err)
		return false
	}
	return manages
}

// imagePullPollInterval is the cadence at which awaitImagePull
// re-reads ImageManager progress. Matches the legacy
// images.ImageManagerClient.PollDownloadProgress ticker. Each tick
// is an informer-cache read, so the interval is cheap.
const imagePullPollInterval = time.Second

// dispatchImagePull bridges CallWorkload{WorkloadPullImage} to the
// ImagePuller port. Runs inside a launchCallWorkload goroutine — so
// blocking-until-terminal semantics are safe here (unlike inside a
// handler).
//
// The ImageManager lifecycle is a two-phase async op: CreateJob
// starts the pull, then Progress must be polled until the bytes
// have landed (Percent>=100) or the pull has Failed. Returning
// as soon as CreateJob succeeds would report the download as
// finished before a single byte landed, which also releases the
// Downloading slot immediately and defeats the one-pull-at-a-time
// FIFO gate in OnPendingTick. Unpack into containerd is not part
// of this wait — Installing owns that window.
//
// Design contract for image-puller adapters:
//
//   - A real transport / infra error uses err != nil directly.
//   - A permanent pull failure (e.g. image doesn't exist, auth
//     denied) surfaces as Progress.Failed with a human-readable
//     Message; LandedReason is set to "ImagePullFailed".
func (r *Runner) dispatchImagePull(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	// ResolveRefs is a HANDLER-SAFE READ per the port comment, but
	// running it here (rather than in buildWorld) keeps the heavy
	// Application-config JSON parse out of the hot Reconcile path
	// for AMs that aren't currently Downloading. Since this
	// dispatcher only fires when a handler emitted
	// CallWorkload{WorkloadPullImage}, the work is naturally
	// scoped to AMs that need it.
	refs, err := r.imagePuller.ResolveRefs(ctx, req.AM, req.AppCfg)
	if err != nil {
		return ports.WorkloadResult{
			LandedReason: "ImageResolveFailed",
			Err:          fmt.Errorf("runner: ImagePuller.ResolveRefs: %w", err),
		}
	}
	if err := r.imagePuller.CreateJob(ctx, req.AM, refs); err != nil {
		return ports.WorkloadResult{
			LandedReason: "ImagePullStartFailed",
			Err:          fmt.Errorf("runner: ImagePuller.CreateJob: %w", err),
		}
	}
	return r.awaitImagePull(ctx, req.AM)
}

// awaitImagePull blocks until every node's bytes have landed
// (Percent>=100) or the pull has failed. ImageManager's
// `completed` state also includes layer unpack into containerd;
// that window used to pin the AM at Downloading/100%. We leave it
// to Installing (helm install + WaitForStartUp) so the UI can
// leave the download bar as soon as the last byte is on disk.
//
// A zero-value Progress (no IM CR visible yet) is not terminal, so
// the loop keeps waiting through the window between CreateJob and
// the informer catching up.
//
// Each observed percentage change is logged and published on the
// app-lifecycle NATS subject, matching legacy
// ImageManagerClient.updateProgress. AM.Status.Progress is still
// not written here — the UI follows the events, not Status.
func (r *Runner) awaitImagePull(ctx context.Context, am *appv1.ApplicationManager) ports.WorkloadResult {
	amName, appName := "", ""
	if am != nil {
		amName = am.Name
		appName = am.Spec.AppName
	}
	ticker := r.clock.NewTicker(imagePullPollInterval)
	defer ticker.Stop()

	// lastProgress is the last NATS/log string ("12.34"). Legacy
	// compared the raw float; we compare the two-decimal rendering
	// so identical UI ticks are not republished.
	lastProgress := ""

	for {
		select {
		case <-ctx.Done():
			// Runner shutdown. The IM CR keeps pulling; the next
			// process to reconcile this AM re-enters Downloading and
			// images.ImageManagerClient.Create adopts the in-flight
			// job rather than restarting it.
			return ports.WorkloadResult{
				LandedReason: "ImagePullAbandoned",
				Err:          fmt.Errorf("runner: image pull %s abandoned: %w", amName, ctx.Err()),
			}

		case <-ticker.C():
			prog, err := r.imagePuller.Progress(ctx, amName)
			if err != nil {
				return ports.WorkloadResult{
					LandedReason: "ImagePullPollFailed",
					Err:          fmt.Errorf("runner: ImagePuller.Progress %s: %w", amName, err),
				}
			}
			progressStr := formatPullProgress(prog)
			if progressStr != lastProgress {
				lastProgress = progressStr
				klog.Infof("runner: image pull progress app=%s am=%s percent=%s complete=%t failed=%t",
					appName, amName, progressStr, prog.Complete, prog.Failed)
				r.publishImagePullProgress(ctx, am, progressStr)
			}
			switch {
			case prog.Failed:
				// Covers both a genuine pull failure and a cancel
				// landing on the IM CR, exactly as legacy did.
				msg := prog.Message
				if msg == "" {
					msg = prog.Reason
				}
				if msg == "" {
					msg = "image pull failed"
				}
				return ports.WorkloadResult{
					LandedReason: "ImagePullFailed",
					Err:          errors.New(msg),
				}
			case prog.Complete || prog.Percent >= 100:
				// Complete is the IM CR's `completed` state (bytes
				// + unpack). Percent>=100 is the byte-only
				// signal: offset==total on every node while the
				// controller is still unpacking. Either is enough
				// to release Downloading; Failed is checked first
				// so 100%+Failed stays a failure.
				return ports.WorkloadResult{}
			}
		}
	}
}

// publishImagePullProgress pushes the current download percent onto
// the per-owner NATS subject, matching legacy
// ImageManagerClient.updateProgress. A recorder error is logged and
// swallowed: failing the poll would strand the AM in Downloading
// because the UI missed a tick.
//
// Field choices are the legacy ones. RawAppName is the catalog name
// (Spec.RawAppName, else AppName), not the AM object name. State is
// forced to Downloading, or Upgrading when Spec.OpType is an upgrade,
// so the market progress bar — which only renders for those two
// states — still appears if the goroutine's AM snapshot is stale.
func (r *Runner) publishImagePullProgress(ctx context.Context, am *appv1.ApplicationManager, progress string) {
	if r.recorder == nil || am == nil {
		return
	}
	ev := ports.AppLifecycleEvent{
		EventID:      fmt.Sprintf("%s-%s-%d", am.Spec.AppOwner, am.Name, r.clock.Now().UnixNano()),
		Name:         am.Spec.AppName,
		RawAppName:   legacyEventRawAppName(am),
		Type:         am.Spec.Type.String(),
		OpType:       string(am.Status.OpType),
		OpID:         am.Status.OpID,
		State:        imagePullEventState(am),
		Progress:     progress,
		User:         am.Spec.AppOwner,
		Title:        apputils.AppTitle(am.Spec.Config),
		Icon:         apputils.AppIcon(am.Spec.Config),
		MarketSource: amAnnotation(am, constants.AppMarketSourceKey),
		ChartOwner:   appcfg.GetChartOwner(am),
		CreateTime:   r.clock.Now(),
	}
	if err := r.recorder.RecordAppEvent(ctx, ev); err != nil {
		klog.Errorf("runner: publish image pull progress app=%s percent=%s: %v", am.Spec.AppName, ev.Progress, err)
	}
}

// formatPullProgress renders the poll snapshot the way legacy
// ImageManagerClient.updateProgress did: strconv.FormatFloat(_, 'f', 2, 64).
func formatPullProgress(prog ports.ImageProgress) string {
	return strconv.FormatFloat(prog.Percent, 'f', 2, 64)
}

// dispatchCancelImagePull bridges CallWorkload{WorkloadCancelImagePull}
// to the ImagePuller port. Fires from the DownloadingCanceling
// handler's first phase — the cancel is the mirror of dispatchImagePull.
//
// Contract on the adapter side (see ports/image.go):
//
//   - CancelJob is idempotent: calling it on an AM whose IM CR
//     already terminated (Complete/Failed/Canceled) MUST return
//     nil so a stale WorkloadReady arriving from the original pull
//     (see PendingActionQueue single-slot semantics) does not
//     re-trigger this dispatch after the fact.
//
//   - The port receives amName (not the AM object) because a
//     cancel-during-cancel can arrive AFTER the AM's namespace has
//     already been deleted — the adapter only needs the name to
//     locate the IM CR (which lives in an admin namespace, not
//     the AM's).
//
// Failure classification mirrors dispatchImagePull: any err from
// CancelJob folds into a WorkloadResult with a Reason the
// downstream DownloadingCanceling.OnWorkloadFailed handler
// unpacks into an ImageCancelFailed → DownloadingCancelFailed
// transition.
func (r *Runner) dispatchCancelImagePull(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	name := ""
	if req.AM != nil {
		name = req.AM.Name
	}
	if err := r.imagePuller.CancelJob(ctx, name); err != nil {
		// Legacy upgrading_canceling_app.go treats a missing IM as
		// already canceled so the AM still proceeds to Stopping.
		if apierrors.IsNotFound(err) {
			klog.Infof("im name=%s not found while canceling image pull, treating as already canceled", name)
			return ports.WorkloadResult{}
		}
		return ports.WorkloadResult{
			LandedReason: "ImageCancelFailed",
			Err:          fmt.Errorf("runner: ImagePuller.CancelJob: %w", err),
		}
	}
	return ports.WorkloadResult{}
}

// validationRejectedError is the sentinel wrapper the Runner returns
// from dispatchValidateInstall when a validator rejects. Handlers
// that need the structured Decision fields can errors.As against it
// (though the common case just consumes WorkloadFailed.Reason /
// Message which the Runner already backfilled from LandedReason /
// err.Error()).
type validationRejectedError struct {
	decision ports.ValidationDecision
}

// Error implements error. Message is the human-readable rejection
// description, prefixed with the validator name for diagnostic
// clarity in logs.
func (e *validationRejectedError) Error() string {
	if e.decision.Validator != "" {
		return fmt.Sprintf("%s: %s", e.decision.Validator, e.decision.Message)
	}
	return e.decision.Message
}

// Decision exposes the underlying ValidationDecision for handlers
// that want more than Reason/Message (e.g. the specific validator
// that rejected). Retrievable via errors.As on WorkloadFailed's
// wrapped error, once the Runner threads the error through — TODO
// when a handler actually needs it.
func (e *validationRejectedError) Decision() ports.ValidationDecision { return e.decision }

// upgradePreparer is the optional WorkloadOps extension that runs
// GetDeployedReleaseVersion + GetAppConfig / GetIndexAndDownloadChart.
// Production helmops implements it; the fake used by handler tests
// does not, and a missing implementation is treated as success so
// those tests still walk the prepare → pull → helm phases.
type upgradePreparer interface {
	PrepareUpgrade(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
}

func (r *Runner) dispatchPrepareUpgrade(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	p, ok := r.workloadOps.(upgradePreparer)
	if !ok {
		return ports.WorkloadResult{}
	}
	return p.PrepareUpgrade(ctx, req)
}

func (r *Runner) dispatchSetExposePorts(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	cfg, _ := req.AppCfg.(*appcfg.ApplicationConfig)
	if cfg == nil {
		return ports.WorkloadResult{}
	}
	if len(cfg.Ports) == 0 {
		return ports.WorkloadResult{}
	}
	if err := apputils.SetExposePorts(ctx, cfg, nil); err != nil {
		return ports.WorkloadResult{LandedReason: "SetExposePortsFailed", Err: err}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return ports.WorkloadResult{LandedReason: "SetExposePortsFailed", Err: err}
	}
	if req.AM == nil {
		return ports.WorkloadResult{}
	}
	// The workload request still carries the ResourceVersion from
	// the Reconcile that dispatched it. Status writes (the
	// Installing phase stamp, download progress) bump that version
	// before this spec write runs. Patch against a fresh read, and
	// retry when another status write wins the race.
	if err := r.patchExposeConfig(ctx, req.AM.Name, raw); err != nil {
		return ports.WorkloadResult{LandedReason: "SetExposePortsFailed", Err: err}
	}
	return ports.WorkloadResult{}
}

func (r *Runner) patchExposeConfig(ctx context.Context, name string, raw []byte) error {
	spec := &ports.SpecPatch{Config: string(raw)}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := r.amStore.Get(ctx, name)
		if err != nil {
			return err
		}
		return r.amStore.PatchSpecAndAnnotations(ctx, name, latest.ResourceVersion, spec, nil)
	})
}

func (r *Runner) dispatchCleanupInstallFailure(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	r.deleteAllocationsForApp(ctx, req)
	if un := r.workloadOps.Uninstall(ctx, req); un.Err != nil {
		return un
	}
	return r.workloadOps.WaitForNamespaceDeleted(ctx, ports.NamespaceWaitRequest{
		Primary:  req.Namespace,
		SharedNS: req.SharedNS,
		Timeout:  req.Timeout,
	})
}

func (r *Runner) dispatchDeleteAllocations(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	r.deleteAllocationsForApp(ctx, req)
	return ports.WorkloadResult{}
}

func (r *Runner) deleteAllocationsForApp(ctx context.Context, req ports.WorkloadRequest) {
	if r.computeOps == nil {
		return
	}
	name, owner := "", ""
	if req.AM != nil {
		name, owner = req.AM.Spec.AppName, req.AM.Spec.AppOwner
	}
	if cfg, ok := req.AppCfg.(*appcfg.ApplicationConfig); ok && cfg != nil {
		if cfg.AppName != "" {
			name = cfg.AppName
		}
		if cfg.OwnerName != "" {
			owner = cfg.OwnerName
		}
	}
	if name == "" {
		return
	}
	if err := r.computeOps.DeleteForApp(ctx, name, owner); err != nil {
		klog.Warningf("runner: DeleteAllocationsForApp %s/%s: %v", owner, name, err)
	}
}

func amToken(am *appv1.ApplicationManager) string {
	if am == nil || am.Annotations == nil {
		return ""
	}
	return am.Annotations[api.AppTokenKey]
}

func amPreUpgradeStopped(am *appv1.ApplicationManager) bool {
	if am == nil || am.Annotations == nil {
		return false
	}
	return am.Annotations[api.AppPreUpgradeStateKey] == string(appv1.Stopped)
}

// podStopWaiter is the optional WorkloadOps extension that polls
// until the app namespace has no live pods. Production helmops
// implements it; the fake used by handler tests does not, and a
// missing implementation is treated as success so those tests
// still walk Stopping → Stopped.
type podStopWaiter interface {
	WaitForPodsGone(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
}

func (r *Runner) dispatchStop(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	if res := r.workloadOps.Stop(ctx, req); res.Err != nil {
		return res
	}
	r.cascadeStopAllPeers(ctx, req)
	if err := r.stopMiddlewareIfNeeded(ctx, req); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if w, ok := r.workloadOps.(podStopWaiter); ok {
		if res := w.WaitForPodsGone(ctx, req); res.Err != nil {
			return res
		}
	}
	stopServer := amAnnotation(req.AM, api.AppStopAllKey) == "true"
	if err := r.deleteAllocationsForTarget(ctx, req, stopServer); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

func (r *Runner) dispatchResume(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	if res := r.workloadOps.Resume(ctx, req); res.Err != nil {
		return res
	}
	if err := r.startMiddlewareIfNeeded(ctx, req); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

func (r *Runner) dispatchUninstall(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	if res := r.uninstallLegacyMongoIfNeeded(ctx, req); res != nil {
		return *res
	}
	uninstallAll := amAnnotation(req.AM, api.AppUninstallAllKey) == "true"
	var res ports.WorkloadResult
	if uninstallAll {
		res = r.workloadOps.UninstallAll(ctx, req)
	} else {
		res = r.workloadOps.Uninstall(ctx, req)
	}
	if res.Err != nil {
		return res
	}
	if err := r.deleteAllocationsForTarget(ctx, req, uninstallAll); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// uninstallLegacyMongoIfNeeded returns a non-nil result when the
// os-platform/mongodb special case applies (success or failure).
// A nil pointer means "fall through to helm uninstall".
func (r *Runner) uninstallLegacyMongoIfNeeded(ctx context.Context, req ports.WorkloadRequest) *ports.WorkloadResult {
	cfg, _ := req.AppCfg.(*appcfg.ApplicationConfig)
	if cfg == nil || cfg.MiddlewareName != "mongodb" || cfg.Namespace != "os-platform" {
		return nil
	}
	if r.middlewareOps == nil || req.AM == nil {
		return nil
	}
	klog.Infof("delete old mongodb ..........")
	mc := &middlewareinstaller.MiddlewareConfig{
		MiddlewareName: req.AM.Spec.AppName,
		Namespace:      req.AM.Spec.AppNamespace,
		OwnerName:      req.AM.Spec.AppOwner,
	}
	if err := r.middlewareOps.UninstallLegacyMongo(ctx, mc); err != nil {
		return &ports.WorkloadResult{LandedReason: "LegacyMongoUninstallFailed", Err: err}
	}
	return &ports.WorkloadResult{}
}

func (r *Runner) stopMiddlewareIfNeeded(ctx context.Context, req ports.WorkloadRequest) error {
	if r.middlewareOps == nil || req.AM == nil {
		return nil
	}
	if req.AM.Spec.Type != appv1.Middleware || !userspace.IsKbMiddlewares(req.AM.Spec.AppName) {
		return nil
	}
	if err := r.middlewareOps.Stop(ctx, req.AM); err != nil {
		klog.Errorf("suspend middleware %s failed %v", req.AM.Spec.AppName, err)
		return err
	}
	return nil
}

func (r *Runner) startMiddlewareIfNeeded(ctx context.Context, req ports.WorkloadRequest) error {
	if r.middlewareOps == nil || req.AM == nil {
		return nil
	}
	if req.AM.Spec.Type != appv1.Middleware || !userspace.IsKbMiddlewares(req.AM.Spec.AppName) {
		return nil
	}
	if err := r.middlewareOps.Start(ctx, req.AM); err != nil {
		klog.Errorf("failed to resume middleware %s,err=%v", req.AM.Spec.AppName, err)
		return err
	}
	return nil
}

func (r *Runner) deleteAllocationsForTarget(ctx context.Context, req ports.WorkloadRequest, includeSharedServer bool) error {
	if r.computeOps == nil {
		return nil
	}
	if err := r.computeOps.DeleteForTarget(ctx, req.AppCfg, includeSharedServer); err != nil {
		name := ""
		if req.AM != nil {
			name = req.AM.Spec.AppName
		}
		klog.Errorf("delete compute allocation for app %s failed %v", name, err)
		return err
	}
	return nil
}

// cascadeStopAllPeers asks peer AMs (same AppName, different owner)
// to stop when a V2 stop-all is in flight. It writes Spec.OpType,
// the operation-request-id, and the parent's stop reason. The
// ApplicationManager update is watched by the lifecycle controller,
// so the peer's next Reconcile re-derives StopOp from that
// persisted request. Conditions match suspending_app.go:exec
// verbatim.
//
// The workload request carries the pre-reconcile AM. A pod-abnormal
// stop stamps AppStopReasonKey / AppStopMessageKey before that
// snapshot's Status is updated, so those annotations win. Status is
// the fallback once stopToStopping has copied them onto the
// Stopping object and cleared the keys.
func (r *Runner) cascadeStopAllPeers(ctx context.Context, req ports.WorkloadRequest) {
	if req.AM == nil || amAnnotation(req.AM, api.AppStopAllKey) != "true" {
		return
	}
	cfg, _ := req.AppCfg.(*appcfg.ApplicationConfig)
	if cfg == nil || cfg.APIVersion != appcfg.V2 {
		return
	}
	klog.Infof("stopping other users' clients for v2 app %s", req.AM.Spec.AppName)
	list, err := r.amStore.List(ctx)
	if err != nil {
		klog.Errorf("failed to list application managers: %v", err)
		return
	}
	for i := range list {
		am := list[i]
		if am.Spec.AppName != req.AM.Spec.AppName || am.Spec.AppOwner == req.AM.Spec.AppOwner {
			continue
		}
		if am.Spec.Type != appv1.App && am.Spec.Type != appv1.Middleware {
			continue
		}
		if am.Status.State == appv1.Stopped || am.Status.State == appv1.Stopping {
			klog.Infof("app %s owner %s already in stopped/stopping state, skip", am.Spec.AppName, am.Spec.AppOwner)
			continue
		}
		// Cascade runs once. Leave a request whose landing state rejects
		// Stop (Upgrade, Uninstall, Cancel). ApplyEnv lands in ApplyingEnv,
		// which still allows Stop, so that peer is stopped here; there is
		// no later pass to catch it.
		if op, pending := domain.UnconsumedOperation(&am); pending && !domain.StopRemainsAllowed(am.Status.State, op) {
			klog.Infof("app %s owner %s has unconsumed %s, skip cascade stop", am.Spec.AppName, am.Spec.AppOwner, op)
			continue
		}
		if !r.engine.Table().AllowsOp(am.Status.State, appv1.StopOp) {
			klog.Infof("app %s owner %s not allowed do stop operation, skip", am.Spec.AppName, am.Spec.AppOwner)
			continue
		}
		ann := map[string]string{
			domain.AnnotationOperationRequestID: uuid.NewString(),
		}
		reason, message := cascadeStopReason(req.AM)
		if reason != "" {
			ann[api.AppStopReasonKey] = reason
		}
		if message != "" {
			ann[api.AppStopMessageKey] = message
		}
		if err := r.amStore.PatchSpecAndAnnotations(ctx, am.Name, am.ResourceVersion, &ports.SpecPatch{
			OpType: appv1.StopOp,
		}, &ports.AnnotationPatch{Set: ann}); err != nil {
			klog.Errorf("cascade stop peer %s failed: %v", am.Name, err)
			continue
		}
		klog.Infof("stopping client for user %s, app %s", am.Spec.AppOwner, am.Spec.AppName)
	}
}

func sharedNamespaces(req ports.WorkloadRequest) []string {
	if len(req.SharedNS) > 0 {
		return req.SharedNS
	}
	if amAnnotation(req.AM, api.AppUninstallAllKey) != "true" {
		return nil
	}
	cfg, _ := req.AppCfg.(*appcfg.ApplicationConfig)
	if cfg == nil || !cfg.IsV2() || !cfg.HasClusterSharedCharts() {
		return nil
	}
	var out []string
	for _, chart := range cfg.SubCharts {
		if !chart.Shared {
			continue
		}
		out = append(out, appcfg.ChartNamespace(&chart, cfg.OwnerName))
	}
	return out
}

// legacyEventRawAppName is the market-card key used by
// ImageManagerClient.updateProgress and PublishAppEventToQueue.
// An empty Spec.RawAppName falls back to the catalog AppName, never
// to the AM object name ({namespace}-{app}).
func legacyEventRawAppName(am *appv1.ApplicationManager) string {
	if am.Spec.RawAppName != "" {
		return am.Spec.RawAppName
	}
	return am.Spec.AppName
}

// imagePullEventState matches ImageManagerClient.updateProgress:
// the poll advertises Downloading, or UpgradeDownloading when the
// in-flight operation is an upgrade.
func imagePullEventState(am *appv1.ApplicationManager) string {
	if am.Spec.OpType == appv1.UpgradeOp {
		return appv1.UpgradeDownloading.String()
	}
	return appv1.Downloading.String()
}

// holdsDownloadSlot reports whether state occupies the single
// cluster-wide image-pull slot shared by install and upgrade.
func holdsDownloadSlot(state appv1.ApplicationManagerState) bool {
	return state == appv1.Downloading || state == appv1.UpgradeDownloading
}

// cascadeStopReason is the reason and message a stop-all peer
// should publish. Annotations are the stop that has not yet been
// copied onto Status; Status covers the reconcile after that copy.
func cascadeStopReason(am *appv1.ApplicationManager) (reason, message string) {
	reason = amAnnotation(am, api.AppStopReasonKey)
	message = amAnnotation(am, api.AppStopMessageKey)
	if am == nil {
		return reason, message
	}
	if reason == "" {
		reason = am.Status.Reason
	}
	if message == "" {
		message = am.Status.Message
	}
	return reason, message
}

func amAnnotation(am *appv1.ApplicationManager, key string) string {
	if am == nil || am.Annotations == nil {
		return ""
	}
	return am.Annotations[key]
}

// minPositiveDuration returns the smallest positive duration among
// its args, or zero if none is positive. Used to fold multiple
// ScheduleTick effects into a single RequeueAfter — the earliest
// wake-up wins.
func minPositiveDuration(a, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// Runner is the impure shell that hosts the pure domain Engine.
// See doc.go for the full flow.
//
// Runner is safe for concurrent Reconcile calls. Internal state is
// either immutable after construction (engine, ports, config) or
// guarded by its own mutex (PendingActionQueue). One goroutine per
// in-flight CallWorkload is spawned via wg so shutdown can wait for
// drain.
type Runner struct {
	// Injected pure domain engine.
	engine *domain.Engine

	// Injected ports — the entire I/O surface. Fake in tests,
	// concrete adapters in production.
	amStore     ports.AMStore
	workloadOps ports.WorkloadOps
	validator   ports.InstallValidator
	imagePuller ports.ImagePuller
	recorder    ports.EventRecorder
	clock       ports.Clock

	// computeOps is OPTIONAL. Used to classify install-validation
	// rejections (shared-server blast radius) and to release
	// allocations on stop / uninstall. Nil degrades those writes
	// to no-ops so handler tests without a compute adapter still
	// walk the state machine.
	computeOps ports.ComputeOps

	// middlewareOps is OPTIONAL. Used to start/stop KubeBlocks
	// middleware alongside resume/stop and to uninstall the
	// pre-KubeBlocks os-platform mongodb chart. Nil skips those
	// steps, matching a cluster with no middleware apps.
	middlewareOps ports.MiddlewareOps

	// Pending-action delivery for async CallWorkload callbacks.
	pending  *PendingActionQueue
	requeuer Requeuer

	// downloadSlot is the process-local mutual exclusion for
	// Downloading. It does not span replicas; it serializes
	// check-then-act inside this process so two concurrent
	// Reconciles cannot both Patch to Downloading.
	downloadSlot downloadSlot

	// workloads tracks per-AM work that is either still running or
	// completed with a callback awaiting consumption.
	workloads workloadRegistry

	// orphans dedupes force-delete of an Application whose
	// ApplicationManager is already gone. The cleanup runs in a
	// goroutine, so a second Reconcile for the same name must not
	// start another namespace delete.
	orphans orphanCleanups

	// Runner-owned lifecycle. wg tracks in-flight CallWorkload
	// goroutines; ctx is the parent context every goroutine
	// derives from, so Stop() cascades cancellation.
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	// inFlight counts launched-but-unfinished CallWorkload
	// goroutines. Incremented synchronously by launchCallWorkload, so
	// it is already accurate when Reconcile returns.
	inFlight atomic.Int64
}

// InFlight reports how many CallWorkload goroutines are running.
//
// Exposed for observability and for tests that need to let a
// long-running dispatcher (the image pull polls until ImageManager
// reaches a terminal state) finish before asserting, rather than
// cancelling it via Stop.
func (r *Runner) InFlight() int { return int(r.inFlight.Load()) }

type workloadState uint8

const (
	workloadRunning workloadState = iota + 1
	workloadCallbackPending
)

// workloadClaim is one AM's slot in the registry. token identifies
// the individual CallWorkload goroutine holding it: a preempted
// goroutine may keep running (helm ignores context cancellation on
// several paths) but its token is gone from the map, which is how
// completeClaim knows to throw its result away.
//
// cancel tears down the goroutine's context when the claim is
// preempted. Ports that poll — WaitForLaunch, WaitForStartUp,
// WaitForNamespaceDeleted, the image-pull loop — return promptly
// once it fires; without it WaitForLaunch in particular polls until
// the process exits, since it has no deadline of its own and the
// state machine never revisits the abandoned goroutine.
type workloadClaim struct {
	state  workloadState
	token  uint64
	kind   domain.WorkloadKind
	cancel context.CancelFunc
}

type workloadRegistry struct {
	mu     sync.Mutex
	claims map[string]workloadClaim

	// nextToken is process-global and never reused, so a token from
	// an abandoned goroutine can never collide with a later claim on
	// the same AM.
	nextToken uint64
}

// claim reserves the AM's single workload slot and returns the token
// the dispatching goroutine must present to completeClaim. cancel is
// the goroutine's own context cancel, invoked if the claim is later
// preempted.
func (r *workloadRegistry) claim(name string, kind domain.WorkloadKind, cancel context.CancelFunc) (uint64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.claims == nil {
		r.claims = make(map[string]workloadClaim)
	}
	if _, exists := r.claims[name]; exists {
		return 0, false
	}
	r.nextToken++
	r.claims[name] = workloadClaim{state: workloadRunning, token: r.nextToken, kind: kind, cancel: cancel}
	return r.nextToken, true
}

// preemptClaim drops whatever claim the AM holds so an interrupting
// CallWorkload can take the slot, and cancels the abandoned
// goroutine's context so polling ports stop instead of running until
// Runner shutdown. Used by the cancel pipelines: a user canceling
// mid-download must reach ImagePuller.CancelJob even though the pull
// goroutine is still polling, and that pull is exactly what the
// cancel is there to stop.
//
// The cancel runs after the lock is released: it is non-blocking, but
// keeping foreign code out of the registry's critical section avoids
// any future lock-ordering surprise.
// cancelInFlight cancels the in-flight workload's context without
// dropping its claim. completeClaim still accepts the goroutine's
// result, so the state machine waits for that callback instead of
// starting another workload against the same helm release.
func (r *workloadRegistry) cancelInFlight(name string) bool {
	r.mu.Lock()
	c, exists := r.claims[name]
	r.mu.Unlock()
	if !exists || c.cancel == nil {
		return false
	}
	c.cancel()
	return true
}

func (r *workloadRegistry) preemptClaim(name string) bool {
	r.mu.Lock()
	c, exists := r.claims[name]
	if exists {
		delete(r.claims, name)
	}
	r.mu.Unlock()

	if !exists {
		return false
	}
	if c.cancel != nil {
		c.cancel()
	}
	return true
}

// completeClaim flips a running claim to callback-pending, keeping
// the AM claimed until Reconcile pops the completion — that closes
// the completion-to-consumption race.
//
// Returns false when this goroutine was preempted while it ran. Its
// result describes work the state machine has already moved past
// (an abandoned image pull reporting failure after the cancel
// landed), so the caller must drop it rather than deliver it.
func (r *workloadRegistry) completeClaim(name string, token uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, exists := r.claims[name]
	if !exists || c.token != token || c.state != workloadRunning {
		return false
	}
	c.state = workloadCallbackPending
	r.claims[name] = c
	return true
}

// acknowledgeCallback releases a claim that is waiting for its
// completion Action to be consumed. Reconcile calls it after Pop
// succeeds. A claim still in workloadRunning is left alone: the
// popped Action is a user operation that arrived while the workload
// goroutine is still running, and clearing the claim would let the
// next Tick start a second one.
//
// Returns true only when a callback-pending claim was deleted. The
// caller must restore that claim if the Reconcile then fails, or the
// completion is back in the queue while World.WorkloadInFlight says
// the slot is free.
func (r *workloadRegistry) acknowledgeCallback(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.claims[name].state == workloadCallbackPending {
		delete(r.claims, name)
		return true
	}
	return false
}

// restoreCallbackPending puts back the callback-pending claim that
// acknowledgeCallback removed, after the popped completion Action
// has been returned to the queue. The goroutine has already finished,
// so the restored claim carries no cancel.
//
// Returns false when the AM already holds a claim. That claim was
// taken while the failed Reconcile was unwinding and is newer than
// the completion being restored, so it must not be replaced.
func (r *workloadRegistry) restoreCallbackPending(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.claims == nil {
		r.claims = make(map[string]workloadClaim)
	}
	if _, exists := r.claims[name]; exists {
		return false
	}
	r.nextToken++
	r.claims[name] = workloadClaim{state: workloadCallbackPending, token: r.nextToken}
	return true
}

// outstanding reports whether this process still holds a claim for
// name: the CallWorkload goroutine is running, or it has finished
// and the completion callback is queued but not yet consumed.
// That is the snapshot copied onto World.WorkloadInFlight so a
// Tick handler can preserve the claim instead of re-dispatching.
func (r *workloadRegistry) outstanding(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.claims[name]
	return ok
}

func (r *workloadRegistry) kind(name string) domain.WorkloadKind {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claims[name].kind
}

// downloadSlot is a process-local exclusive holder for the image-
// pull slot. holder is the AM name currently allowed to be in
// Downloading; empty means free.
type downloadSlot struct {
	mu     sync.Mutex
	holder string
}

func (s *downloadSlot) tryAcquire(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.holder != "" && s.holder != name {
		return false
	}
	s.holder = name
	return true
}

func (s *downloadSlot) release(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.holder == name {
		s.holder = ""
	}
}

// Config bundles every construction-time input the Runner needs.
// Ports and Engine are mandatory; Clock and PendingActionQueue and
// Requeuer receive default values if left nil (real clock, fresh
// queue, no-op requeuer — the last only sensible for synchronous
// tests, override in production).
type Config struct {
	Engine *domain.Engine

	AMStore     ports.AMStore
	WorkloadOps ports.WorkloadOps
	Validator   ports.InstallValidator
	ImagePuller ports.ImagePuller
	Recorder    ports.EventRecorder
	Clock       ports.Clock

	// Optional. When nil, an install-validation rejection is
	// treated as not managing a shared server, so the resulting
	// stop is scoped to the app's own workloads. Stop/uninstall
	// allocation cleanup is also skipped.
	ComputeOps ports.ComputeOps

	// Optional. When nil, stop/resume skip the KubeBlocks
	// middleware step and uninstall skips the legacy mongodb path.
	MiddlewareOps ports.MiddlewareOps

	// Optional. Defaults to a fresh PendingActionQueue.
	Pending *PendingActionQueue

	// Optional. Defaults to NoopRequeuer. Production wiring
	// injects a controller-runtime source.Channel-backed requeuer.
	Requeuer Requeuer
}

// New constructs a Runner from cfg. Returns an error when any
// required field is missing (fast-fail; misconfiguration should
// crash the process, not silently no-op the state machine).
func New(cfg Config) (*Runner, error) {
	if cfg.Engine == nil {
		return nil, errors.New("runner.New: Config.Engine is required")
	}
	if cfg.AMStore == nil {
		return nil, errors.New("runner.New: Config.AMStore is required")
	}
	if cfg.WorkloadOps == nil {
		return nil, errors.New("runner.New: Config.WorkloadOps is required")
	}
	if cfg.Validator == nil {
		return nil, errors.New("runner.New: Config.Validator is required")
	}
	if cfg.ImagePuller == nil {
		return nil, errors.New("runner.New: Config.ImagePuller is required")
	}
	if cfg.Recorder == nil {
		return nil, errors.New("runner.New: Config.Recorder is required")
	}
	if cfg.Clock == nil {
		return nil, errors.New("runner.New: Config.Clock is required")
	}

	pending := cfg.Pending
	if pending == nil {
		pending = NewPendingActionQueue()
	}
	requeuer := cfg.Requeuer
	if requeuer == nil {
		requeuer = NoopRequeuer{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		engine:        cfg.Engine,
		amStore:       cfg.AMStore,
		workloadOps:   cfg.WorkloadOps,
		validator:     cfg.Validator,
		imagePuller:   cfg.ImagePuller,
		recorder:      cfg.Recorder,
		clock:         cfg.Clock,
		computeOps:    cfg.ComputeOps,
		middlewareOps: cfg.MiddlewareOps,
		pending:       pending,
		requeuer:      requeuer,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Stop cancels every in-flight CallWorkload goroutine and any orphan
// Application cleanup, then blocks until they drain. Idempotent.
func (r *Runner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// Pending exposes the PendingActionQueue so external callers (test
// code, or a controller-runtime source.Channel adapter) can Deliver
// synthetic Actions without going through a CallWorkload goroutine.
func (r *Runner) Pending() *PendingActionQueue { return r.pending }

// Request identifies one AM to reconcile. Kept separate from
// controller-runtime's ctrl.Request so this package does not
// transitively import controller-runtime; the apiserver-collapse
// step wires the two types together with a trivial adapter.
type Request struct {
	Name string
}

// ReconcileOutcome is the Runner's answer to a Reconcile call. Maps
// 1:1 to controller-runtime's ctrl.Result (adapter one-liner in the
// apiserver-collapse step).
type ReconcileOutcome struct {
	// RequeueAfter, when > 0, asks controller-runtime to re-invoke
	// Reconcile after the given delay. Derived from ScheduleTick
	// effects.
	RequeueAfter time.Duration

	// Requeue, when true, asks controller-runtime to re-invoke
	// Reconcile immediately without backoff. Set on optimistic-
	// concurrency conflicts (PatchStatus returned IsConflict).
	Requeue bool
}

// errConflict is the sentinel applyPatchAM returns on IsConflict.
// Reconcile unwraps it into ReconcileOutcome.Requeue = true.
var errConflict = errors.New("runner: patch conflict")

// errDownloadSlotBusy is returned when applyPatchAM refuses a
// transition into Downloading because another AM already holds the
// in-process slot. Reconcile unwraps it into RequeueAfter.
var errDownloadSlotBusy = errors.New("runner: download slot busy")

// downloadSlotBusyRequeue matches handlers' pending bed-full backoff
// so a lost race retries on the same cadence as FIFO wait.
const downloadSlotBusyRequeue = 1 * time.Second

// Reconcile is the entry point. Signature is deliberately close to
// controller-runtime's Reconciler so the apiserver-collapse step is
// a one-liner.
//
// Flow (also documented in doc.go):
//
//  1. Load AM. NotFound → release the download slot. If the
//     Application is still there, force-delete its namespace and
//     emit the uninstall events.
//  2. Materialise World (clock, ClusterStats, AppExists).
//  3. Pop pending Action or default to Tick{}.
//  4. engine.Step(world, action). NoTransition / GuardRejected →
//     no-op. Other errors → propagate.
//  5. Dispatch result.Effects. Sentinel errConflict → Requeue=true.
//  6. Return ReconcileOutcome (RequeueAfter, Requeue).
func (r *Runner) Reconcile(ctx context.Context, req Request) (ReconcileOutcome, error) {
	world, ok, err := r.buildWorld(ctx, req.Name)
	if err != nil {
		return ReconcileOutcome{}, err
	}
	if !ok {
		// AM has been deleted. Nothing to reconcile, but a Downloading
		// holder that disappears without a leaving-state patch would
		// otherwise pin the in-process slot until process restart.
		// An Application left behind is force-uninstalled: its
		// namespace is deleted and the uninstall events are emitted.
		r.downloadSlot.release(req.Name)
		if err := r.cleanupOrphanApplication(ctx, req.Name); err != nil {
			return ReconcileOutcome{}, err
		}
		return ReconcileOutcome{}, nil
	}

	// popped is nil when nothing was queued and we synthesised a
	// Tick. Only a real popped Action is worth restoring on failure
	// — a synthetic Tick is regenerated by the next Reconcile.
	popped := r.pending.Pop(req.Name)
	klog.Infof("Reconcile: req.Name: %v", req.Name)
	klog.Infof("Reconcile: popped: %v", popped)
	callbackAcknowledged := false
	if popped != nil {
		klog.Infof("Reconcile: kind: %v", popped.Kind())
		// Clear only a completed claim. A queued user action must
		// never release a workload that is still running.
		callbackAcknowledged = r.workloads.acknowledgeCallback(req.Name)
	}
	// buildWorld ran before the queue pop. Refresh the process-local
	// snapshot so a callback handler sees its acknowledged claim as
	// consumed, while a user Action popped during running work still
	// sees that work as outstanding.
	world.WorkloadInFlight = r.workloads.outstanding(req.Name)
	world.InFlightKind = r.workloads.kind(req.Name)
	klog.Infof("Reconcile: callbackAcknowledged: %v", callbackAcknowledged)
	action := popped
	if action == nil {
		requestID := world.AM.Annotations[domain.AnnotationOperationRequestID]
		if requestID != "" &&
			requestID != world.AM.Status.OpID &&
			r.engine.Table().AllowsOp(world.Current(), world.AM.Spec.OpType) {
			action = domain.TriggerOp{Op: world.AM.Spec.OpType}
		} else {
			action = domain.Tick{}
		}
	}
	// Thread the Action into World so handlers can inspect
	// Action-specific fields (WorkloadFailed.Reason etc.) without
	// changing the HandlerFunc signature. Engine.Step ignores this
	// field for dispatch — it always keys on the Action passed as
	// its third arg.
	world.Action = action

	timedOut, deadlineAfter := r.deadlineState(world)
	// A callback-pending claim means the workload already returned and
	// its result is sitting in popped. Consuming that result wins over
	// this Reconcile happening after the deadline. A still-running
	// workload, a Tick, or a user operation does not: those are replaced
	// with TimeoutFired and the in-flight claim is preempted so a late
	// result is discarded.
	if timedOut && !callbackAcknowledged {
		r.workloads.preemptClaim(req.Name)
		action = domain.TimeoutFired{}
		popped = action
		callbackAcknowledged = false
		world.Action = action
		world.WorkloadInFlight = r.workloads.outstanding(req.Name)
		world.InFlightKind = r.workloads.kind(req.Name)
	}

	res, err := r.engine.Step(ctx, world, action)
	if err != nil {
		// Classify: table-level "no matching edge" is a no-op;
		// anything else is an actual error the controller-runtime
		// should back off on.
		var noTransition domain.ErrNoTransition
		var guardRejected domain.ErrGuardRejected
		if errors.As(err, &noTransition) || errors.As(err, &guardRejected) {
			// The Action genuinely does not apply in this state.
			// Dropping it is correct; restoring would resurrect it
			// against some later, unrelated state.
			return r.withDeadline(ReconcileOutcome{}, deadlineAfter, timedOut), nil
		}
		r.restorePendingAction(req.Name, popped, callbackAcknowledged)
		return ReconcileOutcome{}, fmt.Errorf("runner: engine.Step %s: %w", req.Name, err)
	}

	outcome, dispatchErr := r.dispatchResult(ctx, world, res)
	outcome = r.withDeadline(outcome, deadlineAfter, timedOut)
	klog.Infof("Reconcile: dispatchResult.outcome: %#v", outcome)
	if dispatchErr != nil {
		// Every dispatch failure means the Result did not land in
		// full, so the Action that produced it must survive into the
		// retry the outcome schedules below.
		r.restorePendingAction(req.Name, popped, callbackAcknowledged)
		if errors.Is(dispatchErr, errConflict) {
			outcome.Requeue = true
			return outcome, nil
		}
		if errors.Is(dispatchErr, errDownloadSlotBusy) {
			outcome.RequeueAfter = downloadSlotBusyRequeue
			return outcome, nil
		}
		return outcome, dispatchErr
	}
	return outcome, nil
}

// deadlineState reports whether the AM's persisted state entry time
// has reached its CallWorkload TTL. A live deadline also returns how
// long remains so Reconcile can wake at expiry without an object event.
func (r *Runner) deadlineState(world domain.World) (expired bool, remaining time.Duration) {
	if world.AM == nil {
		return false, 0
	}
	ttl, ok := r.engine.Table().Timeout(world.Current())
	if !ok {
		return false, 0
	}
	deadline := stateEntryTime(world.AM, world.Now).Add(ttl)
	if !world.Now.Before(deadline) {
		return true, 0
	}
	return false, deadline.Sub(world.Now)
}

func (r *Runner) withDeadline(out ReconcileOutcome, remaining time.Duration, expired bool) ReconcileOutcome {
	if expired || remaining <= 0 {
		return out
	}
	out.RequeueAfter = minPositiveDuration(out.RequeueAfter, remaining)
	return out
}

// stateEntryTime is when the AM entered its current state. Legacy
// objects may lack StatusTime; UpdateTime then CreationTimestamp are
// the fallbacks. If none exist, now is used so incomplete data is not
// treated as already expired.
func stateEntryTime(am *appv1.ApplicationManager, now time.Time) time.Time {
	if am != nil && am.Status.StatusTime != nil && !am.Status.StatusTime.IsZero() {
		return am.Status.StatusTime.Time
	}
	if am != nil && am.Status.UpdateTime != nil && !am.Status.UpdateTime.IsZero() {
		return am.Status.UpdateTime.Time
	}
	if am != nil && !am.CreationTimestamp.IsZero() {
		return am.CreationTimestamp.Time
	}
	return now
}

func (r *Runner) restorePendingAction(name string, action domain.Action, callbackAcknowledged bool) {
	if !r.pending.Restore(name, action) || !callbackAcknowledged {
		return
	}
	// A workload may have launched or completed while the failed
	// Reconcile was unwinding. Never replace that fresher claim.
	r.workloads.restoreCallbackPending(name)
}

// buildWorld materialises a domain.World for the Reconcile currently
// in progress. Two-value return + error keeps the "AM not found"
// case unambiguous.
//
// Skeleton implementation notes:
//
//   - ClusterStats.DownloadingCount is computed by AMStore.List and
//     then counting states in-process. In production the AMStore
//     adapter delegates to an informer cache so List is O(1) and
//     the count itself is O(N) over cached objects.
//   - WorkloadStatus is left zero — the informer-backed World
//     builder will populate it later. Handlers that need Workload
//     signals must be added along with the corresponding builder
//     enrichment; the skeleton documents this dependency in the
//     handler's package comment.
func (r *Runner) buildWorld(ctx context.Context, name string) (domain.World, bool, error) {
	am, err := r.amStore.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return domain.World{}, false, nil
		}
		return domain.World{}, false, fmt.Errorf("runner: AMStore.Get %s: %w", name, err)
	}

	appExists := true
	if _, err := r.amStore.GetApplication(ctx, name); err != nil {
		if !apierrors.IsNotFound(err) {
			return domain.World{}, false, fmt.Errorf("runner: AMStore.GetApplication %s: %w", name, err)
		}
		appExists = false
	}

	stats, err := r.computeClusterStats(ctx)
	if err != nil {
		return domain.World{}, false, err
	}

	return domain.World{
		AM:                  am,
		Now:                 r.clock.Now(),
		AppExists:           appExists,
		WorkloadInFlight:    r.workloads.outstanding(name),
		InFlightKind:        r.workloads.kind(name),
		ClusterStats:        stats,
		HasWorkloadReplicas: hasWorkloadReplicas(am),
	}, true, nil
}

// hasWorkloadReplicas materialises domain.World.HasWorkloadReplicas
// from the AM's Spec.Config blob.
//
// A parse failure degrades to false rather than aborting the
// Reconcile: false selects the single-phase install, which is the
// conservative branch (validate before helm, no scale-up), and an
// unparseable Spec.Config will fail the install anyway once
// launchCallWorkload materialises the same blob for the workload
// port — with a real error message attached. Failing here instead
// would report the parse error as a world-build failure and hide it
// behind controller-runtime's backoff.
func hasWorkloadReplicas(am *appv1.ApplicationManager) bool {
	if am == nil || am.Spec.Config == "" {
		return false
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(am.Spec.Config), &cfg); err != nil {
		klog.Warningf("runner: unmarshal Spec.Config for %s: %v; treating as single-phase install", am.Name, err)
		return false
	}
	return cfg.HasWorkloadReplicas()
}

// computeClusterStats folds AMStore.List into the aggregates
// handlers need. Currently only DownloadingCount (used by the
// PendingTickHandler to enforce the "at most one Downloading" gate).
func (r *Runner) computeClusterStats(ctx context.Context) (domain.ClusterStats, error) {
	all, err := r.amStore.List(ctx)
	if err != nil {
		return domain.ClusterStats{}, fmt.Errorf("runner: AMStore.List: %w", err)
	}
	stats := domain.ClusterStats{}
	type fifoCand struct {
		name string
		at   time.Time
	}
	var pending []fifoCand
	for i := range all {
		switch all[i].Status.State {
		case appv1.Downloading, appv1.UpgradeDownloading:
			stats.DownloadingCount++
		case appv1.Pending, appv1.UpgradePending:
			pending = append(pending, fifoCand{
				name: all[i].Name,
				at:   pendingQueueTime(all[i]),
			})
		}
	}
	if len(pending) > 0 {
		sort.Slice(pending, func(i, j int) bool {
			if !pending[i].at.Equal(pending[j].at) {
				return pending[i].at.Before(pending[j].at)
			}
			return pending[i].name < pending[j].name
		})
		stats.NextDownloading = pending[0].name
	}
	return stats, nil
}

// pendingQueueTime is the FIFO key for a Pending AM. Prefer the
// once-written pending.queuedAt payload; fall back to the object's
// creation timestamp so AMs seeded without the key still order
// stably (then by name in computeClusterStats).
func pendingQueueTime(am appv1.ApplicationManager) time.Time {
	if am.Status.Payload != nil {
		if s := am.Status.Payload[domain.PayloadPendingQueuedAt]; s != "" {
			if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
				return t
			}
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t
			}
		}
	}
	if !am.CreationTimestamp.IsZero() {
		return am.CreationTimestamp.Time
	}
	return time.Time{}
}

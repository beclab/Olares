package domain

import (
	"context"
	"os"
	"strings"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// NewDefaultTable returns a TransitionTable populated with every legacy
// edge from pkg/appstate/state_transition.go. Handlers for these edges
// are placeholders that simply set Result.Next; real handlers land in
// pkg/applifecycle/handlers/ in a follow-up PR (todo item "handlers").
//
// See table_test.go for the migration-correctness proofs that assert this
// table matches the legacy tables (StateTransitions,
// OperationAllowedInState, CancelableStates, OperatingStates,
// CancelingStates) plus IsReinstallable and
// IsSafelyDeletable.
func NewDefaultTable() *TransitionTable {
	t := NewTransitionTable()
	setStateSpecs(t)
	setTimeouts(t)
	registerEdges(t)
	return t
}

// DefaultTable is the process-wide transition table. cmd/app-service
// registers handlers on this instance, and operation prechecks call
// DefaultTable.AllowsOp so they see the same TriggerOp edges the
// runner will accept.
var DefaultTable = NewDefaultTable()

// setTimeouts enables deadlines for states whose handlers dispatch
// CallWorkload, and for canceling states that wait on a helm call
// started in the previous state. Pending and UpgradePending only wait
// for the download slot, so they have no deadline. InstallFailed has
// none either: its cleanup retries until the namespace is gone, and a
// same-state write does not refresh StatusTime, so an expired deadline
// would preempt every retry.
//
// Each deadline can be overridden with APPLIFECYCLE_TIMEOUT_<STATE>,
// for example APPLIFECYCLE_TIMEOUT_INSTALLING=5m. The value must be a
// positive time.ParseDuration string. An unset, blank, or invalid
// value keeps the default below.
func setTimeouts(t *TransitionTable) {
	for _, spec := range defaultStateTimeouts() {
		t.SetTimeout(spec.state, resolveStateTimeout(spec.env, spec.fallback))
	}
}

type stateTimeout struct {
	state    State
	env      string
	fallback time.Duration
}

func defaultStateTimeouts() []stateTimeout {
	ten := 10 * time.Minute
	return []stateTimeout{
		{appsv1.Downloading, "APPLIFECYCLE_TIMEOUT_DOWNLOADING", 24 * time.Hour},
		{appsv1.UpgradeDownloading, "APPLIFECYCLE_TIMEOUT_UPGRADE_DOWNLOADING", 24 * time.Hour},
		{appsv1.Installing, "APPLIFECYCLE_TIMEOUT_INSTALLING", 30 * time.Minute},
		{appsv1.Initializing, "APPLIFECYCLE_TIMEOUT_INITIALIZING", time.Hour},
		{appsv1.Upgrading, "APPLIFECYCLE_TIMEOUT_UPGRADING", time.Hour},
		{appsv1.ApplyingEnv, "APPLIFECYCLE_TIMEOUT_APPLYING_ENV", 30 * time.Minute},
		{appsv1.Resuming, "APPLIFECYCLE_TIMEOUT_RESUMING", time.Hour},
		{appsv1.Stopping, "APPLIFECYCLE_TIMEOUT_STOPPING", 30 * time.Minute},
		{appsv1.Uninstalling, "APPLIFECYCLE_TIMEOUT_UNINSTALLING", 30 * time.Minute},
		{appsv1.PendingCanceling, "APPLIFECYCLE_TIMEOUT_PENDING_CANCELING", ten},
		{appsv1.DownloadingCanceling, "APPLIFECYCLE_TIMEOUT_DOWNLOADING_CANCELING", ten},
		{appsv1.UpgradePendingCanceling, "APPLIFECYCLE_TIMEOUT_UPGRADE_PENDING_CANCELING", ten},
		{appsv1.UpgradeDownloadingCanceling, "APPLIFECYCLE_TIMEOUT_UPGRADE_DOWNLOADING_CANCELING", ten},
		{appsv1.InstallingCanceling, "APPLIFECYCLE_TIMEOUT_INSTALLING_CANCELING", ten},
		{appsv1.InitializingCanceling, "APPLIFECYCLE_TIMEOUT_INITIALIZING_CANCELING", ten},
		{appsv1.UpgradingCanceling, "APPLIFECYCLE_TIMEOUT_UPGRADING_CANCELING", ten},
		{appsv1.ResumingCanceling, "APPLIFECYCLE_TIMEOUT_RESUMING_CANCELING", ten},
		{appsv1.ApplyingEnvCanceling, "APPLIFECYCLE_TIMEOUT_APPLYING_ENV_CANCELING", ten},
	}
}

// resolveStateTimeout returns fallback unless env holds a positive
// time.ParseDuration value such as 5m or 1h. Blank, zero, negative,
// and unparsable values keep fallback.
func resolveStateTimeout(env string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(env))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

// PlaceholderHandler is the stub used by NewDefaultTable while real
// per-state handlers are not yet implemented. It only advances the state
// to next and does not emit any effect; the Runner will still perform
// the PatchAM because Engine.Step validates the transition against the
// table, but no side effect fires.
//
// Kept exported so tests and the follow-up handlers package can compose
// against it (e.g. wrap a real handler that only decides Next but reuses
// PlaceholderHandler for its effect construction).
func PlaceholderHandler(next State) HandlerFunc {
	return func(_ context.Context, _ World) Result {
		return Result{Next: next}
	}
}

// setStateSpecs populates per-state metadata for every state, matching
// the legacy classification maps.
func setStateSpecs(t *TransitionTable) {
	// Cancelable matches legacy CancelableStates: the operating states
	// that have an outgoing CancelOp transition.
	cancelable := map[State]bool{
		appsv1.Pending:            true,
		appsv1.Downloading:        true,
		appsv1.UpgradePending:     true,
		appsv1.UpgradeDownloading: true,
		appsv1.Installing:         true,
		appsv1.Initializing:       true,
		appsv1.Resuming:           true,
		appsv1.Upgrading:          true,
		appsv1.ApplyingEnv:        true,
	}

	// Canceling matches legacy CancelingStates: the seven *Canceling
	// mid-cancellation states.
	canceling := map[State]bool{
		appsv1.PendingCanceling:            true,
		appsv1.DownloadingCanceling:        true,
		appsv1.UpgradePendingCanceling:     true,
		appsv1.UpgradeDownloadingCanceling: true,
		appsv1.InstallingCanceling:         true,
		appsv1.InitializingCanceling:       true,
		appsv1.ResumingCanceling:           true,
		appsv1.UpgradingCanceling:          true,
		appsv1.ApplyingEnvCanceling:        true,
	}

	// Operating matches legacy OperatingStates: every state that is
	// actively driving external work. Union of Cancelable + Canceling +
	// Stopping + Uninstalling.
	operating := map[State]bool{
		appsv1.Stopping:     true,
		appsv1.Uninstalling: true,
	}
	for s := range cancelable {
		operating[s] = true
	}
	for s := range canceling {
		operating[s] = true
	}

	// Reinstallable / SafelyDeletable match the same-named
	// legacy funcs: clean terminal states from which a fresh Install is
	// legal.
	reinstallable := map[State]bool{
		appsv1.PendingCanceled:     true,
		appsv1.DownloadingCanceled: true,
		appsv1.InstallingCanceled:  true,
		appsv1.Uninstalled:         true,
		appsv1.DownloadFailed:      true,
	}

	for _, s := range AllStates {
		t.SetSpec(s, StateSpec{
			Operating:       operating[s],
			Canceling:       canceling[s],
			Cancelable:      cancelable[s],
			Reinstallable:   reinstallable[s],
			SafelyDeletable: reinstallable[s],
		})
	}
}

// registerEdges registers every legacy transition from
// StateTransitions[from] = [to...]. The Action attached to each edge is
// chosen per the following priority (this is the "6 tables into 1"
// consolidation):
//
//  1. If (from, op) exists in OperationAllowedInState and op's
//     canonical target state (opTarget below) matches to, the edge
//     fires on TriggerOp{op}.
//  2. Else if to is a *Canceling state and from is that state's
//     matching operating state, the edge fires on TriggerOp{CancelOp}.
//  3. Else if to is a *Failed / *CancelFailed state and from is an
//     operating state, the edge fires on WorkloadFailed{}.
//  4. Else if from is Pending or UpgradePending, the edge fires on
//     Tick{}. Those states only wait for the download slot; they do
//     not dispatch CallWorkload, so a WorkloadReady classification
//     would leave a placeholder that SetOrAddHandler cannot replace.
//  5. Else if from is an operating state (including Canceling), the
//     edge fires on WorkloadReady{}.
//  6. Otherwise the edge fires on Tick{} (reconcile-driven, e.g.
//     "" -> Pending sentinel bootstrap in the legacy flow, terminal ->
//     Pending reinstall recovery). Missing Application CR self-heal
//     leaves Running for Uninstalling, which is the UninstallOp edge.
//
// When two edges from the same state share an Action kind, only the
// FIRST is registered as a dispatchable Transition; every additional
// target is recorded in reachable so Allows() reports it as legal.
// The one handler for that (from, kind) chooses which target to
// return. This keeps table lookup deterministic while still declaring
// the full transition topology.
func registerEdges(t *TransitionTable) {
	// legacy StateTransitions verbatim, expressed as (from, to) pairs.
	// Kept as an explicit slice literal so a reader can diff this
	// against pkg/appstate/state_transition.go without spelunking.
	type edge struct {
		from, to State
	}
	edges := []edge{
		// "" -> Pending (initial InstallOp bootstrap).
		{InitialState, appsv1.Pending},

		// Pending
		{appsv1.Pending, appsv1.Downloading},
		{appsv1.Pending, appsv1.PendingCanceling},

		// Downloading
		{appsv1.Downloading, appsv1.Installing},
		{appsv1.Downloading, appsv1.DownloadFailed},
		{appsv1.Downloading, appsv1.DownloadingCanceling},

		// Installing (multi-target; two of them share a resolution
		// path today — see installing_app.go)
		{appsv1.Installing, appsv1.Initializing},
		{appsv1.Installing, appsv1.InstallFailed},
		{appsv1.Installing, appsv1.InstallingCanceling},
		{appsv1.Installing, appsv1.Stopping},
		{appsv1.Installing, appsv1.Running}, // middleware fast-path

		// Initializing
		{appsv1.Initializing, appsv1.Running},
		{appsv1.Initializing, appsv1.InitializingCanceling},

		// Running. A missing Application CR self-heals through
		// Uninstalling; there is no direct jump to Uninstalled.
		{appsv1.Running, appsv1.Stopping},
		{appsv1.Running, appsv1.UpgradePending},
		{appsv1.Running, appsv1.ApplyingEnv},
		{appsv1.Running, appsv1.Uninstalling},

		// Stopping (workload callback only)
		{appsv1.Stopping, appsv1.Stopped},
		{appsv1.Stopping, appsv1.StopFailed},

		// Upgrading
		{appsv1.Upgrading, appsv1.Initializing},
		{appsv1.Upgrading, appsv1.UpgradeFailed},
		{appsv1.Upgrading, appsv1.UpgradingCanceling},
		{appsv1.Upgrading, appsv1.Stopped}, // upgrade-from-Stopped fast-path

		// Upgrade download queue. Install Pending and UpgradePending share
		// one cluster-wide image-pull slot (Downloading or UpgradeDownloading).
		{appsv1.UpgradePending, appsv1.UpgradeDownloading},
		{appsv1.UpgradePending, appsv1.UpgradePendingCanceling},
		{appsv1.UpgradeDownloading, appsv1.Upgrading},
		{appsv1.UpgradeDownloading, appsv1.UpgradeFailed},
		{appsv1.UpgradeDownloading, appsv1.UpgradeDownloadingCanceling},
		{appsv1.UpgradePendingCanceling, appsv1.Running},
		{appsv1.UpgradePendingCanceling, appsv1.Stopped},
		{appsv1.UpgradePendingCanceling, appsv1.UpgradePendingCancelFailed},
		{appsv1.UpgradeDownloadingCanceling, appsv1.Running},
		{appsv1.UpgradeDownloadingCanceling, appsv1.Stopped},
		{appsv1.UpgradeDownloadingCanceling, appsv1.UpgradeDownloadingCancelFailed},

		// ApplyingEnv
		{appsv1.ApplyingEnv, appsv1.Initializing},
		{appsv1.ApplyingEnv, appsv1.ApplyEnvFailed},
		{appsv1.ApplyingEnv, appsv1.ApplyingEnvCanceling},
		{appsv1.ApplyingEnv, appsv1.Stopping}, // pod_abnormal_suspend fallback
		{appsv1.ApplyingEnv, appsv1.Stopped},  // applyEnv-from-Stopped fast-path

		// Uninstalling (workload callback only; no op is allowed)
		{appsv1.Uninstalling, appsv1.Uninstalled},
		{appsv1.Uninstalling, appsv1.UninstallFailed},

		// *Canceling states each go to their *Canceled or *CancelFailed
		// terminal (or Stopping for Initializing/Resuming/ApplyingEnv).
		// UpgradingCanceling lands in UpgradeFailed so a canceled helm
		// upgrade can be retried without tearing the release down.
		{appsv1.PendingCanceling, appsv1.PendingCanceled},
		{appsv1.PendingCanceling, appsv1.PendingCancelFailed},
		{appsv1.DownloadingCanceling, appsv1.DownloadingCanceled},
		{appsv1.DownloadingCanceling, appsv1.DownloadingCancelFailed},
		{appsv1.InstallingCanceling, appsv1.InstallingCanceled},
		{appsv1.InstallingCanceling, appsv1.InstallingCancelFailed},
		// Install failures share the InstallingCanceling cleanup and
		// land here instead of InstallingCanceled — see
		// handlers.OnInstallingCancelingWorkloadReady. This table
		// entry adds reachability only: registerEdge's generic
		// failure-target classification finds the WorkloadFailed slot
		// already occupied by InstallingCancelFailed. handlers.Register
		// supplies the actual WorkloadReady callback, which selects
		// this target from the persisted terminal intent.
		{appsv1.InstallingCanceling, appsv1.InstallFailed},
		{appsv1.InitializingCanceling, appsv1.Stopping},
		// Ten-minute timeout. A launch callback is not this edge: the
		// canceling handler hands that off to Stopping.
		{appsv1.InitializingCanceling, appsv1.InitializingCancelFailed},
		{appsv1.ResumingCanceling, appsv1.Stopping},
		{appsv1.ResumingCanceling, appsv1.ResumingCancelFailed},
		{appsv1.UpgradingCanceling, appsv1.UpgradeFailed},
		{appsv1.UpgradingCanceling, appsv1.UpgradingCancelFailed},
		{appsv1.ApplyingEnvCanceling, appsv1.Stopping},
		{appsv1.ApplyingEnvCanceling, appsv1.ApplyingEnvCancelFailed},

		// Resuming
		{appsv1.Resuming, appsv1.Initializing},
		{appsv1.Resuming, appsv1.ResumeFailed},
		{appsv1.Resuming, appsv1.ResumingCanceling},
		{appsv1.Resuming, appsv1.Stopping}, // StopOp allowed

		// Stopped (only op-driven exits)
		{appsv1.Stopped, appsv1.Resuming},
		{appsv1.Stopped, appsv1.Uninstalling},
		{appsv1.Stopped, appsv1.UpgradePending},
		{appsv1.Stopped, appsv1.ApplyingEnv},

		// *Failed / terminal-reinstallable recovery paths.
		{appsv1.DownloadFailed, appsv1.Pending},
		{appsv1.InstallFailed, appsv1.Uninstalled},
		{appsv1.InstallFailed, appsv1.Uninstalling},
		{appsv1.StopFailed, appsv1.Stopping},
		{appsv1.StopFailed, appsv1.Uninstalling},
		{appsv1.UpgradeFailed, appsv1.UpgradePending},
		{appsv1.UpgradeFailed, appsv1.Uninstalling},
		{appsv1.ApplyEnvFailed, appsv1.Stopping},
		{appsv1.ApplyEnvFailed, appsv1.ApplyingEnv},
		{appsv1.ApplyEnvFailed, appsv1.Uninstalling},
		{appsv1.ResumeFailed, appsv1.Resuming},
		{appsv1.ResumeFailed, appsv1.Uninstalling},
		{appsv1.ResumeFailed, appsv1.ApplyingEnv},
		{appsv1.UninstallFailed, appsv1.Uninstalling},
		{appsv1.PendingCancelFailed, appsv1.PendingCanceling},
		{appsv1.DownloadingCancelFailed, appsv1.DownloadingCanceling},
		{appsv1.UpgradePendingCancelFailed, appsv1.UpgradePendingCanceling},
		{appsv1.UpgradeDownloadingCancelFailed, appsv1.UpgradeDownloadingCanceling},
		{appsv1.InstallingCancelFailed, appsv1.InstallingCanceling},
		{appsv1.InstallingCancelFailed, appsv1.Uninstalling},
		{appsv1.InitializingCancelFailed, appsv1.InitializingCanceling},
		{appsv1.InitializingCancelFailed, appsv1.Uninstalling},
		{appsv1.UpgradingCancelFailed, appsv1.UpgradingCanceling},
		{appsv1.UpgradingCancelFailed, appsv1.Uninstalling},
		{appsv1.ApplyingEnvCancelFailed, appsv1.ApplyingEnvCanceling},
		{appsv1.ApplyingEnvCancelFailed, appsv1.Uninstalling},
		{appsv1.ResumingCancelFailed, appsv1.ResumingCanceling},

		// Terminal-reinstallable re-target: Install operation on a
		// same-owner AM flips state in place.
		{appsv1.PendingCanceled, appsv1.Pending},
		{appsv1.DownloadingCanceled, appsv1.Pending},
		{appsv1.InstallingCanceled, appsv1.Pending},
		{appsv1.Uninstalled, appsv1.Pending},
	}

	for _, e := range edges {
		registerEdge(t, e.from, e.to)
	}
}

// registerEdge attaches action semantics to a from->to edge per the
// priority described on registerEdges. It always records reachability;
// it only registers a dispatchable Transition when this is the first
// edge from (from, actionKind).
func registerEdge(t *TransitionTable, from, to State) {
	action := actionForEdge(from, to)

	t.AddReachable(from, to)

	// Skip dispatchable registration if an edge for this
	// (from, kind) is already registered. Additional targets stay
	// reachable-only; the handler on the first edge returns the
	// appropriate Next.
	if _, ok := t.Edge(from, action.Kind()); ok {
		return
	}

	t.Add(Transition{
		From:    from,
		On:      action,
		Handler: PlaceholderHandler(to),
	})
}

// actionForEdge picks the Action that fires the from->to transition per
// the priority documented on registerEdges. Kept pure so the choice can
// be re-computed by tests.
func actionForEdge(from, to State) Action {
	// Priority 1: op-driven, with a canonical target (includes CancelOp,
	// which fires whenever to is a *Canceling state and from allows
	// CancelOp — see opForEdge).
	if op, ok := opForEdge(from, to); ok {
		return TriggerOp{Op: op}
	}
	// Priority 2: workload failure into a *Failed / *CancelFailed
	// terminal from an operating source.
	if isOperatingSource(from) && isFailureTarget(to) {
		return WorkloadFailed{}
	}
	// Priority 3: download-queue states advance on reconcile. They are
	// operating, but they do not run CallWorkload, so the generic
	// operating-source rule below must not claim these edges.
	if from == appsv1.Pending || from == appsv1.UpgradePending {
		return Tick{}
	}
	// Priority 4: workload success out of an operating state.
	if isOperatingSource(from) {
		return WorkloadReady{}
	}
	// Priority 5: reconcile-driven (bootstrap / terminal-reinstall).
	return Tick{}
}

// StopRemainsAllowed reports whether StopOp is still legal after op is
// consumed from state. An automatic stop must leave a request alone when
// its landing state rejects Stop: once the Runner moves there, the stop
// could not have been submitted. ApplyEnv lands in ApplyingEnv, which
// still allows Stop.
func StopRemainsAllowed(from State, op OpType) bool {
	next, ok := opLandingState(from, op)
	if !ok {
		return false
	}
	return DefaultTable.AllowsOp(next, appsv1.StopOp)
}

// opLandingState is the state an accepted op enters from `from`. It is
// the inverse of the canonical targets opForEdge recognizes.
func opLandingState(from State, op OpType) (State, bool) {
	if !opAllowedLegacy(from, op) {
		return "", false
	}
	switch op {
	case appsv1.UninstallOp:
		return appsv1.Uninstalling, true
	case appsv1.UpgradeOp:
		return appsv1.UpgradePending, true
	case appsv1.ApplyEnvOp:
		return appsv1.ApplyingEnv, true
	case appsv1.ResumeOp:
		return appsv1.Resuming, true
	case appsv1.StopOp:
		return appsv1.Stopping, true
	case appsv1.InstallOp:
		return appsv1.Pending, true
	case appsv1.CancelOp:
		return cancelLandingState(from)
	default:
		return "", false
	}
}

func cancelLandingState(from State) (State, bool) {
	switch from {
	case appsv1.Pending, appsv1.PendingCancelFailed:
		return appsv1.PendingCanceling, true
	case appsv1.Downloading, appsv1.DownloadingCancelFailed:
		return appsv1.DownloadingCanceling, true
	case appsv1.UpgradePending, appsv1.UpgradePendingCancelFailed:
		return appsv1.UpgradePendingCanceling, true
	case appsv1.UpgradeDownloading, appsv1.UpgradeDownloadingCancelFailed:
		return appsv1.UpgradeDownloadingCanceling, true
	case appsv1.Installing, appsv1.InstallingCancelFailed:
		return appsv1.InstallingCanceling, true
	case appsv1.Initializing, appsv1.InitializingCancelFailed:
		return appsv1.InitializingCanceling, true
	case appsv1.Upgrading, appsv1.UpgradingCancelFailed:
		return appsv1.UpgradingCanceling, true
	case appsv1.ApplyingEnv, appsv1.ApplyingEnvCancelFailed:
		return appsv1.ApplyingEnvCanceling, true
	case appsv1.Resuming, appsv1.ResumingCancelFailed:
		return appsv1.ResumingCanceling, true
	default:
		return "", false
	}
}

// opForEdge returns the operation type that drives from->to when the
// edge matches the canonical (state, op) mapping in
// OperationAllowedInState. When to matches op's canonical target and
// the source state allows that op in the legacy table, the edge is
// classified as op-driven.
func opForEdge(from, to State) (OpType, bool) {
	// CancelOp is the odd op out: its target depends on the source
	// state, so instead of one canonical target we recognize any
	// *Canceling state as the CancelOp landing point (from either an
	// operating state or the corresponding *CancelFailed state).
	if isCancelingState(to) && opAllowedLegacy(from, appsv1.CancelOp) {
		return appsv1.CancelOp, true
	}

	// Canonical op -> target-state mapping. Same values as the
	// legacy state_flow_lineages_test.go's opTarget map.
	switch to {
	case appsv1.Pending:
		// Only real Install triggers can pick Pending as target.
		// The self-loop pattern Pending<-Fooo is InstallOp when
		// the source is a terminal-reinstallable state, otherwise
		// it stays reconcile-driven.
		if opAllowedLegacy(from, appsv1.InstallOp) {
			return appsv1.InstallOp, true
		}
	case appsv1.Uninstalling:
		if opAllowedLegacy(from, appsv1.UninstallOp) {
			return appsv1.UninstallOp, true
		}
	case appsv1.UpgradePending:
		if opAllowedLegacy(from, appsv1.UpgradeOp) {
			return appsv1.UpgradeOp, true
		}
	case appsv1.Upgrading:
		if opAllowedLegacy(from, appsv1.UpgradeOp) {
			return appsv1.UpgradeOp, true
		}
	case appsv1.ApplyingEnv:
		if opAllowedLegacy(from, appsv1.ApplyEnvOp) {
			return appsv1.ApplyEnvOp, true
		}
	case appsv1.Resuming:
		if opAllowedLegacy(from, appsv1.ResumeOp) {
			return appsv1.ResumeOp, true
		}
	case appsv1.Stopping:
		if opAllowedLegacy(from, appsv1.StopOp) {
			return appsv1.StopOp, true
		}
	}
	return "", false
}

// opAllowedLegacy replicates OperationAllowedInState[from][op] as a
// pure lookup. Kept inline (rather than importing pkg/appstate) so
// the domain package remains free of legacy dependencies. The data
// below matches pkg/appstate/state_transition.go:258-388 verbatim;
// table_test.go asserts them byte-for-byte.
func opAllowedLegacy(from State, op OpType) bool {
	table := map[State]map[OpType]bool{
		InitialState:                          {appsv1.InstallOp: true},
		appsv1.Pending:                        {appsv1.CancelOp: true},
		appsv1.Downloading:                    {appsv1.CancelOp: true},
		appsv1.UpgradePending:                 {appsv1.CancelOp: true},
		appsv1.UpgradeDownloading:             {appsv1.CancelOp: true},
		appsv1.Installing:                     {appsv1.CancelOp: true},
		appsv1.Initializing:                   {appsv1.CancelOp: true},
		appsv1.Upgrading:                      {appsv1.CancelOp: true},
		appsv1.ApplyingEnv:                    {appsv1.CancelOp: true, appsv1.StopOp: true},
		appsv1.Resuming:                       {appsv1.CancelOp: true, appsv1.StopOp: true},
		appsv1.PendingCanceled:                {appsv1.InstallOp: true},
		appsv1.DownloadingCanceled:            {appsv1.InstallOp: true},
		appsv1.InstallingCanceled:             {appsv1.InstallOp: true},
		appsv1.Uninstalled:                    {appsv1.InstallOp: true},
		appsv1.DownloadFailed:                 {appsv1.InstallOp: true},
		appsv1.InstallFailed:                  {appsv1.UninstallOp: true},
		appsv1.StopFailed:                     {appsv1.StopOp: true, appsv1.UninstallOp: true},
		appsv1.ResumeFailed:                   {appsv1.ResumeOp: true, appsv1.ApplyEnvOp: true, appsv1.UninstallOp: true},
		appsv1.UninstallFailed:                {appsv1.UninstallOp: true},
		appsv1.UpgradeFailed:                  {appsv1.UninstallOp: true, appsv1.UpgradeOp: true},
		appsv1.ApplyEnvFailed:                 {appsv1.UninstallOp: true, appsv1.ApplyEnvOp: true, appsv1.StopOp: true},
		appsv1.PendingCancelFailed:            {appsv1.CancelOp: true},
		appsv1.DownloadingCancelFailed:        {appsv1.CancelOp: true},
		appsv1.UpgradePendingCancelFailed:     {appsv1.CancelOp: true},
		appsv1.UpgradeDownloadingCancelFailed: {appsv1.CancelOp: true},
		appsv1.InstallingCancelFailed:         {appsv1.CancelOp: true, appsv1.UninstallOp: true},
		appsv1.InitializingCancelFailed:       {appsv1.CancelOp: true, appsv1.UninstallOp: true},
		appsv1.UpgradingCancelFailed:          {appsv1.CancelOp: true, appsv1.UninstallOp: true},
		appsv1.ApplyingEnvCancelFailed:        {appsv1.CancelOp: true, appsv1.UninstallOp: true},
		appsv1.ResumingCancelFailed:           {appsv1.CancelOp: true},
		appsv1.Running:                        {appsv1.UninstallOp: true, appsv1.UpgradeOp: true, appsv1.ApplyEnvOp: true, appsv1.StopOp: true},
		appsv1.Stopped:                        {appsv1.UninstallOp: true, appsv1.UpgradeOp: true, appsv1.ApplyEnvOp: true, appsv1.ResumeOp: true},
	}
	return table[from][op]
}

// LegacyOpsFor returns the operations legacy OperationAllowedInState
// records for state. Exposed for table_test.go's migration-correctness
// proof; not intended for production code.
func LegacyOpsFor(state State) map[OpType]bool {
	out := map[OpType]bool{}
	for _, op := range AllOps {
		if opAllowedLegacy(state, op) {
			out[op] = true
		}
	}
	return out
}

// isCancelingState reports whether s is one of the seven *Canceling
// mid-cancellation states.
func isCancelingState(s State) bool {
	switch s {
	case appsv1.PendingCanceling,
		appsv1.DownloadingCanceling,
		appsv1.UpgradePendingCanceling,
		appsv1.UpgradeDownloadingCanceling,
		appsv1.InstallingCanceling,
		appsv1.InitializingCanceling,
		appsv1.ResumingCanceling,
		appsv1.UpgradingCanceling,
		appsv1.ApplyingEnvCanceling:
		return true
	}
	return false
}

// isFailureTarget reports whether to represents a failure terminal
// (either an operation *Failed or a *CancelFailed).
func isFailureTarget(to State) bool {
	switch to {
	case appsv1.DownloadFailed,
		appsv1.InstallFailed,
		appsv1.StopFailed,
		appsv1.UpgradeFailed,
		appsv1.ApplyEnvFailed,
		appsv1.ResumeFailed,
		appsv1.UninstallFailed,
		appsv1.PendingCancelFailed,
		appsv1.DownloadingCancelFailed,
		appsv1.UpgradePendingCancelFailed,
		appsv1.UpgradeDownloadingCancelFailed,
		appsv1.InstallingCancelFailed,
		appsv1.InitializingCancelFailed,
		appsv1.UpgradingCancelFailed,
		appsv1.ApplyingEnvCancelFailed,
		appsv1.ResumingCancelFailed:
		return true
	}
	return false
}

// isOperatingSource reports whether from is an operating state, mirroring
// legacy OperatingStates. Duplicates data in setStateSpecs so
// actionForEdge can be evaluated independently.
func isOperatingSource(from State) bool {
	switch from {
	case appsv1.Pending,
		appsv1.Downloading,
		appsv1.UpgradePending,
		appsv1.UpgradeDownloading,
		appsv1.Installing,
		appsv1.Initializing,
		appsv1.Resuming,
		appsv1.Upgrading,
		appsv1.ApplyingEnv,
		appsv1.Stopping,
		appsv1.Uninstalling,
		appsv1.PendingCanceling,
		appsv1.DownloadingCanceling,
		appsv1.UpgradePendingCanceling,
		appsv1.UpgradeDownloadingCanceling,
		appsv1.InstallingCanceling,
		appsv1.InitializingCanceling,
		appsv1.ResumingCanceling,
		appsv1.UpgradingCanceling,
		appsv1.ApplyingEnvCanceling:
		return true
	}
	return false
}

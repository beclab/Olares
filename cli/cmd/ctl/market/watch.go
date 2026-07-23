package market

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// watchOp is the lifecycle the user-facing CLI command kicked off and that we
// expect the per-app status row to converge on. Values match the backend
// `OpType` constants in framework/app-service/api/.../appmanager_states.go so
// we can compare directly to `statusRow.OpType` without translation.
type watchOp string

const (
	watchInstall   watchOp = "install"
	watchUpgrade   watchOp = "upgrade"
	watchUninstall watchOp = "uninstall"
	watchStop      watchOp = "stop"
	watchResume    watchOp = "resume"
	watchRestart   watchOp = "restart"
	watchCancel    watchOp = "cancel"
	// watchStatus is the op-agnostic variant used by the `status --watch`
	// command: the user didn't kick off any lifecycle in this CLI
	// invocation, they just want to block until whatever lifecycle is
	// already in flight finishes. Terminal sets are therefore the union
	// of every per-op success/failure set.
	watchStatus watchOp = "status"
)

// State classification mirrors apps/packages/app/src/constant/config.ts.
// progressingStates are merely informational — we never declare them
// terminal; their main use is so a future caller could ask "is this row
// still in motion?". The terminal classification is done per-op via
// successSet / failureSet on watchTarget.
var (
	canceledStates = map[string]bool{
		"pendingCanceled":      true,
		"downloadingCanceled":  true,
		"installingCanceled":   true,
		"initializingCanceled": true,
		"upgradingCanceled":    true,
		"applyingEnvCanceled":  true,
		"resumingCanceled":     true,
	}

	cancelFailedStates = map[string]bool{
		"pendingCancelFailed":     true,
		"downloadingCancelFailed": true,
		"installingCancelFailed":  true,
		"upgradingCancelFailed":   true,
		"applyingEnvCancelFailed": true,
		"resumingCancelFailed":    true,
	}

	operationFailedStates = map[string]bool{
		"downloadFailed":  true,
		"installFailed":   true,
		"uninstallFailed": true,
		"upgradeFailed":   true,
		"stopFailed":      true,
		"resumeFailed":    true,
		"applyEnvFailed":  true,
		"failed":          true,
	}
)

// watchTarget captures everything waitForTerminal needs to decide whether a
// given statusRow means "we're done": the op we issued, who to look for, and
// the per-op success/failure state sets. Built via newWatchTarget so callers
// don't have to reproduce the per-op set lookup table.
type watchTarget struct {
	op      watchOp
	appName string
	source  string

	// matchOpType requires `row.OpType == string(op)` before we accept a
	// state as terminal. This guards against tick-zero false positives:
	// an `upgrade` issued on an already-`running` app would otherwise
	// short-circuit to success before the backend even started the
	// upgrade lifecycle. Cancel sets this to false because the canceled
	// row's OpType remains the underlying op being canceled (install /
	// upgrade / ...), not the literal string "cancel".
	matchOpType bool

	successSet map[string]bool
	failureSet map[string]bool

	// absentMeansSuccess flips the "row vanished from /market/state"
	// signal from "still pending" to "we're done". True for uninstall
	// (the backend may stop reporting the app once it's fully
	// uninstalled) and for status --watch (an in-flight uninstall the
	// user is observing). The seen-first guard (see waitForTerminal)
	// prevents a totally-unknown app name from being misreported as
	// "successfully uninstalled" — except when acceptInitialAbsent is
	// also set, which is the only safe path for uninstall (see below).
	absentMeansSuccess bool

	// acceptInitialAbsent further relaxes absentMeansSuccess: when both
	// are set, the row being absent on tick zero (before we've ever seen
	// it) is itself terminal success. This is required for uninstall
	// when the per-user row was already cleaned up before the DELETE we
	// just issued — the most common case is a multi-user CS app where
	// the user previously ran `uninstall <app>` (which cleared their
	// row) and now re-runs `uninstall <app> --cascade` to tear down the
	// shared sub-charts. The backend accepts the cascade DELETE without
	// re-creating the user's row, so the watcher otherwise hangs until
	// --watch-timeout. Status --watch deliberately does NOT enable this
	// because its caller (runStatusSingle) fetches the initial row
	// before invoking waitForTerminal — a "first-poll absent" there
	// would only mean "row just disappeared between the prefetch and
	// our first poll", which `absentMeansSuccess && seen` already
	// handles. Install / upgrade / clone / cancel never enable this
	// (their absent-at-start state is "not yet provisioned", not "done").
	acceptInitialAbsent bool

	// idempotentSuccess relaxes the OpType gate for verbs whose
	// "target state" is unambiguous about being done: `stop` on
	// already-`stopped` and `resume` on already-`running` are no-ops
	// for the backend, which means it never bumps OpType to `stop` /
	// `resume` and instead leaves the row at `state=target, OpType=""`.
	// With the strict gate that situation looks identical to a still-
	// pending op and we'd hang until --watch-timeout fires. When this
	// flag is set, "state ∈ successSet ∧ OpType == ''" is treated as
	// success: the backend is quiescent at our target state.
	//
	// install / upgrade DELIBERATELY don't get this shortcut because
	// `state=running, OpType=""` could just as well be a stale row
	// from a previous successful install of a different version with
	// the new op not yet picked up by the backend — there's no way to
	// distinguish "no-op success" from "pre-tick-zero" from the row
	// alone, so we keep the strict gate for those.
	idempotentSuccess bool

	// stoppedTerminalSuccess treats a `stopped` row as a terminal SUCCESS,
	// op-agnostically, but ONLY when its statusTime is strictly newer than
	// baselineStatusTime. This is the resume watcher's cancel escape hatch:
	// a resume that gets cancelled (by the user via `market cancel`, or by
	// the backend's TTL) walks resuming -> resumingCanceling -> stopping ->
	// stopped, so `stopped` is the settled outcome — but it is ALSO the
	// pre-resume resting state (you resume a stopped app), byte-identical at
	// tick zero. We reuse restart's statusTime-baseline trick to tell the
	// freshly-cancelled `stopped` from the stale pre-resume row (strict `>`;
	// a missing/unparseable statusTime -> effectiveTime 0 -> never terminal,
	// mirroring the SPA). Unlike requireNewerThanBaseline this is scoped to
	// the single `stopped` state, so the `running` success path (incl. the
	// idempotentSuccess no-op shortcut) is unaffected. Only watchResume sets
	// it; captured baseline comes from runResume, same as restart.
	stoppedTerminalSuccess bool

	// requireNewerThanBaseline gates BOTH success and failure on the row's
	// statusTime being strictly newer than baselineStatusTime. This is the
	// crux of the restart watcher: a completed restart rests at
	// `state=running, OpType=resume` — byte-for-byte identical to the row's
	// pre-restart resting state (OpType is never cleared). The ONLY way to
	// tell "restart just finished" from "restart hasn't started yet" is
	// temporal, and we align with the SPA which orders status rows by
	// statusTime (getEffectiveTime in apps/.../constant/constants.ts).
	//
	// We use STRICT `>` (not `>=`) on purpose: the baseline IS the stale
	// row's statusTime, so `>=` would let the tick-zero stale row satisfy
	// `T0 >= T0` and short-circuit to success — exactly the false positive
	// this whole mechanism exists to prevent.
	//
	// When the row's statusTime is missing/unparseable (effectiveTime == 0)
	// we mirror the SPA (newTime===0 → invalid, skip): the row is NOT
	// treated as terminal and we keep polling until --watch-timeout. Only
	// restart sets this; every other op leaves it false (passesBaseline is
	// then a no-op).
	requireNewerThanBaseline bool
	baselineStatusTime       int64
}

func newWatchTarget(op watchOp, appName, source string) watchTarget {
	t := watchTarget{
		op:          op,
		appName:     appName,
		source:      strings.TrimSpace(source),
		matchOpType: true,
	}
	switch op {
	case watchInstall:
		t.successSet = map[string]bool{"running": true}
		// During install, any *Failed terminates as failure; *Canceled
		// likewise — a concurrent cancel means the install we asked for
		// did not happen, so the user-facing exit code should be
		// non-zero even though the backend reached a "clean" state.
		t.failureSet = unionStateSets(operationFailedStates, canceledStates)
	case watchUpgrade:
		t.successSet = map[string]bool{"running": true}
		t.failureSet = unionStateSets(operationFailedStates, canceledStates)
	case watchUninstall:
		t.successSet = map[string]bool{"uninstalled": true}
		t.failureSet = map[string]bool{"uninstallFailed": true}
		t.absentMeansSuccess = true
		// `uninstall <app>` (and especially `uninstall <app> --cascade`
		// re-run after a previous per-user uninstall) can target a row
		// that the backend has already removed before we get the first
		// poll back. The DELETE we just issued was accepted (we wouldn't
		// be in the watch otherwise — `mc.UninstallApp` returned 200)
		// and absence from /market/state with op=uninstall semantics is
		// itself the terminal state we're looking for. See
		// TestWaitForTerminalUninstallAbsentFromStart.
		t.acceptInitialAbsent = true
	case watchStop:
		t.successSet = map[string]bool{"stopped": true}
		t.failureSet = map[string]bool{"stopFailed": true}
		// `stop` on an already-stopped row is a backend no-op: the row
		// sits at `state=stopped, OpType=""` and the strict gate would
		// otherwise hang waiting for OpType to bounce to "stop".
		t.idempotentSuccess = true
	case watchResume:
		t.successSet = map[string]bool{"running": true}
		// resumingCanceled is intentionally omitted: that transition does
		// not exist (a cancelled resume settles at `stopped`, handled by
		// stoppedTerminalSuccess below). Failure is the resume dying
		// (resumeFailed) or the cancel request itself being rejected
		// (resumingCancelFailed).
		t.failureSet = map[string]bool{
			"resumeFailed":         true,
			"resumingCancelFailed": true,
		}
		// Symmetric to stop: `resume` on an already-running row is a
		// backend no-op and would otherwise hang the watcher.
		t.idempotentSuccess = true
		// A resume that gets cancelled lands on `stopped`; accept it as a
		// terminal (baseline-gated so the pre-resume stopped row at tick
		// zero is not a false positive). See stoppedTerminalSuccess.
		t.stoppedTerminalSuccess = true
	case watchRestart:
		// restart = backend stop THEN resume once stop succeeds. The row
		// therefore passes through `OpType=stop` (stop phase) and
		// `OpType=resume` (resume phase, via `initializing`) before
		// resting at `state=running`. We must NOT reuse watchResume:
		//   - idempotentSuccess would fire on the tick-zero stale
		//     `running, OpType=resume` row (the restart target is by
		//     definition a running app), returning before the cycle even
		//     starts.
		//   - matchOpType=true / OpType=resume can't catch a `stopFailed`
		//     in the stop phase (it carries OpType=stop), so a failed stop
		//     would hang until timeout.
		//
		// Instead: op-agnostic success on `running`, a failure set that
		// spans BOTH phases, and a statusTime baseline gate that alone
		// distinguishes the freshly-completed row from the pre-restart
		// resting row (see requireNewerThanBaseline). No idempotentSuccess.
		t.successSet = map[string]bool{"running": true}
		t.failureSet = map[string]bool{
			"stopFailed":           true,
			"resumeFailed":         true,
			"resumingCanceled":     true,
			"resumingCancelFailed": true,
		}
		t.matchOpType = false
		t.requireNewerThanBaseline = true
	case watchCancel:
		// `cancel` is op-agnostic on settling: once the row leaves its
		// in-flight phase the cancel has done all it can, regardless of
		// which terminal bucket the underlying op landed in. The user
		// only cares whether the request itself was rejected (→
		// *CancelFailed) or whether the row eventually stopped moving.
		//
		// Why this set is wider than canceledStates:
		//   - `downloadFailed` / `installFailed` / `upgradeFailed` /
		//     `stopFailed` / `resumeFailed` / `applyEnvFailed` /
		//     `uninstallFailed`: the underlying op died terminally
		//     during / before cancel landed. From the user's POV the
		//     in-flight op is over (and didn't complete), so cancel
		//     "won". Reporting these as cancel-failures would hang
		//     `--watch` indefinitely on a row that will never move
		//     again — the original regression behind this expansion.
		//   - `running` / `stopped` / `uninstalled`: cancel raced and
		//     lost (op completed before the DELETE landed), OR cancel
		//     of a stop/resume reverted to the prior stable state. The
		//     row is settled either way; the OperationResult's State
		//     field tells the caller what actually happened so they
		//     can decide whether to redo the op.
		//   - canceledStates: original semantics — the dedicated
		//     *Canceled terminal states the SPA emits on cancel-wins.
		//
		// Failure stays narrow on purpose: only *CancelFailed means
		// the cancel request itself was rejected by the backend.
		// Everything else is "the row eventually settled" and is more
		// useful as a non-erroring terminal than as a fatal exit code.
		t.successSet = unionStateSets(
			map[string]bool{
				"running":     true,
				"stopped":     true,
				"uninstalled": true,
			},
			canceledStates,
			operationFailedStates,
		)
		t.failureSet = cancelFailedStates
		t.matchOpType = false
		// Mirror the uninstall watcher's "row vanished mid-watch =
		// terminal" shortcut: a cancel during install of a CS app can
		// trigger backend rollback that prunes the per-user row before
		// the row ever reports a *Canceled state. seen-first guard
		// (acceptInitialAbsent stays false) prevents a wrong app name
		// from being mis-classified as cancel-success.
		t.absentMeansSuccess = true
	case watchStatus:
		// Op-agnostic: any stable resting state counts as success
		// (running for install/upgrade/resume; stopped for stop;
		// uninstalled for uninstall; *Canceled for cancel). Any
		// *Failed or *CancelFailed state still maps to failure so
		// scripts get a non-zero exit when something actually broke.
		t.successSet = unionStateSets(
			map[string]bool{
				"running":     true,
				"stopped":     true,
				"uninstalled": true,
			},
			canceledStates,
		)
		t.failureSet = unionStateSets(operationFailedStates, cancelFailedStates)
		t.matchOpType = false
		// If the row disappears mid-watch (uninstall finishing, app
		// pruned, ...) treat that as terminal — same shortcut the
		// dedicated uninstall watcher uses, but the status flow only
		// enters waitForTerminal after confirming the row was present
		// initially, so this can't fire on a never-installed app.
		t.absentMeansSuccess = true
	}
	return t
}

// watchTimeoutError is the error returned when --watch-timeout elapses before
// the row reaches a terminal state. We surface the last seen state so the
// user can decide whether to extend the timeout or investigate.
type watchTimeoutError struct {
	target watchTarget
	last   *statusRow
}

func (e *watchTimeoutError) Error() string {
	if e.last != nil {
		return fmt.Sprintf("%s '%s' watch timed out (last state: %s, op: %s)",
			e.target.op, e.target.appName,
			valueOrUnknown(e.last.State), valueOrUnknown(e.last.OpType))
	}
	return fmt.Sprintf("%s '%s' watch timed out (no status reported by the backend)",
		e.target.op, e.target.appName)
}

// watchFailureError represents a terminal-failure classification. It exposes
// the row so callers can render a structured OperationResult.
type watchFailureError struct {
	target watchTarget
	row    statusRow
}

func (e *watchFailureError) Error() string {
	parts := []string{fmt.Sprintf("state=%s", e.row.State)}
	if e.row.OpType != "" {
		parts = append(parts, "op="+e.row.OpType)
	}
	if detail := strings.TrimSpace(e.row.Message); detail != "" {
		parts = append(parts, "reason: "+detail)
	}
	return fmt.Sprintf("%s '%s' failed: %s",
		e.target.op, e.target.appName, strings.Join(parts, " "))
}

// waitForTerminal polls /market/state until the row classifies as terminal
// (success or failure) per `t`, or until the deadline / signal.NotifyContext
// fires. The first poll happens immediately so a state that was already
// terminal at issue time (e.g. a `stop` on an already-`stopped` app once the
// backend has switched OpType) returns without a wasted sleep cycle.
func waitForTerminal(parentCtx context.Context, mc *MarketClient, opts *MarketOptions, t watchTarget) (statusRow, error) {
	interval := opts.WatchInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	timeoutDur := opts.WatchTimeout
	if timeoutDur <= 0 {
		timeoutDur = 15 * time.Minute
	}
	deadline := time.Now().Add(timeoutDur)

	// signal.NotifyContext lets us distinguish "user pressed Ctrl-C" from
	// "parent context canceled for some other reason" by checking whether
	// parentCtx is still alive when the derived ctx is done.
	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		last         *statusRow
		seen         bool
		consecErrors int
	)

	for {
		if err := ctx.Err(); err != nil {
			if parentCtx.Err() == nil {
				return statusRow{}, fmt.Errorf("%s '%s' watch canceled by user", t.op, t.appName)
			}
			return statusRow{}, err
		}
		if time.Now().After(deadline) {
			return statusRow{}, &watchTimeoutError{target: t, last: last}
		}

		resp, err := mc.GetMarketState(ctx)
		if err != nil {
			// Transient errors (auth refresh, ephemeral 5xx, network
			// blips) shouldn't kill the watch outright. We surface them
			// to stderr at most once per occurrence and bail out only
			// after several consecutive failures so the user isn't
			// trapped in a pathological retry loop.
			consecErrors++
			opts.info("watch: failed to fetch state: %v (retry %d)", err, consecErrors)
			if consecErrors >= 5 {
				return statusRow{}, fmt.Errorf("%s '%s' watch aborted after %d consecutive errors: %w", t.op, t.appName, consecErrors, err)
			}
			if waitErr := sleepOrCancel(ctx, interval); waitErr != nil {
				continue
			}
			continue
		}
		consecErrors = 0

		row, present := lookupWatchRow(resp, t.appName, t.source)
		switch {
		case !present:
			if t.absentMeansSuccess && (seen || t.acceptInitialAbsent) {
				// Two flavors of "absent → success":
				//   • seen-then-absent: we previously observed the row
				//     in a progressing state and the backend has now
				//     stopped reporting it (the canonical uninstall
				//     completion signal; also fires under
				//     status --watch).
				//   • acceptInitialAbsent (uninstall only): the row was
				//     already gone before our DELETE — typically a
				//     re-run with `--cascade` after the per-user row
				//     was cleared by a prior uninstall. The backend
				//     still processes the cascade payload but never
				//     re-creates the user's row, so without this
				//     branch the watcher hangs until --watch-timeout.
				// Synthesize a row so the caller still has
				// source/state context for output.
				return statusRow{
					Name:   t.appName,
					State:  "uninstalled",
					OpType: string(t.op),
					Source: t.source,
				}, nil
			}
			// Not yet present; keep waiting (e.g. install just submitted).
		default:
			seen = true

			// Only emit an info line when something actually changed,
			// to keep watch output proportional to real progress
			// instead of one line per poll.
			if last == nil || last.State != row.State || last.OpType != row.OpType || last.Progress != row.Progress {
				opts.info("[%s] state=%s op=%s progress=%s source=%s",
					row.Name,
					valueOrUnknown(row.State),
					valueOrUnknown(row.OpType),
					valueOrUnknown(row.Progress),
					valueOrUnknown(row.Source))
			}
			rowCopy := row
			last = &rowCopy

			// Resume-cancel escape hatch: a cancelled resume settles at
			// `stopped`, op-agnostically, but only a row strictly newer
			// than the pre-resume baseline is the fresh cancel (the tick-
			// zero pre-resume `stopped` row shares the baseline statusTime).
			if t.stoppedTerminalSuccess && row.State == appStateStopped && effectiveTime(row) > t.baselineStatusTime {
				return row, nil
			}
			if t.successSet[row.State] && t.passesBaseline(row) {
				switch {
				case t.matchesOpType(row):
					return row, nil
				case t.idempotentSuccess && row.OpType == "":
					// No-op success: the backend treated our request
					// as already-satisfied (stop on stopped / resume
					// on running) and left the row at the target
					// state with no op in flight. Without this branch
					// the watcher would hang until --watch-timeout.
					return row, nil
				}
			}
			if t.matchesOpType(row) && t.failureSet[row.State] && t.passesBaseline(row) {
				return row, &watchFailureError{target: t, row: row}
			}
		}

		if err := sleepOrCancel(ctx, interval); err != nil {
			// Loop top will reclassify the ctx error.
			continue
		}
	}
}

func (t watchTarget) matchesOpType(row statusRow) bool {
	if !t.matchOpType {
		return true
	}
	return row.OpType == string(t.op)
}

// passesBaseline enforces the statusTime baseline gate (restart only). For
// targets that don't set requireNewerThanBaseline it's a no-op. When set, the
// row's statusTime must parse AND be strictly newer than the captured
// baseline; a missing/unparseable statusTime (effectiveTime == 0) is treated
// as "no reliable signal yet" and keeps the watcher polling (mirrors the SPA's
// `newTime === 0 → invalid, skip` in appStore.setAppStatus).
func (t watchTarget) passesBaseline(row statusRow) bool {
	if !t.requireNewerThanBaseline {
		return true
	}
	et := effectiveTime(row)
	if et == 0 {
		return false
	}
	return et > t.baselineStatusTime
}

// effectiveTime mirrors the SPA's getEffectiveTime
// (apps/.../constant/constants.ts): the canonical ordering key for a status
// row is statusTime. Returns unix-millis, or 0 when statusTime is
// absent/unparseable (the SPA's "invalid" sentinel).
func effectiveTime(row statusRow) int64 {
	return parseStatusTime(row.StatusTime)
}

// parseStatusTime parses a backend statusTime (RFC3339, e.g.
// "2026-07-01T08:29:03Z"; fractional seconds accepted) into unix-millis.
// Returns 0 for empty or unparseable input.
func parseStatusTime(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0
	}
	return tm.UnixMilli()
}

func unionStateSets(sets ...map[string]bool) map[string]bool {
	total := 0
	for _, s := range sets {
		total += len(s)
	}
	out := make(map[string]bool, total)
	for _, s := range sets {
		for k := range s {
			out[k] = true
		}
	}
	return out
}

// lookupWatchRow finds the app's current row in a market-state response. It
// mirrors the source-then-fallback logic in runStatusSingle so an app
// installed under a non-default source still surfaces during watch.
func lookupWatchRow(resp *APIResponse, appName, source string) (statusRow, bool) {
	if source != "" {
		if rows, err := parseStatusRows(resp, source, false); err == nil {
			for _, r := range rows {
				if r.Name == appName {
					return r, true
				}
			}
		}
	}
	rows, err := parseStatusRows(resp, "", true)
	if err != nil {
		return statusRow{}, false
	}
	for _, r := range rows {
		if r.Name == appName {
			return r, true
		}
	}
	return statusRow{}, false
}

func sleepOrCancel(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func valueOrUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// runWithWatch is the shared post-mutation flow used by every command that
// adds --watch. When opts.Watch is false it simply prints the existing
// "accepted" result; when true it polls until terminal and folds the final
// state into the OperationResult so JSON callers get a single, fully-resolved
// record on stdout.
func runWithWatch(opts *MarketOptions, mc *MarketClient, accepted OperationResult, target watchTarget) error {
	if !opts.Watch {
		return finishOperation(opts, mc, accepted)
	}

	if !opts.Quiet && !opts.isJSON() {
		// Mirror the "request accepted" line users see without --watch
		// so the watch transitions that follow have context.
		opts.info("%s '%s' requested; watching until terminal state (timeout: %s)...",
			accepted.Operation, accepted.App, opts.WatchTimeout)
	}

	row, err := waitForTerminal(context.Background(), mc, opts, target)
	if err != nil {
		// On terminal-failure / timeout / interrupt, wrap the existing
		// accepted result with whatever we learned about the final
		// state so JSON consumers see structured data instead of a
		// bare error string.
		failed := accepted
		failed.Status = "failed"
		failed.Message = err.Error()
		var fail *watchFailureError
		if errors.As(err, &fail) {
			failed.State = fail.row.State
			failed.Progress = fail.row.Progress
			failed.FinalState = fail.row.State
			failed.FinalOpType = fail.row.OpType
		}
		var to *watchTimeoutError
		if errors.As(err, &to) && to.last != nil {
			failed.State = to.last.State
			failed.Progress = to.last.Progress
			failed.FinalState = to.last.State
			failed.FinalOpType = to.last.OpType
		}
		// Render the structured result first, then return a sentinel
		// so the cobra layer knows we've already reported.
		opts.printResult(failed)
		return errReported
	}

	final := accepted
	final.Status = "success"
	final.State = row.State
	final.Progress = row.Progress
	final.FinalState = row.State
	final.FinalOpType = row.OpType
	final.Message = fmt.Sprintf("%s completed (state=%s)", accepted.Operation, row.State)
	if row.Source != "" {
		final.Source = row.Source
	}
	if !opts.Quiet {
		opts.printResult(final)
	}
	return nil
}

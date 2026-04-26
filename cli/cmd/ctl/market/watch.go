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
	// signal from "still pending" to "we're done". Only true for
	// uninstall, where the backend may stop reporting the app once it's
	// fully uninstalled. We additionally require that we have seen the
	// row at least once during this watch so a totally-unknown app name
	// doesn't get reported as "successfully uninstalled".
	absentMeansSuccess bool
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
	case watchStop:
		t.successSet = map[string]bool{"stopped": true}
		t.failureSet = map[string]bool{"stopFailed": true}
	case watchResume:
		t.successSet = map[string]bool{"running": true}
		t.failureSet = map[string]bool{
			"resumeFailed":         true,
			"resumingCanceled":     true,
			"resumingCancelFailed": true,
		}
	case watchCancel:
		t.successSet = canceledStates
		t.failureSet = cancelFailedStates
		t.matchOpType = false
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
			if t.absentMeansSuccess && seen {
				// We previously saw the row in a progressing state and
				// the backend has now stopped reporting it: treat as
				// a successful uninstall. Synthesize a row so the
				// caller still has source/state context for output.
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

			if t.matchesOpType(row) && t.successSet[row.State] {
				return row, nil
			}
			if t.matchesOpType(row) && t.failureSet[row.State] {
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

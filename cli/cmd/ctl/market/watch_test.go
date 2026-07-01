package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// classifyForTest mirrors the classification waitForTerminal performs on each
// poll: it returns "success" / "failure" / "progressing" for a hypothetical
// row, without invoking the actual poll loop. Kept in lockstep with the
// classifier branches in waitForTerminal (state ∈ successSet ∧ passesBaseline
// ∧ matchesOpType → success; idempotentSuccess shortcut for `state ∈
// successSet ∧ OpType == ""`; matchesOpType ∧ state ∈ failureSet ∧
// passesBaseline → failure; everything else still in motion). Keeping the
// helper local avoids growing the package's exported surface just for unit
// coverage.
func classifyForTest(t watchTarget, row statusRow) string {
	if t.successSet[row.State] && t.passesBaseline(row) {
		if t.matchesOpType(row) {
			return "success"
		}
		if t.idempotentSuccess && row.OpType == "" {
			return "success"
		}
	}
	if t.matchesOpType(row) && t.failureSet[row.State] && t.passesBaseline(row) {
		return "failure"
	}
	return "progressing"
}

func TestClassifierInstallLifecycle(t *testing.T) {
	target := newWatchTarget(watchInstall, "myapp", "market.olares")

	cases := []struct {
		name   string
		row    statusRow
		expect string
	}{
		{"pending", statusRow{State: "pending", OpType: "install"}, "progressing"},
		{"downloading", statusRow{State: "downloading", OpType: "install"}, "progressing"},
		{"installing", statusRow{State: "installing", OpType: "install"}, "progressing"},
		{"initializing", statusRow{State: "initializing", OpType: "install"}, "progressing"},
		{"running with install op", statusRow{State: "running", OpType: "install"}, "success"},
		{"installFailed", statusRow{State: "installFailed", OpType: "install"}, "failure"},
		{"downloadFailed", statusRow{State: "downloadFailed", OpType: "install"}, "failure"},
		// Cancel during install is also terminal-failure for the install
		// CTA: from the user's perspective the install they asked for did
		// not happen.
		{"installingCanceled", statusRow{State: "installingCanceled", OpType: "install"}, "failure"},
		// Stale OpType from a prior lifecycle must not prematurely
		// classify any state.
		{"running with stale upgrade op", statusRow{State: "running", OpType: "upgrade"}, "progressing"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.expect {
				t.Fatalf("classify(%+v) = %s, want %s", c.row, got, c.expect)
			}
		})
	}
}

func TestClassifierUpgradeWaitsForOpTypeFlip(t *testing.T) {
	// Issuing `upgrade` on a row currently `running, op=install` must NOT
	// short-circuit to success on tick zero; only after the backend
	// flips OpType to `upgrade` (plus reaches `running` again) is the
	// upgrade complete.
	target := newWatchTarget(watchUpgrade, "myapp", "market.olares")

	stale := statusRow{State: "running", OpType: "install"}
	if got := classifyForTest(target, stale); got != "progressing" {
		t.Fatalf("stale install OpType should keep progressing, got %s", got)
	}

	mid := statusRow{State: "upgrading", OpType: "upgrade"}
	if got := classifyForTest(target, mid); got != "progressing" {
		t.Fatalf("upgrading should be progressing, got %s", got)
	}

	done := statusRow{State: "running", OpType: "upgrade"}
	if got := classifyForTest(target, done); got != "success" {
		t.Fatalf("running with upgrade op should be success, got %s", got)
	}

	failed := statusRow{State: "upgradeFailed", OpType: "upgrade"}
	if got := classifyForTest(target, failed); got != "failure" {
		t.Fatalf("upgradeFailed should be failure, got %s", got)
	}
}

func TestClassifierUninstall(t *testing.T) {
	target := newWatchTarget(watchUninstall, "myapp", "market.olares")
	if !target.absentMeansSuccess {
		t.Fatalf("uninstall target must set absentMeansSuccess")
	}
	if got := classifyForTest(target, statusRow{State: "uninstalling", OpType: "uninstall"}); got != "progressing" {
		t.Fatalf("uninstalling should be progressing, got %s", got)
	}
	if got := classifyForTest(target, statusRow{State: "uninstalled", OpType: "uninstall"}); got != "success" {
		t.Fatalf("uninstalled should be success, got %s", got)
	}
	if got := classifyForTest(target, statusRow{State: "uninstallFailed", OpType: "uninstall"}); got != "failure" {
		t.Fatalf("uninstallFailed should be failure, got %s", got)
	}
}

func TestClassifierCancelIgnoresOpType(t *testing.T) {
	// cancel's terminal row keeps the *underlying* op (install /
	// upgrade / ...), so matchOpType must be false; otherwise we'd never
	// classify the canceled state as terminal.
	target := newWatchTarget(watchCancel, "myapp", "")
	if target.matchOpType {
		t.Fatalf("cancel target must NOT require OpType match")
	}
	if !target.absentMeansSuccess {
		t.Fatalf("cancel target must opt into absentMeansSuccess so a row vanishing mid-watch is terminal")
	}

	row := statusRow{State: "installingCanceled", OpType: "install"}
	if got := classifyForTest(target, row); got != "success" {
		t.Fatalf("installingCanceled under cancel target should be success, got %s", got)
	}

	failed := statusRow{State: "installingCancelFailed", OpType: "install"}
	if got := classifyForTest(target, failed); got != "failure" {
		t.Fatalf("installingCancelFailed should be failure, got %s", got)
	}
}

// TestClassifierCancelBroadTerminalSet pins the post-cancel terminal
// set against the original regression: a real user-reported case where
// `market cancel firefox --watch` hung after the row settled at
// `downloadFailed` (download died before / during the cancel) or
// `stopped` (rollback brought the partial install down to a stable
// resting state). Before this expansion, those rows fell out of both
// `canceledStates` and `cancelFailedStates` and the watcher polled
// until --watch-timeout.
//
// Mental model: once the cancel request has been accepted, the
// watcher is really asking "did the row stop moving?", same shape as
// `status --watch`. The only state that should *fail* the cancel
// watch is `*CancelFailed` — that's the backend rejecting the cancel
// request itself.
func TestClassifierCancelBroadTerminalSet(t *testing.T) {
	target := newWatchTarget(watchCancel, "myapp", "")

	type tc struct {
		name string
		row  statusRow
		want string
	}

	successes := []tc{
		// The dedicated *Canceled set — original semantics.
		{"pendingCanceled", statusRow{State: "pendingCanceled", OpType: "install"}, "success"},
		{"downloadingCanceled", statusRow{State: "downloadingCanceled", OpType: "install"}, "success"},
		{"installingCanceled", statusRow{State: "installingCanceled", OpType: "install"}, "success"},
		{"upgradingCanceled", statusRow{State: "upgradingCanceled", OpType: "upgrade"}, "success"},
		{"applyingEnvCanceled", statusRow{State: "applyingEnvCanceled", OpType: "applyEnv"}, "success"},
		{"resumingCanceled", statusRow{State: "resumingCanceled", OpType: "resume"}, "success"},

		// The new expansion: underlying op died terminally before /
		// during cancel — cancel won by default, must not hang.
		{"downloadFailed", statusRow{State: "downloadFailed", OpType: "install"}, "success"},
		{"installFailed", statusRow{State: "installFailed", OpType: "install"}, "success"},
		{"upgradeFailed", statusRow{State: "upgradeFailed", OpType: "upgrade"}, "success"},
		{"stopFailed", statusRow{State: "stopFailed", OpType: "stop"}, "success"},
		{"resumeFailed", statusRow{State: "resumeFailed", OpType: "resume"}, "success"},
		{"uninstallFailed", statusRow{State: "uninstallFailed", OpType: "uninstall"}, "success"},
		{"applyEnvFailed", statusRow{State: "applyEnvFailed", OpType: "applyEnv"}, "success"},

		// Stable resting states — cancel raced and lost, OR cancel of
		// a stop/resume reverted to prior stable state. Either way
		// row is settled; OperationResult.State surfaces the actual
		// landing.
		{"running (cancel was too late)", statusRow{State: "running", OpType: "install"}, "success"},
		{"stopped (rollback to stable)", statusRow{State: "stopped", OpType: "stop"}, "success"},
		{"uninstalled (rollback completed)", statusRow{State: "uninstalled", OpType: "uninstall"}, "success"},
	}

	for _, c := range successes {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.want {
				t.Fatalf("classify(%+v) on cancel target = %s, want %s", c.row, got, c.want)
			}
		})
	}

	// Failure branch: ONLY *CancelFailed — the cancel request was
	// rejected by the backend. Everything else (including all of the
	// successes above and any in-flight state) must NOT classify as
	// failure.
	failures := []tc{
		{"pendingCancelFailed", statusRow{State: "pendingCancelFailed", OpType: "install"}, "failure"},
		{"downloadingCancelFailed", statusRow{State: "downloadingCancelFailed", OpType: "install"}, "failure"},
		{"installingCancelFailed", statusRow{State: "installingCancelFailed", OpType: "install"}, "failure"},
		{"upgradingCancelFailed", statusRow{State: "upgradingCancelFailed", OpType: "upgrade"}, "failure"},
		{"applyingEnvCancelFailed", statusRow{State: "applyingEnvCancelFailed", OpType: "applyEnv"}, "failure"},
		{"resumingCancelFailed", statusRow{State: "resumingCancelFailed", OpType: "resume"}, "failure"},
	}
	for _, c := range failures {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.want {
				t.Fatalf("classify(%+v) on cancel target = %s, want %s", c.row, got, c.want)
			}
		})
	}

	// In-flight states must KEEP polling — cancel hasn't reached
	// terminal yet, watcher should wait another tick.
	progressing := []tc{
		{"downloading mid-cancel", statusRow{State: "downloading", OpType: "install"}, "progressing"},
		{"installing mid-cancel", statusRow{State: "installing", OpType: "install"}, "progressing"},
		{"installingCanceling", statusRow{State: "installingCanceling", OpType: "install"}, "progressing"},
		{"upgrading mid-cancel", statusRow{State: "upgrading", OpType: "upgrade"}, "progressing"},
	}
	for _, c := range progressing {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.want {
				t.Fatalf("classify(%+v) on cancel target = %s, want %s", c.row, got, c.want)
			}
		})
	}
}

func TestClassifierStatusOpAgnostic(t *testing.T) {
	target := newWatchTarget(watchStatus, "myapp", "market.olares")
	if target.matchOpType {
		t.Fatalf("status target must not require OpType match")
	}
	if !target.absentMeansSuccess {
		t.Fatalf("status target must opt into absentMeansSuccess so a row vanishing mid-watch is terminal")
	}

	cases := []struct {
		name   string
		row    statusRow
		expect string
	}{
		// Stable resting states for any lifecycle → success.
		{"running after install", statusRow{State: "running", OpType: "install"}, "success"},
		{"running after upgrade", statusRow{State: "running", OpType: "upgrade"}, "success"},
		{"running after resume", statusRow{State: "running", OpType: "resume"}, "success"},
		{"stopped", statusRow{State: "stopped", OpType: "stop"}, "success"},
		{"uninstalled", statusRow{State: "uninstalled", OpType: "uninstall"}, "success"},
		{"installingCanceled", statusRow{State: "installingCanceled", OpType: "install"}, "success"},
		// Any in-flight state keeps polling.
		{"pending", statusRow{State: "pending", OpType: "install"}, "progressing"},
		{"installing", statusRow{State: "installing", OpType: "install"}, "progressing"},
		{"upgrading", statusRow{State: "upgrading", OpType: "upgrade"}, "progressing"},
		{"stopping", statusRow{State: "stopping", OpType: "stop"}, "progressing"},
		// All declared failure states → failure regardless of OpType.
		{"installFailed", statusRow{State: "installFailed", OpType: "install"}, "failure"},
		{"upgradeFailed", statusRow{State: "upgradeFailed", OpType: "upgrade"}, "failure"},
		{"stopFailed", statusRow{State: "stopFailed", OpType: "stop"}, "failure"},
		{"installingCancelFailed", statusRow{State: "installingCancelFailed", OpType: "install"}, "failure"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.expect {
				t.Fatalf("classify(%+v) = %s, want %s", c.row, got, c.expect)
			}
		})
	}
}

func TestWaitForTerminalStatusReachesRunning(t *testing.T) {
	// status --watch is fired against an app that's mid-install; the
	// watcher must recognize `running` as terminal even though no
	// specific op was specified by the caller.
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "initializing", OpType: "install"},
		{State: "running", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStatus, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected status watch success, got %v", err)
	}
	if row.State != "running" {
		t.Fatalf("expected terminal running, got %s", row.State)
	}
}

func TestWaitForTerminalStatusSurfacesFailure(t *testing.T) {
	seq := []statusRow{
		{State: "downloading", OpType: "install"},
		{State: "installFailed", OpType: "install", Message: "image pull error"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStatus, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected installFailed surfaced as failure")
	}
	var fail *watchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected watchFailureError, got %T: %v", err, err)
	}
	if fail.row.State != "installFailed" {
		t.Fatalf("expected installFailed, got %s", fail.row.State)
	}
}

func TestClassifierStopResume(t *testing.T) {
	stopT := newWatchTarget(watchStop, "myapp", "")
	if got := classifyForTest(stopT, statusRow{State: "stopping", OpType: "stop"}); got != "progressing" {
		t.Fatalf("stopping should be progressing, got %s", got)
	}
	if got := classifyForTest(stopT, statusRow{State: "stopped", OpType: "stop"}); got != "success" {
		t.Fatalf("stopped should be success, got %s", got)
	}
	if got := classifyForTest(stopT, statusRow{State: "stopFailed", OpType: "stop"}); got != "failure" {
		t.Fatalf("stopFailed should be failure, got %s", got)
	}

	resumeT := newWatchTarget(watchResume, "myapp", "")
	if got := classifyForTest(resumeT, statusRow{State: "running", OpType: "resume"}); got != "success" {
		t.Fatalf("running under resume target should be success, got %s", got)
	}
	if got := classifyForTest(resumeT, statusRow{State: "resumeFailed", OpType: "resume"}); got != "failure" {
		t.Fatalf("resumeFailed should be failure, got %s", got)
	}
}

func TestClassifierStopAlreadyStopped(t *testing.T) {
	// Reproduces the hang reported on `market stop firefox --watch` when
	// the app is already stopped: backend reports `{state=stopped, op=""}`
	// because it treats our request as a no-op and never bumps OpType to
	// "stop". Without the idempotentSuccess shortcut the watcher would
	// classify this as "progressing" indefinitely. With the shortcut the
	// classifier must accept it as success at tick zero.
	stopT := newWatchTarget(watchStop, "myapp", "")

	if got := classifyForTest(stopT, statusRow{State: "stopped", OpType: ""}); got != "success" {
		t.Fatalf("stop on already-stopped (state=stopped, op='') should be success, got %s", got)
	}
	// A row that is NOT yet at the success state must still progress
	// even when OpType is empty — empty-op alone is not success.
	if got := classifyForTest(stopT, statusRow{State: "running", OpType: ""}); got != "progressing" {
		t.Fatalf("stop with running/empty op should be progressing, got %s", got)
	}
	// Failure states still require the strict OpType gate: a stale
	// `stopFailed` row from some previous lifecycle with op cleared
	// must not be misclassified as a fresh failure of our request.
	if got := classifyForTest(stopT, statusRow{State: "stopFailed", OpType: ""}); got != "progressing" {
		t.Fatalf("stopFailed with empty op should stay progressing (no idempotent shortcut for failure), got %s", got)
	}
}

func TestClassifierResumeAlreadyRunning(t *testing.T) {
	// Symmetric to TestClassifierStopAlreadyStopped: `resume` on an
	// already-running row is a backend no-op; the watcher must short-
	// circuit instead of hanging until --watch-timeout.
	resumeT := newWatchTarget(watchResume, "myapp", "")

	if got := classifyForTest(resumeT, statusRow{State: "running", OpType: ""}); got != "success" {
		t.Fatalf("resume on already-running (state=running, op='') should be success, got %s", got)
	}
	if got := classifyForTest(resumeT, statusRow{State: "stopped", OpType: ""}); got != "progressing" {
		t.Fatalf("resume with stopped/empty op should be progressing, got %s", got)
	}
}

func TestClassifierInstallNoIdempotentShortcut(t *testing.T) {
	// Regression guard: install must NOT borrow the idempotent shortcut.
	// `{state=running, op=""}` is ambiguous — could be a stale row from
	// a prior install of a different version with the new install op
	// not yet picked up. Only `state=running ∧ op=install` may resolve
	// to success.
	target := newWatchTarget(watchInstall, "myapp", "market.olares")
	if got := classifyForTest(target, statusRow{State: "running", OpType: ""}); got != "progressing" {
		t.Fatalf("install must keep strict OpType gate; got %s for empty op", got)
	}
	// Sanity: upgrade behaves the same way.
	upT := newWatchTarget(watchUpgrade, "myapp", "market.olares")
	if got := classifyForTest(upT, statusRow{State: "running", OpType: ""}); got != "progressing" {
		t.Fatalf("upgrade must keep strict OpType gate; got %s for empty op", got)
	}
}

// restartTarget builds a watchRestart target with the given baseline
// statusTime (RFC3339). Mirrors what runRestart does: create the target then
// stamp the pre-restart baseline onto it.
func restartTarget(baseline string) watchTarget {
	tgt := newWatchTarget(watchRestart, "myapp", "market.olares")
	tgt.baselineStatusTime = parseStatusTime(baseline)
	return tgt
}

func TestClassifierRestartBaselineGate(t *testing.T) {
	const baseline = "2026-07-01T08:28:00Z" // pre-restart resting statusTime
	tgt := restartTarget(baseline)

	// tick-zero stale resting row: state=running, op=resume, SAME statusTime
	// as the baseline. This is the exact false-positive the old watchResume
	// reuse produced — must NOT be success (strict `>` baseline).
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "resume", StatusTime: baseline}); got != "progressing" {
		t.Fatalf("stale running row at baseline statusTime must be progressing, got %s", got)
	}
	// An OLDER statusTime than baseline is likewise not terminal.
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "resume", StatusTime: "2026-07-01T08:27:00Z"}); got != "progressing" {
		t.Fatalf("running row older than baseline must be progressing, got %s", got)
	}
	// A strictly-newer running row = the restart actually completed.
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "resume", StatusTime: "2026-07-01T08:29:03Z"}); got != "success" {
		t.Fatalf("running row newer than baseline must be success, got %s", got)
	}
	// Success is op-agnostic: even if the backend leaves op empty, a
	// strictly-newer running row is terminal success.
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "", StatusTime: "2026-07-01T08:29:03Z"}); got != "success" {
		t.Fatalf("newer running row with empty op must still be success (op-agnostic), got %s", got)
	}
}

func TestClassifierRestartMissingStatusTime(t *testing.T) {
	// D2: when statusTime is missing/unparseable (effectiveTime == 0) we
	// mirror the SPA (newTime===0 → invalid) and keep polling rather than
	// declaring terminal. A missing statusTime on a running row must NOT be
	// success, and a missing statusTime on a failure state must NOT be
	// failure.
	tgt := restartTarget("2026-07-01T08:28:00Z")
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "resume", StatusTime: ""}); got != "progressing" {
		t.Fatalf("running row with missing statusTime must be progressing, got %s", got)
	}
	if got := classifyForTest(tgt, statusRow{State: "stopFailed", OpType: "stop", StatusTime: "not-a-timestamp"}); got != "progressing" {
		t.Fatalf("failure row with unparseable statusTime must be progressing, got %s", got)
	}
}

func TestClassifierRestartFailureSpansBothPhases(t *testing.T) {
	// restart = stop THEN resume; a failure in EITHER phase is terminal.
	// The stop phase carries OpType=stop, the resume phase OpType=resume —
	// success/failure detection is op-agnostic (matchOpType=false) so both
	// are caught. Each must also clear the baseline gate.
	newer := "2026-07-01T08:29:00Z"
	tgt := restartTarget("2026-07-01T08:28:00Z")

	for _, c := range []struct {
		state, op string
	}{
		{"stopFailed", "stop"},     // stop phase died
		{"resumeFailed", "resume"}, // resume phase died
		{"resumingCanceled", "resume"},
		{"resumingCancelFailed", "resume"},
	} {
		if got := classifyForTest(tgt, statusRow{State: c.state, OpType: c.op, StatusTime: newer}); got != "failure" {
			t.Fatalf("%s (op=%s) newer than baseline must be failure, got %s", c.state, c.op, got)
		}
	}

	// Intermediate progressing states never terminate.
	for _, c := range []struct {
		state, op string
	}{
		{"stopping", "stop"},
		{"stopped", "stop"},
		{"initializing", "resume"},
		{"resuming", "resume"},
	} {
		if got := classifyForTest(tgt, statusRow{State: c.state, OpType: c.op, StatusTime: newer}); got != "progressing" {
			t.Fatalf("%s (op=%s) must be progressing, got %s", c.state, c.op, got)
		}
	}
}

func TestClassifierRestartNoIdempotentShortcut(t *testing.T) {
	// Regression guard for the original bug: restart must NOT inherit the
	// idempotentSuccess shortcut resume/stop use. Even with an empty op, a
	// running row at (or before) the baseline statusTime is NOT success.
	tgt := restartTarget("2026-07-01T08:28:00Z")
	if tgt.idempotentSuccess {
		t.Fatalf("watchRestart must not enable idempotentSuccess")
	}
	if !tgt.requireNewerThanBaseline {
		t.Fatalf("watchRestart must enable requireNewerThanBaseline")
	}
	if got := classifyForTest(tgt, statusRow{State: "running", OpType: "", StatusTime: "2026-07-01T08:28:00Z"}); got != "progressing" {
		t.Fatalf("running/empty-op at baseline must be progressing, got %s", got)
	}
}

func TestWaitForTerminalRestartSuccess(t *testing.T) {
	// Full stop→resume cycle: baseline captured before POST, then the row
	// walks op=stop → op=resume/initializing → op=resume/running with
	// strictly-increasing statusTime. Only the final running row (newer than
	// baseline) resolves to success.
	baseline := "2026-07-01T08:28:00Z"
	seq := []statusRow{
		// Tick-zero could still show the stale resting row (same statusTime
		// as baseline) before the backend picks up the restart — must be
		// skipped, not treated as success.
		{State: "running", OpType: "resume", StatusTime: baseline},
		{State: "stopping", OpType: "stop", StatusTime: "2026-07-01T08:28:30Z"},
		{State: "stopped", OpType: "stop", StatusTime: "2026-07-01T08:28:39Z"},
		{State: "initializing", OpType: "resume", StatusTime: "2026-07-01T08:29:02Z"},
		{State: "running", OpType: "resume", StatusTime: "2026-07-01T08:29:03Z"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	tgt := newWatchTarget(watchRestart, "myapp", "market.olares")
	tgt.baselineStatusTime = parseStatusTime(baseline)

	row, err := waitForTerminal(context.Background(), mc, opts, tgt)
	if err != nil {
		t.Fatalf("expected restart success, got error: %v", err)
	}
	if row.State != "running" || row.StatusTime != "2026-07-01T08:29:03Z" {
		t.Fatalf("expected terminal running row at 08:29:03, got state=%s statusTime=%s", row.State, row.StatusTime)
	}
}

func TestWaitForTerminalRestartFastComplete(t *testing.T) {
	// The scenario the "observe a transition" approach could NOT handle: the
	// whole cycle finishes before/at our first poll, so we only ever see the
	// final running row — never an intermediate state. The statusTime
	// baseline lets us accept it immediately (newer than baseline) instead of
	// hanging until --watch-timeout.
	baseline := "2026-07-01T08:28:00Z"
	seq := []statusRow{
		{State: "running", OpType: "resume", StatusTime: "2026-07-01T08:29:03Z"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(200*time.Millisecond, 5*time.Millisecond)

	tgt := newWatchTarget(watchRestart, "myapp", "market.olares")
	tgt.baselineStatusTime = parseStatusTime(baseline)

	row, err := waitForTerminal(context.Background(), mc, opts, tgt)
	if err != nil {
		t.Fatalf("expected immediate restart success, got error: %v", err)
	}
	if row.State != "running" {
		t.Fatalf("expected running, got %s", row.State)
	}
}

func TestWaitForTerminalRestartStopPhaseFailure(t *testing.T) {
	// stop phase dies with stopFailed (OpType=stop). The old watchResume
	// reuse would never catch this (its OpType gate required "resume") and
	// hang until timeout. watchRestart's op-agnostic failure set catches it.
	baseline := "2026-07-01T08:28:00Z"
	seq := []statusRow{
		{State: "stopping", OpType: "stop", StatusTime: "2026-07-01T08:28:30Z"},
		{State: "stopFailed", OpType: "stop", StatusTime: "2026-07-01T08:28:40Z", Message: "failed to stop"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	tgt := newWatchTarget(watchRestart, "myapp", "market.olares")
	tgt.baselineStatusTime = parseStatusTime(baseline)

	_, err := waitForTerminal(context.Background(), mc, opts, tgt)
	if err == nil {
		t.Fatalf("expected stop-phase failure, got nil")
	}
	var fail *watchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected watchFailureError, got %T: %v", err, err)
	}
	if fail.row.State != "stopFailed" {
		t.Fatalf("expected stopFailed in error row, got %s", fail.row.State)
	}
}

func TestWaitForTerminalRestartStaleRowTimesOut(t *testing.T) {
	// If the backend NEVER advances past the stale resting row (statusTime
	// stuck at the baseline — e.g. the restart POST silently no-op'd), the
	// watcher must NOT report a false success; it should time out instead.
	baseline := "2026-07-01T08:28:00Z"
	seq := []statusRow{
		{State: "running", OpType: "resume", StatusTime: baseline},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(120*time.Millisecond, 10*time.Millisecond)

	tgt := newWatchTarget(watchRestart, "myapp", "market.olares")
	tgt.baselineStatusTime = parseStatusTime(baseline)

	_, err := waitForTerminal(context.Background(), mc, opts, tgt)
	if err == nil {
		t.Fatalf("expected timeout on stale row, got nil (false success)")
	}
	var to *watchTimeoutError
	if !errors.As(err, &to) {
		t.Fatalf("expected watchTimeoutError, got %T: %v", err, err)
	}
}

// fakeStateServer serves /app-store/api/v2/market/state with a configurable
// queue of states so we can drive waitForTerminal end-to-end without a real
// cluster. It models exactly the response shape parseStatusRows expects.
type fakeStateServer struct {
	mu       sync.Mutex
	idx      int32
	app      string
	source   string
	sequence []statusRow
	missing  bool
	srv      *httptest.Server
}

func newFakeStateServer(t *testing.T, app, source string, seq []statusRow) *fakeStateServer {
	t.Helper()
	f := &fakeStateServer{app: app, source: source, sequence: seq}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/market/state") {
			http.NotFound(w, r)
			return
		}
		row := f.next()
		body := f.envelope(row)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeStateServer) next() (row statusRow) {
	i := atomic.AddInt32(&f.idx, 1) - 1
	if i >= int32(len(f.sequence)) {
		// Stay on the last state forever once the queue is exhausted —
		// makes timeout assertions deterministic.
		return f.sequence[len(f.sequence)-1]
	}
	return f.sequence[i]
}

func (f *fakeStateServer) markMissing() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.missing = true
}

func (f *fakeStateServer) envelope(row statusRow) []byte {
	f.mu.Lock()
	missing := f.missing
	f.mu.Unlock()

	apps := []map[string]interface{}{}
	if !missing {
		apps = append(apps, map[string]interface{}{
			"status": map[string]interface{}{
				"name":       f.app,
				"state":      row.State,
				"opType":     row.OpType,
				"progress":   row.Progress,
				"message":    row.Message,
				"statusTime": row.StatusTime,
			},
		})
	}
	// Mirror the real /market/state shape parseStatusRows expects:
	// the v2 envelope unmarshals resp.Data into MarketStateResponse,
	// whose `user_data.sources[<source>].app_state_latest[].status`
	// path holds the per-app records.
	envelope := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_data": map[string]interface{}{
				"sources": map[string]interface{}{
					f.source: map[string]interface{}{
						"type":             "market",
						"app_state_latest": apps,
					},
				},
			},
		},
	}
	b, _ := json.Marshal(envelope)
	return b
}

func newTestMarketClient(t *testing.T, baseURL string) *MarketClient {
	t.Helper()
	rp := &credential.ResolvedProfile{
		Name:        "test",
		OlaresID:    "tester@olares.test",
		AccessToken: "test-token",
		MarketURL:   baseURL,
	}
	return NewMarketClient(http.DefaultClient, http.DefaultClient, rp, "market.olares")
}

// drain swallows any output runWithWatch / waitForTerminal would emit so
// `go test` output isn't polluted; we still inspect the OperationResult /
// error returned by the API.
func quietOpts(timeout, interval time.Duration) *MarketOptions {
	return &MarketOptions{
		Source:        "market.olares",
		Output:        "json", // suppresses opts.info
		Quiet:         true,
		Watch:         true,
		WatchTimeout:  timeout,
		WatchInterval: interval,
	}
}

func TestWaitForTerminalInstallSuccess(t *testing.T) {
	seq := []statusRow{
		{State: "pending", OpType: "install"},
		{State: "downloading", OpType: "install"},
		{State: "installing", OpType: "install"},
		{State: "running", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if row.State != "running" {
		t.Fatalf("expected terminal state running, got %s", row.State)
	}
}

func TestWaitForTerminalInstallFailure(t *testing.T) {
	seq := []statusRow{
		{State: "pending", OpType: "install"},
		{State: "installFailed", OpType: "install", Message: "image pull error"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
	var fail *watchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected watchFailureError, got %T: %v", err, err)
	}
	if fail.row.State != "installFailed" {
		t.Fatalf("expected installFailed in error row, got %s", fail.row.State)
	}
	if !strings.Contains(err.Error(), "installFailed") {
		t.Fatalf("error message should mention installFailed, got %q", err.Error())
	}
}

func TestWaitForTerminalUninstallAbsent(t *testing.T) {
	seq := []statusRow{
		{State: "uninstalling", OpType: "uninstall"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	// After the first poll, drop the row entirely -> simulates the
	// backend having finished cleanup.
	go func() {
		time.Sleep(20 * time.Millisecond)
		srv.markMissing()
	}()

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchUninstall, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success on row absence, got %v", err)
	}
	if row.State != "uninstalled" {
		t.Fatalf("expected synthesized uninstalled row, got %s", row.State)
	}
}

func TestWaitForTerminalUninstallAbsentFromStart(t *testing.T) {
	// Reproduces the reported hang:
	//
	//   1) `market uninstall ollamav2 --watch` (no --cascade) on a
	//      multi-user CS app clears the user's per-user row.
	//   2) `market uninstall ollamav2 --cascade --watch` re-runs to tear
	//      down the shared sub-charts. The backend accepts the DELETE,
	//      but the user's row was already gone before step 2 so it
	//      never re-appears in /market/state.
	//
	// The fake server here mirrors step 2: the row is missing from the
	// very first poll. Without acceptInitialAbsent the watcher hangs
	// until --watch-timeout fires (~15m by default); with the flag
	// uninstall must classify "first-poll absent" as terminal success
	// — DELETE was accepted upstream and there is nothing left in the
	// user-scoped state to report.
	seq := []statusRow{
		// Placeholder row; the server immediately flips to "missing"
		// before we ever serve it, so its content doesn't matter.
		{State: "ignored", OpType: "ignored"},
	}
	srv := newFakeStateServer(t, "ollamav2", "market.olares", seq)
	srv.markMissing()
	mc := newTestMarketClient(t, srv.srv.URL)
	// Tight timeout: if acceptInitialAbsent isn't wired in, this test
	// fails in ~80ms instead of waiting for the real 15m default.
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchUninstall, "ollamav2", "market.olares"))
	if err != nil {
		t.Fatalf("expected success when row is absent from tick zero, got %v", err)
	}
	if row.State != "uninstalled" {
		t.Fatalf("expected synthesized uninstalled row, got %s", row.State)
	}
	if row.Source != "market.olares" {
		t.Fatalf("expected synthesized row to preserve source, got %q", row.Source)
	}
}

func TestWaitForTerminalStatusDoesNotShortCircuitOnInitialAbsent(t *testing.T) {
	// Regression guard for the symmetric concern: status --watch must
	// NOT borrow uninstall's tick-zero shortcut, otherwise an
	// op-agnostic "what's happening with this app" query against a row
	// that's still being submitted (e.g. user ran `install` without
	// --watch, then immediately `status --watch` while the backend is
	// still creating the row) would falsely resolve to "uninstalled".
	// runStatusSingle in production prevents this from ever firing by
	// fetching the initial row before invoking waitForTerminal, but
	// the classifier itself must independently refuse the shortcut to
	// keep that invariant intact even if a future caller wires status
	// in differently.
	seq := []statusRow{{State: "ignored", OpType: "ignored"}}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	srv.markMissing()
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStatus, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected timeout (status must not short-circuit on initial absent), got success")
	}
	var to *watchTimeoutError
	if !errors.As(err, &to) {
		t.Fatalf("expected watchTimeoutError, got %T: %v", err, err)
	}
}

func TestWaitForTerminalUpgradeWaitsForOpTypeFlip(t *testing.T) {
	// Tick 0 sees the legacy `running, op=install` row from the previous
	// install; only after the backend flips to op=upgrade and reaches
	// running again should we declare success.
	seq := []statusRow{
		{State: "running", OpType: "install"},
		{State: "upgrading", OpType: "upgrade"},
		{State: "running", OpType: "upgrade"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchUpgrade, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success once upgrade lifecycle completes, got %v", err)
	}
	if row.OpType != "upgrade" || row.State != "running" {
		t.Fatalf("expected running/upgrade, got %s/%s", row.State, row.OpType)
	}
}

func TestWaitForTerminalStopOnAlreadyStopped(t *testing.T) {
	// End-to-end reproduction of the reported hang:
	//   `olares-cli market stop firefox --watch`
	// on an app already in `stopped` state. The backend never bumps
	// OpType to "stop" because the request is a no-op, so every poll
	// returns `{state=stopped, op=""}`. Before the idempotentSuccess
	// shortcut the watcher would loop until --watch-timeout fired
	// (~15m default) and then fail with a timeout error.
	seq := []statusRow{
		{State: "stopped", OpType: ""},
		{State: "stopped", OpType: ""},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	// Aggressive timeout: if the shortcut isn't wired in, the test
	// fails in ~80ms instead of waiting for the real 15m default.
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStop, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("stop on already-stopped should succeed immediately, got %v", err)
	}
	if row.State != "stopped" {
		t.Fatalf("expected terminal stopped row, got %+v", row)
	}
	if row.OpType != "" {
		t.Fatalf("expected empty OpType on idempotent success, got %q", row.OpType)
	}
}

func TestWaitForTerminalResumeOnAlreadyRunning(t *testing.T) {
	// Symmetric to TestWaitForTerminalStopOnAlreadyStopped — the
	// idempotent shortcut must let `resume` on an already-running
	// app exit on tick zero instead of waiting for OpType to flip
	// (which the backend never bothers to do for a no-op).
	seq := []statusRow{
		{State: "running", OpType: ""},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchResume, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("resume on already-running should succeed immediately, got %v", err)
	}
	if row.State != "running" || row.OpType != "" {
		t.Fatalf("expected running/empty-op terminal row, got %+v", row)
	}
}

func TestWaitForTerminalCancelLifecycle(t *testing.T) {
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "installingCanceling", OpType: "install"},
		{State: "installingCanceled", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchCancel, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected cancel success, got %v", err)
	}
	if row.State != "installingCanceled" {
		t.Fatalf("expected installingCanceled, got %s", row.State)
	}
}

// TestWaitForTerminalCancelLandsOnDownloadFailed reproduces the
// reported regression: `market cancel firefox --watch` issued while
// firefox was downloading, but the download had already terminally
// failed by the time the cancel reached the backend. The row settles
// at `downloadFailed` and never transitions to `downloadingCanceled`.
// Before the cancel-success-set expansion this watcher hung until
// --watch-timeout fired; after, the cancel watch must terminate
// successfully and surface the actual landed state in the row.
func TestWaitForTerminalCancelLandsOnDownloadFailed(t *testing.T) {
	seq := []statusRow{
		{State: "downloading", OpType: "install"},
		{State: "downloadFailed", OpType: "install", Message: "image pull error: 502"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchCancel, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected cancel-on-downloadFailed to succeed (cancel won by default), got %v", err)
	}
	if row.State != "downloadFailed" {
		t.Fatalf("expected terminal row at downloadFailed, got %s", row.State)
	}
}

// TestWaitForTerminalCancelLandsOnStopped covers the partner regression:
// cancel of an in-flight install settles the row at `stopped` (post-
// rollback stable state, observed on partial-install charts that
// auto-shut-down their components on cancel). Same hang behavior
// before the fix; same expectation now — cancel watch must terminate
// successfully with the landed state surfaced verbatim.
func TestWaitForTerminalCancelLandsOnStopped(t *testing.T) {
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "installingCanceling", OpType: "install"},
		{State: "stopped", OpType: ""},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchCancel, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected cancel-on-stopped to succeed, got %v", err)
	}
	if row.State != "stopped" {
		t.Fatalf("expected terminal row at stopped, got %s", row.State)
	}
}

// TestWaitForTerminalCancelStillFailsOnCancelFailed is the regression
// guard for the other direction: *CancelFailed must still bubble up as
// a watch failure (non-zero exit). It's the only signal the user has
// that the cancel REQUEST itself was rejected (vs. settled).
func TestWaitForTerminalCancelStillFailsOnCancelFailed(t *testing.T) {
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "installingCancelFailed", OpType: "install", Message: "cannot cancel: stage already committed"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchCancel, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected installingCancelFailed to surface as cancel watch failure, got nil")
	}
	var failErr *watchFailureError
	if !errors.As(err, &failErr) {
		t.Fatalf("expected *watchFailureError, got %T (%v)", err, err)
	}
	if failErr.row.State != "installingCancelFailed" {
		t.Fatalf("failure row should carry installingCancelFailed, got %+v", failErr.row)
	}
}

func TestWaitForTerminalTimeoutSurfacesLastState(t *testing.T) {
	// Stuck in `installing` forever: classifier never reaches a terminal
	// set, so the deadline must fire and the error must carry the last
	// observed state.
	seq := []statusRow{{State: "installing", OpType: "install"}}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	var to *watchTimeoutError
	if !errors.As(err, &to) {
		t.Fatalf("expected watchTimeoutError, got %T: %v", err, err)
	}
	if to.last == nil || to.last.State != "installing" {
		t.Fatalf("expected last state installing, got %+v", to.last)
	}
	if !strings.Contains(err.Error(), "installing") {
		t.Fatalf("timeout error must surface last state, got %q", err.Error())
	}
}

// Sanity-guard that the JSON tags on OperationResult haven't drifted: with
// FinalState/FinalOpType empty, the JSON output must NOT contain those keys
// (so existing scripted consumers keep their byte-identical output).
func TestOperationResultJSONOmitsFinalFieldsWhenUnset(t *testing.T) {
	r := OperationResult{App: "a", Operation: "install", Status: "accepted"}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "finalState") || strings.Contains(string(b), "finalOpType") {
		t.Fatalf("non-watch JSON must omit finalState/finalOpType; got %s", b)
	}

	r.FinalState = "running"
	r.FinalOpType = "install"
	b, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"finalState":"running"`) {
		t.Fatalf("watch JSON must include finalState; got %s", b)
	}
}

// helper that builds a printable description of a watchTarget for error
// messages — keeps the test helpers self-contained.
func describeTarget(t watchTarget) string {
	return fmt.Sprintf("op=%s app=%s source=%s matchOpType=%v absentMeansSuccess=%v acceptInitialAbsent=%v idempotentSuccess=%v",
		t.op, t.appName, t.source, t.matchOpType, t.absentMeansSuccess, t.acceptInitialAbsent, t.idempotentSuccess)
}

var _ = describeTarget // referenced by future tests; suppress lint

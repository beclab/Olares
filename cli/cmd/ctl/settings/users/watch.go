package users

// Polling helper for `settings users create/delete --watch`.
//
// The shape mirrors cli/cmd/ctl/market/watch.go so the two surfaces feel
// the same to operators: opt-in `-w/--watch` with companion
// `--watch-timeout` / `--watch-interval` knobs, structured timeout /
// failure errors for JSON consumers, signal-aware cancellation, and a
// modest consecutive-error budget so a network blip doesn't kill an
// otherwise-healthy watch.
//
// Why a separate file: keeping the loop out of mutate.go makes the create
// / delete RunE bodies read as straight-line "POST → maybe wait → print"
// flows and lets the watch test target the loop in isolation.

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// userWatchOp labels the verb we kicked off. It exists for diagnostic
// messages only (the wire status field on accountModifyStatus does not
// echo the op back) but matches market's `watchOp` convention so the
// stderr lines feel uniform across the CLI.
type userWatchOp string

const (
	userWatchCreate userWatchOp = "create"
	userWatchDelete userWatchOp = "delete"
)

// userWatchTarget captures everything the polling loop needs to classify a
// /status response as success/failure. Built via newUserWatchTarget so
// callers do not reproduce the per-op state set lookup.
type userWatchTarget struct {
	op       userWatchOp
	username string

	successSet map[string]bool
	failureSet map[string]bool

	// absentMeansSuccess flips a transport-level 404 from "still
	// running" to "we're done". Only used for delete: the
	// app-service handler currently returns 200 with status=Deleted
	// for unknown users, but a future refactor may return 404 — this
	// keeps the watcher correct under either shape. Safe to enable
	// because runDelete only invokes the watch after a successful
	// DELETE on an existing user, so a 404 cannot be a typo.
	absentMeansSuccess bool
}

func newUserWatchTarget(op userWatchOp, username string) userWatchTarget {
	t := userWatchTarget{op: op, username: username}
	switch op {
	case userWatchCreate:
		t.successSet = map[string]bool{"Created": true}
		// Failed is the explicit provisioning failure. Deleted while
		// we expected Created means the row vanished mid-watch
		// (controller cleanup raced) — treat as failure so JSON
		// consumers get a non-zero exit instead of a hung watcher.
		t.failureSet = map[string]bool{"Failed": true, "Deleted": true}
	case userWatchDelete:
		t.successSet = map[string]bool{"Deleted": true}
		// No DeleteFailed status exists today; timeout is the only
		// failure mode for delete.
		t.failureSet = map[string]bool{}
		t.absentMeansSuccess = true
	}
	return t
}

// userWatchOptions carries the per-invocation knobs the loop needs. We
// keep it tiny and value-typed so the test can drive it directly.
type userWatchOptions struct {
	Timeout  time.Duration
	Interval time.Duration
	Progress bool
}

// userWatchTimeoutError surfaces a --watch-timeout elapse with the
// last-seen status row so callers can render it for the user.
type userWatchTimeoutError struct {
	target userWatchTarget
	last   *accountModifyStatus
}

func (e *userWatchTimeoutError) Error() string {
	if e.last != nil {
		return fmt.Sprintf("%s %q watch timed out (last status: %s)",
			e.target.op, e.target.username, valueOrUnknownStatus(e.last.Status))
	}
	return fmt.Sprintf("%s %q watch timed out (no status reported by the backend)",
		e.target.op, e.target.username)
}

// userWatchFailureError represents a terminal-failure classification.
type userWatchFailureError struct {
	target userWatchTarget
	status accountModifyStatus
}

func (e *userWatchFailureError) Error() string {
	parts := []string{fmt.Sprintf("status=%s", valueOrUnknownStatus(e.status.Status))}
	if detail := strings.TrimSpace(e.status.Message); detail != "" {
		parts = append(parts, "reason: "+detail)
	}
	return fmt.Sprintf("%s %q failed: %s",
		e.target.op, e.target.username, strings.Join(parts, " "))
}

// waitForUserState polls /api/users/{name}/status until the row reaches a
// terminal state per `t`, the deadline elapses, or the user interrupts.
// First poll fires immediately so a state that was already terminal at
// request time (race conditions, replayed mutations) returns without a
// wasted sleep.
//
// The function is shaped after cli/cmd/ctl/market/watch.go::waitForTerminal:
//
//   - signal.NotifyContext wraps parentCtx so SIGINT / SIGTERM produce a
//     "<op> '<name>' watch canceled by user" error distinguishable from a
//     parent-ctx cancellation.
//   - Transient transport errors increment a consecutive-error counter; we
//     bail with "watch aborted after N consecutive errors" once it hits 5.
//   - 401/403 short-circuit immediately (refresh-on-401 lives in the Doer;
//     persistent auth failures should not be retried by us).
//   - Transition logging only fires when the status string actually
//     changes, matching market's `last.State != row.State` rule.
func waitForUserState(parentCtx context.Context, d Doer, opts userWatchOptions, t userWatchTarget) (*accountModifyStatus, error) {
	interval := opts.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	timeoutDur := opts.Timeout
	if timeoutDur <= 0 {
		timeoutDur = 15 * time.Minute
	}
	deadline := time.Now().Add(timeoutDur)

	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	path := "/api/users/" + url.PathEscape(t.username) + "/status"

	var (
		last         *accountModifyStatus
		consecErrors int
	)

	for {
		if err := ctx.Err(); err != nil {
			if parentCtx.Err() == nil {
				return last, fmt.Errorf("%s %q watch canceled by user", t.op, t.username)
			}
			return last, err
		}
		if time.Now().After(deadline) {
			return last, &userWatchTimeoutError{target: t, last: last}
		}

		var st accountModifyStatus
		err := decodeObjectResult(ctx, d, path, &st)
		if err != nil {
			if t.absentMeansSuccess && httpStatusFromErrHint(err, 404) {
				st.Name = t.username
				st.Status = "Deleted"
				return &st, nil
			}
			if httpStatusFromErrHint(err, 401) || httpStatusFromErrHint(err, 403) {
				return last, err
			}
			if ctx.Err() != nil {
				continue
			}
			consecErrors++
			if opts.Progress {
				fmt.Fprintf(os.Stderr,
					"[%s user] transient status poll error (%v); retry in %s (consecutive=%d)…\n",
					t.op, err, formatDurationBrief(interval), consecErrors)
			}
			if consecErrors >= 5 {
				return last, fmt.Errorf("%s %q watch aborted after %d consecutive errors: %w",
					t.op, t.username, consecErrors, err)
			}
			if waitErr := sleepOrCancelUserWatch(ctx, interval); waitErr != nil {
				continue
			}
			continue
		}
		consecErrors = 0

		currentStatus := strings.TrimSpace(st.Status)
		if opts.Progress && (last == nil || strings.TrimSpace(last.Status) != currentStatus) {
			label := currentStatus
			if label == "" {
				label = "…"
			}
			fmt.Fprintf(os.Stderr,
				"[%s user] %q status=%s (next check in %s)…\n",
				t.op, t.username, label, formatDurationBrief(interval))
		}
		stCopy := st
		last = &stCopy

		if t.successSet[currentStatus] {
			return last, nil
		}
		if t.failureSet[currentStatus] {
			return last, &userWatchFailureError{target: t, status: st}
		}

		if err := sleepOrCancelUserWatch(ctx, interval); err != nil {
			continue
		}
	}
}

// sleepOrCancelUserWatch is a per-package copy of market's sleepOrCancel.
// We avoid importing market to keep the dependency graph one-directional
// (settings/users → no peer cmd/ctl trees).
func sleepOrCancelUserWatch(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func valueOrUnknownStatus(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

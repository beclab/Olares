// task.go: polling helper for /api/task/<node>/?task_id=<id>.
//
// The archive endpoints' compress / extract verbs are async: the
// POST returns a task_id and the actual work happens on the
// per-node task queue. The cobra layer's `--wait` flag drives
// this helper to block until the task reaches a terminal
// status, printing periodic progress updates via the supplied
// callback.
//
// Wire shape (mirrors the LarePass web app's olaresTask/index.ts
// getTask call — see apps/packages/app/src/services/olaresTask/
// index.ts):
//
//	GET /api/task/<node>/?task_id=<id>
//	200 → { code, msg, task: { id, status, progress, ... } }
//
// Status vocabulary (from
// services/abstractions/olaresTask/interface.ts OlaresTaskStatus):
//
//	pending, running, paused          ← still in flight
//	completed                         ← success (terminal)
//	failed                            ← terminal error
//	canceled, cancelled               ← cancelled (terminal; both
//	                                    spellings appear in the
//	                                    server's responses)
//
// We treat any other status as "still in flight" so a future
// server-side status (e.g. "queued") doesn't trip a spurious
// failure on old clients.
//
// Mirrors upload.Client.WaitCloudTask deliberately — that
// implementation has been load-tested against the real
// task-queue service for cloud uploads. Diverging here would
// add risk for no benefit; same wire shape, same polling
// strategy.
package archive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DefaultTaskPollInterval is how often WaitTask checks the
// task-status endpoint when the caller doesn't override it.
// 2 s mirrors upload.DefaultCloudTaskPollInterval — both
// endpoints share the same task queue, and 2 s is the same
// compromise between responsiveness and request volume.
const DefaultTaskPollInterval = 2 * time.Second

// Terminal status values. Lowercase as the server sends them.
// The double `canceled`/`cancelled` spelling is intentional —
// the server is inconsistent (see OlaresTaskStatus in the web
// app) so we accept both.
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusPaused    = "paused"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCanceled  = "canceled"
	TaskStatusCancelled = "cancelled"
)

// TaskUpdate is the per-poll snapshot WaitTask passes to its
// progress callback. The cobra layer renders human-friendly
// progress lines from these without needing to know the wire
// envelope.
//
// `Progress` is 0..100 (server-reported). It may stay at 0
// for short tasks where the writer never updates it before
// completing — don't treat 0 as "stuck".
type TaskUpdate struct {
	Status        string  // raw server status: pending / running / paused / ...
	Progress      float64 // 0..100
	CurrentPhase  int     // 1..TotalPhase (when present)
	TotalPhase    int
	TotalFileSize int64
	FailedReason  string
}

// TaskUpdateFunc receives one TaskUpdate per poll while the
// task is non-terminal. Pass nil from the cobra layer when the
// caller doesn't want progress updates.
type TaskUpdateFunc func(TaskUpdate)

// taskQueryEnvelope mirrors the JSON shape the web app's
// olaresTask getTask reads. We keep the type-set conservative:
// only fields we surface in TaskUpdate or branch on are
// decoded. Future fields the server adds are silently ignored
// (json.Unmarshal is lenient by default).
type taskQueryEnvelope struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Task struct {
		ID            string  `json:"id"`
		Status        string  `json:"status"`
		Progress      float64 `json:"progress"`
		CurrentPhase  int     `json:"current_phase"`
		TotalPhase    int     `json:"total_phase"`
		TotalFileSize int64   `json:"total_file_size"`
		FailedReason  string  `json:"failed_reason,omitempty"`
	} `json:"task"`
}

// WaitTask polls /api/task/<node>/?task_id=<taskID> at
// `interval` (or DefaultTaskPollInterval if interval == 0) and
// returns when the task reaches a terminal status:
//
//   - completed                  → nil
//   - failed                     → fmt.Errorf with `failed_reason`
//   - canceled / cancelled       → fmt.Errorf("task ... was cancelled")
//
// onUpdate runs once per poll while the task is non-terminal;
// pass nil when the caller doesn't want progress notifications.
//
// ctx cancellation is honoured between polls AND inside each
// HTTP request via the underlying transport. Transient HTTP
// errors are surfaced verbatim — a task we can't query is
// indistinguishable from a stuck one, and the caller should
// bubble the failure up rather than burn cycles guessing.
//
// `node` MUST be the same node the compress/extract call
// targeted. The task queue is per-node — querying the wrong
// node returns "task not found" which we surface as a clean
// error (the cobra layer's reformatter explains the cause).
func (c *Client) WaitTask(
	ctx context.Context,
	node, taskID string,
	interval time.Duration,
	onUpdate TaskUpdateFunc,
) error {
	if taskID == "" {
		return errors.New("WaitTask: empty taskID")
	}
	if node == "" {
		return errors.New("WaitTask: empty node")
	}
	if interval <= 0 {
		interval = DefaultTaskPollInterval
	}

	q := url.Values{}
	q.Set("task_id", taskID)
	endpoint := c.BaseURL + "/api/task/" + url.PathEscape(node) + "/?" + q.Encode()

	for {
		body, err := c.do(ctx, http.MethodGet, endpoint, nil, "", "")
		if err != nil {
			return fmt.Errorf("query task %s on node %s: %w", taskID, node, err)
		}
		var env taskQueryEnvelope
		if len(body) > 0 {
			if err := json.Unmarshal(body, &env); err != nil {
				return fmt.Errorf("decode task query response for %s: %w (body=%s)",
					taskID, err, truncateBody(body))
			}
		}

		switch env.Task.Status {
		case TaskStatusCompleted:
			return nil
		case TaskStatusFailed:
			reason := env.Task.FailedReason
			if reason == "" {
				reason = "server reported failure with no failed_reason"
			}
			return fmt.Errorf("archive task %s failed: %s", taskID, reason)
		case TaskStatusCanceled, TaskStatusCancelled:
			return fmt.Errorf("archive task %s was cancelled server-side", taskID)
		}

		if onUpdate != nil {
			onUpdate(TaskUpdate{
				Status:        env.Task.Status,
				Progress:      env.Task.Progress,
				CurrentPhase:  env.Task.CurrentPhase,
				TotalPhase:    env.Task.TotalPhase,
				TotalFileSize: env.Task.TotalFileSize,
				FailedReason:  env.Task.FailedReason,
			})
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// CancelTask sends DELETE /api/task/<node>/?task_id=<id> to
// cancel a running task. The server's `OlaresTaskManager.cancelTask`
// is the same wire path. Used by the cobra layer's --wait Ctrl-C
// handler to drop a queued compress / extract task on user
// abort (so the user doesn't end up with a half-built archive
// on disk).
func (c *Client) CancelTask(ctx context.Context, node, taskID string) error {
	if taskID == "" {
		return errors.New("CancelTask: empty taskID")
	}
	if node == "" {
		return errors.New("CancelTask: empty node")
	}
	q := url.Values{}
	q.Set("task_id", taskID)
	endpoint := c.BaseURL + "/api/task/" + url.PathEscape(node) + "/?" + q.Encode()
	_, err := c.do(ctx, http.MethodDelete, endpoint, nil, "", "")
	if err != nil {
		return fmt.Errorf("cancel task %s on node %s: %w", taskID, node, err)
	}
	return nil
}

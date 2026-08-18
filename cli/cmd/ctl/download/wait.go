package download

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewWaitCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout time.Duration
		output  string
	)
	cmd := &cobra.Command{
		Use:   "wait <id>",
		Short: "wait until a download task reaches a terminal status",
		Long: `Poll task info until the download reaches a true terminal status.

Success terminals: completed, seeding (bytes landed; seeding is the
BitTorrent post-complete state).
Failure terminals: error, cancelled, removed.

waiting_to_move and moving are NOT success — the yt-dlp mover is still
relocating bytes to the destination, so wait keeps polling.

An error the server will retry on its own (will_auto_retry) is not a
terminal failure; wait keeps polling until the retry sweep settles it.

Polling interval is 2s. On --timeout expiry the command exits non-zero
and reports the last observed status. This command uses HTTP polling
only; it does not switch to WebSocket watch.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runWait(c.Context(), f, args[0], timeout, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "max wait duration (0 = "+waitDefaultTimeout.String()+")")
	return cmd
}

func runWait(ctx context.Context, f *cmdutil.Factory, idRaw string, timeout time.Duration, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	if timeout < 0 {
		return fmt.Errorf("unsupported --timeout %s (need >= 0)", timeout)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	task, err := waitForTerminal(ctx, pc, id, timeout)
	if err != nil {
		return err
	}
	kind := classifyWaitStatus(task)
	switch format {
	case FormatJSON:
		if err := printJSON(os.Stdout, task); err != nil {
			return err
		}
	default:
		fmt.Printf("Task %d reached %s status=%s\n", task.ID, kind, task.Status)
	}
	if kind == "failure" {
		return fmt.Errorf("task %d ended in status %q", task.ID, task.Status)
	}
	return nil
}

// waitTimeoutError carries the last observed row so callers can still
// report the task id (and hand it to a JSON consumer) after giving up.
type waitTimeoutError struct {
	last DownloadTask
}

func (e *waitTimeoutError) Error() string {
	if e.last.ID == 0 {
		return "wait timed out before the first info response"
	}
	return fmt.Sprintf("wait timed out: task %d still status=%s", e.last.ID, e.last.Status)
}

// waitForTerminal polls GET /api/download/info/<id> until the row
// classifies as terminal, the deadline passes, or the user interrupts.
// Shaped after cmd/ctl/market/watch.go::waitForTerminal.
func waitForTerminal(parentCtx context.Context, pc *preparedClient, id int64, timeout time.Duration) (DownloadTask, error) {
	if timeout <= 0 {
		timeout = waitDefaultTimeout
	}
	deadline := time.Now().Add(timeout)

	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	path := fmt.Sprintf("/api/download/info/%d", id)
	var (
		last         DownloadTask
		consecErrors int
	)
	for {
		if err := ctx.Err(); err != nil {
			if parentCtx.Err() == nil {
				return last, fmt.Errorf("wait on task %d canceled by user", id)
			}
			return last, err
		}
		if time.Now().After(deadline) {
			return last, &waitTimeoutError{last: last}
		}

		var task DownloadTask
		if err := doGet(ctx, pc.doer, path, &task); err != nil {
			if ctx.Err() != nil {
				continue
			}
			consecErrors++
			if consecErrors >= waitMaxConsecErrors {
				return last, fmt.Errorf("wait on task %d aborted after %d consecutive errors: %w", id, consecErrors, err)
			}
			fmt.Fprintf(os.Stderr, "wait: transient info poll error (%v); retry in %s (consecutive=%d)\n",
				err, waitPollInterval, consecErrors)
			sleepOrCancelWait(ctx, waitPollInterval)
			continue
		}
		consecErrors = 0
		last = task
		if kind := classifyWaitStatus(task); kind != "pending" {
			return task, nil
		}
		sleepOrCancelWait(ctx, waitPollInterval)
	}
}

func sleepOrCancelWait(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// classifyWaitStatus returns "success", "failure", or "pending".
// Mover phases and a will_auto_retry error are pending: the bytes are
// not on disk yet, and the server's retry sweep still owns that row.
func classifyWaitStatus(task DownloadTask) string {
	switch task.Status {
	case "completed", "seeding":
		return "success"
	case "error":
		if task.WillAutoRetry {
			return "pending"
		}
		return "failure"
	case "cancelled", "removed":
		return "failure"
	default:
		return "pending"
	}
}

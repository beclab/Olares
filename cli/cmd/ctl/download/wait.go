package download

import (
	"context"
	"fmt"
	"os"
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

Polling interval is 2s. On --timeout expiry the command exits non-zero
and prints the current status. This command uses HTTP polling only;
it does not switch to WebSocket watch.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runWait(c.Context(), f, args[0], timeout, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "max wait duration (0 = no limit)")
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
	kind := classifyWaitStatus(task.Status)
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

// waitForTerminal polls GET /api/download/info/<id> until a terminal
// status or timeout. timeout<=0 means wait indefinitely (still
// cancelled by ctx).
func waitForTerminal(ctx context.Context, pc *preparedClient, id int64, timeout time.Duration) (DownloadTask, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var last DownloadTask
	for {
		var task DownloadTask
		path := fmt.Sprintf("/api/download/info/%d", id)
		if err := doGet(ctx, pc.doer, path, &task); err != nil {
			if ctx.Err() != nil && last.ID != 0 {
				return last, fmt.Errorf("wait timed out: task %d still status=%s", last.ID, last.Status)
			}
			if ctx.Err() != nil {
				return DownloadTask{}, fmt.Errorf("wait timed out before first info response")
			}
			return DownloadTask{}, err
		}
		last = task
		kind := classifyWaitStatus(task.Status)
		if kind == "success" || kind == "failure" {
			return task, nil
		}
		timer := time.NewTimer(waitPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, fmt.Errorf("wait timed out: task %d still status=%s", last.ID, last.Status)
		case <-timer.C:
		}
	}
}

// classifyWaitStatus returns "success", "failure", or "pending".
// Mover phases are pending so scripts wait for bytes on disk.
func classifyWaitStatus(status string) string {
	switch status {
	case "completed", "seeding":
		return "success"
	case "error", "cancelled", "removed":
		return "failure"
	default:
		return "pending"
	}
}

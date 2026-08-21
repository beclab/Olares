package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
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
Failure terminals: error, cancelled, removed. Terminal status is
decided from status alone — will_auto_retry does not keep wait
polling; it only changes the hint printed on an error row.

waiting_to_move and moving are NOT success — the yt-dlp mover is still
relocating bytes to the destination, so wait keeps polling.

Table mode prints a watching header on stdout, then compact ticks
on stderr ("[id] downloading 21.4%  elapsed=12s") on status changes
or percent jumps of at least 5%. If nothing moves, a heartbeat
reprints the current tick every 10s so a stall does not look hung.
A market-style summary is printed when the wait ends (completed,
failed, or timed out). This command uses HTTP polling only; it does
not switch to WebSocket watch.`,
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
	if format != FormatJSON {
		fmt.Printf("watching task %d until terminal (timeout: %s)...\n", id, resolveWaitTimeout(timeout))
	}
	task, err := waitForTerminal(ctx, pc, id, timeout, waitProgressWriter(format))
	if err != nil {
		emitWaitFinish(format, task, err)
		return err
	}
	kind := classifyWaitStatus(task)
	if format == FormatJSON {
		if err := printJSON(os.Stdout, task); err != nil {
			return err
		}
		if kind == "failure" {
			emitWaitFailureDetails(os.Stderr, task, format)
			return waitFailureError(task)
		}
		return nil
	}
	if kind == "failure" {
		emitWaitOutcome(os.Stderr, task, "failure")
		return waitFailureError(task)
	}
	emitWaitOutcome(os.Stdout, task, "success")
	return nil
}

func resolveWaitTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return waitDefaultTimeout
	}
	return timeout
}

func waitProgressWriter(format Format) io.Writer {
	if format == FormatJSON {
		return nil
	}
	return os.Stderr
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
// progress, when non-nil, receives compact ticks (table mode).
func waitForTerminal(parentCtx context.Context, pc *preparedClient, id int64, timeout time.Duration, progress io.Writer) (DownloadTask, error) {
	timeout = resolveWaitTimeout(timeout)
	deadline := time.Now().Add(timeout)

	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	path := fmt.Sprintf("/api/download/info/%d", id)
	var (
		last         DownloadTask
		consecErrors int
		prog         waitProgress
	)
	if progress != nil {
		prog.w = progress
	}
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
		kind := classifyWaitStatus(task)
		if kind == "pending" {
			prog.emit(task)
		}
		if kind != "pending" {
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
// Terminal status is decided from status alone. will_auto_retry is
// copy for the error hint, not a reason to keep polling.
func classifyWaitStatus(task DownloadTask) string {
	switch task.Status {
	case "completed", "seeding":
		return "success"
	case "error", "cancelled", "removed":
		return "failure"
	default:
		return "pending"
	}
}

func waitFailureError(task DownloadTask) error {
	if msg := strings.TrimSpace(task.ErrMsg); msg != "" {
		return fmt.Errorf("task %d ended in status %q: %s", task.ID, task.Status, msg)
	}
	return fmt.Errorf("task %d ended in status %q", task.ID, task.Status)
}

func waitRetryHint(task DownloadTask) string {
	if task.Status != "error" {
		return ""
	}
	if task.WillAutoRetry {
		return fmt.Sprintf("server will auto-retry this task; inspect later with `olares-cli knowledge download info %d`", task.ID)
	}
	return "this error is final (will_auto_retry=false)"
}

func emitWatchingStart(w io.Writer, task DownloadTask, timeout time.Duration) error {
	if w == nil {
		w = os.Stdout
	}
	_, err := fmt.Fprintf(w, "Created task %d; watching until terminal (timeout: %s)...\n", task.ID, resolveWaitTimeout(timeout))
	if err != nil {
		return fmt.Errorf("write created task %d: %w", task.ID, err)
	}
	if prov := strings.TrimSpace(task.DownloadProvider); prov != "" {
		fmt.Fprintf(w, "  provider: %s\n", prov)
	}
	fmt.Fprintf(w, "  name: %s\n", displayName(task))
	return nil
}

func emitWaitFinish(format Format, task DownloadTask, waitErr error) {
	if format == FormatJSON || task.ID == 0 {
		return
	}
	var timeoutErr *waitTimeoutError
	if waitErr != nil && errors.As(waitErr, &timeoutErr) {
		emitWaitOutcome(os.Stderr, task, "timeout")
	}
}

// emitWaitOutcome prints the market-style verdict after ticks stop.
func emitWaitOutcome(w io.Writer, task DownloadTask, kind string) {
	if w == nil || task.ID == 0 {
		return
	}
	switch kind {
	case "success":
		fmt.Fprintf(w, "download '%d': completed (status=%s)\n", task.ID, orDash(task.Status))
	case "timeout":
		fmt.Fprintf(w, "download '%d': watch timed out (status=%s)\n", task.ID, orDash(task.Status))
	default:
		msg := strings.TrimSpace(task.ErrMsg)
		if msg == "" {
			msg = orDash(task.Status)
		}
		fmt.Fprintf(w, "download '%d' failed: %s\n", task.ID, msg)
	}
	if prov := strings.TrimSpace(task.DownloadProvider); prov != "" {
		fmt.Fprintf(w, "  provider: %s\n", prov)
	}
	fmt.Fprintf(w, "  name: %s\n", displayName(task))
	if path := strings.TrimSpace(task.Path); path != "" {
		fmt.Fprintf(w, "  path: %s\n", path)
	}
	if kind == "failure" {
		if cat := strings.TrimSpace(task.ErrCategory); cat != "" {
			fmt.Fprintf(w, "  err_category: %s\n", cat)
		}
		if hint := waitRetryHint(task); hint != "" {
			fmt.Fprintf(w, "  %s\n", hint)
		}
		if hint := taskCookieHint(task); hint != "" {
			fmt.Fprintln(w, hint)
		}
	}
}

// emitWaitFailureDetails writes JSON-mode retry / cookie hints to stderr.
func emitWaitFailureDetails(w io.Writer, task DownloadTask, format Format) {
	if w == nil || task.ID == 0 || format != FormatJSON {
		return
	}
	if hint := waitRetryHint(task); hint != "" {
		fmt.Fprintln(w, hint)
	}
	if hint := taskCookieHint(task); hint != "" {
		fmt.Fprintln(w, hint)
	}
}

// waitProgress prints compact ticks: `[6] downloading 21.4%  elapsed=12s`.
// Status changes always print. Percent-only ticks need a jump of
// waitProgressPercentStep. Unchanged polls still reprint after
// waitProgressHeartbeat so a stall is visible. The first still-waiting
// 0% poll is skipped because the watching header already said created.
type waitProgress struct {
	w           io.Writer
	seen        bool
	started     time.Time
	lastPrint   time.Time
	lastStatus  string
	lastPercent float32
	lastErrMsg  string
}

func (p *waitProgress) emit(task DownloadTask) {
	if p == nil || p.w == nil {
		return
	}
	now := time.Now()
	if !p.seen {
		p.seen = true
		p.started = now
		p.lastStatus = task.Status
		p.lastPercent = task.Percent
		p.lastErrMsg = task.ErrMsg
		if task.Status == "waiting" && task.Percent == 0 {
			p.lastPrint = now
			return
		}
		p.print(task, now)
		return
	}
	statusChanged := task.Status != p.lastStatus
	errChanged := task.ErrMsg != p.lastErrMsg
	percentJump := task.Percent-p.lastPercent >= waitProgressPercentStep || task.Percent < p.lastPercent
	stalled := waitProgressHeartbeat > 0 && now.Sub(p.lastPrint) >= waitProgressHeartbeat
	if !statusChanged && !errChanged && !percentJump && !stalled {
		return
	}
	p.print(task, now)
}

func (p *waitProgress) print(task DownloadTask, now time.Time) {
	p.lastStatus = task.Status
	p.lastPercent = task.Percent
	p.lastErrMsg = task.ErrMsg
	p.lastPrint = now
	elapsed := now.Sub(p.started).Truncate(time.Second)
	fmt.Fprintf(p.w, "[%d] %s %.1f%%  elapsed=%s\n", task.ID, orDash(task.Status), task.Percent, elapsed)
}

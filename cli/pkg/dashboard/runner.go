package dashboard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// RunOnce is the per-iteration callback every leaf command supplies. The
// watch loop calls it once per tick; the leaf is responsible for building
// and emitting (or returning) one Envelope.
//
// `iter` is 1-based. `now` is the wall clock anchor for the current
// iteration — the leaf should use it to compute time windows when
// --since is set.
//
// Returning a non-nil error signals iteration failure. The watch loop
// will:
//
//   - emit a single NDJSON line with Meta.Error / Meta.ErrorKind populated
//     to keep the stream intact (JSON output);
//   - log a warning to stderr (table output);
//   - bump the consecutive-failure counter; after FailureThreshold (default
//     3) consecutive failures, the loop exits non-zero;
//   - on credential.ErrTokenInvalidated / ErrNotLoggedIn, exit immediately
//     non-zero regardless of failure count (re-running login is a hard
//     prerequisite).
//
// One-shot commands (no --watch) reuse RunOnce too: the runner just calls
// it exactly once and returns its error.
type RunOnce func(ctx context.Context, iter int, now time.Time) (Envelope, error)

// Runner is the per-command "execute me with the user's --watch flags"
// orchestrator. Embed CommonFlags into your command, register a RunOnce,
// and call Runner.Run(ctx) from RunE.
type Runner struct {
	// Flags is the resolved CommonFlags after Validate() ran.
	Flags *CommonFlags

	// Recommended is the SPA-side polling cadence (typically pulled from
	// the metric catalog). 0 means "this command is not poll-able"; in
	// that case --watch is rejected up front. Otherwise it's the default
	// for --watch-interval when the user didn't set one.
	Recommended time.Duration

	// FailureThreshold is the consecutive-failure cap before the watch
	// loop bails out. Defaults to 3.
	FailureThreshold int

	// Stdout / Stderr default to os.Stdout / os.Stderr. Tests override
	// them.
	Stdout io.Writer
	Stderr io.Writer

	// Now is the clock used by the watch loop. nil = time.Now. Tests
	// inject a fake clock to drive the ticker deterministically.
	Now func() time.Time

	// Sleep is the scheduling primitive. nil = time.NewTimer-based; tests
	// inject a fake sleeper that advances Now in lockstep.
	Sleep func(ctx context.Context, d time.Duration) error

	// Iter is the per-iteration callback. Required. We name the field
	// Iter (not Run) so it doesn't collide with the entry-point method
	// Runner.Run.
	Iter RunOnce
}

// Run is the entry point. With --watch off, calls RunOnce exactly once and
// emits one Envelope. With --watch on, runs the ticker until iterations /
// timeout / failure threshold / SIGINT triggers an exit.
func (r *Runner) Run(ctx context.Context) error {
	if r.Iter == nil {
		return errors.New("watch: Iter is required")
	}
	stdout := r.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := r.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	now := r.Now
	if now == nil {
		now = time.Now
	}
	sleep := r.Sleep
	if sleep == nil {
		sleep = sleepCtx
	}
	threshold := r.FailureThreshold
	if threshold <= 0 {
		threshold = 3
	}

	// SIGINT/SIGTERM cancel the context so in-flight requests bail out
	// promptly. Cancelled context errors are downgraded to nil so a
	// graceful Ctrl-C doesn't render as a failure.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := installSignalHandler(cancel)
	defer stop()

	// One-shot path.
	if !r.Flags.Watch {
		env, err := r.Iter(ctx, 1, now())
		if err != nil {
			return r.fatal(err)
		}
		return r.emit(stdout, env)
	}

	// Watch path.
	if r.Recommended <= 0 {
		return errors.New("--watch is not supported for this command (no recommended polling cadence)")
	}
	interval := r.Flags.WatchInterval
	if interval <= 0 {
		interval = r.Recommended
	} else if interval < r.Recommended {
		fmt.Fprintf(stderr,
			"warning: --watch-interval=%s is faster than the SPA's recommended cadence of %s; "+
				"this may stress the BFF\n",
			interval, r.Recommended)
	}

	deadline := time.Time{}
	if r.Flags.WatchTimeout > 0 {
		deadline = now().Add(r.Flags.WatchTimeout)
	}

	failures := 0
	for iter := 1; ; iter++ {
		t := now()
		if !deadline.IsZero() && !t.Before(deadline) {
			return nil
		}

		env, err := r.Iter(ctx, iter, t)
		if err != nil {
			// Hard exits — no recovery possible.
			if isFatalCredErr(err) {
				return r.fatal(err)
			}
			failures++
			r.emitFailure(stdout, stderr, env, err, iter, r.Flags.WatchIterations)
			if failures >= threshold {
				return fmt.Errorf("watch: aborted after %d consecutive failures: %w", failures, err)
			}
		} else {
			failures = 0
			env.Meta.Iteration = iter
			if r.Flags.WatchIterations > 0 {
				env.Meta.TotalIterations = r.Flags.WatchIterations
			}
			if err := r.emit(stdout, env); err != nil {
				return err
			}
		}

		if r.Flags.WatchIterations > 0 && iter >= r.Flags.WatchIterations {
			return nil
		}

		// Sleep till the next tick. We compute the "remaining" against
		// the deadline so a long-running iteration doesn't push us
		// past --watch-timeout.
		remaining := interval
		if !deadline.IsZero() {
			rem := time.Until(deadline)
			if rem < remaining {
				remaining = rem
			}
		}
		if remaining <= 0 {
			return nil
		}
		if err := sleep(ctx, remaining); err != nil {
			// Cancelled / SIGINT: graceful exit.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
	}
}

// emit writes one Envelope to stdout in the active format. Errors from the
// writer (typically EPIPE when the user piped into `head`) are surfaced so
// the caller can downgrade them to a graceful exit if desired.
func (r *Runner) emit(w io.Writer, env Envelope) error {
	if r.Flags.Output == OutputJSON {
		return WriteJSON(w, env)
	}
	// Table mode is leaf-driven: leaves render their own table after
	// returning. Re-rendering here would double-print.
	return nil
}

// emitFailure prints the per-iteration failure record. JSON mode keeps the
// NDJSON stream alive by emitting an envelope with Meta.Error set; table
// mode just logs to stderr (we don't want to splat half-rendered tables on
// stdout).
func (r *Runner) emitFailure(stdout, stderr io.Writer, env Envelope, err error, iter, total int) {
	kind := ClassifyTransportErr(err)
	if r.Flags.Output == OutputJSON {
		if env.Kind == "" {
			// No partial envelope — emit a minimal placeholder so the
			// NDJSON stream still has a kind / meta block.
			env.Kind = "dashboard.iteration.error"
		}
		env.Meta.Iteration = iter
		if total > 0 {
			env.Meta.TotalIterations = total
		}
		env.Meta.Error = err.Error()
		env.Meta.ErrorKind = kind
		_ = WriteJSON(stdout, env)
		return
	}
	fmt.Fprintf(stderr, "warning: iteration %d failed (%s): %v\n", iter, kind, err)
}

// fatal prints the err and returns it unchanged. Cobra's RunE passes it up
// to the root command, which (with SilenceErrors=true) leaves stderr clean.
func (r *Runner) fatal(err error) error {
	return err
}

// isFatalCredErr returns true for the typed credential errors that mean
// "no amount of retrying will recover; the user must re-login". The watch
// loop short-circuits on these regardless of the failure counter.
func isFatalCredErr(err error) bool {
	var inv *credential.ErrTokenInvalidated
	if errors.As(err, &inv) {
		return true
	}
	var nli *credential.ErrNotLoggedIn
	if errors.As(err, &nli) {
		return true
	}
	return false
}

// sleepCtx is a context-aware time.Sleep. Returns nil when d elapses; a
// context.Canceled / DeadlineExceeded when ctx fires first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	tm := time.NewTimer(d)
	defer tm.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-tm.C:
		return nil
	}
}

// installSignalHandler routes SIGINT / SIGTERM to cancel(). Returns a stop
// func the caller defers to release the signal.Notify channel.
func installSignalHandler(cancel context.CancelFunc) func() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-sig:
			cancel()
		case <-done:
		}
	}()
	return func() {
		signal.Stop(sig)
		close(done)
	}
}

package pod

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// LogsOptions captures every per-call tunable for `cluster pod logs`
// / `cluster container logs`. Exported so the container-package alias
// can build the same struct without a mirror declaration.
//
// Defaults are applied inside RunLogs so the alias and the direct
// invocation share identical behavior.
type LogsOptions struct {
	// Container scopes the request to a single container. Required
	// when the target pod has more than one container; otherwise the
	// sole container is auto-selected.
	Container string

	// TailLines limits the initial fetch to the last N lines. 0 means
	// "no tail" (server returns the whole log buffer it has — same
	// semantics as kubectl). Subsequent --follow polls always use
	// sinceTime, never tailLines, so this only affects the first
	// fetch.
	TailLines int

	// SinceSeconds restricts the initial fetch to logs newer than N
	// seconds ago. 0 disables. Mutually exclusive with TailLines on
	// the wire (the server picks one); we let TailLines win when both
	// are set, mirroring kubectl.
	SinceSeconds int

	// LimitBytes caps the response body size. 0 disables. Useful when
	// pulling logs from a chatty container without --follow.
	LimitBytes int

	// Timestamps asks the server to prefix each line with an RFC3339
	// timestamp. Default true (the SPA pins it the same way) so
	// users can correlate output across containers without an extra
	// flag flip.
	Timestamps bool

	// Follow turns on the polling loop. When false, RunLogs makes
	// one request and returns. When true, RunLogs blocks until the
	// caller cancels (Ctrl-C / SIGTERM / context).
	Follow bool

	// Interval is the polling interval used when Follow is true. 0
	// defaults to 2s. Polling-based --follow is a deliberate choice
	// (mirrors `cluster <verb> --watch` policy and the SPA's own 5s
	// realtime tick) — we don't open a long-lived chunked connection.
	Interval time.Duration

	// Previous fetches the previous container instance's log buffer
	// (after a crash / restart). Maps to the upstream `?previous=true`
	// flag. Mutually exclusive with Follow on the wire.
	Previous bool
}

// NewLogsCommand: `olares-cli cluster pod logs <ns/pod | pod>
// [-n NS] [-c NAME] [--tail N] [-f] ...`.
//
// Calls SPA's getContainersLogs
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:104):
// `/api/v1/namespaces/<ns>/pods/<name>/log` — K8s native logs API.
//
// Container selection:
//   - --container <name>: explicit, no preflight.
//   - omitted: we GET the pod first; sole container auto-selected;
//     multi-container pods error with the available container list
//     (kubectl behavior, no surprise picks).
//
// --follow uses polling (sinceTime advances per tick), NOT chunked
// streaming. Same rationale as `olares-cli market <verb> --watch`:
// uniform behavior across long-running CLI verbs and a single source
// of truth for the polling interval / Ctrl-C plumbing.
func NewLogsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		tail      int
		since     time.Duration
		limitB    int
		ts        bool
		follow    bool
		interval  time.Duration
		previous  bool
	)
	cmd := &cobra.Command{
		Use:   "logs <ns/pod | pod>",
		Short: "stream a pod container's log buffer (--follow polls, doesn't stream)",
		Long: `Print a pod container's log buffer.

Identity follows the same "<namespace>/<pod>" or "-n <ns> <pod>"
convention as ` + "`cluster pod get`" + `. Container selection:

  --container <name>    explicit; passed through verbatim
  (omitted)             auto-select if the pod has exactly one
                        container; error otherwise (with the list)

--follow turns on polling-based tail (the next request uses
sinceTime=<previous fetch start>). This matches the rest of the
` + "`olares-cli market` --watch" + ` family — no long-lived chunked
streams, Ctrl-C cancels cleanly, and the polling interval is tunable
via --interval (default 2s).

--previous fetches the previous container instance's logs (after a
crash). The upstream API rejects --previous with --follow.

Output is forwarded verbatim. With --timestamps (default true) the
server prefixes each line with an RFC3339 timestamp, which is what
the SPA pins as well so output is correlatable across windows.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			if previous && follow {
				return fmt.Errorf("--follow and --previous are mutually exclusive")
			}
			return RunLogs(c.Context(), o, ns, name, LogsOptions{
				Container:    container,
				TailLines:    tail,
				SinceSeconds: int(since / time.Second),
				LimitBytes:   limitB,
				Timestamps:   ts,
				Follow:       follow,
				Interval:     interval,
				Previous:     previous,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare pod name)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container to read logs from (required when the pod has multiple containers)")
	cmd.Flags().IntVar(&tail, "tail", 200, "show the last N lines on the initial fetch (0 = unlimited; --follow always advances by sinceTime after the first fetch)")
	cmd.Flags().DurationVar(&since, "since", 0, "show logs newer than this duration ago on the initial fetch (e.g. 5m, 1h); 0 = unlimited")
	cmd.Flags().IntVar(&limitB, "limit-bytes", 0, "cap the response body size in bytes (0 = unlimited)")
	cmd.Flags().BoolVar(&ts, "timestamps", true, "ask the server to prefix every line with an RFC3339 timestamp")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "keep polling for new lines until interrupted (Ctrl-C to stop)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval when --follow is set")
	cmd.Flags().BoolVar(&previous, "previous", false, "fetch the previous container instance's logs (after a crash); incompatible with --follow")
	return cmd
}

// RunLogs is the exported entry point used by both `cluster pod logs`
// and `cluster container logs`. Container-package wrappers should
// always populate Container explicitly (the user already had to name
// it to invoke the alias) so we skip the auto-resolve preflight; the
// pod-package wrapper falls back to the auto-resolve path.
//
// Returns nil on graceful exit (Ctrl-C / SIGTERM / context cancel).
func RunLogs(ctx context.Context, o *clusteropts.ClusterOptions, namespace, podName string, opts LogsOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}

	container := strings.TrimSpace(opts.Container)
	if container == "" {
		// Preflight: GET the pod, auto-select its sole container, or
		// error with the full container list when ambiguous. Same
		// behavior as `kubectl logs` without -c on a multi-container
		// pod. The extra round trip is the cost of friendliness.
		p, err := Get(ctx, o, namespace, podName)
		if err != nil {
			return err
		}
		container, err = pickContainer(p)
		if err != nil {
			return err
		}
	}

	base := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log",
		url.PathEscape(namespace), url.PathEscape(podName))

	// First fetch: tail / since / limit-bytes apply here.
	q := url.Values{}
	q.Set("container", container)
	if opts.Timestamps {
		q.Set("timestamps", "true")
	}
	if opts.Previous {
		q.Set("previous", "true")
	}
	if opts.TailLines > 0 {
		q.Set("tailLines", fmt.Sprintf("%d", opts.TailLines))
	}
	if opts.SinceSeconds > 0 && opts.TailLines == 0 {
		// K8s rejects (sinceSeconds + tailLines together) loosely —
		// it picks one server-side. We mirror kubectl's "tail wins
		// if both supplied" policy by only sending sinceSeconds when
		// tail is unset.
		q.Set("sinceSeconds", fmt.Sprintf("%d", opts.SinceSeconds))
	}
	if opts.LimitBytes > 0 {
		q.Set("limitBytes", fmt.Sprintf("%d", opts.LimitBytes))
	}

	// Anchor the polling cursor BEFORE the first fetch so any line
	// the server returns is older than `nextSince` and won't be
	// duplicated by the first --follow tick.
	nextSince := time.Now()
	body, err := clusterclient.GetRaw(ctx, client, base+"?"+q.Encode())
	if err != nil {
		return err
	}
	writeChunk(o, body)

	if !opts.Follow {
		return nil
	}

	// signal.NotifyContext follows the same convention used by
	// `market --watch` (cli/cmd/ctl/market/watch.go). Distinguishes
	// "user pressed Ctrl-C" (return nil) from "parent context died
	// for some other reason" (return ctx.Err()).
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	interval := opts.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}

	consecErr := 0
	for {
		if err := sleepCtx(ctx, interval); err != nil {
			// Either parent ctx cancel or the user pressed Ctrl-C —
			// graceful exit.
			return nil
		}

		// Capture the cursor for the NEXT iteration BEFORE issuing
		// this fetch so any line the server emits during this fetch
		// is included by the next sinceTime window.
		nextNext := time.Now()

		q := url.Values{}
		q.Set("container", container)
		if opts.Timestamps {
			q.Set("timestamps", "true")
		}
		q.Set("sinceTime", nextSince.Format(time.RFC3339Nano))
		if opts.LimitBytes > 0 {
			q.Set("limitBytes", fmt.Sprintf("%d", opts.LimitBytes))
		}

		body, err := clusterclient.GetRaw(ctx, client, base+"?"+q.Encode())
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			// Mirror waitForTerminal in market/watch.go: tolerate up
			// to 5 consecutive errors so a transient 5xx / network
			// blip doesn't kill an otherwise-healthy --follow.
			consecErr++
			o.Info("logs: failed to fetch (retry %d): %v", consecErr, err)
			if consecErr >= 5 {
				return fmt.Errorf("logs: aborted after %d consecutive errors: %w", consecErr, err)
			}
			continue
		}
		consecErr = 0
		writeChunk(o, body)
		nextSince = nextNext
	}
}

// pickContainer is the single-vs-multi container selection rule shared
// by every logs entry point. Multi-container pods get a deterministic
// list of names so the user can copy/paste a --container choice
// without going hunting in `cluster pod get`.
func pickContainer(p *Pod) (string, error) {
	switch len(p.Spec.Containers) {
	case 0:
		return "", fmt.Errorf("pod %s/%s has no containers in spec", p.Metadata.Namespace, p.Metadata.Name)
	case 1:
		return p.Spec.Containers[0].Name, nil
	}
	names := make([]string, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		names = append(names, c.Name)
	}
	return "", fmt.Errorf("pod %s/%s has multiple containers (%s); pick one with --container",
		p.Metadata.Namespace, p.Metadata.Name, strings.Join(names, ", "))
}

// writeChunk forwards a server response body to stdout, normalizing
// trailing newlines so consecutive polling chunks don't collapse into
// a wall-of-text run-on. Empty body means "no new lines this tick" —
// stay silent so the user's terminal isn't spammed with blank lines.
//
// Honors --quiet by routing through ClusterOptions.WriteStdout (same
// guard as `cluster {pod,workload,...} yaml`); a quiet --follow run
// still polls and surfaces hard errors, just doesn't echo log lines.
func writeChunk(o *clusteropts.ClusterOptions, body []byte) {
	if len(body) == 0 {
		return
	}
	_ = o.WriteStdout(body)
	if !o.Quiet && !strings.HasSuffix(string(body), "\n") {
		fmt.Fprintln(os.Stdout)
	}
}

// sleepCtx is the cancellation-aware sleep used by RunLogs's polling
// loop. Same shape as cli/cmd/ctl/market/watch.go::sleepOrCancel —
// repeated here rather than imported because that helper is private
// to package market. Trivial enough to duplicate; not worth lifting
// to a shared utility just for two callers.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

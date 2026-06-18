package pod

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterexec"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ExecParams is the shared input for `cluster pod exec` and the
// `cluster container exec` alias. Exported so Task 5's
// `cluster container exec` verb can delegate to RunExec without
// re-deriving the flag set.
type ExecParams struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
	Stdin     bool
	TTY       bool
	AssumeYes bool
	Timeout   time.Duration
	MaxBytes  int
}

// execJSON is the one-shot `-o json` result shape. ExitCode is a pointer so a
// timeout serializes as null.
type execJSON struct {
	Namespace  string   `json:"namespace"`
	Pod        string   `json:"pod"`
	Container  string   `json:"container"`
	Command    []string `json:"command"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	ExitCode   *int     `json:"exitCode"`
	Truncated  bool     `json:"truncated"`
	DurationMs int64    `json:"durationMs"`
}

// NewExecCommand: `olares-cli cluster pod exec <ns/pod | pod> [-c C]
// [-it] -- CMD [args...]`.
//
// One-shot (default) is the AI-friendly path: separated stdout/stderr,
// exit-code propagation, bounded by --timeout and --max-output-bytes.
// `-it` is the human path: TTY + terminal attach with a y/N confirm.
func NewExecCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		stdinFlag bool
		ttyFlag   bool
		assumeYes bool
		timeout   time.Duration
		maxBytes  int
	)
	cmd := &cobra.Command{
		Use:   "exec <ns/pod | pod> [-c CONTAINER] [-it] -- CMD [args...]",
		Short: "run a command inside a container (one-shot; -it for an interactive shell)",
		Long: `Run a command inside a container.

One-shot (default): everything after ` + "`--`" + ` is the argv run in the
container (no implicit shell). stdout/stderr are captured separately and the
container's exit code becomes this command's exit code. Bounded by --timeout
and --max-output-bytes so a hung/chatty command can't stall or flood callers.
Use ` + "`-- sh -c '...'`" + ` for pipes/redirects or multi-step repairs.

Interactive (-i -t / -it): allocate a TTY and attach your terminal, like
` + "`kubectl exec -it`" + `. Requires a local terminal and prompts for
confirmation (--yes skips). Default command is ` + "`sh`" + ` when none given.

NOTE: changes made inside a running container are ephemeral — a pod restart
reverts them. Durable fixes go through the image / ConfigMap / workload spec
(see ` + "`cluster workload`" + `).
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			if dash == -1 {
				if len(args) != 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--` (e.g. exec mypod -- ls)", args[1:])
				}
				target = args[0]
			} else {
				if dash < 1 {
					return fmt.Errorf("missing <pod> before `--`")
				}
				target = args[0]
				command = args[dash:]
			}
			ns, podName, err := clusteropts.SplitNsName(namespace, target)
			if err != nil {
				return err
			}
			return RunExec(c.Context(), o, ExecParams{
				Namespace: ns, Pod: podName, Container: container,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				AssumeYes: assumeYes, Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional is a bare pod name)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required for multi-container pods)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container (interactive -it only)")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt for interactive (-it) exec")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

// RunExec is the exported entry point for the exec verb, shared with
// Task 5's `cluster container exec` alias. It resolves the container
// (auto-selecting the sole container, or erroring with the candidate
// list for multi-container pods), then dispatches to the one-shot or
// interactive transport.
func RunExec(ctx context.Context, o *clusteropts.ClusterOptions, p ExecParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	f := o.Factory()

	container := strings.TrimSpace(p.Container)
	if container == "" {
		pod, err := Get(ctx, o, p.Namespace, p.Pod)
		if err != nil {
			return err
		}
		container, err = pickContainer(pod)
		if err != nil {
			return err
		}
	}

	token, err := f.ValidAccessToken(ctx)
	if err != nil {
		return err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	opts := clusterexec.Options{
		Namespace: p.Namespace, Pod: p.Pod, Container: container,
		Command: p.Command, Stdin: p.Stdin, TTY: p.TTY,
	}

	if p.TTY {
		if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("-t/--tty requires an interactive terminal; for non-interactive use run one-shot (drop -it and pass `-- CMD`)")
		}
		if len(opts.Command) == 0 {
			opts.Command = []string{"sh"}
		}
		if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Open an interactive shell in %s/%s [container %s]?", p.Namespace, p.Pod, container),
			p.AssumeYes); err != nil {
			return err
		}
		exit, rerr := clusterexec.RunInteractive(ctx, rp, token, opts, os.Stdin, os.Stdout)
		if rerr != nil {
			return rerr
		}
		if exit != nil && *exit != 0 {
			os.Exit(*exit)
		}
		return nil
	}

	if p.Stdin {
		return fmt.Errorf("-i/--stdin is only supported with -t/--tty (interactive); for one-shot file writes use `-- sh -c \"sed -i ...\"`, `printf ... | tee`, or a heredoc `-- sh -c 'cat > /path <<EOF ... EOF'`")
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("no command given; pass `-- CMD [args...]` (or use -it for an interactive shell)")
	}
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}
	start := time.Now()
	res, rerr := clusterexec.RunOneShot(ctx, rp, token, opts, p.MaxBytes)
	dur := time.Since(start)
	if rerr != nil {
		if errors.Is(rerr, context.DeadlineExceeded) {
			renderOneShot(o, p, container, res, nil, dur, true)
			return fmt.Errorf("command timed out after %s", p.Timeout)
		}
		return rerr
	}
	renderOneShot(o, p, container, res, res.ExitCode, dur, false)
	if res.ExitCode != nil && *res.ExitCode != 0 {
		os.Exit(*res.ExitCode)
	}
	return nil
}

func renderOneShot(o *clusteropts.ClusterOptions, p ExecParams, container string, res clusterexec.Result, exit *int, dur time.Duration, timedOut bool) {
	if o.IsJSON() {
		_ = o.PrintJSON(execJSON{
			Namespace: p.Namespace, Pod: p.Pod, Container: container,
			Command: p.Command, Stdout: string(res.Stdout), Stderr: string(res.Stderr),
			ExitCode: exit, Truncated: res.Truncated, DurationMs: dur.Milliseconds(),
		})
		return
	}
	if o.Quiet {
		return
	}
	_, _ = os.Stdout.Write(res.Stdout)
	_, _ = os.Stderr.Write(res.Stderr)
	if res.Truncated {
		fmt.Fprintln(os.Stderr, "[output truncated: --max-output-bytes reached]")
	}
	if timedOut {
		fmt.Fprintln(os.Stderr, "[timed out]")
	}
}

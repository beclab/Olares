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
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/picker"
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
// `-it` is the human path: TTY allocation + terminal attach (no prompt; the
// TTY requirement itself keeps non-terminal AI callers on the one-shot path).
func NewExecCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		stdinFlag bool
		ttyFlag   bool
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
` + "`kubectl exec -it`" + `. Requires a local terminal (a non-TTY caller such
as an AI tool call is refused with guidance to use one-shot instead). Default
command is ` + "`sh`" + ` when none given.

With -it and NO target, an interactive picker lists every container visible to
your profile (type to filter, arrows to move, enter to select). Add -n <ns> to
scope the picker to one namespace.

NOTE: changes made inside a running container are ephemeral — a pod restart
reverts them. Durable fixes go through the image / ConfigMap / workload spec
(see ` + "`cluster workload`" + `).
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			switch {
			case dash == -1:
				if len(args) > 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--` (e.g. exec mypod -- ls)", args[1:])
				}
				if len(args) == 1 {
					target = args[0]
				}
			default:
				if dash > 1 {
					return fmt.Errorf("unexpected args before `--`: %q", args[1:dash])
				}
				if dash == 1 {
					target = args[0]
				}
				command = args[dash:]
			}

			// No target + -it → interactive picker. Without -it we keep the
			// old "target required" contract (one-shot must be explicit).
			if target == "" {
				if !ttyFlag {
					return fmt.Errorf("missing <pod>; give a target (e.g. exec ns/pod -- ls) or add -it to pick a container interactively")
				}
				// Preflight the backend version before showing the picker so an
				// unsupported backend fails fast with the upgrade hint instead
				// of popping the picker only to reject the selection in RunExec.
				if err := requireExecBackendVersion(c.Context(), o.Factory()); err != nil {
					return err
				}
				ns, podName, ctr, canceled, perr := PickInteractiveTarget(c.Context(), o, namespace)
				if perr != nil {
					return perr
				}
				if canceled {
					return nil
				}
				return RunExec(c.Context(), o, ExecParams{
					Namespace: ns, Pod: podName, Container: ctr,
					Command: command, Stdin: stdinFlag, TTY: ttyFlag,
					Timeout: timeout, MaxBytes: maxBytes,
				})
			}

			ns, podName, err := clusteropts.SplitNsName(namespace, target)
			if err != nil {
				return err
			}
			return RunExec(c.Context(), o, ExecParams{
				Namespace: ns, Pod: podName, Container: container,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional is a bare pod name)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required for multi-container pods)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container (interactive -it only)")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

// execMinOlaresVersion is the first Olares OS line whose ControlHub edge nginx
// exposes the exec WebSocket route (`/api/v1/namespaces/.../pods/.../exec`).
// Earlier backends close the upgrade, so `cluster {pod,container} exec` is
// gated on this version.
const execMinOlaresVersion = "1.12.7"

// requireExecBackendVersion is the client-side version preflight for exec. It
// mirrors the files feature gates: >= 1.12.7 allowed; a detected-but-older
// backend is rejected with an upgrade hint; an undetectable version is rejected
// with the shared profile-refresh hint. The backend version is cached per
// profile at login, so in the common case this adds no network round-trip.
func requireExecBackendVersion(ctx context.Context, f *cmdutil.Factory) error {
	ok, err := f.OlaresBackendAtLeast(ctx, execMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"`cluster exec` requires Olares >= %s (the ControlHub exec route was added then), but the backend "+
				"version could not be determined: %v",
			execMinOlaresVersion, err)
	}
	if !ok {
		got := "unknown"
		if v, verr := f.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		return fmt.Errorf(
			"`cluster exec` requires Olares >= %s, but this backend is %s; upgrade the Olares system to use exec",
			execMinOlaresVersion, got)
	}
	return nil
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

	// Feature gate: the exec WebSocket route is only reachable once the system
	// ControlHub app's edge nginx allows/upgrades `/api/v1/.../exec`, which
	// shipped in Olares 1.12.7. On older backends the dial fails with an opaque
	// handshake error, so we fail fast with an actionable version message.
	if err := requireExecBackendVersion(ctx, f); err != nil {
		return err
	}

	// Permission gate: mirror the ControlHub SPA's per-namespace exec rule
	// client-side so the main account can't open a shell in a sub-account's
	// container (and non-admins stay confined to their own namespaces). This
	// matches the SPA hiding the Terminal button via hasPermission(namespace);
	// viewing/listing is untouched. Fails fast before any Get/dial.
	if err := gateExecPermission(ctx, o, p.Namespace); err != nil {
		return err
	}

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
		// Inject a prompt that names the target on every line so the user can
		// always tell which pod/container this shell is attached to — even deep
		// in scrollback (the kubelet doesn't expose the container name inside the
		// container; only the pod name shows up as the hostname). This fires both
		// when no command was given (default shell) and when the user explicitly
		// asked for a bare login shell like `-- sh` / `-- bash`, since that's the
		// common interactive case.
		label := fmt.Sprintf("%s/%s/%s", p.Namespace, p.Pod, container)
		if len(opts.Command) == 0 {
			opts.Command = defaultInteractiveShell(label)
		} else {
			opts.Command = injectPromptLabel(opts.Command, label)
		}
		// No confirmation here: opening an interactive shell isn't itself
		// destructive, and the TTY requirement above already keeps non-terminal
		// callers (AI tool calls) off this path — matching `kubectl/docker exec
		// -it`, which don't prompt either.
		//
		// Show a "Connecting ..." status before the (potentially slow) WebSocket
		// handshake so the user isn't left staring at a blank screen, then
		// replace it with "Connected ..." once RunInteractive reports the dial
		// succeeded. On a TTY the status is an animated spinner (same as the
		// interactive picker's loading state); on non-TTY it's a plain line.
		target := fmt.Sprintf("%s/%s [container %s]", p.Namespace, p.Pod, container)
		stderrTTY := term.IsTerminal(int(os.Stderr.Fd()))
		var sp *picker.Spinner
		if stderrTTY {
			sp = picker.StartSpinner(func() string { return "Connecting to " + target + "\u2026" })
		} else {
			fmt.Fprintf(os.Stderr, "Connecting to %s ...\n", target)
		}
		stopSpinner := func() {
			if sp != nil {
				sp.Stop() // clears the spinner line and blocks until stopped
				sp = nil
			}
		}
		onConnected := func() {
			stopSpinner()
			if stderrTTY {
				fmt.Fprintf(os.Stderr, "\033[1;32mConnected to %s\033[0m  \033[2m(Ctrl-D to exit)\033[0m\n", target)
			} else {
				fmt.Fprintf(os.Stderr, "Connected to %s  (Ctrl-D to exit)\n", target)
			}
		}
		exit, rerr := clusterexec.RunInteractive(ctx, rp, token, opts, os.Stdin, os.Stdout, onConnected)
		stopSpinner() // no-op if onConnected already stopped it; clears the line on dial failure
		if rerr != nil {
			return rerr
		}
		// ssh-style farewell so the user always knows the session ended (a clean
		// Ctrl-D exit and a mid-session drop otherwise look identical — you just
		// land back at the local prompt).
		if stderrTTY {
			fmt.Fprintf(os.Stderr, "\033[2mConnection to %s closed.\033[0m\n", target)
		} else {
			fmt.Fprintf(os.Stderr, "Connection to %s closed.\n", target)
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
		// The socket closed before the exit-status frame: render whatever
		// partial output was captured (with a null exit code), then fail so the
		// run isn't mistaken for a clean success.
		if errors.Is(rerr, clusterexec.ErrClosedBeforeExit) {
			renderOneShot(o, p, container, res, nil, dur, false)
			return rerr
		}
		return rerr
	}
	renderOneShot(o, p, container, res, res.ExitCode, dur, false)
	if res.ExitCode != nil && *res.ExitCode != 0 {
		os.Exit(*res.ExitCode)
	}
	return nil
}

// defaultInteractiveShell builds the command used when the user runs `-it`
// without an explicit `-- CMD`. It prefers `bash` (readline gives arrow-key
// history and line editing) and falls back to `sh` when the image ships no
// bash — many minimal images link /bin/sh to dash/busybox, which echoes
// `^[[A` on the arrow keys instead of recalling history. The PS1 export names
// the target and is inherited by whichever shell we finally exec.
func defaultInteractiveShell(label string) []string {
	return []string{"sh", "-c",
		fmt.Sprintf("export PS1='[%s] $PWD $ '; if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi", label)}
}

// injectPromptLabel wraps a bare interactive shell so its prompt always shows
// the target [ns/pod/container] plus the working directory. Only single-element
// commands whose basename is a known shell are wrapped; anything else (e.g.
// `-- python`, `-- bash -lc '...'`) is returned untouched. The label is literal
// text so it never depends on prompt-escape support; the cwd uses each shell's
// native mechanism ($PWD for the POSIX family, %~ for zsh). The exported prompt
// is inherited by the `exec`'d interactive shell.
func injectPromptLabel(cmd []string, label string) []string {
	if len(cmd) != 1 {
		return cmd
	}
	shell := cmd[0]
	base := shell
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	switch base {
	case "sh", "bash", "ash", "dash":
		return []string{shell, "-c",
			fmt.Sprintf("export PS1='[%s] $PWD $ '; exec %s", label, shell)}
	case "zsh":
		return []string{shell, "-c",
			fmt.Sprintf("export PROMPT='[%s] %%~ %%# '; exec %s", label, shell)}
	}
	return cmd
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

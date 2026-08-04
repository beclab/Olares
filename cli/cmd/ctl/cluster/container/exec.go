package container

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewExecCommand: `olares-cli cluster container exec
// <ns/pod/container | ns/pod | pod> [-n NS] [-c NAME] [-it] -- CMD [args...]`.
//
// Thin alias over `cluster pod exec` — same wire, same semantics. The only
// difference is the positional grammar: container may be supplied as the third
// path segment. Delegates to pod.RunExec.
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
		Use:   "exec <ns/pod/container | ns/pod | pod> [-c NAME] [-it] -- CMD [args...]",
		Short: "run a command inside a container (one-shot; -it for an interactive shell)",
		Long: `Run a command inside a container (alias of ` + "`cluster pod exec`" + `).

Identity grammar adds a three-segment positional <ns>/<pod>/<container>; the
two-segment <ns>/<pod> + --container and bare <pod> + -n/-c forms also work.

The container name is mandatory for this verb. (` + "`cluster pod exec`" + `
auto-selects the sole container of a single-container pod when --container
is omitted; the container alias asks you to identify it explicitly — via the
third path segment, ` + "`<ns>/<pod>`" + ` + --container, or bare ` + "`<pod>`" + ` + -n/-c.)

It shares the same execution semantics as ` + "`cluster pod exec`" + ` — one-shot
vs -it, --timeout, --max-output-bytes, and -o json all behave identically;
the only divergence is requiring an explicit container.

With -it and NO target, an interactive picker lists every container visible to
your profile (type to filter, arrows to move, enter to select). Add -n <ns> to
scope the picker to one namespace.
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			switch {
			case dash == -1:
				if len(args) > 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--` (e.g. exec ns/pod/ctr -- ls)", args[1:])
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

			// No target + -it → interactive picker (the picked container is
			// authoritative; -c is ignored in this path).
			if target == "" {
				if !ttyFlag {
					return fmt.Errorf("missing <pod>; give a target (e.g. exec ns/pod/ctr -- ls) or add -it to pick a container interactively")
				}
				ns, podName, ctr, canceled, perr := pod.PickInteractiveTarget(c.Context(), o, namespace)
				if perr != nil {
					return perr
				}
				if canceled {
					return nil
				}
				return pod.RunExec(c.Context(), o, pod.ExecParams{
					Namespace: ns, Pod: podName, Container: ctr,
					Command: command, Stdin: stdinFlag, TTY: ttyFlag,
					Timeout: timeout, MaxBytes: maxBytes,
				})
			}

			ns, podName, ctr, err := splitNsPodContainer(namespace, container, target)
			if err != nil {
				return err
			}
			return pod.RunExec(c.Context(), o, pod.ExecParams{
				Namespace: ns, Pod: podName, Container: ctr,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional doesn't include one)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required when the positional doesn't include one)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container (interactive -it only)")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

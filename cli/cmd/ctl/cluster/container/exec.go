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
		assumeYes bool
		timeout   time.Duration
		maxBytes  int
	)
	cmd := &cobra.Command{
		Use:   "exec <ns/pod/container | ns/pod | pod> [-c NAME] [-it] -- CMD [args...]",
		Short: "run a command inside a container (one-shot; -it for an interactive shell)",
		Long: `Run a command inside a container (alias of ` + "`cluster pod exec`" + `).

Identity grammar adds a three-segment positional <ns>/<pod>/<container>; the
two-segment <ns>/<pod> + --container and bare <pod> + -n/-c forms also work.
Everything else (one-shot vs -it, --timeout, --max-output-bytes, -o json) is
identical to ` + "`cluster pod exec`" + `; this verb just delegates.
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			if dash == -1 {
				if len(args) != 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--`", args[1:])
				}
				target = args[0]
			} else {
				if dash < 1 {
					return fmt.Errorf("missing <pod> before `--`")
				}
				target = args[0]
				command = args[dash:]
			}
			ns, podName, ctr, err := splitNsPodContainer(namespace, container, target)
			if err != nil {
				return err
			}
			return pod.RunExec(c.Context(), o, pod.ExecParams{
				Namespace: ns, Pod: podName, Container: ctr,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				AssumeYes: assumeYes, Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional doesn't include one)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required when the positional doesn't include one)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container (interactive -it only)")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt for interactive (-it) exec")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

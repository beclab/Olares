package workload

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewStopCommand: `olares-cli cluster workload stop <ns/name | name>
// --kind X [-n NS] [--watch] [--yes]`.
//
// Thin alias for `cluster workload scale --replicas=0 --yes`.
// Justified because the SPA exposes a labeled "STOP" button on the
// workload detail page (apps/.../Workloads/Detail.vue:208-209) — CLI
// users will look for the verb by name, not infer "scale to 0".
//
// DaemonSets are rejected the same way `workload scale` rejects them
// (no replicas concept). To "stop" a DaemonSet you'd delete the
// workload entirely (`cluster workload delete`) since DaemonSet has
// no scale-to-zero idiom.
func NewStopCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
		watch     bool
		interval  time.Duration
		timeout   time.Duration
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "stop <ns/name | name>",
		Short: "stop a Deployment / StatefulSet (alias for `scale --replicas=0`)",
		Long: `Stop a Deployment or StatefulSet by scaling its replicas to 0.

Equivalent to ` + "`cluster workload scale ... --replicas=0`" + ` and
mirrors the SPA's "STOP" button. ConfirmDestructive prompt is shown
unless --yes is passed.

DaemonSets cannot be "stopped" via this verb (no replicas concept) —
delete the workload (` + "`cluster workload delete`" + `) if you really
want to take it offline.

--watch chains into rollout-status convergence so you get
"stopped and confirmed terminated" in one command.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			plural, err := NormalizeKind(kindRaw)
			if err != nil {
				return err
			}
			if plural == KindAll {
				return fmt.Errorf("--kind must be one of: deployment, statefulset (not %q)", kindRaw)
			}
			if plural == "daemonsets" {
				return fmt.Errorf("DaemonSets cannot be stopped via scale; use `cluster workload delete` to remove them")
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return RunScale(c.Context(), o, ns, name, plural, 0, watch, interval, timeout, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset (REQUIRED)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "after stopping, wait for the rollout to converge (replicas=0)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	o.AddOutputFlags(cmd)
	return cmd
}

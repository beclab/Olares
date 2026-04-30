package workload

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewStartCommand: `olares-cli cluster workload start <ns/name | name>
// --kind X --replicas N [-n NS] [--watch]`.
//
// Thin alias for `cluster workload scale --replicas=N`. Justified for
// the same reason as `stop` — the SPA pairs them ("STOP" / "START")
// and CLI users will look for the verb by name.
//
// --replicas is REQUIRED here (we don't have a "previous replicas
// count" cached anywhere; the SPA infers it from spec but for the CLI
// it's safer to make the operator name the target). No
// ConfirmDestructive — starting a stopped workload is non-destructive.
//
// DaemonSets are rejected the same way `workload scale` rejects them.
func NewStartCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
		replicas  int
		watch     bool
		interval  time.Duration
		timeout   time.Duration
	)
	cmd := &cobra.Command{
		Use:   "start <ns/name | name>",
		Short: "start a stopped Deployment / StatefulSet (alias for `scale --replicas=N`)",
		Long: `Start a stopped Deployment or StatefulSet by scaling its replicas
to --replicas N (must be >= 1).

Equivalent to ` + "`cluster workload scale ... --replicas=N`" + ` and
mirrors the SPA's "START" button. No confirmation prompt because
starting a stopped workload is non-destructive.

DaemonSets cannot be "started" via this verb (no replicas concept) —
re-create the DaemonSet object via your normal apply pipeline if
needed.

--watch chains into rollout-status convergence so you get
"started and Ready" in one command.
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
				return fmt.Errorf("DaemonSets cannot be started via scale; recreate the object instead")
			}
			if replicas < 1 {
				return fmt.Errorf("--replicas must be >= 1 for `start`, got %d (use `cluster workload scale --replicas=0` to scale to zero)", replicas)
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			// assumeYes=true at the call site — start is non-
			// destructive, no prompt needed.
			return RunScale(c.Context(), o, ns, name, plural, replicas, watch, interval, timeout, true)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset (REQUIRED)")
	cmd.Flags().IntVar(&replicas, "replicas", 1, "desired replica count (REQUIRED; >= 1)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "after starting, wait for the rollout to converge")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	o.AddOutputFlags(cmd)
	return cmd
}

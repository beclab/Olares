package pod

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRestartCommand: `olares-cli cluster pod restart <ns/name | name>
// [-n NS] [--yes] [--grace-period N]`.
//
// Wire-identical to `cluster pod delete` — the SPA's restartPods is
// bit-identical to deletePod (apps/.../controlPanelCommon/network/
// index.ts, both call DELETE on the same URL). Offered as a separate
// verb because operators reach for "restart" by name when they want
// the controller to recreate a pod, and we want the CLI's verb shelf
// to match the SPA's button shelf.
//
// Implementation reuses pod.RunDelete with opName="restart" so the
// JSON output and confirmation prompt say "restart" while the wire
// call stays the canonical DELETE.
func NewRestartCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace   string
		assumeYes   bool
		gracePeriod int
	)
	cmd := &cobra.Command{
		Use:   "restart <ns/name | name>",
		Short: "restart one Pod (alias for `delete` — controller recreates the pod)",
		Long: `Restart one Pod by deleting it; the owning controller (Deployment /
StatefulSet / DaemonSet / Job / ReplicaSet) recreates it.

Wire-identical to ` + "`cluster pod delete`" + ` — the SPA's restartPods is
bit-identical to deletePod, both ` + "`DELETE /api/v1/namespaces/<ns>/pods/<name>`" + `.
The verb name is the only difference.

For standalone (controller-less) pods this is effectively a delete
(no recreation). Pass --yes to skip the confirmation prompt for
scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return RunDelete(c.Context(), o, ns, name, assumeYes, gracePeriod, "restart")
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().IntVar(&gracePeriod, "grace-period", -1, "graceful termination period in seconds (-1 = pod default; 0 = immediate)")
	o.AddOutputFlags(cmd)
	return cmd
}

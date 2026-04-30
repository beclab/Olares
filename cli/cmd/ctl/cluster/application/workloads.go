package application

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewWorkloadsCommand: `olares-cli cluster application workloads
// <namespace> [--kind ...] [-l ...] [--limit ...] [-o ...]`.
//
// Convenience alias for `cluster workload list -n <namespace> ...`.
// Most user workflows arrive here from `cluster application list`
// then "show me what's running in this space"; offering it as a
// dedicated verb keeps that pivot a one-shot command rather than
// requiring two flag flips.
//
// All filtering happens server-side via workload.RunList — there is
// no client-side namespace inference or scope expansion here (the
// security model is server-decides; see olares-cluster SKILL.md).
func NewWorkloadsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		kindRaw       string
		labelSelector string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "workloads <namespace>",
		Short: "list workloads inside one application space (alias for `cluster workload list -n <ns>`)",
		Long: `List Deployment / StatefulSet / DaemonSet workloads inside one
application space (Namespace).

Equivalent to ` + "`cluster workload list -n <namespace>`" + ` — the verb
just makes the application-side pivot from ` + "`application list`" + `
explicit. --kind defaults to "all"; pass deployment / statefulset /
daemonset (singular or plural; "deploy" / "sts" / "ds" also accepted)
to scope to one kind. --label / --limit are forwarded verbatim.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns := strings.TrimSpace(args[0])
			if ns == "" {
				return fmt.Errorf("namespace must be non-empty")
			}
			return workload.RunList(c.Context(), o, ns, kindRaw, labelSelector, limit)
		},
	}
	cmd.Flags().StringVar(&kindRaw, "kind", workload.KindAll, "workload kind: all | deployment | statefulset | daemonset")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter workloads (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items per kind to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

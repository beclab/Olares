package application

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewPodsCommand: `olares-cli cluster application pods <namespace>
// [-l ...] [--field-selector ...] [--limit ...] [--page ...] [--all] [-o ...]`.
//
// Convenience alias for `cluster pod list -n <namespace> ...`,
// symmetric with `cluster application workloads`. Same server-side
// scoping rules apply, same pagination semantics
// (--limit / --page / --all all forwarded verbatim).
func NewPodsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		labelSelector string
		fieldSelector string
	)
	cmd := &cobra.Command{
		Use:   "pods <namespace>",
		Short: "list pods inside one application space (alias for `cluster pod list -n <ns>`)",
		Long: `List pods inside one application space (Namespace).

Equivalent to ` + "`cluster pod list -n <namespace>`" + ` — the verb just
makes the application-side pivot from ` + "`application list`" + ` explicit.
--label / --field-selector / --limit / --page / --all are forwarded verbatim.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns := strings.TrimSpace(args[0])
			if ns == "" {
				return fmt.Errorf("namespace must be non-empty")
			}
			// Mirror `cluster pod list -n <ns>`'s empty-result UX so
			// the application-side pivot doesn't silently show an
			// empty header table — see pod.RunList for why count is
			// returned instead of folding the message into RunList
			// itself.
			n, err := pod.RunList(c.Context(), o, p, ns, labelSelector, fieldSelector)
			if err != nil {
				return err
			}
			if n == 0 && !o.IsJSON() && !o.Quiet {
				fmt.Fprintf(os.Stderr, "No pods found in %s namespace.\n", ns)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter pods (K8s syntax)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "field selector to filter pods (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

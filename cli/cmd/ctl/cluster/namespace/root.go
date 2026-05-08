// Package namespace implements `olares-cli cluster namespace ...` —
// raw K8s Namespace inspection visible to the active profile.
//
// Note: this is the K8s framing of the same resource the
// `cluster application` tree exposes from the Olares ApplicationSpace
// angle. Both ultimately hit `/api/v1/namespaces/<ns>` for detail;
// the distinction is in how lists are rendered and what extra
// metadata is surfaced. Use `cluster namespace ...` when you want
// the kubectl-style flat list (PHASE / AGE), `cluster application
// ...` when you want the SPA's workspace grouping.
package namespace

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewNamespaceCommand assembles `olares-cli cluster namespace`.
// Today's verbs are `list` (KubeSphere paginated) and `get` (K8s
// native).
func NewNamespaceCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"namespaces", "ns"},
		Short:   "Inspect K8s Namespaces visible to the active profile",
		Long: `Inspect Kubernetes Namespaces visible to the active profile.

Endpoints (all under https://control-hub.<terminus>):
  list  /kapis/resources.kubesphere.io/v1alpha3/namespaces
  get   /api/v1/namespaces/<ns>

For the Olares ApplicationSpace framing (workspace-grouped),
see ` + "`cluster application list`" + `.
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))

	return cmd
}

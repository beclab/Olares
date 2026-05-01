// Package node implements `olares-cli cluster node ...` —
// cluster node inspection visible to the active profile.
//
// Like every other verb in the cluster tree, server-side scoping
// decides what's visible (e.g. a non-admin profile may receive an
// empty list or 403 when listing nodes). CLI never gates locally;
// see olares-cluster SKILL.md for the security rationale.
//
// This is the per-user K8s view of nodes — for host-side node
// install / join / drain operations see `olares-cli node` (the
// kubeconfig-based tree).
package node

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewNodeCommand assembles `olares-cli cluster node`. Today's verbs
// are read-only `list` + `get`. Mutating verbs (cordon / drain) live
// outside this tree (they require kubeconfig).
func NewNodeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes"},
		Short:   "Inspect K8s nodes visible to the active profile (per-user view)",
		Long: `Inspect Kubernetes nodes visible to the active profile.

Endpoints (all under https://control-hub.<terminus>):
  list   /kapis/resources.kubesphere.io/v1alpha3/nodes
  get    /kapis/resources.kubesphere.io/v1alpha3/nodes/<node>

For host-side node operations (install / join / drain) see
"olares-cli node" — that's a different tree that uses kubeconfig.
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

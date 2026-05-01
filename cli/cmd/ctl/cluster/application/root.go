// Package application implements `olares-cli cluster application ...`.
//
// "Application" here refers to the Olares ApplicationSpace
// abstraction over K8s Namespaces — the same surface the ControlHub
// SPA exposes under "Application Spaces" (see
// apps/packages/app/src/apps/controlHub/pages/ApplicationSpaces).
// Each visible application space is a Namespace; the SPA groups them
// by KubeSphere workspace to surface the user-vs-system distinction
// the cluster's RBAC enforces.
//
// Boundary note: this tree is the runtime-state view. App-store
// lifecycle (install / uninstall / start / stop / upgrade) belongs
// to `olares-cli market`, which goes through the market service and
// has its own typed client (cmd/ctl/market). The two trees are
// complementary — `cluster application list` shows what's actually
// running on the cluster you can see, `market list` shows what's
// installable / installed from the user's app perspective.
//
// Phase 1d (initial slice) ships only `list`. Later phases bring
// `get` (per-application detail) and `workloads` / `pods` (drill-
// down into a specific application's K8s objects).
package application

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewApplicationCommand assembles `olares-cli cluster application`.
func NewApplicationCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "List Olares application spaces (Namespaces grouped by workspace)",
		Long: `List Olares application spaces — Namespaces visible to the active
profile, grouped by their KubeSphere workspace.

Backed by https://control-hub.<terminus>/capi/namespaces/group, the
same endpoint the ControlHub SPA uses for its "Application Spaces"
sidebar. The server returns groups already scoped to what the active
token can see; CLI does no client-side filtering.

For app-store lifecycle (install / uninstall / start / stop / upgrade)
see "olares-cli market"; this tree is the runtime-state view of the
resulting K8s namespaces.
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewWorkloadsCommand(f))
	cmd.AddCommand(NewPodsCommand(f))
	cmd.AddCommand(NewStatusCommand(f))

	return cmd
}

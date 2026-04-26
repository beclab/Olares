// Package gpu implements the `olares-cli settings gpu` subtree (Settings ->
// GPU). Backed by user-service's gpu.controller.ts. There's also a top-level
// `olares-cli gpu` tree from earlier work that talks via kubeconfig — this
// one is the profile-based settings surface that mirrors the Settings UI.
package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGPUCommand returns the `settings gpu` parent.
func NewGPUCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gpu",
		Short: "GPU mode + per-app GPU assignment (Settings -> GPU)",
		Long: `Inspect GPU device list, mode, and per-app assignment.

Subcommands will be added in subsequent phases:
  Phase 1: list
  Phase 2: mode, assign, unassign, bulk-assign

Note: this differs from the top-level "olares-cli gpu" command — that one
talks to the cluster via kubeconfig. "settings gpu" is the profile-based
edge-API surface that mirrors the SPA's Settings -> GPU page.
`,
	}
	cmd.SilenceUsage = true
	return cmd
}

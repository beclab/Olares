package compute

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewComputeCommand returns the `settings compute` parent — the new-version
// "Accelerator" surface (compute-resources). Version-gated to
// Olares >= 1.12.6; on older backends use `settings gpu list` instead.
//
// Subcommands are flat verbs that mirror the SPA's Accelerator pages:
//
//	list                            inspect nodes / devices (DEVICE-ID) / bindings
//	unbind <app>                    unbind (and stop) an app's compute
//	set-type <node> <device>        switch a device's support type
func NewComputeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "AI compute (accelerator) resources — Settings -> Accelerator (Olares >= 1.12.6)",
		Long: `Inspect and manage AI compute (accelerator) resources — the same surface the
desktop SPA exposes under Settings -> Accelerator.

Backed by the compute-resources APIs introduced in Olares 1.12.6; on a 1.12.5
backend these routes do not exist — use the legacy "olares-cli settings gpu
list" instead.

Subcommands:
  list                            list nodes, devices (with DEVICE-ID) and bindings
  unbind <app>                    unbind an app from its device(s) (stops the app)
  set-type <node> <device> ...    switch a device's support type

Typical flow: run "list" first; its per-node header gives <node> and the
DEVICE-ID column gives <device> for "set-type".

Note: this differs from the top-level "olares-cli gpu" command (kubeconfig).
"settings compute" is the profile-based edge-API surface that mirrors the SPA.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(
		newListCommand(f),
		newUnbindCommand(f),
		newSetTypeCommand(f),
	)
	return cmd
}

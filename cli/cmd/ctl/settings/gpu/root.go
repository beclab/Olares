package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGPUCommand returns the `settings gpu` parent (Settings -> Compute
// Acceleration). Its subcommand surface and help text adapt to the LOCALLY
// CACHED backend version (no network call at construction time, so `--help`
// and shell completion stay fast and offline):
//
//   - Olares 1.12.6+ (or unknown cache): list + bindings + unbind + support-type
//     (the compute-resources model).
//   - Olares 1.12.5: only list (the legacy /api/gpu/list device view); the
//     1.12.6-only verbs are hidden and the help notes the requirement.
//
// Visibility is advisory: runtime dispatch (WithOlaresClient) re-detects the
// version and the capability gate returns a clear "requires Olares >= 1.12.6"
// error if a hidden verb is invoked against an older backend anyway, so a
// stale cache never produces a wrong action — only stale help.
func NewGPUCommand(f *cmdutil.Factory) *cobra.Command {
	newSurface := true
	if v, ok := f.CachedOlaresBackendVersion(); ok {
		newSurface = usesComputeResources(v)
	}

	cmd := &cobra.Command{
		Use:   "gpu",
		Short: "Compute acceleration: GPU devices + per-app bindings (Settings -> GPU)",
		Long:  gpuLong(newSurface),
	}
	cmd.SilenceUsage = true

	cmd.AddCommand(NewListCommand(f))

	bindings := NewBindingsCommand(f)
	unbind := NewUnbindCommand(f)
	supportType := NewSupportTypeCommand(f)
	if !newSurface {
		// Legacy backend: keep these reachable (the runtime gate gives a
		// precise error) but hide them from help/completion so the
		// advertised surface matches what actually works.
		bindings.Hidden = true
		unbind.Hidden = true
		supportType.Hidden = true
	}
	cmd.AddCommand(bindings)
	cmd.AddCommand(unbind)
	cmd.AddCommand(supportType)

	return cmd
}

func gpuLong(newSurface bool) string {
	if newSurface {
		return `Inspect and manage compute-acceleration resources (Olares 1.12.6+).

Subcommands:
  list                       list nodes / GPU devices and their support type + bindings
  bindings <app>             show an app's compute (GPU) bindings
  unbind <app>               release an app's compute bindings (suspends the app)
  support-type set <node> <device> <type>
                             switch a device's support type
                             (Exclusive | MemorySlice | TimeSlice | MemoryShared)

Note: this differs from the top-level "olares-cli gpu" command, which talks to
the cluster via kubeconfig. "settings gpu" is the profile-based edge-API surface
that mirrors the SPA's Settings -> GPU page.`
	}
	return `Inspect GPU devices (Olares 1.12.5).

Subcommands:
  list   list GPU devices, mode, and per-app assignment

The compute-resources verbs (bindings, unbind, support-type) require Olares
1.12.6 or newer and are hidden on this backend. Upgrade Olares to manage
per-app GPU bindings and device support types from the CLI.

Note: this differs from the top-level "olares-cli gpu" command, which talks to
the cluster via kubeconfig.`
}

// Package container implements `olares-cli cluster container ...` —
// per-container drill-down inside a Pod.
//
// "container" is a virtual noun in this CLI: a container is always
// scoped to a Pod, and every fetch goes through the same
// `/api/v1/namespaces/<ns>/pods/<name>` endpoint that
// `cluster pod get` uses. We re-project the response into per-
// container views (one row per spec.containers[*]; one section per
// container's env block) so users don't have to JSONPath their way
// into the pod object to answer "what env vars does the X container
// of pod Y have?".
//
// No new HTTP surface is added here — it's all rendering on top of
// pod.Get. That means container verbs inherit the same server-side
// scoping (a 404 from pod.Get propagates as-is) and the same auth
// recovery story as `cluster pod ...`.
package container

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewContainerCommand assembles `olares-cli cluster container`.
// Today's verbs are the read-only Phase 1c slice (list / env) plus
// Phase 2's `container logs`, which routes per-container via
// /api/v1/namespaces/<ns>/pods/<name>/log?container=<c> (same
// endpoint the SPA uses for its container log viewer).
func NewContainerCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "container",
		Aliases: []string{"containers", "ctr"},
		Short:   "Inspect containers inside a pod (image / state / env)",
		Long: `Inspect containers inside a pod.

Both verbs take the parent pod as a "<namespace>/<pod>" positional or
"-n <ns>" + bare "<pod>"; "container env" additionally accepts an
optional --container <name> to scope to one container's env block
(default: print every container's env, grouped).

No new HTTP surface — every fetch reuses ` + "`cluster pod get`" + `'s
endpoint (/api/v1/namespaces/<ns>/pods/<name>).
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewEnvCommand(f))
	cmd.AddCommand(NewLogsCommand(f))

	return cmd
}

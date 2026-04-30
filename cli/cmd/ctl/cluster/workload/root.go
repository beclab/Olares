// Package workload implements `olares-cli cluster workload ...` —
// read-side workload (Deployment / StatefulSet / DaemonSet)
// inspection for the active user's profile.
//
// The KubeSphere paginated paths (`/kapis/resources.kubesphere.io/
// v1alpha3/<kind>` cross-ns and
// `/kapis/.../namespaces/<ns>/<kind>` per-ns) drive list verbs;
// per-resource detail uses the K8s native path
// (`/apis/apps/v1/namespaces/<ns>/<kind>/<name>`). Server-side
// scoping decides what's visible — CLI never filters or gates based
// on the cached cluster context (see olares-cluster SKILL.md).
package workload

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewWorkloadCommand assembles `olares-cli cluster workload`. Verbs
// are added incrementally; today's set is the read-only Phase 1b
// slice (list / get / yaml). Phase 4 will bring scale + restart
// using the same merge-patch+json plumbing the SPA's
// patchWorkloadsControler uses (see
// apps/packages/app/src/apps/controlPanelCommon/network/index.ts:372).
func NewWorkloadCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workload",
		Aliases: []string{"workloads", "wl"},
		Short:   "Inspect Deployments / StatefulSets / DaemonSets visible to the active profile",
		Long: `Inspect workloads on the Olares cluster from the active profile's
ControlHub view.

"workload" here means the K8s controller resources Deployment,
StatefulSet, and DaemonSet — the same set the ControlHub SPA exposes
under "Workloads" inside an Application Space. Verbs accept --kind
to scope to one of the three; the default for "list" is "all", which
fans out one request per kind and merges the results into a single
table with a KIND column.

Endpoints (all under https://control-hub.<terminus>):
  list      /kapis/resources.kubesphere.io/v1alpha3/<kind>
            /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/<kind>
  get/yaml  /apis/apps/v1/namespaces/<ns>/<kind>/<name>

By default list-style verbs do NOT pass a namespace to the server
(cross-ns mode), so the response is the union of every namespace the
active profile can see. Pass -n / --namespace to scope explicitly.
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewYAMLCommand(f))
	cmd.AddCommand(NewRolloutStatusCommand(f))
	cmd.AddCommand(NewScaleCommand(f))
	cmd.AddCommand(NewRestartCommand(f))
	cmd.AddCommand(NewStopCommand(f))
	cmd.AddCommand(NewStartCommand(f))

	return cmd
}

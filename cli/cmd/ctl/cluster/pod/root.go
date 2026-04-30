// Package pod implements `olares-cli cluster pod ...` — read-side
// pod inspection for the active user's profile.
//
// All verbs here go through pkg/clusterclient.Client (which talks to
// https://control-hub.<terminus>) and decode either KubeSphere
// {items, totalItems} envelopes (`/kapis/resources.kubesphere.io/v1alpha3/...`)
// or K8s-native shapes (`/api/v1/namespaces/<ns>/pods/<name>`,
// `/api/v1/namespaces/<ns>/events`). Per-user namespace scoping is
// enforced server-side; CLI verbs never filter or gate based on the
// locally cached cluster context.
package pod

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewPodCommand assembles `olares-cli cluster pod`. Verbs are added
// incrementally; today's set is the read-only Phase 1a slice (list /
// get / yaml / events). Phase 2 brings logs, Phase 3 brings --watch
// onto get, Phase 6 brings restart / delete.
func NewPodCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pod",
		Short: "Inspect Pods visible to the active profile (cross-namespace by default)",
		Long: `Inspect Pods on the Olares cluster from the active profile's
ControlHub view.

By default, list-style verbs do NOT pass a namespace to the server,
so the response is the union of every namespace the active profile
can see (the SPA does the same in its Pods page). Pass -n / --namespace
to scope explicitly. The server, not the CLI, decides what's visible
under either mode.

Endpoints (all under https://control-hub.<terminus>):
  list          /kapis/resources.kubesphere.io/v1alpha3/pods
                /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/pods
  get / yaml    /api/v1/namespaces/<ns>/pods/<name>
  events        /api/v1/namespaces/<ns>/events
                  (filtered to involvedObject.kind=Pod, name=<pod>)
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewYAMLCommand(f))
	cmd.AddCommand(NewEventsCommand(f))

	return cmd
}

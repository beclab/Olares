// Package job implements `olares-cli cluster job ...` — read- and
// write-side Job inspection (and the KubeSphere-flavored "rerun"
// action) for the active user's profile.
//
// The KubeSphere paginated path
// (`/kapis/resources.kubesphere.io/v1alpha3/jobs` cross-ns and
// `/kapis/.../namespaces/<ns>/jobs` per-ns) drives `job list`;
// per-resource detail uses the K8s native path
// (`/apis/batch/v1/namespaces/<ns>/jobs/<name>`). Server-side
// scoping decides what's visible — CLI never filters or gates based
// on the cached cluster context (see olares-cluster SKILL.md).
//
// Mutations:
//
//   - `job rerun` calls KubeSphere's operations API
//     (`/kapis/operations.kubesphere.io/v1alpha2/.../jobs/<name>?
//     action=rerun&resourceVersion=<rv>`) — NOT a client-side spec
//     clone. Same flow the SPA's JobsDetails.vue toolbar uses.
package job

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewJobCommand assembles `olares-cli cluster job`. Today's set is
// the Phase 5 slice (list / get / yaml / pods / events / rerun).
// CronJobs are a separate noun (`cluster cronjob`) since the SPA
// models them that way and the API versions differ
// (`apis/batch/v1` for jobs, `apis/batch/v1beta1` for cronjobs).
func NewJobCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Inspect K8s Jobs visible to the active profile (and rerun them)",
		Long: `Inspect K8s Jobs (apis/batch/v1) on the Olares cluster from the
active profile's ControlHub view.

Endpoints (all under https://control-hub.<terminus>):
  list           /kapis/resources.kubesphere.io/v1alpha3/jobs
                 /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/jobs
  get / yaml     /apis/batch/v1/namespaces/<ns>/jobs/<name>
  events         /api/v1/namespaces/<ns>/events
                   (filtered to involvedObject.kind=Job, name=<job>)
  pods           /kapis/.../namespaces/<ns>/pods?labelSelector=controller-uid=<uid>
                   (two-step: GET job → reuse "cluster pod list")
  rerun          POST /kapis/operations.kubesphere.io/v1alpha2/
                       namespaces/<ns>/jobs/<name>
                       ?action=rerun&resourceVersion=<rv>
                   (KubeSphere operations API, no body)

For scheduled jobs see "cluster cronjob".
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewYAMLCommand(f))
	cmd.AddCommand(NewPodsCommand(f))
	cmd.AddCommand(NewEventsCommand(f))
	cmd.AddCommand(NewRerunCommand(f))

	return cmd
}

// Package cronjob implements `olares-cli cluster cronjob ...` —
// CronJob inspection plus suspend/resume mutations.
//
// Modeled as a separate noun (rather than folding into `cluster job`)
// because:
//   - The SPA also models them separately (jobType enum in
//     apps/.../controlPanelCommon/network/network.ts has cronjobs +
//     jobs as siblings).
//   - Even though both live under apis/batch/v1 today, the verb sets
//     differ (only cronjobs have suspend/resume; only jobs have
//     rerun) and the SPA itself models them as siblings, so keeping
//     one package per noun matches the call shape users expect.
//     (Historical note: cronjobs used to live under apis/batch/v1beta1
//     pre-K8s 1.25; we no longer reference that path.)
//   - The verb sets differ: only cronjobs have suspend/resume; only
//     jobs have rerun. Splitting keeps each package's --help focused.
//
// Server-side scoping decides what's visible — CLI never filters or
// gates based on the cached cluster context (see olares-cluster
// SKILL.md).
package cronjob

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewCronJobCommand assembles `olares-cli cluster cronjob`. Today's
// set is the Phase 5 slice (list / get / yaml / jobs / suspend /
// resume).
func NewCronJobCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cronjob",
		Aliases: []string{"cronjobs", "cj"},
		Short:   "Inspect and suspend/resume K8s CronJobs visible to the active profile",
		Long: `Inspect K8s CronJobs (apis/batch/v1) on the Olares cluster from
the active profile's ControlHub view, and suspend/resume them.

Endpoints (all under https://control-hub.<terminus>):
  list           /kapis/resources.kubesphere.io/v1alpha3/cronjobs
                 /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/cronjobs
  get / yaml     /apis/batch/v1/namespaces/<ns>/cronjobs/<name>
  jobs           (1) GET cronjob for UID (+ jobTemplate labels if any);
                 (2) GET /apis/batch/v1/namespaces/<ns>/jobs
                       [?labelSelector=<derived>]  -- label is an
                       apiserver-side pre-narrow when present;
                 (3) client-side filter by ownerReferences
                       [uid==cronjob.uid, controller=true, kind=CronJob]
                       — the K8s-native source of truth for parent
                       /child binding.
  suspend        PATCH apis/batch/v1/.../cronjobs/<name>
                   body {"spec":{"suspend":true}}
                   Content-Type application/merge-patch+json
  resume         same path; body {"spec":{"suspend":false}}

For one-shot Jobs see "cluster job".
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewYAMLCommand(f))
	cmd.AddCommand(NewJobsCommand(f))
	cmd.AddCommand(NewSuspendCommand(f))
	cmd.AddCommand(NewResumeCommand(f))

	return cmd
}

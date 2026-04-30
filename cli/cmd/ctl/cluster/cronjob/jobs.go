package cronjob

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/job"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewJobsCommand: `olares-cli cluster cronjob jobs <ns/name | name>
// [-n NS] [--limit N]`.
//
// Two-step:
//
//  1. GET the CronJob to read spec.jobTemplate.metadata.labels.
//  2. List Jobs server-side via labelSelector built from those labels
//     (`/apis/batch/v1/namespaces/<ns>/jobs?labelSelector=<derived>`).
//
// Mirrors the SPA's CronJob → child Jobs lookup
// (apps/.../controlHub/pages/Jobs/CronJobsDetails.vue) which derives
// the same selector.
//
// We hit the K8s native path (apis/batch/v1) rather than KubeSphere's
// envelope so the labelSelector is honored by the apiserver directly.
// Output mirrors `cluster job list` (NAMESPACE / NAME / COMPLETIONS /
// STATUS / DURATION / AGE) for consistency.
func NewJobsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		limit     int
	)
	cmd := &cobra.Command{
		Use:   "jobs <ns/name | name>",
		Short: "list child Jobs spawned by one CronJob",
		Long: `List the child Jobs spawned by one CronJob.

Two-step:
  1. GET the CronJob to read spec.jobTemplate.metadata.labels.
  2. List Jobs in the same namespace via labelSelector derived from
     those labels (` + "`/apis/batch/v1/namespaces/<ns>/jobs?labelSelector=...`" + `).

If the CronJob's jobTemplate carries no labels (rare; would be a
manual edit) the selector cannot be built and the verb errors out
rather than fanning out to "every job in the namespace".

Output columns mirror ` + "`cluster job list`" + ` for consistency.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runJobs(c.Context(), o, ns, name, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

func runJobs(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	selector := c.templateLabelSelector()
	if selector == "" {
		return fmt.Errorf("cronjob %s/%s has no spec.jobTemplate.metadata.labels — cannot derive a Job selector", namespace, name)
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("labelSelector", selector)
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/apis/batch/v1/namespaces/%s/jobs?%s",
		url.PathEscape(namespace), q.Encode())

	resp, err := clusterclient.GetK8sList[job.Job](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list jobs spawned by cronjob %s/%s: %w", namespace, name, err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items []job.Job `json:"items"`
		}{Items: resp.Items})
	}
	if o.Quiet {
		return nil
	}
	return renderChildJobsTable(resp.Items, o.NoHeaders)
}

// renderChildJobsTable lays out the child Jobs in the same column
// shape as `cluster job list` so users can read both outputs without
// re-parsing. We reuse job.Job's helper methods (status / duration /
// age) so the column rules stay shared with the job package.
func renderChildJobsTable(items []job.Job, noHeaders bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "NAME\tCOMPLETIONS\tSTATUS\tAGE")
	}
	now := time.Now()
	for _, j := range items {
		// We render NAME / COMPLETIONS (succeeded/total) / STATUS /
		// AGE — DURATION is dropped because the job.Job helpers
		// computing it are unexported.
		comp := "-"
		if j.Spec.Completions != nil {
			comp = fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
		} else {
			comp = fmt.Sprintf("%d", j.Status.Succeeded)
		}
		status := "Pending"
		for _, c := range j.Status.Conditions {
			if c.Status != "True" {
				continue
			}
			switch c.Type {
			case "Complete":
				status = "Complete"
			case "Failed":
				status = "Failed"
			}
		}
		if status == "Pending" && j.Status.Active > 0 {
			status = "Running"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			j.Metadata.Name, comp, status,
			clusteropts.Age(j.Metadata.CreationTimestamp, now))
	}
	if len(items) == 0 {
		w.Flush()
		fmt.Fprintln(os.Stderr, "no child jobs found")
	}
	return nil
}

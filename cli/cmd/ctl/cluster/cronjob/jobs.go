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
// Lists the child Jobs spawned by one CronJob.
//
// Why ownerReferences and not labelSelector
// -----------------------------------------
//
// Earlier revisions of this verb relied on the SPA's trick of
// rebuilding a labelSelector from `spec.jobTemplate.metadata.labels`
// and asking the apiserver to filter by it. That worked for CronJobs
// authored in the KubeSphere/SPA UI (which auto-stamps the template
// with labels) but errored out on kubectl/yaml-authored CronJobs (the
// common case), because the standard K8s spec does NOT require
// jobTemplate labels and most users don't set them — see the
// user-visible "has no spec.jobTemplate.metadata.labels — cannot
// derive a Job selector" failure.
//
// The K8s-native binding between a CronJob and its child Jobs is
// `metadata.ownerReferences` (the Job carries a controller=true
// OwnerReference pointing back to the parent CronJob by UID). This
// works regardless of whether the user remembered to set template
// labels. So we always rely on UID matching as the source of truth.
//
// Two-step flow (label pre-filter optimization, then verify):
//
//  1. GET the CronJob to read `metadata.uid` (the authoritative key)
//     plus `spec.jobTemplate.metadata.labels` (the optimization key).
//  2. If labels are non-empty, ask the apiserver to pre-narrow with
//     labelSelector (saves wire bytes on big namespaces); otherwise
//     fetch every Job in the namespace.
//  3. Either way, run the client-side `ownerReferences` filter on
//     the candidate set. Labels are best-effort — Jobs can share
//     labels across CronJobs (or with manually-created Jobs in the
//     same namespace) — so the UID check is the final authority.
//
// `--limit N` caps the items DISPLAYED after filtering (default 100),
// not the apiserver request size — the pre-filter (when usable)
// keeps the wire transfer small enough that we can fetch generously
// and filter locally.
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
  1. GET the CronJob to read its UID (and any labels on
     spec.jobTemplate.metadata.labels as an optimization key).
  2. List candidate Jobs from the namespace — using
     ` + "`labelSelector=<derived>`" + ` when the jobTemplate carries labels
     (apiserver prefilter), or unfiltered otherwise.
  3. Filter client-side by ownerReferences[uid == <cronjob.uid>,
     controller=true, kind=CronJob]. The UID match is the source of
     truth; labels are only used to keep the wire response small.

This handles both KubeSphere/SPA-authored CronJobs (which carry
template labels) and kubectl/yaml-authored CronJobs (which usually
don't) — the verb no longer errors out on missing labels.

` + "`--limit N`" + ` caps the items displayed after filtering (default
100).

Output columns mirror ` + "`cluster job list`" + ` for consistency.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if limit < 0 {
				return fmt.Errorf("--limit must be >= 0, got %d", limit)
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runJobs(c.Context(), o, ns, name, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to display after filtering (0 = unlimited)")
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
	if c.Metadata.UID == "" {
		return fmt.Errorf("cronjob %s/%s has no metadata.uid — server response missing the field", namespace, name)
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}

	// When jobTemplate carries labels, ask the apiserver to pre-narrow
	// the candidate set — saves wire bytes on namespaces with many
	// unrelated Jobs. When labels are absent (the common kubectl case),
	// fall back to listing every Job in the namespace. Either way the
	// client-side UID filter below is the actual authority.
	q := url.Values{}
	if sel := c.templateLabelSelector(); sel != "" {
		q.Set("labelSelector", sel)
	}
	listPath := fmt.Sprintf("/apis/batch/v1/namespaces/%s/jobs", url.PathEscape(namespace))
	if encoded := q.Encode(); encoded != "" {
		listPath += "?" + encoded
	}

	resp, err := clusterclient.GetK8sList[job.Job](ctx, client, listPath)
	if err != nil {
		return fmt.Errorf("list jobs spawned by cronjob %s/%s: %w", namespace, name, err)
	}

	// Client-side UID filter — this is the K8s-native source of
	// truth for parent/child binding. We retain the order the
	// apiserver returned (which is reasonably stable across calls
	// for native list endpoints) so repeat runs diff cleanly.
	children := make([]job.Job, 0, len(resp.Items))
	for _, j := range resp.Items {
		if isChildOfCronJob(j, c.Metadata.UID) {
			children = append(children, j)
		}
	}

	totalMatched := len(children)
	truncated := false
	if limit > 0 && len(children) > limit {
		children = children[:limit]
		truncated = true
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items        []job.Job `json:"items"`
			TotalMatched int       `json:"totalMatched"`
			Limit        int       `json:"limit"`
			Truncated    bool      `json:"truncated,omitempty"`
		}{Items: children, TotalMatched: totalMatched, Limit: limit, Truncated: truncated})
	}
	if o.Quiet {
		return nil
	}
	return renderChildJobsTable(children, totalMatched, truncated, limit, o.NoHeaders)
}

// isChildOfCronJob reports whether the given Job is a controller-owned
// child of the CronJob identified by `parentUID`. Mirrors the K8s
// garbage-collector's notion of ownership:
//
//   - `Controller == true` — the parent is THE controller, not just a
//     soft reference written by some operator;
//   - `UID == parentUID` — UIDs are guaranteed unique across renames /
//     recreates, so this is the only safe equality;
//   - `Kind == "CronJob"` — defensive; in practice the apiserver won't
//     hand back a foreign Kind sharing a UID, but pinning the type
//     avoids any future surprise.
//
// We deliberately do NOT match on `Name` — a CronJob deleted and
// recreated with the same name gets a fresh UID, so name-matching
// would attribute children to the wrong (current) parent.
func isChildOfCronJob(j job.Job, parentUID string) bool {
	if parentUID == "" {
		return false
	}
	for _, o := range j.Metadata.OwnerReferences {
		if !o.Controller {
			continue
		}
		if o.UID != parentUID {
			continue
		}
		if o.Kind != "" && o.Kind != "CronJob" {
			continue
		}
		return true
	}
	return false
}

// renderChildJobsTable lays out the child Jobs in the same column
// shape as `cluster job list` so users can read both outputs without
// re-parsing. We reuse the exported job.Job.StatusLabel() helper so
// the STATUS column matches `cluster job list` exactly — otherwise a
// suspended Job would show as "Suspended" in `cluster job list` and
// "Pending" here, and the case-insensitivity / Failing-state handling
// would silently diverge.
//
// `totalMatched` is the count BEFORE --limit truncation so users can
// see "showing 5 of 17" — same pattern as the paginated list verbs.
func renderChildJobsTable(items []job.Job, totalMatched int, truncated bool, limit int, noHeaders bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "NAME\tCOMPLETIONS\tSTATUS\tAGE")
	}
	now := time.Now()
	for _, j := range items {
		// NAME / COMPLETIONS (succeeded/total) / STATUS / AGE.
		// DURATION is dropped because the job.Job duration helper
		// is unexported and adding a fifth column here would push
		// the table past the SPA's lazy-load tree layout.
		comp := "-"
		if j.Spec.Completions != nil {
			comp = fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
		} else {
			comp = fmt.Sprintf("%d", j.Status.Succeeded)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			j.Metadata.Name, comp, j.StatusLabel(),
			clusteropts.Age(j.Metadata.CreationTimestamp, now))
	}
	w.Flush()

	switch {
	case len(items) == 0 && totalMatched == 0:
		fmt.Fprintln(os.Stderr, "no child jobs found")
	case truncated:
		fmt.Fprintf(os.Stderr, "(showing %d of %d matched — pass --limit %d to see more)\n",
			len(items), totalMatched, totalMatched)
	}
	return nil
}

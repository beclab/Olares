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
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGetCommand: `olares-cli cluster cronjob get <ns/name | name>
// [-n NS] [-o table|json]`.
//
// Hits the K8s native CronJob endpoint
// `/apis/batch/v1/namespaces/<ns>/cronjobs/<name>`.
//
// History note: the SPA's `getCornJobsDetail` (apps/.../network/index.ts)
// historically passed `apis/batch/v1beta1` from `API_VERSIONS.cronjobs`.
// That API version was removed in Kubernetes 1.25; Olares clusters
// only serve the v1 endpoint, so the CLI uses v1 even though the SPA
// constant is still stale.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "get <ns/name | name>",
		Short: "show one CronJob's details (K8s native shape)",
		Long: `Show one CronJob's full detail.

Identity may be passed as a single "<namespace>/<name>" positional or
as a bare "<name>" with -n <namespace>.

In table mode the output is a vertical key/value summary including
SCHEDULE / TIME-ZONE (when set) / SUSPEND / CONCURRENCY-POLICY /
STARTING-DEADLINE / SUCCESSFUL-JOBS-HISTORY-LIMIT /
FAILED-JOBS-HISTORY-LIMIT / ACTIVE jobs / LAST-SCHEDULE /
LAST-SUCCESSFUL / Job Template Selector.
In JSON mode the typed view is forwarded; for byte-perfect output use
` + "`cluster cronjob yaml`" + `.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runGet(c.Context(), o, ns, name)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	o.AddOutputFlags(cmd)
	return cmd
}

// Get is the exported single-CronJob fetcher used by sibling verbs:
// `cronjob jobs` needs spec.jobTemplate.metadata.labels to derive the
// child-Job selector, and the suspend/resume verbs use this to pick
// up resourceVersion (PATCH is RV-aware on the apiserver side).
func Get(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) (*CronJob, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return nil, err
	}
	path := buildGetPath(namespace, name)
	var c CronJob
	if err := clusterclient.GetK8sObject(ctx, client, path, &c); err != nil {
		return nil, fmt.Errorf("get cronjob %s/%s: %w", namespace, name, err)
	}
	if c.Kind == "" {
		c.Kind = "CronJob"
	}
	return &c, nil
}

func buildGetPath(namespace, name string) string {
	return fmt.Sprintf("/apis/batch/v1/namespaces/%s/cronjobs/%s",
		url.PathEscape(namespace), url.PathEscape(name))
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) error {
	c, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if o.IsJSON() {
		return o.PrintJSON(*c)
	}
	if o.Quiet {
		return nil
	}
	return renderGetTable(*c)
}

func renderGetTable(c CronJob) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	now := time.Now()

	fmt.Fprintf(w, "Name:\t%s\n", c.Metadata.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", clusteropts.DashIfEmpty(c.Metadata.Namespace))
	fmt.Fprintf(w, "Schedule:\t%s\n", clusteropts.DashIfEmpty(c.Spec.Schedule))
	if c.Spec.TimeZone != nil && *c.Spec.TimeZone != "" {
		fmt.Fprintf(w, "Time Zone:\t%s\n", *c.Spec.TimeZone)
	}
	fmt.Fprintf(w, "Suspend:\t%s\n", c.suspendLabel())
	if c.Spec.ConcurrencyPolicy != "" {
		fmt.Fprintf(w, "Concurrency Policy:\t%s\n", c.Spec.ConcurrencyPolicy)
	}
	if c.Spec.StartingDeadlineSeconds != nil {
		fmt.Fprintf(w, "Starting Deadline:\t%ds\n", *c.Spec.StartingDeadlineSeconds)
	}
	if c.Spec.SuccessfulJobsHistoryLimit != nil {
		fmt.Fprintf(w, "Successful Jobs History Limit:\t%d\n", *c.Spec.SuccessfulJobsHistoryLimit)
	}
	if c.Spec.FailedJobsHistoryLimit != nil {
		fmt.Fprintf(w, "Failed Jobs History Limit:\t%d\n", *c.Spec.FailedJobsHistoryLimit)
	}
	fmt.Fprintf(w, "Active Jobs:\t%s\n", c.activeJobsLabel())
	fmt.Fprintf(w, "Last Schedule:\t%s\n", c.lastScheduleLabel(now))
	if c.Status.LastSuccessfulTime != "" {
		fmt.Fprintf(w, "Last Successful:\t%s ago\n", clusteropts.Age(c.Status.LastSuccessfulTime, now))
	}
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(c.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", c.age(now))
	if sel := c.templateLabelSelector(); sel != "" {
		fmt.Fprintf(w, "Job Template Selector:\t%s\n", sel)
	}
	return nil
}

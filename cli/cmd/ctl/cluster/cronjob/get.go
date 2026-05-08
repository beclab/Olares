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
// Calls SPA's getCornJobsDetail (apps/.../network/index.ts):
// `/apis/batch/v1beta1/namespaces/<ns>/cronjobs/<name>`.
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
SCHEDULE / SUSPEND / CONCURRENCY-POLICY / ACTIVE jobs / LAST-SCHEDULE.
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
	return fmt.Sprintf("/apis/batch/v1beta1/namespaces/%s/cronjobs/%s",
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
	fmt.Fprintf(w, "Suspend:\t%s\n", c.suspendLabel())
	if c.Spec.ConcurrencyPolicy != "" {
		fmt.Fprintf(w, "Concurrency Policy:\t%s\n", c.Spec.ConcurrencyPolicy)
	}
	fmt.Fprintf(w, "Active Jobs:\t%s\n", c.activeJobsLabel())
	fmt.Fprintf(w, "Last Schedule:\t%s\n", c.lastScheduleLabel(now))
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(c.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", c.age(now))
	if sel := c.templateLabelSelector(); sel != "" {
		fmt.Fprintf(w, "Job Template Selector:\t%s\n", sel)
	}
	return nil
}

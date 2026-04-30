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

// NewListCommand: `olares-cli cluster cronjob list [-n NS] [-l SEL]
// [--limit N] [-o table|json] [--no-headers] [--quiet]`.
//
// Calls SPA's getCronjobs path
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts):
// `/kapis/resources.kubesphere.io/v1alpha3/cronjobs` (cross-ns) or
// `/kapis/.../namespaces/<ns>/cronjobs` (per-ns).
//
// Defaults to cross-namespace. NAMESPACE column appears in cross-ns
// mode.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace     string
		labelSelector string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list CronJobs visible to the active profile",
		Long: `List K8s CronJobs visible to the active profile.

Without -n, returns the union of every namespace the active token can
see. The output table includes a NAMESPACE column.

--label uses K8s label-selector syntax (e.g. "app=foo").
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, namespace, labelSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter cronjobs (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, namespace, labelSelector string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := buildListPath(namespace, labelSelector, limit)
	resp, err := clusterclient.GetKubeSphereList[CronJob](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list cronjobs: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []CronJob `json:"items"`
			TotalItems int       `json:"totalItems"`
		}{Items: resp.Items, TotalItems: resp.TotalItems})
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(resp.Items, namespace == "", o.NoHeaders, len(resp.Items) < resp.TotalItems, resp.TotalItems)
}

func buildListPath(namespace, label string, limit int) string {
	base := "/kapis/resources.kubesphere.io/v1alpha3/cronjobs"
	if namespace != "" {
		base = "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
			url.PathEscape(namespace) + "/cronjobs"
	}
	q := url.Values{}
	if label != "" {
		q.Set("labelSelector", label)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func renderListTable(items []CronJob, showNamespace, noHeaders, paged bool, totalItems int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST-SCHEDULE\tAGE")
		} else {
			fmt.Fprintln(w, "NAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST-SCHEDULE\tAGE")
		}
	}
	now := time.Now()
	for _, c := range items {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				clusteropts.DashIfEmpty(c.Metadata.Namespace), c.Metadata.Name,
				clusteropts.DashIfEmpty(c.Spec.Schedule), c.suspendLabel(),
				len(c.Status.Active), c.lastScheduleLabel(now), c.age(now))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				c.Metadata.Name,
				clusteropts.DashIfEmpty(c.Spec.Schedule), c.suspendLabel(),
				len(c.Status.Active), c.lastScheduleLabel(now), c.age(now))
		}
	}
	if paged {
		w.Flush()
		fmt.Fprintf(os.Stderr, "(showing %d of %d total — pass --limit %d to see more)\n",
			len(items), totalItems, totalItems)
	}
	return nil
}

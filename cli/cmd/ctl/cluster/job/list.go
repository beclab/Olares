package job

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster job list [-n NS] [-l SEL]
// [--limit N] [--page N] [--all] [-o table|json] [--no-headers] [--quiet]`.
//
// Calls SPA's getJobs
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts):
// `/kapis/resources.kubesphere.io/v1alpha3/jobs` (cross-ns) or
// `/kapis/.../namespaces/<ns>/jobs` (per-ns).
//
// Defaults to cross-namespace (no -n) so the response is the union of
// every namespace the active token can see, mirroring the rest of the
// cluster list verbs. NAMESPACE column appears in cross-ns mode.
// Pagination matches the SPA wire shape; --all drains every page.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		namespace     string
		labelSelector string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Jobs visible to the active profile",
		Long: `List K8s Jobs visible to the active profile.

Without -n, returns the union of every namespace the active token can
see. The output table includes a NAMESPACE column.

--label uses K8s label-selector syntax (e.g. "app=foo").

Pagination: --limit sets the page size (default 100). --page picks one
1-indexed page (default 1). --all drains every page until exhausted
and is mutually exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, p, namespace, labelSelector)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter jobs (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, p *clusteropts.PaginationOptions, namespace, labelSelector string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.Validate(); err != nil {
		return err
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	items, total, err := clusteropts.FetchAllKubeSphere[Job](ctx, client, p, func(page int) string {
		return buildListPath(namespace, labelSelector, p, page)
	})
	if err != nil {
		return fmt.Errorf("list jobs: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []Job `json:"items"`
			TotalItems int   `json:"totalItems"`
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			All        bool  `json:"all,omitempty"`
		}{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit, All: p.All})
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(items, namespace == "", o.NoHeaders, p, total)
}

func buildListPath(namespace, label string, p *clusteropts.PaginationOptions, page int) string {
	base := "/kapis/resources.kubesphere.io/v1alpha3/jobs"
	if namespace != "" {
		base = "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
			url.PathEscape(namespace) + "/jobs"
	}
	q := url.Values{}
	if label != "" {
		q.Set("labelSelector", label)
	}
	p.AppendQueryForPage(q, page)
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func renderListTable(items []Job, showNamespace, noHeaders bool, p *clusteropts.PaginationOptions, totalItems int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tCOMPLETIONS\tSTATUS\tDURATION\tAGE")
		} else {
			fmt.Fprintln(w, "NAME\tCOMPLETIONS\tSTATUS\tDURATION\tAGE")
		}
	}
	now := time.Now()
	for _, j := range items {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(j.Metadata.Namespace), j.Metadata.Name,
				j.completionsLabel(), j.status(),
				j.duration(now), j.age(now))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				j.Metadata.Name,
				j.completionsLabel(), j.status(),
				j.duration(now), j.age(now))
		}
	}
	w.Flush()
	clusteropts.PrintPageHint(os.Stderr, p, len(items), totalItems)
	return nil
}

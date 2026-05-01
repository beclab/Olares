package pod

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

// NewListCommand: `olares-cli cluster pod list [-n ns] [-l selector]
// [--limit N] [--page N] [--all] [-o table|json] [--no-headers] [--quiet]`.
//
// Calls the SPA's getPodsList endpoint
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:52):
// `/kapis/resources.kubesphere.io/v1alpha3/pods` for the cross-namespace
// case (no -n) or
// `/kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/pods` for
// scoped-namespace.
//
// -n IS NOT defaulted — when omitted, we explicitly send the
// cross-namespace request and let the server return everything the
// active token can see. Output adds a NAMESPACE column in that mode
// so the result is still grep-able. CLI does not infer a "default
// namespace" from username or any other client-side heuristic; the
// security model is server-decides (see skills/olares-cluster).
//
// Pagination matches the SPA wire shape (limit + 1-indexed page).
// Use --all to drain every page in one command.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		namespace     string
		labelSelector string
		fieldSelector string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list pods visible to the active profile",
		Long: `List pods.

Without -n, returns the union of every namespace the server allows
the active profile to see (KubeSphere /kapis/resources.kubesphere.io
/v1alpha3/pods). The output table includes a NAMESPACE column.

With -n, scopes to a single namespace
(/kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/pods); the
server still decides whether you have access (a 403 means the namespace
exists but isn't visible to your token).

--label uses K8s label-selector syntax (e.g. "app=foo,tier=frontend").
--field-selector forwards K8s field selectors verbatim
(e.g. "spec.nodeName=node-1").

Pagination: --limit sets the page size (default 100). --page picks one
1-indexed page (default 1). --all drains every page until exhausted
and is mutually exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return RunList(c.Context(), o, p, namespace, labelSelector, fieldSelector)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter pods (K8s syntax)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "field selector to filter pods (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

// RunList is the exported entry point so sibling packages (e.g.
// cluster/application) can share the same fetch + render path
// without duplicating the table layout. opts and pagination are both
// required; pass a fresh clusteropts.NewClusterOptions(f) +
// clusteropts.NewPaginationOptions() when calling from outside cobra.
// namespace="" means cross-namespace.
func RunList(
	ctx context.Context,
	o *clusteropts.ClusterOptions,
	p *clusteropts.PaginationOptions,
	namespace, labelSelector, fieldSelector string,
) error {
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

	items, total, err := clusteropts.FetchAllKubeSphere[Pod](ctx, client, p, func(page int) string {
		return buildListPath(namespace, labelSelector, fieldSelector, p, page)
	})
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []Pod `json:"items"`
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

// buildListPath assembles the KubeSphere pods endpoint plus query
// string. Splits the namespace decision into the path itself (rather
// than a query param) so we get the cross-namespace endpoint when -n
// is empty.
//
// page is the 1-indexed page number for THIS request — driven by
// FetchAllKubeSphere's drain loop in --all mode, or p.Page in
// single-page mode.
func buildListPath(namespace, label, field string, p *clusteropts.PaginationOptions, page int) string {
	base := "/kapis/resources.kubesphere.io/v1alpha3/pods"
	if namespace != "" {
		// PathEscape is the right tool here: namespace is a path
		// segment, can in principle contain weird chars (it won't in
		// practice but defense-in-depth doesn't cost anything).
		base = "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" + url.PathEscape(namespace) + "/pods"
	}
	q := url.Values{}
	if label != "" {
		q.Set("labelSelector", label)
	}
	if field != "" {
		q.Set("fieldSelector", field)
	}
	p.AppendQueryForPage(q, page)
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func renderListTable(items []Pod, showNamespace, noHeaders bool, p *clusteropts.PaginationOptions, total int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	if !noHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tNODE\tIP")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\tNODE\tIP")
		}
	}

	now := time.Now()
	for _, p := range items {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(p.Metadata.Namespace),
				p.Metadata.Name,
				p.readyCount(),
				clusteropts.DashIfEmpty(p.statusReason()),
				p.totalRestarts(),
				p.age(now),
				clusteropts.DashIfEmpty(p.Spec.NodeName),
				clusteropts.DashIfEmpty(p.Status.PodIP),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
				p.Metadata.Name,
				p.readyCount(),
				clusteropts.DashIfEmpty(p.statusReason()),
				p.totalRestarts(),
				p.age(now),
				clusteropts.DashIfEmpty(p.Spec.NodeName),
				clusteropts.DashIfEmpty(p.Status.PodIP),
			)
		}
	}

	w.Flush()
	clusteropts.PrintPageHint(os.Stderr, p, len(items), total)
	return nil
}

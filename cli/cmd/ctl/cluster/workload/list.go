package workload

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster workload list [-n NS]
// [--kind all|deployment|statefulset|daemonset] [-l SEL] [--limit N]
// [--page N] [--all] [-o table|json] [--no-headers] [--quiet]`.
//
// Default --kind is "all" — the verb fans out one KubeSphere request
// per kind in SupportedKinds and merges the results into a single
// table with a KIND column. --kind <single> issues one request and
// drops the KIND column from the output.
//
// Pagination is per-kind. In --all mode each kind drains independently
// (so a deployment-heavy ns doesn't starve statefulsets). In --page
// mode the same page index is requested from every kind — useful for
// "give me page 2 across the board" symmetry.
//
// Calls SPA's getWorkloadList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:297):
// `/kapis/resources.kubesphere.io/v1alpha3/<kind>` cross-ns or
// `/kapis/.../namespaces/<ns>/<kind>` per-ns.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		namespace     string
		kindRaw       string
		labelSelector string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list workloads visible to the active profile",
		Long: `List Deployment / StatefulSet / DaemonSet workloads.

Without -n, returns the union of every namespace the server allows
the active profile to see. The output table includes a NAMESPACE
column.

--kind defaults to "all" (one request per kind, merged into a single
table with a KIND column). Pass --kind deployment / statefulset /
daemonset (singular or plural; "deploy" / "sts" / "ds" also accepted)
to scope to a single kind and drop the KIND column.

--label uses K8s label-selector syntax (e.g. "app=foo,tier=frontend").

Pagination: --limit sets the page size per kind (default 100). --page
picks one 1-indexed page per kind (default 1). --all drains every
page until exhausted (per-kind, independently) and is mutually
exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return RunList(c.Context(), o, p, namespace, kindRaw, labelSelector)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVar(&kindRaw, "kind", KindAll, "workload kind: all | deployment | statefulset | daemonset")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter workloads (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

// RunList is the exported entry point so sibling packages
// (cluster/application) can share the same fetch + render path.
func RunList(
	ctx context.Context,
	o *clusteropts.ClusterOptions,
	p *clusteropts.PaginationOptions,
	namespace, kindRaw, labelSelector string,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.Validate(); err != nil {
		return err
	}
	plural, err := NormalizeKind(kindRaw)
	if err != nil {
		return err
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}

	kindsToFetch := []string{plural}
	multi := false
	if plural == KindAll {
		kindsToFetch = []string{"deployments", "statefulsets", "daemonsets"}
		multi = true
	}

	type kindResult struct {
		Kind  string     `json:"kind"`
		Items []Workload `json:"items"`
		Total int        `json:"totalItems"`
	}
	var collected []kindResult
	for _, k := range kindsToFetch {
		items, total, err := clusteropts.FetchAllKubeSphere[Workload](ctx, client, p, func(page int) string {
			return buildListPath(k, namespace, labelSelector, p, page)
		})
		if err != nil {
			return fmt.Errorf("list %s: %w", k, err)
		}
		// Some KubeSphere versions strip Kind from list items —
		// stamp it from the per-call context so the KIND column is
		// always populated downstream.
		for i := range items {
			if items[i].Kind == "" {
				items[i].Kind = SingularKind(k)
			}
		}
		collected = append(collected, kindResult{Kind: k, Items: items, Total: total})
	}

	if o.IsJSON() {
		// JSON envelope mirrors the single-kind verbs (page/limit/all
		// alongside items/totalItems) so scripts can paginate against
		// any list verb with the same parser. For multi-kind, items
		// and totalItems are wrapped per-kind under `kinds`.
		type jsonOut struct {
			Kinds []kindResult `json:"kinds,omitempty"`
			Items []Workload   `json:"items,omitempty"`
			Total int          `json:"totalItems,omitempty"`
			Page  int          `json:"page"`
			Limit int          `json:"limit"`
			All   bool         `json:"all,omitempty"`
		}
		out := jsonOut{Page: p.Page, Limit: p.Limit, All: p.All}
		if multi {
			out.Kinds = collected
		} else {
			out.Items = collected[0].Items
			out.Total = collected[0].Total
		}
		return o.PrintJSON(out)
	}
	if o.Quiet {
		return nil
	}

	// Flatten + render a single table. Sort by namespace, kind, name
	// so cross-kind output is diff-friendly across runs.
	type row struct {
		Workload
		KindPlural string
	}
	var rows []row
	pagedKinds := []string{}
	for _, r := range collected {
		for _, w := range r.Items {
			rows = append(rows, row{Workload: w, KindPlural: r.Kind})
		}
		// In single-page mode we may have fetched only a subset of
		// each kind. In --all mode, FetchAllKubeSphere already drained
		// so this branch never trips. Per-kind hint defers to the
		// shared format used by single-kind verbs by computing the
		// same range string.
		if !p.All && len(r.Items) < r.Total {
			pageStart := (p.Page-1)*p.Limit + 1
			pageEnd := pageStart + len(r.Items) - 1
			if pageEnd >= pageStart {
				pagedKinds = append(pagedKinds,
					fmt.Sprintf("%s (items %d-%d of %d)", r.Kind, pageStart, pageEnd, r.Total))
			} else {
				pagedKinds = append(pagedKinds,
					fmt.Sprintf("%s (no items on page %d; total %d)", r.Kind, p.Page, r.Total))
			}
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Metadata.Namespace != rows[j].Metadata.Namespace {
			return rows[i].Metadata.Namespace < rows[j].Metadata.Namespace
		}
		if rows[i].Kind != rows[j].Kind {
			return rows[i].Kind < rows[j].Kind
		}
		return rows[i].Metadata.Name < rows[j].Metadata.Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	showNamespace := namespace == ""
	showKind := multi
	if !o.NoHeaders {
		switch {
		case showNamespace && showKind:
			fmt.Fprintln(w, "NAMESPACE\tKIND\tNAME\tREADY\tAGE")
		case showNamespace && !showKind:
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tAGE")
		case !showNamespace && showKind:
			fmt.Fprintln(w, "KIND\tNAME\tREADY\tAGE")
		default:
			fmt.Fprintln(w, "NAME\tREADY\tAGE")
		}
	}
	now := time.Now()
	for _, r := range rows {
		ready := r.Workload.Ready(r.KindPlural)
		age := r.Workload.Age(now)
		switch {
		case showNamespace && showKind:
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(r.Metadata.Namespace), clusteropts.DashIfEmpty(r.Kind), r.Metadata.Name, ready, age)
		case showNamespace && !showKind:
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(r.Metadata.Namespace), r.Metadata.Name, ready, age)
		case !showNamespace && showKind:
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(r.Kind), r.Metadata.Name, ready, age)
		default:
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Metadata.Name, ready, age)
		}
	}
	if len(pagedKinds) > 0 {
		w.Flush()
		fmt.Fprintf(os.Stderr, "(some kinds were paged: %s — pass --page %d or --all to see more)\n",
			strings.Join(pagedKinds, "; "), p.Page+1)
	}
	if len(rows) == 0 {
		w.Flush()
		fmt.Fprintln(os.Stderr, "no workloads visible to this profile")
	}
	return nil
}

// buildListPath assembles the KubeSphere workload endpoint plus
// query string. Splits the namespace decision into the path itself
// rather than a query param so we hit the cross-namespace endpoint
// when -n is empty.
func buildListPath(kindPlural, namespace, label string, p *clusteropts.PaginationOptions, page int) string {
	base := "/kapis/resources.kubesphere.io/v1alpha3/" + kindPlural
	if namespace != "" {
		base = "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
			url.PathEscape(namespace) + "/" + kindPlural
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

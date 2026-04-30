package workload

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster workload list [-n NS]
// [--kind all|deployment|statefulset|daemonset] [-l SEL] [--limit N]
// [-o table|json] [--no-headers] [--quiet]`.
//
// Default --kind is "all" — the verb fans out one KubeSphere request
// per kind in SupportedKinds and merges the results into a single
// table with a KIND column. --kind <single> issues one request and
// drops the KIND column from the output.
//
// Calls SPA's getWorkloadList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:297):
// `/kapis/resources.kubesphere.io/v1alpha3/<kind>` cross-ns or
// `/kapis/.../namespaces/<ns>/<kind>` per-ns.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace     string
		kindRaw       string
		labelSelector string
		limit         int
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
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return RunList(c.Context(), o, namespace, kindRaw, labelSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVar(&kindRaw, "kind", KindAll, "workload kind: all | deployment | statefulset | daemonset")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter workloads (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items per kind to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

// RunList is the exported entry point so sibling packages
// (cluster/application) can share the same fetch + render path.
func RunList(ctx context.Context, o *clusteropts.ClusterOptions, namespace, kindRaw, labelSelector string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
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
		path := buildListPath(k, namespace, labelSelector, limit)
		resp, err := clusterclient.GetKubeSphereList[Workload](ctx, client, path)
		if err != nil {
			return fmt.Errorf("list %s: %w", k, err)
		}
		// Some KubeSphere versions strip Kind from list items —
		// stamp it from the per-call context so the KIND column is
		// always populated downstream.
		for i := range resp.Items {
			if resp.Items[i].Kind == "" {
				resp.Items[i].Kind = SingularKind(k)
			}
		}
		collected = append(collected, kindResult{Kind: k, Items: resp.Items, Total: resp.TotalItems})
	}

	if o.IsJSON() {
		if multi {
			return o.PrintJSON(collected)
		}
		return o.PrintJSON(struct {
			Items      []Workload `json:"items"`
			TotalItems int        `json:"totalItems"`
		}{Items: collected[0].Items, TotalItems: collected[0].Total})
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
		if len(r.Items) < r.Total {
			pagedKinds = append(pagedKinds, fmt.Sprintf("%s (%d of %d)", r.Kind, len(r.Items), r.Total))
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
		fmt.Fprintf(os.Stderr, "(some kinds were paged: %v — pass --limit higher to see more)\n", pagedKinds)
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
func buildListPath(kindPlural, namespace, label string, limit int) string {
	base := "/kapis/resources.kubesphere.io/v1alpha3/" + kindPlural
	if namespace != "" {
		base = "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
			url.PathEscape(namespace) + "/" + kindPlural
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

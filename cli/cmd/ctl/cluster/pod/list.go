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
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster pod list [-n ns] [-l selector]
// [--limit N] [-o table|json] [--no-headers] [--quiet]`.
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
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace     string
		labelSelector string
		fieldSelector string
		limit         int
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
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return RunList(c.Context(), o, namespace, labelSelector, fieldSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter pods (K8s syntax)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "field selector to filter pods (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap; KubeSphere returns the page even if more exist)")
	o.AddOutputFlags(cmd)
	return cmd
}

// RunList is the exported entry point so sibling packages (e.g.
// cluster/application) can share the same fetch + render path
// without duplicating the table layout. opts is required; pass a
// fresh clusteropts.NewClusterOptions(f) when calling from outside
// cobra. namespace="" means cross-namespace.
func RunList(ctx context.Context, o *clusteropts.ClusterOptions, namespace, labelSelector, fieldSelector string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}

	path := buildListPath(namespace, labelSelector, fieldSelector, limit)
	resp, err := clusterclient.GetKubeSphereList[Pod](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []Pod `json:"items"`
			TotalItems int   `json:"totalItems"`
		}{Items: resp.Items, TotalItems: resp.TotalItems})
	}
	return renderListTable(resp.Items, namespace == "", o.NoHeaders, len(resp.Items) < resp.TotalItems, resp.TotalItems)
}

// buildListPath assembles the KubeSphere pods endpoint plus query
// string. Splits the namespace decision into the path itself (rather
// than a query param) so we get the cross-namespace endpoint when -n
// is empty.
func buildListPath(namespace, label, field string, limit int) string {
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
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

func renderListTable(items []Pod, showNamespace, noHeaders, paged bool, totalItems int) error {
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
				dashIfEmpty(p.Metadata.Namespace),
				p.Metadata.Name,
				p.readyCount(),
				dashIfEmpty(p.statusReason()),
				p.totalRestarts(),
				p.age(now),
				dashIfEmpty(p.Spec.NodeName),
				dashIfEmpty(p.Status.PodIP),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
				p.Metadata.Name,
				p.readyCount(),
				dashIfEmpty(p.statusReason()),
				p.totalRestarts(),
				p.age(now),
				dashIfEmpty(p.Spec.NodeName),
				dashIfEmpty(p.Status.PodIP),
			)
		}
	}

	if paged {
		// Server-side pagination capped this batch; print a soft
		// hint so users know to bump --limit.
		w.Flush()
		fmt.Fprintf(os.Stderr, "(showing %d of %d total — pass --limit %d to see more)\n",
			len(items), totalItems, totalItems)
	}
	return nil
}

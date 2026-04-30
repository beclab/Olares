package namespace

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

// NewListCommand: `olares-cli cluster namespace list [-l SEL]
// [--limit N] [-o table|json]`.
//
// Calls SPA's getNamespacesList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:256):
// `/kapis/resources.kubesphere.io/v1alpha3/namespaces`. Server-side
// scoping decides what's visible; CLI never filters or expands.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		labelSelector string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list K8s namespaces visible to the active profile",
		Long: `List K8s namespaces visible to the active profile.

Output (table mode): NAME, PHASE, WORKSPACE, AGE — flat kubectl-
style listing. WORKSPACE is read from the kubesphere.io/workspace
label when present, "-" otherwise.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, labelSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter namespaces (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
}

// nsItem is the per-namespace projection we render. Mirrors the
// subset of corev1.Namespace `kubectl get namespaces` displays plus
// the kubesphere workspace label.
type nsItem struct {
	Metadata struct {
		Name              string            `json:"name"`
		CreationTimestamp string            `json:"creationTimestamp,omitempty"`
		Labels            map[string]string `json:"labels,omitempty"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase,omitempty"`
	} `json:"status,omitempty"`
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, labelSelector string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	q := url.Values{}
	if labelSelector != "" {
		q.Set("labelSelector", labelSelector)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/kapis/resources.kubesphere.io/v1alpha3/namespaces"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	resp, err := clusterclient.GetKubeSphereList[nsItem](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list namespaces: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []nsItem `json:"items"`
			TotalItems int      `json:"totalItems"`
		}{Items: resp.Items, TotalItems: resp.TotalItems})
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(resp.Items, o.NoHeaders, len(resp.Items) < resp.TotalItems, resp.TotalItems)
}

func renderListTable(items []nsItem, noHeaders bool, paged bool, totalItems int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "NAME\tPHASE\tWORKSPACE\tAGE")
	}
	now := time.Now()
	for _, ns := range items {
		ws := "-"
		if v, ok := ns.Metadata.Labels["kubesphere.io/workspace"]; ok && v != "" {
			ws = v
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			ns.Metadata.Name,
			clusteropts.DashIfEmpty(ns.Status.Phase),
			ws,
			clusteropts.Age(ns.Metadata.CreationTimestamp, now),
		)
	}
	if paged {
		w.Flush()
		fmt.Fprintf(os.Stderr, "(showing %d of %d total — pass --limit %d to see more)\n",
			len(items), totalItems, totalItems)
	}
	return nil
}

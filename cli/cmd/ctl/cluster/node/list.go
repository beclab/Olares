package node

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

// NewListCommand: `olares-cli cluster node list [-l SEL] [--limit N]
// [--page N] [--all] [-o table|json]`.
//
// Calls SPA's getNodesList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:46):
// `/kapis/resources.kubesphere.io/v1alpha3/nodes`. Server-side
// scoping decides what's visible.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var labelSelector string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list K8s nodes visible to the active profile",
		Long: `List Kubernetes nodes visible to the active profile.

Output (table mode): NAME, STATUS, ROLES, AGE, VERSION, INTERNAL-IP
— same shape kubectl uses. STATUS = Ready /
"Ready,SchedulingDisabled" / NotReady / Unknown derived from the
Ready condition + spec.unschedulable. ROLES are read from
node-role.kubernetes.io/* labels.

Pagination: --limit sets the page size (default 100). --page picks one
1-indexed page (default 1). --all drains every page until exhausted
and is mutually exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, p, labelSelector)
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter nodes (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, p *clusteropts.PaginationOptions, labelSelector string) error {
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
	items, total, err := clusteropts.FetchAllKubeSphere[Node](ctx, client, p, func(page int) string {
		q := url.Values{}
		if labelSelector != "" {
			q.Set("labelSelector", labelSelector)
		}
		p.AppendQueryForPage(q, page)
		path := "/kapis/resources.kubesphere.io/v1alpha3/nodes"
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
		return path
	})
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []Node `json:"items"`
			TotalItems int    `json:"totalItems"`
			Page       int    `json:"page"`
			Limit      int    `json:"limit"`
			All        bool   `json:"all,omitempty"`
		}{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit, All: p.All})
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(items, o.NoHeaders, p, total)
}

func renderListTable(items []Node, noHeaders bool, p *clusteropts.PaginationOptions, totalItems int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION\tINTERNAL-IP")
	}
	now := time.Now()
	for _, n := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			n.Metadata.Name,
			n.StatusLabel(),
			n.Roles(),
			n.Age(now),
			clusteropts.DashIfEmpty(n.KubeletVersion()),
			n.InternalIP(),
		)
	}
	w.Flush()
	clusteropts.PrintPageHint(os.Stderr, p, len(items), totalItems)
	return nil
}

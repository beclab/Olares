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
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster node list [-l SEL] [--limit N]
// [-o table|json]`.
//
// Calls SPA's getNodesList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:46):
// `/kapis/resources.kubesphere.io/v1alpha3/nodes`. Server-side
// scoping decides what's visible.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		labelSelector string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list K8s nodes visible to the active profile",
		Long: `List Kubernetes nodes visible to the active profile.

Output (table mode): NAME, STATUS, ROLES, AGE, VERSION, INTERNAL-IP
— same shape kubectl uses. STATUS = Ready /
"Ready,SchedulingDisabled" / NotReady / Unknown derived from the
Ready condition + spec.unschedulable. ROLES are read from
node-role.kubernetes.io/* labels.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, labelSelector, limit)
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter nodes (K8s syntax)")
	cmd.Flags().IntVar(&limit, "limit", 100, "max items to fetch in one request (server-side cap)")
	o.AddOutputFlags(cmd)
	return cmd
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
	path := "/kapis/resources.kubesphere.io/v1alpha3/nodes"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	resp, err := clusterclient.GetKubeSphereList[Node](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []Node `json:"items"`
			TotalItems int    `json:"totalItems"`
		}{Items: resp.Items, TotalItems: resp.TotalItems})
	}
	return renderListTable(resp.Items, o.NoHeaders, len(resp.Items) < resp.TotalItems, resp.TotalItems)
}

func renderListTable(items []Node, noHeaders, paged bool, totalItems int) error {
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
			dashIfEmpty(n.KubeletVersion()),
			n.InternalIP(),
		)
	}
	if paged {
		w.Flush()
		fmt.Fprintf(os.Stderr, "(showing %d of %d total — pass --limit %d to see more)\n",
			len(items), totalItems, totalItems)
	}
	return nil
}

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
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster namespace list [-l SEL]
// [--limit N] [--page N] [--all] [-o table|json]`.
//
// Calls SPA's getNamespacesList
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:256):
// `/kapis/resources.kubesphere.io/v1alpha3/namespaces`. Server-side
// scoping decides what's visible; CLI never filters or expands.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var labelSelector string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list K8s namespaces visible to the active profile",
		Long: `List K8s namespaces visible to the active profile.

Output (table mode): NAME, PHASE, WORKSPACE, AGE — flat kubectl-
style listing. WORKSPACE is read from the kubesphere.io/workspace
label when present, "-" otherwise.

Pagination: --limit sets the page size (default 100). --page picks one
1-indexed page (default 1). --all drains every page until exhausted
and is mutually exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, p, labelSelector)
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter namespaces (K8s syntax)")
	p.AddPaginationFlags(cmd)
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
	items, total, err := clusteropts.FetchAllKubeSphere[nsItem](ctx, client, p, func(page int) string {
		q := url.Values{}
		if labelSelector != "" {
			q.Set("labelSelector", labelSelector)
		}
		p.AppendQueryForPage(q, page)
		path := "/kapis/resources.kubesphere.io/v1alpha3/namespaces"
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
		return path
	})
	if err != nil {
		return fmt.Errorf("list namespaces: %w", err)
	}
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items      []nsItem `json:"items"`
			TotalItems int      `json:"totalItems"`
			Page       int      `json:"page"`
			Limit      int      `json:"limit"`
			All        bool     `json:"all,omitempty"`
		}{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit, All: p.All})
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(items, o.NoHeaders, p, total)
}

func renderListTable(items []nsItem, noHeaders bool, p *clusteropts.PaginationOptions, totalItems int) error {
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
	w.Flush()
	clusteropts.PrintPageHint(os.Stderr, p, len(items), totalItems)
	return nil
}

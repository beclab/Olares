package application

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
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster application list [-o table|json]
// [--label SELECTOR] [--no-headers] [--quiet]`
//
// Calls SPA's getNamespacesGroup
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:264):
// `/capi/namespaces/group`. The endpoint returns a non-enveloped
// array of {title, data:[Namespace]} entries — unlike the KubeSphere
// /kapis/* paths there is no {items, totalItems} wrapper, so we
// decode straight into []NamespaceGroup.
//
// SPA passes labelSelector="kubesphere.io/workspace!=
// kubesphere.io/devopsproject" by default to hide internal devops
// projects; we forward --label verbatim so users can choose their
// own filter. Empty (unset) means "no filter, return what the server
// considers visible".
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		labelSelector string
		sortBy        string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list application spaces visible to the active profile, grouped by workspace",
		Long: `List application spaces (K8s Namespaces) visible to the active
profile, grouped by their KubeSphere workspace.

Server-side scope: /capi/namespaces/group only returns workspaces /
namespaces the active token can see. CLI never filters or expands
that set client-side.

Default --label filters out devops-project namespaces (matches the
SPA default). Pass --label "" to include every namespace the server
returns; pass any other selector to forward it verbatim.

Output:
  table  WORKSPACE | NAMESPACE | AGE  (one row per namespace)
  json   the raw "[{title, data:[Namespace]}, ...]" array verbatim
         so scripts can keep the workspace grouping intact.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, labelSelector, sortBy)
		},
	}
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "kubesphere.io/workspace!=kubesphere.io/devopsproject",
		`label selector forwarded to the server (set to "" to include every visible namespace)`)
	cmd.Flags().StringVar(&sortBy, "sort-by", "createTime",
		"server-side sort key (mirrors the SPA's getNamespacesGroup default)")
	o.AddOutputFlags(cmd)
	return cmd
}

// NamespaceGroup is one entry in the /capi/namespaces/group response.
// `title` is the workspace name; `data` is the list of Namespaces
// scoped to that workspace.
type NamespaceGroup struct {
	Title string      `json:"title"`
	Data  []Namespace `json:"data"`
}

// Namespace is the minimal corev1.Namespace view used to render
// list rows. Only metadata fields are exercised today; spec / status
// stay out until a verb actually needs them.
type Namespace struct {
	Metadata struct {
		Name              string            `json:"name"`
		CreationTimestamp string            `json:"creationTimestamp,omitempty"`
		Labels            map[string]string `json:"labels,omitempty"`
		Annotations       map[string]string `json:"annotations,omitempty"`
	} `json:"metadata"`
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, labelSelector, sortBy string) error {
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
	if sortBy != "" {
		q.Set("sortBy", sortBy)
	}
	path := "/capi/namespaces/group"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var groups []NamespaceGroup
	if err := client.DoJSON(ctx, "GET", path, nil, &groups); err != nil {
		return fmt.Errorf("list application spaces: %w", err)
	}

	if o.IsJSON() {
		return o.PrintJSON(groups)
	}
	return renderListTable(groups, o.NoHeaders)
}

func renderListTable(groups []NamespaceGroup, noHeaders bool) error {
	// Stable group order (server already sorts by createTime within a
	// group, but cross-group order isn't guaranteed). Sorting by
	// title keeps repeated invocations diff-friendly.
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Title < groups[j].Title
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "WORKSPACE\tNAMESPACE\tAGE")
	}

	now := time.Now()
	any := false
	for _, g := range groups {
		for _, ns := range g.Data {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				clusteropts.DashIfEmpty(g.Title),
				clusteropts.DashIfEmpty(ns.Metadata.Name),
				clusteropts.Age(ns.Metadata.CreationTimestamp, now),
			)
			any = true
		}
	}
	if !any {
		w.Flush()
		fmt.Fprintln(os.Stderr, "no application spaces visible to this profile")
	}
	return nil
}


package pod

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
--field-selector accepts kubectl-style field selectors and translates
them to the equivalent KubeSphere /kapis/resources.kubesphere.io/v1alpha3
filter params (the upstream endpoint does NOT understand the raw
"status.phase=Running" wire syntax). Supported fields:
  status.phase        -> filters by pod phase (Running / Pending / ...)
  spec.nodeName       -> filters by the scheduled node
  metadata.name       -> exact pod name (comma-separated for multi-name)
  metadata.namespace  -> namespace (rarely needed: prefer -n)
Only the '=' and '==' operators are recognized; '!=' and set-based
selectors are rejected with an error so users don't get a silently
empty result.

Pagination: --limit sets the page size (default 100). --page picks one
1-indexed page (default 1). --all drains every page until exhausted
and is mutually exclusive with --page > 1.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			// kubectl-style "No resources found" on empty list. We
			// only print it in human-readable modes; JSON consumers
			// already see {items: [], totalItems: 0} and don't need
			// the courtesy line cluttering their parsed output.
			n, err := RunList(c.Context(), o, p, namespace, labelSelector, fieldSelector)
			if err != nil {
				return err
			}
			if n == 0 && !o.IsJSON() && !o.Quiet {
				if namespace == "" {
					fmt.Fprintln(os.Stderr, "No pods found.")
				} else {
					fmt.Fprintf(os.Stderr, "No pods found in %s namespace.\n", namespace)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter pods (K8s syntax)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "K8s-style field selector (supported: status.phase, spec.nodeName, metadata.name, metadata.namespace; only '=' / '==' operators)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

// RunList is the exported entry point so sibling packages (e.g.
// cluster/application, cluster/job) can share the same fetch + render
// path without duplicating the table layout. opts and pagination are
// both required; pass a fresh clusteropts.NewClusterOptions(f) +
// clusteropts.NewPaginationOptions() when calling from outside cobra.
// namespace="" means cross-namespace.
//
// Returns the number of items rendered (post-filter, post-pagination)
// so callers can layer their own empty-result UX on top — e.g.
// `cluster job pods` prints a Job-specific hint about garbage-collected
// pods instead of the generic "No pods found." that `cluster pod list`
// uses. The count is the length of the items slice that actually went
// out the door (not the server-side totalItems), so JSON-mode callers
// can also distinguish "page returned no rows" from "no items match".
func RunList(
	ctx context.Context,
	o *clusteropts.ClusterOptions,
	p *clusteropts.PaginationOptions,
	namespace, labelSelector, fieldSelector string,
) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.Validate(); err != nil {
		return 0, err
	}
	client, err := o.Prepare()
	if err != nil {
		return 0, err
	}

	// Translate the kubectl-style --field-selector into KubeSphere's
	// per-field query params up front so we surface a clear error
	// (rather than a silently empty list) when the user passes an
	// unsupported field or operator. See translatePodFieldSelector
	// for the full mapping table.
	fieldQ, err := translatePodFieldSelector(fieldSelector)
	if err != nil {
		return 0, err
	}

	items, total, err := clusteropts.FetchAllKubeSphere[Pod](ctx, client, p, func(page int) string {
		return buildListPath(namespace, labelSelector, fieldQ, p, page)
	})
	if err != nil {
		return 0, fmt.Errorf("list pods: %w", err)
	}

	if o.IsJSON() {
		return len(items), o.PrintJSON(struct {
			Items      []Pod `json:"items"`
			TotalItems int   `json:"totalItems"`
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			All        bool  `json:"all,omitempty"`
		}{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit, All: p.All})
	}
	if o.Quiet {
		return len(items), nil
	}
	return len(items), renderListTable(items, namespace == "", o.NoHeaders, p, total)
}

// buildListPath assembles the KubeSphere pods endpoint plus query
// string. Splits the namespace decision into the path itself (rather
// than a query param) so we get the cross-namespace endpoint when -n
// is empty.
//
// page is the 1-indexed page number for THIS request — driven by
// FetchAllKubeSphere's drain loop in --all mode, or p.Page in
// single-page mode.
//
// fieldQ carries the already-translated KubeSphere filter params
// (status, nodeName, names, namespace) derived from the user's
// kubectl-style --field-selector input. See translatePodFieldSelector.
func buildListPath(namespace, label string, fieldQ url.Values, p *clusteropts.PaginationOptions, page int) string {
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
	for k, vs := range fieldQ {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	p.AppendQueryForPage(q, page)
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

// translatePodFieldSelector maps kubectl-style field selector terms
// (e.g. "status.phase=Running,spec.nodeName=node-1") to the per-field
// query parameters the KubeSphere /kapis/resources.kubesphere.io
// /v1alpha3/[namespaces/<ns>/]pods endpoint actually recognizes.
//
// Why the translation exists: the KubeSphere v1alpha3 handler does
// NOT honor the raw K8s `fieldSelector=status.phase=Running` wire
// syntax — its pod-filter implementation only switches on the bespoke
// filter keys `nodeName`, `pvcName`, `serviceName`, `status` (and the
// shared `name` / `names` / `namespace`). Passing through the raw
// kubectl syntax silently returns an empty list because the unknown
// `fieldSelector` filter key falls into the default-false branch of
// every per-resource filter func upstream.
//
// We expose the small subset that has a direct equivalent. Unsupported
// fields and the `!=` operator return an explicit error so users see a
// clear failure instead of an empty-but-successful response.
//
// Mapping (left = kubectl input, right = KubeSphere query param):
//
//	status.phase        -> status     (matches pod.Status.Phase exactly)
//	spec.nodeName       -> nodeName   (exact node-name match)
//	metadata.name       -> names      (KubeSphere: comma-separated exact)
//	metadata.namespace  -> namespace  (exact ns match — rarely useful; -n is cleaner)
//
// Operators: only `=` and `==` (treated as synonyms). Set-based
// selectors (`in`, `notin`) are not supported by the upstream filter
// and are rejected here.
func translatePodFieldSelector(sel string) (url.Values, error) {
	out := url.Values{}
	if strings.TrimSpace(sel) == "" {
		return out, nil
	}
	for _, raw := range strings.Split(sel, ",") {
		term := strings.TrimSpace(raw)
		if term == "" {
			continue
		}
		if strings.Contains(term, "!=") {
			return nil, fmt.Errorf("--field-selector: %q uses the '!=' operator which the upstream KubeSphere pods endpoint does not support", term)
		}
		var lhs, rhs string
		switch {
		case strings.Contains(term, "=="):
			parts := strings.SplitN(term, "==", 2)
			lhs, rhs = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		case strings.Contains(term, "="):
			parts := strings.SplitN(term, "=", 2)
			lhs, rhs = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		default:
			return nil, fmt.Errorf("--field-selector: %q is not a valid term (expected key=value)", term)
		}
		if lhs == "" || rhs == "" {
			return nil, fmt.Errorf("--field-selector: %q has empty key or value", term)
		}
		switch lhs {
		case "status.phase":
			out.Set("status", rhs)
		case "spec.nodeName":
			out.Set("nodeName", rhs)
		case "metadata.name":
			out.Set("names", rhs)
		case "metadata.namespace":
			out.Set("namespace", rhs)
		default:
			supported := []string{"status.phase", "spec.nodeName", "metadata.name", "metadata.namespace"}
			sort.Strings(supported)
			return nil, fmt.Errorf("--field-selector: field %q is not supported (supported: %s)",
				lhs, strings.Join(supported, ", "))
		}
	}
	return out, nil
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

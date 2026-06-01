package workload

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewImagesCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		namespace     string
		kindRaw       string
		labelSelector string
	)
	cmd := &cobra.Command{
		Use:   "images [IMAGE]",
		Short: "list container images referenced by workload pod templates",
		Long: `List container images referenced by Deployment / StatefulSet /
DaemonSet / Job / CronJob pod templates.

--kind defaults to "all" (every kind above). Pass a single kind —
deployment / statefulset / daemonset / job / cronjob (singular, plural,
or the short forms deploy / sts / ds) — to scope.

Pass an optional IMAGE argument to answer "where is this image
referenced?" — output is filtered to the workloads whose containers use
that image (tag- and digest-normalized, so "nginx" matches
"docker.io/library/nginx:latest"). An IMAGE lookup always scans the
whole cluster (pagination is ignored) so the answer can't miss refs on
a later page.

Without an IMAGE argument the listing is paginated by default. Pass
--page / --limit to walk large clusters deliberately, or --all to drain
every page.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			image := ""
			if len(args) == 1 {
				image = strings.TrimSpace(args[0])
			}
			return runImages(c.Context(), o, p, namespace, kindRaw, labelSelector, image)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVar(&kindRaw, "kind", KindAll, "workload kind: all | deployment | statefulset | daemonset | job | cronjob")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter workloads (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

// resolveImageScanKinds maps the --kind value to a primary fetch arg
// (KindAll or a single Deployment/StatefulSet/DaemonSet plural, "" when
// none) plus the batch kinds (jobs / cronjobs) to scan. Unlike the
// shared NormalizeKind — which only models the three controller kinds
// the mutating verbs support — the images verb also accepts job /
// cronjob and folds both into "all".
func resolveImageScanKinds(kindRaw string) (primary string, batch []string, err error) {
	switch strings.ToLower(strings.TrimSpace(kindRaw)) {
	case KindAll, "":
		return KindAll, batchKinds, nil
	case "job", "jobs":
		return "", []string{"jobs"}, nil
	case "cronjob", "cronjobs":
		return "", []string{"cronjobs"}, nil
	}
	plural, nerr := NormalizeKind(kindRaw)
	if nerr != nil {
		return "", nil, fmt.Errorf("unsupported kind %q (want one of: all, deployment, statefulset, daemonset, job, cronjob)", kindRaw)
	}
	return plural, nil, nil
}

type workloadImageRef struct {
	Namespace     string `json:"namespace,omitempty"`
	Kind          string `json:"kind"`
	Workload      string `json:"workload"`
	ContainerType string `json:"containerType"`
	Container     string `json:"container"`
	Image         string `json:"image"`
}

type workloadImageScan struct {
	Kind          string `json:"kind"`
	ReturnedItems int    `json:"returnedItems"`
	TotalItems    int    `json:"totalItems"`
	HasMore       bool   `json:"hasMore,omitempty"`
}

func runImages(ctx context.Context, o *clusteropts.ClusterOptions, p *clusteropts.PaginationOptions, namespace, kindRaw, labelSelector, imageQuery string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.Validate(); err != nil {
		return err
	}
	// A targeted "where is this image referenced?" lookup must be
	// complete: a paged subset would miss refs on later pages and
	// wrongly report the image as unreferenced. Force a full scan.
	if imageQuery != "" {
		p.All = true
	}
	primary, batch, err := resolveImageScanKinds(kindRaw)
	if err != nil {
		return err
	}

	var (
		refs []workloadImageRef
		scan []workloadImageScan
	)
	if primary != "" {
		collected, _, ferr := fetchWorkloads(ctx, o, p, namespace, primary, labelSelector, true)
		if ferr != nil {
			return ferr
		}
		refs = append(refs, collectWorkloadImageRefs(collected)...)
		scan = append(scan, summarizeImageScan(collected, p)...)
	}
	if len(batch) > 0 {
		batchRefs, batchScan, berr := fetchBatchImageRefs(ctx, o, p, namespace, labelSelector, batch)
		if berr != nil {
			return berr
		}
		refs = append(refs, batchRefs...)
		scan = append(scan, batchScan...)
	}
	if imageQuery != "" {
		refs = filterRefsByImage(refs, imageQuery)
	}
	sortImageRefs(refs)
	if o.IsJSON() {
		return o.PrintJSON(struct {
			Refs  []workloadImageRef  `json:"refs"`
			Scan  []workloadImageScan `json:"scan"`
			Page  int                 `json:"page"`
			Limit int                 `json:"limit"`
			All   bool                `json:"all,omitempty"`
		}{Refs: refs, Scan: scan, Page: p.Page, Limit: p.Limit, All: p.All})
	}
	if o.Quiet {
		return nil
	}
	return renderImagesTable(refs, scan, o.NoHeaders, imageQuery)
}

func collectWorkloadImageRefs(results []workloadKindResult) []workloadImageRef {
	var refs []workloadImageRef
	for _, result := range results {
		for _, w := range result.Items {
			if w.Spec.Template == nil {
				continue
			}
			base := workloadImageRef{
				Namespace: w.Metadata.Namespace,
				Kind:      nonEmpty(w.Kind, SingularKind(result.Kind)),
				Workload:  w.Metadata.Name,
			}
			for _, c := range w.Spec.Template.Spec.InitContainers {
				if c.Image == "" {
					continue
				}
				ref := base
				ref.ContainerType = "initContainer"
				ref.Container = c.Name
				ref.Image = c.Image
				refs = append(refs, ref)
			}
			for _, c := range w.Spec.Template.Spec.Containers {
				if c.Image == "" {
					continue
				}
				ref := base
				ref.ContainerType = "container"
				ref.Container = c.Name
				ref.Image = c.Image
				refs = append(refs, ref)
			}
		}
	}
	return refs
}

// filterRefsByImage keeps only the references whose image matches the
// query, tag- and digest-normalized via imageMatchKeys (so "nginx"
// matches "docker.io/library/nginx:latest", and a digest pin matches
// by digest).
func filterRefsByImage(refs []workloadImageRef, query string) []workloadImageRef {
	want := map[string]struct{}{}
	for _, key := range imageMatchKeys(query) {
		want[key] = struct{}{}
	}
	var out []workloadImageRef
	for _, ref := range refs {
		for _, key := range imageMatchKeys(ref.Image) {
			if _, ok := want[key]; ok {
				out = append(out, ref)
				break
			}
		}
	}
	return out
}

// sortImageRefs orders refs deterministically so JSON and table output
// is stable across runs regardless of per-kind fetch / pagination order.
func sortImageRefs(refs []workloadImageRef) {
	sort.SliceStable(refs, func(i, j int) bool {
		a, b := refs[i], refs[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Workload != b.Workload {
			return a.Workload < b.Workload
		}
		if a.ContainerType != b.ContainerType {
			return a.ContainerType < b.ContainerType
		}
		if a.Container != b.Container {
			return a.Container < b.Container
		}
		return a.Image < b.Image
	})
}

func summarizeImageScan(results []workloadKindResult, p *clusteropts.PaginationOptions) []workloadImageScan {
	out := make([]workloadImageScan, 0, len(results))
	for _, result := range results {
		out = append(out, workloadImageScan{
			Kind:          result.Kind,
			ReturnedItems: len(result.Items),
			TotalItems:    result.Total,
			HasMore:       !p.All && len(result.Items) < result.Total,
		})
	}
	return out
}

func renderImagesTable(refs []workloadImageRef, scan []workloadImageScan, noHeaders bool, imageQuery string) error {
	if len(refs) == 0 {
		if imageQuery != "" {
			fmt.Fprintf(os.Stdout, "no workloads reference %q\n", imageQuery)
		} else {
			fmt.Fprintln(os.Stdout, "no workload images")
		}
		return printImageScanHint(scan)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if !noHeaders {
		fmt.Fprintln(tw, "NAMESPACE\tKIND\tWORKLOAD\tTYPE\tCONTAINER\tIMAGE")
	}
	for _, ref := range refs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			clusteropts.DashIfEmpty(ref.Namespace),
			clusteropts.DashIfEmpty(ref.Kind),
			clusteropts.DashIfEmpty(ref.Workload),
			clusteropts.DashIfEmpty(ref.ContainerType),
			clusteropts.DashIfEmpty(ref.Container),
			clusteropts.DashIfEmpty(ref.Image),
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return printImageScanHint(scan)
}

func printImageScanHint(scan []workloadImageScan) error {
	var paged []string
	for _, s := range scan {
		if s.HasMore {
			paged = append(paged, fmt.Sprintf("%s (%d of %d)", s.Kind, s.ReturnedItems, s.TotalItems))
		}
	}
	if len(paged) == 0 {
		return nil
	}
	_, err := fmt.Fprintf(os.Stderr, "(scan is paged: %s — pass --page <next> or --all to see more)\n", strings.Join(paged, "; "))
	return err
}

func nonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

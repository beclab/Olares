package workload

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/distribution/reference"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/containerdimages"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// containerdImagesVerb is the human-readable verb interpolated into the
// admin role-gate / permission-error hints.
const containerdImagesVerb = "list local containerd images"

// NewDoctorImagesCommand backs `olares-cli doctor images` — a local
// containerd image inventory annotated with how many workload container
// references point at each image. `--unused` narrows the view to images
// nothing references (orphans / prune candidates).
func NewDoctorImagesCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace     string
		labelSelector string
		registry      string
		unusedOnly    bool
	)
	cmd := &cobra.Command{
		Use:   "images",
		Short: "local containerd images with workload reference counts",
		Long: `List local containerd images (Settings -> Advanced) annotated with
how many workload pod-template references point at each one, across
Deployment / StatefulSet / DaemonSet / Job / CronJob.

The cluster is always scanned in full, so the reference count — and the
"unused" verdict — is complete; there is no pagination to reason about.
Use -n / -l to deliberately narrow the scan (the counts and the unused
verdict are then scoped to that selection).

--unused shows only images with zero references (prune candidates),
sorted largest-first, with a reclaimable-size summary.
Pause / sandbox images (repo tag contains "pause") are excluded — they
are runtime-pinned, never user-prunable. Digest-pinned references
(repo@sha256:...) match by digest as well as by tag.

Caveat: the local image list reflects the node that serves the request
(the control node), while workload references are cluster-wide. An
image present only on a worker node won't appear here, so treat the
"unused" set as "images on the control node that no workload
references".`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDoctorImages(c.Context(), f, o, namespace, labelSelector, registry, unusedOnly)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "scope workloads to a single namespace (default: all namespaces visible to your profile)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "label selector to filter workloads (K8s syntax)")
	cmd.Flags().StringVar(&registry, "registry", "", "filter local images by registry before comparison")
	cmd.Flags().BoolVar(&unusedOnly, "unused", false, "show only images with zero workload references")
	o.AddOutputFlags(cmd)
	return cmd
}

// imageUsage is one local image plus the workload references that point
// at it. Image is embedded so id/size/repo_tags/repo_digests stay at the
// top level of the JSON object.
type imageUsage struct {
	containerdimages.Image
	Refs   int                `json:"refs"`
	UsedBy []workloadImageRef `json:"used_by,omitempty"`
}

func runDoctorImages(ctx context.Context, f *cmdutil.Factory, o *clusteropts.ClusterOptions, namespace, labelSelector, registry string, unusedOnly bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// /api/containerd/images is admin-floor (mirrors `settings advanced
	// images`). Soft-gate on the cached role and translate a server 403
	// into the same refresh-and-retry hint.
	if err := gateContainerdImages(ctx, f); err != nil {
		return err
	}
	local, err := containerdimages.List(ctx, f, registry)
	if err != nil {
		return wrapContainerdImagesErr(ctx, f, err)
	}

	// A usage / unused verdict is only meaningful over a complete scan,
	// so always drain every page of every kind.
	p := &clusteropts.PaginationOptions{Limit: 100, Page: 1, All: true}
	collected, _, err := fetchWorkloads(ctx, o, p, namespace, KindAll, labelSelector, true)
	if err != nil {
		return err
	}
	refs := collectWorkloadImageRefs(collected)
	scan := summarizeImageScan(collected, p)

	batchRefs, batchScan, err := fetchBatchImageRefs(ctx, o, p, namespace, labelSelector, batchKinds)
	if err != nil {
		return err
	}
	refs = append(refs, batchRefs...)
	scan = append(scan, batchScan...)

	usage := computeImageUsage(local, refs)
	summary := summarizeImageUsage(usage)
	if unusedOnly {
		usage = filterUnused(usage)
		sortBySizeDesc(usage)
	}

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Images     []imageUsage        `json:"images"`
			Summary    imageUsageSummary   `json:"summary"`
			UnusedOnly bool                `json:"unused_only,omitempty"`
			Scan       []workloadImageScan `json:"scan"`
		}{Images: usage, Summary: summary, UnusedOnly: unusedOnly, Scan: scan})
	}
	if o.Quiet {
		return nil
	}
	return renderImageUsageTable(usage, summary, unusedOnly, o.NoHeaders)
}

// imageUsageSummary is the prune-decision header: how many images were
// scanned, how many are orphaned, and how much disk the orphans hold.
type imageUsageSummary struct {
	Total            int   `json:"total"`
	Unused           int   `json:"unused"`
	ReclaimableBytes int64 `json:"reclaimable_bytes"`
}

func summarizeImageUsage(usage []imageUsage) imageUsageSummary {
	s := imageUsageSummary{Total: len(usage)}
	for _, u := range usage {
		if u.Refs == 0 {
			s.Unused++
			s.ReclaimableBytes += u.Size
		}
	}
	return s
}

func sortBySizeDesc(usage []imageUsage) {
	sort.SliceStable(usage, func(i, j int) bool {
		if usage[i].Size != usage[j].Size {
			return usage[i].Size > usage[j].Size
		}
		ti, tj := firstRepoTag(usage[i].Image), firstRepoTag(usage[j].Image)
		if ti != tj {
			return ti < tj
		}
		return usage[i].ID < usage[j].ID
	})
}

// computeImageUsage joins local images with workload references, counting
// distinct references per image (digest- and tag-aware) and skipping
// runtime-pinned pause images. Output is deterministically sorted.
func computeImageUsage(images []containerdimages.Image, refs []workloadImageRef) []imageUsage {
	keyToRefs := map[string][]int{}
	for i, ref := range refs {
		for _, key := range imageMatchKeys(ref.Image) {
			keyToRefs[key] = append(keyToRefs[key], i)
		}
	}
	var out []imageUsage
	for _, image := range images {
		if isPauseImage(image) {
			continue
		}
		seen := map[int]struct{}{}
		for _, key := range localImageKeys(image) {
			for _, idx := range keyToRefs[key] {
				seen[idx] = struct{}{}
			}
		}
		u := imageUsage{Image: image, Refs: len(seen)}
		for idx := range seen {
			u.UsedBy = append(u.UsedBy, refs[idx])
		}
		sortImageRefs(u.UsedBy)
		out = append(out, u)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti, tj := firstRepoTag(out[i].Image), firstRepoTag(out[j].Image)
		if ti != tj {
			return ti < tj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func filterUnused(usage []imageUsage) []imageUsage {
	var out []imageUsage
	for _, u := range usage {
		if u.Refs == 0 {
			out = append(out, u)
		}
	}
	return out
}

func firstRepoTag(image containerdimages.Image) string {
	if len(image.RepoTags) > 0 {
		return image.RepoTags[0]
	}
	return ""
}

// isPauseImage flags runtime sandbox images the same way the backend
// PruneImages path does (daemon/pkg/containerd/api.go: skip any tag
// containing "pause"). These are pinned by the CRI and must never be
// surfaced as user-prunable orphans.
func isPauseImage(image containerdimages.Image) bool {
	for _, tag := range image.RepoTags {
		if strings.Contains(tag, "pause") {
			return true
		}
	}
	return false
}

// localImageKeys is the set of identities a local image can be matched
// by: its content digest (ID), every repo tag, and every repo digest —
// each normalized through imageMatchKeys so a workload reference that
// pins by tag OR by digest can match.
func localImageKeys(image containerdimages.Image) []string {
	var keys []string
	if id := strings.TrimSpace(image.ID); id != "" {
		keys = append(keys, id)
	}
	for _, tag := range image.RepoTags {
		keys = append(keys, imageMatchKeys(tag)...)
	}
	for _, digest := range image.RepoDigests {
		keys = append(keys, imageMatchKeys(digest)...)
	}
	return dedupeStrings(keys)
}

// imageMatchKeys expands a single image reference into the keys it can
// be matched on: the raw string, the tag-normalized canonical name, and
// — when the reference is digest-pinned — the bare digest (sha256:...)
// so it lines up with a local image's repo digest.
func imageMatchKeys(image string) []string {
	image = strings.TrimSpace(image)
	if image == "" {
		return nil
	}
	keys := []string{image}
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return keys
	}
	keys = append(keys, reference.TagNameOnly(named).String())
	if digested, ok := named.(reference.Digested); ok {
		keys = append(keys, digested.Digest().String())
	}
	return dedupeStrings(keys)
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// gateContainerdImages short-circuits when the cached role is provably
// below admin, mirroring settings/internal/preflight.Gate (which we
// cannot import from here). Any local lookup failure falls through so a
// flaky cache never blocks a legitimate call — the server stays
// authoritative via wrapContainerdImagesErr.
func gateContainerdImages(ctx context.Context, f *cmdutil.Factory) error {
	if f == nil {
		return nil
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil || rp == nil {
		return nil
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil || cfg == nil {
		return nil
	}
	return whoami.PreflightRole(cfg, rp.OlaresID, whoami.RoleAdmin, containerdImagesVerb)
}

func wrapContainerdImagesErr(ctx context.Context, f *cmdutil.Factory, err error) error {
	if err == nil {
		return nil
	}
	var olaresID string
	if f != nil {
		if rp, rerr := f.ResolveProfile(ctx); rerr == nil && rp != nil {
			olaresID = rp.OlaresID
		}
	}
	return whoami.WrapPermissionErr(err, olaresID, containerdImagesVerb)
}

func renderImageUsageTable(usage []imageUsage, summary imageUsageSummary, unusedOnly, noHeaders bool) error {
	if len(usage) == 0 {
		if unusedOnly {
			fmt.Fprintln(os.Stdout, "no unused images")
		} else {
			fmt.Fprintln(os.Stdout, "no images")
		}
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if !noHeaders {
		fmt.Fprintln(tw, "ID\tSIZE\tREFS\tREPO TAGS")
	}
	for _, u := range usage {
		tags := "-"
		if len(u.RepoTags) > 0 {
			tags = strings.Join(u.RepoTags, ",")
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			containerdimages.ShortID(u.ID), containerdimages.HumanBytes(u.Size), u.Refs, tags)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	printImageUsageSummary(summary, unusedOnly)
	return nil
}

// printImageUsageSummary writes the prune-decision footer to stderr so
// it never pollutes the parseable table on stdout.
func printImageUsageSummary(summary imageUsageSummary, unusedOnly bool) {
	reclaimable := containerdimages.HumanBytes(summary.ReclaimableBytes)
	if unusedOnly {
		fmt.Fprintf(os.Stderr, "(%d unused images · ~%s reclaimable)\n", summary.Unused, reclaimable)
		return
	}
	fmt.Fprintf(os.Stderr, "(%d images · %d unused · ~%s reclaimable)\n", summary.Total, summary.Unused, reclaimable)
}

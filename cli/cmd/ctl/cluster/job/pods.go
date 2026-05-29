package job

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewPodsCommand: `olares-cli cluster job pods <ns/name | name>
// [-n NS] [-l ...] [--field-selector ...] [--limit N] [--page N] [--all]`.
//
// Two-step: GET the Job, then list pods server-side using the
// K8s-native binding the Job controller itself uses to find its
// children — `spec.selector.matchLabels`. Mirrors the SPA's Job
// detail page (PodContainer) which pulls the same field via
// ObjectMapper.jobs and feeds it to getNameSpacePodsList.
//
// Why spec.selector.matchLabels (not a hardcoded controller-uid=)
// ---------------------------------------------------------------
//
// In K8s 1.27 the auto-generated Job selector switched from the
// legacy `controller-uid=<uid>` to `batch.kubernetes.io/controller-uid
// =<uid>` (see KEP-3850). The legacy labels are still stamped on Pods
// when the LegacyJobMetadata feature gate is enabled, but the *selector*
// the controller uses is whatever the apiserver wrote into
// spec.selector — so reading it off the Job removes our dependency on
// label-name compatibility, AND it correctly handles
// `manualSelector: true` Jobs (where the user picks the labels).
//
// When .spec.selector is missing for whatever reason (very old
// clusters, manually-crafted bare Job objects), we fall back to the
// historical `controller-uid=<uid>` clause so this command never
// silently degrades to a cross-Job listing.
//
// --label / --field-selector are appended to the derived selector so
// users can further filter on top. Pagination (--limit / --page /
// --all) is forwarded verbatim.
func NewPodsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	p := clusteropts.NewPaginationOptions()
	var (
		namespace     string
		labelSelector string
		fieldSelector string
	)
	cmd := &cobra.Command{
		Use:   "pods <ns/name | name>",
		Short: "list pods controlled by one Job (uses the Job's spec.selector.matchLabels)",
		Long: `List pods controlled by one Job.

Two-step: GET the Job, then list pods server-side via
labelSelector=<spec.selector.matchLabels>. This is the same selector
the Job controller itself uses to find its child pods, so it stays
correct across K8s versions (the auto-generated selector switched
from "controller-uid" to "batch.kubernetes.io/controller-uid" in K8s
1.27) and across manualSelector=true Jobs.

If the Job's spec.selector is missing entirely (very old or
hand-crafted bare Job objects), the verb falls back to the historical
labelSelector=controller-uid=<job.uid> clause so it never silently
degrades to a cross-Job listing.

--label and --field-selector are appended to the derived selector so
additional filtering happens server-side as well; the CLI never
filters or scopes pods locally.

When no pods match (e.g. a Completed Job whose pods were garbage-
collected by ttlSecondsAfterFinished or pruned by the CronJob history
limit), the verb prints a "No pods found" hint to stderr and exits 0
— the empty result is informational, not an error.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runPods(c.Context(), o, p, ns, name, labelSelector, fieldSelector)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVarP(&labelSelector, "label", "l", "", "additional label selector to filter pods (K8s syntax; ANDed with controller-uid=<uid>)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "field selector to filter pods (K8s syntax)")
	p.AddPaginationFlags(cmd)
	o.AddOutputFlags(cmd)
	return cmd
}

func runPods(ctx context.Context, o *clusteropts.ClusterOptions, p *clusteropts.PaginationOptions, namespace, name, extraLabel, fieldSelector string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	j, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if j.Metadata.UID == "" {
		return fmt.Errorf("job %s/%s has no metadata.uid — server response missing the field", namespace, name)
	}

	selector, source := jobPodSelector(*j)
	if selector == "" {
		// Should be unreachable: jobPodSelector always returns the
		// controller-uid fallback when the spec.selector path yields
		// nothing. Defense-in-depth so a future refactor can't drop
		// the safety net silently.
		return fmt.Errorf("job %s/%s: cannot derive a pod selector (spec.selector empty and metadata.uid blank)",
			namespace, name)
	}
	if extraLabel != "" {
		// K8s label selectors AND comma-separated clauses, so `,` is
		// the safe join character — same convention kubectl uses
		// internally when it merges multiple --selector arguments.
		selector += "," + extraLabel
	}

	n, err := pod.RunList(ctx, o, p, namespace, selector, fieldSelector)
	if err != nil {
		return err
	}
	if n == 0 && !o.IsJSON() && !o.Quiet {
		// Friendly hint that distinguishes "command broken" from
		// "command worked, there just aren't any pods" — the most
		// common confusion point. Calling out the selector lets the
		// user double-check with `kubectl get pods -l <selector>`
		// when they suspect a label mismatch (manualSelector Jobs,
		// or a cluster with an unusual feature-gate config).
		fmt.Fprintf(os.Stderr,
			"No pods found for job %s/%s (selector %q via %s).\n"+
				"Hint: completed Jobs often lose their pods to ttlSecondsAfterFinished or to the parent CronJob's successfulJobsHistoryLimit. Run `olares-cli cluster job get %s/%s` to confirm the Job is still around, or `olares-cli cluster job yaml %s/%s` to inspect spec.selector and spec.ttlSecondsAfterFinished.\n",
			namespace, name, selector, source, namespace, name, namespace, name)
	}
	return nil
}

// jobPodSelector picks the selector clause we'll send as
// labelSelector when listing the Job's child Pods, plus a short
// human-readable label describing where the clause came from (used
// in the empty-result hint so the user can tell which path was
// taken without having to read the code).
//
// Preference order:
//
//  1. spec.selector.matchLabels — the K8s-native source of truth.
//     This is what the Job controller itself uses to find its
//     children, so it stays correct across K8s versions and across
//     manualSelector=true Jobs. We join the entries with the
//     comma-AND syntax kubectl uses; key order is sorted so the
//     output (and any user-facing error / hint string) stays
//     deterministic across calls.
//  2. metadata.uid — historical fallback for bare Job objects that
//     somehow have no spec.selector. Should be vanishingly rare in
//     practice but we'd rather print a working selector than fail
//     hard.
//
// Returns ("", "") only if both the spec.selector path and the UID
// path produce nothing — runPods treats that as an error so we never
// silently degrade to a cross-Job pod listing.
func jobPodSelector(j Job) (selector string, source string) {
	if j.Spec.Selector != nil && len(j.Spec.Selector.MatchLabels) > 0 {
		keys := make([]string, 0, len(j.Spec.Selector.MatchLabels))
		for k := range j.Spec.Selector.MatchLabels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+"="+j.Spec.Selector.MatchLabels[k])
		}
		return strings.Join(parts, ","), "spec.selector.matchLabels"
	}
	if j.Metadata.UID != "" {
		return "controller-uid=" + j.Metadata.UID, "metadata.uid fallback"
	}
	return "", ""
}

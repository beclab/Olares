package workload

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRestartCommand: `olares-cli cluster workload restart
// <ns/name | name> --kind X [-n NS] [--yes] [--concurrency N]`.
//
// SPA-aligned restart (matches confirmHandler2 in
// apps/.../controlHub/pages/ApplicationSpaces/Workloads/Detail.vue):
//
//  1. GET the workload to read spec.selector.matchLabels (the
//     authoritative source — we don't reconstruct it from convention).
//  2. List pods server-side via labelSelector=<rebuilt selector>
//     (`/api/v1/namespaces/<ns>/pods`).
//  3. ConfirmDestructive showing the count and pod names so the user
//     sees exactly which pods are about to be deleted.
//  4. Parallel `DELETE /api/v1/namespaces/<ns>/pods/<name>` with
//     bounded concurrency (default 5).
//
// The controller then recreates each pod from the workload's template.
// We deliberately do NOT use the kubectl-style `restartedAt`
// annotation trick — that would diverge from the SPA's behavior the
// user already knows, and it requires PATCHing the template (which
// some user-facing constraints disallow).
func NewRestartCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace   string
		kindRaw     string
		assumeYes   bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "restart <ns/name | name>",
		Short: "restart a workload by deleting its pods (controller recreates them)",
		Long: `Restart a workload by deleting all of its pods; the controller
then recreates them from the workload template.

Steps:
  1. GET the workload to read spec.selector.matchLabels.
  2. List pods via the rebuilt labelSelector.
  3. Confirm with the operator (showing pod names + count).
  4. Parallel DELETE /api/v1/namespaces/<ns>/pods/<name>, bounded by
     --concurrency (default 5).

This matches the SPA's "Restart" button in the workload detail view
exactly — same selector lookup, same delete-pods semantics. Pods
recreate one by one as the controller observes the deletions, so
expect a brief readiness dip until each replacement is Ready.

Pass --yes to skip the confirmation prompt for scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			plural, err := NormalizeKind(kindRaw)
			if err != nil {
				return err
			}
			if plural == KindAll {
				return fmt.Errorf("--kind must be one of: deployment, statefulset, daemonset (not %q)", kindRaw)
			}
			if concurrency < 1 {
				return fmt.Errorf("--concurrency must be >= 1, got %d", concurrency)
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runRestart(c.Context(), o, ns, name, plural, assumeYes, concurrency)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().IntVar(&concurrency, "concurrency", 5, "max parallel DELETEs (defense against API server hammering)")
	o.AddOutputFlags(cmd)
	return cmd
}

// restartResult is the JSON-mode shape emitted on success. The list
// of pods deleted (and their per-pod success / error) lets scripts
// post-process partial failures (e.g. retry the failed ones).
type restartResult struct {
	Operation string           `json:"operation"`
	Kind      string           `json:"kind"`
	Namespace string           `json:"namespace"`
	Name      string           `json:"name"`
	Selector  string           `json:"selector"`
	Deleted   []podDeleteState `json:"deleted"`
	Failures  int              `json:"failures"`
}

type podDeleteState struct {
	Pod     string `json:"pod"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// minimalPod is just enough of corev1.Pod to populate the delete-list
// confirmation prompt. Re-declared (not shared with cluster/pod) to
// keep this package leaf-independent.
type minimalPod struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
	} `json:"metadata"`
}

func runRestart(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string, assumeYes bool, concurrency int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}

	// 1. GET the workload to read spec.selector.matchLabels.
	var w Workload
	if err := clusterclient.GetK8sObject(ctx, client, buildGetPath(namespace, kindPlural, name), &w); err != nil {
		return fmt.Errorf("get %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
	}
	selector := buildLabelSelector(w.Spec.Selector.MatchLabels)
	if selector == "" {
		return fmt.Errorf("%s %s/%s has no spec.selector.matchLabels — cannot find pods to restart",
			SingularKind(kindPlural), namespace, name)
	}

	// 2. List pods with that selector (K8s native; we don't need the
	// KubeSphere envelope's metrics enrichment for a delete loop).
	q := url.Values{}
	q.Set("labelSelector", selector)
	listPath := fmt.Sprintf("/api/v1/namespaces/%s/pods?%s", url.PathEscape(namespace), q.Encode())
	resp, err := clusterclient.GetK8sList[minimalPod](ctx, client, listPath)
	if err != nil {
		return fmt.Errorf("list pods for %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
	}
	pods := resp.Items
	if len(pods) == 0 {
		// Nothing to restart — surface this as "no-op" rather than
		// silently succeeding so the operator knows the lookup was
		// performed but matched nothing.
		if o.IsJSON() {
			return o.PrintJSON(restartResult{
				Operation: "restart",
				Kind:      SingularKind(kindPlural),
				Namespace: namespace,
				Name:      name,
				Selector:  selector,
				Deleted:   []podDeleteState{},
			})
		}
		if !o.Quiet {
			fmt.Fprintf(os.Stderr, "no pods match selector %q for %s %s/%s — nothing to restart\n",
				selector, SingularKind(kindPlural), namespace, name)
		}
		return nil
	}

	// Stable order so the prompt + JSON output diff cleanly across
	// runs — server returns items in apiserver-internal order which
	// can shift between calls.
	sort.SliceStable(pods, func(i, j int) bool {
		return pods[i].Metadata.Name < pods[j].Metadata.Name
	})
	names := make([]string, 0, len(pods))
	for _, p := range pods {
		names = append(names, p.Metadata.Name)
	}

	// 3. ConfirmDestructive — show count + (truncated) names so the
	// operator can spot a wrong selector before pods start dying.
	preview := strings.Join(names, ", ")
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Restart %s %s/%s by deleting %d pods (%s)? The controller will recreate them",
			SingularKind(kindPlural), namespace, name, len(pods), preview),
		assumeYes); err != nil {
		return err
	}

	// 4. Parallel DELETE bounded by concurrency. We use a buffered
	// semaphore + WaitGroup rather than errgroup because we want to
	// continue past per-pod failures rather than aborting on the
	// first one (errgroup cancels siblings on first error).
	results := make([]podDeleteState, len(pods))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, p := range pods {
		i, p := i, p
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			delPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s",
				url.PathEscape(namespace), url.PathEscape(p.Metadata.Name))
			err := client.DoJSON(ctx, "DELETE", delPath, nil, nil)
			if err != nil {
				results[i] = podDeleteState{Pod: p.Metadata.Name, Success: false, Error: err.Error()}
				return
			}
			results[i] = podDeleteState{Pod: p.Metadata.Name, Success: true}
		}()
	}
	wg.Wait()

	failures := 0
	for _, r := range results {
		if !r.Success {
			failures++
		}
	}

	if o.IsJSON() {
		return o.PrintJSON(restartResult{
			Operation: "restart",
			Kind:      SingularKind(kindPlural),
			Namespace: namespace,
			Name:      name,
			Selector:  selector,
			Deleted:   results,
			Failures:  failures,
		})
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "deleted %d/%d pods for %s %s/%s\n",
			len(results)-failures, len(results), SingularKind(kindPlural), namespace, name)
		for _, r := range results {
			if !r.Success {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", r.Pod, r.Error)
			}
		}
	}
	if failures > 0 {
		return fmt.Errorf("restart %s %s/%s: %d of %d pod deletes failed",
			SingularKind(kindPlural), namespace, name, failures, len(results))
	}
	return nil
}

// buildLabelSelector turns matchLabels into a K8s "key=value,..."
// label selector string with a stable key order. Equivalent to the
// SPA's helper that joins the same map; we do it here because the
// REST endpoint takes a flat string, not a typed selector.
func buildLabelSelector(matchLabels map[string]string) string {
	if len(matchLabels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(matchLabels))
	for k := range matchLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+matchLabels[k])
	}
	return strings.Join(pairs, ",")
}

package job

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

// NewRerunCommand: `olares-cli cluster job rerun <ns/name | name>
// [-n NS] [--yes] [--concurrency N]`.
//
// History note: this verb used to POST to KubeSphere's operations API
// (`/kapis/operations.kubesphere.io/v1alpha2/.../jobs/<name>?
// action=rerun&resourceVersion=<rv>`). Olares's control-hub nginx
// does NOT route that path — the SPA's "Rerun" button is broken for
// the same reason — so the CLI now mimics the same effect
// client-side by deleting the Job's child Pods. The Job controller
// then creates new Pods to satisfy `spec.parallelism` /
// `spec.completions`, which is exactly what KubeSphere's operations
// action did server-side.
//
// Flow (mirrors `cluster workload restart`):
//
//  1. GET the Job to read .metadata.uid + check terminal state.
//  2. List Pods server-side via labelSelector=controller-uid=<uid>
//     (same trick used by `cluster job pods`).
//  3. ConfirmDestructive showing the count + pod names.
//  4. Parallel `DELETE /api/v1/namespaces/<ns>/pods/<name>` bounded
//     by --concurrency (default 5).
//
// Limitation: a Job that has already reached Complete=True or
// Failed=True is terminal — deleting its Pods won't trigger a
// re-execution because the Job controller has stopped scheduling.
// We surface that as a clear error so the user knows to delete-and-
// recreate the Job instead of being silently confused by "0 new
// pods came up".
func NewRerunCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace   string
		assumeYes   bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "rerun <ns/name | name>",
		Short: "rerun one Job by deleting its pods (controller reschedules new attempts)",
		Long: `Rerun one Job by deleting all of its child Pods; the Job
controller then creates new Pods to satisfy spec.parallelism /
spec.completions.

Flow:
  1. GET the Job to read its .metadata.uid and check terminal state.
  2. List Pods via labelSelector=controller-uid=<uid>.
  3. Confirm with the operator (showing pod names + count).
  4. Parallel DELETE /api/v1/namespaces/<ns>/pods/<name>, bounded by
     --concurrency (default 5).

This is a client-side equivalent of KubeSphere's "action=rerun"
operations endpoint — that endpoint is not exposed on Olares
clusters, but deleting the Job's Pods produces the same observable
effect for running Jobs.

Limitation: a Job that has already terminated (Complete=True or
Failed=True) cannot be rerun in-place — the Job controller stops
scheduling Pods once it sees a terminal condition. The verb refuses
that case and asks you to delete-and-recreate the Job instead.

This is mutating: existing Pods are killed and new ones are launched.
Pass --yes to skip the confirmation prompt for scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if concurrency < 1 {
				return fmt.Errorf("--concurrency must be >= 1, got %d", concurrency)
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runRerun(c.Context(), o, ns, name, assumeYes, concurrency)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().IntVar(&concurrency, "concurrency", 5, "max parallel pod DELETEs (defense against apiserver hammering)")
	o.AddOutputFlags(cmd)
	return cmd
}

// rerunResult is the JSON-mode shape emitted on success. The list of
// pods deleted (and their per-pod success / error) lets scripts post-
// process partial failures (e.g. retry the failed ones). Mirrors
// workload.restartResult so consumers reading either verb's JSON
// don't need a separate parser.
type rerunResult struct {
	Operation string           `json:"operation"`
	Namespace string           `json:"namespace"`
	Job       string           `json:"job"`
	UID       string           `json:"uid"`
	Selector  string           `json:"selector"`
	Deleted   []podDeleteState `json:"deleted"`
	Failures  int              `json:"failures"`
}

// podDeleteState mirrors workload.podDeleteState — re-declared (not
// shared with cluster/workload) to keep this package leaf-independent,
// same convention used by workload/restart.go's minimalPod.
type podDeleteState struct {
	Pod     string `json:"pod"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// minimalPod is just enough of corev1.Pod to populate the prompt and
// per-pod result rows. Re-declared (not shared with cluster/pod) to
// keep this package leaf-independent — same convention used by
// workload/restart.go.
type minimalPod struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
	} `json:"metadata"`
}

func runRerun(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, assumeYes bool, concurrency int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. GET the Job. We need its UID for the pod selector, and we
	// also need to inspect status.conditions so we can refuse Jobs
	// that have already terminated (deleting their Pods would not
	// trigger a re-execution — the controller has stopped).
	j, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if j.Metadata.UID == "" {
		return fmt.Errorf("job %s/%s has no metadata.uid — server response missing the field", namespace, name)
	}
	if terminal, reason := jobTerminalState(*j); terminal {
		return fmt.Errorf("job %s/%s has already terminated (%s) — rerun cannot resurrect a terminal Job; delete and recreate the Job instead",
			namespace, name, reason)
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}

	// 2. List Pods owned by this Job using the K8s-native selector
	// the Job controller itself wrote into spec.selector (see
	// pods.go::jobPodSelector for why we don't hardcode
	// controller-uid=<uid> anymore — K8s 1.27+ switched the auto-
	// generated label key, and manualSelector=true Jobs pick their
	// own). If spec.selector is empty we fall back to the legacy
	// UID-based clause so this still works on ancient clusters.
	selector, _ := jobPodSelector(*j)
	if selector == "" {
		return fmt.Errorf("job %s/%s: cannot derive a pod selector (spec.selector empty and metadata.uid blank)",
			namespace, name)
	}
	q := url.Values{}
	q.Set("labelSelector", selector)
	listPath := fmt.Sprintf("/api/v1/namespaces/%s/pods?%s", url.PathEscape(namespace), q.Encode())
	resp, err := clusterclient.GetK8sList[minimalPod](ctx, client, listPath)
	if err != nil {
		return fmt.Errorf("list pods for job %s/%s: %w", namespace, name, err)
	}
	pods := resp.Items
	if len(pods) == 0 {
		return fmt.Errorf("job %s/%s has no child pods (selector %q matched nothing) — nothing to rerun. If the Job has not yet scheduled any pod, wait and try again",
			namespace, name, selector)
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

	// 3. ConfirmDestructive showing the count + (truncated) names so
	// the operator can spot a wrong identity before pods start dying.
	preview := strings.Join(names, ", ")
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Rerun job %s/%s by deleting %d pod(s) (%s)? The Job controller will create new pods to satisfy spec.parallelism / spec.completions",
			namespace, name, len(pods), preview),
		assumeYes); err != nil {
		return err
	}

	// 4. Parallel DELETE bounded by concurrency. We use a buffered
	// semaphore + WaitGroup rather than errgroup because we want to
	// continue past per-pod failures rather than aborting on the
	// first one (errgroup cancels siblings on first error). Same
	// shape as workload/restart.go.
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
		if err := o.PrintJSON(rerunResult{
			Operation: "rerun",
			Namespace: namespace,
			Job:       name,
			UID:       j.Metadata.UID,
			Selector:  selector,
			Deleted:   results,
			Failures:  failures,
		}); err != nil {
			return err
		}
	} else if !o.Quiet {
		fmt.Fprintf(os.Stdout, "deleted %d/%d pod(s) for job %s/%s; controller will reschedule new attempts\n",
			len(results)-failures, len(results), namespace, name)
		for _, r := range results {
			if !r.Success {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", r.Pod, r.Error)
			}
		}
	}
	if failures > 0 {
		return fmt.Errorf("rerun job %s/%s: %d of %d pod deletes failed", namespace, name, failures, len(results))
	}
	return nil
}

// jobTerminalState reports whether the given Job has reached a
// terminal phase (Complete=True or Failed=True). The Job controller
// stops scheduling once a terminal condition is set, so a client-side
// "rerun" (delete pods) would be a no-op — better to refuse loudly.
//
// We rely on status.conditions rather than status.completionTime
// because the K8s controller sets the Complete condition before the
// timestamp in some pathological races, and Failed is never reflected
// by completionTime at all. The reason string is propagated to the
// error message so the user can tell "Complete" from "BackoffLimit
// Exceeded" without re-reading the YAML.
func jobTerminalState(j Job) (bool, string) {
	for _, c := range j.Status.Conditions {
		if c.Status != "True" {
			continue
		}
		switch strings.ToLower(c.Type) {
		case "complete":
			return true, "Complete"
		case "failed":
			if c.Reason != "" {
				return true, "Failed: " + c.Reason
			}
			return true, "Failed"
		}
	}
	return false, ""
}

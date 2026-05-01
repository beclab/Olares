package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewStatusCommand: `olares-cli cluster application status <namespace>
// [--watch] [--interval D] [--events N] [-o table|json]`.
//
// CLI-original aggregation — the ControlHub SPA does NOT have a single
// per-namespace health dashboard. Instead it spreads the same
// information across separate Pods / Workloads tabs. This verb fans
// out the underlying GETs (workloads × 3 kinds + pods + events) in
// parallel and renders one consolidated view, so a script answer
// "is application X healthy?" without crossing five terminal screens.
//
// The fan-out goes through the same endpoints `cluster workload list`,
// `cluster pod list`, and `cluster pod events` use — server-side
// scoping is therefore identical (no client-side filtering / role
// gating; the security model stays "server decides").
//
// --watch repaints (table) or emits one JSONL object (json) on each
// tick. Errors on any single fan-out lane don't abort the snapshot:
// we surface a "(failed: ...)" placeholder so users still see what
// IS available even when one of the three fetches fails.
func NewStatusCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		watch    bool
		interval time.Duration
		eventsN  int
	)
	cmd := &cobra.Command{
		Use:   "status <namespace>",
		Short: "consolidated workloads/pods/events overview for one application space",
		Long: `One-shot health overview for one application space (Namespace).

Fans out three GETs in parallel:
  1. workloads list (deployments + statefulsets + daemonsets)
       /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/<kind>
  2. pods list
       /kapis/resources.kubesphere.io/v1alpha3/namespaces/<ns>/pods
  3. recent events
       /api/v1/namespaces/<ns>/events

Then renders three sections: per-kind workload READY counts, pod
phase buckets (Running / Pending / Succeeded / Failed / Unknown),
and the most recent --events events sorted by lastTimestamp desc
(default 5).

--watch repaints (table) or emits one JSON object per tick (JSONL)
on --interval (default 2s) until Ctrl-C. Same signal + 5-error
tolerance plumbing as the rest of the cluster --watch verbs.

There is NO equivalent SPA endpoint that returns this aggregated
shape — the ControlHub UI spreads the same information across
separate Pods / Workloads / Events tabs. CLI just stitches them.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns := strings.TrimSpace(args[0])
			if ns == "" {
				return fmt.Errorf("namespace must be non-empty")
			}
			// Refuse --interval without --watch: the flag only governs
			// the polling cadence inside the watch loop, so silently
			// ignoring it on a one-shot snapshot would hide the misuse.
			if c.Flags().Changed("interval") && !watch {
				return fmt.Errorf("--interval requires --watch")
			}
			return runStatus(c.Context(), o, ns, watch, interval, eventsN)
		},
	}
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "re-fetch and re-render until interrupted (Ctrl-C to stop)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval when --watch is set")
	cmd.Flags().IntVar(&eventsN, "events", 5, "number of recent events to include in the snapshot (0 = none)")
	o.AddOutputFlags(cmd)
	return cmd
}

// readyTotal is the canonical per-kind workload pair used in the JSON
// output (and as the input to the table-mode renderer). Two ints are
// enough to convey "X of Y workloads of this kind have all their
// replicas / pods ready".
type readyTotal struct {
	Ready int `json:"ready"`
	Total int `json:"total"`
}

// podPhaseCounts buckets the namespace's pods by Phase. Mirrors
// kubectl's "STATUS" column except aggregated; "Total" is provided
// up-front so consumers don't have to sum the buckets.
type podPhaseCounts struct {
	Running   int `json:"running"`
	Pending   int `json:"pending"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Unknown   int `json:"unknown"`
	Total     int `json:"total"`
}

// appStatus is the JSON-mode shape (and internal carrier between
// fetch + render). Errors per fan-out lane stay on the snapshot so
// the renderer can surface them inline rather than aborting the
// whole tick.
type appStatus struct {
	Namespace string                `json:"namespace"`
	FetchedAt string                `json:"fetchedAt"`
	Workloads map[string]readyTotal `json:"workloads"`
	Pods      podPhaseCounts        `json:"pods"`
	Events    []pod.Event           `json:"events"`

	WorkloadsErr string `json:"workloadsError,omitempty"`
	PodsErr      string `json:"podsError,omitempty"`
	EventsErr    string `json:"eventsError,omitempty"`
}

// fetchStatus performs the three parallel GETs and assembles a
// snapshot. Per-lane errors are stashed on the snapshot rather than
// returned so a partial view is still useful (e.g. events 5xx
// shouldn't black out the workloads/pods sections).
//
// Implementation note: we deliberately do NOT use errgroup.WithContext
// here. errgroup cancels its context the moment any goroutine returns
// an error, which would cascade-cancel the in-flight HTTP requests on
// the OTHER lanes and turn a single-lane failure into "everything
// failed". That contradicts the per-lane partial-failure intent above.
// Instead each lane runs against the parent ctx (so user-level Ctrl-C
// still cancels everything) and writes its own success/error slot.
func fetchStatus(ctx context.Context, c *clusterclient.Client, namespace string, eventsN int) appStatus {
	out := appStatus{
		Namespace: namespace,
		FetchedAt: time.Now().Format(time.RFC3339),
		Workloads: map[string]readyTotal{},
	}

	// Per-kind workload fan-out (3 parallel GETs). We fetch all three
	// concurrently so the polling interval bound is "the slowest one"
	// rather than "their sum".
	type wlBucket struct {
		kindPlural string
		bucket     readyTotal
	}
	wlResults := make(chan wlBucket, len(workload.SupportedKinds))

	var (
		wg         sync.WaitGroup
		wlErrs     []string
		wlErrsMu   sync.Mutex
		podsBucket podPhaseCounts
		podsErr    error
		events     []pod.Event
		eventsErr  error
	)

	for _, k := range workload.SupportedKinds {
		k := k
		wg.Add(1)
		go func() {
			defer wg.Done()
			plural, _ := workload.NormalizeKind(k)
			path := buildNamespaceKindPath(namespace, plural)
			resp, err := clusterclient.GetKubeSphereList[workload.Workload](ctx, c, path)
			if err != nil {
				wlErrsMu.Lock()
				wlErrs = append(wlErrs, fmt.Sprintf("%s: %v", plural, err))
				wlErrsMu.Unlock()
				return
			}
			ready, total := 0, len(resp.Items)
			for _, w := range resp.Items {
				if isWorkloadReady(w, plural) {
					ready++
				}
			}
			wlResults <- wlBucket{kindPlural: plural, bucket: readyTotal{Ready: ready, Total: total}}
		}()
	}

	// Pods. Single goroutine writer for podsBucket / podsErr — no mutex.
	wg.Add(1)
	go func() {
		defer wg.Done()
		path := "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
			url.PathEscape(namespace) + "/pods"
		resp, err := clusterclient.GetKubeSphereList[pod.Pod](ctx, c, path)
		if err != nil {
			podsErr = fmt.Errorf("pods: %w", err)
			return
		}
		podsBucket = bucketPods(resp.Items)
	}()

	// Events. Fetch namespace-wide events; sort+truncate client-side
	// (same approach as `cluster pod events` for portability across
	// kube-apiserver versions). Skipped entirely when --events 0 to
	// save a roundtrip per tick.
	if eventsN > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path := fmt.Sprintf("/api/v1/namespaces/%s/events", url.PathEscape(namespace))
			resp, err := clusterclient.GetK8sList[pod.Event](ctx, c, path)
			if err != nil {
				eventsErr = fmt.Errorf("events: %w", err)
				return
			}
			events = mostRecentEvents(resp.Items, eventsN)
		}()
	}

	wg.Wait()
	close(wlResults)

	for r := range wlResults {
		out.Workloads[workload.SingularKind(r.kindPlural)] = r.bucket
	}
	if len(wlErrs) > 0 {
		out.WorkloadsErr = strings.Join(wlErrs, "; ")
	}
	if podsErr != nil {
		out.PodsErr = podsErr.Error()
	}
	if eventsErr != nil {
		out.EventsErr = eventsErr.Error()
	}
	out.Pods = podsBucket
	out.Events = events
	return out
}

// buildNamespaceKindPath constructs the per-kind, per-namespace
// KubeSphere list URL. Kept tiny because `cluster workload list`
// already has its own buildListPath (private) and we don't want
// status.go reaching into another file's helpers; same path though.
func buildNamespaceKindPath(namespace, kindPlural string) string {
	return "/kapis/resources.kubesphere.io/v1alpha3/namespaces/" +
		url.PathEscape(namespace) + "/" + kindPlural
}

// isWorkloadReady is the per-row "all replicas/pods ready" predicate
// we use to populate the READY column counts. Looser than the
// rollout-status convergence rule (we don't require observedGeneration
// to match — the application overview cares about steady-state, not
// rollout progress). For DaemonSets, "ready" means every desired
// pod is ready; for Deployment/StatefulSet, every desired replica.
func isWorkloadReady(w workload.Workload, kindPlural string) bool {
	switch kindPlural {
	case "deployments", "statefulsets":
		desired := 0
		if w.Spec.Replicas != nil {
			desired = *w.Spec.Replicas
		} else {
			desired = w.Status.Replicas
		}
		// scaled-to-zero counts as "ready" — there's nothing to fail.
		if desired == 0 {
			return true
		}
		return w.Status.ReadyReplicas == desired
	case "daemonsets":
		desired := w.Status.DesiredNumberScheduled
		if desired == 0 {
			return true
		}
		return w.Status.NumberReady == desired
	}
	return false
}

// bucketPods sorts pods into the five canonical Phase buckets used by
// the table-mode "Pods:" line. Unknown / empty Phase values land in
// the Unknown bucket (rather than silently dropping) so the Total
// always equals the sum.
func bucketPods(items []pod.Pod) podPhaseCounts {
	out := podPhaseCounts{Total: len(items)}
	for _, p := range items {
		switch p.Status.Phase {
		case "Running":
			out.Running++
		case "Pending":
			out.Pending++
		case "Succeeded":
			out.Succeeded++
		case "Failed":
			out.Failed++
		default:
			out.Unknown++
		}
	}
	return out
}

// mostRecentEvents returns the last N events sorted newest-first by
// lastTimestamp. Falls back to firstTimestamp / creationTimestamp
// when lastTimestamp is empty (mirrors `cluster pod events`'s
// eventSortKey semantics — same precedence).
func mostRecentEvents(events []pod.Event, n int) []pod.Event {
	sort.SliceStable(events, func(i, j int) bool {
		return eventTimeKey(events[i]).After(eventTimeKey(events[j]))
	})
	if n > 0 && len(events) > n {
		events = events[:n]
	}
	return events
}

func eventTimeKey(e pod.Event) time.Time {
	for _, ts := range []string{e.LastTimestamp, e.FirstTimestamp, e.Metadata.CreationTimestamp} {
		if ts == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t
		}
	}
	return time.Time{}
}

func runStatus(ctx context.Context, o *clusteropts.ClusterOptions, namespace string, watch bool, interval time.Duration, eventsN int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := o.Prepare()
	if err != nil {
		return err
	}

	render := func(s appStatus) error {
		if o.IsJSON() {
			return o.PrintJSON(s)
		}
		if o.Quiet {
			return nil
		}
		return renderStatusTable(s)
	}

	if !watch {
		s := fetchStatus(ctx, c, namespace, eventsN)
		return render(s)
	}

	if interval <= 0 {
		interval = 2 * time.Second
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	clearable := !o.IsJSON() && term.IsTerminal(int(os.Stdout.Fd()))
	first := true
	for {
		if !first {
			if err := sleepCtx(ctx, interval); err != nil {
				return nil
			}
		}
		first = false

		s := fetchStatus(ctx, c, namespace, eventsN)
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		if clearable {
			fmt.Fprint(os.Stdout, "\x1b[2J\x1b[H")
			fmt.Fprintf(os.Stdout, "Watching application %q every %s (Ctrl-C to exit). Last fetch: %s\n\n",
				namespace, interval, s.FetchedAt)
		}
		if err := render(s); err != nil {
			return err
		}
	}
}

// renderStatusTable lays out the three sections in a fixed order
// (Workloads, Pods, Recent Events) so users get a stable visual
// mental model across runs. Empty sections still print a header so
// the layout doesn't shift between ticks under --watch.
func renderStatusTable(s appStatus) error {
	fmt.Fprintf(os.Stdout, "Application: %s\n", s.Namespace)
	fmt.Fprintf(os.Stdout, "Fetched:     %s\n\n", s.FetchedAt)

	// Workloads section.
	fmt.Fprintln(os.Stdout, "Workloads:")
	if s.WorkloadsErr != "" {
		fmt.Fprintf(os.Stdout, "  (failed: %s)\n", s.WorkloadsErr)
	} else if len(s.Workloads) == 0 {
		fmt.Fprintln(os.Stdout, "  (none)")
	} else {
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  KIND\tREADY")
		// Stable kind order so output diffs cleanly across runs.
		var kinds []string
		for k := range s.Workloads {
			kinds = append(kinds, k)
		}
		sort.Strings(kinds)
		for _, k := range kinds {
			b := s.Workloads[k]
			fmt.Fprintf(tw, "  %s\t%d/%d\n", k, b.Ready, b.Total)
		}
		tw.Flush()
	}

	// Pods section.
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Pods:")
	if s.PodsErr != "" {
		fmt.Fprintf(os.Stdout, "  (failed: %s)\n", s.PodsErr)
	} else if s.Pods.Total == 0 {
		fmt.Fprintln(os.Stdout, "  (none)")
	} else {
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  PHASE\tCOUNT")
		fmt.Fprintf(tw, "  Running\t%d\n", s.Pods.Running)
		fmt.Fprintf(tw, "  Pending\t%d\n", s.Pods.Pending)
		fmt.Fprintf(tw, "  Succeeded\t%d\n", s.Pods.Succeeded)
		fmt.Fprintf(tw, "  Failed\t%d\n", s.Pods.Failed)
		fmt.Fprintf(tw, "  Unknown\t%d\n", s.Pods.Unknown)
		fmt.Fprintf(tw, "  TOTAL\t%d\n", s.Pods.Total)
		tw.Flush()
	}

	// Events section.
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Recent Events:")
	if s.EventsErr != "" {
		fmt.Fprintf(os.Stdout, "  (failed: %s)\n", s.EventsErr)
	} else if len(s.Events) == 0 {
		fmt.Fprintln(os.Stdout, "  (none)")
	} else {
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE")
		now := time.Now()
		for _, e := range s.Events {
			ts := e.LastTimestamp
			if ts == "" {
				ts = e.Metadata.CreationTimestamp
			}
			obj := e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\n",
				clusteropts.Age(ts, now)+" ago",
				clusteropts.DashIfEmpty(e.Type),
				clusteropts.DashIfEmpty(e.Reason),
				clusteropts.DashIfEmpty(obj),
				clusteropts.DashIfEmpty(e.Message),
			)
		}
		tw.Flush()
	}
	return nil
}

// sleepCtx is the cancellation-aware sleep used by the polling loop.
// Re-declared here to keep the application package independent of
// the workload / pod packages' equivalents.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

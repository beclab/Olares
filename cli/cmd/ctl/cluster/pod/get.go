package pod

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewGetCommand: `olares-cli cluster pod get <ns/name> [-o table|json]`.
//
// Calls SPA's getPodDetail
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:58):
// `/api/v1/namespaces/<ns>/pods/<name>` — K8s native shape.
//
// Pod identity is taken as a single positional `<namespace>/<name>`
// argument so the verb composes nicely with shell pipelines. We
// support `-n <ns> <name>` as well for symmetry with kubectl: when -n
// is supplied, the positional argument is the bare pod name.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		watch     bool
		interval  time.Duration
	)

	cmd := &cobra.Command{
		Use:   "get <ns/name | name>",
		Short: "show one pod's details (K8s native shape)",
		Long: `Show one pod's full detail.

Identity may be passed as a single "<namespace>/<name>" positional or
as a bare "<name>" with -n <namespace>. Without -n the positional
form is required so we don't guess a namespace.

In table mode, the output is a vertical key/value summary plus per-
container rows. In json mode the response body is forwarded verbatim
(no envelope wrapping) so the shape matches kube-apiserver exactly.

--watch repeats the GET on --interval (default 2s). In table mode
the screen is cleared and redrawn between ticks (when stdout is a
terminal); in JSON mode each tick emits one JSON object on its own
line (JSONL stream). Ctrl-C exits cleanly.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			if watch {
				return runGetWatch(c.Context(), o, ns, name, interval)
			}
			return runGet(c.Context(), o, ns, name)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "re-fetch and re-render until interrupted (Ctrl-C to stop)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval when --watch is set")
	o.AddOutputFlags(cmd)
	return cmd
}

// Get is the exported single-pod fetcher used by sibling packages
// (cluster/container) that need to project pod contents — containers
// list, env vars, etc. Returns the typed Pod without rendering.
//
// Same HTTP path as the `cluster pod get` verb; same server-side
// scoping rules apply (a 404 means the pod doesn't exist OR your
// token can't see it).
func Get(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) (*Pod, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s",
		url.PathEscape(namespace), url.PathEscape(name))
	var p Pod
	if err := clusterclient.GetK8sObject(ctx, client, path, &p); err != nil {
		return nil, fmt.Errorf("get pod %s/%s: %w", namespace, name, err)
	}
	return &p, nil
}

func runGet(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string) error {
	p, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	if o.IsJSON() {
		return o.PrintJSON(*p)
	}
	return renderGetTable(*p)
}

// runGetWatch is the --watch variant of runGet: same fetch, repeated
// on `interval` until the caller interrupts.
//
// Table mode: clear-screen + redraw on each tick when stdout is a TTY
// so the output is readable as a live dashboard. When stdout is NOT a
// TTY (piped to a file / grep), we skip the clear and just stream
// repeated tables — that keeps `--watch | tee` useful without
// littering the file with ANSI escapes.
//
// JSON mode: emit one JSON object per tick (JSONL stream). No clear,
// no separators between objects beyond the trailing newline that
// json.Encoder adds — matches the convention `kubectl get -o json
// --watch` consumers (jq -c, tooling, etc.) already expect.
//
// Transient error policy mirrors the logs polling loop: tolerate up
// to 5 consecutive HTTP errors before giving up so a transient 5xx /
// network blip doesn't kill an otherwise-healthy `--watch`. Auth
// failures (401/403) and other terminal errors propagate immediately
// — they won't fix themselves on the next tick.
func runGetWatch(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, interval time.Duration) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	clearable := !o.IsJSON() && term.IsTerminal(int(os.Stdout.Fd()))
	consecErr := 0
	first := true
	for {
		if !first {
			if err := sleepCtx(ctx, interval); err != nil {
				return nil
			}
		}
		first = false

		p, err := Get(ctx, o, namespace, name)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			consecErr++
			o.Info("watch: failed to fetch pod (retry %d): %v", consecErr, err)
			if consecErr >= 5 {
				return fmt.Errorf("watch: aborted after %d consecutive errors: %w", consecErr, err)
			}
			continue
		}
		consecErr = 0

		if o.IsJSON() {
			if err := o.PrintJSON(*p); err != nil {
				return err
			}
			continue
		}

		if clearable {
			// "\x1b[2J" clears the screen, "\x1b[H" homes the cursor
			// to top-left. Same convention `top` / `htop` use; works
			// on every terminal Cursor / iTerm / GNOME Terminal /
			// Linux console emit. We deliberately don't gate on
			// $TERM — any TTY honors these two sequences.
			fmt.Fprint(os.Stdout, "\x1b[2J\x1b[H")
			fmt.Fprintf(os.Stdout, "Watching pod %s/%s every %s (Ctrl-C to exit). Last fetch: %s\n\n",
				namespace, name, interval, time.Now().Format(time.RFC3339))
		}
		if err := renderGetTable(*p); err != nil {
			return err
		}
	}
}

func renderGetTable(p Pod) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Name:\t%s\n", p.Metadata.Name)
	fmt.Fprintf(w, "Namespace:\t%s\n", clusteropts.DashIfEmpty(p.Metadata.Namespace))
	fmt.Fprintf(w, "Node:\t%s\n", clusteropts.DashIfEmpty(p.Spec.NodeName))
	fmt.Fprintf(w, "Status:\t%s\n", clusteropts.DashIfEmpty(p.statusReason()))
	fmt.Fprintf(w, "Phase:\t%s\n", clusteropts.DashIfEmpty(p.Status.Phase))
	fmt.Fprintf(w, "Pod IP:\t%s\n", clusteropts.DashIfEmpty(p.Status.PodIP))
	fmt.Fprintf(w, "Host IP:\t%s\n", clusteropts.DashIfEmpty(p.Status.HostIP))
	fmt.Fprintf(w, "Ready:\t%s\n", p.readyCount())
	fmt.Fprintf(w, "Restarts:\t%d\n", p.totalRestarts())
	fmt.Fprintf(w, "QoS:\t%s\n", clusteropts.DashIfEmpty(p.Status.QOSClass))
	fmt.Fprintf(w, "Service Account:\t%s\n", clusteropts.DashIfEmpty(p.Spec.ServiceAccount))
	if !p.Spec.HostNetwork {
		fmt.Fprintf(w, "Host Network:\tfalse\n")
	} else {
		fmt.Fprintf(w, "Host Network:\ttrue\n")
	}
	fmt.Fprintf(w, "Created:\t%s\n", clusteropts.DashIfEmpty(p.Metadata.CreationTimestamp))
	fmt.Fprintf(w, "Age:\t%s\n", p.age(time.Now()))

	// Owner references — surface "controlled by" so users can pivot
	// to `cluster workload get` when they wonder why a pod they
	// deleted came back.
	if len(p.Metadata.OwnerReferences) > 0 {
		var owners []string
		for _, o := range p.Metadata.OwnerReferences {
			lbl := o.Kind + "/" + o.Name
			if !o.Controller {
				lbl += " (non-controller)"
			}
			owners = append(owners, lbl)
		}
		fmt.Fprintf(w, "Controlled By:\t%s\n", strings.Join(owners, ", "))
	}

	// Conditions — abbreviated to type=status (reason if not empty).
	// Full Reason/Message stays in --output json so users have the
	// full story when needed.
	if len(p.Status.Conditions) > 0 {
		var lines []string
		for _, c := range p.Status.Conditions {
			s := c.Type + "=" + c.Status
			if c.Reason != "" {
				s += " (" + c.Reason + ")"
			}
			lines = append(lines, s)
		}
		fmt.Fprintf(w, "Conditions:\t%s\n", strings.Join(lines, ", "))
	}

	// Container summary — one row per container so users can see
	// image / restart count / ready state at a glance without diving
	// into json. spec.containers is the source of truth for the set;
	// status.containerStatuses provides the runtime overlay.
	if len(p.Spec.Containers) > 0 {
		w.Flush()
		fmt.Fprintln(os.Stdout)

		cw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer cw.Flush()
		fmt.Fprintln(cw, "CONTAINER\tIMAGE\tREADY\tRESTARTS\tSTATE")

		statusByName := map[string]PodContainerStatus{}
		for _, cs := range p.Status.ContainerStatuses {
			statusByName[cs.Name] = cs
		}
		for _, c := range p.Spec.Containers {
			cs, ok := statusByName[c.Name]
			ready := "-"
			restarts := "-"
			state := "-"
			if ok {
				if cs.Ready {
					ready = "true"
				} else {
					ready = "false"
				}
				restarts = fmt.Sprintf("%d", cs.RestartCount)
				state = describeContainerState(cs.State)
			}
			fmt.Fprintf(cw, "%s\t%s\t%s\t%s\t%s\n", c.Name, c.Image, ready, restarts, state)
		}
	}

	return nil
}

// describeContainerState turns a containerStatus.state map into a
// short label suitable for a single column. Mirrors `kubectl describe`'s
// State block in spirit but compressed to one line.
func describeContainerState(state map[string]interface{}) string {
	if state == nil {
		return "-"
	}
	if _, ok := state["running"]; ok {
		return "Running"
	}
	if w, ok := state["waiting"].(map[string]interface{}); ok {
		if reason, ok := w["reason"].(string); ok && reason != "" {
			return "Waiting (" + reason + ")"
		}
		return "Waiting"
	}
	if t, ok := state["terminated"].(map[string]interface{}); ok {
		reason, _ := t["reason"].(string)
		ec, _ := t["exitCode"].(float64)
		if reason != "" {
			return fmt.Sprintf("Terminated (%s, exit %d)", reason, int(ec))
		}
		return fmt.Sprintf("Terminated (exit %d)", int(ec))
	}
	return "-"
}

package workload

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRolloutStatusCommand: `olares-cli cluster workload rollout-status
// <ns/name | name> --kind X [-n NS] [--watch] [--interval D] [--timeout D]`.
//
// Polls the same K8s native endpoint `cluster workload get` uses
// (`/apis/apps/v1/namespaces/<ns>/<kind>/<name>`) and applies a
// kind-aware convergence rule that mirrors `kubectl rollout status`:
//
//   - Deployment / StatefulSet: observedGeneration == metadata.generation
//     AND updatedReplicas == spec.replicas AND readyReplicas == spec.replicas
//   - DaemonSet:                observedGeneration == metadata.generation
//     AND updatedNumberScheduled == desiredNumberScheduled
//     AND numberReady == desiredNumberScheduled
//
// Without --watch: one GET, prints a one-line status, exits 0 if
// converged or 2 if not — handy for shell scripts that just want a
// yes/no answer.
//
// With --watch: re-poll on --interval (default 2s) until converged,
// --timeout elapses (default 10m), or Ctrl-C. Same signal +
// 5-error tolerance plumbing as `cluster pod logs --follow`. Each
// state change emits one line (table mode) or one JSON object
// (JSONL mode); steady ticks are silent so users only see real
// progress.
func NewRolloutStatusCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
		watch     bool
		interval  time.Duration
		timeout   time.Duration
	)
	cmd := &cobra.Command{
		Use:   "rollout-status <ns/name | name>",
		Short: "report whether a workload's rollout has converged (kubectl rollout status)",
		Long: `Report whether a Deployment / StatefulSet / DaemonSet has converged
to its desired state.

Convergence rule (kind-aware, mirrors kubectl rollout status):
  Deployment / StatefulSet  observedGeneration == metadata.generation
                            AND updatedReplicas == spec.replicas
                            AND readyReplicas   == spec.replicas
  DaemonSet                 observedGeneration == metadata.generation
                            AND updatedNumberScheduled == desiredNumberScheduled
                            AND numberReady             == desiredNumberScheduled

Without --watch: one GET, prints a single status line, exits 0 if
converged or 2 if not. Useful in scripts:

  olares-cli cluster workload rollout-status foo/bar --kind deploy && echo ok

With --watch: re-poll on --interval (default 2s) until converged,
--timeout (default 10m) elapses, or Ctrl-C. Each state change emits
one line in table mode or one JSON object in JSONL mode; steady
ticks are silent so the output matches real progress 1:1.
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
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runRolloutStatus(c.Context(), o, ns, name, plural, watch, interval, timeout)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "poll until converged or interrupted (Ctrl-C to stop)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval when --watch is set")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	o.AddOutputFlags(cmd)
	return cmd
}

// rolloutSnapshot is the JSONL element emitted in --output json mode
// (one per tick on state change). Exposed as a struct rather than
// inlined so consumers can rely on a stable field set.
type rolloutSnapshot struct {
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Ready      string `json:"ready"`
	Updated    string `json:"updated,omitempty"`
	Generation string `json:"generation,omitempty"`
	Phase      string `json:"phase"`
	Converged  bool   `json:"converged"`
	FetchedAt  string `json:"fetchedAt"`
}

// converged reports whether the workload's status has caught up with
// the spec for its kind. Mirrors `kubectl rollout status`'s success
// definition rather than the looser "ready==desired" the list table
// uses, so a rollout that's mid-flight (observedGeneration lagging)
// is still treated as in-progress even when readyReplicas already
// matches.
func converged(w Workload, kindPlural string) bool {
	if w.Metadata.Generation > 0 && w.Status.ObservedGeneration < w.Metadata.Generation {
		return false
	}
	switch kindPlural {
	case "deployments", "statefulsets":
		desired := 0
		if w.Spec.Replicas != nil {
			desired = *w.Spec.Replicas
		} else {
			desired = w.Status.Replicas
		}
		return w.Status.UpdatedReplicas == desired && w.Status.ReadyReplicas == desired
	case "daemonsets":
		desired := w.Status.DesiredNumberScheduled
		return w.Status.UpdatedNumberScheduled == desired && w.Status.NumberReady == desired
	}
	return false
}

// snapshot builds the per-tick rolloutSnapshot used in both table-line
// rendering (via line()) and JSONL output. Kept tiny so renderers stay
// dumb projections of the wire shape.
func snapshot(w Workload, kindPlural string, now time.Time) rolloutSnapshot {
	desired := 0
	updated := 0
	switch kindPlural {
	case "deployments", "statefulsets":
		if w.Spec.Replicas != nil {
			desired = *w.Spec.Replicas
		} else {
			desired = w.Status.Replicas
		}
		updated = w.Status.UpdatedReplicas
	case "daemonsets":
		desired = w.Status.DesiredNumberScheduled
		updated = w.Status.UpdatedNumberScheduled
	}
	ready := fmt.Sprintf("%d/%d", w.Status.ReadyReplicas, desired)
	if kindPlural == "daemonsets" {
		ready = fmt.Sprintf("%d/%d", w.Status.NumberReady, desired)
	}
	c := converged(w, kindPlural)
	phase := "rolling out"
	if c {
		phase = "converged"
	}
	gen := ""
	if w.Metadata.Generation > 0 {
		gen = fmt.Sprintf("%d/%d observed", w.Status.ObservedGeneration, w.Metadata.Generation)
	}
	return rolloutSnapshot{
		Kind:       SingularKind(kindPlural),
		Namespace:  w.Metadata.Namespace,
		Name:       w.Metadata.Name,
		Ready:      ready,
		Updated:    fmt.Sprintf("%d/%d updated", updated, desired),
		Generation: gen,
		Phase:      phase,
		Converged:  c,
		FetchedAt:  now.Format(time.RFC3339),
	}
}

// stateKey is the change-detection key used to decide whether to emit
// a new line in --watch mode. Two consecutive snapshots with the
// same key are squelched so output stays proportional to real
// progress (mirrors market/watch.go's "only emit on change" policy).
func stateKey(s rolloutSnapshot) string {
	return s.Phase + "|" + s.Ready + "|" + s.Updated + "|" + s.Generation
}

// line is the human-readable table-mode rendering of one snapshot.
// Kept intentionally compact (one line, fixed prefix) so it tails
// nicely under --watch and slots into shell pipelines.
func (s rolloutSnapshot) line() string {
	parts := []string{s.Ready + " ready", s.Updated}
	if s.Generation != "" {
		parts = append(parts, "generation "+s.Generation)
	}
	parts = append(parts, s.Phase)
	out := fmt.Sprintf(`%s "%s/%s" rollout: `, s.Kind, s.Namespace, s.Name)
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

func runRolloutStatus(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string, watch bool, interval, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	path := buildGetPath(namespace, kindPlural, name)

	fetch := func(ctx context.Context) (rolloutSnapshot, error) {
		var w Workload
		if err := clusterclient.GetK8sObject(ctx, client, path, &w); err != nil {
			return rolloutSnapshot{}, fmt.Errorf("get %s %s/%s: %w", SingularKind(kindPlural), namespace, name, err)
		}
		// Stamp namespace/name from the call context — the server
		// MAY return them empty in pathological cases; the user
		// asked for this identity.
		if w.Metadata.Namespace == "" {
			w.Metadata.Namespace = namespace
		}
		if w.Metadata.Name == "" {
			w.Metadata.Name = name
		}
		return snapshot(w, kindPlural, time.Now()), nil
	}

	if !watch {
		s, err := fetch(ctx)
		if err != nil {
			return err
		}
		if o.IsJSON() {
			if err := o.PrintJSON(s); err != nil {
				return err
			}
		} else if !o.Quiet {
			fmt.Fprintln(os.Stdout, s.line())
		}
		if !s.Converged {
			// Non-zero exit so scripts can branch without parsing
			// the line. Sentinel keeps cobra from re-printing the
			// error (we already rendered the snapshot).
			return clusteropts.ErrReported
		}
		return nil
	}

	// --watch path. signal.NotifyContext mirrors the pattern shared
	// by `pod logs --follow` and `pod get --watch` — Ctrl-C exits
	// nil so scripts that interrupt voluntarily don't get a
	// non-zero status from us.
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var deadline <-chan time.Time
	if timeout > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		deadline = t.C
	}

	consecErr := 0
	var lastKey string
	first := true
	for {
		select {
		case <-deadline:
			return fmt.Errorf("rollout-status: timed out after %s without convergence", timeout)
		default:
		}
		if !first {
			if err := sleepCtx(ctx, interval); err != nil {
				return nil
			}
		}
		first = false

		s, err := fetch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			consecErr++
			o.Info("rollout-status: failed to fetch (retry %d): %v", consecErr, err)
			if consecErr >= 5 {
				return fmt.Errorf("rollout-status: aborted after %d consecutive errors: %w", consecErr, err)
			}
			continue
		}
		consecErr = 0

		key := stateKey(s)
		if key != lastKey {
			lastKey = key
			if o.IsJSON() {
				if err := o.PrintJSON(s); err != nil {
					return err
				}
			} else if !o.Quiet {
				fmt.Fprintln(os.Stdout, s.line())
			}
		}
		if s.Converged {
			return nil
		}
	}
}

// sleepCtx is the cancellation-aware sleep used by the polling loop.
// Re-declared here (rather than shared with pod/logs.go::sleepCtx) so
// the workload package stays compile-time independent of the pod
// package — the helper is trivial enough that duplication beats a
// shared-utility crate.
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

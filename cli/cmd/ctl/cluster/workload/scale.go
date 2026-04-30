package workload

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewScaleCommand: `olares-cli cluster workload scale <ns/name | name>
// --kind X --replicas N [-n NS] [--watch] [--interval D] [--timeout D]
// [--yes]`.
//
// PATCHes the workload with `{"spec":{"replicas":N}}` using
// Content-Type `application/merge-patch+json`. Same plumbing the SPA's
// patchWorkloadsControler uses
// (apps/.../controlPanelCommon/network/index.ts:372).
//
// Confirmation:
//
//   - --replicas=0 (scale-to-zero / "stop") triggers ConfirmDestructive
//     because it pauses traffic the same way `workload stop` does.
//     Other replica counts are reversible and silent.
//
// --watch chains into the rollout-status convergence loop so users get
// "scaled and waited for ready" in one command. Skipped for
// DaemonSets (which can't be scaled — server would reject the PATCH
// with 422 anyway, but we surface a clearer error before the wire).
func NewScaleCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		kindRaw   string
		replicas  int
		watch     bool
		interval  time.Duration
		timeout   time.Duration
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "scale <ns/name | name>",
		Short: "scale a Deployment / StatefulSet to N replicas",
		Long: `Scale a Deployment or StatefulSet to --replicas N.

PATCHes ` + "`/apis/apps/v1/namespaces/<ns>/<kind>/<name>`" + ` with
` + "`{\"spec\":{\"replicas\":N}}`" + ` (Content-Type
` + "`application/merge-patch+json`" + `).

DaemonSets are NOT scalable — kind=daemonset (or its aliases) is
rejected up-front with a clear error rather than letting the server
return a 422.

--replicas=0 (scale-to-zero) triggers a confirmation prompt because
it pauses traffic — same as ` + "`workload stop`" + `. Other replica counts
are reversible and don't prompt.

--watch chains into ` + "`rollout-status --watch`" + `: after the PATCH
succeeds we poll the same endpoint until the controller's
observedGeneration / updatedReplicas / readyReplicas all match the
new spec. --interval / --timeout drive that loop.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			plural, err := NormalizeKind(kindRaw)
			if err != nil {
				return err
			}
			if plural == KindAll {
				return fmt.Errorf("--kind must be one of: deployment, statefulset (not %q)", kindRaw)
			}
			if plural == "daemonsets" {
				return fmt.Errorf("DaemonSets are not scalable (replicas don't apply); use `cluster workload restart` to roll their pods")
			}
			if replicas < 0 {
				return fmt.Errorf("--replicas must be >= 0, got %d", replicas)
			}
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runScale(c.Context(), o, ns, name, plural, replicas, watch, interval, timeout, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset (REQUIRED; daemonsets are not scalable)")
	cmd.Flags().IntVar(&replicas, "replicas", -1, "desired replica count (REQUIRED; >= 0)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "after scaling, wait for the rollout to converge")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt for --replicas=0")
	o.AddOutputFlags(cmd)
	return cmd
}

// scaleResult is the JSON-mode shape emitted on success. We synthesize
// a stable summary rather than forwarding the (verbose) post-PATCH
// workload object — JSON consumers care about whether the scale took
// effect, not about every field of the object.
type scaleResult struct {
	Operation string `json:"operation"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int    `json:"replicas"`
}

// RunScale is the exported scale entry point used by `workload stop`
// and `workload start` (which are thin aliases over scale-with-known-
// replicas). Keeps the PATCH plumbing in one place.
func RunScale(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string, replicas int, watch bool, interval, timeout time.Duration, assumeYes bool) error {
	return runScale(ctx, o, namespace, name, kindPlural, replicas, watch, interval, timeout, assumeYes)
}

func runScale(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name, kindPlural string, replicas int, watch bool, interval, timeout time.Duration, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if replicas == 0 {
		if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Scale %s %s/%s to 0 replicas? Traffic will be paused", SingularKind(kindPlural), namespace, name),
			assumeYes); err != nil {
			return err
		}
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": replicas,
		},
	}
	path := buildGetPath(namespace, kindPlural, name)
	var patched Workload
	if err := clusterclient.Patch(ctx, client, path, "application/merge-patch+json", body, &patched); err != nil {
		return fmt.Errorf("scale %s %s/%s to %d: %w", SingularKind(kindPlural), namespace, name, replicas, err)
	}

	if !watch {
		result := scaleResult{
			Operation: "scale",
			Kind:      SingularKind(kindPlural),
			Namespace: namespace,
			Name:      name,
			Replicas:  replicas,
		}
		if o.IsJSON() {
			return o.PrintJSON(result)
		}
		if !o.Quiet {
			fmt.Fprintf(os.Stdout, "%s %s/%s scaled to %d replicas\n",
				SingularKind(kindPlural), namespace, name, replicas)
		}
		return nil
	}

	// --watch path: chain straight into the rollout-status convergence
	// loop so users get one consolidated experience. We re-invoke
	// runRolloutStatus with watch=true rather than duplicating the
	// polling code — single source of convergence truth.
	if !o.Quiet && !o.IsJSON() {
		fmt.Fprintf(os.Stdout, "%s %s/%s scaled to %d replicas; waiting for rollout to converge\n",
			SingularKind(kindPlural), namespace, name, replicas)
	}
	return runRolloutStatus(ctx, o, namespace, name, kindPlural, true, interval, timeout)
}

package dev

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRevertCommand: `olares-cli dev revert <ns/name> --kind K
// [-c CONTAINER] [-w]`.
func NewRevertCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		namespace string
		kindRaw   string
		container string
		image     string
		watch     bool
		interval  time.Duration
		timeout   time.Duration
		jsonOut   bool
	)
	cmd := &cobra.Command{
		Use:   "revert [<ns/name | name>]",
		Short: "restore the images a workload ran before `dev deploy`",
		Long: `Put a workload back the way ` + "`dev deploy`" + ` found it.

Two ways to say which workloads, mirroring ` + "`dev deploy`" + `:

  <ns/name> --kind K   revert one workload.
  --image <ref>        revert every workload currently running that
                       image — the counterpart of deploying by
                       --replaces, so a component can be undone without
                       remembering where it landed.

Reads the ` + workload.PreviousImagesAnnotation + ` annotation and
restores both the image and the imagePullPolicy recorded there, then
clears the entries it consumed. Restoring the policy matters: set-image
forces IfNotPresent so a side-loaded build is usable at all, and leaving
that behind would pin the workload to whatever happens to be cached on
the node and quietly defeat the next chart upgrade.

Without -c every recorded container is restored. An entry whose
container has since disappeared from the pod template is reported and
skipped rather than re-added.

Find what is currently overridden with ` + "`olares-cli dev status`" + `.
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if (len(args) == 1) == (strings.TrimSpace(image) != "") {
				return fmt.Errorf("pass exactly one of <ns/name> or --image <ref>")
			}
			if c.Flags().Changed("interval") && !watch {
				return fmt.Errorf("--interval requires --watch")
			}
			if c.Flags().Changed("timeout") && !watch {
				return fmt.Errorf("--timeout requires --watch")
			}
			mode := workload.OutputMode{JSON: jsonOut, Quiet: jsonOut}

			var targets []workload.RevertParams
			if len(args) == 1 {
				ns, name, err := splitNsName(namespace, args[0])
				if err != nil {
					return err
				}
				if strings.TrimSpace(kindRaw) == "" {
					return fmt.Errorf("--kind is required when naming a workload")
				}
				plural, err := workload.NormalizeKind(kindRaw)
				if err != nil {
					return err
				}
				if plural == workload.KindAll {
					return fmt.Errorf("--kind must name one kind, not %q", kindRaw)
				}
				targets = append(targets, workload.RevertParams{
					Namespace: ns, Name: name, KindPlural: plural, Container: container,
					Watch: watch, Interval: interval, Timeout: timeout,
				})
			} else {
				resolved, err := revertTargetsForImage(c.Context(), f, mode, namespace, image, container, watch, interval, timeout)
				if err != nil {
					return err
				}
				if len(resolved) == 0 {
					return fmt.Errorf("no workload is running %s — nothing to revert\n"+
						"List what is overridden with `olares-cli dev status`", image)
				}
				targets = resolved
			}

			var all []workload.RevertResult
			var firstErr error
			for _, t := range targets {
				results, err := workload.RevertImages(c.Context(), f, mode, t)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %s %s/%s: %v\n",
						workload.SingularKind(t.KindPlural), t.Namespace, t.Name, err)
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				all = append(all, results...)
			}

			if jsonOut {
				if err := printJSON(all); err != nil {
					return err
				}
				return firstErr
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAMESPACE\tNAME\tCONTAINER\tFROM\tTO\tNOTE")
			for _, r := range all {
				note := r.Skipped
				if note == "" {
					note = "restored"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
					r.Namespace, r.Name, r.Container, dashIfEmpty(r.From), dashIfEmpty(r.To), note)
			}
			tw.Flush()
			return firstErr
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (with a bare name), or the namespace to scan (with --image)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (required when naming a workload)")
	cmd.Flags().StringVar(&image, "image", "", "revert every workload currently running this image")
	cmd.Flags().StringVarP(&container, "container", "c", "", "restore only this container (default: every recorded container)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "wait for the rollout to converge")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit the results as JSON")
	return cmd
}

// revertTargetsForImage turns "--image beclab/app-service:dev" into one
// RevertParams per workload running it.
//
// It resolves by the image the workloads run *now* (the dev tag), not by
// what they will be restored to, because that is the thing the caller
// knows: `make dev-revert C=app-service` knows it pushed
// beclab/app-service:dev and has no idea which released tag the chart
// pinned. Deduplication is by workload, since one workload with two
// overridden containers is a single revert.
func revertTargetsForImage(
	ctx context.Context,
	f *cmdutil.Factory,
	mode workload.OutputMode,
	namespace, image, container string,
	watch bool,
	interval, timeout time.Duration,
) ([]workload.RevertParams, error) {
	refs, err := workload.FindImageRefsFor(ctx, f, mode, namespace, image)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var out []workload.RevertParams
	for _, ref := range refs {
		if container != "" && ref.Container != container {
			continue
		}
		plural, err := ref.KindPlural()
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping: %v\n", err)
			continue
		}
		key := ref.Namespace + "/" + plural + "/" + ref.Workload
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, workload.RevertParams{
			Namespace:  ref.Namespace,
			Name:       ref.Workload,
			KindPlural: plural,
			Container:  container,
			Watch:      watch,
			Interval:   interval,
			Timeout:    timeout,
		})
	}
	return out, nil
}

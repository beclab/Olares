package dev

import (
	"bufio"
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

// NewDeployCommand: `olares-cli dev deploy <image> --replaces <image>
// [--to ns/name] [--kind K] [-c CONTAINER] [-w] [--yes]`.
func NewDeployCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		replaces   string
		to         string
		kindRaw    string
		container  string
		namespace  string
		pullPolicy string
		watch      bool
		interval   time.Duration
		timeout    time.Duration
		assumeYes  bool
		jsonOut    bool
	)
	cmd := &cobra.Command{
		Use:   "deploy <image>",
		Short: "point workloads at a dev image (resolving targets by the image they replace)",
		Long: `Repoint one or more workloads at <image>.

Two ways to say which workloads:

  --replaces <image>   find every workload currently referencing that
                       image and repoint all of them. The lookup is the
                       same full-cluster, tag- and digest-normalized scan
                       ` + "`cluster workload images <IMAGE>`" + ` runs, so it
                       cannot miss a reference on a later page.

  --to <ns/name>       name one workload explicitly (needs --kind).

--replaces is the normal path when working on this repo: you know the
released tag baked into the chart, not where it is used.

Every target is listed and confirmed once before anything is patched —
an image referenced by several workloads repoints all of them, and that
is worth seeing first.

Patching goes through the same code path as
` + "`cluster workload set-image`" + `, including the annotation that lets
` + "`dev revert`" + ` restore the original image and pull policy.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			image := strings.TrimSpace(args[0])
			if image == "" {
				return fmt.Errorf("image is required")
			}
			if (replaces == "") == (to == "") {
				return fmt.Errorf("pass exactly one of --replaces <image> or --to <ns/name>")
			}
			if c.Flags().Changed("interval") && !watch {
				return fmt.Errorf("--interval requires --watch")
			}
			if c.Flags().Changed("timeout") && !watch {
				return fmt.Errorf("--timeout requires --watch")
			}
			return runDeploy(c.Context(), f, deployParams{
				Image:      image,
				Replaces:   replaces,
				To:         to,
				KindRaw:    kindRaw,
				Container:  container,
				Namespace:  namespace,
				PullPolicy: pullPolicy,
				Watch:      watch,
				Interval:   interval,
				Timeout:    timeout,
				AssumeYes:  assumeYes,
				JSON:       jsonOut,
			})
		},
	}
	cmd.Flags().StringVar(&replaces, "replaces", "", "repoint every workload currently referencing this image")
	cmd.Flags().StringVar(&to, "to", "", "repoint one workload, as ns/name (requires --kind)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind for --to: deployment | statefulset | daemonset")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (optional when the pod template has one container)")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "restrict the --replaces scan to one namespace")
	cmd.Flags().StringVar(&pullPolicy, "pull-policy", workload.DefaultDevPullPolicy,
		`imagePullPolicy to set alongside the image; "" leaves it untouched`)
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "wait for each rollout to converge")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit the per-target results as JSON")
	return cmd
}

type deployParams struct {
	Image      string
	Replaces   string
	To         string
	KindRaw    string
	Container  string
	Namespace  string
	PullPolicy string
	Watch      bool
	Interval   time.Duration
	Timeout    time.Duration
	AssumeYes  bool
	JSON       bool
}

// target is one resolved (workload, container) pair to patch.
type target struct {
	Namespace  string
	Name       string
	KindPlural string
	Kind       string
	Container  string
	From       string
}

func runDeploy(ctx context.Context, f *cmdutil.Factory, p deployParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	mode := workload.OutputMode{JSON: p.JSON, Quiet: p.JSON}

	targets, err := resolveTargets(ctx, f, mode, p)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no workload references %s — nothing to repoint\n"+
			"Check what is actually deployed with `olares-cli cluster workload images %s`",
			p.Replaces, p.Replaces)
	}

	if !p.JSON {
		printTargets(os.Stderr, targets, p.Image)
	}
	if err := confirmTargets(len(targets), p.AssumeYes, p.JSON); err != nil {
		return err
	}

	results := make([]workload.SetImageResult, 0, len(targets))
	var firstErr error
	for _, t := range targets {
		res, err := workload.SetImage(ctx, f, mode, workload.SetImageParams{
			Namespace:   t.Namespace,
			Name:        t.Name,
			KindPlural:  t.KindPlural,
			Container:   t.Container,
			Image:       p.Image,
			PullPolicy:  p.PullPolicy,
			Watch:       p.Watch,
			Interval:    p.Interval,
			Timeout:     p.Timeout,
			SkipConfirm: true,
		})
		if err != nil {
			// Keep going: a partially applied fan-out is far easier to
			// finish or unwind than one that stopped at an arbitrary
			// point without saying which targets were done.
			fmt.Fprintf(os.Stderr, "error: %s %s/%s: %v\n", t.Kind, t.Namespace, t.Name, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		results = append(results, res)
	}

	if p.JSON {
		return printJSON(results)
	}
	return firstErr
}

func resolveTargets(ctx context.Context, f *cmdutil.Factory, mode workload.OutputMode, p deployParams) ([]target, error) {
	if p.To != "" {
		ns, name, err := splitNsName(p.Namespace, p.To)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(p.KindRaw) == "" {
			return nil, fmt.Errorf("--kind is required with --to")
		}
		plural, err := workload.NormalizeKind(p.KindRaw)
		if err != nil {
			return nil, err
		}
		if plural == workload.KindAll {
			return nil, fmt.Errorf("--kind must name one kind, not %q", p.KindRaw)
		}
		return []target{{
			Namespace:  ns,
			Name:       name,
			KindPlural: plural,
			Kind:       workload.SingularKind(plural),
			Container:  p.Container,
		}}, nil
	}

	refs, err := workload.FindImageRefsFor(ctx, f, mode, p.Namespace, p.Replaces)
	if err != nil {
		return nil, err
	}
	var out []target
	for _, ref := range refs {
		plural, err := ref.KindPlural()
		if err != nil {
			// Jobs / CronJobs can reference the image but cannot be
			// repointed. Say so instead of dropping them silently.
			fmt.Fprintf(os.Stderr, "skipping: %v\n", err)
			continue
		}
		if p.Container != "" && ref.Container != p.Container {
			continue
		}
		out = append(out, target{
			Namespace:  ref.Namespace,
			Name:       ref.Workload,
			KindPlural: plural,
			Kind:       ref.Kind,
			Container:  ref.Container,
			From:       ref.Image,
		})
	}
	return out, nil
}

func printTargets(w *os.File, targets []target, image string) {
	fmt.Fprintf(w, "About to repoint %d container(s) at %s:\n\n", len(targets), image)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tKIND\tNAME\tCONTAINER\tCURRENT")
	for _, t := range targets {
		from := t.From
		if from == "" {
			from = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", t.Namespace, t.Kind, t.Name, dashIfEmpty(t.Container), from)
	}
	tw.Flush()
	fmt.Fprintln(w)
}

// confirmTargets asks once for the whole fan-out.
//
// It does not reuse clusteropts.ConfirmDestructive because that lives
// under the cluster tree's internal/ directory; the prompt contract
// (default no, "y"/"yes" to proceed, --yes to skip) is kept identical
// on purpose.
func confirmTargets(count int, assumeYes, jsonMode bool) error {
	if assumeYes {
		return nil
	}
	if jsonMode {
		return fmt.Errorf("--json requires --yes (there is nobody to answer the confirmation prompt)")
	}
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		return fmt.Errorf("refusing to repoint %d container(s) without confirmation: "+
			"stdin is not a terminal — pass --yes to proceed non-interactively", count)
	}
	fmt.Fprintf(os.Stderr, "Repoint these %d container(s)? Pods will be recreated [y/N]: ", count)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted")
	}
}

func splitNsName(nsFlag, arg string) (string, string, error) {
	if strings.Contains(arg, "/") {
		parts := strings.SplitN(arg, "/", 2)
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid target %q (want ns/name)", arg)
		}
		return parts[0], parts[1], nil
	}
	if strings.TrimSpace(nsFlag) == "" {
		return "", "", fmt.Errorf("target %q is a bare name — pass -n/--namespace or use ns/name", arg)
	}
	return nsFlag, arg, nil
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

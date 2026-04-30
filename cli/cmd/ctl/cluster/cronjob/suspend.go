package cronjob

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSuspendCommand: `olares-cli cluster cronjob suspend
// <ns/name | name> [-n NS] [--yes]`.
//
// PATCHes the CronJob with `{"spec":{"suspend":true}}` using
// Content-Type `application/merge-patch+json` (matches the SPA's
// toggleJob flow). Pauses scheduled runs without deleting the
// CronJob, and is reversed by `cronjob resume`.
//
// Wrapped in ConfirmDestructive — suspending production schedules
// is reversible but visibly disruptive (next run won't fire), so
// scripts must opt in via --yes.
func NewSuspendCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "suspend <ns/name | name>",
		Short: "suspend one CronJob (set spec.suspend=true)",
		Long: `Suspend one CronJob — sets spec.suspend=true via merge-patch+json.

Equivalent to the SPA's "Suspend" toggle. Reversed by
` + "`cluster cronjob resume`" + `.

This is mutating: future scheduled runs are paused until resumed.
Pass --yes to skip the confirmation prompt for scripted use.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runToggle(c.Context(), o, ns, name, true, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	o.AddOutputFlags(cmd)
	return cmd
}

// toggleResult is the JSON-mode shape emitted on success. We
// synthesize a stable summary rather than forwarding the (verbose)
// post-PATCH CronJob object — JSON consumers care about whether the
// suspend bit changed, not about every field of the object.
type toggleResult struct {
	Operation       string `json:"operation"`
	Namespace       string `json:"namespace"`
	CronJob         string `json:"cronjob"`
	Suspend         bool   `json:"suspend"`
	ResourceVersion string `json:"resourceVersion"`
}

// runToggle is the shared body of suspend / resume. The only
// difference between the two verbs is the suspend value (true /
// false) and the prompt message; everything else (PATCH endpoint,
// merge-patch+json content type, RV-aware GET, ConfirmDestructive
// wrapping) is identical.
//
// resume passes assumeYes=true at the call site (re-enabling is
// non-destructive), so the prompt only ever fires for suspend.
func runToggle(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, suspend, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := Get(ctx, o, namespace, name)
	if err != nil {
		return err
	}
	currently := false
	if c.Spec.Suspend != nil {
		currently = *c.Spec.Suspend
	}
	if currently == suspend {
		// No-op short-circuit. Tell the user (in non-quiet mode)
		// rather than silently swallow — operators sometimes flip
		// the wrong bit and this is the cheapest way to surface it.
		if !o.Quiet {
			verb := "suspended"
			if !suspend {
				verb = "active"
			}
			fmt.Fprintf(os.Stderr, "cronjob %s/%s is already %s — no change\n", namespace, name, verb)
		}
		if o.IsJSON() {
			return o.PrintJSON(toggleResult{
				Operation:       toggleVerb(suspend),
				Namespace:       namespace,
				CronJob:         name,
				Suspend:         suspend,
				ResourceVersion: c.Metadata.ResourceVersion,
			})
		}
		return nil
	}

	if suspend {
		if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Suspend cronjob %s/%s? Future scheduled runs will be paused", namespace, name),
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
			"suspend": suspend,
		},
	}
	path := buildGetPath(namespace, name)
	var patched CronJob
	if err := clusterclient.Patch(ctx, client, path, "application/merge-patch+json", body, &patched); err != nil {
		return fmt.Errorf("%s cronjob %s/%s: %w", toggleVerb(suspend), namespace, name, err)
	}

	result := toggleResult{
		Operation:       toggleVerb(suspend),
		Namespace:       namespace,
		CronJob:         name,
		Suspend:         suspend,
		ResourceVersion: patched.Metadata.ResourceVersion,
	}
	if o.IsJSON() {
		return o.PrintJSON(result)
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "cronjob %s/%s: spec.suspend=%v (resourceVersion=%s)\n",
			namespace, name, suspend, patched.Metadata.ResourceVersion)
	}
	return nil
}

func toggleVerb(suspend bool) string {
	if suspend {
		return "suspend"
	}
	return "resume"
}

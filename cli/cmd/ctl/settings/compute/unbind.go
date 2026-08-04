package compute

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings compute unbind <app>`
//
// Backed by DELETE /api/apps/<app>/compute-resources/bindings. Mirrors the
// SPA's UnbindAppDialog: unbinding releases the app from its device(s) and
// stops the app. A multi-card app releases every linked card at once.
//
// (The per-app GET /api/apps/:name/compute-resources/bindings is unused by
// the SPA — the Accelerator page reads bindings from /api/compute-resources,
// so `compute list` is the inspection path here too.)
func newUnbindCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		skipConfirm bool
		output      string
	)
	cmd := &cobra.Command{
		Use:   "unbind <app>",
		Short: "unbind an app from its accelerator device(s) (stops the app)",
		Long: `Unbind an app from its accelerator device(s).

This mirrors the SPA's "Unbind" action: the app is unbound and stops
simultaneously. A multi-card app releases every linked card at once; a
single-card app releases only its one binding.

By default you must type the whole word "yes" when prompted. Use --yes to
skip confirmation (for scripting).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "unbind compute bindings"); err != nil {
				return err
			}
			if err := requireComputeBackendVersion(ctx, f); err != nil {
				return err
			}
			app := strings.TrimSpace(args[0])
			if app == "" {
				return fmt.Errorf("app name is required")
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runUnbind(ctx, f, app, skipConfirm, format), "unbind compute bindings")
		},
	}
	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "skip interactive confirmation (dangerous)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runUnbind(ctx context.Context, f *cmdutil.Factory, app string, skipConfirm bool, format Format) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	// Read current bindings from /compute-resources (same source the SPA
	// uses) so the confirmation can state the real card count and whether
	// this is a multi-card release.
	var nodes []computeNode
	if err := doGetEnvelope(ctx, pc.doer, "/api/compute-resources", &nodes); err != nil {
		return err
	}
	cards, multiCard := appBindings(nodes, app)
	if len(cards) == 0 {
		return fmt.Errorf("app %q has no compute bindings to unbind (see `olares-cli settings compute list`)", app)
	}

	if !skipConfirm {
		if multiCard {
			fmt.Fprintf(os.Stderr,
				"This will unbind app %q from ALL %d linked card(s) and stop the app.\n"+
					"Type 'yes' to continue: ", app, len(cards))
		} else {
			fmt.Fprintf(os.Stderr,
				"This will unbind app %q from its accelerator device and stop the app.\n"+
					"Type 'yes' to continue: ", app)
		}
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) != "yes" {
			return fmt.Errorf("aborting unbind: confirmation was not yes")
		}
	}

	path := "/api/apps/" + url.PathEscape(app) + "/compute-resources/bindings"
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", path, nil, nil); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, map[string]interface{}{
			"app":            app,
			"status":         "unbound",
			"released_cards": len(cards),
		})
	default:
		_, err := fmt.Fprintf(os.Stdout, "unbound app %q from compute (%d card(s) released; the app was stopped)\n", app, len(cards))
		return err
	}
}

// appBindings returns the device ids an app is bound to across all nodes, and
// whether any of those bindings is a multi-card binding (spec.supportMultiCards).
func appBindings(nodes []computeNode, app string) (cards []string, multiCard bool) {
	for _, node := range nodes {
		for _, d := range node.Devices {
			for _, b := range d.Bindings {
				if b.AppName != app {
					continue
				}
				cards = append(cards, d.ID)
				if b.Spec != nil && b.Spec.SupportMultiCards {
					multiCard = true
				}
			}
		}
	}
	return cards, multiCard
}

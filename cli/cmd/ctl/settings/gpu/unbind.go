package gpu

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/edge"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings gpu unbind <app>` (Olares 1.12.6+)
//
// Releases an app's compute (GPU) bindings via
// olaresclient.ComputeOps.ReleaseAppBindings (DELETE
// /api/apps/<app>/compute-resources/bindings). The backend implements release
// as a suspend, so this stops the app — a destructive action that prompts for
// confirmation unless --yes is passed.
func NewUnbindCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "unbind <app>",
		Short: "release an app's compute bindings, suspending it (Olares 1.12.6+)",
		Long: `Release the compute (GPU) device bindings held by an app.

The backend frees the bindings by suspending the app, so the app stops until
it is bound and resumed again. Prompts for confirmation by default; pass --yes
to skip the prompt for automation. Non-TTY stdin without --yes is a hard error.

Example:
  olares-cli settings gpu unbind ollama
  olares-cli settings gpu unbind ollama --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "release GPU bindings"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runUnbind(ctx, f, args[0], assumeYes), "release GPU bindings")
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required for non-TTY automation)")
	return cmd
}

func runUnbind(ctx context.Context, f *cmdutil.Factory, appName string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	doer, _, err := edge.New(ctx, f)
	if err != nil {
		return err
	}

	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Release compute bindings for %q? This suspends (stops) the app.", appName),
		assumeYes); err != nil {
		return err
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		if _, err := c.ReleaseAppBindings(ctx, doer, appName); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "released compute bindings for %s (app suspended)\n", appName)
		return nil
	})
}

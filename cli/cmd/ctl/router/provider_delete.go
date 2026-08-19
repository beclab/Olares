package router

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider delete <name|id>`
// DELETE /console/api/providers/:id
//
// The row goes, and so does everything that referenced it: its models, the
// default-model choices pointing at them, per-model settings, and quotas scoped
// to them. There is no archive state and no undo; the credentials are gone with
// their history.
//
// `provider update --status disabled` is the reversible alternative, and is
// what you want for an upstream you are only trying to stop traffic to.

func newProviderDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "delete <name|id>",
		Short: "remove a provider and everything that referenced it",
		Long: `Delete a provider.

This removes the provider, its models, the default-model choices that pointed at
those models, their per-model settings, and any quota scoped to them. Credential
history goes too. Nothing is archived and nothing can be recovered; re-creating
the provider afterwards starts from an empty state.

If the goal is to stop traffic rather than to forget the upstream, use
"provider update --status disabled" instead: routing stops immediately and
everything else is preserved, so re-enabling restores the previous behaviour.

A provider registered for a model application cannot be deleted here. Its
lifecycle belongs to the application — uninstall the application and the
provider goes with it.

Confirmation is required. --yes skips the prompt, and is mandatory when stdin
is not a terminal so an unattended script cannot destroy state by accident.

Example:
  olares-cli router provider delete stale-openai --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderDelete(c.Context(), f, args[0], assumeYes, output)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderDelete(ctx context.Context, f *cmdutil.Factory, ref string, assumeYes bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	if found.isMarketSourced() {
		return marketOwnedErr(found, "delete")
	}

	detail, err := getProvider(ctx, pc, found.ID)
	if err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Delete provider %q and its %d models, along with the defaults, settings and quotas pointing at them?",
			found.Name, len(detail.Models)),
		assumeYes); err != nil {
		return err
	}

	var res struct {
		ID      string `json:"id"`
		Deleted bool   `json:"deleted"`
	}
	path := epProvider(found.ID)
	if err := pc.router.doJSON(ctx, "DELETE", path, nil, &res); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	_, err = fmt.Fprintf(os.Stdout, "deleted provider %s (%s)\n", found.Name, res.ID)
	return err
}

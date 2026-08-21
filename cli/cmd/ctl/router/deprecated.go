package router

// Paths that moved, kept answering for one release.
//
// Every command here is hidden and delegates to whatever replaced it, so there
// is one implementation and not two drifting copies. Cobra prints the
// `Deprecated` line before running, which is the only difference a caller sees.
// The file is meant to be deleted whole rather than pruned entry by entry.

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// newDeprecatedDefaultCommand is `router default`, which turned out to be three
// ways of saying `router route`. A category is a route whose kind Router owns
// rather than a noun of its own: the old subtree read the same rows through the
// same filter and wrote the same PATCH, so it was a `--kind` flag wearing a
// verb, and the price was a help text on each side explaining it was not the
// other one.
func newDeprecatedDefaultCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "default",
		Short:  "moved to `router route`",
		Hidden: true,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDeprecatedDefaultShowCommand(f))
	cmd.AddCommand(newDeprecatedDefaultToggleCommand(f, true))
	cmd.AddCommand(newDeprecatedDefaultToggleCommand(f, false))
	return cmd
}

func newDeprecatedDefaultShowCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:        "show",
		Short:      "moved to `router route list --kind default`",
		Deprecated: "use `olares-cli router route list --kind default` instead.",
		Hidden:     true,
		Args:       cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runRouteList(c.Context(), f, routeKindDefault, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedDefaultToggleCommand(f *cmdutil.Factory, on bool) *cobra.Command {
	var output string
	verb := "disable"
	if on {
		verb = "enable"
	}
	cmd := &cobra.Command{
		Use:        verb + " <category>",
		Short:      "moved to `router route " + verb + "`",
		Deprecated: "use `olares-cli router route " + verb + " <category>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			enabled := on
			return runRoutePatch(c.Context(), f, args[0], "", &enabled, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

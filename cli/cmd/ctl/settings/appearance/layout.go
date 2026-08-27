package appearance

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance layout reset`
//
// Backed by user-service's desktop-layout.controller.ts, which drops the
// launchpad and dock layouts and broadcasts desktop_layout_updated to any
// open session. The arrangement is not recoverable afterwards.
const layoutResetPath = "/api/desktop/layout/reset"

// layoutResetResult names the surfaces the upstream actually cleared.
type layoutResetResult struct {
	ChangedSurfaces []string `json:"changedSurfaces"`
}

func NewLayoutCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layout",
		Short: "desktop layout",
		Long: `Manage the desktop layout (Settings -> Appearance > Reset desktop
layout).

Subcommands:
  reset                 restore the default launchpad and dock arrangement
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newLayoutResetCommand(f))
	return cmd
}

func newLayoutResetCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "reset the desktop layout to its defaults",
		Long: `Reset the desktop layout to its defaults.

This discards the launchpad ordering, folders and dock arrangement for
the current user. There is no undo, and any open Olares desktop session
re-renders immediately.

Examples:
  olares-cli settings appearance layout reset
  olares-cli settings appearance layout reset --yes
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runLayoutReset(c.Context(), f, assumeYes)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	return cmd
}

func runLayoutReset(ctx context.Context, f *cmdutil.Factory, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Gate before the prompt: never ask the user to confirm something
	// this backend cannot carry out.
	if err := requireAppearanceBackendVersion(ctx, f,
		"settings appearance layout reset", "the desktop layout reset route"); err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		"Reset the desktop layout, discarding the current launchpad and dock arrangement?",
		assumeYes); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var result layoutResetResult
	if err := doMutateEnvelope(ctx, pc.doer, "POST", layoutResetPath, map[string]string{}, &result); err != nil {
		return err
	}
	if len(result.ChangedSurfaces) == 0 {
		fmt.Println("Desktop layout reset.")
		return nil
	}
	fmt.Printf("Desktop layout reset (%s).\n", strings.Join(result.ChangedSurfaces, ", "))
	return nil
}

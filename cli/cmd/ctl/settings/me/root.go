// Package me implements the `olares-cli settings me` subtree — the
// 13th, non-canonical, "self-service" area that mirrors the SPA's
// avatar/Person dropdown.
//
// This area is intentionally outside the 12 docs sections at
// https://docs.olares.com/manual/olares/settings/. It's a CLI convenience
// that bundles a small set of self-service items every authenticated user
// (including owner) routinely uses:
//
//	whoami        cached role + olaresId           (alias for `olares-cli profile whoami`)
//	version       current OS version
//
// All `me` verbs are roleNormal-floor: every authenticated user can call
// them. Browser-bound / TermiPass-bound Person sub-pages (Hardware QR,
// VaultActiveSession, OlaresSpace, Authority) are intentionally excluded —
// see plan.md's "Self-service sub-tree" section for the rationale.
package me

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewMeCommand returns the `settings me` parent. The parent prints help
// by default; subcommands cover whoami / version.
func NewMeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Self-service settings for the current user (whoami / version)",
		Long: `Self-service "about me" subcommands.

This is the 13th, non-canonical sub-tree under "settings" — it exists so
the SPA's avatar/Person dropdown items have an obvious CLI home without
bloating the 12 documented Settings sections. Every verb here is callable
by any authenticated user (owner / admin / user).

Subcommands:
  whoami                  alias for "olares-cli profile whoami"
  version                 current OS version
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewWhoamiCommand(f))
	cmd.AddCommand(NewVersionCommand(f))
	cmd.AddCommand(NewPasswordCommand(f))
	return cmd
}

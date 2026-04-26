// Package appearance implements the `olares-cli settings appearance` subtree
// (Settings -> Appearance). Backed by user-service's wallpaper.controller.ts
// — only the language/theme JSON read + language set fall in CLI scope; the
// wallpaper image picker is intentionally browser-bound and out of scope.
package appearance

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppearanceCommand returns the `settings appearance` parent.
//
// Phase 1 will add `get` (read /api/wallpaper/config/system); Phase 2 will
// add `language set` (POST /api/wallpaper/update/language). Until then the
// parent prints its own help, confirming the umbrella wires through.
func NewAppearanceCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appearance",
		Short: "Appearance settings (language, theme view)",
		Long: `Read and update appearance preferences (Settings -> Appearance).

Subcommands will be added in subsequent phases:
  Phase 1: get
  Phase 2: language set

Wallpaper image upload + theme picker stay in the SPA — they are browser
blob/picker flows with no useful CLI surface.
`,
	}
	cmd.SilenceUsage = true
	return cmd
}

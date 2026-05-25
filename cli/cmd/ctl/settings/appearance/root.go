// Package appearance implements the `olares-cli settings appearance` subtree
// (Settings -> Appearance). Backed by user-service's wallpaper.controller.ts
// — only the language/theme JSON read + language set fall in CLI scope; the
// wallpaper image picker is intentionally browser-bound and out of scope.
package appearance

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppearanceCommand returns the `settings appearance` parent: read
// the language / locale and update the system language. Wallpaper +
// theme picker stay in the SPA (browser blob/picker flows).
func NewAppearanceCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appearance",
		Short: "Appearance settings (language, locale)",
		Long: `Read and update appearance preferences (Settings -> Appearance).

Subcommands:
  get                              show language + locale
  language set --value <code>      update the system language

Wallpaper image upload + theme picker stay in the SPA — they are browser
blob/picker flows with no useful CLI surface.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewLanguageCommand(f))
	return cmd
}

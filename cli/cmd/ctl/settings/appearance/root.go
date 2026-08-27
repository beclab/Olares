// Package appearance implements the `olares-cli settings appearance`
// subtree (Settings -> Appearance): locale, theme, desktop widget
// preferences, wallpaper, and the desktop layout reset.
//
// See common.go for how the three backing services are reached.
package appearance

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppearanceCommand returns the `settings appearance` parent.
func NewAppearanceCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appearance",
		Short: "Appearance settings (locale, theme, widgets, wallpaper, layout)",
		Long: `Read and update appearance preferences (Settings -> Appearance).

Subcommands:
  get                                   show the whole page in one read
  language set <locale>                 update the system language
  theme set <light|dark>                update the theme preference
  widget set [flags]                    update the desktop widget preferences
  wallpaper set <surface> <bg>          select a wallpaper
  wallpaper style set <surface> <mode>  set the wallpaper fill mode
  wallpaper upload <surface> --file     upload a local image and select it
  wallpaper delete <surface> <bg>       remove an uploaded image
  layout reset                          restore the default desktop layout

"get" is the only read verb: the write subtrees have no "get" of their
own, because reading Appearance means reading the page.

A theme written here is what apps and the login flow observe. It does not
change an Olares desktop already open in a browser, whose light/dark state
comes from a cookie the SPA never sends to the backend.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewLanguageCommand(f))
	cmd.AddCommand(NewThemeCommand(f))
	cmd.AddCommand(NewWidgetCommand(f))
	cmd.AddCommand(NewWallpaperCommand(f))
	cmd.AddCommand(NewLayoutCommand(f))
	return cmd
}

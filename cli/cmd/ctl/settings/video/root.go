// Package video implements the `olares-cli settings video` subtree (Settings
// -> Video). Backed by the /api/files/video/config slice of user-service's
// files.controller.ts.
package video

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewVideoCommand returns the `settings video` parent: read the
// playback configuration blob exposed by user-service at
// /api/files/video/config. The matching write verb (config set) is
// out of scope for now.
func NewVideoCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Video preferences (Settings -> Video)",
		Long: `Read video playback preferences (single config blob exposed by
user-service at /api/files/video/config).

Subcommands:
  config get

Out of scope for now:
  config set
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewConfigCommand(f))
	return cmd
}

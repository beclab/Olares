// Package video implements the `olares-cli settings video` subtree (Settings
// -> Video). Backed by the /api/files/video/config slice of user-service's
// files.controller.ts.
package video

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewVideoCommand returns the `settings video` parent. Phase 1 ships
// `config get`; Phase 2 will add `config set`.
func NewVideoCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Video preferences (Settings -> Video)",
		Long: `Read and update video playback preferences (single config blob exposed by
user-service at /api/files/video/config).

Subcommands:
  config get                                              (Phase 1)

Subcommands landing in Phase 2:
  config set
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewConfigCommand(f))
	return cmd
}

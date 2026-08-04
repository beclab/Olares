package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/download"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewKnowledgeCommand assembles the `olares-cli knowledge` subtree.
// Identity and transport come from the active profile (same as market /
// files / settings).
//
// Phase 1 only exposes `knowledge download` (download-server task centre).
// Further knowledge-facing verbs can register here later.
func NewKnowledgeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Knowledge / Wise related APIs via the active profile",
		Long: `Knowledge-facing command tree (profile-authenticated).

Currently:

  download   download-server task centre (Settings /download edge)

Requires Olares 1.12.7+ for download verbs. This is not the top-level
"download" command (installer packages) and not "files download"
(pull a file from Drive).

Run "olares-cli knowledge <verb> --help" for details.
`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}

	cmd.AddCommand(download.NewDownloadCommand(f))
	return cmd
}

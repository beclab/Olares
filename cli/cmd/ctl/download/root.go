package download

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewDownloadCommand assembles `olares-cli knowledge download`.
// Identity and transport come from the active profile (same as market /
// files / settings). Requires Olares >= 1.12.7.
//
// Naming: this tree is the download-server *task centre*. It is not
// top-level `download` (installer packages) and not `files download`
// (pull a file from files-backend).
func NewDownloadCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Manage download-server tasks (Settings /download edge)",
		Long: `Manage download tasks via download-server.

Invoked as: olares-cli knowledge download <verb>

Requires Olares 1.12.7+. Calls go through the active profile's Settings
URL with a /download prefix (user-service relay → download provider →
download-server). Auth is the profile access token (X-Authorization);
the gateway injects X-Bfl-User — do not set it from the CLI.

This is the download *task centre* (create / list / pause / …). It is
not the top-level "download" command (installer packages) and not
"files download" (copy a file out of Drive).

Verb families:

  lifecycle   create, list, info, pause, resume, cancel, remove
  probe       inspect
  prefs       prefs get, prefs set
  sync        unfinished, sync
  torrent     torrent inspect, stats, peers, files, seed stop|resume
  file        file exists, file check, file remove
  settings    settings get, settings set  (download-server global config)

Universal flags:

  -o, --output {table,json}
      --app <name>     default wise (create / list / prefs)

Run "olares-cli knowledge download <verb> --help" for verb-specific flags.
`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewCreateCommand(f))
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewInfoCommand(f))
	cmd.AddCommand(NewPauseCommand(f))
	cmd.AddCommand(NewResumeCommand(f))
	cmd.AddCommand(NewCancelCommand(f))
	cmd.AddCommand(NewRemoveCommand(f))
	cmd.AddCommand(NewInspectCommand(f))
	cmd.AddCommand(NewPrefsCommand(f))
	cmd.AddCommand(NewUnfinishedCommand(f))
	cmd.AddCommand(NewSyncCommand(f))
	cmd.AddCommand(NewTorrentCommand(f))
	cmd.AddCommand(NewFileCommand(f))
	cmd.AddCommand(NewCookiesCommand(f))
	cmd.AddCommand(NewSettingsCommand(f))
	return cmd
}

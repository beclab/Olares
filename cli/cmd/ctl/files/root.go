package files

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewFilesCommand returns the `files` parent command, ready to be added to
// the olares-cli root.
//
// Phase 1 surface (only one verb, intentionally minimal — Phase 2 adds
// cat / cp / mv / rm / mkdir):
//
//	files ls <fileType>/<extend>[/<subPath>] [--json]
//
// The Factory is supplied by the root command so credential resolution and
// HTTP-client setup happen once per process — and so the global `--profile`
// flag wired up at the root can flow through here unchanged.
//
// See cmd/ctl/files/path.go for the front-end path schema and
// docs/notes/olares-cli-auth-profile-config.md for the broader Phase 1
// design (this is the demo command that closes Phase 1).
func NewFilesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "interact with the per-user files-backend (list, ...)",
		Long: `Talk to the Olares per-user files-backend over its /api/resources REST surface.

Every resource is addressed by a 3-segment "front-end path":

    <fileType>/<extend>[/<subPath>]

where:

    fileType   storage class:  drive | cache | sync | external |
                               awss3 | dropbox | google | tencent |
                               share | internal
    extend     volume / repo / account inside that class:
                  drive  -> Home or Data
                  cache  -> node name
                  sync   -> seafile repo id
                  cloud  -> account key
    subPath    path inside <extend> (root if omitted)

Examples:

    olares-cli files ls drive/Home/
    olares-cli files ls drive/Home/Documents
    olares-cli files ls drive/Data/
    olares-cli files ls sync/<repo_id>/
`,
	}
	for _, sub := range []*cobra.Command{
		NewLsCommand(f),
		NewUploadCommand(f),
		NewDownloadCommand(f),
		NewCatCommand(f),
		NewRmCommand(f),
	} {
		// Same rationale as cmd/ctl/profile/root.go: bad creds / network /
		// path-not-found errors are already actionable; don't bury them under
		// a usage dump.
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}

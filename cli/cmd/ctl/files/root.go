package files

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewFilesCommand returns the `files` parent command, ready to be added to
// the olares-cli root.
//
// Current verbs (mkdir is the next obvious add):
//
//	files ls       — list a directory                  (cmd/ctl/files/ls.go)
//	files upload   — resumable chunked upload          (cmd/ctl/files/upload.go)
//	files download — single-file or recursive pull     (cmd/ctl/files/download.go)
//	files cat      — stream a file to stdout           (cmd/ctl/files/cat.go)
//	files rm       — batched DELETE                    (cmd/ctl/files/rm.go)
//	files cp       — server-side copy via paste        (cmd/ctl/files/cp.go)
//	files mv       — server-side move via paste        (cmd/ctl/files/cp.go, action="move")
//	files rename   — synchronous in-place rename       (cmd/ctl/files/rename.go)
//	files share    — create / list / remove shares     (cmd/ctl/files/share.go,
//	                  internal: cross-user             cmd/ctl/files/share_create.go)
//	                  public:   external link
//	                  smb:      Samba network share
//	files repos    — list / inspect Sync (Seafile)     (cmd/ctl/files/repos.go,
//	                  libraries (repo_id catalog)      internal/files/repos/repos.go)
//
// cp / mv share a single PATCH /api/paste/<node>/ wire path (see
// cmd/ctl/files/cp.go and internal/files/cp/cp.go); the only
// difference is the action verb in the JSON body. `rename` is a
// distinct synchronous PATCH /api/resources/.../?destination=... call
// (see cmd/ctl/files/rename.go and internal/files/rename/rename.go) —
// no <node> URL segment, no task_id, basename-only payload. `share`
// fans out across the /api/share/share_path/ surface (see
// internal/files/share/share.go); the three creation flavors converge
// on the same POST endpoint and disambiguate via the `share_type`
// field in the JSON body.
//
// The Factory is supplied by the root command so credential resolution and
// HTTP-client setup happen once per process. Identity is whichever profile
// `olares-cli profile use` (or `profile login` / `profile import`) most
// recently selected; there is no per-invocation override flag.
//
// See cmd/ctl/files/path.go for the front-end path schema and
// docs/notes/olares-cli-auth-profile-config.md for the broader
// auth / profile design.
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
		NewCpCommand(f),
		NewMvCommand(f),
		NewRenameCommand(f),
		NewShareCommand(f),
		NewReposCommand(f),
	} {
		// Same rationale as cmd/ctl/profile/root.go: bad creds / network /
		// path-not-found errors are already actionable; don't bury them under
		// a usage dump.
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}

package search

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search dirs ...`
//
// Backed by /api/search/monitorsetting/include-directory/full_content.
// Same wire shape as excludes — search3's {code, message, data:
// string[]} envelope.
//
// Phase 2 will add `add` / `rm` (PUT/DELETE on the .../part path with
// `{include_directory: [...]}`).
func NewDirsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dirs",
		Short: "indexed directories (Settings -> Search > File Search)",
		Long: `Inspect the directories the search index is currently watching for
full-content indexing.

Subcommands:
  list                                                    (Phase 1)

Subcommands landing in Phase 2:
  add <path>..., rm <path>...
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDirsListCommand(f))
	return cmd
}

func newDirsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list directories under full-content indexing",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDirsList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runDirsList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var rows []string
	if err := doGetEnvelope(ctx, pc.doer, "/api/search/monitorsetting/include-directory/full_content", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderStringList(os.Stdout, rows, "no indexed directories")
	}
}

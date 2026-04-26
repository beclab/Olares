package search

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search excludes ...`
//
// Backed by /api/search/monitorsetting/exclude-pattern, which the SPA's
// FileSearch.vue uses to render the "exclude pattern" list. The wire
// body is search3's {code, message, data: string[]} envelope.
//
// Phase 2 will add `add` / `rm` (PUT/DELETE the same path with a
// `{exclude_pattern: [...]}` body).
func NewExcludesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "excludes",
		Short: "exclude-pattern list (Settings -> Search > File Search)",
		Long: `Inspect the search index's exclude-pattern list.

Subcommands:
  list                                                    (Phase 1)

Subcommands landing in Phase 2:
  add <pattern>..., rm <pattern>...
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newExcludesListCommand(f))
	return cmd
}

func newExcludesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list current exclude-patterns",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runExcludesList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runExcludesList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	if err := doGetEnvelope(ctx, pc.doer, "/api/search/monitorsetting/exclude-pattern", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderStringList(os.Stdout, rows, "no exclude patterns")
	}
}

func renderStringList(w io.Writer, rows []string, emptyMsg string) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, emptyMsg)
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintln(w, r); err != nil {
			return err
		}
	}
	return nil
}

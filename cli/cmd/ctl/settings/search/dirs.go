package search

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search dirs ...`
//
// Backed by /api/search/monitorsetting/include-directory/full_content.
// Same wire shape as excludes — search3's {code, message, data:
// string[]} envelope. `add` / `rm` POST against
// /include-directory/full_content/part with body
// `{include_directory: [...]}`.
//
// SPA reference: apps/packages/app/src/api/settings/search.ts
//   addSearchDirectories(values)    -> PUT    /full_content/part
//   deleteSearchDirectories(values) -> DELETE /full_content/part
func NewDirsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dirs",
		Short: "indexed directories (Settings -> Search > File Search)",
		Long: `Inspect and edit the directories the search index is watching for
full-content indexing.

Subcommands:
  list
  add <path>...
  rm  <path>...
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDirsListCommand(f))
	cmd.AddCommand(newDirsAddCommand(f))
	cmd.AddCommand(newDirsRmCommand(f))
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

func newDirsAddCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add <path>...",
		Short: "add one or more directories to full-content indexing",
		Long: `Add one or more directories to the full-content index. Paths must
exist on the Olares filesystem; the server appends them to the
existing list (it does not replace it).

Example:
  olares-cli settings search dirs add /data/Documents /data/Notes
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDirsAdd(c.Context(), f, args)
		},
	}
}

func newDirsRmCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <path>...",
		Short: "remove one or more directories from full-content indexing",
		Long: `Remove one or more directories from the full-content index. Paths
must match existing entries verbatim; use "list" to see what is
currently configured.

Example:
  olares-cli settings search dirs rm /data/Notes
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDirsRm(c.Context(), f, args)
		},
	}
}

func runDirsAdd(ctx context.Context, f *cmdutil.Factory, paths []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	clean := dedupTrim(paths)
	if len(clean) == 0 {
		return fmt.Errorf("no non-empty directory paths supplied")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string][]string{"include_directory": clean}
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", "/api/search/monitorsetting/include-directory/full_content/part", body, nil); err != nil {
		return err
	}
	fmt.Printf("Added %d directory(ies) to the full-content index.\n", len(clean))
	return nil
}

func runDirsRm(ctx context.Context, f *cmdutil.Factory, paths []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	clean := dedupTrim(paths)
	if len(clean) == 0 {
		return fmt.Errorf("no non-empty directory paths supplied")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string][]string{"include_directory": clean}
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", "/api/search/monitorsetting/include-directory/full_content/part", body, nil); err != nil {
		return err
	}
	fmt.Printf("Removed %d directory(ies) from the full-content index.\n", len(clean))
	return nil
}

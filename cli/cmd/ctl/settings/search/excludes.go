package search

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search excludes ...`
//
// Backed by /api/search/monitorsetting/exclude-pattern, which the SPA's
// FileSearch.vue uses to render the "exclude pattern" list. The wire
// body is search3's {code, message, data: string[]} envelope.
//
// Phase 2 adds `add` / `rm` against /exclude-pattern/part with the body
// `{exclude_pattern: [...]}`.
//
// SPA reference: apps/packages/app/src/api/settings/search.ts
//   addExcludePattern(values)    -> PUT    /exclude-pattern/part
//   deleteExcludePattern(values) -> DELETE /exclude-pattern/part
//
// Both endpoints expect *additions / removals*, not the full new list,
// so the CLI verbs accept one or more patterns as positional args.
func NewExcludesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "excludes",
		Short: "exclude-pattern list (Settings -> Search > File Search)",
		Long: `Inspect and edit the search index's exclude-pattern list.

Subcommands:
  list                                                    (Phase 1)
  add <pattern>...                                        (Phase 2)
  rm  <pattern>...                                        (Phase 2)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newExcludesListCommand(f))
	cmd.AddCommand(newExcludesAddCommand(f))
	cmd.AddCommand(newExcludesRmCommand(f))
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

func newExcludesAddCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add <pattern>...",
		Short: "add one or more exclude patterns",
		Long: `Add one or more exclude patterns to the search index. Patterns are
glob-style; the server appends them to the existing list (it does not
replace it). Use "list" first to see what is already there.

Example:
  olares-cli settings search excludes add "node_modules" ".cache/*"
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runExcludesAdd(c.Context(), f, args)
		},
	}
}

func newExcludesRmCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <pattern>...",
		Short: "remove one or more exclude patterns",
		Long: `Remove one or more exclude patterns from the search index. The
patterns must match existing entries verbatim; use "list" to see what
is currently configured.

Example:
  olares-cli settings search excludes rm "node_modules"
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runExcludesRm(c.Context(), f, args)
		},
	}
}

func runExcludesAdd(ctx context.Context, f *cmdutil.Factory, patterns []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	clean := dedupTrim(patterns)
	if len(clean) == 0 {
		return fmt.Errorf("no non-empty exclude patterns supplied")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string][]string{"exclude_pattern": clean}
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", "/api/search/monitorsetting/exclude-pattern/part", body, nil); err != nil {
		return err
	}
	fmt.Printf("Added %d exclude pattern(s).\n", len(clean))
	return nil
}

func runExcludesRm(ctx context.Context, f *cmdutil.Factory, patterns []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	clean := dedupTrim(patterns)
	if len(clean) == 0 {
		return fmt.Errorf("no non-empty exclude patterns supplied")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string][]string{"exclude_pattern": clean}
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", "/api/search/monitorsetting/exclude-pattern/part", body, nil); err != nil {
		return err
	}
	fmt.Printf("Removed %d exclude pattern(s).\n", len(clean))
	return nil
}

// dedupTrim drops empty entries and de-duplicates the slice while
// preserving the original order — both PUT and DELETE silently accept
// duplicates, but echoing a sane summary back to the user requires the
// CLI to know how many distinct values it actually sent.
func dedupTrim(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

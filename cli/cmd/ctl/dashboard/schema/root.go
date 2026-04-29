// Package schema hosts the `olares-cli dashboard schema` introspection
// command. The verb is leaf-only (no subcommands) so this directory
// follows the simplest settings-style layout: a single root.go that
// exports the cobra factory.
package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// common holds a pointer to the dashboard root's CommonFlags so the
// schema verb honours --output even though it doesn't talk to any
// remote endpoint.
var common *pkgdashboard.CommonFlags

// NewSchemaCommand registers the `dashboard schema` introspection verb.
// With no args it dumps an index of all known kinds + the corresponding
// schema filenames; with one arg (or a space-separated path) it returns
// the JSON Schema (draft-07) document for that command path.
//
// Two distinct output forms are intentional:
//
//   - `dashboard schema`           → an Envelope-shaped index (kind=
//     dashboard.schema.index, items=[{path, kind, schema_file}]). Mirrors
//     the dual-shape contract; agents can pipe it through the same parser
//     used for any other leaf command.
//
//   - `dashboard schema <path>`    → the raw JSON Schema document for that
//     command. NOT wrapped in an envelope (it's already the meta-level
//     contract; wrapping would self-reference).
func NewSchemaCommand(cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:   "schema [command-path]",
		Short: "Introspect the JSON Schemas served by `olares-cli dashboard`",
		Example: `  # Index of every command + its schema file:
  olares-cli dashboard schema

  # Schema for a specific command (use space-separated path):
  olares-cli dashboard schema overview cpu`,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			return runSchema(c.Context(), c.OutOrStdout(), args)
		},
	}
	return cmd
}

func runSchema(_ context.Context, w io.Writer, args []string) error {
	entries := pkgdashboard.LoadSchemaIndex()

	if len(args) == 0 {
		env := pkgdashboard.Envelope{
			Kind: pkgdashboard.KindSchemaIndex,
			Meta: pkgdashboard.Meta{},
			Items: func() []pkgdashboard.Item {
				items := make([]pkgdashboard.Item, len(entries))
				for i, e := range entries {
					items[i] = pkgdashboard.Item{
						Raw: map[string]any{
							"path":        e.Path,
							"kind":        e.Kind,
							"schema_file": e.File,
						},
						Display: map[string]any{
							"path":        e.Path,
							"kind":        e.Kind,
							"schema_file": e.File,
						},
					}
				}
				return items
			}(),
		}
		switch common.Output {
		case pkgdashboard.OutputJSON:
			return pkgdashboard.WriteJSON(w, env)
		default:
			cols := []pkgdashboard.TableColumn{
				{Header: "PATH", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "path") }},
				{Header: "KIND", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "kind") }},
				{Header: "SCHEMA", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "schema_file") }},
			}
			return pkgdashboard.WriteTable(w, cols, env.Items)
		}
	}

	want := strings.Join(args, " ")
	for _, e := range entries {
		if e.Path == want || e.Path == "dashboard "+want {
			payload, err := pkgdashboard.ReadSchemaFile(e.File)
			if err != nil {
				return fmt.Errorf("read embedded schema %s: %w", e.File, err)
			}
			var doc map[string]any
			if err := json.Unmarshal(payload, &doc); err != nil {
				return fmt.Errorf("parse embedded schema %s: %w", e.File, err)
			}
			out, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return fmt.Errorf("re-marshal schema %s: %w", e.File, err)
			}
			_, err = w.Write(append(out, '\n'))
			return err
		}
	}
	return fmt.Errorf("no schema known for command path %q (try `olares-cli dashboard schema` for the index)", want)
}

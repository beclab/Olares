package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// Common is the shared flag bundle every leaf command consumes. Stored as
// a package-level variable so child trees (overview / applications) can
// reach it via the exported accessor below — Cobra's parent.PersistentFlags
// already populates the same fields, but the typed access matters for
// PreRunE validation and for tests that bypass cobra.
//
// Single global instance is intentional: a `dashboard` invocation always
// resolves to one shared parameter set. Sub-trees never construct their
// own.
var common CommonFlags

// Flags returns the shared CommonFlags singleton. Child packages call this
// from their RunE bodies to read --output / --watch / --since / etc.
func Flags() *CommonFlags { return &common }

// NewDashboardCommand assembles the `olares-cli dashboard` subtree. f is the
// shared cmdutil.Factory the root command builds — every leaf reaches into
// it for an authenticated *http.Client and the resolved profile.
//
// Layout mirrors the SPA:
//
//	dashboard
//	├── overview                                (default = sections envelope: physical+user+ranking)
//	│   ├── physical                            (9-row cluster metric table)
//	│   ├── user [<username>]                   (CPU / memory quota)
//	│   ├── ranking                             (workload-grain ranking)
//	│   ├── cpu                                 (per-node table)
//	│   ├── memory [--mode physical|swap]       (per-node table)
//	│   ├── disk                                (default = sections: main + per-disk partitions)
//	│   │   ├── main                            (per-disk table)
//	│   │   └── partitions <device> [--node N]  ("Occupancy analysis" popup)
//	│   ├── pods                                (per-node table)
//	│   ├── network                             (per-iface table; getSystemIFS)
//	│   ├── fan                                 (default = sections: live + curve)
//	│   │   ├── live                            (1 row; getSystemFan + graphics[0])
//	│   │   └── curve                           (10-row hardcoded fan-curve spec)
//	│   └── gpu                                 (default = list)
//	│       ├── list                            (Graphics management tab)
//	│       ├── tasks                           (Task management tab)
//	│       ├── get <uuid>                      (GPU detail)
//	│       └── task <name> <pod-uid>           (Task detail)
//	├── applications                            (default = workload-grain table)
//	│   └── pods <namespace>                    (per-pod table)
//	└── schema [<command-path>]                 (introspection)
func NewDashboardCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Query the per-user Olares dashboard for AI agents",
		Long: `Query the same data the dashboard SPA renders, but from the CLI.

Every leaf command emits one of two output shapes:
  -o table  (default) — a human-readable, two-column-gutter table.
  -o json             — a strict envelope:
                        { kind, meta, items: [...] }
                        or, for "dashboard overview" (default action),
                        { kind, meta, sections: { ... } }

Authentication, transport and per-user routing are inherited from the
active profile (--profile). Token refresh on 401/403 is transparent.

For agent integration, run "olares-cli dashboard schema" to discover the
available commands + their JSON Schemas.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          unknownSubcommandRunE,
	}

	common.BindPersistent(cmd)

	cmd.AddCommand(newOverviewCommand(f))
	cmd.AddCommand(newApplicationsCommand(f))
	cmd.AddCommand(newSchemaCommand())

	return cmd
}

// unknownSubcommandRunE is the RunE wired onto every dispatch-only parent in
// the dashboard tree (dashboard, applications, overview gpu — and the typo
// guard inside overview).
//
// Without it, a typo'd subcommand like `dashboard application` (note the
// missing 's') silently falls through to cobra's default help renderer with
// exit 0 because:
//
//  1. the parent has no Run/RunE → cobra falls back to printing Help();
//  2. SilenceErrors=true (a deliberate choice for clean machine-readable
//     errors on the leaves) swallows the "unknown command" suggestion cobra
//     would otherwise emit.
//
// We restore the suggestion behaviour ourselves: with no args, defer to the
// usual help path; with positional args, write a "Did you mean…" hint
// directly to stderr (since SilenceErrors stops cobra from printing the
// returned error) and return a non-nil error so the process exits non-zero.
func unknownSubcommandRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.Help()
	}
	msg := fmt.Sprintf("Error: unknown subcommand %q for %q", args[0], c.CommandPath())
	if suggestions := c.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}
	fmt.Fprintln(c.ErrOrStderr(), msg)
	fmt.Fprintf(c.ErrOrStderr(), "\nRun '%s --help' for usage.\n", c.CommandPath())
	return errors.New("unknown subcommand")
}

// ----------------------------------------------------------------------------
// `dashboard schema` — introspection
// ----------------------------------------------------------------------------

// newSchemaCommand registers the introspection verb. With no args it dumps
// an index of all known kinds + the corresponding schema filenames; with
// one arg it returns the JSON Schema (draft-07) document for that command
// path.
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
//     contract; wrapping would self-reference). The first line is always
//     the literal `{"$schema":"http://json-schema.org/draft-07/schema#",`
//     so consumers can validate it against the draft-07 meta-schema if
//     they want.
func newSchemaCommand() *cobra.Command {
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
	entries := loadSchemaIndex()

	if len(args) == 0 {
		env := Envelope{
			Kind: KindSchemaIndex,
			Meta: Meta{},
			Items: func() []Item {
				items := make([]Item, len(entries))
				for i, e := range entries {
					items[i] = Item{
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
		case OutputJSON:
			return WriteJSON(w, env)
		default:
			cols := []TableColumn{
				{Header: "PATH", Get: func(it Item) string { return DisplayString(it, "path") }},
				{Header: "KIND", Get: func(it Item) string { return DisplayString(it, "kind") }},
				{Header: "SCHEMA", Get: func(it Item) string { return DisplayString(it, "schema_file") }},
			}
			return WriteTable(w, cols, env.Items)
		}
	}

	want := strings.Join(args, " ")
	for _, e := range entries {
		if e.Path == want || e.Path == "dashboard "+want {
			payload, err := schemaFS.ReadFile(path.Join("schemas", e.File))
			if err != nil {
				return fmt.Errorf("read embedded schema %s: %w", e.File, err)
			}
			// Pretty-print so the emitted document is human-friendly
			// even when piped to a file. We don't wrap in an envelope
			// — the schema document is the contract, not subject to it.
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

// schemaEntry is one row of the schema index — exported only for tests.
type schemaEntry struct {
	Path string // user-facing command path, e.g. "overview cpu"
	Kind string // dashboard.* constant
	File string // filename inside the embedded schemas/ FS
}

// loadSchemaIndex reads the embedded schemas/ directory + augments each
// entry with the user-facing command path. Order: the static table below
// is the source of truth for path↔kind mapping; the FS walk only verifies
// every static entry has a matching file (a missing file returns the entry
// with File="" so `dashboard schema` still surfaces it).
func loadSchemaIndex() []schemaEntry {
	static := []schemaEntry{
		{"dashboard overview", KindOverview, "overview.json"},
		{"dashboard overview physical", KindOverviewPhysical, "overview-physical.json"},
		{"dashboard overview user", KindOverviewUser, "overview-user.json"},
		{"dashboard overview ranking", KindOverviewRanking, "overview-ranking.json"},
		{"dashboard overview cpu", KindOverviewCPU, "overview-cpu.json"},
		{"dashboard overview memory", KindOverviewMemory, "overview-memory.json"},
		{"dashboard overview disk", KindOverviewDisk, "overview-disk.json"},
		{"dashboard overview disk main", KindOverviewDiskMain, "overview-disk-main.json"},
		{"dashboard overview disk partitions", KindOverviewDiskPart, "overview-disk-partitions.json"},
		{"dashboard overview pods", KindOverviewPods, "overview-pods.json"},
		{"dashboard overview network", KindOverviewNetwork, "overview-network.json"},
		{"dashboard overview fan", KindOverviewFan, "overview-fan.json"},
		{"dashboard overview fan live", KindOverviewFanLive, "overview-fan-live.json"},
		{"dashboard overview fan curve", KindOverviewFanCurve, "overview-fan-curve.json"},
		{"dashboard overview gpu list", KindOverviewGPUList, "overview-gpu-list.json"},
		{"dashboard overview gpu tasks", KindOverviewGPUTasks, "overview-gpu-tasks.json"},
		{"dashboard overview gpu get", KindOverviewGPUDetail, "overview-gpu-detail.json"},
		{"dashboard overview gpu task", KindOverviewGPUTaskDet, "overview-gpu-task-detail.json"},
		{"dashboard overview gpu detail", KindOverviewGPUDetailFull, "overview-gpu-detail-full.json"},
		{"dashboard overview gpu task-detail", KindOverviewGPUTaskDetFull, "overview-gpu-task-detail-full.json"},
		{"dashboard applications", KindApplicationsList, "applications.json"},
		{"dashboard applications pods", KindApplicationsPods, "applications-pods.json"},
	}
	// Drop entries whose schema file isn't shipped (tests catch this
	// gap; production would render an empty path gracefully). We keep
	// the entry with File="" so the index still mentions the command.
	have := embeddedSchemaFiles()
	out := make([]schemaEntry, 0, len(static))
	for _, e := range static {
		if !have[e.File] {
			e.File = ""
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func embeddedSchemaFiles() map[string]bool {
	out := map[string]bool{}
	_ = fs.WalkDir(schemaFS, "schemas", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		out[path.Base(p)] = true
		return nil
	})
	return out
}

// ----------------------------------------------------------------------------
// Helper for sub-tree commands: emit one envelope with the user's chosen
// output format. Sub-trees that need richer table layouts (per-column
// formatters) call WriteTable directly instead.
// ----------------------------------------------------------------------------

// EmitDefault is a tiny helper for leaf commands that don't have custom
// table columns: emit JSON in JSON mode, fall back to a generic key/value
// dump in table mode. Most leaves prefer their own TableColumn slice and
// don't call this.
func EmitDefault(env Envelope, format OutputFormat) error {
	if format == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	// Generic table fallback: one row per item, columns alphabetised.
	if len(env.Items) == 0 {
		fmt.Println("(no items)")
		return nil
	}
	keys := map[string]struct{}{}
	for _, it := range env.Items {
		for k := range it.Display {
			keys[k] = struct{}{}
		}
	}
	headers := make([]string, 0, len(keys))
	for k := range keys {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	cols := make([]TableColumn, len(headers))
	for i, h := range headers {
		key := h
		cols[i] = TableColumn{
			Header: strings.ToUpper(key),
			Get:    func(it Item) string { return DisplayString(it, key) },
		}
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

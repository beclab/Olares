package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search status`
//
// Backed by /api/search/task/stats/merged, which user-service @All-proxies
// to search3. The SPA polls this every 10s and reads two fields off the
// inner data: `status` (a free-form label like "completed"/"running") and
// `full_content_task_error` (best-effort failure list / map per file
// indexer). See settings/src/pages/settings/Search/FileSearch.vue:244
// `const { status, full_content_task_error } = await getSearchTaskStatus()`.
//
// The backend returns `data: {status, full_content_task_error, ...}`
// (object). We decode into searchTaskStats and surface both fields;
// full_content_task_error is held as json.RawMessage so future schema
// bumps (currently object or null) round-trip through `--output json`
// without further code churn.
func NewStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show the search index task status",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// searchTaskStats mirrors the inner shape of /api/search/task/stats/merged
// after the BFL/search3 envelope is unwrapped. Only the two fields the
// SPA reads today are typed; everything else round-trips via the JSON
// output path on `--output json`.
type searchTaskStats struct {
	Status               string          `json:"status"`
	FullContentTaskError json.RawMessage `json:"full_content_task_error,omitempty"`
}

func runStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var s searchTaskStats
	if err := doGetEnvelope(ctx, pc.doer, "/api/search/task/stats/merged", &s); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, s)
	default:
		return renderStatus(os.Stdout, s)
	}
}

func renderStatus(w io.Writer, s searchTaskStats) error {
	status := s.Status
	if status == "" {
		status = "(unknown)"
	}
	if _, err := fmt.Fprintf(w, "Status:   %s\n", status); err != nil {
		return err
	}
	failures := formatTaskFailures(s.FullContentTaskError)
	if _, err := fmt.Fprintf(w, "Failures: %s\n", failures); err != nil {
		return err
	}
	return nil
}

// formatTaskFailures collapses the upstream full_content_task_error
// payload into one human-readable line. Today the SPA renders this as a
// list section; the CLI table view only has room for a summary, and the
// raw bytes are still available via `--output json`.
func formatTaskFailures(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "(none)"
	}
	trimmed := string(raw)
	switch trimmed {
	case "null", "{}", "[]", `""`:
		return "(none)"
	}
	// best-effort: try to count entries when it's an array or an object
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		if len(arr) == 0 {
			return "(none)"
		}
		return fmt.Sprintf("%d (use -o json for details)", len(arr))
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		if len(obj) == 0 {
			return "(none)"
		}
		return fmt.Sprintf("%d (use -o json for details)", len(obj))
	}
	return "(use -o json for details)"
}

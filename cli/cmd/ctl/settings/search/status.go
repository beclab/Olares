package search

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search status`
//
// Backed by /api/search/task/stats/merged, which user-service @All-proxies
// to search3. The SPA polls this every 10s and treats the value as a
// status string ("completed" / "running"). The wire body is search3's
// {code, message, data: string} envelope.
//
// Phase 2 will add `rebuild` (POST /api/search/task/rebuild).
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

	var status string
	if err := doGetEnvelope(ctx, pc.doer, "/api/search/task/stats/merged", &status); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, map[string]string{"status": status})
	default:
		return renderStatus(os.Stdout, status)
	}
}

func renderStatus(w io.Writer, status string) error {
	if status == "" {
		status = "(unknown)"
	}
	_, err := fmt.Fprintf(w, "Status: %s\n", status)
	return err
}

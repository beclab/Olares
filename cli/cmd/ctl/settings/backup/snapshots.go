package backup

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings backup snapshots ...`
//
// Backed by /apis/backup/v1/plans/backup/<backup-id>/snapshots?offset=&limit=.
// Returns a BFL envelope around { snapshots: [...], totalCount: <int> }.
//
// The SPA's BackupDetail page passes a backupId from the URL, so we
// take it as a positional arg.
//
// Phase 6 adds `snapshots get <backup-id> <snapshot-id>`,
// `snapshots create <backup-id>`, and `snapshots cancel <backup-id> <snapshot-id>`.
func NewSnapshotsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "backup snapshots for a single plan",
		Long: `Inspect snapshots taken by a backup plan.

Subcommands:
  list <backup-id>                                        (Phase 1)

Subcommands landing in Phase 6:
  get <backup-id> <snapshot-id>,
  create <backup-id>,
  cancel <backup-id> <snapshot-id>
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSnapshotsListCommand(f))
	return cmd
}

func newSnapshotsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var offset, limit int
	cmd := &cobra.Command{
		Use:   "list <backup-id>",
		Short: "list snapshots for a single backup plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSnapshotsList(c.Context(), f, args[0], offset, limit, output)
		},
	}
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset")
	cmd.Flags().IntVar(&limit, "limit", 50, "pagination limit (matches the SPA default)")
	addOutputFlag(cmd, &output)
	return cmd
}

// backupSnapshot mirrors the BackupSnapshot TypeScript interface
// (constant/index.ts: SnapshotInfo + status + progress).
type backupSnapshot struct {
	ID       string `json:"id"`
	CreateAt int64  `json:"createAt"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type backupSnapshotListResponse struct {
	Snapshots  []backupSnapshot `json:"snapshots"`
	TotalCount int              `json:"totalCount"`
}

func runSnapshotsList(ctx context.Context, f *cmdutil.Factory, backupID string, offset, limit int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if backupID == "" {
		return fmt.Errorf("backup-id is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("offset", fmt.Sprintf("%d", offset))
	q.Set("limit", fmt.Sprintf("%d", limit))
	path := fmt.Sprintf("/apis/backup/v1/plans/backup/%s/snapshots?%s",
		url.PathEscape(backupID), q.Encode())

	var resp backupSnapshotListResponse
	if err := doGetEnvelope(ctx, pc.doer, path, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderSnapshotsTable(os.Stdout, resp.Snapshots)
	}
}

func renderSnapshotsTable(w io.Writer, rows []backupSnapshot) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no snapshots")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSTATUS\tPROGRESS\tSIZE\tCREATED"); err != nil {
		return err
	}
	for _, s := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d%%\t%s\t%s\n",
			nonEmpty(s.ID),
			nonEmpty(s.Status),
			s.Progress,
			humanBytes(s.Size),
			fmtUnix(s.CreateAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// humanBytes mirrors the helper in settings/advanced — duplicated
// rather than imported to keep area packages independent (matching
// the convention of per-area common.go files).
func humanBytes(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffix := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), suffix)
}

package backup

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
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
// Per-snapshot delete (a non-cancel remove of a completed snapshot) is
// intentionally not exposed: backup-server does not have a route for it
// and the SPA only wires the cancel button on running / pending
// snapshots.
func NewSnapshotsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "backup snapshots for a single plan",
		Long: `Inspect or trigger snapshots taken by a backup plan.

Subcommands:
  list   <backup-id>
  run    <backup-id>
  cancel <backup-id> <snapshot-id> [--yes]
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSnapshotsListCommand(f))
	cmd.AddCommand(newSnapshotsRunCommand(f))
	cmd.AddCommand(newSnapshotsCancelCommand(f))
	return cmd
}

func newSnapshotsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var offset, limit int
	cmd := &cobra.Command{
		Use:   "list <backup-id>",
		Short: "list snapshots for a single backup plan (run 'backup plans list' first to find a <backup-id>)",
		Long: `List snapshots taken by a single backup plan.

Argument shape: <backup-id> is REQUIRED — snapshots are stored
plan-scoped upstream (/apis/backup/v1/plans/backup/<id>/snapshots) so
there is no "all snapshots" endpoint to list. Run

  olares-cli settings backup plans list

first to discover the plan IDs, then pass one of them here.

Examples:
  olares-cli settings backup snapshots list <backup-id>
  olares-cli settings backup snapshots list <backup-id> -o json --limit 200
`,
		Args: cobra.ExactArgs(1),
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
//
// Wire note: `size` is a *decimal string* on the wire even though the
// SPA's BackupSnapshot interface declares `size: number` — backup-server
// formats it via handlers.ParseSnapshotSize (`fmt.Sprintf("%d", *size)`,
// see framework/backup-server/pkg/handlers/helper.go), so we decode as
// string and humanize at render time. The same shape applies to
// `backupPlan.Size` in plans.go.
//
// `progress` is an integer in *basis points* (0–10000 where 10000 =
// 100.00%); we render it via formatProgressBP at the table layer, not
// here, so JSON output preserves the raw wire value for downstream
// scripts that prefer to do their own math. The same convention
// applies to backupPlan.Progress (plans.go) and restorePlan.Progress
// (restore/plans.go).
type backupSnapshot struct {
	ID       string `json:"id"`
	CreateAt int64  `json:"createAt"`
	Size     string `json:"size"`
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
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(s.ID),
			nonEmpty(s.Status),
			formatProgressBP(s.Progress),
			humanBytesString(s.Size),
			fmtUnix(s.CreateAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// newSnapshotsRunCommand wraps POST /apis/backup/v1/plans/backup/<id>/snapshots
// with body {event: "create"} — the SPA's "Run now" button on
// BackupDetail (BackupDetail.vue:325).
func newSnapshotsRunCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "run <backup-id>",
		Short: "trigger an ad-hoc snapshot for a backup plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSnapshotsRun(c.Context(), f, args[0])
		},
	}
}

func runSnapshotsRun(ctx context.Context, f *cmdutil.Factory, backupID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	backupID = strings.TrimSpace(backupID)
	if backupID == "" {
		return fmt.Errorf("run requires a backup-id")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/apis/backup/v1/plans/backup/%s/snapshots",
		url.PathEscape(backupID))
	body := map[string]string{"event": "create"}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", path, body, nil); err != nil {
		return err
	}
	fmt.Printf("Triggered ad-hoc snapshot for backup plan %q.\n", backupID)
	return nil
}

// newSnapshotsCancelCommand wraps DELETE /apis/backup/v1/plans/backup/<bid>/snapshots/<sid>
// with body {event: "cancel"}. Note: backup-server uses DELETE-with-body
// (BackupSnapshotDetail.vue:176-179); axios's `data:` quirk made the
// SPA explicit about it, and we match that by sending the body even
// though DELETE-with-body is unusual.
func newSnapshotsCancelCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "cancel <backup-id> <snapshot-id>",
		Short: "cancel a running or pending snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runSnapshotsCancel(c.Context(), f, args[0], args[1], assumeYes)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	return cmd
}

func runSnapshotsCancel(ctx context.Context, f *cmdutil.Factory, backupID, snapshotID string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	backupID = strings.TrimSpace(backupID)
	snapshotID = strings.TrimSpace(snapshotID)
	if backupID == "" || snapshotID == "" {
		return fmt.Errorf("cancel requires both backup-id and snapshot-id")
	}
	if !assumeYes {
		if err := confirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf("Cancel snapshot %q on plan %q?", snapshotID, backupID)); err != nil {
			return err
		}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/apis/backup/v1/plans/backup/%s/snapshots/%s",
		url.PathEscape(backupID), url.PathEscape(snapshotID))
	body := map[string]string{"event": "cancel"}
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", path, body, nil); err != nil {
		return err
	}
	fmt.Printf("Cancelled snapshot %q on plan %q.\n", snapshotID, backupID)
	return nil
}

// humanBytesString accepts the raw `size` field as it arrives on the
// wire (a decimal byte count formatted as a string by backup-server,
// e.g. "1234567"). On a successful int64 parse it formats via
// humanBytes; otherwise it falls back to the trimmed raw value, or
// "-" when empty — so a future backend that switches to pre-formatted
// human strings (e.g. "12.06 KiB") still renders cleanly.
func humanBytesString(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "-"
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return humanBytes(n)
	}
	return s
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

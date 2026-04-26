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

// `olares-cli settings backup plans ...`
//
// Backed by /apis/backup/v1/plans/backup?offset=&limit= on the BFL
// backup-server. The SPA's Settings -> Backup page calls this with a
// fixed limit of 50 and pages forward via offset; Phase 1 keeps the
// same UX (single page, --limit / --offset flags).
//
// Phase 6 will add `plans get / create / update / delete` and the
// pause/resume verbs.
func NewPlansCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "backup plans (Settings -> Backup)",
		Long: `Manage backup plans on the BFL backup-server.

Subcommands:
  list   list backup plans                                (Phase 1)

Subcommands landing in Phase 6:
  get <id>, create, update, delete <id>,
  pause <id>, resume <id>
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPlansListCommand(f))
	return cmd
}

func newPlansListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var offset, limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list backup plans",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPlansList(c.Context(), f, offset, limit, output)
		},
	}
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset")
	cmd.Flags().IntVar(&limit, "limit", 50, "pagination limit (matches the SPA default)")
	addOutputFlag(cmd, &output)
	return cmd
}

// backupPlan mirrors the BackupPlan TypeScript interface in
// apps/.../constant/index.ts. We deliberately keep only the fields
// the table view uses; the JSON output round-trips the raw object,
// so unknown fields aren't lost there.
type backupPlan struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	BackupType          string `json:"backupType"`
	BackupAppTypeName   string `json:"backupAppTypeName"`
	Size                string `json:"size"`
	RestoreSize         string `json:"restoreSize"`
	Path                string `json:"path"`
	Progress            int    `json:"progress"`
	NextBackupTimestamp int64  `json:"nextBackupTimestamp"`
	Location            string `json:"location"`
	LocationConfigName  string `json:"locationConfigName"`
	Status              string `json:"status"`
	CreateAt            int64  `json:"createAt"`
}

type backupPlanListResponse struct {
	Backups []backupPlan `json:"backups"`
	// totalCount isn't surfaced by the SPA list view; we still decode
	// it best-effort so a future paginated table can use it.
	TotalCount int `json:"totalCount"`
}

func runPlansList(ctx context.Context, f *cmdutil.Factory, offset, limit int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
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
	path := "/apis/backup/v1/plans/backup?" + q.Encode()

	var resp backupPlanListResponse
	if err := doGetEnvelope(ctx, pc.doer, path, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderPlansTable(os.Stdout, resp.Backups)
	}
}

func renderPlansTable(w io.Writer, rows []backupPlan) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no backup plans")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tID\tTYPE\tLOCATION\tSTATUS\tPROGRESS\tSIZE\tNEXT-RUN\tCREATED"); err != nil {
		return err
	}
	for _, p := range rows {
		typ := nonEmpty(p.BackupType)
		if p.BackupAppTypeName != "" {
			typ = fmt.Sprintf("%s/%s", typ, p.BackupAppTypeName)
		}
		loc := nonEmpty(p.Location)
		if p.LocationConfigName != "" {
			loc = fmt.Sprintf("%s/%s", loc, p.LocationConfigName)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d%%\t%s\t%s\t%s\n",
			nonEmpty(p.Name),
			nonEmpty(p.ID),
			typ,
			loc,
			nonEmpty(p.Status),
			p.Progress,
			nonEmpty(p.Size),
			fmtUnix(p.NextBackupTimestamp),
			fmtUnix(p.CreateAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

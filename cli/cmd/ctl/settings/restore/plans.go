package restore

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

// `olares-cli settings restore plans ...`
//
// Backed by /apis/backup/v1/plans/restore?offset=&limit= on the BFL
// backup-server. Returns a BFL envelope around { restores: [...] }
// (and probably a totalCount). The SPA pages forward via offset
// with a fixed limit of 50; Phase 1 mirrors that.
func NewPlansCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "restore plans (Settings -> Restore)",
		Long: `Manage restore plans on the BFL backup-server.

Subcommands:
  list   list restore plans                               (Phase 1)

Subcommands landing in Phase 6:
  get <id>, create-from-snapshot, create-from-url, cancel <id>,
  check-url <url> --password <pw>
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
		Short: "list restore plans",
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

// restorePlan mirrors the RestorePlan TypeScript interface in
// apps/.../constant/index.ts.
type restorePlan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	CreateAt          int64  `json:"createAt"`
	EndAt             int64  `json:"endAt"`
	SnapshotTime      int64  `json:"snapshotTime"`
	Progress          int    `json:"progress"`
	Status            string `json:"status"`
	BackupAppTypeName string `json:"backupAppTypeName"`
	BackupType        string `json:"backupType"`
}

type restorePlanListResponse struct {
	Restores   []restorePlan `json:"restores"`
	TotalCount int           `json:"totalCount"`
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
	path := "/apis/backup/v1/plans/restore?" + q.Encode()

	var resp restorePlanListResponse
	if err := doGetEnvelope(ctx, pc.doer, path, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderPlansTable(os.Stdout, resp.Restores)
	}
}

func renderPlansTable(w io.Writer, rows []restorePlan) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no restore plans")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tID\tTYPE\tSTATUS\tPROGRESS\tSNAPSHOT-TIME\tCREATED\tEND"); err != nil {
		return err
	}
	for _, p := range rows {
		typ := nonEmpty(p.BackupType)
		if p.BackupAppTypeName != "" {
			typ = fmt.Sprintf("%s/%s", typ, p.BackupAppTypeName)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d%%\t%s\t%s\t%s\n",
			nonEmpty(p.Name),
			nonEmpty(p.ID),
			typ,
			nonEmpty(p.Status),
			p.Progress,
			fmtUnix(p.SnapshotTime),
			fmtUnix(p.CreateAt),
			fmtUnix(p.EndAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

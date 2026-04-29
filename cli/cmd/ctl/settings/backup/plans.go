package backup

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings backup plans ...`
//
// Backed by /apis/backup/v1/plans/backup?offset=&limit= on the BFL
// backup-server. The SPA's Settings -> Backup page calls this with a
// fixed limit of 50 and pages forward via offset; the CLI keeps the
// same UX (single page, --limit / --offset flags).
//
// Plan create + policy update are intentionally out of scope: their
// wire shape requires a full BackupPolicy + LocationConfig vector that
// the SPA assembles from a multi-step form. Encoding that in CLI flags
// would be a poor UX; we'd rather wait for either a
// `--from-file plan.json` mode or for the upstream to expose a
// higher-level "create from defaults" shortcut.
func NewPlansCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "backup plans (Settings -> Backup)",
		Long: `Manage backup plans on the BFL backup-server.

Subcommands:
  list                          list backup plans
  delete <id> [--yes]           delete a backup plan
  pause  <id>                   pause a backup plan
  resume <id>                   resume a paused plan

Out of scope for now (need a richer flag/file UX before shipping):
  create / update    (full BackupPolicy + LocationConfig vector)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPlansListCommand(f))
	cmd.AddCommand(newPlansDeleteCommand(f))
	cmd.AddCommand(newPlansPauseCommand(f))
	cmd.AddCommand(newPlansResumeCommand(f))
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

// newPlansDeleteCommand wraps DELETE /apis/backup/v1/plans/backup/<id>.
// The SPA wires this to the trash button on each plan row in
// BackupDetail.vue:404. We add a destructive-confirmation prompt
// (consistent with `vpn devices delete`) because deleting a plan also
// orphans every snapshot it produced.
func newPlansDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "delete a backup plan",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPlansDelete(c.Context(), f, args[0], assumeYes)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	return cmd
}

func runPlansDelete(ctx context.Context, f *cmdutil.Factory, id string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("delete requires a plan id")
	}
	if !assumeYes {
		if err := confirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf("Delete backup plan %q (and orphan its snapshots)?", id)); err != nil {
			return err
		}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/apis/backup/v1/plans/backup/" + url.PathEscape(id)
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("Deleted backup plan %q.\n", id)
	return nil
}

func newPlansPauseCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id>",
		Short: "pause a backup plan (skips upcoming scheduled runs)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPlansEvent(c.Context(), f, args[0], "pause")
		},
	}
}

func newPlansResumeCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id>",
		Short: "resume a paused backup plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPlansEvent(c.Context(), f, args[0], "resume")
		},
	}
}

func runPlansEvent(ctx context.Context, f *cmdutil.Factory, id, event string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%s requires a plan id", event)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/apis/backup/v1/plans/backup/" + url.PathEscape(id)
	body := map[string]string{"event": event}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", path, body, nil); err != nil {
		return err
	}
	verb := "Paused"
	if event == "resume" {
		verb = "Resumed"
	}
	fmt.Printf("%s backup plan %q.\n", verb, id)
	return nil
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

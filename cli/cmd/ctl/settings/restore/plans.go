package restore

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

// `olares-cli settings restore plans ...`
//
// Backed by /apis/backup/v1/plans/restore?offset=&limit= on the BFL
// backup-server. Returns a BFL envelope around { restores: [...] }
// (and probably a totalCount). The SPA pages forward via offset
// with a fixed limit of 50; Phase 1 mirrors that.
//
// Phase 6 lands `check-url`, `create-from-snapshot`, `create-from-url`,
// and `cancel`. `update` and `delete` (non-cancel) are intentionally
// out of scope: backup-server has no routes for them — see
// `framework/backup-server/pkg/modules/backup/v1/register.go`.
func NewPlansCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "restore plans (Settings -> Restore)",
		Long: `Manage restore plans on the BFL backup-server.

Subcommands:
  list                                                          (Phase 1)
  check-url --backup-url URL [--password ... | --password-stdin]
                                                                (Phase 6)
  create-from-snapshot --snapshot-id SID --path PATH            (Phase 6)
  create-from-url --backup-url URL --path PATH [--dir DIR]
                  [--password ... | --password-stdin]           (Phase 6)
  cancel <id> [--yes]                                           (Phase 6)

Out of scope (no backup-server route):
  update / delete (non-cancel)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPlansListCommand(f))
	cmd.AddCommand(newCheckURLCommand(f))
	cmd.AddCommand(newCreateFromSnapshotCommand(f))
	cmd.AddCommand(newCreateFromURLCommand(f))
	cmd.AddCommand(newPlansCancelCommand(f))
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

// newCheckURLCommand wraps POST /apis/backup/v1/plans/restore/checkurl —
// the SPA's "validate this restic-style URL" probe. Returns the
// candidate snapshot list along with totalCount; we render either a
// summary table (default) or the raw JSON.
func newCheckURLCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		backupURL    string
		password     string
		passwordStd  bool
		offset, limit int
		output       string
	)
	cmd := &cobra.Command{
		Use:   "check-url",
		Short: "list snapshots reachable at a restic / kopia URL",
		Long: `Probe a restic / kopia-style backup URL with a password and return the
list of snapshots backup-server can see at that URL.

Examples:
  olares-cli settings restore plans check-url \
    --backup-url s3:s3.amazonaws.com/bucket/repo --password "$REPO_PW"

  echo -n "$REPO_PW" | olares-cli settings restore plans check-url \
    --backup-url s3:s3.amazonaws.com/bucket/repo --password-stdin
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCheckURL(c.Context(), f, backupURL, password, passwordStd, offset, limit, output)
		},
	}
	cmd.Flags().StringVar(&backupURL, "backup-url", "", "restic/kopia-style URL to probe (required)")
	cmd.Flags().StringVar(&password, "password", "", "repository password (use --password-stdin in scripts)")
	cmd.Flags().BoolVar(&passwordStd, "password-stdin", false, "read the repository password from stdin once")
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset on the candidate snapshot list")
	cmd.Flags().IntVar(&limit, "limit", 50, "pagination limit on the candidate snapshot list")
	addOutputFlag(cmd, &output)
	_ = cmd.MarkFlagRequired("backup-url")
	return cmd
}

type checkURLRequest struct {
	BackupURL string `json:"backupUrl"`
	Password  string `json:"password"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

type checkURLSnapshot struct {
	ID           string `json:"id"`
	Time         string `json:"time"`
	SnapshotTime int64  `json:"snapshotTime"`
	Size         int64  `json:"size"`
}

type checkURLResponse struct {
	BackupPath string             `json:"backupPath"`
	BackupType string             `json:"backupType"`
	TotalCount int                `json:"totalCount"`
	Snapshots  []checkURLSnapshot `json:"snapshots"`
}

func runCheckURL(ctx context.Context, f *cmdutil.Factory, backupURL, password string, passwordStdin bool, offset, limit int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pw, err := readPasswordOnce(password, passwordStdin, "Repository password: ")
	if err != nil {
		return err
	}
	if pw == "" {
		return fmt.Errorf("password is required (use --password or --password-stdin)")
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := checkURLRequest{
		BackupURL: backupURL,
		Password:  pw,
		Offset:    offset,
		Limit:     limit,
	}
	var resp checkURLResponse
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/apis/backup/v1/plans/restore/checkurl", body, &resp); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderCheckURL(os.Stdout, resp)
}

func renderCheckURL(w io.Writer, r checkURLResponse) error {
	fmt.Fprintf(w, "BACKUP-TYPE: %s\nBACKUP-PATH: %s\nTOTAL:       %d\n\n",
		nonEmpty(r.BackupType), nonEmpty(r.BackupPath), r.TotalCount)
	if len(r.Snapshots) == 0 {
		_, err := fmt.Fprintln(w, "no snapshots reachable at this URL")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tTIME\tSNAPSHOT-TIME\tSIZE"); err != nil {
		return err
	}
	for _, s := range r.Snapshots {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n",
			nonEmpty(s.ID), nonEmpty(s.Time), fmtUnix(s.SnapshotTime), s.Size); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// newCreateFromSnapshotCommand wraps POST /apis/backup/v1/plans/restore
// with body { snapshotId, path } — the SPA's "restore an existing
// snapshot" flow (RestoreExistingBackup.vue:115). The SPA disables
// this route in routes-settings.ts:323-327 today, but the underlying
// API still accepts it; we keep the verb so CLI scripts that already
// know a snapshot id can drive a restore without going through a URL
// probe.
func newCreateFromSnapshotCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		snapshotID string
		path       string
	)
	cmd := &cobra.Command{
		Use:   "create-from-snapshot",
		Short: "create a restore plan from an existing snapshot id",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCreateFromSnapshot(c.Context(), f, snapshotID, path)
		},
	}
	cmd.Flags().StringVar(&snapshotID, "snapshot-id", "", "snapshot id from `settings backup snapshots list` (required)")
	cmd.Flags().StringVar(&path, "path", "", "restore path on the target Olares (required)")
	_ = cmd.MarkFlagRequired("snapshot-id")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runCreateFromSnapshot(ctx context.Context, f *cmdutil.Factory, snapshotID, path string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	snapshotID = strings.TrimSpace(snapshotID)
	path = strings.TrimSpace(path)
	if snapshotID == "" || path == "" {
		return fmt.Errorf("--snapshot-id and --path are required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]string{"snapshotId": snapshotID, "path": path}
	var resp restorePlan
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/apis/backup/v1/plans/restore", body, &resp); err != nil {
		return err
	}
	fmt.Printf("Created restore plan %q from snapshot %q.\n", resp.ID, snapshotID)
	return nil
}

// newCreateFromURLCommand wraps POST /apis/backup/v1/plans/restore
// with body { backupUrl, password, path, dir? } — the SPA's
// "restore from a custom URL" flow (MultiRestoreOptions.vue:340-348).
func newCreateFromURLCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		backupURL    string
		password     string
		passwordStd  bool
		path         string
		dir          string
	)
	cmd := &cobra.Command{
		Use:   "create-from-url",
		Short: "create a restore plan from a restic / kopia-style URL",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCreateFromURL(c.Context(), f, backupURL, password, passwordStd, path, dir)
		},
	}
	cmd.Flags().StringVar(&backupURL, "backup-url", "", "restic/kopia-style URL (required)")
	cmd.Flags().StringVar(&password, "password", "", "repository password (use --password-stdin in scripts)")
	cmd.Flags().BoolVar(&passwordStd, "password-stdin", false, "read the repository password from stdin once")
	cmd.Flags().StringVar(&path, "path", "", "restore path on the target Olares (required)")
	cmd.Flags().StringVar(&dir, "dir", "", "subdirectory inside the backup to restore (optional)")
	_ = cmd.MarkFlagRequired("backup-url")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runCreateFromURL(ctx context.Context, f *cmdutil.Factory, backupURL, password string, passwordStdin bool, path, dir string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	backupURL = strings.TrimSpace(backupURL)
	path = strings.TrimSpace(path)
	if backupURL == "" || path == "" {
		return fmt.Errorf("--backup-url and --path are required")
	}
	pw, err := readPasswordOnce(password, passwordStdin, "Repository password: ")
	if err != nil {
		return err
	}
	if pw == "" {
		return fmt.Errorf("password is required (use --password or --password-stdin)")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]string{
		"backupUrl": backupURL,
		"password":  pw,
		"path":      path,
	}
	if dir != "" {
		body["dir"] = dir
	}
	var resp restorePlan
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/apis/backup/v1/plans/restore", body, &resp); err != nil {
		return err
	}
	fmt.Printf("Created restore plan %q from URL.\n", resp.ID)
	return nil
}

// newPlansCancelCommand wraps DELETE /apis/backup/v1/plans/restore/<id>
// with body {event: "cancel"} — same DELETE-with-body convention as
// the snapshot cancel verb (axios's `data:` second-argument).
func newPlansCancelCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "cancel <id>",
		Short: "cancel a running or pending restore plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPlansCancel(c.Context(), f, args[0], assumeYes)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	return cmd
}

func runPlansCancel(ctx context.Context, f *cmdutil.Factory, id string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("cancel requires a plan id")
	}
	if !assumeYes {
		if err := confirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf("Cancel restore plan %q?", id)); err != nil {
			return err
		}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/apis/backup/v1/plans/restore/" + url.PathEscape(id)
	body := map[string]string{"event": "cancel"}
	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", path, body, nil); err != nil {
		return err
	}
	fmt.Printf("Cancelled restore plan %q.\n", id)
	return nil
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

package download

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewSyncCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app     string
		limit   int
		since   string
		sinceID int64
		all     bool
		output  string
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "incremental cursor sync of download tasks",
		Long: `Incrementally pull tasks by a composite cursor (GET /api/download/sync).

The server keys the feed on (updated_at, id): it returns rows whose
updated_at is newer than --since, or equal to --since with an id greater
than --since-id, in (updated_at, id) ascending order. Unlike an id-only
feed this DOES surface progress updates to tasks you already saw, because
any change bumps updated_at. Remember the cursor of the last row you saw
(printed after each page) and pass it back as --since / --since-id to fetch
only newer changes.

--since accepts the local time shown in the list/sync table (e.g.
2026-07-15T23:03 or "2026-07-15 23:03"), a bare date (2026-07-15), or a
zoned RFC3339 value (2026-07-15T15:03:00Z); a value without a zone is read
in your local timezone, so no UTC math is needed. Omit it for a full drain
from the beginning. --all drains every page (advancing the cursor for you)
and prints the combined result; without --all one page is fetched and, if
more remain, the next --since / --since-id cursor is printed.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSync(c.Context(), f, app, limit, since, sinceID, all, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().IntVar(&limit, "limit", 0, "page size (0 = server default, max 100)")
	cmd.Flags().StringVar(&since, "since", "", "cursor lower bound on updated_at; local time (2026-07-15T23:03), a date, or RFC3339 (empty = from start)")
	cmd.Flags().Int64Var(&sinceID, "since-id", 0, "cursor tie-breaker: id lower bound for rows whose updated_at equals --since")
	cmd.Flags().BoolVar(&all, "all", false, "drain every page, advancing the cursor for you")
	return cmd
}

func runSync(ctx context.Context, f *cmdutil.Factory, app string, limit int, sinceRaw string, sinceID int64, all bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	since, err := parseSince(sinceRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if all {
		return runSyncAll(ctx, pc, app, limit, since, sinceID, format)
	}
	res, err := fetchSyncPage(ctx, pc, app, limit, since, sinceID)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		if err := renderTasksTable(os.Stdout, res.Items); err != nil {
			return err
		}
		if res.HasMore {
			nextSince, nextID := res.NextCursor()
			fmt.Printf("\nmore available; next --since %s --since-id %d\n",
				nextSince.UTC().Format(time.RFC3339Nano), nextID)
		}
		return nil
	}
}

// sinceZonedLayouts carry an explicit zone and are parsed as-is (this is how
// the next-cursor line we print round-trips exactly). sinceLocalLayouts omit a
// zone and are interpreted in the machine's local timezone, so a user can paste
// the local time the list/sync table prints without doing any UTC math — the
// query param is still converted to UTC downstream (see fetchSyncPage).
var (
	sinceZonedLayouts = []string{time.RFC3339Nano, time.RFC3339}
	sinceLocalLayouts = []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
)

// parseSince turns the --since flag into a cursor timestamp. An empty string is
// the zero time (full drain). Inputs carrying a zone (RFC3339) are honoured
// as-is; zone-less inputs (e.g. "2026-07-15T23:03" or "2026-07-15 23:03:05",
// matching the table's local-time column) are read in the local timezone.
func parseSince(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range sinceZonedLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	for _, layout := range sinceLocalLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid --since %q (want local time e.g. 2026-07-15T23:03 or \"2026-07-15 23:03\", a date 2026-07-15, or RFC3339 2026-07-15T15:03:00Z)", raw)
}

func fetchSyncPage(ctx context.Context, pc *preparedClient, app string, limit int, since time.Time, sinceID int64) (SyncResult, error) {
	q := url.Values{}
	if a := strings.TrimSpace(app); a != "" {
		q.Set("app", a)
	}
	if !since.IsZero() {
		q.Set("since", since.UTC().Format(time.RFC3339Nano))
	}
	if sinceID > 0 {
		q.Set("since_id", strconv.FormatInt(sinceID, 10))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var res SyncResult
	if err := doGet(ctx, pc.doer, "/api/download/sync"+encodeQuery(q), &res); err != nil {
		return SyncResult{}, err
	}
	return res, nil
}

func runSyncAll(ctx context.Context, pc *preparedClient, app string, limit int, since time.Time, sinceID int64, format Format) error {
	var acc []DownloadTask
	curSince, curID := since, sinceID
	for {
		res, err := fetchSyncPage(ctx, pc, app, limit, curSince, curID)
		if err != nil {
			return err
		}
		acc = append(acc, res.Items...)
		nextSince, nextID := res.NextCursor()
		// Stop on the last page, an empty page, or a cursor that fails to
		// advance, so a misbehaving server can never spin us forever.
		if !res.HasMore || len(res.Items) == 0 ||
			(!nextSince.After(curSince) && nextID <= curID) {
			break
		}
		curSince, curID = nextSince, nextID
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, SyncResult{Items: acc, HasMore: false})
	default:
		return renderTasksTable(os.Stdout, acc)
	}
}

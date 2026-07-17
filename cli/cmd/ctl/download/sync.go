package download

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewSyncCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app    string
		limit  int
		after  int64
		all    bool
		output string
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "incremental cursor sync of download tasks",
		Long: `Incrementally pull tasks by id cursor (GET /api/download/sync).

sync returns tasks whose id is greater than --after (ascending order),
including finished ones. It is an incremental full pull keyed on id, not a
change feed: it does not surface progress updates to tasks you already saw,
only tasks with a larger id. Remember the largest id you saw and pass it as
--after next time to fetch only newer tasks.

--all drains every page (following the server cursor) and prints the
combined result. Without --all, one page is fetched and, if more pages
remain, the next --after cursor is printed.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSync(c.Context(), f, app, limit, after, all, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().IntVar(&limit, "limit", 0, "page size (0 = server default, max 500)")
	cmd.Flags().Int64Var(&after, "after", 0, "cursor: only tasks with id greater than this (0 = from start)")
	cmd.Flags().BoolVar(&all, "all", false, "drain every page following the server cursor")
	return cmd
}

func runSync(ctx context.Context, f *cmdutil.Factory, app string, limit int, after int64, all bool, outputRaw string) error {
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
	if all {
		return runSyncAll(ctx, pc, app, limit, after, format)
	}
	res, err := fetchSyncPage(ctx, pc, app, limit, after)
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
			fmt.Printf("\nmore available; next --after %d\n", res.NextCursor)
		}
		return nil
	}
}

func fetchSyncPage(ctx context.Context, pc *preparedClient, app string, limit int, after int64) (SyncResult, error) {
	q := url.Values{}
	if a := strings.TrimSpace(app); a != "" {
		q.Set("app", a)
	}
	if after > 0 {
		q.Set("after", strconv.FormatInt(after, 10))
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

func runSyncAll(ctx context.Context, pc *preparedClient, app string, limit int, after int64, format Format) error {
	var acc []DownloadTask
	cursor := after
	last := after
	for {
		res, err := fetchSyncPage(ctx, pc, app, limit, cursor)
		if err != nil {
			return err
		}
		acc = append(acc, res.Items...)
		if len(res.Items) > 0 {
			last = res.NextCursor
		}
		// Stop on the last page or when the cursor fails to advance, so a
		// misbehaving server can never spin us in an infinite loop.
		if !res.HasMore || len(res.Items) == 0 || res.NextCursor <= cursor {
			break
		}
		cursor = res.NextCursor
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, SyncResult{Items: acc, HasMore: false, NextCursor: last})
	default:
		return renderTasksTable(os.Stdout, acc)
	}
}

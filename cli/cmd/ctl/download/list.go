package download

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app      string
		status   string
		page     int
		pageSize int
		all      bool
		allApps  bool
		output   string
	)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list download tasks",
		Long: `List download tasks for the active profile.

--all pages through /api/download/list until every matching row is
collected (distinct from sync --all, which drains the sync cursor).
--all-apps lists across every app; it cannot be combined with an
explicit --app. Without --all-apps, --app defaults to larepass.

--status is validated locally against the server task-status enum;
illegal values fail before any request.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			appChanged := c.Flags().Changed("app")
			return runList(c.Context(), f, app, status, page, pageSize, all, allApps, appChanged, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&status, "status", "", "filter by status (one of: "+validTaskStatusValues+")")
	cmd.Flags().IntVar(&page, "page", 0, "page number (0 = server default; ignored with --all)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (0 = server default; with --all, chunk size defaults to 100)")
	cmd.Flags().BoolVar(&all, "all", false, "fetch every page for the current filters (not sync --all)")
	cmd.Flags().BoolVar(&allApps, "all-apps", false, "list tasks across all apps (mutually exclusive with --app)")
	return cmd
}

func runList(ctx context.Context, f *cmdutil.Factory, app, status string, page, pageSize int, all, allApps, appChanged bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if allApps && appChanged {
		return fmt.Errorf("--all-apps cannot be combined with --app")
	}
	appFilter := ""
	if !allApps {
		app, err = validateApp(app)
		if err != nil {
			return err
		}
		appFilter = app
	}
	if err := validateTaskStatus(status); err != nil {
		return err
	}
	if err := validateNonNegativeFlag("--page", page); err != nil {
		return err
	}
	if err := validateNonNegativeFlag("--page-size", pageSize); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var result ListResult
	if all {
		result, err = fetchListAll(ctx, pc, appFilter, status, pageSize)
	} else {
		result, err = fetchListPage(ctx, pc, appFilter, status, page, pageSize)
	}
	if err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderListTable(os.Stdout, result)
	}
}

func fetchListPage(ctx context.Context, pc *preparedClient, app, status string, page, pageSize int) (ListResult, error) {
	q := url.Values{}
	if app != "" {
		q.Set("app", app)
	}
	if s := strings.TrimSpace(status); s != "" {
		q.Set("status", s)
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	var result ListResult
	if err := doGet(ctx, pc.doer, "/api/download/list"+encodeQuery(q), &result); err != nil {
		return ListResult{}, err
	}
	return result, nil
}

func fetchListAll(ctx context.Context, pc *preparedClient, app, status string, pageSize int) (ListResult, error) {
	if pageSize <= 0 {
		pageSize = listPageSizeDefault
	}
	// The server clamps a larger request down without saying so, which
	// would make the short-page guard below stop after one page.
	if pageSize > listPageSizeMax {
		pageSize = listPageSizeMax
	}
	var acc []DownloadTask
	var total int64
	page := 1
	for {
		res, err := fetchListPage(ctx, pc, app, status, page, pageSize)
		if err != nil {
			return ListResult{}, err
		}
		if page == 1 {
			total = res.Total
		}
		acc = append(acc, res.List...)
		if len(res.List) == 0 || int64(len(acc)) >= total {
			break
		}
		// Guard against a misbehaving server that never advances.
		if len(res.List) < pageSize {
			break
		}
		page++
	}
	return ListResult{List: acc, Total: total}, nil
}

func renderListTable(w io.Writer, result ListResult) error {
	if err := renderTasksTable(w, result.List); err != nil {
		return err
	}
	if result.Total > 0 {
		fmt.Fprintf(w, "\n%d of %d\n", len(result.List), result.Total)
	}
	return nil
}

const (
	taskTableColPad    = 2
	taskTableNameWidth = 40
	taskTableURLWidth  = 48
)

// renderTasksTable prints the shared task table (list / sync / unfinished).
// Columns are padded by terminal display width, not rune count: a CJK
// title is two columns per character, and tabwriter treating it as one
// shoved SOURCE to the right of every ASCII row.
func renderTasksTable(w io.Writer, tasks []DownloadTask) error {
	headers := []string{"ID", "STATUS", "PROVIDER", "PERCENT", "NAME", "SOURCE", "APP", "UPDATED"}
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		rows = append(rows, []string{
			strconv.FormatInt(t.ID, 10),
			orDash(t.Status),
			orDash(t.DownloadProvider),
			fmt.Sprintf("%.1f%%", t.Percent),
			clipDisplay(displayName(t), taskTableNameWidth),
			clipDisplay(orDash(t.URL), taskTableURLWidth),
			orDash(t.App),
			formatTime(t.UpdatedAt),
		})
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if n := runewidth.StringWidth(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}
	if err := writeDisplayRow(w, headers, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writeDisplayRow(w, row, widths); err != nil {
			return err
		}
	}
	return nil
}

func writeDisplayRow(w io.Writer, cells []string, widths []int) error {
	var b strings.Builder
	for i, cell := range cells {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", taskTableColPad))
		}
		b.WriteString(padDisplay(cell, widths[i]))
	}
	b.WriteByte('\n')
	_, err := io.WriteString(w, b.String())
	return err
}

func padDisplay(s string, n int) string {
	w := runewidth.StringWidth(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

// clipDisplay clips s to n terminal columns. CJK titles are two
// columns wide; cutting on runes left NAME visually overflowing the
// budget and broke the SOURCE column for every other row.
func clipDisplay(s string, n int) string {
	if n <= 0 || runewidth.StringWidth(s) <= n {
		return s
	}
	return runewidth.Truncate(s, n, "...")
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

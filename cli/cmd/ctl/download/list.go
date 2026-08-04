package download

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app      string
		status   string
		page     int
		pageSize int
		output   string
	)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list download tasks",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), f, app, status, page, pageSize, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&status, "status", "", "filter by status (downloading, paused, …)")
	cmd.Flags().IntVar(&page, "page", 0, "page number (0 = server default)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (0 = server default)")
	return cmd
}

func runList(ctx context.Context, f *cmdutil.Factory, app, status string, page, pageSize int, outputRaw string) error {
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

	q := url.Values{}
	if a := strings.TrimSpace(app); a != "" {
		q.Set("app", a)
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
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderListTable(os.Stdout, result)
	}
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

// renderTasksTable prints the shared task table (list / sync / unfinished).
func renderTasksTable(w io.Writer, tasks []DownloadTask) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tPROVIDER\tPERCENT\tNAME\tAPP\tUPDATED")
	for _, t := range tasks {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			orDash(t.Status),
			orDash(t.DownloadProvider),
			fmt.Sprintf("%.1f%%", t.Percent),
			truncate(displayName(t), 48),
			orDash(t.App),
			formatTime(t.UpdatedAt),
		)
	}
	return tw.Flush()
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

// formatUnix renders a unix-seconds timestamp (cookies / settings) or "-".
func formatUnix(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	return time.Unix(sec, 0).Local().Format("2006-01-02 15:04")
}

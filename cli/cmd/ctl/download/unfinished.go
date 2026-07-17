package download

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewUnfinishedCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app    string
		output string
	)
	cmd := &cobra.Command{
		Use:   "unfinished",
		Short: "list tasks that have not reached a terminal state",
		Long: `List unfinished download tasks (GET /api/download/unfinished).

Returns tasks that are not in a terminal state (downloading / pending /
paused / …), optionally filtered by --app.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUnfinished(c.Context(), f, app, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	return cmd
}

func runUnfinished(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
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
	var result ListResult
	if err := doGet(ctx, pc.doer, "/api/download/unfinished"+encodeQuery(q), &result); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderListTable(os.Stdout, result)
	}
}

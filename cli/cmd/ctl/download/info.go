package download

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewInfoCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "info <id>",
		Short: "show download task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runInfo(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runInfo(ctx context.Context, f *cmdutil.Factory, idRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var task DownloadTask
	path := fmt.Sprintf("/api/download/info/%d", id)
	if err := doGet(ctx, pc.doer, path, &task); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, task)
	default:
		return renderInfo(os.Stdout, task)
	}
}

func renderInfo(w io.Writer, t DownloadTask) error {
	fields := [][2]string{
		{"ID", strconv.FormatInt(t.ID, 10)},
		{"Status", orDash(t.Status)},
		{"Provider", orDash(t.DownloadProvider)},
		{"App", orDash(t.App)},
		{"Percent", fmt.Sprintf("%.1f%%", t.Percent)},
		{"Size", strconv.FormatInt(t.Size, 10)},
		{"Downloaded", strconv.FormatInt(t.DownloadedBytes, 10)},
		{"Path", orDash(t.Path)},
		{"Name", displayName(t)},
		{"URL", orDash(t.URL)},
		{"Error", orDash(t.ErrMsg)},
		{"ErrCategory", orDash(t.ErrCategory)},
		{"Created", formatTime(t.CreatedAt)},
		{"Updated", formatTime(t.UpdatedAt)},
	}
	for _, kv := range fields {
		fmt.Fprintf(w, "%-12s %s\n", kv[0]+":", kv[1])
	}
	return nil
}

func parseTaskID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid task id %q (need a positive integer)", raw)
	}
	return id, nil
}

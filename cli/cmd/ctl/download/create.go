package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app      string
		path     string
		name     string
		quality  string
		formatID string
		extraRaw string
		output   string
	)
	cmd := &cobra.Command{
		Use:   "create <url>",
		Short: "create a download task",
		Long: `Create a download task (POST /api/download).

--quality maps to extra.ytdlp_quality; --format-id maps to extra.format_id.
--extra accepts a JSON object of string values merged into extra (wins last).
--path accepts drive/Home/... (or a files resource URL).`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCreate(c.Context(), f, args[0], app, path, name, quality, formatID, extraRaw, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", "", "destination path (drive/Home/... or files resource URL)")
	cmd.Flags().StringVar(&name, "name", "", "suggested file_name")
	cmd.Flags().StringVar(&quality, "quality", "", "yt-dlp quality preset (best, 2160p, 1080p, …)")
	cmd.Flags().StringVar(&formatID, "format-id", "", "yt-dlp format_id override")
	cmd.Flags().StringVar(&extraRaw, "extra", "", "JSON object merged into extra (string values)")
	return cmd
}

func runCreate(ctx context.Context, f *cmdutil.Factory, rawURL, app, path, name, quality, formatID, extraRaw, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	app = strings.TrimSpace(app)
	if app == "" {
		app = defaultApp
	}

	extra := map[string]string{}
	if strings.TrimSpace(extraRaw) != "" {
		var parsed map[string]string
		if err := json.Unmarshal([]byte(extraRaw), &parsed); err != nil {
			return fmt.Errorf("--extra must be a JSON object of string values: %w", err)
		}
		for k, v := range parsed {
			extra[k] = v
		}
	}
	if q := strings.TrimSpace(quality); q != "" {
		extra["ytdlp_quality"] = q
	}
	if fid := strings.TrimSpace(formatID); fid != "" {
		extra["format_id"] = fid
	}

	req := NewDownloadReq{
		URL:      rawURL,
		App:      app,
		Path:     strings.TrimSpace(path),
		FileName: strings.TrimSpace(name),
	}
	if len(extra) > 0 {
		req.Extra = extra
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var task DownloadTask
	if err := doMutate(ctx, pc.doer, "POST", "/api/download", req, &task); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, task)
	default:
		fmt.Printf("Created task %d  status=%s  provider=%s  name=%s\n",
			task.ID, task.Status, task.DownloadProvider, displayName(task))
		return nil
	}
}

package download

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app         string
		path        string
		name        string
		quality     string
		formatID    string
		extraRaw    string
		torrentFile string
		selectFiles string
		output      string
	)
	cmd := &cobra.Command{
		Use:   "create [url]",
		Short: "create a download task",
		Long: `Create a download task (POST /api/download).

Quote the URL. A URL with ?, & or = must be wrapped in single quotes,
otherwise the shell splits it on & and drops the query string:
  olares-cli knowledge download create 'https://host/v?a=1&b=2'

Torrent / magnet:
  A magnet link is passed as an ordinary URL argument:
    olares-cli knowledge download create 'magnet:?xt=urn:btih:...'
  A local .torrent file is uploaded with --torrent (base64); the URL
  argument may be omitted in that case:
    olares-cli knowledge download create --torrent ./x.torrent --select-files 1,3
--torrent reads a local .torrent file and sends it as extra.torrent_file_b64.
--select-files takes a comma-separated list of 1-based file indices (as
reported by "torrent inspect"), passed through as extra.selected_files.

--quality maps to extra.ytdlp_quality (one of: ` + ytdlpQualityValues + `).
--format-id maps to extra.format_id.
--extra accepts a JSON object of string values merged into extra (wins last).
--path must start with drive/Home/ or drive/Data/, e.g.
  --path drive/Home/Pictures/
The first segment is literally "drive"; the second is "Home" or "Data"
(case-sensitive). A full API URL also works:
  --path 'https://files.<user>.olares.cn/api/resources/drive/Home/Pictures/'
NOT accepted: the browser address like .../Files/Home/... , or a bare
Home/... without the drive/ prefix (both fail as unsupported file type).
Defaults to ` + defaultDownloadPath + ` (aligned with wise). Pass --path ""
to send an empty path (e.g. HuggingFace cache mode) and let the server decide.

HuggingFace: the destination is picked by extra._hf_dest, not by --path/--name.
  local (default when _hf_dest is unset): lands under <path>/<repoID>/. --path
         applies; --name is unnecessary (the repo id is the folder name).
  cache: shared HF_HOME (Files UI: /Common/huggingface/). --path and --name are
         ignored; send --path "" to match wise.
Set HF options via --extra (keys map to hf CLI flags), e.g.:
  --extra '{"_hf_dest":"cache"}'
  --extra '{"_hf_dest":"local","token":"hf_xxx","revision":"v1.0","include":"*.safetensors"}'
Note wise defaults HF to cache; this CLI defaults to local unless you pass _hf_dest.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			rawURL := ""
			if len(args) > 0 {
				rawURL = args[0]
			}
			return runCreate(c.Context(), f, rawURL, app, path, name, quality, formatID, extraRaw, torrentFile, selectFiles, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&path, "path", defaultDownloadPath, "destination starting with drive/Home/ or drive/Data/ (e.g. drive/Home/Pictures/); \"\" lets the server decide")
	cmd.Flags().StringVar(&name, "name", "", "suggested file_name (ignored for HuggingFace: repo id / cache layout wins)")
	cmd.Flags().StringVar(&quality, "quality", "", "yt-dlp quality preset (one of: "+ytdlpQualityValues+")")
	cmd.Flags().StringVar(&formatID, "format-id", "", "yt-dlp format_id override")
	cmd.Flags().StringVar(&extraRaw, "extra", "", "JSON object merged into extra (string values)")
	cmd.Flags().StringVar(&torrentFile, "torrent", "", "local .torrent file to upload (base64); the URL argument may be omitted")
	cmd.Flags().StringVar(&selectFiles, "select-files", "", "comma-separated 1-based file indices for a multi-file torrent (e.g. 1,3,5)")
	return cmd
}

func runCreate(ctx context.Context, f *cmdutil.Factory, rawURL, app, path, name, quality, formatID, extraRaw, torrentFile, selectFiles, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	rawURL = strings.TrimSpace(rawURL)
	torrentFile = strings.TrimSpace(torrentFile)
	if rawURL == "" && torrentFile == "" {
		return fmt.Errorf("provide a URL/magnet argument or --torrent <file>")
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
	if torrentFile != "" {
		raw, err := os.ReadFile(torrentFile)
		if err != nil {
			return fmt.Errorf("read torrent file: %w", err)
		}
		extra["torrent_file_b64"] = base64.StdEncoding.EncodeToString(raw)
	}
	if sf := strings.TrimSpace(selectFiles); sf != "" {
		extra["selected_files"] = sf
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

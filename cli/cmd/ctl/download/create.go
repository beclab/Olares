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
		Long: `Create a download task from a URL, magnet link, or local .torrent file.

Quote the URL. A URL with ?, & or = must be wrapped in single quotes,
otherwise your shell may split it.

Torrent / magnet:
  A magnet link is passed as the URL argument. For a local .torrent file,
  use --torrent and omit the URL.
--select-files takes a comma-separated list of 1-based file indices (as
reported by "torrent inspect"). Pass --select-files all, or omit the flag,
to download every file.

--quality accepts: ` + ytdlpQualityValues + `.
--format-id selects a specific yt-dlp format.
--extra accepts additional provider options as a JSON object of strings.
--path is normalized like download-server CreateFileParam:
  drive/Home/… or drive/Data/…, /api/resources/drive/…, or a full Files
  API URL. "Home" and "Data" are case-sensitive. Browser Files UI paths
  and bare Home/… are rejected. The default is ` + defaultDownloadPath + `.
Pass --path "" when the provider should choose the destination.

For HuggingFace, set _hf_dest in --extra:
  local (default) downloads under <path>/<repoID>/.
  cache downloads to the shared HuggingFace cache and ignores --path/--name.

Re-submitting the same URL always creates a new task (no duplicate 409).`,
		Example: `  # URL
  olares-cli knowledge download create 'https://host/v?a=1&b=2'

  # magnet link
  olares-cli knowledge download create 'magnet:?xt=urn:btih:...'

  # selected files from a local torrent
  olares-cli knowledge download create --torrent ./x.torrent --select-files 1,3

  # HuggingFace shared cache
  olares-cli knowledge download create https://huggingface.co/org/repo --path "" --extra '{"_hf_dest":"cache"}'`,
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
	cmd.Flags().StringVar(&path, "path", defaultDownloadPath, "destination: drive/Home|Data/…, Files API URL, or \"\" for HF cache")
	cmd.Flags().StringVar(&name, "name", "", "suggested file_name (ignored for HuggingFace: repo id / cache layout wins)")
	cmd.Flags().StringVar(&quality, "quality", "", "yt-dlp quality preset (one of: "+ytdlpQualityValues+")")
	cmd.Flags().StringVar(&formatID, "format-id", "", "yt-dlp format_id override")
	cmd.Flags().StringVar(&extraRaw, "extra", "", "JSON object merged into extra (string values)")
	cmd.Flags().StringVar(&torrentFile, "torrent", "", "local .torrent file to upload (base64); the URL argument may be omitted")
	cmd.Flags().StringVar(&selectFiles, "select-files", "", "comma-separated 1-based file indices for a multi-file torrent (e.g. 1,3,5), or \"all\" (= omit = every file)")
	return cmd
}

func validateYTDLPQuality(raw string, required bool) error {
	quality := strings.TrimSpace(raw)
	if quality == "" {
		if required {
			return fmt.Errorf("--quality is required")
		}
		return nil
	}
	switch quality {
	case "best", "2160p", "1080p", "720p", "480p", "360p", "audio":
		return nil
	default:
		return fmt.Errorf("unsupported --quality %q (allowed: %s)", raw, ytdlpQualityValues)
	}
}

// normalizeSelectFiles validates and normalises the --select-files flag into
// the extra.selected_files CSV, using the same validator as
// `torrent files --select` so both paths agree. It returns ok=false (omit the
// key) for a blank input or an empty/"all" selection — on create an absent
// selected_files already means "download every file" — and a non-nil error for
// bad tokens (0, negatives, non-integers), which fail locally instead of
// round-tripping to a server 400.
func normalizeSelectFiles(raw string) (csv string, ok bool, err error) {
	if strings.TrimSpace(raw) == "" {
		return "", false, nil
	}
	sel, err := parseSelectedIndices(raw)
	if err != nil {
		return "", false, err
	}
	if len(sel) == 0 {
		return "", false, nil
	}
	return joinIndicesCSV(sel), true, nil
}

// readTorrentFile validates a local torrent path before upload: it must
// end in .torrent (case-insensitive), exist, and be non-empty. flag is the
// CLI flag name used in error messages (e.g. "--torrent" or "--file") so
// create and torrent inspect can share the validator without pointing at
// the wrong flag. A missing path gets a quote hint for shell users.
func readTorrentFile(path, flag string) ([]byte, error) {
	if flag == "" {
		flag = "--torrent"
	}
	if !strings.HasSuffix(strings.ToLower(path), ".torrent") {
		return nil, fmt.Errorf("unsupported %s file %q (need a .torrent file); if the path contains spaces, quote it: %s './name with spaces.torrent'", flag, path, flag)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("read torrent file: %w; if the path contains spaces, quote it: %s './name with spaces.torrent'", err, flag)
		}
		return nil, fmt.Errorf("read torrent file: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("unsupported %s file %q (file is empty)", flag, path)
	}
	return raw, nil
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
	app, err = validateApp(app)
	if err != nil {
		return err
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
	if err := validateYTDLPQuality(quality, false); err != nil {
		return err
	}
	if q := strings.TrimSpace(quality); q != "" {
		extra["ytdlp_quality"] = q
	}
	if q, ok := extra["ytdlp_quality"]; ok {
		if err := validateYTDLPQuality(q, false); err != nil {
			return fmt.Errorf("invalid ytdlp_quality in --extra: %w", err)
		}
	}
	if fid := strings.TrimSpace(formatID); fid != "" {
		extra["format_id"] = fid
	}
	if torrentFile != "" {
		raw, err := readTorrentFile(torrentFile, "--torrent")
		if err != nil {
			return err
		}
		extra["torrent_file_b64"] = base64.StdEncoding.EncodeToString(raw)
	}
	csv, ok, err := normalizeSelectFiles(selectFiles)
	if err != nil {
		return err
	}
	if ok {
		extra["selected_files"] = csv
	}

	normalizedPath, err := normalizeDownloadPath(path, true)
	if err != nil {
		return err
	}

	req := NewDownloadReq{
		URL:      rawURL,
		App:      app,
		Path:     normalizedPath,
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

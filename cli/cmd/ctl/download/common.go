// Package download hosts `olares-cli knowledge download` — the download-server
// task centre via Settings edge `https://settings.<terminus>/download/...`.
// Distinct from top-level `download` (installer packages) and `files download`.
package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

const (
	minOlaresVersion = "1.12.7"
	defaultApp       = "wise"
	// allowedApps is the --app whitelist (kept in sync with validateApp).
	// Default remains wise; larepass is the TermiPass / LarePass namespace.
	allowedApps = "wise, larepass"
	// ytdlpQualityValues is the server-accepted --quality enum
	// (download-server services.IsValidYtdlpQuality). Kept here so the
	// --help text stays in sync with what the backend validates.
	ytdlpQualityValues = "best, 2160p, 1080p, 720p, 480p, 360p, audio"
	// defaultDownloadPath mirrors the wise front-end default landing
	// directory (Termipass collect-site downloadFile: opts.path ||
	// 'Downloads/', normalised to drive/Home/Downloads/). The manager
	// itself applies no default (empty task.Path lands at the PVC root),
	// so the CLI seeds this to match what wise users see.
	defaultDownloadPath = "drive/Home/Downloads/"
	// ytdlpMarketInstall is the next-step CTA when the yt-dlp provider
	// is unreachable (often not installed). Chart name matches the
	// Market listing / shared service namespace ytdlpv3-shared.
	ytdlpMarketInstall = "olares-cli market install ytdlpv3"
	// syncLimitMax matches the download-server sync page-size cap.
	syncLimitMax = 100
)

var taskActionPathRE = regexp.MustCompile(`^/api/download/(?:pause|resume|cancel|info)/(\d+)(?:/|$)`)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func parseFormat(s string) (Format, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "output format: table, json")
}

func addAppFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVar(target, "app", defaultApp, "app namespace for the download task (one of: "+allowedApps+"; default: wise)")
}

// validateApp trims --app, defaults empty to wise, and rejects values
// outside the whitelist. Returns the normalised app name.
func validateApp(raw string) (string, error) {
	app := strings.TrimSpace(raw)
	if app == "" {
		return defaultApp, nil
	}
	switch app {
	case "wise", "larepass":
		return app, nil
	default:
		return "", fmt.Errorf("unsupported --app %q (allowed: %s)", raw, allowedApps)
	}
}

// validateNonNegativeFlag rejects negative paging / size flags. Zero
// means "server default" and stays valid.
func validateNonNegativeFlag(name string, v int) error {
	if v < 0 {
		return fmt.Errorf("unsupported %s %d (need >= 0; 0 = server default)", name, v)
	}
	return nil
}

// validateLimit rejects negative sync --limit and enforces the server
// max of 100. Zero means "server default".
func validateLimit(limit int) error {
	if limit < 0 {
		return fmt.Errorf("unsupported --limit %d (need >= 0; 0 = server default, max %d)", limit, syncLimitMax)
	}
	if limit > syncLimitMax {
		return fmt.Errorf("unsupported --limit %d (max %d)", limit, syncLimitMax)
	}
	return nil
}

// validateSinceID rejects a negative --since-id. Zero is omitted from
// the query (full drain / no tie-breaker).
func validateSinceID(id int64) error {
	if id < 0 {
		return fmt.Errorf("unsupported --since-id %d (need >= 0)", id)
	}
	return nil
}

// validateDrivePath requires drive/Home/ or drive/Data/ (case-sensitive),
// matching the create --path contract used by download-server.
func validateDrivePath(raw string) error {
	path := strings.TrimSpace(raw)
	if path == "" {
		return fmt.Errorf("--path is required")
	}
	if strings.HasPrefix(path, "drive/Home/") || strings.HasPrefix(path, "drive/Data/") {
		return nil
	}
	return fmt.Errorf("unsupported --path %q (need a path starting with drive/Home/ or drive/Data/)", raw)
}

// ytdlpUnavailableHint is printed when inspect reports Available=false
// (or create fails because the yt-dlp daemon is unreachable).
func ytdlpUnavailableHint() string {
	return "yt-dlp is not available; install it with `" + ytdlpMarketInstall + "`"
}

// shouldHintYTDLPUnavailable reports whether a server message points at
// a missing / unreachable yt-dlp provider.
func shouldHintYTDLPUnavailable(message string) bool {
	lower := strings.ToLower(message)
	if !(strings.Contains(lower, "yt-dlp") || strings.Contains(lower, "ytdlp")) {
		return false
	}
	return strings.Contains(lower, "unavailable") ||
		strings.Contains(lower, "unreachable") ||
		strings.Contains(lower, "not available") ||
		strings.Contains(lower, "not installed") ||
		strings.Contains(lower, "missing") ||
		strings.Contains(lower, "no such") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "dial")
}

// shouldHintInspectTimeout reports whether an inspect failure looks like
// the server's short probe deadline (channel / RSS URLs).
func shouldHintInspectTimeout(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "deadline exceeded") ||
		strings.Contains(lower, "timed out")
}

// Doer is the JSON transport used by download verbs (whoami.HTTPClient).
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
}

// edgeBase is SettingsURL + "/download" (user-service DownloadController strip prefix).
func edgeBase(rp *credential.ResolvedProfile) string {
	if rp == nil {
		return ""
	}
	return strings.TrimRight(rp.SettingsURL, "/") + "/download"
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: download not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "knowledge download",
		MinVersion: minOlaresVersion,
		Reason:     "settings /download edge + download provider",
	}); err != nil {
		return nil, err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedClient{
		profile: rp,
		doer:    whoami.NewHTTPClient(hc, edgeBase(rp), rp.OlaresID),
	}, nil
}

// dsEnvelope is download-server's response shape: success code 200 (or 0),
// single object in data, list+total (or list+has_more) at the top level.
type dsEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	List    json.RawMessage `json:"list"`
	Total   *int64          `json:"total"`
	HasMore *bool           `json:"has_more"`
	// Batch lifecycle responses carry succeeded/failed at the top level
	// (alongside code), not under data — see BatchResult.
	Succeeded json.RawMessage `json:"succeeded"`
	Failed    json.RawMessage `json:"failed"`
}

func doGet(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutate(ctx, d, "GET", path, nil, out)
}

func doMutate(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env dsEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		var httpErr *whoami.HTTPError
		if errors.As(err, &httpErr) {
			if recovery := taskErrorRecovery(method, path, httpErr.Status, string(httpErr.Body), body); recovery != "" {
				return fmt.Errorf("%w%s", err, recovery)
			}
		}
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		recovery := taskErrorRecovery(method, path, env.Code, msg, body)
		if msg == "" {
			return fmt.Errorf("%s %s: code %d%s", method, path, env.Code, recovery)
		}
		return fmt.Errorf("%s %s: code %d: %s%s", method, path, env.Code, msg, recovery)
	}
	if out == nil {
		return nil
	}
	// List-shaped responses: decode into *ListResult (or any type that
	// expects {"list","total","has_more"}).
	if lr, ok := out.(*ListResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &lr.List); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.Total != nil {
			lr.Total = *env.Total
		}
		return nil
	}
	if cl, ok := out.(*CookieListResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &cl.List); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.Total != nil {
			cl.Total = *env.Total
		}
		return nil
	}
	if br, ok := out.(*BatchResult); ok {
		if len(env.Succeeded) > 0 {
			if err := json.Unmarshal(env.Succeeded, &br.Succeeded); err != nil {
				return fmt.Errorf("%s %s: decode succeeded: %w", method, path, err)
			}
		}
		if len(env.Failed) > 0 {
			if err := json.Unmarshal(env.Failed, &br.Failed); err != nil {
				return fmt.Errorf("%s %s: decode failed: %w", method, path, err)
			}
		}
		return nil
	}
	// Sync-shaped responses: top-level {list, has_more}. Same "list" slot
	// as the list endpoint, so decode from env.List (not env.Data) plus the
	// has_more flag; the composite cursor is derived client-side.
	if sr, ok := out.(*SyncResult); ok {
		if len(env.List) > 0 {
			if err := json.Unmarshal(env.List, &sr.Items); err != nil {
				return fmt.Errorf("%s %s: decode list: %w", method, path, err)
			}
		}
		if env.HasMore != nil {
			sr.HasMore = *env.HasMore
		}
		return nil
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

func taskErrorRecovery(method, path string, status int, message string, body interface{}) string {
	lowerMsg := strings.ToLower(message)
	taskID := ""
	if taskPathMatch := taskActionPathRE.FindStringSubmatch(path); len(taskPathMatch) > 0 {
		taskID = taskPathMatch[1]
	} else if path == "/api/download/remove" {
		switch req := body.(type) {
		case RemoveReq:
			taskID = fmt.Sprintf("%d", req.TaskID)
		case *RemoveReq:
			if req != nil {
				taskID = fmt.Sprintf("%d", req.TaskID)
			}
		}
	}
	switch {
	case method == "POST" && path == "/api/download" && status == 409 &&
		(strings.Contains(lowerMsg, "already") ||
			strings.Contains(lowerMsg, "exist") ||
			strings.Contains(lowerMsg, "registered")):
		return "; inspect existing tasks with `olares-cli knowledge download list`"
	case method == "POST" && path == "/api/download" && shouldHintYTDLPUnavailable(message):
		return "; " + ytdlpUnavailableHint()
	case strings.Contains(path, "/api/url/inspect") && shouldHintInspectTimeout(message):
		return "; channel/RSS probes can exceed the server inspect timeout; retry later or create the task directly if the URL is known good"
	case taskID != "" && status == 409:
		return fmt.Sprintf(
			"; wait for the move to finish, then retry; inspect progress with `olares-cli knowledge download info %s`",
			taskID,
		)
	case taskID != "" &&
		(status == 404 || strings.Contains(lowerMsg, "task not found")):
		return "; refresh task IDs with `olares-cli knowledge download list`"
	}
	return ""
}

func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func encodeQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func displayName(t DownloadTask) string {
	if strings.TrimSpace(t.FileName) != "" {
		return t.FileName
	}
	if strings.TrimSpace(t.URL) != "" {
		return t.URL
	}
	return "-"
}

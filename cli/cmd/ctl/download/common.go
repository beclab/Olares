// Package download hosts `olares-cli knowledge download` — the download-server
// task centre via Settings edge `https://settings.<terminus>/download/...`.
// Distinct from top-level `download` (installer packages) and `files download`.
package download

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

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
	// listPageSizeDefault matches download-server TaskRepository
	// defaultPageSize so list --all chunks align with the server.
	listPageSizeDefault = 100
	// listPageSizeMax matches download-server TaskRepository maxPageSize.
	// The server clamps silently, so --all must page at the real size or
	// it stops early and under-reports.
	listPageSizeMax = 1000
	// headerIdempotencyKey is the create replay key (RFC 9110 style).
	headerIdempotencyKey = "Idempotency-Key"
)

// waitPollInterval is how often wait / create --wait re-queries info.
// Mutable in tests so polling cases finish quickly.
var waitPollInterval = 2 * time.Second

const (
	// waitDefaultTimeout matches the market / users watch commands so
	// every long-poll surface in the CLI gives up after the same wait.
	waitDefaultTimeout = 15 * time.Minute
	// waitMaxConsecErrors is the transient-error budget before wait
	// gives up, mirroring market watch.
	waitMaxConsecErrors = 5
)

// validTaskStatuses mirrors download-server models.validTaskStatuses
// (IsValidTaskStatus). Kept as a CSV for --help and a set for local
// --status validation so illegal values fail before any HTTP call.
const validTaskStatusValues = "downloading, paused, cancelled, error, completed, waiting, removed, preparing, waiting_to_move, moving, seeding"

var validTaskStatusSet = map[string]struct{}{
	"downloading":     {},
	"paused":          {},
	"cancelled":       {},
	"error":           {},
	"completed":       {},
	"waiting":         {},
	"removed":         {},
	"preparing":       {},
	"waiting_to_move": {},
	"moving":          {},
	"seeding":         {},
}

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

// validateTaskStatus rejects unknown --status values locally so the
// CLI fails closed instead of round-tripping to an empty list.
func validateTaskStatus(raw string) error {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if _, ok := validTaskStatusSet[s]; !ok {
		return fmt.Errorf("unsupported --status %q (allowed: %s)", raw, validTaskStatusValues)
	}
	return nil
}

// newIdempotencyKey returns a fresh create key for one user invoke.
func newIdempotencyKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("cli-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

const resourcesPrefix = "/api/resources/"

// normalizeDownloadPath mirrors download-server CreateFileParam for CLI
// fail-fast: empty (when allowed), bare drive/Home|Data/..., /api/resources/...
// , or a full Files API URL. Returns the bare drive/... form to POST.
func normalizeDownloadPath(raw string, allowEmpty bool) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("--path is required")
	}
	resource, err := extractDriveResourcePath(path)
	if err != nil {
		return "", err
	}
	if !isDriveHomeOrDataPath(resource) {
		return "", fmt.Errorf("unsupported --path %q (need drive/Home/… or drive/Data/…, a Files API /api/resources/… URL, or \"\" for HF cache)", raw)
	}
	return resource, nil
}

func extractDriveResourcePath(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("unsupported --path %q (invalid URL): %w", raw, err)
		}
		s = u.Path
	}
	// Anchored on purpose: only a leading /api/resources/ is the Files API
	// prefix. A bare drive/… destination may legitimately contain those
	// segments (a folder named api), and matching mid-string would truncate
	// the path and then reject it.
	s = strings.TrimPrefix(s, resourcesPrefix)
	s = strings.TrimLeft(s, "/")
	if s == "" {
		return "", fmt.Errorf("unsupported --path %q (no resource path)", raw)
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == ".." {
			return "", fmt.Errorf("unsupported --path %q (path traversal not allowed)", raw)
		}
	}
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("unsupported --path %q (need drive/Home/… or drive/Data/…)", raw)
	}
	if strings.ToLower(parts[0]) != "drive" {
		return "", fmt.Errorf("unsupported --path %q (only drive destinations are accepted)", raw)
	}
	if parts[1] != "Home" && parts[1] != "Data" {
		return "", fmt.Errorf("unsupported --path %q (drive extend must be Home or Data)", raw)
	}
	// Keep the case-sensitive wire form the server expects.
	parts[0] = "drive"
	return strings.Join(parts, "/"), nil
}

func isDriveHomeOrDataPath(resource string) bool {
	return resource == "drive/Home" ||
		resource == "drive/Data" ||
		strings.HasPrefix(resource, "drive/Home/") ||
		strings.HasPrefix(resource, "drive/Data/")
}

// validateDrivePath requires a non-empty drive/Home or drive/Data destination
// (used by file remove).
func validateDrivePath(raw string) error {
	_, err := normalizeDownloadPath(raw, false)
	return err
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
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	ErrorCode string          `json:"error_code"`
	Data      json.RawMessage `json:"data"`
	List      json.RawMessage `json:"list"`
	Total     *int64          `json:"total"`
	HasMore   *bool           `json:"has_more"`
	// Batch lifecycle responses carry succeeded/failed at the top level
	// (alongside code), not under data — see BatchResult.
	Succeeded json.RawMessage `json:"succeeded"`
	Failed    json.RawMessage `json:"failed"`
}

type errorBody struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code"`
}

func parseErrorBody(status int, raw string) (code int, message, errorCode string) {
	code = status
	message = strings.TrimSpace(raw)
	var eb errorBody
	if err := json.Unmarshal([]byte(raw), &eb); err == nil {
		if eb.Code != 0 {
			code = eb.Code
		}
		if strings.TrimSpace(eb.Message) != "" {
			message = strings.TrimSpace(eb.Message)
		}
		errorCode = strings.TrimSpace(eb.ErrorCode)
	}
	return code, message, errorCode
}

func doGet(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutate(ctx, d, "GET", path, nil, out)
}

func doMutate(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env dsEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		var httpErr *whoami.HTTPError
		if errors.As(err, &httpErr) {
			status, msg, errCode := parseErrorBody(httpErr.Status, string(httpErr.Body))
			if recovery := taskErrorRecovery(method, path, status, msg, errCode, body); recovery != "" {
				return fmt.Errorf("%w%s", err, recovery)
			}
		}
		return err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		recovery := taskErrorRecovery(method, path, env.Code, msg, env.ErrorCode, body)
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

// cookieRequiredCode is download-server's "this URL needs cookies".
const cookieRequiredCode = 501

// cookieErrorCodes are the error codes a stored login cookie can fix.
// Besides 501, they are the yt-dlp daemon's private_resource /
// authorization_failed / bot_detected, which clear once the caller's
// session cookies are forwarded.
var cookieErrorCodes = map[int]bool{cookieRequiredCode: true, 507: true, 511: true, 512: true}

// cookieErrorCategories are the symbolic twins of cookieErrorCodes; the
// daemon may send a category without a recognised numeric code.
var cookieErrorCategories = map[string]bool{
	"cookie_required":      true,
	"private_resource":     true,
	"authorization_failed": true,
	"bot_detected":         true,
}

func isCookieRecoverable(errorCode int, errorCategory string) bool {
	if cookieErrorCodes[errorCode] {
		return true
	}
	return cookieErrorCategories[strings.ToLower(strings.TrimSpace(errorCategory))]
}

// cookieHostFromURL reduces a URL to the bare host the cookie store and
// download-server's lookup walk terminate on. Empty when the URL carries
// no host (magnet links, bare paths), in which case the caller falls
// back to a placeholder.
//
// Matches download-server GetPrimaryDomain (rightmost two labels) plus
// the SPA's bare-host convention: lowercased, leading "www." stripped.
// Deliberately not a Public Suffix List eTLD+1 — same naive heuristic
// download-server uses as its climb terminator.
func cookieHostFromURL(rawURL string) string {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return host
}

// cookieRequiredHint is the copy-pasteable next step for a failure a
// login cookie fixes. Cookies live under `settings integration cookie`,
// a different command family, so the hint has to spell the whole
// command out rather than say "upload a cookie first".
func cookieRequiredHint(rawURL string) string {
	domain := cookieHostFromURL(rawURL)
	if domain == "" {
		domain = "<domain>"
	}
	return "this URL needs your login cookies; import them with " +
		"`olares-cli settings integration cookie import --domain " + domain +
		" --file cookies.txt`"
}

func taskErrorRecovery(method, path string, status int, message, errorCode string, body interface{}) string {
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

	createURL := ""
	if method == "POST" && path == "/api/download" {
		switch req := body.(type) {
		case NewDownloadReq:
			createURL = req.URL
		case *NewDownloadReq:
			if req != nil {
				createURL = req.URL
			}
		}
	}

	// Prefer stable machine-readable codes from download-server; message
	// matching remains as a fallback for older servers / unknown codes.
	switch errorCode {
	case "task_in_mover_phase":
		if taskID != "" {
			return fmt.Sprintf(
				"; wait for the move to finish, then retry; inspect progress with `olares-cli knowledge download info %s`",
				taskID,
			)
		}
	case "illegal_pause_status":
		if taskID != "" {
			return fmt.Sprintf(
				"; pause only works while downloading or waiting; inspect status with `olares-cli knowledge download info %s`",
				taskID,
			)
		}
	case "dependency_unavailable":
		if shouldHintYTDLPUnavailable(message) || message == "" {
			return "; " + ytdlpUnavailableHint()
		}
		return "; dependency unavailable; retry later or check provider status"
	case "task_not_found":
		return "; refresh task IDs with `olares-cli knowledge download list`"
	case "invalid_path", "invalid_params":
		if method == "POST" && path == "/api/download" {
			return "; check --path/--url/--extra (duplicate URL creates a new task; use `olares-cli knowledge download list` to inspect existing ones)"
		}
	}

	switch {
	// 501 is the download-server contract for "this URL needs cookies".
	case status == cookieRequiredCode:
		return "; " + cookieRequiredHint(createURL)
	case method == "POST" && path == "/api/download" && shouldHintYTDLPUnavailable(message):
		return "; " + ytdlpUnavailableHint()
	case method == "POST" && path == "/api/download" && status == 400:
		return "; check --path/--url/--extra (duplicate URL creates a new task; use `olares-cli knowledge download list` to inspect existing ones)"
	case strings.Contains(path, "/api/url/inspect") && shouldHintInspectTimeout(message):
		return "; channel/RSS probes can exceed the server inspect timeout; retry later or create the task directly if the URL is known good"
	case taskID != "" && status == 409:
		// resume / cancel / remove during yt-dlp mover phase
		return fmt.Sprintf(
			"; wait for the move to finish, then retry; inspect progress with `olares-cli knowledge download info %s`",
			taskID,
		)
	case taskID != "" && strings.Contains(path, "/pause/") && status == 400:
		return fmt.Sprintf(
			"; pause only works while downloading or waiting; inspect status with `olares-cli knowledge download info %s`",
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

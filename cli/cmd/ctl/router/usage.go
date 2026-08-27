package router

import (
	"context"
	"crypto/md5" //nolint:gosec // the appid is defined as an MD5 prefix, not chosen here
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router usage …` — what has been called, and what it cost.
//
// GET /console/api/spend-logs
// GET /console/api/spend-logs/summary
// GET /console/api/spend-logs/export.csv
//
// Router writes one row per model call, whoever made it. An admin sees the
// workspace; anyone else sees their own calls and cannot ask for another
// person's — the filters that would say "somebody else" are refused rather than
// ignored, so a narrower answer is never mistaken for a complete one.
//
// Cost is Router's own arithmetic, from the prices on the model row. A model
// priced wrongly, or not at all, produces a total that is wrong in exactly that
// way: this is an accounting of what Router believes, not an invoice.

// spendLog mirrors one row of Router's spend table, whole.
//
// Whole rather than the columns the table below prints, because -o json prints
// this struct: a column left out here is a column that vanishes from the JSON,
// which is the one output somebody scripts against.
//
// The quantity columns are pointers and their absence carries meaning. A call
// is priced by whichever unit family its mode belongs to — tokens, seconds,
// images, videos, queries — and nil means no number arrived, which is not a
// measured zero. A VAD pass over silence stores 0.0 seconds and costs nothing;
// an engine that reported no duration stores nil and also costs nothing, and
// only this difference says which.
type spendLog struct {
	ID              int64   `json:"id"`
	AttemptID       *string `json:"attempt_id,omitempty"`
	RequestID       *string `json:"request_id,omitempty"`
	APIKeyID        *string `json:"api_key_id,omitempty"`
	UserID          *string `json:"user_id,omitempty"`
	CallerAppID     *string `json:"caller_app_id,omitempty"`
	ProviderID      *string `json:"provider_id,omitempty"`
	ProviderModelID *string `json:"provider_model_id,omitempty"`
	ModelName       string  `json:"model_name"`
	Mode            string  `json:"mode"`
	// Op is the operation within the mode: which media operation, which
	// audio route. A mode alone does not say what was asked for.
	Op            *string `json:"op,omitempty"`
	PromptTokens  int64   `json:"prompt_tokens"`
	CompletionTok int64   `json:"completion_tokens"`
	TotalTokens   int64   `json:"total_tokens"`
	CacheCreation *int64  `json:"cache_creation_input_tokens,omitempty"`
	CacheRead     *int64  `json:"cache_read_input_tokens,omitempty"`
	ReasoningTok  *int64  `json:"reasoning_tokens,omitempty"`
	CostUSD       float64 `json:"cost_usd"`
	Status        string  `json:"status"`
	HTTPStatus    int     `json:"http_status"`
	ErrorCode     *string `json:"error_code,omitempty"`
	LatencyMS     int64   `json:"latency_ms"`
	// TTFTMS is how long the caller waited for the first token, and exists
	// only for a stream. QueueMS is the part of the latency the engine did
	// not spend working, and only a local engine reports the timings it is
	// derived from.
	TTFTMS   *int64 `json:"ttft_ms,omitempty"`
	QueueMS  *int64 `json:"queue_ms,omitempty"`
	Streamed bool   `json:"streamed"`

	AudioInputSeconds  *float64 `json:"audio_input_seconds,omitempty"`
	AudioOutputSeconds *float64 `json:"audio_output_seconds,omitempty"`
	Images             *int64   `json:"images,omitempty"`
	ImageSize          *string  `json:"image_size,omitempty"`
	ImageQuality       *string  `json:"image_quality,omitempty"`
	Videos             *int64   `json:"videos,omitempty"`
	VideoSeconds       *float64 `json:"video_seconds,omitempty"`
	Queries            *int64   `json:"queries,omitempty"`

	// SessionID groups the calls one task made. It is the caller's own id or
	// nothing: Router cannot invent a task boundary.
	SessionID *string   `json:"session_id,omitempty"`
	IP        *string   `json:"ip,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// spendSummaryRow is one bucket. A bucket that mixes modes carries several
// quantities at once, which is not a contradiction: cost is the sum across unit
// families, so the quantities behind it are too.
type spendSummaryRow struct {
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	IconURL   string   `json:"icon_url,omitempty"`
	Owners    []string `json:"owners,omitempty"`
	Installed bool     `json:"installed,omitempty"`
	// Shared is what makes Owners readable: a shared application is one
	// deployment serving the cluster, so its owners are whoever installed it
	// rather than everyone who may call it.
	Shared             bool    `json:"shared,omitempty"`
	Requests           int64   `json:"requests"`
	CostUSD            float64 `json:"cost_usd"`
	TotalTokens        int64   `json:"total_tokens"`
	PromptTokens       int64   `json:"prompt_tokens"`
	CompletionTok      int64   `json:"completion_tokens"`
	AudioInputSeconds  float64 `json:"audio_input_seconds"`
	AudioOutputSeconds float64 `json:"audio_output_seconds"`
	Images             int64   `json:"images"`
	Videos             int64   `json:"videos"`
	VideoSeconds       float64 `json:"video_seconds"`
	Queries            int64   `json:"queries"`
}

// spendTotals are the same figures however the calls are grouped, so both
// summary shapes embed one copy.
type spendTotals struct {
	TotalRequests        int64   `json:"total_requests"`
	TotalSuccessRequests int64   `json:"total_success_requests"`
	TotalCostUSD         float64 `json:"total_cost_usd"`
	TotalTokens          int64   `json:"total_tokens"`
	// The counters for everything not priced by the token. Tokens alone
	// report a window of image and audio work as having done nothing.
	TotalAudioSeconds float64 `json:"total_audio_seconds"`
	TotalImages       int64   `json:"total_images"`
	TotalVideos       int64   `json:"total_videos"`
	TotalVideoSeconds float64 `json:"total_video_seconds"`
	TotalQueries      int64   `json:"total_queries"`
	AvgTPS            float64 `json:"avg_tps"`
}

type spendSummary struct {
	Dim       string            `json:"dim"`
	Items     []spendSummaryRow `json:"items"`
	Truncated bool              `json:"truncated"`
	spendTotals
}

// spendMultiSummary is what several groupings at once answer with: the buckets
// filed under the dimension they belong to, and one copy of the totals, which
// are the same calls counted the same way however they are grouped.
type spendMultiSummary struct {
	Dims map[string]struct {
		Items     []spendSummaryRow `json:"items"`
		Truncated bool              `json:"truncated"`
	} `json:"dims"`
	spendTotals
}

// spendFilter is the one filter every usage route shares.
type spendFilter struct {
	UserRef     string
	CallerRef   string
	KeyRef      string
	ProviderRef string
	ModelRef    string
	Status      string
	Mode        string
	SessionID   string
	Since       string
	Until       string
	Tag         string
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

// The vocabularies Router validates against, copied because there is nothing
// to read them from. A value outside either is refused here rather than sent:
// Router answers a 400, and one round trip earlier the message can name the
// whole set.
var (
	spendStatuses = []string{"success", "failed", "canceled", "in_progress"}
	spendModes    = []string{
		"chat", "responses", "embedding", "rerank",
		"image_generation", "video_generation", "music_generation", "model3d_generation",
		"translate", "ocr", "search", "scrape", "audio", "passthrough",
	}
	spendSortable = []string{
		"created_at", "cost_usd", "total_tokens", "latency_ms", "ttft_ms", "model_name",
	}
)

// statusInProgress is the one non-terminal status, and the reason cost needs a
// third rendering. A running call has been admitted and has not been priced, so
// its zero is not a price.
const statusInProgress = "in_progress"

var spendDims = []string{"model", "provider", "user", "caller_app", "day", "hour"}

// dimHour is the one grouping that cannot share a request with another. Router
// refuses the combination rather than answering it, and the reason is its own:
// an hourly series is unbounded where the others are a bounded set of names, so
// batching it would make the response size depend on the window.
const dimHour = "hour"

func NewUsageCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "what has been called, and what it cost",
		Long: `Report the model calls Router has served.

An admin sees the whole workspace. Anyone else sees their own calls, and asking
for somebody else's is refused rather than quietly narrowed, so a small answer is
never mistaken for the whole picture.

Cost is Router's arithmetic over the prices on each model row, not a bill from
anybody. A model with no price contributes nothing to the total, which is worth
knowing before treating a number here as spend.

Subcommands:
  summary    totals grouped by model, provider, person, app, day or hour
  list       individual calls, newest first
  export     the same rows as CSV
  retention  how long the individual calls are kept

The first three take the same filters, so a total and the calls behind it are one
flag apart. A total outlives the calls it was made of: totals are kept per day
forever, and the per-call rows are deleted on the window "retention" reports.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newUsageSummaryCommand(f))
	cmd.AddCommand(newUsageListCommand(f))
	cmd.AddCommand(newUsageExportCommand(f))
	cmd.AddCommand(newUsageRetentionCommand(f))
	return cmd
}

func addSpendFilterFlags(cmd *cobra.Command, fl *spendFilter) {
	cmd.Flags().StringVar(&fl.UserRef, "user", "", "only this user's calls, by name or id (admin only)")
	cmd.Flags().StringVar(&fl.CallerRef, "caller-app", "", "only this application's calls, by title, app name or appid (admin only)")
	// --caller was this flag's name, and `quota` spelled the same concept
	// --caller-app with help text identical word for word. One name now; the
	// old one keeps working unannounced rather than breaking a script.
	cmd.Flags().StringVar(&fl.CallerRef, "caller", "", "")
	_ = cmd.Flags().MarkHidden("caller")
	cmd.Flags().StringVar(&fl.KeyRef, "key", "", "only calls made with this key, by name, prefix or id")
	cmd.Flags().StringVar(&fl.ProviderRef, "provider", "", "only calls to this provider, by name or id")
	cmd.Flags().StringVar(&fl.ModelRef, "model", "", "only calls to this model, as <provider>/<model>")
	cmd.Flags().StringVar(&fl.Status, "status", "", "success, failed, canceled or in_progress (still running)")
	cmd.Flags().StringVar(&fl.Mode, "mode", "", "only calls in this mode: "+strings.Join(spendModes, ", "))
	cmd.Flags().StringVar(&fl.SessionID, "session", "", "only the calls one task made, by the session id it sent")
	cmd.Flags().StringVar(&fl.Since, "since", "", "calls at or after this time, or a span like 24h or 7d")
	cmd.Flags().StringVar(&fl.Until, "until", "", "calls before this time")
	cmd.Flags().StringVar(&fl.Tag, "tag", "", "only calls carrying this tag")
}

// resolveSpendQuery turns the human-facing filter into query parameters.
//
// Names are resolved here rather than passed through, because every one of these
// routes filters by id: a provider name sent as-is comes back as an empty page,
// which reads as "nothing happened" instead of "that is not how you ask".
func resolveSpendQuery(ctx context.Context, pc *preparedClient, fl spendFilter) (url.Values, error) {
	q := url.Values{}
	if s := strings.TrimSpace(fl.UserRef); s != "" {
		id, err := resolveUserID(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		q.Set("user_id", id)
	}
	if s := strings.TrimSpace(fl.CallerRef); s != "" {
		id, err := resolveCallerAppID(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		q.Set("caller_app_id", id)
	}
	if s := strings.TrimSpace(fl.KeyRef); s != "" {
		found, err := resolveKey(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		q.Set("api_key_id", found.ID)
	}
	if s := strings.TrimSpace(fl.ProviderRef); s != "" {
		found, err := resolveProvider(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		q.Set("provider_id", found.ID)
	}
	if s := strings.TrimSpace(fl.ModelRef); s != "" {
		row, err := resolveModel(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		q.Set("provider_model_id", row.ProviderModelID)
	}
	if s := strings.ToLower(strings.TrimSpace(fl.Status)); s != "" {
		if !containsString(spendStatuses, s) {
			return nil, fmt.Errorf("--status must be one of %s, not %q",
				strings.Join(spendStatuses, ", "), fl.Status)
		}
		q.Set("status", s)
	}
	if s := strings.ToLower(strings.TrimSpace(fl.Mode)); s != "" {
		if !containsString(spendModes, s) {
			return nil, fmt.Errorf("--mode must be one of %s, not %q",
				strings.Join(spendModes, ", "), fl.Mode)
		}
		q.Set("mode", s)
	}
	if s := strings.TrimSpace(fl.SessionID); s != "" {
		q.Set("session_id", s)
	}
	if s := strings.ToLower(strings.TrimSpace(fl.SortBy)); s != "" {
		if !containsString(spendSortable, s) {
			return nil, fmt.Errorf("--sort-by must be one of %s, not %q",
				strings.Join(spendSortable, ", "), fl.SortBy)
		}
		q.Set("sort_by", s)
	}
	if s := strings.ToLower(strings.TrimSpace(fl.SortOrder)); s != "" {
		if s != "asc" && s != "desc" {
			return nil, fmt.Errorf("--sort-order must be asc or desc, not %q", fl.SortOrder)
		}
		q.Set("sort_order", s)
	}
	if s := strings.TrimSpace(fl.Since); s != "" {
		when, err := parseSinceOrInstant(s)
		if err != nil {
			return nil, fmt.Errorf("--since %w", err)
		}
		q.Set("since", when.UTC().Format(time.RFC3339))
	}
	if s := strings.TrimSpace(fl.Until); s != "" {
		when, err := parseSinceOrInstant(s)
		if err != nil {
			return nil, fmt.Errorf("--until %w", err)
		}
		q.Set("until", when.UTC().Format(time.RFC3339))
	}
	if s := strings.TrimSpace(fl.Tag); s != "" {
		q.Set("tag", s)
	}
	if fl.Limit > 0 {
		q.Set("limit", strconv.Itoa(fl.Limit))
	}
	if fl.Offset > 0 {
		q.Set("offset", strconv.Itoa(fl.Offset))
	}
	return q, nil
}

// callerAppBuckets is the appid-to-name index, and the only one on the wire.
//
// An application does not have a row in Router. The platform vouches for it at
// the edge and the request arrives carrying an appid, so there is nothing to
// register and nothing to archive — which is why there is no application list to
// read and this asks the spend summary instead.
//
// include_idle is what makes it an index rather than a report: it widens the
// dimension into every application the directory says is installed, so an app
// that has never called still has a name here. The other direction is covered
// too, and matters more — an uninstalled application keeps its spend, and its
// bucket then carries the appid as its own label because no name for it exists
// anywhere.
//
// Admin only, like every caller_app read. Nothing here is worth a fallback: a
// non-admin cannot filter by application at all.
func callerAppBuckets(ctx context.Context, pc *preparedClient) ([]spendSummaryRow, error) {
	q := url.Values{}
	q.Set("dim", "caller_app")
	q.Set("include_idle", "true")
	return collection[spendSummaryRow](ctx, pc, withQuery(epSpendSummary, q))
}

// resolveCallerAppID turns what somebody typed into the appid the filter takes.
//
// Three forms, because an appid is not something anybody reads. The title is
// what the desktop shows. The application name is what a manifest and a log
// carry, and it maps to an appid by a hash the platform defines — computed here
// and then *matched against a bucket that exists*, never sent on its own, so a
// typo produces a refusal rather than a filter that silently matches nothing.
// And the appid itself is accepted, since it is what a spend row shows.
func resolveCallerAppID(ctx context.Context, pc *preparedClient, ref string) (string, error) {
	ref, err := requireRef(ref, "an application name, title or appid")
	if err != nil {
		return "", err
	}
	rows, err := callerAppBuckets(ctx, pc)
	if err != nil {
		return "", err
	}
	hashed := appIDFromName(ref)
	known := make([]string, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		if strings.EqualFold(r.Key, ref) || r.Key == hashed || strings.EqualFold(r.Label, ref) {
			return r.Key, nil
		}
		known = append(known, nonEmpty(r.Label))
	}
	sort.Strings(known)
	return "", missing{
		noun:  "application",
		ref:   ref,
		known: dedupeStrings(known),
		have:  "installed or having called are",
		none:  "no application is installed and none has called",
		note: "An application is named by its title, its Olares application name, or the appid a " +
			"spend row shows.",
	}.err()
}

// appIDFromName is the platform's own derivation: the first eight hex digits of
// the MD5 of the application name, hashed exactly as written (ADR-23). A system
// application carries its name instead, which the exact-match branch covers.
//
// MD5 is not a security decision here; it is the identifier's definition, and
// computing anything else would name a different application.
func appIDFromName(name string) string {
	sum := md5.Sum([]byte(strings.TrimSpace(name))) //nolint:gosec // the platform defines the appid this way
	return hex.EncodeToString(sum[:])[:8]
}

// parseSinceOrInstant accepts an absolute time and also a span, because "the
// last day" is how anybody actually asks and computing the instant by hand is
// the sort of arithmetic a tool should do.
func parseSinceOrInstant(s string) (time.Time, error) {
	if d, err := parseTTL(s); err == nil {
		return time.Now().Add(-d), nil
	}
	t, err := parseInstant(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is neither a time nor a span; use 7d, 24h, 2026-08-01 or an RFC3339 instant", s)
	}
	return t, nil
}

func newUsageSummaryCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		dim    string
		fl     spendFilter
	)
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "totals grouped by model, provider, person, app, day or hour",
		Long: `Total the calls Router has served, grouped one or several ways.

--by chooses the grouping: model, provider, user, caller_app, day or hour. Name
several, comma-separated, and each grouping is reported in turn from a single
request — the same calls counted the same way, so the totals underneath are one
figure rather than one per table.

The answer carries the top 100 buckets by cost per grouping, so a workspace with
more than that many models is showing you the expensive ones rather than all of
them.

"hour" cannot be combined with anything: an hourly series grows with the window
where the other groupings are a bounded set of names, so Router answers it on
its own.

FAILED is worth reading alongside cost: a failed call still took time and may
still have been charged upstream, and a group that is mostly failures is a
misconfiguration rather than usage.

Examples:
  olares-cli router usage summary --by model --since 7d
  olares-cli router usage summary --by day --since 30d
  olares-cli router usage summary --by model,provider,user --since 7d
  olares-cli router usage summary --by user --status failed
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUsageSummary(c.Context(), f, dim, fl, output)
		},
	}
	cmd.Flags().StringVar(&dim, "by", "model",
		"group by: "+strings.Join(spendDims, ", ")+"; several, comma-separated, are reported in one request")
	addSpendFilterFlags(cmd, &fl)
	addOutputFlag(cmd, &output)
	return cmd
}

func runUsageSummary(ctx context.Context, f *cmdutil.Factory, dim string, fl spendFilter, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	dims, err := parseSpendDims(dim)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q, err := resolveSpendQuery(ctx, pc, fl)
	if err != nil {
		return err
	}
	if len(dims) > 1 {
		q.Set("dims", strings.Join(dims, ","))
		var multi spendMultiSummary
		if err := pc.router.doJSON(ctx, "GET", withQuery(epSpendSummary, q), nil, &multi); err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, multi)
		}
		return renderUsageSummaries(os.Stdout, dims, &multi)
	}
	q.Set("dim", dims[0])
	var sum spendSummary
	if err := pc.router.doJSON(ctx, "GET", withQuery(epSpendSummary, q), nil, &sum); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, sum)
	}
	return renderUsageSummary(os.Stdout, &sum)
}

// parseSpendDims reads --by. Duplicates collapse rather than producing the same
// table twice, and the order asked for is the order reported, since a caller
// listing model before day is describing what they want to read first.
func parseSpendDims(raw string) ([]string, error) {
	var dims []string
	seen := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		if !containsString(spendDims, name) {
			return nil, fmt.Errorf("--by must name groupings from %s, not %q",
				strings.Join(spendDims, ", "), name)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		dims = append(dims, name)
	}
	if len(dims) == 0 {
		return nil, fmt.Errorf("--by needs a grouping: one of %s", strings.Join(spendDims, ", "))
	}
	if len(dims) > 1 && seen[dimHour] {
		return nil, fmt.Errorf("--by %s cannot be combined with another grouping: an hourly series grows "+
			"with the window, so Router answers it on its own. Ask for it separately", dimHour)
	}
	return dims, nil
}

func renderUsageSummary(w io.Writer, sum *spendSummary) error {
	if len(sum.Items) == 0 {
		_, err := fmt.Fprintln(w, "no calls match. Nothing has gone through Router in this window, "+
			"or the filters exclude everything that has.")
		return err
	}
	if err := renderSummaryBuckets(w, sum.Dim, sum.Items); err != nil {
		return err
	}
	if sum.Truncated {
		if _, err := fmt.Fprintf(w, "\nOnly the first %d buckets are shown; this is not the whole set.\n",
			len(sum.Items)); err != nil {
			return err
		}
	}
	return renderSummaryTotals(w, &sum.spendTotals)
}

// renderUsageSummaries prints one table per grouping and one set of totals. The
// totals are stated once because they are one figure: every grouping counts the
// same calls, and repeating the same number under each table would read as
// though the tables were of different things.
func renderUsageSummaries(w io.Writer, dims []string, multi *spendMultiSummary) error {
	if multi.TotalRequests == 0 {
		_, err := fmt.Fprintln(w, "no calls match. Nothing has gone through Router in this window, "+
			"or the filters exclude everything that has.")
		return err
	}
	for i, dim := range dims {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		bucket, ok := multi.Dims[dim]
		if !ok {
			// Asked for and not answered. Saying so beats an empty table,
			// which would read as "no calls grouped this way".
			if _, err := fmt.Fprintf(w, "%s: this Router did not report this grouping.\n",
				strings.ToUpper(dim)); err != nil {
				return err
			}
			continue
		}
		if err := renderSummaryBuckets(w, dim, bucket.Items); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return renderSummaryTotals(w, &multi.spendTotals)
}

// renderSummaryBuckets prints the quantity columns the buckets actually carry.
//
// Fixed token columns would report a window of image and audio work as three
// zeros beside a real cost, which reads as an arithmetic error rather than as a
// different unit family. A bucket mixing modes carries several quantities at
// once, so several columns can appear together.
func renderSummaryBuckets(w io.Writer, dim string, items []spendSummaryRow) error {
	if len(items) == 0 {
		_, err := fmt.Fprintf(w, "%s: nothing in this window.\n", strings.ToUpper(dim))
		return err
	}
	var tokens, images, videos, audio, queries bool
	for i := range items {
		it := &items[i]
		tokens = tokens || it.TotalTokens > 0
		images = images || it.Images > 0
		videos = videos || it.Videos > 0 || it.VideoSeconds > 0
		audio = audio || it.AudioInputSeconds > 0 || it.AudioOutputSeconds > 0
		queries = queries || it.Queries > 0
	}
	// Nothing measurable in any bucket: keep the token columns rather than a
	// table of two columns, so the shape of the report does not depend on
	// whether the calls in it happened to report a quantity.
	if !tokens && !images && !videos && !audio && !queries {
		tokens = true
	}
	headers := []string{strings.ToUpper(dim), "REQUESTS", "COST"}
	if tokens {
		headers = append(headers, "TOKENS", "IN", "OUT")
	}
	if images {
		headers = append(headers, "IMAGES")
	}
	if videos {
		headers = append(headers, "VIDEOS", "VIDEO S")
	}
	if audio {
		headers = append(headers, "AUDIO IN S", "AUDIO OUT S")
	}
	if queries {
		headers = append(headers, "QUERIES")
	}
	t := newTable(w, headers...)
	for i := range items {
		it := &items[i]
		label := it.Label
		if strings.TrimSpace(label) == "" {
			label = it.Key
		}
		cells := []string{nonEmpty(label), strconv.FormatInt(it.Requests, 10), money(it.CostUSD)}
		if tokens {
			cells = append(cells,
				strconv.FormatInt(it.TotalTokens, 10),
				strconv.FormatInt(it.PromptTokens, 10),
				strconv.FormatInt(it.CompletionTok, 10))
		}
		if images {
			cells = append(cells, strconv.FormatInt(it.Images, 10))
		}
		if videos {
			cells = append(cells, strconv.FormatInt(it.Videos, 10), shortSeconds(it.VideoSeconds))
		}
		if audio {
			cells = append(cells, shortSeconds(it.AudioInputSeconds), shortSeconds(it.AudioOutputSeconds))
		}
		if queries {
			cells = append(cells, strconv.FormatInt(it.Queries, 10))
		}
		t.row(cells...)
	}
	return t.flush()
}

func renderSummaryTotals(w io.Writer, tot *spendTotals) error {
	failed := tot.TotalRequests - tot.TotalSuccessRequests
	if _, err := fmt.Fprintf(w, "\n%d requests, %d of them failed, %s",
		tot.TotalRequests, failed, money(tot.TotalCostUSD)); err != nil {
		return err
	}
	// Every quantity family that carried anything, so a report of audio or
	// image work is not summarized by the one number that is zero on it.
	quantities := make([]string, 0, 5)
	if tot.TotalTokens > 0 {
		quantities = append(quantities, fmt.Sprintf("%d tokens", tot.TotalTokens))
	}
	if tot.TotalAudioSeconds > 0 {
		quantities = append(quantities, shortSeconds(tot.TotalAudioSeconds)+"s of audio")
	}
	if tot.TotalImages > 0 {
		quantities = append(quantities,
			fmt.Sprintf("%d %s", tot.TotalImages, plural(int(tot.TotalImages), "image", "images")))
	}
	if tot.TotalVideos > 0 {
		quantities = append(quantities,
			fmt.Sprintf("%d %s", tot.TotalVideos, plural(int(tot.TotalVideos), "video", "videos")))
	}
	if tot.TotalVideoSeconds > 0 {
		quantities = append(quantities, shortSeconds(tot.TotalVideoSeconds)+"s of video")
	}
	if tot.TotalQueries > 0 {
		quantities = append(quantities,
			fmt.Sprintf("%d %s", tot.TotalQueries, plural(int(tot.TotalQueries), "query", "queries")))
	}
	if len(quantities) == 0 {
		quantities = append(quantities, fmt.Sprintf("%d tokens", tot.TotalTokens))
	}
	if _, err := fmt.Fprintf(w, ", %s", strings.Join(quantities, ", ")); err != nil {
		return err
	}
	if tot.AvgTPS > 0 {
		if _, err := fmt.Fprintf(w, ", averaging %.1f tokens/s", tot.AvgTPS); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "."); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "Cost is computed from the prices on each model row; a model with no price adds nothing to it.")
	return err
}

// shortSeconds prints a duration the way it arrived, without padding: a
// measured 0.5 and a measured 12 both read as themselves, where the fixed three
// decimals a subtitle boundary needs would make a billed minute look rounded.
func shortSeconds(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func money(v float64) string {
	switch {
	case v == 0:
		return "$0"
	case v < 0.01:
		return "$" + strconv.FormatFloat(v, 'f', 6, 64)
	default:
		return "$" + strconv.FormatFloat(v, 'f', 2, 64)
	}
}

func newUsageListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		fl     spendFilter
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "individual calls, newest first",
		Long: `List the model calls Router has served.

One row per call. WHO is the key, person or application Router billed it to;
STATUS is the outcome, and a failed call carries the error code Router returned.

USAGE is the quantity the call was priced by, which is not tokens for most of
what Router serves: audio is priced by the second, images by the picture, video
by both, and search and OCR by the query. The column shows whichever quantity
arrived, so a dash means nothing measurable came back.

--status in_progress lists the calls being served right now. Their cost reads
"pending" rather than $0: the row is written when the call starts and priced when
it ends, and a zero there would be read as free.

--sort-by ranks the whole matching set rather than the page you were handed, so
"the most expensive call this week" is one flag rather than an export.

Use this to explain a total rather than to compute one: "usage summary" is what
adds up.

Examples:
  olares-cli router usage list --since 1h
  olares-cli router usage list --status failed --limit 20
  olares-cli router usage list --status in_progress
  olares-cli router usage list --mode image_generation --since 7d
  olares-cli router usage list --sort-by cost_usd --limit 10
  olares-cli router usage list --session task-4711
  olares-cli router usage list --key ci -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUsageList(c.Context(), f, fl, output)
		},
	}
	addSpendFilterFlags(cmd, &fl)
	// Sorting is only offered here. The summary orders its buckets by cost and
	// the export streams every match, so on those two the flag would be a
	// setting Router ignores.
	cmd.Flags().StringVar(&fl.SortBy, "sort-by", "",
		"rank the whole matching set by: "+strings.Join(spendSortable, ", ")+" (default created_at)")
	cmd.Flags().StringVar(&fl.SortOrder, "sort-order", "", "asc or desc (default desc)")
	cmd.Flags().IntVar(&fl.Limit, "limit", 50, "how many calls to return (1-1000)")
	cmd.Flags().IntVar(&fl.Offset, "offset", 0, "how many calls to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

func runUsageList(ctx context.Context, f *cmdutil.Factory, fl spendFilter, outputRaw string) error {
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
	q, err := resolveSpendQuery(ctx, pc, fl)
	if err != nil {
		return err
	}
	var env page[spendLog]
	if err := pc.router.doJSON(ctx, "GET", withQuery(epSpendLogs, q), nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderUsageList(ctx, pc, os.Stdout, env.Items, env.Total, env.Offset)
}

func renderUsageList(ctx context.Context, pc *preparedClient, w io.Writer, items []spendLog, total, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no calls match.")
		return err
	}
	who := spendActorLabels(ctx, pc, items)
	t := newTable(w, "WHEN", "MODEL", "MODE", "WHO", "STATUS", "USAGE", "COST", "LATENCY")
	for i := range items {
		it := &items[i]
		status := nonEmpty(it.Status)
		if it.ErrorCode != nil && *it.ErrorCode != "" {
			status = it.Status + ": " + *it.ErrorCode
		}
		t.row(
			it.CreatedAt.Local().Format("2006-01-02 15:04:05"),
			clip(nonEmpty(it.ModelName), 28), nonEmpty(spendOp(it)),
			clip(spendActor(it, who), 24), clip(status, 30),
			spendQuantity(it), spendCost(it),
			fmt.Sprintf("%dms", it.LatencyMS))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if err := spendZeroNotes(w, items); err != nil {
		return err
	}
	return pageFooter(w, len(items), total, offset)
}

// spendOp is the MODE column: the mode, and the operation within it when Router
// recorded one. "image_generation" alone does not say whether a picture was made
// or edited, and the two are priced the same but are not the same call.
func spendOp(it *spendLog) string {
	if it.Op != nil && strings.TrimSpace(*it.Op) != "" && *it.Op != it.Mode {
		return it.Mode + "/" + *it.Op
	}
	return it.Mode
}

// spendQuantity is what the call was priced by, read off whichever columns
// carried a number rather than switched on the mode.
//
// Reading the columns is what keeps this honest as Router grows modes: a row
// states its own unit family by which quantity it filled in, so a mode this
// build has never heard of still reports the right figure. Tokens are the
// fallback rather than the first choice, because they are the one quantity that
// is present and zero on calls priced by something else.
func spendQuantity(it *spendLog) string {
	var parts []string
	if it.Images != nil {
		parts = append(parts, fmt.Sprintf("%d img", *it.Images))
	}
	if it.Videos != nil {
		parts = append(parts, fmt.Sprintf("%d vid", *it.Videos))
	}
	if it.VideoSeconds != nil {
		parts = append(parts, shortSeconds(*it.VideoSeconds)+"s")
	}
	if it.AudioInputSeconds != nil {
		parts = append(parts, shortSeconds(*it.AudioInputSeconds)+"s in")
	}
	if it.AudioOutputSeconds != nil {
		parts = append(parts, shortSeconds(*it.AudioOutputSeconds)+"s out")
	}
	if it.Queries != nil {
		parts = append(parts, fmt.Sprintf("%d q", *it.Queries))
	}
	if len(parts) == 0 && it.TotalTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d tok", it.TotalTokens))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// spendCost keeps a running call's cost out of the money column. Router writes
// the row when it admits the call and prices it when the call ends, so the zero
// in between is an absence rather than an amount.
func spendCost(it *spendLog) string {
	if it.Status == statusInProgress {
		return "pending"
	}
	return money(it.CostUSD)
}

// spendZeroNotes explains the zeros on the page, because a zero cost has four
// readings and the column shows one glyph for all of them: still running, priced
// at nothing, measured with no rate to apply, or moved audio nobody measured.
func spendZeroNotes(w io.Writer, items []spendLog) error {
	var running, unpriced, unmetered int
	for i := range items {
		it := &items[i]
		if it.Status == statusInProgress {
			running++
		}
		for _, tag := range it.Tags {
			switch tag {
			case "unpriced":
				unpriced++
			case "audio_unmetered":
				unmetered++
			}
		}
	}
	notes := make([]string, 0, 3)
	if running > 0 {
		notes = append(notes, fmt.Sprintf(
			"%d %s still running. Cost is settled when the call ends, so it is pending rather than zero.",
			running, plural(running, "call is", "calls are")))
	}
	if unpriced > 0 {
		notes = append(notes, fmt.Sprintf(
			"%d %s tagged unpriced: the quantity was measured and the model row has no rate for it, "+
				"so the $0 is a missing price rather than free traffic. Music and 3D generation are "+
				"there permanently today.", unpriced, plural(unpriced, "call is", "calls are")))
	}
	if unmetered > 0 {
		notes = append(notes, fmt.Sprintf(
			"%d audio %s tagged audio_unmetered: the engine reported no duration, and audio is priced "+
				"by the second, so nothing could be charged.",
			unmetered, plural(unmetered, "call is", "calls are")))
	}
	for _, note := range notes {
		if _, err := fmt.Fprintln(w, note); err != nil {
			return err
		}
	}
	return nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// spendActor names who a call was billed to. A key is the most specific answer,
// then the application, then the person — in that order, because a call made
// with a key is best explained by the key even though the person owns it.
func spendActor(it *spendLog, labels map[string]string) string {
	if it.APIKeyID != nil {
		if l := labels["key:"+*it.APIKeyID]; l != "" {
			return l
		}
	}
	if it.CallerAppID != nil {
		if l := labels["app:"+*it.CallerAppID]; l != "" {
			return l
		}
	}
	if it.UserID != nil {
		if l := labels["user:"+*it.UserID]; l != "" {
			return l
		}
		return *it.UserID
	}
	return "-"
}

func spendActorLabels(ctx context.Context, pc *preparedClient, items []spendLog) map[string]string {
	out := map[string]string{}
	var wantKey, wantApp, wantUser bool
	for i := range items {
		wantKey = wantKey || items[i].APIKeyID != nil
		wantApp = wantApp || items[i].CallerAppID != nil
		wantUser = wantUser || items[i].UserID != nil
	}
	if wantKey {
		if keys, err := listKeys(ctx, pc); err == nil {
			for i := range keys {
				out["key:"+keys[i].ID] = keys[i].Name
			}
		}
	}
	if wantApp {
		// A best effort, and it fails for a reader who is not an admin — who
		// then sees the appid itself, which is what the row carries. Their own
		// calls are not an application's, so this is the rare case.
		if rows, err := callerAppBuckets(ctx, pc); err == nil {
			for i := range rows {
				out["app:"+rows[i].Key] = nonEmpty(rows[i].Label)
			}
		}
	}
	if wantUser {
		for id, name := range userLabels(ctx, pc) {
			out["user:"+id] = name
		}
	}
	return out
}

func newUsageExportCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		out string
		fl  spendFilter
	)
	cmd := &cobra.Command{
		Use:   "export",
		Short: "the matching calls as CSV",
		Long: `Write the matching calls to a CSV file.

Every row that matches is written, not just the first page, which is why this
exists alongside "usage list": there is no --limit to raise and no pagination to
walk.

--out writes to a file; without it the CSV goes to stdout, for piping.

Examples:
  olares-cli router usage export --since 30d --out usage.csv
  olares-cli router usage export --status failed | head
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUsageExport(c.Context(), f, fl, out)
		},
	}
	addSpendFilterFlags(cmd, &fl)
	cmd.Flags().StringVar(&out, "out", "", "write to this file instead of stdout")
	return cmd
}

func runUsageExport(ctx context.Context, f *cmdutil.Factory, fl spendFilter, outPath string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q, err := resolveSpendQuery(ctx, pc, fl)
	if err != nil {
		return err
	}
	resp, err := pc.router.doStream(ctx, withQuery(epSpendExportCSV, q))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	sink := io.Writer(os.Stdout)
	if p := strings.TrimSpace(outPath); p != "" {
		file, cerr := os.Create(p)
		if cerr != nil {
			return fmt.Errorf("create %s: %w", p, cerr)
		}
		defer func() { _ = file.Close() }()
		sink = file
	}
	written, err := io.Copy(sink, resp.Body)
	if err != nil {
		// Router streams the rows and cannot report a mid-stream failure in the
		// body, so a truncated file is the failure. Saying how far it got is the
		// only way to know the file is incomplete.
		return fmt.Errorf("the export stopped after %d bytes and the output is incomplete: %w", written, err)
	}
	if p := strings.TrimSpace(outPath); p != "" {
		_, werr := fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", written, p)
		return werr
	}
	return nil
}

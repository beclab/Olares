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

type spendLog struct {
	ID              int64     `json:"id"`
	RequestID       *string   `json:"request_id,omitempty"`
	APIKeyID        *string   `json:"api_key_id,omitempty"`
	UserID          *string   `json:"user_id,omitempty"`
	CallerAppID     *string   `json:"caller_app_id,omitempty"`
	ProviderID      *string   `json:"provider_id,omitempty"`
	ProviderModelID *string   `json:"provider_model_id,omitempty"`
	ModelName       string    `json:"model_name"`
	Mode            string    `json:"mode"`
	PromptTokens    int64     `json:"prompt_tokens"`
	CompletionTok   int64     `json:"completion_tokens"`
	TotalTokens     int64     `json:"total_tokens"`
	ReasoningTokens *int64    `json:"reasoning_tokens,omitempty"`
	CostUSD         float64   `json:"cost_usd"`
	Status          string    `json:"status"`
	HTTPStatus      int       `json:"http_status"`
	ErrorCode       *string   `json:"error_code,omitempty"`
	LatencyMS       int64     `json:"latency_ms"`
	Tags            []string  `json:"tags,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type spendSummaryRow struct {
	Key           string  `json:"key"`
	Label         string  `json:"label"`
	Installed     bool    `json:"installed,omitempty"`
	Requests      int64   `json:"requests"`
	CostUSD       float64 `json:"cost_usd"`
	TotalTokens   int64   `json:"total_tokens"`
	PromptTokens  int64   `json:"prompt_tokens"`
	CompletionTok int64   `json:"completion_tokens"`
}

type spendSummary struct {
	Dim                  string            `json:"dim"`
	Items                []spendSummaryRow `json:"items"`
	TotalRequests        int64             `json:"total_requests"`
	TotalSuccessRequests int64             `json:"total_success_requests"`
	TotalCostUSD         float64           `json:"total_cost_usd"`
	TotalTokens          int64             `json:"total_tokens"`
	AvgTPS               float64           `json:"avg_tps"`
}

// spendMultiSummary is what several groupings at once answer with: the buckets
// filed under the dimension they belong to, and one copy of the totals, which
// are the same calls counted the same way however they are grouped.
type spendMultiSummary struct {
	Dims map[string]struct {
		Items     []spendSummaryRow `json:"items"`
		Truncated bool              `json:"truncated"`
	} `json:"dims"`
	TotalRequests        int64   `json:"total_requests"`
	TotalSuccessRequests int64   `json:"total_success_requests"`
	TotalCostUSD         float64 `json:"total_cost_usd"`
	TotalTokens          int64   `json:"total_tokens"`
	AvgTPS               float64 `json:"avg_tps"`
}

// spendFilter is the one filter every usage route shares.
type spendFilter struct {
	UserRef     string
	CallerRef   string
	KeyRef      string
	ProviderRef string
	ModelRef    string
	Status      string
	Since       string
	Until       string
	Tag         string
	Limit       int
	Offset      int
}

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
	cmd.Flags().StringVar(&fl.CallerRef, "caller", "", "only this application's calls, by title, app name or appid (admin only)")
	cmd.Flags().StringVar(&fl.KeyRef, "key", "", "only calls made with this key, by name, prefix or id")
	cmd.Flags().StringVar(&fl.ProviderRef, "provider", "", "only calls to this provider, by name or id")
	cmd.Flags().StringVar(&fl.ModelRef, "model", "", "only calls to this model, as <provider>/<model>")
	cmd.Flags().StringVar(&fl.Status, "status", "", "success, failed or canceled")
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
		if s != "success" && s != "failed" && s != "canceled" {
			return nil, fmt.Errorf("--status must be success, failed or canceled, not %q", fl.Status)
		}
		q.Set("status", s)
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
	return renderSummaryTotals(w, sum.TotalRequests, sum.TotalSuccessRequests,
		sum.TotalCostUSD, sum.TotalTokens, sum.AvgTPS)
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
	return renderSummaryTotals(w, multi.TotalRequests, multi.TotalSuccessRequests,
		multi.TotalCostUSD, multi.TotalTokens, multi.AvgTPS)
}

func renderSummaryBuckets(w io.Writer, dim string, items []spendSummaryRow) error {
	if len(items) == 0 {
		_, err := fmt.Fprintf(w, "%s: nothing in this window.\n", strings.ToUpper(dim))
		return err
	}
	t := newTable(w, strings.ToUpper(dim), "REQUESTS", "COST", "TOKENS", "IN", "OUT")
	for i := range items {
		it := &items[i]
		label := it.Label
		if strings.TrimSpace(label) == "" {
			label = it.Key
		}
		t.row(nonEmpty(label), strconv.FormatInt(it.Requests, 10), money(it.CostUSD),
			strconv.FormatInt(it.TotalTokens, 10), strconv.FormatInt(it.PromptTokens, 10),
			strconv.FormatInt(it.CompletionTok, 10))
	}
	return t.flush()
}

func renderSummaryTotals(w io.Writer, requests, succeeded int64, cost float64, tokens int64, tps float64) error {
	failed := requests - succeeded
	if _, err := fmt.Fprintf(w, "\n%d requests, %d of them failed, %s, %d tokens",
		requests, failed, money(cost), tokens); err != nil {
		return err
	}
	if tps > 0 {
		if _, err := fmt.Fprintf(w, ", averaging %.1f tokens/s", tps); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "."); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "Cost is computed from the prices on each model row; a model with no price adds nothing to it.")
	return err
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

Use this to explain a total rather than to compute one: "usage summary" is what
adds up.

Examples:
  olares-cli router usage list --since 1h
  olares-cli router usage list --status failed --limit 20
  olares-cli router usage list --key ci -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUsageList(c.Context(), f, fl, output)
		},
	}
	addSpendFilterFlags(cmd, &fl)
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
	t := newTable(w, "WHEN", "MODEL", "MODE", "WHO", "STATUS", "TOKENS", "COST", "LATENCY")
	for i := range items {
		it := &items[i]
		status := nonEmpty(it.Status)
		if it.ErrorCode != nil && *it.ErrorCode != "" {
			status = it.Status + ": " + *it.ErrorCode
		}
		t.row(
			it.CreatedAt.Local().Format("2006-01-02 15:04:05"),
			clip(nonEmpty(it.ModelName), 28), nonEmpty(it.Mode),
			clip(spendActor(it, who), 24), clip(status, 30),
			strconv.FormatInt(it.TotalTokens, 10), money(it.CostUSD),
			fmt.Sprintf("%dms", it.LatencyMS))
	}
	if err := t.flush(); err != nil {
		return err
	}
	return pageFooter(w, len(items), total, offset)
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

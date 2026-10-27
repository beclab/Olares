// Package search implements the `olares-cli search` command tree: the CLI
// counterpart of the Olares Desktop global search dialog.
//
// SPA reference: apps/packages/app/src/api/common/search.ts
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/bflenvelope"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

const (
	appFilesV2     = "files_v2"
	appGoogleDrive = "google_drive"
	appDropbox     = "dropbox"
	appSeafile     = "seafile"
	appKnowledge   = "knowledge"

	// searchSessionAppMinOlaresVersion is the first Olares line on which
	// search3 serves the knowledge partition and federated cloud sources.
	//
	// For google_drive and dropbox this mirrors TermiPass's
	// searchOSVersionLargeThan12_7 gate directly. knowledge has no such
	// gate in the SPA -- it offers the Wise source whenever Wise is
	// installed -- but it lands on the same floor anyway, because only
	// the new Wise feeds search3's knowledge partition and it can only be
	// installed on Olares >= 1.12.7. Supporting the older Wise is
	// explicitly out of scope for the CLI.
	searchSessionAppMinOlaresVersion = "1.12.7"

	searchTypeAggregate = "aggregate"
	searchTypeFileName  = "file_name"

	// initPageSize is the fixed number of hits /search/init returns (it
	// ignores the requested limit). moreMaxLimit is the upper bound the
	// backend enforces on /search/more's limit.
	initPageSize = 20
	moreMaxLimit = 100

	// codeNoMoreResults is the envelope code /search/more returns when the
	// requested offset is past the end of the cached result set. It carries an
	// empty data array and should be treated as "no results", not an error.
	codeNoMoreResults = -3
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

type pagingOptions struct {
	limit  int
	offset int
	output string
}

func parseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

func parseSearchType(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", searchTypeAggregate:
		return searchTypeAggregate, nil
	case searchTypeFileName:
		return searchTypeFileName, nil
	default:
		return "", fmt.Errorf("unsupported --type %q (allowed: aggregate, file_name)", s)
	}
}

func parseKeyword(args []string) (string, error) {
	keyword := strings.TrimSpace(strings.Join(args, " "))
	if keyword == "" {
		return "", fmt.Errorf("a non-empty search keyword is required")
	}
	return keyword, nil
}

func (o *pagingOptions) validate() (Format, error) {
	format, err := parseFormat(o.output)
	if err != nil {
		return "", err
	}
	if o.limit < 0 {
		return "", fmt.Errorf("--limit must not be negative (0 prints every result)")
	}
	if o.offset < 0 {
		return "", fmt.Errorf("--offset must not be negative")
	}
	return format, nil
}

// unlimited reports whether every hit should be printed. Only the federated
// search can honor this: it runs the whole job to completion before anything
// can be shown, so holding results back would hide work already done.
func (o *pagingOptions) unlimited() bool {
	return o.limit <= 0
}

// windowed reports whether only part of the result set will be printed. It is
// what separates "you asked for 20" from "the backend owes us the rest".
func (o *pagingOptions) windowed() bool {
	return o.offset > 0 || o.limit > 0
}

// sessionPaging resolves "print everything" down to a page for the synchronous
// search3 session API. That path pages server-side over a cached result set
// with no total to work from, so it keeps the page size the Desktop dialog
// uses rather than walking the cache to its end.
func (o pagingOptions) sessionPaging() pagingOptions {
	if o.unlimited() {
		o.limit = initPageSize
	}
	return o
}

// registerPagingFlags wires the shared paging flags. defaultLimit is per
// command because only the federated search can print an entire result set:
// commands that fall back to the session API keep a page.
func registerPagingFlags(cmd *cobra.Command, o *pagingOptions, defaultLimit int) {
	usage := "maximum number of results"
	if defaultLimit <= 0 {
		usage += " (0 prints every result)"
	}
	cmd.Flags().IntVarP(&o.limit, "limit", "l", defaultLimit, usage)
	cmd.Flags().IntVar(&o.offset, "offset", 0, "result offset for pagination")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "output format: table, json")
}

func newDoer(ctx context.Context, f *cmdutil.Factory) (*whoami.HTTPClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: search not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID), nil
}

// doEnvelope issues an authenticated request and unwraps the {code,
// message, data} envelope into out. out may be nil for fire-and-forget
// calls (e.g. cancel).
func doEnvelope(ctx context.Context, d *whoami.HTTPClient, method, path string, body, out interface{}) error {
	return doEnvelopeAllowing(ctx, d, method, path, body, out)
}

// doEnvelopeAllowing behaves like doEnvelope but additionally treats the given
// soft codes as success, decoding whatever data they carry. This lets callers
// tolerate non-fatal upstream codes (e.g. /search/more's "no more results").
func doEnvelopeAllowing(ctx context.Context, d *whoami.HTTPClient, method, path string, body, out interface{}, softCodes ...int) error {
	var env bflenvelope.Envelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	return bflenvelope.Data(method, path, env, out, softCodes...)
}

// searchPage is what will be printed, together with the size of the result set
// it came from. total lets the footer distinguish a window the user asked for
// from hits the backend reported but never delivered; it is 0 when the size is
// genuinely unknown -- the legacy session API pages server-side and never says
// how much it cached.
type searchPage struct {
	items    []resultItem
	offset   int
	total    int
	windowed bool
}

// remaining reports how many hits past this window the backend still holds,
// and 0 whenever that cannot be known.
func (p searchPage) remaining() int {
	return remainingResults(p.total, p.offset, len(p.items))
}

func remainingResults(total, offset, shown int) int {
	rest := total - (offset + shown)
	if total <= 0 || rest <= 0 {
		return 0
	}
	return rest
}

// resultItem captures the fields the desktop SPA reads off each result
// row, across both the /init (Drive/Knowledge/Files) and /sync shapes.
type resultItem struct {
	Title       string          `json:"title"`
	ResourceURI string          `json:"resource_uri,omitempty"`
	Path        string          `json:"path,omitempty"`
	RepoName    string          `json:"repo_name,omitempty"`
	Highlight   json.RawMessage `json:"highlight,omitempty"`
	// HighlightField names the field each Highlight entry was produced from,
	// positionally. Both are `string | string[]` in the Desktop's own type
	// (apps/packages/app/src/utils/interface/search.ts), hence raw.
	HighlightField json.RawMessage `json:"highlight_field,omitempty"`
	// Left raw because `meta` is source-specific: a shape this CLI does not
	// model must not fail the whole result set.
	Meta json.RawMessage `json:"meta,omitempty"`

	Raw json.RawMessage `json:"-"`
}

func (it resultItem) location() string {
	if it.ResourceURI != "" {
		return it.ResourceURI
	}
	return it.Path
}

// displayTitle is the name to print for a row. The federated index leaves
// `title` empty on rows that matched on content alone, and the Desktop then
// falls back to the title excerpt; when there is no excerpt either, the file
// name is still sitting in the location, and printing that beats "(untitled)"
// next to a perfectly good path.
func (it resultItem) displayTitle() string {
	if it.Title != "" {
		return it.Title
	}
	if titled := it.highlightFor("title"); titled != "" {
		return titled
	}
	if base := locationBase(it.location()); base != "" {
		return base
	}
	return "(untitled)"
}

// locationBase recovers a file name from a result location, which is a
// slash-separated path on every source that indexes files.
func locationBase(location string) string {
	trimmed := strings.TrimRight(location, "/")
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
		return trimmed[idx+1:]
	}
	return trimmed
}

// snippet is the excerpt printed under a result. When the row says which field
// each excerpt came from, only the content one is worth printing: the title
// excerpt merely repeats the line above it.
func (it resultItem) snippet() string {
	if len(jsonStrings(it.HighlightField)) > 0 {
		if content := it.highlightFor("content"); content != "" {
			return content
		}
		if it.highlightFor("title") != "" {
			return ""
		}
	}
	return joinHighlights(jsonStrings(it.Highlight))
}

// highlightFor returns the excerpt produced for one field. `highlight` is
// positional against `highlight_field`, which is how the Desktop reads it
// (apps/packages/app/src/components/search/SearchItemsComponent.vue).
func (it resultItem) highlightFor(field string) string {
	entries := jsonStrings(it.Highlight)
	for i, name := range jsonStrings(it.HighlightField) {
		if name == field && i < len(entries) {
			return cleanHighlight(entries[i])
		}
	}
	return ""
}

// libraryName returns the Sync library's display name when the hit carries one.
// Federated seafile hits report it in `meta.repo_name`; legacy /api/search/sync
// rows put it at the top level.
func (it resultItem) libraryName() string {
	if it.RepoName != "" {
		return it.RepoName
	}
	var meta struct {
		RepoName string `json:"repo_name"`
	}
	if err := json.Unmarshal(it.Meta, &meta); err != nil {
		return ""
	}
	return meta.RepoName
}

// locationLine renders the result's location, annotated with the Sync library's
// display name when there is one. A seafile location carries only the repo id,
// which is unreadable on its own but is also what makes the location a valid
// files path (`files ls sync/<repo_id>/`) — hence an annotation rather than a
// substitution.
func (it resultItem) locationLine() string {
	location := it.location()
	if location == "" {
		return ""
	}
	if library := it.libraryName(); library != "" {
		return fmt.Sprintf("%s (%s)", location, library)
	}
	return location
}

// paginateRaw applies client-side offset/limit windowing to raw result
// rows. It exists for endpoints that do not paginate server-side (sync), and
// for trimming an over-sized server page down to the requested limit (drive
// init). Callers must pass offset >= 0 and limit > 0 (pagingOptions.validate
// guarantees this).
func paginateRaw(rows []json.RawMessage, offset, limit int) []json.RawMessage {
	if offset >= len(rows) {
		return nil
	}
	end := len(rows)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return rows[offset:end]
}

// needsMorePage reports whether the requested [offset, offset+limit) window
// extends beyond what /search/init already returned and therefore requires a
// /search/more call. A short init page (fewer than initPageSize hits) means
// the cached result set is exhausted, so the window can always be served from
// the init rows.
func needsMorePage(offset, limit, initLen int) bool {
	return offset+limit > initLen && initLen >= initPageSize
}

// clampMoreLimit caps a requested limit to the backend's /search/more maximum.
func clampMoreLimit(limit int) int {
	if limit > moreMaxLimit {
		return moreMaxLimit
	}
	return limit
}

func decodeResultRows(rawRows []json.RawMessage) ([]resultItem, error) {
	items := make([]resultItem, 0, len(rawRows))
	for _, raw := range rawRows {
		var it resultItem
		if err := json.Unmarshal(raw, &it); err != nil {
			return nil, fmt.Errorf("decode search result: %w", err)
		}
		it.Raw = raw
		items = append(items, it)
	}
	return items, nil
}

func printSearchResults(format Format, page searchPage) error {
	switch format {
	case FormatJSON:
		if err := printResultsJSON(os.Stdout, page.items); err != nil {
			return err
		}
		// stdout stays a plain array of hits for whoever is parsing it; the
		// note that it is a partial one goes to stderr.
		return writeTruncationNote(os.Stderr, len(page.items), page.total, page.remaining(), page.windowed)
	default:
		return renderResults(os.Stdout, page)
	}
}

func printResultsJSON(w io.Writer, items []resultItem) error {
	rows := make([]json.RawMessage, 0, len(items))
	for _, it := range items {
		rows = append(rows, it.Raw)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

func renderResults(w io.Writer, page searchPage) error {
	items := page.items
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no results")
		return err
	}
	for i, it := range items {
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, it.displayTitle()); err != nil {
			return err
		}
		if loc := it.locationLine(); loc != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", loc); err != nil {
				return err
			}
		}
		if snippet := it.snippet(); snippet != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", snippet); err != nil {
				return err
			}
		}
	}
	return writeResultCount(w, len(items), page.total, page.remaining(), page.windowed)
}

// writeResultCount renders the trailing count. A short window is only worth
// remarking on when the reader can act on it, and what they can do depends on
// why it is short: their own --offset/--limit, or hits the job counted but
// never handed over.
func writeResultCount(w io.Writer, shown, total, remaining int, windowed bool) error {
	if remaining <= 0 {
		_, err := fmt.Fprintf(w, "\n%d result(s)\n", shown)
		return err
	}
	if windowed {
		_, err := fmt.Fprintf(w, "\n%d of %d result(s); raise --limit or page with --offset for the other %d\n",
			shown, total, remaining)
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d of %d result(s); the search reported %d more but never delivered them\n",
		shown, total, remaining)
	return err
}

// writeTruncationNote is writeResultCount's counterpart for JSON output, where
// the count cannot ride along in the payload without changing its shape.
func writeTruncationNote(w io.Writer, shown, total, remaining int, windowed bool) error {
	if remaining <= 0 {
		return nil
	}
	if windowed {
		_, err := fmt.Fprintf(w, "search: printed %d of %d result(s); raise --limit or page with --offset for the other %d\n",
			shown, total, remaining)
		return err
	}
	_, err := fmt.Fprintf(w, "search: printed %d of %d result(s); the search reported %d more but never delivered them\n",
		shown, total, remaining)
	return err
}

// jsonStrings decodes a `string | string[]` field into a slice. An unmodelled
// shape yields nothing rather than failing the row.
func jsonStrings(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		return many
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}
	}
	return nil
}

func joinHighlights(entries []string) string {
	cleaned := make([]string, 0, len(entries))
	for _, entry := range entries {
		if c := cleanHighlight(entry); c != "" {
			cleaned = append(cleaned, c)
		}
	}
	return strings.Join(cleaned, highlightEllipsis)
}

// highlightElision is the marker search3 leaves where it dropped the text
// between two excerpt windows of the same field. Only a run of exactly this
// width is one: hashes that are really part of a document must survive, and
// the cost of guessing wrong is mangled file content.
const highlightElision = "#######"

const highlightEllipsis = " … "

var highlightTags = strings.NewReplacer("<hi>", "", "</hi>", "")

// cleanHighlight turns one backend excerpt into a printable line. `<hi>` tags
// are the Desktop's bold markers, and an elision reads as the same ellipsis
// this CLI already puts between excerpts.
func cleanHighlight(s string) string {
	cleaned := replaceElisions(highlightTags.Replace(s))
	cleaned = strings.TrimSpace(strings.Join(strings.Fields(cleaned), " "))
	// An excerpt that survives as nothing but elisions says nothing.
	if strings.Trim(cleaned, "… ") == "" {
		return ""
	}
	return cleaned
}

func replaceElisions(s string) string {
	if !strings.Contains(s, highlightElision) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '#' {
			b.WriteByte(s[i])
			i++
			continue
		}
		end := i
		for end < len(s) && s[end] == '#' {
			end++
		}
		if end-i == len(highlightElision) {
			b.WriteString(highlightEllipsis)
		} else {
			b.WriteString(s[i:end])
		}
		i = end
	}
	return b.String()
}

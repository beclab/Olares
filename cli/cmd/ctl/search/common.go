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

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

const (
	appFilesV2 = "files_v2"

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
	if o.limit <= 0 {
		return "", fmt.Errorf("--limit must be a positive integer")
	}
	if o.offset < 0 {
		return "", fmt.Errorf("--offset must not be negative")
	}
	return format, nil
}

func registerPagingFlags(cmd *cobra.Command, o *pagingOptions) {
	cmd.Flags().IntVarP(&o.limit, "limit", "l", 20, "maximum number of results")
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

// bflEnvelope is search3's (and BFL's) shared response shape. code 0/200
// mean success; data carries the typed payload.
type bflEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
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
	var env bflEnvelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	ok := env.Code == 0 || env.Code == 200
	for _, c := range softCodes {
		if env.Code == c {
			ok = true
			break
		}
	}
	if !ok {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("%s %s: upstream returned code %d", method, path, env.Code)
		}
		return fmt.Errorf("%s %s: upstream returned code %d: %s", method, path, env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

// resultItem captures the fields the desktop SPA reads off each result
// row, across both the /init (Drive/Knowledge/Files) and /sync shapes.
type resultItem struct {
	Title       string          `json:"title"`
	ResourceURI string          `json:"resource_uri,omitempty"`
	Path        string          `json:"path,omitempty"`
	RepoName    string          `json:"repo_name,omitempty"`
	Highlight   json.RawMessage `json:"highlight,omitempty"`

	Raw json.RawMessage `json:"-"`
}

func (it resultItem) location() string {
	if it.ResourceURI != "" {
		return it.ResourceURI
	}
	return it.Path
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

func printSearchResults(format Format, items []resultItem) error {
	switch format {
	case FormatJSON:
		return printResultsJSON(os.Stdout, items)
	default:
		return renderResults(os.Stdout, items)
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

func renderResults(w io.Writer, items []resultItem) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no results")
		return err
	}
	for i, it := range items {
		title := it.Title
		if title == "" {
			title = "(untitled)"
		}
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, title); err != nil {
			return err
		}
		if loc := it.location(); loc != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", loc); err != nil {
				return err
			}
		}
		if snippet := highlightSnippet(it.Highlight); snippet != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", snippet); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(w, "\n%d result(s)\n", len(items))
	return err
}

func highlightSnippet(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return stripHighlightTags(single)
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		return stripHighlightTags(strings.Join(many, " … "))
	}
	return ""
}

func stripHighlightTags(s string) string {
	replacer := strings.NewReplacer("<hi>", "", "</hi>", "")
	return strings.TrimSpace(strings.Join(strings.Fields(replacer.Replace(s)), " "))
}

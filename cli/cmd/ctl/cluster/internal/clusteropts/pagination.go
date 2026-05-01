package clusteropts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
)

// PaginationOptions backs the `--limit / --page / --all` flag trio
// shared by every cluster list verb. Values match the KubeSphere
// /kapis/* paginated list contract documented at
// pkg/clusterclient/decode.go::ListResponse — the same `limit` /
// `page` query params the SPA's Pagination type sends (see
// apps/packages/app/src/apps/controlPanelCommon/network/network.ts:181).
//
// Page is 1-indexed (matching the SPA wire shape and KubeSphere
// server convention). Limit is the per-request cap; in --all mode it
// becomes the page size for the drain loop (smaller = more requests,
// larger = fewer requests with bigger responses).
//
// All and Page (when > 1) are mutually exclusive: pass --all to drain
// every page, or --page N to fetch exactly one page. The mutex is
// enforced by Validate, not by cobra, so the conflict surfaces as a
// readable error message rather than a generic flag-parser complaint.
type PaginationOptions struct {
	Limit int
	Page  int
	All   bool
}

// NewPaginationOptions returns a PaginationOptions with the canonical
// defaults — page size 100, page 1, --all off. Use this when
// constructing PaginationOptions outside cobra (e.g. shared between a
// command and a wrapper) so defaults stay in one place.
func NewPaginationOptions() *PaginationOptions {
	return &PaginationOptions{Limit: 100, Page: 1}
}

// AddPaginationFlags wires --limit / --page / --all onto cmd. Help
// strings call out the mutex with --all explicitly so users see it in
// `cmd --help` instead of only at runtime.
//
// Defaults: --limit 100, --page 1, --all false. Same trio is
// registered identically across every list verb so the help text is
// uniform.
func (p *PaginationOptions) AddPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&p.Limit, "limit", 100, "max items per request (KubeSphere page size)")
	cmd.Flags().IntVar(&p.Page, "page", 1, "1-indexed page to fetch (mutually exclusive with --all when N > 1)")
	cmd.Flags().BoolVar(&p.All, "all", false, "fetch every page until exhausted; ignores --page")
}

// Validate enforces the --all/--page mutex and clamps degenerate
// values to sane defaults (so a user who passes --limit 0 still gets
// a working request). MUST be called once from each list verb's RunE
// before the first network round-trip.
//
// Returns a typed error when --all is combined with an explicit
// --page > 1 (page == 1 is the default and silently coexists with
// --all so users can write `--all` alone).
func (p *PaginationOptions) Validate() error {
	if p.All && p.Page > 1 {
		return errors.New("--all and --page are mutually exclusive (omit --page or pass --page 1)")
	}
	if p.Limit < 1 {
		p.Limit = 100
	}
	if p.Page < 1 {
		p.Page = 1
	}
	return nil
}

// AppendQueryForPage writes `limit=<L>` and `page=<N>` into q for a
// single round-trip targeted at page N. Caller-side helper used by
// per-verb buildListPath functions; FetchAllKubeSphere drives the
// page parameter itself in --all mode and just passes the right N to
// the request closure.
//
// The "page > 0" guard skips the page param entirely on N == 0,
// matching the KubeSphere convention that omitting `page` is
// equivalent to page=1 (handy for callers that want a clean URL on
// the default request).
func (p *PaginationOptions) AppendQueryForPage(q url.Values, page int) {
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
}

// FetchAllKubeSphere is the canonical KubeSphere-paginated drain
// loop. When p.All is false, fires one request for p.Page and returns
// the items + reported total. When p.All is true, walks page=1..N
// until any of three stop conditions trips (in declaration order):
//
//  1. resp.TotalItems == 0          → empty list, nothing to do
//  2. len(accumulated) >= TotalItems → full coverage, done
//  3. len(resp.Items) < p.Limit     → server signaled "this was the last page"
//
// `request` builds the URL for a given page; callers own the path
// + querystring shape so per-verb selectors (label, field, kind)
// stay in their own buildListPath helpers. Use AppendQueryForPage to
// stamp limit + page onto the url.Values the request closure builds.
//
// Returns (items, totalItems, err). totalItems is the value reported
// by the LAST successful response — for --all callers this is the
// authoritative total; for single-page callers (`!p.All`) this lets
// the renderer print "(showing N of M total — pass --page <next>)"
// hints when len(items) < total.
//
// We deliberately do NOT impose a hard request-count cap. If a server
// regression causes "every page returns exactly limit items but
// totalItems lies" the loop would spin — but that's a server bug we
// want surfaced, not silently capped to an arbitrary "200 requests"
// limit that would let users believe they got everything when they
// didn't. Real KubeSphere doesn't lie, and a hung CLI is a louder
// signal than truncated data.
func FetchAllKubeSphere[T any](
	ctx context.Context,
	c *clusterclient.Client,
	p *PaginationOptions,
	request func(page int) string,
) (items []T, total int, err error) {
	if !p.All {
		path := request(p.Page)
		resp, err := clusterclient.GetKubeSphereList[T](ctx, c, path)
		if err != nil {
			return nil, 0, err
		}
		return resp.Items, resp.TotalItems, nil
	}

	var all []T
	lastTotal := 0
	for page := 1; ; page++ {
		path := request(page)
		resp, err := clusterclient.GetKubeSphereList[T](ctx, c, path)
		if err != nil {
			return nil, 0, fmt.Errorf("page %d: %w", page, err)
		}
		all = append(all, resp.Items...)
		lastTotal = resp.TotalItems

		if resp.TotalItems == 0 {
			return all, 0, nil
		}
		if len(all) >= resp.TotalItems {
			return all, lastTotal, nil
		}
		if len(resp.Items) < p.Limit {
			return all, lastTotal, nil
		}
	}
}

// PrintPageHint emits a one-line stderr hint when the rendered page
// is a strict subset of the available data, telling the user how to
// see more (`--page <next>` or `--all`). Centralized here so every
// list verb gets identical wording (the previous "(showing N of M
// total — pass --limit X to see more)" hint was copy-pasted across
// 6 files and inconsistent — some used --limit, all suggested the
// wrong knob).
//
// Behavior:
//
//   - --all mode: never prints (the caller already drained everything)
//   - returnedItems >= total or total == 0: never prints (full coverage)
//   - otherwise: "(showing items A-B of TOTAL — pass --page <N+1> or
//     --all to see more)" where A,B are the 1-indexed range covered
//     by the current page
//
// Suppressed when out is nil so JSON / --quiet callers can pass nil
// to opt out cleanly without an extra branch at the call site.
func PrintPageHint(out io.Writer, p *PaginationOptions, returnedItems, total int) {
	if out == nil {
		return
	}
	if p == nil || p.All {
		return
	}
	if total == 0 || returnedItems >= total {
		return
	}
	if p.Limit < 1 {
		return
	}
	pageStart := (p.Page-1)*p.Limit + 1
	pageEnd := pageStart + returnedItems - 1
	if pageEnd < pageStart {
		// returnedItems == 0 on a page past the end: "(no items
		// on page N; total has M items)" is the more useful
		// shape than the negative-range default.
		fmt.Fprintf(out, "(no items on page %d; total has %d items — pass --page 1 or --all)\n", p.Page, total)
		return
	}
	fmt.Fprintf(out, "(showing items %d-%d of %d — pass --page %d or --all to see more)\n",
		pageStart, pageEnd, total, p.Page+1)
}

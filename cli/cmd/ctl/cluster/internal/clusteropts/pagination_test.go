package clusteropts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// TestValidate_Mutex covers the --all / --page mutex (the user-visible
// rule) plus the silent --all + default --page=1 acceptance (so users
// can write `--all` alone without explicit --page 1).
func TestValidate_Mutex(t *testing.T) {
	for _, tc := range []struct {
		name    string
		opts    PaginationOptions
		wantErr bool
	}{
		{"all alone", PaginationOptions{Limit: 100, Page: 1, All: true}, false},
		{"page alone", PaginationOptions{Limit: 100, Page: 5, All: false}, false},
		{"defaults", PaginationOptions{Limit: 100, Page: 1, All: false}, false},
		{"all + explicit page=2", PaginationOptions{Limit: 100, Page: 2, All: true}, true},
		{"all + page=10", PaginationOptions{Limit: 100, Page: 10, All: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestValidate_ClampsDegenerate makes sure "obviously broken inputs"
// (zero / negative values from a typo or scripting bug) get clamped
// to the documented defaults instead of generating an empty URL or
// page=0 request KubeSphere would reject opaquely.
func TestValidate_ClampsDegenerate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		in        PaginationOptions
		wantLimit int
		wantPage  int
	}{
		{"limit zero", PaginationOptions{Limit: 0, Page: 1}, 100, 1},
		{"limit negative", PaginationOptions{Limit: -5, Page: 1}, 100, 1},
		{"page zero", PaginationOptions{Limit: 50, Page: 0}, 50, 1},
		{"page negative", PaginationOptions{Limit: 50, Page: -1}, 50, 1},
		{"both broken", PaginationOptions{Limit: -1, Page: -1}, 100, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.in
			if err := p.Validate(); err != nil {
				t.Fatalf("Validate() err = %v", err)
			}
			if p.Limit != tc.wantLimit || p.Page != tc.wantPage {
				t.Errorf("after Validate, got (limit=%d, page=%d), want (limit=%d, page=%d)",
					p.Limit, p.Page, tc.wantLimit, tc.wantPage)
			}
		})
	}
}

// item is the minimal payload type used by the FetchAllKubeSphere
// tests below — just enough to be JSON-decodable into a real
// ListResponse[item] so the generic plumbing exercises the same
// codepath production callers use.
type item struct {
	Name string `json:"name"`
}

// pagedServer is the canonical KubeSphere-style backend for these
// tests: returns up to limit items per page, advertises totalItems,
// and counts hits per page so tests can assert the drain loop fired
// exactly the right number of requests.
type pagedServer struct {
	*httptest.Server
	hits  atomic.Int32
	pages map[int][]item // page -> items returned
	total int
}

func newPagedServer(t *testing.T, total, limit int) *pagedServer {
	ps := &pagedServer{total: total, pages: map[int][]item{}}
	// Pre-compute exactly which items each page returns. Page is
	// 1-indexed; last page may be short.
	for p := 1; ; p++ {
		start := (p - 1) * limit
		if start >= total {
			break
		}
		end := start + limit
		if end > total {
			end = total
		}
		var items []item
		for i := start; i < end; i++ {
			items = append(items, item{Name: fmt.Sprintf("it-%d", i)})
		}
		ps.pages[p] = items
	}
	ps.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps.hits.Add(1)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		items, ok := ps.pages[page]
		if !ok {
			items = nil
		}
		w.Header().Set("Content-Type", "application/json")
		// Hand-craft the JSON so test stays decoupled from
		// pkg/clusterclient's encoder choice.
		var sb strings.Builder
		sb.WriteString(`{"items":[`)
		for i, it := range items {
			if i > 0 {
				sb.WriteString(",")
			}
			fmt.Fprintf(&sb, `{"name":%q}`, it.Name)
		}
		fmt.Fprintf(&sb, `],"totalItems":%d}`, ps.total)
		_, _ = w.Write([]byte(sb.String()))
	}))
	t.Cleanup(ps.Close)
	return ps
}

// newClient wires a clusterclient.Client targeting the test server.
// We bypass refreshingTransport entirely (the pagination layer
// doesn't care about auth — that's tested in cmdutil/factory_test.go);
// the default http.Client suffices.
func newClient(srv *pagedServer) *clusterclient.Client {
	return clusterclient.NewClient(http.DefaultClient, &credential.ResolvedProfile{
		ControlHubURL: srv.URL,
		OlaresID:      "test@olares.com",
	})
}

// requestFor mirrors the per-verb buildListPath shape: builds the
// /kapis/... path with limit + page query string. Same helper the
// production list verbs will use via PaginationOptions.AppendQueryForPage.
func requestFor(p *PaginationOptions) func(int) string {
	return func(page int) string {
		q := url.Values{}
		p.AppendQueryForPage(q, page)
		return "/kapis/test/items?" + q.Encode()
	}
}

// TestFetchAllKubeSphere_SinglePage covers the !All path: one GET
// per call, page taken from p.Page, no drain loop. The test asserts
// the URL carries page=2 (the request closure must be driven with
// p.Page, not page=1 hard-coded).
func TestFetchAllKubeSphere_SinglePage(t *testing.T) {
	srv := newPagedServer(t, 250, 100)
	c := newClient(srv)
	p := &PaginationOptions{Limit: 100, Page: 2, All: false}

	items, total, err := FetchAllKubeSphere[item](context.Background(), c, p, requestFor(p))
	if err != nil {
		t.Fatalf("FetchAllKubeSphere err = %v", err)
	}
	if srv.hits.Load() != 1 {
		t.Errorf("server hits = %d, want 1 (single-page mode)", srv.hits.Load())
	}
	if total != 250 {
		t.Errorf("total = %d, want 250", total)
	}
	if len(items) != 100 {
		t.Errorf("items = %d, want 100 (page 2 of 250 with limit 100)", len(items))
	}
	if items[0].Name != "it-100" {
		t.Errorf("items[0].Name = %q, want it-100 (page=2 should start at index 100)", items[0].Name)
	}
}

// TestFetchAllKubeSphere_DrainsToTotal: 250 items / limit 100 →
// pages 1,2,3 with the last returning 50 items. The drain loop must
// hit the server exactly 3 times and accumulate all 250.
//
// The third stop condition (len(resp.Items) < limit) is what fires
// here — the second stop (len(all) >= total) would also fire on
// page 3, but the order in FetchAllKubeSphere is total-first so
// "exactly 250 received" wins.
func TestFetchAllKubeSphere_DrainsToTotal(t *testing.T) {
	srv := newPagedServer(t, 250, 100)
	c := newClient(srv)
	p := &PaginationOptions{Limit: 100, Page: 1, All: true}

	items, total, err := FetchAllKubeSphere[item](context.Background(), c, p, requestFor(p))
	if err != nil {
		t.Fatalf("FetchAllKubeSphere err = %v", err)
	}
	if srv.hits.Load() != 3 {
		t.Errorf("server hits = %d, want 3 (250 items / limit 100 = 2 full + 1 short page)", srv.hits.Load())
	}
	if total != 250 {
		t.Errorf("total = %d, want 250", total)
	}
	if len(items) != 250 {
		t.Fatalf("items = %d, want 250", len(items))
	}
	// Spot-check page boundaries to make sure the drain loop didn't
	// re-fetch the same page or skip one.
	if items[0].Name != "it-0" || items[99].Name != "it-99" || items[100].Name != "it-100" || items[249].Name != "it-249" {
		t.Errorf("page boundaries wrong: items[0]=%q items[99]=%q items[100]=%q items[249]=%q",
			items[0].Name, items[99].Name, items[100].Name, items[249].Name)
	}
}

// TestFetchAllKubeSphere_StopsOnExactTotal: 200 items / limit 100 →
// pages 1,2 each returning exactly 100. The "len(all) >= total"
// stop condition (#2) must fire after page 2 even though that page
// was a full page (the third stop would NOT fire — len(items) ==
// limit). Without stop #2 the loop would issue a wasteful 3rd
// request that returns an empty page.
func TestFetchAllKubeSphere_StopsOnExactTotal(t *testing.T) {
	srv := newPagedServer(t, 200, 100)
	c := newClient(srv)
	p := &PaginationOptions{Limit: 100, Page: 1, All: true}

	items, total, err := FetchAllKubeSphere[item](context.Background(), c, p, requestFor(p))
	if err != nil {
		t.Fatalf("FetchAllKubeSphere err = %v", err)
	}
	if srv.hits.Load() != 2 {
		t.Errorf("server hits = %d, want exactly 2 (no wasteful 3rd request when len(all) >= total)", srv.hits.Load())
	}
	if total != 200 {
		t.Errorf("total = %d, want 200", total)
	}
	if len(items) != 200 {
		t.Errorf("items = %d, want 200", len(items))
	}
}

// TestFetchAllKubeSphere_EmptyList: totalItems=0 → 1 request, drain
// stops immediately. Stop condition #1 (TotalItems == 0) prevents
// an infinite-loop pathology when a server returns "no items, total
// 0" but somehow makes p.Limit non-comparable to len(items)==0
// (it would always be < limit, but the explicit zero-check makes
// the intent obvious to readers).
func TestFetchAllKubeSphere_EmptyList(t *testing.T) {
	srv := newPagedServer(t, 0, 100)
	c := newClient(srv)
	p := &PaginationOptions{Limit: 100, Page: 1, All: true}

	items, total, err := FetchAllKubeSphere[item](context.Background(), c, p, requestFor(p))
	if err != nil {
		t.Fatalf("FetchAllKubeSphere err = %v", err)
	}
	if srv.hits.Load() != 1 {
		t.Errorf("server hits = %d, want 1 (empty list = one round-trip then stop)", srv.hits.Load())
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(items) != 0 {
		t.Errorf("items = %d, want 0", len(items))
	}
}

// TestPrintPageHint covers the four behaviorally-distinct branches:
// --all suppresses, full-coverage suppresses, mid-stream emits a
// "more" pointer, and past-end emits a recovery hint. Without these
// the previous copy-pasted "(showing N of M — pass --limit X)" hints
// across 6 verbs drifted; centralizing the format here keeps them in
// lockstep.
func TestPrintPageHint(t *testing.T) {
	for _, tc := range []struct {
		name      string
		p         PaginationOptions
		returned  int
		total     int
		wantSub   string // empty = expect no output
		wantEmpty bool
	}{
		{"all mode silent", PaginationOptions{Limit: 100, Page: 1, All: true}, 100, 500, "", true},
		{"full coverage silent", PaginationOptions{Limit: 100, Page: 1}, 50, 50, "", true},
		{"empty list silent", PaginationOptions{Limit: 100, Page: 1}, 0, 0, "", true},
		{"mid-stream prompts next page", PaginationOptions{Limit: 100, Page: 1}, 100, 250, "items 1-100 of 250 — pass --page 2", false},
		{"page 2 prompts page 3", PaginationOptions{Limit: 100, Page: 2}, 100, 250, "items 101-200 of 250 — pass --page 3", false},
		{"past-end recovery hint", PaginationOptions{Limit: 100, Page: 99}, 0, 250, "no items on page 99", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder
			PrintPageHint(&sb, &tc.p, tc.returned, tc.total)
			got := sb.String()
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("got %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("got %q, want substring %q", got, tc.wantSub)
			}
		})
	}
}

// TestAppendQueryForPage covers the small URL-stamping helper. Two
// branches matter: both keys present when limit > 0 && page > 0; and
// the omit-page branch when callers want a clean URL on the default
// (page=0 means "let the server default to page 1").
func TestAppendQueryForPage(t *testing.T) {
	for _, tc := range []struct {
		name     string
		opts     PaginationOptions
		page     int
		wantKeys []string
	}{
		{"both present", PaginationOptions{Limit: 50}, 3, []string{"limit", "page"}},
		{"page omitted", PaginationOptions{Limit: 50}, 0, []string{"limit"}},
		{"limit zero skipped", PaginationOptions{Limit: 0}, 2, []string{"page"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := url.Values{}
			tc.opts.AppendQueryForPage(q, tc.page)
			for _, k := range tc.wantKeys {
				if q.Get(k) == "" {
					t.Errorf("missing %s in %v", k, q)
				}
			}
		})
	}
}

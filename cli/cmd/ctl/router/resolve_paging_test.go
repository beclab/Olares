package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// pagedModels serves epProviderModels the way Router does: a window plus the
// count before it, with the window capped however large the caller asked for.
type pagedModels struct {
	rows  []adminModelRow
	cap   int
	reads int
}

func (p *pagedModels) serve(w http.ResponseWriter, r *http.Request) {
	p.reads++
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > p.cap {
		limit = p.cap
	}
	window := []adminModelRow{}
	if offset < len(p.rows) {
		end := offset + limit
		if end > len(p.rows) {
			end = len(p.rows)
		}
		window = p.rows[offset:end]
	}
	_ = json.NewEncoder(w).Encode(page[adminModelRow]{
		Items: window, Total: len(p.rows), Limit: limit, Offset: offset,
	})
}

func newPagedClient(t *testing.T, count int) (*preparedClient, *pagedModels) {
	t.Helper()
	rows := make([]adminModelRow, 0, count)
	for i := 0; i < count; i++ {
		rows = append(rows, adminModelRow{
			ProviderModelID: fmt.Sprintf("pm-%04d", i),
			ProviderName:    "openrouter",
			ProviderSource:  "manual",
			ProviderStatus:  "active",
			Model: providerModelRow{
				Name: fmt.Sprintf("model-%04d", i), Enabled: true, Status: "active",
			},
		})
	}
	srv := &pagedModels{rows: rows, cap: pageCeiling}
	ts := httptest.NewServer(http.HandlerFunc(srv.serve))
	t.Cleanup(ts.Close)
	return &preparedClient{router: newRouterClient(ts.Client(), ts.URL, "alice@example.com")}, srv
}

// The read asked for exactly the ceiling Router allows and then trusted what
// came back, so a deployment past that many models silently lost the tail. It
// is a plausible size: one `provider sync-models` against a large vendor
// catalogue is several hundred rows.
func TestEveryConfiguredModelIsRead(t *testing.T) {
	pc, srv := newPagedClient(t, pageCeiling*2+37)

	rows, err := listAllModels(context.Background(), pc)
	if err != nil {
		t.Fatalf("read the models: %v", err)
	}
	if len(rows) != pageCeiling*2+37 {
		t.Errorf("read %d models, want all %d", len(rows), pageCeiling*2+37)
	}
	if srv.reads != 3 {
		t.Errorf("made %d requests, want one per page", srv.reads)
	}
}

// The truncation never announced itself. It arrived as a refusal naming a model
// that is configured and running, with a list of suggestions it is not in —
// which reads as "somebody deleted it", and sends the reader to recreate a row
// that already exists.
func TestAModelPastTheFirstPageStillResolves(t *testing.T) {
	pc, _ := newPagedClient(t, pageCeiling+5)
	ctx := context.Background()

	last := fmt.Sprintf("openrouter/model-%04d", pageCeiling+4)
	row, err := resolveModel(ctx, pc, last)
	if err != nil {
		t.Fatalf("resolve %s: %v", last, err)
	}
	if row.Model.Name != fmt.Sprintf("model-%04d", pageCeiling+4) {
		t.Errorf("resolved to %s, want the row on the second page", row.Model.Name)
	}
}

// Paging must not cost the memo: the whole point of reading once per process
// is that a verb resolving two references does not read the catalogue twice,
// and that matters more once a read is three round trips rather than one.
func TestThePagedReadIsStillMemoized(t *testing.T) {
	pc, srv := newPagedClient(t, pageCeiling+1)
	ctx := context.Background()

	if _, err := listAllModels(ctx, pc); err != nil {
		t.Fatalf("first read: %v", err)
	}
	if _, err := listAllModels(ctx, pc); err != nil {
		t.Fatalf("second read: %v", err)
	}
	if srv.reads != 2 {
		t.Errorf("made %d requests for two reads of a two-page list, want 2", srv.reads)
	}
}

// A route that reports more rows than it will hand over cannot be paged by
// anybody. Returning the short list would put us back where we started, so the
// read fails and says what it got.
func TestAShortPageIsReportedRatherThanTruncated(t *testing.T) {
	pc, srv := newPagedClient(t, pageCeiling)
	// Claim twice as many rows as there are to serve.
	srv.rows = srv.rows[:pageCeiling]
	lying := &pagedModels{rows: srv.rows, cap: pageCeiling}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		window := []adminModelRow{}
		if offset == 0 {
			window = lying.rows
		}
		_ = json.NewEncoder(w).Encode(page[adminModelRow]{
			Items: window, Total: len(lying.rows) * 2, Limit: pageCeiling, Offset: offset,
		})
	}))
	t.Cleanup(ts.Close)
	pc = &preparedClient{router: newRouterClient(ts.Client(), ts.URL, "alice@example.com")}

	if _, err := listAllModels(context.Background(), pc); err == nil {
		t.Fatal("a list that stopped short came back as though it were complete")
	}
}

package search

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// runSessionSearch runs a search3 session-based search for the given app
// partition (files_v2 / google_drive / dropbox / knowledge). It bootstraps
// /api/search/init, pages via /api/search/more when the requested window
// extends past the first page, and best-effort cancels the session on exit.
func runSessionSearch(ctx context.Context, f *cmdutil.Factory, keyword, app, searchType string, o *pagingOptions) ([]resultItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	doer, err := newDoer(ctx, f)
	if err != nil {
		return nil, err
	}

	reqid := uuid.NewString()
	defer func() {
		_ = doEnvelope(ctx, doer, "POST", "/api/search/cancel",
			map[string]interface{}{"reqid": reqid}, nil)
	}()

	// init runs the search, caches the full (server-capped) result set, and
	// returns only the first initPageSize hits. It ignores offset/limit, so we
	// don't send them.
	initBody := map[string]interface{}{
		"reqid":   reqid,
		"keyword": keyword,
		"type":    searchType,
		"app":     app,
	}
	var initRows []json.RawMessage
	if err := doEnvelope(ctx, doer, "POST", "/api/search/init", initBody, &initRows); err != nil {
		return nil, err
	}

	// Honor --offset/--limit client-side. If the requested window already lies
	// within what init returned -- or init returned a short final page (fewer
	// than initPageSize hits means the cache holds no more) -- serve it
	// directly. Otherwise page the exact window via /search/more, whose limit
	// must stay within the backend's 1-100 range; a past-the-end offset comes
	// back as codeNoMoreResults, which we treat as an empty result set.
	var window []json.RawMessage
	if needsMorePage(o.offset, o.limit, len(initRows)) {
		moreBody := map[string]interface{}{
			"reqid":  reqid,
			"offset": o.offset,
			"limit":  clampMoreLimit(o.limit),
		}
		if err := doEnvelopeAllowing(ctx, doer, "POST", "/api/search/more", moreBody, &window, codeNoMoreResults); err != nil {
			return nil, err
		}
	} else {
		window = paginateRaw(initRows, o.offset, o.limit)
	}

	return decodeResultRows(window)
}

// requireSessionAppBackendVersion is the fail-closed preflight for the
// search3 app partitions that only exist on Olares >= 1.12.7 (google_drive,
// dropbox, knowledge). Older backends would return an opaque empty set or
// an unexpected error code; reject up front with an actionable upgrade
// message instead. The version is cached per profile at login, so in the
// common case this adds no network round-trip.
func requireSessionAppBackendVersion(ctx context.Context, f *cmdutil.Factory, verb, reason string) error {
	return cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       verb,
		MinVersion: searchSessionAppMinOlaresVersion,
		Reason:     reason,
		Fallback:   "upgrade the Olares system, or use `olares-cli search drive` for local Drive files",
	})
}

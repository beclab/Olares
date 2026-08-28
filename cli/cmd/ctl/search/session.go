package search

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// runSessionSearch runs a legacy search3 session for the given app partition.
// It bootstraps
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

	// Honor --offset/--limit client-side. /search/more's limit must stay within
	// the backend's 1-100 range, and a past-the-end offset comes back as
	// codeNoMoreResults, which we treat as an empty result set.
	var window []json.RawMessage
	if w := resolveSessionWindow(o, len(initRows)); w.needsMore {
		moreBody := map[string]interface{}{
			"reqid":  reqid,
			"offset": w.offset,
			"limit":  clampMoreLimit(w.limit),
		}
		if err := doEnvelopeAllowing(ctx, doer, "POST", "/api/search/more", moreBody, &window, codeNoMoreResults); err != nil {
			return nil, err
		}
	} else {
		window = paginateRaw(initRows, w.offset, w.limit)
	}

	return decodeResultRows(window)
}

// sessionWindow is how the requested --offset/--limit will be served from a
// session: out of the rows /search/init already returned, or by asking
// /search/more for the exact window.
type sessionWindow struct {
	offset    int
	limit     int
	needsMore bool
}

// resolveSessionWindow decides how to serve o against an init page of initLen
// rows. It resolves the window itself rather than trusting the caller to,
// because every primitive below it assumes a positive limit: "print every
// result" would otherwise read as a zero-width window that is always already
// satisfied, and would reach /search/more outside its 1-100 range.
//
// A window that lies within what init returned -- or any window at all once
// init returns a short final page, since fewer than initPageSize hits means
// the cache holds no more -- is served from the init rows.
func resolveSessionWindow(o *pagingOptions, initLen int) sessionWindow {
	paging := o.sessionPaging()
	return sessionWindow{
		offset:    paging.offset,
		limit:     paging.limit,
		needsMore: needsMorePage(paging.offset, paging.limit, initLen),
	}
}

// requireSessionAppBackendVersion is the fail-closed preflight for legacy
// search3 app partitions that only exist on Olares >= 1.12.7 (knowledge).
// Older backends would return an opaque empty set or
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

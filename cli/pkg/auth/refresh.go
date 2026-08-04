package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/olares"
)

// ErrRefreshUnauthorized is returned (wrapped) by Refresh when Authelia itself
// rejects the refresh-token grant. This is the only signal callers should treat
// as "the grant is dead, mark it invalidated and force re-login". Any other
// error from /api/refresh (transport hiccup, gateway denial, 5xx, malformed
// body) is treated as transient and surfaced verbatim so the caller can retry.
//
// A 401/403 status on its own does NOT qualify — see isAppRejection for why the
// response body has to look like Authelia's.
//
// The refresher in cli/pkg/credential/refresher.go uses errors.Is(err,
// ErrRefreshUnauthorized) to gate the MarkInvalidated → ErrTokenInvalidated
// path; do NOT collapse this with the generic error case.
var ErrRefreshUnauthorized = errors.New("refresh token rejected by server")

// RefreshRequest is the input to a single /api/refresh call. AccessToken is
// optional — the web client passes the (possibly expired) current token in
// `X-Authorization` and the server tolerates an empty value during bootstrap,
// so the CLI's `profile import` path leaves it blank.
type RefreshRequest struct {
	AuthURL            string
	RefreshToken       string
	AccessToken        string // optional, sent verbatim as X-Authorization when set
	InsecureSkipVerify bool
	Timeout            time.Duration

	// Location selects the http.Transport resolver for the refresh round-trip
	// (the `host` position resolves auth.<terminus> via the in-cluster DNS).
	// Zero value → LocationExternal, preserving the historical public path.
	Location olares.Location
}

type refreshBody struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    Token  `json:"data"`
}

// Refresh exchanges a refresh_token for a new Token via POST /api/refresh.
//
// Phase 1 uses this in two places:
//  1. `profile import` — bootstrap an access_token from a user-supplied
//     refresh_token (no current access_token to pass).
//  2. (Phase 2) Background refresh when the stored access_token is near expiry.
//
// The wire format mirrors apps/packages/app/src/utils/account.ts `refresh_token`:
// POST `<authURL>/api/refresh` with `{"refreshToken": "..."}`, optionally
// carrying `X-Authorization: <currentAccessToken>`. Response envelope is
// `{"status": "OK", "data": Token}` (same shape as /api/firstfactor).
func Refresh(ctx context.Context, req RefreshRequest) (*Token, error) {
	if req.AuthURL == "" {
		return nil, errors.New("AuthURL is required")
	}
	if req.RefreshToken == "" {
		return nil, errors.New("RefreshToken is required")
	}
	client := newHTTPClient(req.Timeout, req.Location, req.InsecureSkipVerify)

	headers := map[string]string{}
	if req.AccessToken != "" {
		headers["X-Authorization"] = req.AccessToken
	}
	resp, err := postJSON(ctx, client, req.AuthURL+"/api/refresh", refreshBody{RefreshToken: req.RefreshToken}, headers)
	if err != nil {
		return nil, fmt.Errorf("/api/refresh: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read /api/refresh body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// An Authelia rejection means the refresh-token grant itself is
		// no longer honored — the only recovery is a fresh login. Wrap
		// with the typed sentinel so credential.Refresher can stamp
		// InvalidatedAt and surface ErrTokenInvalidated. Any other
		// non-200 stays an opaque error (treated as transient).
		if isAppRejection(resp.StatusCode, raw) {
			return nil, fmt.Errorf("%w: /api/refresh returned HTTP %d: %s", ErrRefreshUnauthorized, resp.StatusCode, truncate(raw))
		}
		return nil, fmt.Errorf("/api/refresh returned HTTP %d: %s", resp.StatusCode, truncate(raw))
	}
	var parsed refreshResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse /api/refresh body: %w (body=%s)", err, truncate(raw))
	}
	if !strings.EqualFold(parsed.Status, "OK") {
		msg := parsed.Status
		if parsed.Message != "" {
			msg = msg + ": " + parsed.Message
		}
		return nil, fmt.Errorf("refresh failed: %s", msg)
	}
	if parsed.Data.AccessToken == "" {
		return nil, errors.New("refresh returned empty access_token")
	}
	// The server occasionally omits a fresh refresh_token (rotating policy
	// disabled). Fall back to the caller-supplied one so the next refresh has
	// something to send.
	if parsed.Data.RefreshToken == "" {
		parsed.Data.RefreshToken = req.RefreshToken
	}
	return &parsed.Data, nil
}

// isAppRejection reports whether an unsuccessful /api/refresh response is
// Authelia turning down the grant, as opposed to something in front of it
// refusing to forward the request at all.
//
// 401/403 are Authelia's own codes. 459 is Olares' edge (Authelia ext-authz
// wired through l4-bfl-proxy) signalling "auth failed"; /api/refresh sits on
// auth.<terminus> which usually doesn't pass through that filter, but we treat
// it the same to stay consistent with the SPA's mapping in
// apps/packages/app/src/platform/platformAjaxSender.ts:89 (459 →
// ErrorCode.TOKE_INVILID).
//
// The status code alone is not enough. Envoy answers requests that fail its
// RBAC policy — reaching the public entrance while the instance is VPN-only,
// say — with a bare `403 RBAC: access denied` in text/plain, having never
// forwarded anything to Authelia. The grant is untouched in that case, so
// stamping InvalidatedAt would force a pointless re-login and, because the
// stamp is persisted to a shared store, also strand every other client that
// does reach the instance over a working route. Authelia phrases its own
// rejections as JSON objects (`{"status":"KO",...}` from /api/refresh,
// `{"fa2":false,...}` from the 459 filter), so require one before declaring the
// grant dead.
func isAppRejection(status int, body []byte) bool {
	if status != http.StatusUnauthorized &&
		status != http.StatusForbidden &&
		status != 459 {
		return false
	}
	var probe map[string]json.RawMessage
	return json.Unmarshal(body, &probe) == nil && probe != nil
}

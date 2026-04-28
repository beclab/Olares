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
)

// ErrRefreshUnauthorized is returned (wrapped) by Refresh when the server
// rejects the refresh-token grant with HTTP 401/403. This is the only signal
// callers should treat as "the grant is dead, mark it invalidated and force
// re-login". Any other error from /api/refresh (transport hiccup, 5xx, malformed
// body) is treated as transient and surfaced verbatim so the caller can retry.
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
	client := newHTTPClient(req.Timeout, req.InsecureSkipVerify)

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
		// 401/403 means the refresh-token grant itself is no longer
		// honored — the only recovery is a fresh login. Wrap with the
		// typed sentinel so credential.Refresher can stamp
		// InvalidatedAt and surface ErrTokenInvalidated. Any other
		// non-200 stays an opaque error (treated as transient).
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
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

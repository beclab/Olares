// Package auth provides olares-cli's authentication primitives: JWT expiry
// extraction, password-based login (/api/firstfactor + /api/secondfactor/totp),
// refresh-token bootstrap (/api/refresh), and the on-disk token store.
//
// jwt.go intentionally exposes ONLY ExpiresAt(). The CLI does NOT verify JWT
// signatures (it has no signing key), so all other claims (`username`,
// `groups`, `mfa`, `jid`, ...) are untrusted and must not leak into UX. The
// only JWT field treated as a "hint" is `exp`, because faking it can only
// trigger a self-inflicted 401 from the server. See §7.5 of
// docs/notes/olares-cli-auth-profile-config.md for the full rationale.
package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// expClaim is the minimal JWT payload subset we ever decode. We deliberately
// avoid Unmarshalling into a richer struct so reviewers can audit at a glance
// that no other claim ever escapes this package.
type expClaim struct {
	Exp int64 `json:"exp"`
}

// ExpiresAt decodes only the `exp` claim of a JWT (header.payload.signature)
// and returns it as a time.Time. It does NOT verify the signature. Use the
// returned value as a client-side hint only; the server remains the source of
// truth for token validity.
//
// Returns an error if the input doesn't look like a JWT, the payload can't be
// base64url-decoded, or the JSON is malformed. Tokens with no `exp` claim
// produce (zero time, ErrNoExpClaim).
func ExpiresAt(token string) (time.Time, error) {
	if token == "" {
		return time.Time{}, errors.New("token is empty")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("token does not look like a JWT (want 3 segments, got %d)", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Some encoders include `=` padding; tolerate that.
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("decode payload: %w", err)
		}
	}
	var c expClaim
	if err := json.Unmarshal(payload, &c); err != nil {
		return time.Time{}, fmt.Errorf("parse payload: %w", err)
	}
	if c.Exp == 0 {
		return time.Time{}, ErrNoExpClaim
	}
	return time.Unix(c.Exp, 0), nil
}

// ErrNoExpClaim is returned by ExpiresAt when the JWT payload has no `exp`
// field. Callers can treat this as "unknown expiry" and decide their own
// policy (Phase 1 conservatively treats unknown as "trust the token until the
// server says otherwise").
var ErrNoExpClaim = errors.New("jwt has no exp claim")

// IsExpired returns true if ExpiresAt(token) is non-zero AND in the past
// relative to now (or within `skew` of now). Tokens with no exp claim or
// malformed tokens return (false, err).
//
// `skew` is treated as a non-negative leeway; pass 0 for exact comparison.
func IsExpired(token string, now time.Time, skew time.Duration) (bool, error) {
	exp, err := ExpiresAt(token)
	if err != nil {
		return false, err
	}
	return !now.Add(skew).Before(exp), nil
}

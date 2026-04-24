package auth

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// Token mirrors the Authelia /api/firstfactor + /api/secondfactor/totp +
// /api/refresh response payload shared by Olares.
//
// We keep only the fields the CLI actually persists or inspects. The wire
// format historically also returns `expires_in`, `expires_at`, `fa2`, etc;
// the CLI ignores the time fields (auth.ExpiresAt(AccessToken) is the source
// of truth) but does honor `fa2` to detect when a TOTP step is required.
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	FA2          bool   `json:"fa2,omitempty"`
}

// LoginRequest captures everything Login needs to perform first-factor (and,
// if needed, second-factor TOTP) authentication.
//
// AuthURL is the Olares auth subdomain base, e.g. "https://auth.alice.olares.com".
// The CLI POSTs to AuthURL + "/api/firstfactor" and AuthURL + "/api/secondfactor/totp".
//
// LocalName is the bare username (the part before `@` of the olaresId).
// The web app uses this as `username` in the request body.
//
// TerminusName is "<local>.<domain>"; it's used to construct the second-factor
// `targetUrl` field (https://desktop.<terminusName>/) which the auth backend
// echoes back as the redirect target.
//
// TOTP is optional — supply it when the account has 2FA enabled. If the
// first-factor response indicates FA2 is required and TOTP is empty, Login
// returns ErrTOTPRequired so the caller can prompt and retry.
type LoginRequest struct {
	AuthURL            string
	LocalName          string
	TerminusName       string
	Password           string
	TOTP               string
	InsecureSkipVerify bool
	Timeout            time.Duration // zero → 10s default
}

// ErrTOTPRequired is returned from Login when the first-factor response
// reports FA2 is needed but the caller didn't supply a TOTP code. Callers
// (e.g. `profile login`) can prompt the user and call Login again with TOTP set.
var ErrTOTPRequired = errors.New("two-factor authentication is required: re-run with --totp <code>")

// Login executes the password login flow:
//  1. POST /api/firstfactor with the salted-MD5 password.
//  2. If the response says fa2 is required, POST /api/secondfactor/totp with
//     the supplied TOTP code (or return ErrTOTPRequired if none was given).
//
// On success the returned Token contains the freshly minted access_token and
// refresh_token (the second-factor response overrides them when present).
//
// The function uses a short-lived http.Client with a cookie jar so that the
// Authelia session cookie set on /api/firstfactor is automatically attached
// to /api/secondfactor/totp — mirroring `withCredentials: true` in the web
// implementation in apps/packages/app/src/utils/BindTerminusBusiness.ts.
func Login(ctx context.Context, req LoginRequest) (*Token, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}
	client := newHTTPClient(req.Timeout, req.InsecureSkipVerify)

	tok, err := postFirstFactor(ctx, client, req)
	if err != nil {
		return nil, err
	}
	if !tok.FA2 {
		return tok, nil
	}
	if req.TOTP == "" {
		return nil, ErrTOTPRequired
	}
	tok2, err := postSecondFactorTOTP(ctx, client, req, tok.AccessToken)
	if err != nil {
		return nil, err
	}
	// Carry forward whatever the second-factor response refreshed; keep the
	// first-factor SessionID as a fallback if the server returned an empty one.
	if tok2.SessionID == "" {
		tok2.SessionID = tok.SessionID
	}
	return tok2, nil
}

func validateLoginRequest(req LoginRequest) error {
	switch {
	case req.AuthURL == "":
		return errors.New("AuthURL is required")
	case req.LocalName == "":
		return errors.New("LocalName is required")
	case req.TerminusName == "":
		return errors.New("TerminusName is required")
	case req.Password == "":
		return errors.New("Password is required")
	}
	return nil
}

// passwordSalt mirrors the `passwordAddSort` helper in
// pkg/wizard/auth.go (and its TS counterpart in BindTerminusBusiness.ts):
// MD5 of `<password>@Olares2025`. The salt is a public, account-independent
// constant — it's NOT a security feature, just a wire-format quirk the auth
// backend expects.
func passwordSalt(password string) string {
	hash := md5.Sum([]byte(password + "@Olares2025"))
	return fmt.Sprintf("%x", hash)
}

type firstFactorBody struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	KeepMeLoggedIn bool   `json:"keepMeLoggedIn"`
	RequestMethod  string `json:"requestMethod"`
	TargetURL      string `json:"targetURL"`
	AcceptCookie   bool   `json:"acceptCookie"`
}

type firstFactorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    Token  `json:"data"`
}

func postFirstFactor(ctx context.Context, client *http.Client, req LoginRequest) (*Token, error) {
	body := firstFactorBody{
		Username:       req.LocalName,
		Password:       passwordSalt(req.Password),
		KeepMeLoggedIn: false,
		RequestMethod:  "POST",
		// Always declare the desktop subdomain as the post-login redirect target.
		// Authelia's `fa2` flag in the response is computed against this URL via
		// its access-control policy, and only the desktop.<terminusName>/ rule
		// requires 2FA. Sending the auth or vault URL would silently downgrade
		// the response to 1FA, hiding the fact that the account has 2FA enabled.
		// See apps/packages/app/src/utils/account.ts (onFirstFactor) for the
		// matching web behavior.
		TargetURL:    "https://desktop." + req.TerminusName + "/",
		AcceptCookie: true,
	}
	resp, err := postJSON(ctx, client, req.AuthURL+"/api/firstfactor?hideCookie=true", body, nil)
	if err != nil {
		return nil, fmt.Errorf("/api/firstfactor: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read /api/firstfactor body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("/api/firstfactor returned HTTP %d: %s", resp.StatusCode, truncate(raw))
	}
	var parsed firstFactorResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse /api/firstfactor body: %w (body=%s)", err, truncate(raw))
	}
	if !strings.EqualFold(parsed.Status, "OK") {
		msg := parsed.Status
		if parsed.Message != "" {
			msg = msg + ": " + parsed.Message
		}
		return nil, fmt.Errorf("first-factor authentication failed: %s", msg)
	}
	return &parsed.Data, nil
}

type secondFactorBody struct {
	TargetURL string `json:"targetUrl"`
	Token     string `json:"token"`
}

func postSecondFactorTOTP(ctx context.Context, client *http.Client, req LoginRequest, firstFactorAccessToken string) (*Token, error) {
	// `targetUrl` echoes the eventual redirect destination the web app would
	// be sent to after a successful second factor. The auth backend validates
	// its scheme/host but otherwise just relays it back, so we hard-code the
	// desktop subdomain pattern to match BindTerminusBusiness.ts.
	body := secondFactorBody{
		TargetURL: "https://desktop." + req.TerminusName + "/",
		Token:     req.TOTP,
	}
	headers := map[string]string{
		"X-Authorization": firstFactorAccessToken,
		"X-Unauth-Error":  "Non-Redirect",
	}
	resp, err := postJSON(ctx, client, req.AuthURL+"/api/secondfactor/totp", body, headers)
	if err != nil {
		return nil, fmt.Errorf("/api/secondfactor/totp: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read /api/secondfactor/totp body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("/api/secondfactor/totp returned HTTP %d: %s", resp.StatusCode, truncate(raw))
	}
	var parsed firstFactorResponse // identical envelope as first-factor
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse /api/secondfactor/totp body: %w (body=%s)", err, truncate(raw))
	}
	if !strings.EqualFold(parsed.Status, "OK") {
		msg := parsed.Status
		if parsed.Message != "" {
			msg = msg + ": " + parsed.Message
		}
		return nil, fmt.Errorf("second-factor authentication failed: %s", msg)
	}
	return &parsed.Data, nil
}

// postJSON marshals `body` as JSON, posts it to `url` via `client`, and
// returns the raw response. Callers must close resp.Body.
func postJSON(ctx context.Context, client *http.Client, url string, body any, headers map[string]string) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	return client.Do(httpReq)
}

// newHTTPClient returns an http.Client suitable for auth flows: short timeout,
// cookie jar (so first-factor session cookies attach to second-factor), and
// optional InsecureSkipVerify for dev environments.
func newHTTPClient(timeout time.Duration, insecure bool) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	jar, _ := cookiejar.New(nil)
	c := &http.Client{
		Timeout: timeout,
		Jar:     jar,
	}
	if insecure {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- dev override gated behind explicit flag
		}
	}
	return c
}

// truncate caps a body snippet for inclusion in error messages.
func truncate(b []byte) string {
	const max = 256
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}

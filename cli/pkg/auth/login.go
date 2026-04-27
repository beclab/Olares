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

	"github.com/beclab/Olares/cli/pkg/olares"
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

// LoginRequest is the input shared by FirstFactor (low-level, equivalent
// to the TS onFirstFactor primitive) and Login (high-level wrapper around
// FirstFactor + optional /api/secondfactor/totp, equivalent to the TS
// loginTerminus flow).
//
// Field semantics mirror the TS web reference 1:1; if you change anything
// here, also re-read those two TS functions and keep them aligned:
//
//   - apps/packages/app/src/utils/account.ts L7-71               (onFirstFactor)
//   - apps/packages/app/src/utils/BindTerminusBusiness.ts L353-446 (loginTerminus)
//
// AuthURL is the Olares auth base, e.g. "https://auth.alice.olares.com".
// The CLI POSTs to AuthURL + "/api/firstfactor" and AuthURL + "/api/secondfactor/totp".
//
// LocalName is the bare username (the part before `@` of the olaresId).
// The web app uses this as `username` in the request body.
//
// TerminusName is "<local>.<domain>"; it's only used to derive the
// `targetURL` form field (vault.<name>/server by default,
// desktop.<name>/ when NeedTwoFactor is true) and the second-factor
// `targetUrl`.
//
// TOTP is optional — supply it when the account has 2FA enabled. Login
// returns ErrTOTPRequired when 2FA is needed (tok.FA2 || NeedTwoFactor)
// but TOTP is empty. FirstFactor never reads TOTP.
//
// NeedTwoFactor mirrors the `needTwoFactor` parameter on TS onFirstFactor:
// when true, swap targetURL from `vault.<name>/server` to
// `desktop.<name>/`. This is the ONLY thing it does in the Go API.
//
// Authelia's per-URL access policy is what makes `fa2` flip to true in
// the response — the vault URL maps to a 1FA policy, the desktop URL
// maps to a 2FA policy. Callers that want the server to honestly tell
// them whether the account has 2FA enabled (e.g. `profile login`'s
// initial probe) MUST pass NeedTwoFactor=true so Authelia evaluates the
// 2FA policy. NeedTwoFactor does NOT participate in Login's escalation
// gate — Login uses `tok.FA2` from the server only; see Login's doc for
// why we diverge from TS's `tok.FA2 || needTwoFactor` here.
//
// AcceptCookie mirrors the `acceptCookie` parameter on TS onFirstFactor;
// it is passed through verbatim into the request body. callers known to
// follow up with /api/secondfactor/totp pass true (so Authelia sets the
// session cookie that the second-factor request needs); the
// activation/signup caller (cli/pkg/wizard.UserBindTerminus) passes false.
type LoginRequest struct {
	AuthURL            string
	LocalName          string
	TerminusName       string
	Password           string
	TOTP               string
	NeedTwoFactor      bool
	AcceptCookie       bool
	InsecureSkipVerify bool
	Timeout            time.Duration // zero → 10s default
}

// ErrTOTPRequired is returned from Login when the first-factor response
// reports FA2 is needed but the caller didn't supply a TOTP code. Callers
// (e.g. `profile login`) can prompt the user and call Login again with TOTP set.
var ErrTOTPRequired = errors.New("two-factor authentication is required: re-run with --totp <code>")

// FirstFactor performs a single POST /api/firstfactor and returns the raw
// token. Mirrors apps/packages/app/src/utils/account.ts:onFirstFactor (L7-71)
// 1:1: it does NOT inspect or act on the response's `fa2` flag — choosing
// whether to escalate to /api/secondfactor/totp is the caller's job.
//
// Two callers exist today:
//
//   - Login (this file) wraps FirstFactor and does the
//     `(tok.FA2 || NeedTwoFactor)` escalation, mirroring TS loginTerminus.
//   - cli/pkg/wizard.UserBindTerminus uses FirstFactor directly and
//     ignores fa2, mirroring TS userBindTerminus — at signup time there is
//     no MFA seed yet, so the 1st-factor access_token is what the
//     subsequent signup endpoints need.
func FirstFactor(ctx context.Context, req LoginRequest) (*Token, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}
	client := newHTTPClient(req.Timeout, req.InsecureSkipVerify)
	return firstFactorWithClient(ctx, client, req)
}

// Login executes the full password login flow:
//
//  1. POST /api/firstfactor with the salted-MD5 password (via FirstFactor).
//  2. If the server reports `tok.FA2`, POST /api/secondfactor/totp with the
//     supplied TOTP code (or return ErrTOTPRequired if none was given).
//
// Mirrors apps/packages/app/src/utils/BindTerminusBusiness.ts:loginTerminus
// (L353-446), with one deliberate divergence: the gate is `tok.FA2` only,
// not the TS `tok.FA2 || needTwoFactor`. The TS code OR's in
// `needTwoFactor` so the web UI can *force* 2FA when it locally knows the
// user has it but the server hasn't reported it (defensive UI-state
// pattern). The CLI has no such caller-side knowledge — it can only
// trust whatever the server says — and gating on the OR would make
// non-2FA users (who get fa2=false) hit a spurious ErrTOTPRequired the
// moment a caller passes NeedTwoFactor=true to probe with the desktop
// targetURL (e.g. `profile login`).
//
// FirstFactor and the optional second-factor POST share a single
// http.Client (with cookie jar) so the Authelia session cookie set on
// /api/firstfactor automatically attaches to /api/secondfactor/totp,
// mirroring `withCredentials: true` in the TS axios instance.
func Login(ctx context.Context, req LoginRequest) (*Token, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}
	client := newHTTPClient(req.Timeout, req.InsecureSkipVerify)

	tok, err := firstFactorWithClient(ctx, client, req)
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

// PasswordSalt is the md5(`<password>@Olares2025`) wire-format the Authelia
// backend expects on /api/firstfactor and on the bfl
// /iam/v1alpha1/users/<name>/password reset endpoint. The salt string is a
// public, account-independent constant — it is NOT a security feature, only
// a quirk we have to reproduce on every code path that talks to those two
// endpoints. The TS counterpart is `passwordAddSort` in
// apps/packages/app/src/utils/BindTerminusBusiness.ts.
//
// Exported so cli/pkg/wizard.ResetPassword can reuse the same implementation
// instead of carrying its own copy — having two copies invites silent drift
// the day someone changes the salt server-side.
func PasswordSalt(password string) string {
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

// firstFactorWithClient is the shared implementation behind FirstFactor and
// Login. Splitting it out lets Login reuse the cookie-jarred client across
// /api/firstfactor and /api/secondfactor/totp without re-dialling.
//
// targetURL derivation matches apps/packages/app/src/utils/account.ts
// L19-26: vault.<name>/server by default, desktop.<name>/ when the caller
// asks for the 2FA-bearing policy via NeedTwoFactor.
func firstFactorWithClient(ctx context.Context, client *http.Client, req LoginRequest) (*Token, error) {

	id, err := olares.ParseID(req.TerminusName)
	if err != nil {
		return nil, err
	}

	targetURL := id.VaultURL("")
	if req.NeedTwoFactor {
		targetURL = id.DesktopURL("")
	}
	body := firstFactorBody{
		Username:       req.LocalName,
		Password:       PasswordSalt(req.Password),
		KeepMeLoggedIn: false,
		RequestMethod:  "POST",
		TargetURL:      targetURL,
		AcceptCookie:   req.AcceptCookie,
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

	id, err := olares.ParseID(req.TerminusName)
	if err != nil {
		return nil, err
	}

	body := secondFactorBody{
		TargetURL: id.DesktopURL(""),
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

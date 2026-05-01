// Package cmdutil holds the shared "Factory" that command implementations
// reach into instead of constructing their own clients / loading their own
// config / etc. This is the olares-cli analogue of lark-cli's
// cmdutil.Factory.
//
// The Factory wires three things together:
//   - the lazily-resolved credential chain (env + on-disk profile),
//   - a credential.Refresher that owns /api/refresh + InvalidatedAt
//     bookkeeping with cross-process locking,
//   - one or more *http.Client instances whose RoundTripper transparently
//     injects the active access_token via X-Authorization (NOT
//     Authorization: Bearer — see refreshingTransport rationale below) AND
//     refreshes on 401/403, retrying the original request once.
//
// The refreshingTransport is shared via a *tokenCell so both the
// timed (HTTPClient) and untimed (HTTPClientWithoutTimeout) clients see
// the same access_token at all times — a refresh on one updates the other.
package cmdutil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// preflightSkew is the safety window applied when deciding whether to
// proactively refresh a non-replayable request's token. If the JWT's exp
// is within this margin of now (or already past), we rotate the token
// BEFORE handing the request body to the base transport.
//
// Why this matters: streaming bodies (typically *os.File for `files
// upload` chunks) get consumed by the first send and cannot be replayed
// on a 401. Without pre-flight the user would see a single failed chunk
// and have to re-run the whole upload command. 60s comfortably covers
// client↔server clock drift plus the time from local JWT decode to the
// request actually landing on the server.
//
// Replayable bodies (every JSON / files-cat / files-rm / market verb)
// deliberately skip pre-flight: the reactive 401 path is one extra HTTP
// round-trip but avoids decoding the JWT on every request, and the
// outcome is identical from the user's perspective.
const preflightSkew = 60 * time.Second

// Factory is the dependency-injection seam for olares-cli commands. Build
// one with NewFactory at the root command level and pass it (or a closure
// that closes over it) into command constructors.
//
// All accessors are lazy and memoized — calling HTTPClient(ctx) multiple
// times reuses the same resolved profile + client.
type Factory struct {
	// ProfileOverride, when non-empty, forces ResolveProfile to look up this
	// profile instead of the currently-selected one. Wired from the root
	// command's persistent `--profile` flag.
	ProfileOverride string

	credentialOnce sync.Once
	credentialErr  error
	credential     *credential.CredentialProvider

	resolveOnce sync.Once
	resolveErr  error
	resolved    *credential.ResolvedProfile

	refresherOnce sync.Once
	refresher     *credential.Refresher

	// tokenCell is the shared mutable access_token cell. Both the timed
	// and untimed http.Clients hold the same pointer so a refresh
	// triggered through one is immediately visible to the other.
	tokenCellOnce sync.Once
	tokenCell     *tokenCell

	clientOnce sync.Once
	client     *http.Client

	uploadClientOnce sync.Once
	uploadClient     *http.Client
}

// NewFactory builds a fresh Factory. Cheap; intended to be called once per
// process from the root command.
func NewFactory() *Factory {
	return &Factory{}
}

// Credential returns the lazily-constructed credential chain.
func (f *Factory) Credential() (*credential.CredentialProvider, error) {
	f.credentialOnce.Do(func() {
		def, err := credential.NewDefaultProvider()
		if err != nil {
			f.credentialErr = fmt.Errorf("init default credential provider: %w", err)
			return
		}
		f.credential = credential.NewCredentialProvider(
			credential.NewEnvProvider(),
			def,
		)
	})
	return f.credential, f.credentialErr
}

// ResolveProfile returns the active profile fully resolved (URLs + token).
// Memoized; subsequent calls return the same ResolvedProfile.
func (f *Factory) ResolveProfile(ctx context.Context) (*credential.ResolvedProfile, error) {
	f.resolveOnce.Do(func() {
		cred, err := f.Credential()
		if err != nil {
			f.resolveErr = err
			return
		}
		rp, err := cred.Resolve(ctx, f.ProfileOverride)
		if err != nil {
			f.resolveErr = err
			return
		}
		f.resolved = rp
	})
	return f.resolved, f.resolveErr
}

// Refresher returns the lazily-constructed cross-process token refresher.
// Exported for tests; production code reaches it through HTTPClient's
// transport.
func (f *Factory) Refresher() *credential.Refresher {
	f.refresherOnce.Do(func() {
		f.refresher = credential.NewRefresher()
	})
	return f.refresher
}

// HTTPClient returns the standard http.Client used for short JSON requests.
// Its RoundTripper:
//   - injects the active access_token via X-Authorization on every request,
//   - on 401/403, calls Refresher() to rotate the token and retries the
//     request once (only when req.GetBody is set — i.e. the body is
//     replayable),
//   - returns the original error / 401 if refresh itself fails or the body
//     can't be replayed.
//
// The 30s overall timeout matches the previous behavior. For streaming
// uploads (large multipart bodies) use HTTPClientWithoutTimeout.
func (f *Factory) HTTPClient(ctx context.Context) (*http.Client, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	f.clientOnce.Do(func() {
		f.client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: f.newRefreshingTransport(rp),
		}
	})
	return f.client, nil
}

// HTTPClientWithoutTimeout returns an http.Client with no overall timeout,
// suitable for streaming uploads / long-running multipart requests. It
// shares the same access_token cell as HTTPClient so a refresh triggered
// through either client is immediately visible to the other.
func (f *Factory) HTTPClientWithoutTimeout(ctx context.Context) (*http.Client, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	f.uploadClientOnce.Do(func() {
		f.uploadClient = &http.Client{
			Transport: f.newRefreshingTransport(rp),
		}
	})
	return f.uploadClient, nil
}

// tokenCell stores the current access_token under a RWMutex so concurrent
// RoundTrip calls can read it cheaply and a successful refresh can swap it
// in with one writer-side lock.
type tokenCell struct {
	mu    sync.RWMutex
	token string
}

func (c *tokenCell) snapshot() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

func (c *tokenCell) update(newToken string) {
	if newToken == "" {
		return
	}
	c.mu.Lock()
	c.token = newToken
	c.mu.Unlock()
}

func (f *Factory) sharedTokenCell(initial string) *tokenCell {
	f.tokenCellOnce.Do(func() {
		f.tokenCell = &tokenCell{token: initial}
	})
	return f.tokenCell
}

// newRefreshingTransport builds a refreshingTransport bound to rp. Both
// HTTPClient and HTTPClientWithoutTimeout reuse the same Refresher and the
// same tokenCell so a refresh on one is immediately visible on the other.
func (f *Factory) newRefreshingTransport(rp *credential.ResolvedProfile) *refreshingTransport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	if rp.InsecureSkipVerify {
		base.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	return &refreshingTransport{
		base:               base,
		olaresID:           rp.OlaresID,
		authURL:            rp.AuthURL,
		insecureSkipVerify: rp.InsecureSkipVerify,
		refresher:          f.Refresher(),
		token:              f.sharedTokenCell(rp.AccessToken),
	}
}

// refreshingTransport injects the current access_token via X-Authorization,
// and on 401/403 transparently refreshes the token and retries the original
// request once.
//
// Why X-Authorization (not the standard Authorization: Bearer)?
// Olares' edge stack — Authelia ext-authz wired through l4-bfl-proxy —
// inspects `X-Authorization` (and Cookie) to identify the user; see
// framework/l4-bfl-proxy/internal/translator/translator.go (RequestHeaders
// allow-list) and the BFL backend's
// framework/bfl/pkg/apiserver/filters.go (UserAuthorizationTokenKey =
// "X-Authorization"). The standard Authorization header is filtered out by
// the edge before it reaches per-user services, so X-Authorization is the
// only value that round-trips to the backend today. The web app does the
// same thing in apps/packages/app/src/platform/platformAjaxSender.ts.
//
// Retry is best-effort: we only retry when req.GetBody is set (or the body
// is nil). Streaming bodies (e.g. files upload chunks backed by *os.File)
// can't be replayed; their 401 falls through verbatim and the user re-runs
// the command. http.NewRequest sets GetBody automatically for the standard
// rewindable body types (*bytes.Reader, *bytes.Buffer, *strings.Reader)
// which covers all current JSON callers.
type refreshingTransport struct {
	base               http.RoundTripper
	olaresID           string
	authURL            string
	insecureSkipVerify bool
	refresher          *credential.Refresher
	token              *tokenCell

	// now is the clock used by preflightRefresh's expiry check. nil =
	// time.Now (production); tests override it to drive the skew
	// window deterministically.
	now func() time.Time
}

func (t *refreshingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	snap := t.token.snapshot()

	// Pre-flight refresh for non-replayable bodies. Once a streaming
	// body (e.g. *os.File chunk for files upload) is handed to the
	// base transport, the first byte that goes on the wire consumes
	// the read offset; a 401 response can't trigger a retry because
	// we have no way to rewind. Defensively check the JWT's exp claim
	// locally and rotate the token here instead.
	//
	// Replayable bodies fall through to the reactive 401 path below —
	// it's one extra HTTP round-trip in the rare expiry case, but
	// avoids a JWT decode on every request.
	if !canRetry(req) {
		newAT, ok, err := t.preflightRefresh(req.Context(), snap)
		if err != nil {
			return nil, err
		}
		if ok {
			snap = newAT
		}
	}

	resp, err := t.send(req, snap)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		return resp, nil
	}
	if !canRetry(req) {
		// Non-replayable body — pre-flight already had its chance
		// (and either rotated the token or decided not to). Either
		// way, body is consumed; surface the 401 verbatim.
		return resp, nil
	}
	// Drain + close so the underlying connection can be reused.
	drainAndClose(resp)

	newAT, rerr := t.refresher.Refresh(req.Context(), t.olaresID, t.authURL, snap, t.insecureSkipVerify)
	if rerr != nil {
		// Refresh itself failed (network, 5xx) or the grant is dead
		// (ErrTokenInvalidated, ErrNotLoggedIn). Surface the typed
		// error so the caller's reformatter can render the
		// "run profile login" CTA.
		return nil, rerr
	}
	t.token.update(newAT)

	retryReq, err := cloneWithBody(req)
	if err != nil {
		return nil, fmt.Errorf("clone request for retry: %w", err)
	}
	return t.send(retryReq, newAT)
}

// preflightRefresh decides whether snap is close enough to expiring
// that a non-replayable request must rotate it BEFORE sending. Returns:
//
//   - (newAT, true, nil)  — refresh fired, caller should send with newAT
//   - ("",   false, nil)  — token still fresh, no exp claim, malformed,
//     or empty; caller should send with the original snap
//   - ("",   false, err)  — refresh attempted and failed; caller MUST
//     return err without sending (the streaming body is still intact)
//
// Refresher.Refresh has its own in-process mutex + cross-process flock,
// so concurrent chunk uploads collapse to a single /api/refresh hit
// even when they all decide to pre-flight at the same moment.
func (t *refreshingTransport) preflightRefresh(ctx context.Context, snap string) (string, bool, error) {
	if snap == "" {
		return "", false, nil
	}
	expired, err := auth.IsExpired(snap, t.clock(), preflightSkew)
	if err != nil {
		// No exp claim / non-JWT / malformed payload. We can't make
		// a meaningful local decision; let the request fly. For a
		// non-replayable body that means the user may see a 401,
		// but refusing to send a known-otherwise-valid token is
		// strictly worse UX (would break tokens issued without exp).
		return "", false, nil
	}
	if !expired {
		return "", false, nil
	}
	newAT, rerr := t.refresher.Refresh(ctx, t.olaresID, t.authURL, snap, t.insecureSkipVerify)
	if rerr != nil {
		return "", false, rerr
	}
	t.token.update(newAT)
	return newAT, true, nil
}

func (t *refreshingTransport) clock() time.Time {
	if t.now != nil {
		return t.now()
	}
	return time.Now()
}

// send injects access_token (when non-empty) and forwards to the base
// transport. Three pieces of auth state ride together — see the SPA's
// successful /capi/app/detail request (captured 2026-05-01) for the
// exact recipe:
//
//  1. X-Authorization: <jwt>
//     The Olares edge stack (Authelia ext-authz → l4-bfl-proxy) reads
//     this for identity. Set on every request as the primary auth
//     signal. See framework/l4-bfl-proxy/internal/translator/translator.go
//     (RequestHeaders allow-list) and framework/bfl/pkg/apiserver/
//     filters.go (UserAuthorizationTokenKey = "X-Authorization").
//
//  2. Cookie: auth_token=<jwt>
//     Per-service hosts (dashboard.<terminus>, control-hub.<terminus>,
//     files.<terminus>, market.<terminus>) emit `Set-Cookie:
//     auth_token=<jwt>; domain=<terminus>; path=/` on every 200, and
//     several /capi/* handlers (notably /capi/app/detail and
//     /capi/namespaces/group) read identity from THIS cookie rather
//     than X-Authorization — without it they reply `500 Not Login`
//     (BFL) or `403 system:anonymous` (controlhub→K8s API server).
//     The cookie value is the same JWT we put in X-Authorization.
//
//     The earlier KI-12 / KI-15 / KI-16 note (committed 2026-04-28)
//     said `auth_token` cookie broke 31 verbs because Authelia treated
//     it as its own session slot. That symptom is real but only on
//     desktop.<terminus>, where Authelia's ext-authz session lives.
//     The CLI doesn't talk to desktop.<terminus> — every command goes
//     to a per-service host, where this cookie IS the canonical
//     identity carrier (proven by the browser's successful 200s on
//     dashboard.<terminus>/capi/app/detail using exactly this header).
//
//  3. X-Unauth-Error: Non-Redirect
//     Tells the edge to return JSON 401 instead of an HTML 302 redirect
//     to /login when auth fails. Without it, an expired token surfaces
//     as a JSON-decode parse error (the historical "400 Bad Request
//     HTML response" symptom referenced in the KI-12/15/16 note).
//
// Clones the request so the caller's *http.Request is left untouched
// (some callers retry at a higher level). AddCookie appends to any
// pre-existing Cookie header rather than replacing it; in the unlikely
// event the caller pre-set its own auth_token cookie both values are
// sent (per RFC 6265 §5.4 the server picks one, typically the first).
func (t *refreshingTransport) send(req *http.Request, token string) (*http.Response, error) {
	if token == "" {
		return t.base.RoundTrip(req)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("X-Authorization", token)
	clone.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	clone.Header.Set("X-Unauth-Error", "Non-Redirect")
	return t.base.RoundTrip(clone)
}

// canRetry reports whether req's body can be replayed. A request with no
// body is trivially replayable; a request with a body is replayable only
// when http.NewRequest already populated GetBody (set automatically for
// *bytes.Reader, *bytes.Buffer, *strings.Reader).
func canRetry(req *http.Request) bool {
	if req.Body == nil || req.Body == http.NoBody {
		return true
	}
	return req.GetBody != nil
}

// cloneWithBody clones req for a retry. It uses GetBody to obtain a fresh
// reader; callers must have already gated on canRetry.
func cloneWithBody(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body == nil || req.Body == http.NoBody {
		return clone, nil
	}
	if req.GetBody == nil {
		return nil, errors.New("request has a non-replayable body")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("rewind body: %w", err)
	}
	clone.Body = body
	return clone, nil
}

// drainAndClose drains and closes resp.Body so the underlying connection
// can be re-used by the retry. Errors are intentionally ignored — by the
// time we get here the response is already a 401/403 we're discarding.
func drainAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

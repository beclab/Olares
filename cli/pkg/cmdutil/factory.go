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
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/beclab/Olares/cli/pkg/access"
	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// locationCooldown is how long after an "every probe failed" result the CLI
// skips a fresh (slow) re-probe and fails fast instead. Short enough to
// self-heal quickly once connectivity returns, long enough that a burst of
// commands during an outage doesn't each pay the full probe cost.
const locationCooldown = 30 * time.Second

// locationProbeBudget bounds a single reprobe triggered by a mid-request
// network error, so a switch attempt can't hang a command much longer than
// the original request would have. Derived from the probe timeouts (plus a
// small buffer) rather than hard-coded, so it can never silently truncate the
// final external probe when those timeouts change.
var locationProbeBudget = access.MaxProbeDuration() + time.Second

// statusAuthFailureOlares459 is the non-standard status code Olares' edge
// stack (Authelia ext-authz wired through l4-bfl-proxy) returns when an
// otherwise-valid request fails authentication — typically because the
// X-Authorization JWT has expired or 2FA needs re-arming. The body looks
// like `{"fa2":false,"method":"...","session_id":"<edge-jwt>","target_url":"..."}`.
//
// The web app treats 459 as a refresh trigger, parallel to 401:
// `apps/packages/app/src/platform/platformAjaxSender.ts:89` maps it to
// `ErrorCode.TOKE_INVILID`. We mirror that here — without this, every
// expired-token request through the per-service hosts (files.<terminus>,
// dashboard.<terminus>, etc.) returns 459 verbatim and refreshingTransport
// never gets the chance to rotate the token. The 401 path is still kept
// for endpoints that DON'T sit behind the Authelia edge (notably
// /api/refresh on auth.<terminus>, which the SPA also special-cases).
const statusAuthFailureOlares459 = 459

// isAuthFailureStatus reports whether resp.StatusCode means "request was
// rejected because the caller's token is no longer accepted, please get
// a fresh one and retry". Centralised so the two callers (RoundTrip's
// reactive path and any future pre-validation) stay in sync.
func isAuthFailureStatus(status int) bool {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, statusAuthFailureOlares459:
		return true
	}
	return false
}

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

	// locationState is the shared mutable connection-method cell (current
	// Location + its base http.Transport). Both http.Clients hold the same
	// pointer so a network-error-triggered switch through one is immediately
	// visible to the other.
	locationStateOnce sync.Once
	locationState     *locationState

	clientOnce sync.Once
	client     *http.Client

	uploadClientOnce sync.Once
	uploadClient     *http.Client

	// backendVersion memoizes the detected Olares backend version (see
	// olares_version.go). backendVersionMu guards the cell because
	// RefreshOlaresBackendVersion may overwrite it after the initial
	// sync.Once resolution.
	backendVersionOnce sync.Once
	backendVersionMu   sync.Mutex
	backendVersion     *semver.Version
	backendVersionErr  error
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
		// Identity is always the currently-selected profile; there is no
		// per-invocation override flag (see cli/cmd/ctl/root.go for why).
		// `profile login` / `profile import` reach the provider directly
		// with an explicit --olares-id, bypassing this Factory path.
		rp, err := cred.Resolve(ctx, "")
		if err != nil {
			f.resolveErr = err
			return
		}
		f.maybeBackfillLocation(ctx, rp)
		f.resolved = rp
	})
	return f.resolved, f.resolveErr
}

// maybeBackfillLocation lazily probes + persists the network position for a
// pre-existing profile that predates the Location field (rp.Location empty /
// invalid). It is a no-op for env-resolved profiles (nothing local to write),
// for profiles with a pinned auth URL override, for already-known locations,
// and while a recent outage cooldown is in effect (use the external defaults
// and fail fast). On a successful probe it re-derives rp's URLs in place and
// writes the result to config.json best-effort. On ErrUnreachable it leaves
// the external defaults and does NOT persist, so the next command re-probes.
func (f *Factory) maybeBackfillLocation(ctx context.Context, rp *credential.ResolvedProfile) {
	if rp == nil || rp.Source != "default" || rp.AuthURLOverride != "" {
		return
	}
	if rp.Location.Valid() {
		return
	}
	if inLocationCooldown(rp.OlaresID, time.Now()) {
		return
	}
	id, err := olares.ParseID(rp.OlaresID)
	if err != nil {
		return
	}
	probeCtx, cancel := context.WithTimeout(ctx, locationProbeBudget)
	defer cancel()
	loc, err := access.ProbeLocation(probeCtx, id, rp.LocalURLPrefix, rp.InsecureSkipVerify)
	if err != nil {
		return
	}
	rp.ApplyLocation(loc)
	var cfg cliconfig.MultiProfileConfig
	_ = cfg.SetLocation(rp.OlaresID, string(loc), time.Now().Unix())
}

// inLocationCooldown reports whether the profile keyed by olaresID recorded an
// "every probe failed" result within the cooldown window ending at now.
func inLocationCooldown(olaresID string, now time.Time) bool {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return false
	}
	p := cfg.FindByOlaresID(olaresID)
	if p == nil || p.LocationUnreachableAt == 0 {
		return false
	}
	return now.Unix()-p.LocationUnreachableAt < int64(locationCooldown/time.Second)
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

// ValidAccessToken returns an access token fresh enough for an immediate
// request, refreshing via /api/refresh if the cached token is within
// preflightSkew of expiry. Used by callers that bypass HTTPClient's transport
// (e.g. the exec WebSocket handshake) and therefore can't rely on the reactive
// 401-refresh path. On a missing/non-JWT exp claim the token is returned as-is
// (a 401 at use time will surface the login CTA).
func (f *Factory) ValidAccessToken(ctx context.Context) (string, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return "", err
	}
	cell := f.sharedTokenCell(rp.AccessToken)
	tok := cell.snapshot()
	if tok == "" {
		tok = rp.AccessToken
	}
	expired, err := auth.IsExpired(tok, time.Now(), preflightSkew)
	if err != nil || !expired {
		return tok, nil
	}
	newAT, rerr := f.Refresher().Refresh(ctx, rp.OlaresID, rp.AuthURL, tok, rp.InsecureSkipVerify)
	if rerr != nil {
		return "", rerr
	}
	cell.update(newAT)
	return newAT, nil
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

// locationState is the shared, mutable connection-method cell: the current
// Location plus the base http.Transport configured for it. A network-error
// reprobe swaps both under mu.
//
// unreachableMarked tracks whether an outage cooldown stamp is believed to be
// persisted for this profile — either carried over from a previous run or set
// by markUnreachable this run. clearUnreachable flips it false on the first
// success after a mark (doing exactly one config write to lift the cooldown)
// and markUnreachable re-arms it, so back-to-back outage/recovery cycles in a
// long-lived process (e.g. a chunked upload) are each handled rather than
// collapsing to a single per-process clear.
type locationState struct {
	mu                sync.Mutex
	loc               olares.Location
	base              http.RoundTripper
	unreachableMarked atomic.Bool
}

// sharedLocationState lazily builds the process-wide locationState from rp's
// (possibly backfilled) Location. Both http.Clients share the pointer so a
// switch is visible to all in-flight requests.
func (f *Factory) sharedLocationState(rp *credential.ResolvedProfile) *locationState {
	f.locationStateOnce.Do(func() {
		loc := rp.Location
		if !loc.Valid() {
			loc = olares.LocationExternal
		}
		ls := &locationState{
			loc:  loc,
			base: access.Transport(loc, rp.InsecureSkipVerify),
		}
		// Seed the marker from disk so the first success this run clears a
		// cooldown stamp left behind by a previous (failed) run.
		ls.unreachableMarked.Store(hasUnreachableStamp(rp.OlaresID))
		f.locationState = ls
	})
	return f.locationState
}

// hasUnreachableStamp reports whether the profile keyed by olaresID currently
// has a persisted outage cooldown stamp. A read error is treated as "no
// stamp" (best-effort).
func hasUnreachableStamp(olaresID string) bool {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return false
	}
	p := cfg.FindByOlaresID(olaresID)
	return p != nil && p.LocationUnreachableAt != 0
}

// newRefreshingTransport builds a refreshingTransport bound to rp. Both
// HTTPClient and HTTPClientWithoutTimeout reuse the same Refresher, tokenCell
// and locationState so a refresh or location switch through one is immediately
// visible on the other.
func (f *Factory) newRefreshingTransport(rp *credential.ResolvedProfile) *refreshingTransport {
	id, _ := olares.ParseID(rp.OlaresID)
	return &refreshingTransport{
		id:                 id,
		olaresID:           rp.OlaresID,
		localPrefix:        rp.LocalURLPrefix,
		authURLOverride:    rp.AuthURLOverride,
		insecureSkipVerify: rp.InsecureSkipVerify,
		loc:                f.sharedLocationState(rp),
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
	id                 olares.ID
	olaresID           string
	localPrefix        string
	authURLOverride    string
	insecureSkipVerify bool
	loc                *locationState
	refresher          *credential.Refresher
	token              *tokenCell

	// now is the clock used by preflightRefresh's expiry check. nil =
	// time.Now (production); tests override it to drive the skew
	// window deterministically.
	now func() time.Time
}

// authURL derives the /api/refresh base for the current Location, honoring a
// pinned auth URL override when present.
func (t *refreshingTransport) authURL(loc olares.Location) string {
	if t.authURLOverride != "" {
		return t.authURLOverride
	}
	return t.id.Endpoints(loc, t.localPrefix).Auth
}

func (t *refreshingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	snap := t.token.snapshot()

	// Snapshot the current connection method (Location + base transport).
	t.loc.mu.Lock()
	curLoc := t.loc.loc
	base := t.loc.base
	t.loc.mu.Unlock()

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
		newAT, ok, err := t.preflightRefresh(req.Context(), snap, curLoc)
		if err != nil {
			return nil, err
		}
		if ok {
			snap = newAT
		}
	}

	resp, err := t.send(base, req, snap)
	if err != nil {
		// Transport-layer failure (DNS / connect / timeout). If it's the
		// kind a different connection method might fix AND the body is
		// replayable, re-probe the Location and retry exactly once against
		// the new method. TLS errors, caller-cancels and anything that
		// produced an HTTP response are NOT switchable.
		if canRetry(req) && access.IsSwitchableNetErr(err, req.Context()) {
			if newLoc, newBase, switched := t.ensureSwitched(req.Context(), curLoc); switched {
				retryReq, rerr := cloneWithBody(req)
				if rerr != nil {
					return nil, err
				}
				retryReq.URL = t.id.RebaseURL(retryReq.URL, newLoc, t.localPrefix)
				retryReq.Host = ""
				return t.send(newBase, retryReq, snap)
			}
			// No connection method worked (cooldown, or every probe failed):
			// surface the friendly, classified "unreachable" message while
			// keeping the raw transport error reachable via Unwrap.
			return nil, access.NewUnreachable(t.olaresID, err)
		}
		return nil, err
	}
	// A real HTTP response means the path is up; lift any outage cooldown.
	t.clearUnreachable()

	if !isAuthFailureStatus(resp.StatusCode) {
		return resp, nil
	}
	if !canRetry(req) {
		// Non-replayable body — pre-flight already had its chance
		// (and either rotated the token or decided not to). Either
		// way, body is consumed; surface the auth failure verbatim.
		return resp, nil
	}
	// Drain + close so the underlying connection can be reused.
	drainAndClose(resp)

	newAT, rerr := t.refresher.Refresh(req.Context(), t.olaresID, t.authURL(curLoc), snap, t.insecureSkipVerify, curLoc)
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
	return t.send(base, retryReq, newAT)
}

// ensureSwitched re-probes the Location after a switchable network error and,
// if a usable method is found, swaps the shared base transport to it. It
// returns (newLoc, newBase, true) when the caller should retry, or
// (failedLoc, currentBase, false) when no switch happened (cooldown, every
// probe failed, or the caller's context expired).
//
// Concurrency mirrors the Refresher's "lock then re-check": if another
// goroutine already switched away from failedLoc while we waited on the mutex,
// we adopt their result instead of probing again.
func (t *refreshingTransport) ensureSwitched(ctx context.Context, failedLoc olares.Location) (olares.Location, http.RoundTripper, bool) {
	// Fast path under the lock: a peer may have already switched, or we may be
	// in cooldown — both decided without paying for a (slow) probe.
	t.loc.mu.Lock()
	if t.loc.loc != failedLoc {
		loc, base := t.loc.loc, t.loc.base
		t.loc.mu.Unlock()
		return loc, base, true
	}
	if t.inCooldown() {
		base := t.loc.base
		t.loc.mu.Unlock()
		return failedLoc, base, false
	}
	t.loc.mu.Unlock()

	// Probe WITHOUT holding the lock, so concurrent requests can keep reading
	// the current connection method (RoundTrip's quick snapshot) instead of
	// blocking for the whole probe budget behind us.
	probeCtx, cancel := context.WithTimeout(ctx, locationProbeBudget)
	defer cancel()
	newLoc, probeErr := access.ProbeLocation(probeCtx, t.id, t.localPrefix, t.insecureSkipVerify)

	// Commit under the lock, re-checking for a switch that landed while we
	// probed (mirrors the Refresher's "lock then re-compare").
	t.loc.mu.Lock()
	if t.loc.loc != failedLoc {
		// A peer already switched away from failedLoc — adopt their result
		// rather than stomping it with ours.
		loc, base := t.loc.loc, t.loc.base
		t.loc.mu.Unlock()
		return loc, base, true
	}
	if probeErr != nil {
		base := t.loc.base
		t.loc.mu.Unlock()
		// Every method failed: stamp the cooldown so back-to-back commands
		// fail fast. Keep the last-known-good Location (don't downgrade it).
		// Done outside the lock so the disk/flock write doesn't stall readers.
		if access.IsUnreachable(probeErr) {
			t.markUnreachable()
		}
		return failedLoc, base, false
	}
	// newLoc == failedLoc means a transient blip at the same position; rebuild
	// the base (drop stale idle conns) and retry once. A different position is
	// a genuine switch — update the shared state and persist it.
	newBase := access.Transport(newLoc, t.insecureSkipVerify)
	t.loc.loc = newLoc
	t.loc.base = newBase
	switched := newLoc != failedLoc
	t.loc.mu.Unlock()
	if switched {
		t.persistLocation(newLoc)
	}
	return newLoc, newBase, true
}

// persistLocation best-effort writes a freshly switched-to Location to
// config.json (and clears any outage cooldown via SetLocation). SetLocation
// re-reads config under the config lock, so an empty receiver is fine.
func (t *refreshingTransport) persistLocation(loc olares.Location) {
	var cfg cliconfig.MultiProfileConfig
	_ = cfg.SetLocation(t.olaresID, string(loc), t.clock().Unix())
}

// markUnreachable best-effort stamps the outage cooldown after an
// every-probe-failed result, and arms the marker so the next success lifts it.
func (t *refreshingTransport) markUnreachable() {
	var cfg cliconfig.MultiProfileConfig
	if err := cfg.SetLocationUnreachable(t.olaresID, t.clock().Unix()); err != nil {
		return
	}
	t.loc.unreachableMarked.Store(true)
}

// clearUnreachable best-effort lifts the outage cooldown after a successful
// response. The CAS gate makes this a no-op (no disk read/write) unless an
// outage was actually marked since the last clear, so a chunked upload's
// steady-state successes don't touch config — while a later outage re-arms it
// so a subsequent recovery is still cleared.
func (t *refreshingTransport) clearUnreachable() {
	if !t.loc.unreachableMarked.CompareAndSwap(true, false) {
		return
	}
	var cfg cliconfig.MultiProfileConfig
	if err := cfg.ClearLocationUnreachable(t.olaresID); err != nil {
		// Re-arm: we didn't actually clear, so a future success should retry.
		t.loc.unreachableMarked.Store(true)
	}
}

// inCooldown reports whether this profile recorded an every-probe-failed
// result within the cooldown window. Caller must hold no assumptions about
// disk state; a read error is treated as "not in cooldown".
func (t *refreshingTransport) inCooldown() bool {
	return inLocationCooldown(t.olaresID, t.clock())
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
func (t *refreshingTransport) preflightRefresh(ctx context.Context, snap string, loc olares.Location) (string, bool, error) {
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
	newAT, rerr := t.refresher.Refresh(ctx, t.olaresID, t.authURL(loc), snap, t.insecureSkipVerify, loc)
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
func (t *refreshingTransport) send(base http.RoundTripper, req *http.Request, token string) (*http.Response, error) {
	if token == "" {
		return base.RoundTrip(req)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("X-Authorization", token)
	clone.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	clone.Header.Set("X-Unauth-Error", "Non-Redirect")
	return base.RoundTrip(clone)
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

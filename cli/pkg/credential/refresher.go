package credential

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/internal/lockfile"
	"github.com/beclab/Olares/cli/pkg/auth"
)

// Refresher is the runtime token-refresh primitive: given a (likely-expired)
// access_token observed by an HTTP caller, it returns a freshly-rotated one,
// performing /api/refresh at most once across all goroutines AND across all
// concurrent olares-cli processes.
//
// The lark-cli pattern we mirror has three layers of de-duplication:
//
//  1. **Process-wide mutex** (`mu`). Every goroutine in the same Refresher
//     instance serializes through this; the loser waits and reads whatever
//     the winner persisted.
//
//  2. **Compare-after-Get**. The first thing the locked goroutine does is
//     re-read StoredToken from the keychain. If `stored.AccessToken` is
//     already different from the caller's `currentAccessToken`, somebody
//     (this process or another one) refreshed while we were blocked on the
//     mutex — return the stored token without touching the network.
//
//  3. **On-disk flock** (cli/internal/lockfile). Cross-process serialization.
//     We acquire it AFTER the in-process mutex+compare so multiple
//     goroutines from the same process collapse to a single flock contender.
//     After acquiring flock, we re-read StoredToken AGAIN — the winning
//     process may have finished and released its flock between our two
//     reads.
//
// Only when all three checks agree the token is still stale do we POST
// /api/refresh. On 401/403 from refresh we stamp InvalidatedAt and return
// ErrTokenInvalidated; on transport errors we surface them verbatim so the
// caller can retry the whole command.
type Refresher struct {
	store auth.TokenStore
	mu    sync.Mutex
	now   func() time.Time

	// timeout bounds a single Refresh call's flock-acquire window so a
	// stuck peer process can't hang the CLI forever. A misbehaving peer
	// stuck inside /api/refresh longer than this gets bypassed and we'll
	// race them on the actual POST — that's strictly better than blocking
	// the caller indefinitely.
	flockTimeout time.Duration
}

// NewRefresher returns a production Refresher backed by the keychain token
// store.
func NewRefresher() *Refresher {
	return &Refresher{
		store:        auth.NewTokenStore(),
		now:          time.Now,
		flockTimeout: 30 * time.Second,
	}
}

// NewRefresherWith is the test seam — pass any TokenStore + clock.
func NewRefresherWith(store auth.TokenStore, now func() time.Time) *Refresher {
	return &Refresher{store: store, now: now, flockTimeout: 30 * time.Second}
}

// Refresh exchanges the stored refresh_token for a fresh access_token.
//
// `currentAccessToken` is what the caller observed in its (now-401)
// response; if the keychain holds a different access_token at any of the
// double-check points, we return the stored value without making a network
// call. This is what lets concurrent goroutines / processes collapse to a
// single /api/refresh.
//
// Errors:
//   - *ErrNotLoggedIn          — no token entry exists for olaresID
//   - *ErrTokenInvalidated     — a previous refresh failure (or this one)
//     marked the grant unusable; user must re-authenticate
//   - context.Canceled / .DeadlineExceeded — caller's ctx fired while we
//     were blocked on flock
//   - any other error          — transport / 5xx / decode failure (transient,
//     caller may retry the whole command)
func (r *Refresher) Refresh(
	ctx context.Context,
	olaresID, authURL, currentAccessToken string,
	insecureSkipVerify bool,
) (string, error) {
	if olaresID == "" {
		return "", errors.New("olaresID is required")
	}
	if authURL == "" {
		return "", errors.New("authURL is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// In-process double-check. If another goroutine already rotated the
	// token while we were waiting on r.mu, we're done — no flock, no POST.
	if newAT, ok, err := r.alreadyFresh(olaresID, currentAccessToken); err != nil || ok {
		return newAT, err
	}

	// Cross-process serialization. Bound the wait so a stuck peer can't
	// hang us indefinitely.
	lockPath, err := lockfile.RefreshLockPath(olaresID)
	if err != nil {
		return "", fmt.Errorf("derive refresh lock path: %w", err)
	}
	lockCtx, cancel := context.WithTimeout(ctx, r.flockTimeout)
	defer cancel()
	release, err := lockfile.Acquire(lockCtx, lockPath)
	if err != nil {
		// Bubble the original ctx error if that's why we failed (so a
		// caller-side cancel surfaces as context.Canceled, not as a
		// deadline-exceeded from the inner ctx).
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("acquire refresh lock for %s: %w", olaresID, err)
	}
	defer release() //nolint:errcheck // best-effort lock release

	// Cross-process double-check: another olares-cli may have refreshed
	// between our two reads. Re-read the store while holding flock so the
	// next read sees a stable answer.
	if newAT, ok, err := r.alreadyFresh(olaresID, currentAccessToken); err != nil || ok {
		return newAT, err
	}

	// Still stale → actually call /api/refresh.
	stored, err := r.store.Get(olaresID)
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return "", &ErrNotLoggedIn{OlaresID: olaresID}
		}
		return "", fmt.Errorf("read token store: %w", err)
	}
	if stored.RefreshToken == "" {
		// We have an access_token but no refresh_token to rotate it
		// with. Treat as "needs login" — same UX as if the entry was
		// missing, since we can't recover from this client-side.
		return "", &ErrNotLoggedIn{OlaresID: olaresID}
	}

	tok, err := auth.Refresh(ctx, auth.RefreshRequest{
		AuthURL:            authURL,
		RefreshToken:       stored.RefreshToken,
		AccessToken:        stored.AccessToken,
		InsecureSkipVerify: insecureSkipVerify,
		Timeout:            15 * time.Second,
	})
	if err != nil {
		// 401/403 from /api/refresh = the grant itself is dead. Mark
		// it so subsequent commands skip the network round-trip and
		// go straight to "run profile login".
		if errors.Is(err, auth.ErrRefreshUnauthorized) {
			at := r.now()
			// MarkInvalidated is best-effort: failing to persist
			// the stamp doesn't change the user-facing CTA
			// ("run profile login"), and the next refresh attempt
			// will hit the same 401 → re-mark, so the situation
			// is self-healing. Surface the persistence failure on
			// stderr for ops/debugging but DO NOT swallow the
			// typed error — that would deny callers'
			// reformatters the chance to render the CTA.
			if mErr := r.store.MarkInvalidated(olaresID, at); mErr != nil {
				fmt.Fprintf(os.Stderr,
					"warning: failed to persist invalidated marker for %s: %v (refresh attempt will repeat next run)\n",
					olaresID, mErr)
			}
			return "", &ErrTokenInvalidated{OlaresID: olaresID, InvalidatedAt: at}
		}
		return "", err
	}

	newStored := auth.StoredToken{
		OlaresID:     olaresID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		SessionID:    tok.SessionID,
		GrantedAt:    r.now().UnixMilli(),
	}
	if err := r.store.Set(newStored); err != nil {
		return "", fmt.Errorf("persist refreshed token: %w", err)
	}
	return tok.AccessToken, nil
}

// alreadyFresh returns (newAT, true, nil) if the keychain holds an
// access_token that differs from currentAccessToken (somebody already
// refreshed) and the entry is not invalidated. Returns (_, false, nil) when
// the caller should proceed to the next step. Returns a typed error for
// permanent conditions (not-logged-in, invalidated).
func (r *Refresher) alreadyFresh(olaresID, currentAccessToken string) (string, bool, error) {
	stored, err := r.store.Get(olaresID)
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return "", false, &ErrNotLoggedIn{OlaresID: olaresID}
		}
		return "", false, fmt.Errorf("read token store: %w", err)
	}
	if stored.InvalidatedAt > 0 {
		return "", false, &ErrTokenInvalidated{
			OlaresID:      olaresID,
			InvalidatedAt: time.UnixMilli(stored.InvalidatedAt),
		}
	}
	if stored.AccessToken != "" && stored.AccessToken != currentAccessToken {
		return stored.AccessToken, true, nil
	}
	return "", false, nil
}

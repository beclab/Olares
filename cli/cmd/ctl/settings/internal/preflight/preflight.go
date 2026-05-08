// Package preflight is the soft role-gate adapter shared by every
// `olares-cli settings` area. Each area's RunE calls Gate at the top to
// short-circuit on a stale-but-clearly-low cached role, and Wrap at the
// bottom to translate a server-side 403 / 401 into the canonical
// "refresh + retry" hint.
//
// The role floor for every verb tracks the SPA's `useAdminStore` rules
// (apps/.../stores/settings/admin.ts):
//
//	normal — every authenticated user
//	admin  — `useAdminStore.isAdmin` (admin OR owner)
//	owner  — `useAdminStore.isOwner` (owner only)
//
// Verbs that are visible to every authenticated user pass an empty
// required role (or skip Gate entirely); admin-floor verbs pass
// whoami.RoleAdmin; the rare owner-only verbs pass whoami.RoleOwner.
//
// Gate never reaches out to the network — it consults the cached role
// recorded by `profile login` / `profile whoami` in
// ~/.olares-cli/config.json. Empty / unknown cache always passes
// through, so first-run profiles created before the role cache
// existed keep working without a forced re-login. The server stays
// authoritative; Wrap is the second half that catches drift the cache
// hasn't seen yet.
package preflight

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Gate short-circuits a verb when the cached role is provably below
// `required`. It looks up the active profile via the factory (cheap;
// memoized after the first call) and the cached role via cliconfig.
//
// On any local failure (factory not wired, config unreadable, profile
// not found) it falls through silently — the server is the only
// authoritative source for "can this user do this", so a flaky cache
// must never block a legitimate API call.
//
// `verbDescr` is the human-readable verb name interpolated into the
// resulting error if the gate trips. Keep it lowercase, present-tense,
// no trailing punctuation (e.g. "list users", "set FRP server").
func Gate(ctx context.Context, f *cmdutil.Factory, required, verbDescr string) error {
	if f == nil || required == "" {
		return nil
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil || rp == nil {
		return nil
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil || cfg == nil {
		return nil
	}
	return whoami.PreflightRole(cfg, rp.OlaresID, required, verbDescr)
}

// Wrap turns a server-side 403 / 401 into the canonical "refresh +
// retry" hint, tagged with the verb the call site was attempting.
// Non-permission errors and nil pass through unchanged.
//
// Designed to be called as the last expression of a RunE:
//
//	return preflight.Wrap(ctx, f, doMutateEnvelope(...), "rename device")
func Wrap(ctx context.Context, f *cmdutil.Factory, err error, verbDescr string) error {
	if err == nil {
		return nil
	}
	var olaresID string
	if f != nil {
		if rp, rerr := f.ResolveProfile(ctx); rerr == nil && rp != nil {
			olaresID = rp.OlaresID
		}
	}
	return whoami.WrapPermissionErr(err, olaresID, verbDescr)
}

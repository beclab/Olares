package profile

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// eagerWhoami runs a best-effort GET /api/backend/v1/user-info right after
// `profile login` / `profile import` has persisted the new credentials. It
// populates ProfileConfig.OwnerRole / WhoamiRefreshedAt so subsequent
// `settings` preflight checks have something to compare against without a
// follow-up `profile whoami --refresh`.
//
// "Best-effort" means: any failure is downgraded to a one-line stderr
// warning and the function returns nil. Login / import succeeded — the
// user shouldn't have to debug a transient backend hiccup just because
// the role pre-fetch didn't land. The next time they run `profile whoami`
// or any `settings` verb, the cache will populate naturally.
//
// Why we don't reuse cmdutil.Factory.HTTPClient: that http.Client memoizes
// the access token from the FIRST ResolveProfile call in the process. In
// login / import we want to talk to the backend with the JUST-MINTED
// token, not whatever Factory previously cached (which may be tied to a
// different active profile when --no-switch was passed). NewHTTPClientWithToken
// builds a fresh client and injects the new token explicitly, sidestepping
// the issue entirely.
func eagerWhoami(
	ctx context.Context,
	cfg *cliconfig.MultiProfileConfig,
	profile cliconfig.ProfileConfig,
	accessToken string,
) {
	if accessToken == "" {
		return
	}
	id, err := olares.ParseID(profile.OlaresID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: skipped post-login role fetch: %v\n", err)
		return
	}

	desktopURL := id.DesktopURL(profile.LocalURLPrefix)
	client := whoami.NewHTTPClientWithToken(desktopURL, profile.OlaresID, accessToken, profile.InsecureSkipVerify)

	// Tight ceiling: this is a "while you wait" hop. If the backend is
	// slow we don't want to delay the success message any longer than we
	// have to — the user has already authenticated, the role can populate
	// later via `whoami --refresh`.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := whoami.FetchAndCache(ctx, client, cfg, profile.OlaresID, time.Now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: post-login role fetch failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "         (run `olares-cli profile whoami --refresh` later to populate the role cache)")
		return
	}
	if res != nil && res.Info.OwnerRole != "" {
		fmt.Printf("role: %s\n", whoami.FriendlyLabel(res.Info.OwnerRole))
	}
}

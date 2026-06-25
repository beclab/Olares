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

// eagerDetectTimeout bounds the post-login/import detect. The network position
// is already known here (no probe), so this only needs to cover the role +
// backend-version round-trips; keep it short so a slow backend can't stall the
// otherwise-finished login for long.
const eagerDetectTimeout = 8 * time.Second

// eagerDetect runs the unified detect (role + backend version) right after
// `profile login` / `profile import` has persisted the new credentials, and
// prints a one-line-per-fact summary. The network position was already probed
// during login/import and stored on profile.Location, so it is reused here
// (no second probe) — the role/version fetches go through that detected
// connection method rather than a hard-wired public URL.
//
// Best-effort: any failure downgrades to a stderr warning and returns.
// Login/import already succeeded — a transient backend hiccup must not shadow
// it; the caches populate on the next version/role-aware command (or
// `profile whoami --refresh`).
//
// Why NewHTTPClientWithToken (inside DetectAndCache) rather than the Factory
// http.Client: the latter memoizes the access token from the first
// ResolveProfile call, which under --no-switch may be tied to a different
// active profile. The detect path injects the just-minted token explicitly.
func eagerDetect(
	ctx context.Context,
	cfg *cliconfig.MultiProfileConfig,
	profile cliconfig.ProfileConfig,
	accessToken string,
) {
	if accessToken == "" || cfg == nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, eagerDetectTimeout)
	defer cancel()

	d, err := whoami.DetectAndCache(ctx, whoami.DetectInput{
		Cfg:           cfg,
		OlaresID:      profile.OlaresID,
		LocalPrefix:   profile.LocalURLPrefix,
		Insecure:      profile.InsecureSkipVerify,
		AccessToken:   accessToken,
		KnownLocation: olares.Location(profile.Location),
		Now:           time.Now,
	})
	if d != nil {
		if d.Location != "" {
			fmt.Printf("location: %s\n", d.Location)
		}
		if d.RoleLabel != "" {
			fmt.Printf("role: %s\n", d.RoleLabel)
		}
		if d.BackendVersion != "" {
			fmt.Printf("version: %s\n", d.BackendVersion)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: post-login detect did not fully complete: %v\n", err)
		fmt.Fprintln(os.Stderr, "         (run `olares-cli profile whoami --refresh` later to populate the cache)")
	}
}

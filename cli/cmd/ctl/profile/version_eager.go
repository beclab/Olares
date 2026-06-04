package profile

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// eagerBackendVersion runs a best-effort GET /api/olares-info right after
// `profile login` / `profile import` has persisted the new credentials, and
// caches the backend osVersion into ProfileConfig.BackendVersion. This is the
// version-cache analogue of eagerWhoami's role pre-fetch: it makes the
// version-aware help / subcommand visibility (e.g. `settings gpu`, `settings
// network overlay`) accurate from the very first command after login, rather
// than only after some version-touching command happens to run.
//
// Best-effort means: any failure downgrades to a one-line stderr warning and
// returns. Login/import already succeeded — a transient backend hiccup must
// not shadow it; the cache will populate naturally on the next version-aware
// command via the TTL path.
//
// We can't piggyback on eagerWhoami: /api/backend/v1/user-info returns only
// {name, owner_role}, not osVersion — the version lives on the separate
// /api/olares-info endpoint. Like eagerWhoami we use NewHTTPClientWithToken
// (the just-minted token) rather than cmdutil.Factory.HTTPClient, which
// memoizes the token from the first ResolveProfile and may be tied to a
// different active profile under --no-switch.
func eagerBackendVersion(
	ctx context.Context,
	cfg *cliconfig.MultiProfileConfig,
	profile cliconfig.ProfileConfig,
	accessToken string,
) {
	if accessToken == "" || cfg == nil {
		return
	}
	id, err := olares.ParseID(profile.OlaresID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: skipped post-login version fetch: %v\n", err)
		return
	}
	desktopURL := id.DesktopURL(profile.LocalURLPrefix)
	client := whoami.NewHTTPClientWithToken(desktopURL, profile.OlaresID, accessToken, profile.InsecureSkipVerify)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var env struct {
		Data struct {
			OsVersion string `json:"osVersion"`
		} `json:"data"`
	}
	if err := client.DoJSON(ctx, "GET", "/api/olares-info", nil, &env); err != nil {
		fmt.Fprintf(os.Stderr, "warning: post-login version fetch failed: %v\n", err)
		return
	}
	osVersion := strings.TrimSpace(env.Data.OsVersion)
	if osVersion == "" {
		return
	}
	v, err := utils.ParseOlaresVersionString(osVersion)
	if err != nil {
		return
	}
	if _, err := cfg.SetBackendVersion(profile.OlaresID, v.Original(), time.Now().Unix()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not cache Olares backend version: %v\n", err)
	}
}

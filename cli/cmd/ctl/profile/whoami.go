package profile

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// NewWhoamiCommand: `olares-cli profile whoami [--refresh] [-o table|json]`
//
// Reports the active profile's identity + role, defaulting to the locally
// cached value (config.json's ownerRole / whoamiRefreshedAt fields). Use
// --refresh to force a server roundtrip against /api/backend/v1/user-info
// — the same endpoint the SPA's admin store hits in
// apps/packages/app/src/stores/settings/admin.ts (`get_user_info`).
//
// This command intentionally has two aliases under the settings tree:
//   - `olares-cli settings users me`   (canonical "Settings -> Users -> me")
//   - `olares-cli settings me whoami`  (canonical "Settings -> Person ->
//                                        whoami")
//
// All three call into the same pkg/whoami.Run driver so behavior, output
// shapes, and cache-write semantics stay identical no matter which surface
// the user reaches for.
//
// Why --refresh (rather than always hitting the server): every settings
// subcommand performs a soft preflight using OwnerRole — making the cache
// the cheap default keeps `whoami` a single round-trip per session in the
// common case, and lets users still reconcile after a role change with a
// single keystroke.
//
// Output:
//   - table (default): two human lines — identity + freshness, plus a
//     "role changed" notice when --refresh detected drift.
//   - json: pkg/whoami.Display verbatim, so scripts can branch on
//     {"role":"owner"} without parsing prose.
func NewWhoamiCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		refresh   bool
		outputRaw string
	)
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "show the current profile's identity and role",
		Long: `Show the active profile's olaresId, friendly name, and role
("Owner" / "Admin" / "User") on the target Olares instance.

Defaults to the locally cached role written by ` + "`profile login`" + ` /
` + "`profile import`" + ` / a previous ` + "`whoami --refresh`" + `. Pass --refresh to
force a fresh GET against /api/backend/v1/user-info and update the
cache; if the role changed since the last refresh you'll see a
"role changed: X -> Y" notice.

Aliases:
  olares-cli settings users me
  olares-cli settings me whoami
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runWhoami(c.Context(), f, refresh, outputRaw)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force a fresh /api/backend/v1/user-info roundtrip and update the cached role")
	cmd.Flags().StringVarP(&outputRaw, "output", "o", "table", "output format: table, json")
	return cmd
}

// runWhoami is the cobra-side glue: it resolves the active profile + http
// client, parses the --output flag, then delegates to pkg/whoami.Run for
// the actual cache/server policy and rendering.
func runWhoami(ctx context.Context, f *cmdutil.Factory, refresh bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return fmt.Errorf("internal error: profile whoami not wired with cmdutil.Factory")
	}

	format, err := whoami.ParseOutput(outputRaw)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}

	client := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)
	return whoami.Run(ctx, client, cfg, rp.OlaresID, refresh, format, nil, os.Stdout)
}

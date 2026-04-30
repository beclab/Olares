package apps

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps auth-level set <app> <entrance> --level X`.
//
// Sets the authorization_level for a single entrance. Mirrors the
// SPA's stores/settings/application.ts setupAuthLevel() and the
// "Authorization" panel on the per-app entrance page.
//
// Wire shape (BFL-proxied; no user-service controller):
//
//   POST /api/applications/<app>/<entrance>/setup/auth-level
//        body: { authorization_level: "private"|"public"|"internal" }
//
// The auth-level is also surfaced as a column in `apps entrances list`
// output (read-side); we don't expose a dedicated `auth-level get`
// because the read is already covered there. Restricting the CLI to a
// single `set` mirrors the upstream API which is POST-only too.
//
// Role: per-app config write; the SPA gates on isAdmin. We rely on
// server-side preflight (a normal user gets a 403 with the usual hint).

// validAuthLevels covers AUTH_LEVEL from constant/index.ts:314-318.
var validAuthLevels = map[string]struct{}{
	"private":  {},
	"public":   {},
	"internal": {},
}

// NewAuthLevelCommand returns the `settings apps auth-level` parent.
// Single-verb but we still nest it under a parent so the help text
// has somewhere to land — and so a future `auth-level get` (if the
// upstream ever exposes one) has a natural home.
func NewAuthLevelCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth-level",
		Short: "per-entrance authorization level (Settings -> App -> Entrance -> Authorization)",
		Long: `Set the authorization level for a single entrance.

Subcommands:
  set <app> <entrance> --level private|public|internal

Levels:
  private   only the app's owner can reach the entrance
  public    any authenticated user can reach the entrance
  internal  only intra-cluster traffic can reach the entrance

The current level for each entrance is shown in the AUTH LEVEL column
of "olares-cli settings apps entrances list" — there is no separate
get verb because no GET endpoint exists upstream.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAuthLevelSetCommand(f))
	return cmd
}

func newAuthLevelSetCommand(f *cmdutil.Factory) *cobra.Command {
	var level string
	cmd := &cobra.Command{
		Use:   "set <app> <entrance>",
		Short: "set the per-entrance authorization level",
		Long: `Set the authorization level for an entrance.

--level values:
  private   only the app's owner can reach the entrance
  public    any authenticated user can reach the entrance
  internal  only intra-cluster traffic can reach the entrance

The change is enforced by the upstream proxy on the next request;
existing authenticated sessions are NOT terminated retroactively.

Examples:
  olares-cli settings apps auth-level set files file --level private
  olares-cli settings apps auth-level set files file --level public
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runAuthLevelSet(c.Context(), f, args[0], args[1], level)
		},
	}
	cmd.Flags().StringVar(&level, "level", "", "authorization level: private | public | internal")
	_ = cmd.MarkFlagRequired("level")
	return cmd
}

func runAuthLevelSet(ctx context.Context, f *cmdutil.Factory, app, entrance, level string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	return runAuthLevelSetWithDoer(ctx, pc.doer, app, entrance, level)
}

// runAuthLevelSetWithDoer is the wire-level core of `apps auth-level set`.
// Split out so tests can drive the validation + body shape directly
// through a fakeDoer without the cmdutil.Factory dependency.
func runAuthLevelSetWithDoer(ctx context.Context, d Doer, app, entrance, level string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app, entrance = strings.TrimSpace(app), strings.TrimSpace(entrance)
	if app == "" || entrance == "" {
		return fmt.Errorf("both <app> and <entrance> are required")
	}
	level = strings.ToLower(strings.TrimSpace(level))
	if _, ok := validAuthLevels[level]; !ok {
		return fmt.Errorf("--level %q is not one of private|public|internal", level)
	}
	body := map[string]string{"authorization_level": level}
	path := "/api/applications/" + url.PathEscape(app) + "/" + url.PathEscape(entrance) + "/setup/auth-level"
	if err := doMutateEnvelope(ctx, d, "POST", path, body, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "set auth level for %s/%s to %q\n", app, entrance, level)
	return nil
}

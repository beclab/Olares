package appearance

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/userenv"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance theme set ...`
//
// The theme is stored in the OLARES_USER_THEME UserEnv, which
// build/user-env.yaml declares the single source of truth: BFL's
// config-system endpoint and the env settings page both read and write
// it, and app-service injects it into apps. We write it through
// /api/env/userenvs — the same vector `settings advanced env user`
// drives — because that is the only route to it that does not need a
// JWS-signed activation payload.
//
// The Settings SPA's own light/dark toggle only writes a browser cookie
// and never calls the backend, so a value set here is what apps and the
// login flow observe, not what an already-open browser session looks
// like. That asymmetry is a frontend gap, not a CLI limitation.
//
// Role: Appearance is in the normal-user menu (admin.ts:101-103), and
// a UserEnv is per-user. No PreflightRole check.

// themeValues mirrors the options declared for OLARES_USER_THEME in
// build/user-env.yaml. app-service validates the value against that same
// declaration (EnvVarSpec.ValidateValue -> "value not in options"), so
// this check only moves the rejection earlier — there is no route that
// writes a value the declaration does not list.
var themeValues = []string{"light", "dark"}

func NewThemeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme",
		Short: "system theme preference",
		Long: `Update the system theme preference (Settings -> Appearance > Theme).

The current value is shown by "settings appearance get".

Subcommands:
  set <light|dark>      update the theme preference
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newThemeSetCommand(f))
	return cmd
}

func newThemeSetCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <light|dark>",
		Short: "update the system theme preference",
		Long: `Update the system theme preference.

The value is written to the OLARES_USER_THEME user environment variable,
which the platform treats as the source of truth for the theme: apps
receive it, and the login flow honors it.

An Olares desktop already open in a browser will not change appearance:
its light/dark state comes from a cookie the SPA sets locally and never
sends to the backend.

Allowed values:
  light
  dark

The platform declares these two and rejects anything else, so a value a
newer release adds needs a newer CLI rather than a flag to bypass this.

Examples:
  olares-cli settings appearance theme set dark
  olares-cli settings appearance theme set light
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			value, err := resolveThemeValue(args[0])
			if err != nil {
				return err
			}
			return runThemeSet(c.Context(), f, value)
		},
	}
}

func resolveThemeValue(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("a theme value is required (allowed: %s)", strings.Join(themeValues, ", "))
	}
	for _, v := range themeValues {
		if value == v {
			return value, nil
		}
	}
	return "", fmt.Errorf("unsupported theme %q (allowed: %s)", raw, strings.Join(themeValues, ", "))
}

func runThemeSet(ctx context.Context, f *cmdutil.Factory, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireAppearanceBackendVersion(ctx, f,
		"settings appearance theme set", "the "+userenv.ThemeEnvName+" user env"); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	updates := map[string]string{userenv.ThemeEnvName: value}
	if err := userenv.SetValues(ctx, pc.doer, userenv.UserEnvsPath, updates); err != nil {
		return err
	}
	fmt.Printf("System theme updated to %q.\n", value)
	return nil
}

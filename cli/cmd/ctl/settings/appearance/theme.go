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
// NOT REGISTERED in root.go, and deliberately kept: the live theme is a
// browser cookie (theme_name, written and read only by the SPA, never
// sent upstream), which a CLI cannot set. The OLARES_USER_THEME UserEnv
// this writes has no shipped reader — nothing under apps/ references it,
// the greeter reads only the login wallpaper, and no app manifest binds
// it through valueFrom — so exposing the verb would report success on a
// write nobody observes. Re-register it once that UserEnv drives the UI.
//
// The write itself goes through /api/env/userenvs — the same vector
// `settings advanced env user` drives — because that is the only route
// to it that does not need a JWS-signed activation payload.
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

The value is written to the OLARES_USER_THEME user environment variable.

No Olares interface reads it yet: a desktop's light/dark state comes from
a cookie the SPA sets locally and never sends to the backend, so this
does not change how anything looks.

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

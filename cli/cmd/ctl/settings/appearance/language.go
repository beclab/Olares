package appearance

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance language ...`
//
// Backed by user-service's POST /api/wallpaper/update/language, which
// forwards to /bfl/settings/v1alpha1/config-system/language.
//
// SPA reference: apps/packages/app/src/stores/settings/background.ts
//   updateLanguage(language) -> axios.post('/api/wallpaper/update/language',
//                                          { language })
//
// We mirror the SPA's supportLanguages whitelist client-side
// (apps/.../i18n/index.ts:12 — currently `en-US` and `zh-CN`) because
// neither user-service nor BFL validate the value today: an unknown
// code would land in the config-system CRD verbatim and the SPA's
// i18n loader would silently fall back to defaultLanguage on the next
// session — i.e. the call "succeeds" but nothing changes. The CLI
// catches that early so callers get a real error instead of a
// false-positive.
//
// `--force` is the escape hatch for the unusual case where the SPA
// has shipped a new locale before this CLI build catches up. It
// bypasses the local whitelist and PUTs the value through; the
// upstream remains the final authority.
//
// Role: Appearance is in the normal-user menu (admin.ts:101-103). No
// PreflightRole check.

func NewLanguageCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "language",
		Short: "system language preference",
		Long: `Read or update the system language preference (Settings ->
Appearance > Language).

The current value can be inspected via "settings appearance get".

Subcommands:
  set <locale>          update the system language
                        (e.g. "set en-US"; --value also accepted)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newLanguageSetCommand(f))
	return cmd
}

// supportedLocales is the client-side whitelist mirrored from the
// SPA's apps/.../i18n/index.ts supportLanguages list. Keep in sync
// when SPA adds a locale; until then, callers can use --force to
// bypass the check.
var supportedLocales = []string{"en-US", "zh-CN"}

// newLanguageSetCommand registers `appearance language set [<locale>]`.
//
// Argument shape: a positional <locale> is the canonical form (matches
// the SKILL doc + SPA copy "language set en-US"); --value is kept as a
// strict-flag alternative for users who prefer flag-only invocations.
// Exactly one of the two MUST be supplied; if both are passed and
// disagree we error out rather than silently picking one. The
// positional shape mirrors every other "verb <obj>" command in this
// tree, which all take their primary subject positionally.
func newLanguageSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		value string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "set <locale>",
		Short: "update the system language preference (e.g. set en-US)",
		Long: `Update the system language preference. The value is a locale code
the SPA's language picker emits.

Allowed locales (matches SPA's i18n bundle list):
  en-US
  zh-CN

The locale can be passed as a positional argument or as --value. Pass
exactly one of the two; passing both with conflicting values is an
error.

Pass --force to bypass the client-side whitelist. Use only when the
SPA has shipped a new locale ahead of this CLI build — the upstream
will accept any string today, so a typo with --force will silently
land in the config-system CRD and the SPA will fall back to the
default locale on the next session.

Examples:
  olares-cli settings appearance language set en-US
  olares-cli settings appearance language set zh-CN
  olares-cli settings appearance language set --value en-US
  olares-cli settings appearance language set ja-JP --force
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			resolved, err := resolveLanguageValue(args, value)
			if err != nil {
				return err
			}
			if err := validateLocale(resolved, force); err != nil {
				return err
			}
			return runLanguageSet(c.Context(), f, resolved)
		},
	}
	cmd.Flags().StringVar(&value, "value", "", "locale code to set (e.g. en-US, zh-CN); same as the positional <locale>")
	cmd.Flags().BoolVar(&force, "force", false, "bypass the client-side whitelist (use only when the SPA has shipped a new locale ahead of this CLI build)")
	return cmd
}

// validateLocale enforces the client-side supportedLocales whitelist.
// `force=true` short-circuits to nil so callers can opt into writing
// a locale the CLI doesn't yet know about. The trim is symmetric with
// resolveLanguageValue's trim so callers that bypass that helper still
// get the same forgiving behavior.
func validateLocale(value string, force bool) error {
	value = strings.TrimSpace(value)
	if force {
		return nil
	}
	for _, l := range supportedLocales {
		if value == l {
			return nil
		}
	}
	return fmt.Errorf("unsupported locale %q (allowed: %s; pass --force to bypass for forward compatibility)",
		value, strings.Join(supportedLocales, ", "))
}

// resolveLanguageValue picks the locale from <args> or --value. Empty
// after Trim is treated as "not supplied"; conflicting non-empty values
// are an explicit error so a stale --value alias can't silently
// override a positional intent (or vice versa).
func resolveLanguageValue(args []string, flagValue string) (string, error) {
	pos := ""
	if len(args) == 1 {
		pos = strings.TrimSpace(args[0])
	}
	flag := strings.TrimSpace(flagValue)
	switch {
	case pos == "" && flag == "":
		return "", fmt.Errorf("a locale code is required (e.g. \"set en-US\" or --value en-US)")
	case pos != "" && flag != "" && pos != flag:
		return "", fmt.Errorf("conflicting locale: positional %q vs --value %q (pass only one)", pos, flag)
	case pos != "":
		return pos, nil
	default:
		return flag, nil
	}
}

func runLanguageSet(ctx context.Context, f *cmdutil.Factory, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("a locale code is required (e.g. \"set en-US\" or --value en-US)")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	body := map[string]string{"language": value}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/wallpaper/update/language", body, nil); err != nil {
		return err
	}
	fmt.Printf("System language updated to %q.\n", value)
	return nil
}

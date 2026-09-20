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
// Termipass writes the same field from the same route
// (packages/app/src/stores/settings/background.ts requestUpdateLanguage),
// so its picker is the reference for what this endpoint accepts — see
// supportedLocales below.
//
// The two SPAs reading this field do not carry the same bundles: the
// Olares desktop ships en-US and zh-CN only, so one of the locales
// added in 1.12.7 renders in LarePass while the browser desktop falls
// back to English.
//
// `--force` is the escape hatch for the unusual case where a SPA has
// shipped a locale before this CLI build catches up. It bypasses both
// the whitelist and the version requirement; the upstream remains the
// final authority.
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
                        (` + localeSummary() + `)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newLanguageSetCommand(f))
	return cmd
}

// localeOption is a locale code with the name the picker shows for it,
// and the Olares release whose SPA carries its translation.
type localeOption struct {
	code  string
	label string
	since string
}

// localeMinVersionExtended is the release that added five locales.
// Settings gates its own picker on the same boundary, treating any
// 1.12.7 prerelease as in (Termipass ActiveOlaresReminderDialog.vue:
// compareOlaresVersion(version, "1.12.7-0") >= 0), which is also what
// OlaresBackendAtLeast does with a prerelease.
const localeMinVersionExtended = "1.12.7"

// supportedLocales mirrors the SPA's supportLanguages picker
// (Termipass packages/app/src/i18n/index.ts). It stays a client-side
// whitelist because neither user-service nor BFL validate the value:
// an unknown code lands in the config-system CRD verbatim and the SPA
// falls back to its default on the next session, so the call
// "succeeds" while nothing changes.
//
// Add a locale here when the SPA ships its bundle, with the release
// that ships it; until then callers can use --force.
var supportedLocales = []localeOption{
	{code: "en-US", label: "English"},
	{code: "zh-CN", label: "简体中文"},
	{code: "de-DE", label: "Deutsch", since: localeMinVersionExtended},
	{code: "es-ES", label: "Español", since: localeMinVersionExtended},
	{code: "it-IT", label: "Italiano", since: localeMinVersionExtended},
	{code: "fr-FR", label: "Français", since: localeMinVersionExtended},
	{code: "ja-JP", label: "日本語", since: localeMinVersionExtended},
}

// localeSummary is the one-line form for the parent help, kept in step
// with supportedLocales so adding a locale cannot leave a stale count
// behind. It assumes the gated ones share a single release, which is
// what the list holds today.
func localeSummary() string {
	var always []string
	gated := 0
	for _, l := range supportedLocales {
		if l.since == "" {
			always = append(always, l.code)
			continue
		}
		gated++
	}
	summary := strings.Join(always, ", ")
	if gated > 0 {
		summary += fmt.Sprintf(", and %d more on Olares >= %s", gated, localeMinVersionExtended)
	}
	return summary
}

// localeChoices lists the locales the way the picker labels them,
// grouped so the ones needing a newer Olares are named as such. CJK
// labels make column alignment unreliable, hence the inline shape.
func localeChoices() string {
	var always, gated []string
	for _, l := range supportedLocales {
		entry := fmt.Sprintf("%s (%s)", l.code, l.label)
		if l.since == "" {
			always = append(always, entry)
			continue
		}
		gated = append(gated, entry)
	}
	out := "  " + strings.Join(always, ", ")
	if len(gated) > 0 {
		out += fmt.Sprintf("\n  Olares >= %s also has: %s", localeMinVersionExtended, strings.Join(gated, ", "))
	}
	return out
}

// describeLocale is how a stored code reads in the table: the code the
// page shows, plus the name it shows beside it. A code this CLI does
// not carry — one --force wrote, or one a newer SPA added — is printed
// as stored rather than guessed at.
func describeLocale(code string) string {
	code = strings.TrimSpace(code)
	for _, l := range supportedLocales {
		if strings.EqualFold(code, l.code) {
			return fmt.Sprintf("%s (%s)", l.code, l.label)
		}
	}
	return nonEmpty(code)
}

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
		Long: fmt.Sprintf(`Update the system language preference. The value is a locale code
the SPA's language picker emits.

Allowed locales (matches the SPA's i18n bundle list):
%s

Case does not matter: "en-us" sets en-US.

The locale can be passed as a positional argument or as --value. Pass
exactly one of the two; passing both with conflicting values is an
error.

Pass --force to bypass both the list and the version requirement. Use
only when the SPA has shipped a locale ahead of this CLI build — the
upstream will accept any string today, so a forced typo lands in the
config-system CRD and the SPA falls back to its default locale on the
next session, i.e. the write "succeeds" and nothing changes.

Examples:
  olares-cli settings appearance language set en-US
  olares-cli settings appearance language set zh-CN
  olares-cli settings appearance language set de-DE
  olares-cli settings appearance language set --value en-US
  olares-cli settings appearance language set ko-KR --force
`, localeChoices()),
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			resolved, err := resolveLanguageValue(args, value)
			if err != nil {
				return err
			}
			canonical, err := resolveLocale(c.Context(), f, resolved, force)
			if err != nil {
				return err
			}
			return runLanguageSet(c.Context(), f, canonical)
		},
	}
	cmd.Flags().StringVar(&value, "value", "", "locale code to set (e.g. en-US); same as the positional <locale>")
	cmd.Flags().BoolVar(&force, "force", false, "bypass the whitelist and the version requirement (use only when a SPA has shipped a locale ahead of this CLI build)")
	return cmd
}

// resolveLocale enforces the client-side whitelist and returns the
// canonical spelling, because the SPA's i18n bundle is keyed by it:
// "en-us" has to be stored as "en-US" to load anything. Matching
// case-insensitively is what this subtree does with surfaces, fill modes
// and date formats.
//
// The backend version is only consulted for a locale that needs one, so
// setting en-US or zh-CN still works when the version cannot be
// established.
//
// `force=true` passes the value through untouched — the CLI cannot know
// the canonical spelling of a locale it does not carry, and the upstream
// is the authority there.
func resolveLocale(ctx context.Context, f *cmdutil.Factory, value string, force bool) (string, error) {
	value = strings.TrimSpace(value)
	if force {
		return value, nil
	}
	for _, l := range supportedLocales {
		if !strings.EqualFold(value, l.code) {
			continue
		}
		if l.since == "" {
			return l.code, nil
		}
		ok, err := f.OlaresBackendAtLeast(ctx, l.since)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("locale %s (%s) needs Olares >= %s, where its translation shipped; pass --force to write it anyway",
				l.code, l.label, l.since)
		}
		return l.code, nil
	}
	return "", fmt.Errorf("unsupported locale %q; allowed locales:\n%s\npass --force to write a locale this CLI does not carry",
		value, localeChoices())
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

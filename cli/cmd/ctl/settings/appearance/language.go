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
// The picker in Settings -> Appearance writes free-form locale codes
// (e.g. "en", "zh-CN"); we don't validate against a hardcoded list here
// since the supported set evolves with each release. The server is the
// source of truth — if the value is rejected, surface the BFL error
// verbatim.
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
  set <locale>          update the system language        (Phase 2)
                        (e.g. "set en-US"; --value also accepted)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newLanguageSetCommand(f))
	return cmd
}

// newLanguageSetCommand registers `appearance language set [<locale>]`.
//
// Argument shape: a positional <locale> is the canonical form (matches
// the SKILL doc + SPA copy "language set en-US"); --value is kept as a
// strict-flag alternative for users who prefer larksuite/cli-style
// flag-only invocations. Exactly one of the two MUST be supplied; if
// both are passed and disagree we error out rather than silently picking
// one. The previous shape (required --value, NoArgs) was rejected by
// the smoke matrix as KI-17 because every other "verb <obj>" verb in
// this tree takes its primary subject positionally.
func newLanguageSetCommand(f *cmdutil.Factory) *cobra.Command {
	var value string
	cmd := &cobra.Command{
		Use:   "set <locale>",
		Short: "update the system language preference (e.g. set en-US)",
		Long: `Update the system language preference. The value is a locale code
the SPA's language picker emits (e.g. "en", "zh-CN"); the server defines
the supported set and will reject unknown codes.

The locale can be passed as a positional argument or as --value. Pass
exactly one of the two; passing both with conflicting values is an
error.

Examples:
  olares-cli settings appearance language set en-US
  olares-cli settings appearance language set zh-CN
  olares-cli settings appearance language set --value en
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			resolved, err := resolveLanguageValue(args, value)
			if err != nil {
				return err
			}
			return runLanguageSet(c.Context(), f, resolved)
		},
	}
	cmd.Flags().StringVar(&value, "value", "", "locale code to set (e.g. en, zh-CN); same as the positional <locale>")
	return cmd
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

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
  set --value <code>    update the system language        (Phase 2)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newLanguageSetCommand(f))
	return cmd
}

func newLanguageSetCommand(f *cmdutil.Factory) *cobra.Command {
	var value string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "update the system language preference",
		Long: `Update the system language preference. The value is a locale code
the SPA's language picker emits (e.g. "en", "zh-CN"); the server defines
the supported set and will reject unknown codes.

Examples:
  olares-cli settings appearance language set --value en
  olares-cli settings appearance language set --value zh-CN
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runLanguageSet(c.Context(), f, value)
		},
	}
	cmd.Flags().StringVar(&value, "value", "", "locale code to set (e.g. en, zh-CN)")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func runLanguageSet(ctx context.Context, f *cmdutil.Factory, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("--value must be a non-empty locale code (e.g. en, zh-CN)")
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

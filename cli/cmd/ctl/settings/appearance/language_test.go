package appearance

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// backendVersion pins what OlaresBackendAtLeast resolves against, the
// way version_gate_test.go does, so the gated locales can be exercised
// without a cluster.
func backendVersion(t *testing.T, version string) *cmdutil.Factory {
	t.Helper()
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })
	viper.Set(cmdutil.FlagOlaresVersion, version)
	return cmdutil.NewFactory()
}

func TestResolveLocale(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		version  string
		force    bool
		want     string
		wantErr  bool
		wantSubs []string
	}{
		{name: "allowed_en", value: "en-US", want: "en-US"},
		{name: "allowed_zh", value: "zh-CN", want: "zh-CN"},
		{name: "trim_then_match", value: "  en-US  ", want: "en-US"},
		// The SPA's i18n bundle is keyed by the canonical spelling, so a
		// case-insensitive match has to store that rather than the input.
		{name: "lowercased_is_canonicalized", value: "en-us", want: "en-US"},
		{name: "mixed_case_is_canonicalized", value: "ZH-cn", want: "zh-CN"},
		// The five locales that shipped in 1.12.7.
		{name: "added_locale_on_1_12_7", value: "de-DE", version: "1.12.7", want: "de-DE"},
		{name: "added_locale_canonicalized", value: "ja-jp", version: "1.12.7", want: "ja-JP"},
		// A dated build is its own line, and so is a prerelease: Settings
		// compares against 1.12.7-0 for the same reason.
		{name: "added_locale_on_daily", value: "fr-FR", version: "1.12.7-20260819", want: "fr-FR"},
		{name: "added_locale_on_prerelease", value: "es-ES", version: "1.12.7-alpha.4", want: "es-ES"},
		{
			name:     "added_locale_below_gate",
			value:    "de-DE",
			version:  "1.12.6",
			wantErr:  true,
			wantSubs: []string{"de-DE", "Deutsch", "Olares >= 1.12.7", "--force"},
		},
		{
			name:     "unknown_rejected",
			value:    "xx",
			wantErr:  true,
			wantSubs: []string{`"xx"`, "en-US (English)", "zh-CN (简体中文)", "de-DE (Deutsch)", "--force"},
		},
		{
			name:     "empty_rejected",
			value:    "",
			wantErr:  true,
			wantSubs: []string{"--force"},
		},
		// Forced values pass through untouched: the CLI cannot know the
		// canonical spelling of a locale it does not carry.
		{name: "force_overrides_unknown", value: "xx", force: true, want: "xx"},
		{name: "force_keeps_case", value: "ko-kr", force: true, want: "ko-kr"},
		{name: "force_overrides_the_gate", value: "de-DE", version: "1.12.5", force: true, want: "de-DE"},
		{name: "force_overrides_empty", value: "", force: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// An ungated locale must not need a version at all, so those
			// cases run against a version that cannot be resolved.
			version := tc.version
			if version == "" {
				version = "not-a-version"
			}
			f := backendVersion(t, version)
			got, err := resolveLocale(context.Background(), f, tc.value, tc.force)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveLocale(%q, %v) returned nil; want error", tc.value, tc.force)
				}
				msg := err.Error()
				for _, sub := range tc.wantSubs {
					if !strings.Contains(msg, sub) {
						t.Errorf("error %q does not contain %q", msg, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveLocale(%q, %v) returned error %v; want nil", tc.value, tc.force, err)
			}
			if got != tc.want {
				t.Errorf("resolveLocale(%q, %v) = %q; want %q", tc.value, tc.force, got, tc.want)
			}
		})
	}
}

// A gated locale on an undetectable backend has to say so rather than
// read as an unknown code.
func TestResolveLocaleReportsAnUnknownBackendVersion(t *testing.T) {
	f := backendVersion(t, "not-a-version")
	_, err := resolveLocale(context.Background(), f, "de-DE", false)
	if err == nil || !strings.Contains(err.Error(), "profile list --refresh-version") {
		t.Fatalf("expected the shared refresh hint, got %v", err)
	}
}

func TestDescribeLocale(t *testing.T) {
	for value, want := range map[string]string{
		"zh-CN": "zh-CN (简体中文)",
		"de-DE": "de-DE (Deutsch)",
		// Case-insensitive, and canonicalized on the way out.
		"en-us": "en-US (English)",
		// A code this CLI does not carry is shown as stored.
		"ko-KR": "ko-KR",
		"":      nonEmpty(""),
	} {
		if got := describeLocale(value); got != want {
			t.Errorf("describeLocale(%q) = %q; want %q", value, got, want)
		}
	}
}

// The parent help is where a reader learns the value domain, so it has
// to follow supportedLocales rather than restate a count by hand.
func TestLanguageHelpFollowsTheLocaleList(t *testing.T) {
	var gated []string
	for _, l := range supportedLocales {
		if l.since != "" {
			gated = append(gated, l.code)
		}
	}
	long := NewLanguageCommand(cmdutil.NewFactory()).Long
	want := fmt.Sprintf("en-US, zh-CN, and %d more on Olares >= %s", len(gated), localeMinVersionExtended)
	if !strings.Contains(long, want) {
		t.Errorf("parent help does not carry %q:\n%s", want, long)
	}
	// Help is offline text: no help line may run past a narrow terminal.
	for _, line := range strings.Split(long, "\n") {
		if len(line) > 72 {
			t.Errorf("help line is %d chars: %q", len(line), line)
		}
	}
}

// The gate exists because the desktop and LarePass ship different
// bundles; moving it silently would change which locales are accepted.
func TestLocaleChoicesNamesTheGate(t *testing.T) {
	if localeMinVersionExtended != "1.12.7" {
		t.Fatalf("localeMinVersionExtended = %q, want 1.12.7", localeMinVersionExtended)
	}
	choices := localeChoices()
	for _, want := range []string{
		"en-US (English)", "zh-CN (简体中文)",
		"Olares >= 1.12.7 also has:",
		"de-DE (Deutsch)", "es-ES (Español)", "it-IT (Italiano)",
		"fr-FR (Français)", "ja-JP (日本語)",
	} {
		if !strings.Contains(choices, want) {
			t.Errorf("choices missing %q:\n%s", want, choices)
		}
	}
	// The ungated pair leads; a reader should not have to work out which
	// half applies to their backend.
	if strings.Index(choices, "en-US") > strings.Index(choices, "de-DE") {
		t.Errorf("gated locales are listed before the ungated ones:\n%s", choices)
	}
}

func TestResolveLanguageValue(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		flag     string
		want     string
		wantErr  bool
		wantSubs []string
	}{
		{name: "positional_only", args: []string{"en-US"}, want: "en-US"},
		{name: "flag_only", flag: "zh-CN", want: "zh-CN"},
		{name: "positional_trimmed", args: []string{"  en-US  "}, want: "en-US"},
		{name: "flag_trimmed", flag: "  zh-CN  ", want: "zh-CN"},
		{name: "matching_both_ok", args: []string{"en-US"}, flag: "en-US", want: "en-US"},
		{
			name:     "neither_supplied_errors",
			wantErr:  true,
			wantSubs: []string{"locale code is required"},
		},
		{
			name:     "conflict_errors",
			args:     []string{"en-US"},
			flag:     "zh-CN",
			wantErr:  true,
			wantSubs: []string{"conflicting locale", `"en-US"`, `"zh-CN"`},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveLanguageValue(tc.args, tc.flag)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveLanguageValue(%v, %q) returned %q nil; want error", tc.args, tc.flag, got)
				}
				msg := err.Error()
				for _, sub := range tc.wantSubs {
					if !strings.Contains(msg, sub) {
						t.Errorf("error %q does not contain %q", msg, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveLanguageValue(%v, %q) errored: %v", tc.args, tc.flag, err)
			}
			if got != tc.want {
				t.Errorf("resolveLanguageValue(%v, %q) = %q; want %q", tc.args, tc.flag, got, tc.want)
			}
		})
	}
}

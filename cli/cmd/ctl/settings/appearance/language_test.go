package appearance

import (
	"strings"
	"testing"
)

func TestValidateLocale(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		force    bool
		wantErr  bool
		wantSubs []string
	}{
		{name: "allowed_en", value: "en-US"},
		{name: "allowed_zh", value: "zh-CN"},
		{name: "trim_then_match", value: "  en-US  "},
		{
			name:     "unknown_rejected",
			value:    "xx",
			wantErr:  true,
			wantSubs: []string{`"xx"`, "allowed: en-US, zh-CN", "--force"},
		},
		{
			name:     "empty_rejected",
			value:    "",
			wantErr:  true,
			wantSubs: []string{"--force"},
		},
		{name: "force_overrides_unknown", value: "xx", force: true},
		{name: "force_overrides_empty", value: "", force: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateLocale(tc.value, tc.force)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("validateLocale(%q, %v) returned nil; want error", tc.value, tc.force)
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
				t.Fatalf("validateLocale(%q, %v) returned error %v; want nil", tc.value, tc.force, err)
			}
		})
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

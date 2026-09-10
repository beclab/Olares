package appearance

import (
	"strings"
	"testing"
)

func TestResolveThemeValue(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		want     string
		wantErr  bool
		wantSubs []string
	}{
		{name: "light", raw: "light", want: "light"},
		{name: "dark", raw: "dark", want: "dark"},
		{name: "trimmed_and_lowercased", raw: "  Dark  ", want: "dark"},
		{
			name:     "unknown_rejected",
			raw:      "auto",
			wantErr:  true,
			wantSubs: []string{`"auto"`, "allowed: light, dark"},
		},
		{
			name:     "empty_rejected",
			raw:      "   ",
			wantErr:  true,
			wantSubs: []string{"theme value is required"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveThemeValue(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveThemeValue(%q) = %q, nil; want error", tc.raw, got)
				}
				for _, sub := range tc.wantSubs {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("error %q does not contain %q", err, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveThemeValue(%q) errored: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("resolveThemeValue(%q) = %q; want %q", tc.raw, got, tc.want)
			}
		})
	}
}

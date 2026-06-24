package manifest

import (
	"strings"
	"testing"
)

func TestCheckNoTemplateSyntax(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErr   bool
		wantLine  int
		wantInMsg []string
	}{
		{
			name:    "clean modern manifest",
			content: "olaresManifest.version: 0.12.0\nmetadata:\n  name: foo\n",
			wantErr: false,
		},
		{
			name:    "double-quoted scalar shields a lone `{{`",
			content: "metadata:\n  description: \"uses {{ markers for code blocks\"\n",
			wantErr: false,
		},
		{
			name:    "double-quoted scalar shields a lone `}}`",
			content: "metadata:\n  description: \"ends with literal }} sequence\"\n",
			wantErr: false,
		},
		{
			name:    "double-quoted scalar shields a paired `{{ ... }}`",
			content: "spec:\n  description: \"use {{ .Values.foo }} as a placeholder\"\n",
			wantErr: false,
		},
		{
			name:    "single-quoted scalar shields a paired `{{ ... }}`",
			content: "spec:\n  description: 'use {{ .Values.foo }} as a placeholder'\n",
			wantErr: false,
		},
		{
			name:    "single-quoted scalar with embedded `''` escape stays inside the scalar",
			content: "spec:\n  description: 'it''s safe to mention {{ .x }} inside docs'\n",
			wantErr: false,
		},
		{
			name:    "block scalar `|` shields placeholders on every body line",
			content: "spec:\n  description: |\n    Multi-line text\n    that shows {{ .Values.foo }}\n    and {{ .Values.bar }} too.\n",
			wantErr: false,
		},
		{
			name:    "folded block scalar `>-` shields placeholders",
			content: "spec:\n  description: >-\n    Folded text\n    mentioning {{ .x }}\n    inline.\n",
			wantErr: false,
		},
		{
			name:    "double-quoted scalar with backslash-escaped brace stays inside the scalar",
			content: "spec:\n  description: \"path \\\\ then {{ .x }}\"\n",
			wantErr: false,
		},
		{
			name:     "unquoted `{{ ... }}` value is rejected",
			content:  "line1\nline2\nmetadata:\n  name: {{ .Values.bfl.username }}\n",
			wantErr:  true,
			wantLine: 4,
			wantInMsg: []string{
				"template syntax",
				"{{",
				"0.12.0",
			},
		},
		{
			name:     "unquoted lone `{{` mid-scalar is rejected",
			content:  "metadata:\n  name: hello {{ no closer here\n",
			wantErr:  true,
			wantLine: 2,
			wantInMsg: []string{
				`"{{"`,
			},
		},
		{
			name:     "unquoted lone `}}` mid-scalar is rejected",
			content:  "metadata:\n  name: trailing }} braces\n",
			wantErr:  true,
			wantLine: 2,
			wantInMsg: []string{
				`"}}"`,
			},
		},
		{
			name:     "block scalar exits on dedent and a placeholder afterwards still trips",
			content:  "spec:\n  description: |\n    Documented {{ .ok }} here.\n  other: {{ .leaks }}\n",
			wantErr:  true,
			wantLine: 4,
		},
		{
			name:    "comment containing `|` does not open a fake block scalar",
			content: "spec: foo # use | for pipes\nname: {{ .x }}\n",
			wantErr: true,
			wantInMsg: []string{
				"line 2:",
			},
		},
		{
			name:    "two unquoted placeholders on different lines: first wins",
			content: "k1: {{ a }}\nk2: {{ b }}\n",
			wantErr: true,
			wantInMsg: []string{
				"line 1:",
			},
		},
		{
			name:    "single brace in flow style is left alone",
			content: "metadata:\n  labels: {a: 1, b: 2}\n",
			wantErr: false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := checkNoTemplateSyntax([]byte(tc.content))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if tc.wantLine != 0 {
					prefix := "line " + itoa(tc.wantLine) + ":"
					if !strings.Contains(err.Error(), prefix) {
						t.Errorf("error %q must include %q", err.Error(), prefix)
					}
				}
				for _, want := range tc.wantInMsg {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q must include %q", err.Error(), want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestMaskYAMLStringRegions_PreservesLength documents the masker's
// length-preservation contract: callers report match offsets against
// the original content using indices found in the masked copy, so any
// byte the masker touches must be replaced (not deleted).
func TestMaskYAMLStringRegions_PreservesLength(t *testing.T) {
	cases := []string{
		"plain: value\n",
		"q: \"with {{ braces }}\"\n",
		"q: 'with {{ braces }}'\n",
		"desc: |\n  block {{ x }} scalar\n  more {{ y }} text\n",
		"mixed: \"a\"\nbare: {{ leak }}\n",
	}
	for _, c := range cases {
		got := maskYAMLStringRegions([]byte(c))
		if len(got) != len(c) {
			t.Errorf("masker changed length for %q: got %d, want %d", c, len(got), len(c))
		}
	}
}

// itoa is a tiny stdlib-free decimal helper kept local so the test
// file does not need to pull strconv in just for two test labels.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

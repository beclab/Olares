package manifest

import (
	"strings"
	"testing"
)

func TestPeek_Plain(t *testing.T) {
	h, err := Peek([]byte(`olaresManifest.version: 0.12.0
apiVersion: v1
metadata:
  name: x
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "0.12.0" || h.APIVersion != "v1" {
		t.Fatalf("unexpected header: %+v", h)
	}
}

func TestPeek_WithHelmTemplateBlocks(t *testing.T) {
	// Legacy manifests carry Go-template blocks that gopkg.in/yaml.v3 refuses
	// to parse. The probe must still return the dispatch fields from the
	// column-0 lines.
	content := []byte(`olaresManifest.version: "0.8.1"
apiVersion: v1
entrances:
{{- if and .Values.admin .Values.bfl.username (eq .Values.admin .Values.bfl.username) }}
- name: a
  port: 1
{{- else }}
- name: b
  port: 2
{{- end }}
olaresManifest.type: app
`)
	h, err := Peek(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "0.8.1" {
		t.Fatalf("OlaresVersion = %q, want 0.8.1", h.OlaresVersion)
	}
	if h.APIVersion != "v1" {
		t.Fatalf("APIVersion = %q, want v1", h.APIVersion)
	}
}

func TestPeek_QuotedAndComments(t *testing.T) {
	cases := []struct {
		name    string
		content string
		version string
		api     string
	}{
		{
			name:    "double quotes",
			content: `olaresManifest.version: "1.2.3"` + "\n" + `apiVersion: "v2"`,
			version: "1.2.3",
			api:     "v2",
		},
		{
			name:    "single quotes",
			content: `olaresManifest.version: '1.2.3'` + "\n" + `apiVersion: 'v2'`,
			version: "1.2.3",
			api:     "v2",
		},
		{
			name:    "trailing comment",
			content: "olaresManifest.version: 0.11.9 # legacy\napiVersion: v1 # stable\n",
			version: "0.11.9",
			api:     "v1",
		},
		{
			name:    "comment-only line",
			content: "# header\nolaresManifest.version: 0.15.0\n# trailer\napiVersion: v2\n",
			version: "0.15.0",
			api:     "v2",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h, err := Peek([]byte(tc.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h.OlaresVersion != tc.version || h.APIVersion != tc.api {
				t.Fatalf("got %+v, want {%q %q}", h, tc.version, tc.api)
			}
		})
	}
}

func TestPeek_IgnoresIndentedOccurrences(t *testing.T) {
	// A nested `apiVersion:` inside metadata/spec/etc. must not override the
	// real column-0 field. The probe should stop at the first top-level
	// match.
	content := []byte(`olaresManifest.version: 0.13.0
apiVersion: v1
metadata:
  apiVersion: shouldnt-be-picked
spec:
  olaresManifest.version: shouldnt-be-picked-either
`)
	h, err := Peek(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "0.13.0" {
		t.Fatalf("OlaresVersion = %q, want 0.13.0", h.OlaresVersion)
	}
	if h.APIVersion != "v1" {
		t.Fatalf("APIVersion = %q, want v1", h.APIVersion)
	}
}

func TestPeek_Missing(t *testing.T) {
	h, err := Peek([]byte("metadata:\n  name: x\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "" || h.APIVersion != "" {
		t.Fatalf("expected empty header, got %+v", h)
	}
}

func TestPeek_EmptyInput(t *testing.T) {
	h, err := Peek(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "" || h.APIVersion != "" {
		t.Fatalf("expected empty header, got %+v", h)
	}
}

func TestPeek_ListItemIsIgnored(t *testing.T) {
	// A `- apiVersion:` line inside an array must not be picked up.
	content := []byte(`olaresManifest.version: 0.12.5
- apiVersion: decoy
apiVersion: v1
`)
	h, err := Peek(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.APIVersion != "v1" {
		t.Fatalf("APIVersion = %q, want v1", h.APIVersion)
	}
}

func TestPeek_StopsAfterBothFound(t *testing.T) {
	// Sanity check: after both fields are extracted, the scanner should stop
	// and therefore not care about any malformed tail. We exercise this by
	// providing a tail that would be problematic if re-scanned.
	content := []byte(`olaresManifest.version: 0.9.0
apiVersion: v1
this is not: valid: yaml: at: all
`)
	h, err := Peek(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "0.9.0" || h.APIVersion != "v1" {
		t.Fatalf("unexpected header: %+v", h)
	}
}

func TestPeek_HandlesCRLF(t *testing.T) {
	content := []byte(strings.Join([]string{
		"olaresManifest.version: 0.10.0",
		"apiVersion: v1",
		"",
	}, "\r\n"))
	h, err := Peek(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.OlaresVersion != "0.10.0" || h.APIVersion != "v1" {
		t.Fatalf("unexpected header: %+v", h)
	}
}

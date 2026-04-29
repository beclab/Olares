package oac

import (
	"strings"
	"testing"
)

func TestRenderManifestTemplate_SubstitutesValues(t *testing.T) {
	tmpl := []byte(`olaresManifest.version: 0.8.1
apiVersion: v1
metadata:
  name: {{ .Values.bfl.username }}-app
  admin: {{ .Values.admin }}
`)
	out, err := renderManifestTemplate(tmpl, "alice", "root")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "alice-app") {
		t.Fatalf("expected username substitution, got %q", got)
	}
	if !strings.Contains(got, "admin: root") {
		t.Fatalf("expected admin substitution, got %q", got)
	}
}

func TestRenderManifestTemplate_FallsBackToDefaults(t *testing.T) {
	tmpl := []byte(`{{ .Values.bfl.username }}-{{ .Values.admin }}`)
	out, err := renderManifestTemplate(tmpl, "", "")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if string(out) != defaultRenderOwner+"-"+defaultRenderAdmin {
		t.Fatalf("unexpected fallback output: %q", out)
	}
}

// TestLintBothOwnerScenarios_BadPath documents that LintBothOwnerScenarios
// does not silently succeed when the target directory is missing: the
// folder-layout check (which runs before any owner-scenario expansion)
// still trips and surfaces a real error to the caller.
//
// The previous version of this test double-asserted "empty admin must
// error" / "empty user must error" against the same bogus path, but
// today the API does not require admin or user arguments — the two
// identical branches were leftover from an earlier signature.
func TestLintBothOwnerScenarios_BadPath(t *testing.T) {
	if err := LintBothOwnerScenarios("/no/such/path"); err == nil {
		t.Fatal("expected error for a non-existent chart directory")
	}
}

func TestPeekManifestVersions(t *testing.T) {
	raw := []byte(`olaresManifest.version: "0.12.0"
apiVersion: v2
metadata:
  name: demo
`)
	v, err := PeekManifestVersions(raw)
	if err != nil {
		t.Fatalf("peek: %v", err)
	}
	if v.APIVersion != "v2" {
		t.Fatalf("APIVersion: want v2, got %q", v.APIVersion)
	}
	if v.OlaresManifestVersion != "0.12.0" {
		t.Fatalf("OlaresManifestVersion: want 0.12.0, got %q", v.OlaresManifestVersion)
	}
}

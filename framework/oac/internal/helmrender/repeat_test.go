package helmrender

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemplate(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, "templates", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestRejectTemplateRepeat_FlagsAction(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "data:\n  payload: {{ repeat 2048 (repeat 8 \"x\") }}\n")
	err := rejectTemplateRepeat(dir)
	if err == nil {
		t.Fatal("expected repeat in a template action to be rejected")
	}
	if !strings.Contains(err.Error(), "templates/cm.yaml") || !strings.Contains(err.Error(), "repeat") {
		t.Fatalf("error should name the template and repeat, got: %v", err)
	}
}

func TestRejectTemplateRepeat_WhitespaceControl(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "data: {{- repeat 1 \"x\" -}}\n")
	if err := rejectTemplateRepeat(dir); err == nil {
		t.Fatal("expected {{- repeat -}} to be rejected")
	}
}

func TestRejectTemplateRepeat_IgnoresNonActionText(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "data:\n  note: do not repeat this\n  name: {{ .Release.Name }}\n")
	if err := rejectTemplateRepeat(dir); err != nil {
		t.Fatalf("prose and ordinary actions must pass: %v", err)
	}
}

func TestRejectTemplateRepeat_AllowsFieldAndVariable(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "data:\n  a: {{ .Values.repeat }}\n  b: {{ $repeat }}\n  c: {{ .repeat }}\n  d: {{ $.repeat }}\n")
	if err := rejectTemplateRepeat(dir); err != nil {
		t.Fatalf("field and variable names must pass: %v", err)
	}
}

func TestRejectTemplateRepeat_IgnoresRepeatedWord(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "data: {{ print \"repeated\" }}\n")
	if err := rejectTemplateRepeat(dir); err != nil {
		t.Fatalf("the word repeated must not be treated as the repeat function: %v", err)
	}
}

func TestRender_RejectsRepeatBeforeHelm(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "cm.yaml", "payload: {{ repeat 2 \"x\" }}\n")
	_, err := Render(dir, map[string]interface{}{}, "demo", "")
	if err == nil {
		t.Fatal("expected Render to reject repeat")
	}
	if !strings.Contains(err.Error(), "repeat") {
		t.Fatalf("error should mention repeat, got: %v", err)
	}
	if strings.Contains(err.Error(), "Chart.yaml") {
		t.Fatalf("helm loaded the chart before the repeat check: %v", err)
	}
}

package chartfolder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	olm "github.com/beclab/Olares/framework/oac/internal/manifest"
)

// stubManifest is the minimum manifest.Manifest needed by CheckConsistency
// and CheckWithTitle. The remaining methods return zero values because the
// chart-folder layer never reads them.
type stubManifest struct{ name, version string }

func (s stubManifest) APIVersion() string            { return "v1" }
func (s stubManifest) ConfigVersion() string         { return "0.13.0" }
func (s stubManifest) ConfigType() string            { return "app" }
func (s stubManifest) AppName() string               { return s.name }
func (s stubManifest) AppVersion() string            { return s.version }
func (s stubManifest) Entrances() []olm.EntranceInfo { return nil }
func (s stubManifest) OptionsImages() []string       { return nil }
func (s stubManifest) PermissionAppData() bool       { return false }
func (s stubManifest) Raw() any                      { return nil }

// writeFile creates a file at p with contents.
func writeFile(t *testing.T, p, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// makeChart builds a well-formed chart folder under t.TempDir. Callers can
// override which sub-pieces are present by deleting after creation.
func makeChart(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	writeFile(t, filepath.Join(root, "Chart.yaml"), "apiVersion: v2\nname: "+name+"\nversion: 1.2.3\n")
	writeFile(t, filepath.Join(root, "values.yaml"), "")
	writeFile(t, filepath.Join(root, "templates", "deployment.yaml"), "")
	writeFile(t, filepath.Join(root, "OlaresManifest.yaml"), "olaresManifest.version: 0.13.0\n")
	return root
}

func TestIsValidFolderName(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"abc", true},
		{"abc123", true},
		{"a", true},
		{"", false},
		{"ABC", false},
		{"ab-c", false},
		{strings.Repeat("a", 30), true},
		{strings.Repeat("a", 31), false},
	}
	for _, tc := range cases {
		if got := isValidFolderName(tc.in); got != tc.want {
			t.Errorf("isValidFolderName(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsReservedWord(t *testing.T) {
	for _, n := range []string{"user", "system", "kube", "Bytetrade"} {
		if !isReservedWord(n) {
			t.Errorf("%q should be reserved", n)
		}
	}
	for _, n := range []string{"firefox", "myapp", "xy"} {
		if isReservedWord(n) {
			t.Errorf("%q should NOT be reserved", n)
		}
	}
}

func TestValidCategories(t *testing.T) {
	if validCategories(nil) {
		t.Error("empty list must not validate")
	}
	if !validCategories([]string{"AI", "Productivity"}) {
		t.Error("known categories should validate")
	}
	if validCategories([]string{"AI", "Bogus"}) {
		t.Error("unknown category should fail")
	}
}

func TestCheckLayout_Happy(t *testing.T) {
	root := makeChart(t, "myapp")
	if err := CheckLayout(root); err != nil {
		t.Fatalf("expected layout to pass: %v", err)
	}
}

func TestCheckLayout_BadFolderName(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "Bad-Name")
	writeFile(t, filepath.Join(bad, "Chart.yaml"), "name: x\n")
	err := CheckLayout(bad)
	if err == nil {
		t.Fatal("expected invalid folder name error")
	}
	if !strings.Contains(err.Error(), "invalid folder name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckLayout_MissingPieces(t *testing.T) {
	cases := []struct {
		name   string
		remove string
		want   string
	}{
		{"missing Chart.yaml", "Chart.yaml", "missing Chart.yaml"},
		{"missing values.yaml", "values.yaml", "missing values.yaml"},
		{"missing templates", "templates", "missing templates folder"},
		{"missing OlaresManifest", "OlaresManifest.yaml", "missing OlaresManifest.yaml"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := makeChart(t, "myapp")
			if err := os.RemoveAll(filepath.Join(root, tc.remove)); err != nil {
				t.Fatalf("remove: %v", err)
			}
			err := CheckLayout(root)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q missing %q", err, tc.want)
			}
		})
	}
}

func TestLoadChart_Happy(t *testing.T) {
	root := makeChart(t, "firefox")
	c, err := LoadChart(root)
	if err != nil {
		t.Fatalf("LoadChart: %v", err)
	}
	if c.Name != "firefox" || c.Version != "1.2.3" || c.APIVersion != "v2" {
		t.Fatalf("unexpected chart: %+v", c)
	}
}

func TestLoadChart_MissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := LoadChart(root)
	if err == nil {
		t.Fatal("expected error: missing Chart.yaml")
	}
}

func TestLoadChart_FieldsRequired(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{"missing apiVersion", "name: x\nversion: 1.0.0\n", "apiVersion field empty"},
		{"missing name", "apiVersion: v2\nversion: 1.0.0\n", "name field empty"},
		{"missing version", "apiVersion: v2\nname: x\n", "version field empty"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "Chart.yaml"), tc.body)
			_, err := LoadChart(root)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q missing %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoadChart_Malformed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Chart.yaml"), "::not yaml::")
	_, err := LoadChart(root)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCheckConsistency_Happy(t *testing.T) {
	root := makeChart(t, "myapp")
	chart := &Chart{APIVersion: "v2", Name: "myapp", Version: "1.2.3"}
	m := stubManifest{name: "myapp", version: "1.2.3"}
	if err := CheckConsistency(root, chart, m); err != nil {
		t.Fatalf("CheckConsistency: %v", err)
	}
}

func TestCheckConsistency_NameMismatch(t *testing.T) {
	root := makeChart(t, "myapp")
	chart := &Chart{APIVersion: "v2", Name: "different", Version: "1.2.3"}
	m := stubManifest{name: "myapp", version: "1.2.3"}
	err := CheckConsistency(root, chart, m)
	if err == nil {
		t.Fatal("expected name mismatch error")
	}
	if !strings.Contains(err.Error(), "name must be the same") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckConsistency_VersionMismatch(t *testing.T) {
	root := makeChart(t, "myapp")
	chart := &Chart{APIVersion: "v2", Name: "myapp", Version: "1.2.3"}
	m := stubManifest{name: "myapp", version: "9.9.9"}
	err := CheckConsistency(root, chart, m)
	if err == nil {
		t.Fatal("expected version mismatch error")
	}
	if !strings.Contains(err.Error(), "Version must be the same") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckWithTitle_Happy(t *testing.T) {
	root := makeChart(t, "myapp")
	chart := &Chart{APIVersion: "v2", Name: "myapp", Version: "1.2.3"}
	m := stubManifest{name: "myapp", version: "1.2.3"}
	title := TitleInfo{Folder: "myapp", Version: "1.2.3"}
	if err := CheckWithTitle(root, chart, m, title, []string{"AI"}); err != nil {
		t.Fatalf("CheckWithTitle: %v", err)
	}
}

func TestCheckWithTitle_RejectsReserved(t *testing.T) {
	root := makeChart(t, "system")
	chart := &Chart{APIVersion: "v2", Name: "system", Version: "1.2.3"}
	m := stubManifest{name: "system", version: "1.2.3"}
	title := TitleInfo{Folder: "system", Version: "1.2.3"}
	err := CheckWithTitle(root, chart, m, title, []string{"AI"})
	if err == nil {
		t.Fatal("expected reserved-folder error")
	}
	if !strings.Contains(err.Error(), "reserved foldername") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckWithTitle_RejectsBadCategories(t *testing.T) {
	root := makeChart(t, "myapp")
	chart := &Chart{APIVersion: "v2", Name: "myapp", Version: "1.2.3"}
	m := stubManifest{name: "myapp", version: "1.2.3"}
	title := TitleInfo{Folder: "myapp", Version: "1.2.3"}
	err := CheckWithTitle(root, chart, m, title, []string{"Bogus"})
	if err == nil {
		t.Fatal("expected unsupported category error")
	}
	if !strings.Contains(err.Error(), "unsupported category") {
		t.Fatalf("unexpected error: %v", err)
	}
}

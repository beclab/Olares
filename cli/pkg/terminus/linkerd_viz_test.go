package terminus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkerdVizValuesPath_vendor(t *testing.T) {
	dir := t.TempDir()
	vals := filepath.Join(dir, linkerdVizValuesFileName)
	if err := os.WriteFile(vals, []byte("prometheus:\n  enabled: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := linkerdVizValuesPath(dir)
	if got != vals {
		t.Fatalf("got %q want %q", got, vals)
	}
}

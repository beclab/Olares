package skills

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

const releaseVersion = "1.12.7-cli.3"

// installStore writes a store whose probe skill declares the given version,
// standing in for an install made by an older binary.
func installStore(t *testing.T, home, declared string) string {
	t.Helper()
	store := filepath.Join(home, ".agents", "skills")
	if _, err := skillsuite.Export(store); err != nil {
		t.Fatalf("export: %v", err)
	}
	entry := filepath.Join(store, probe, skillsuite.EntryFile)
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatalf("read %s: %v", entry, err)
	}
	meta, err := skillsuite.ParseMeta(source)
	if err != nil {
		t.Fatalf("parse %s: %v", entry, err)
	}
	rewritten := strings.Replace(string(source), "version: "+meta.Version, "version: "+declared, 1)
	if rewritten == string(source) {
		t.Fatalf("could not rewrite the version in %s", entry)
	}
	if err := os.WriteFile(entry, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write %s: %v", entry, err)
	}
	return store
}

func TestNoticeNamesTheDriftAndTheFix(t *testing.T) {
	home := sandboxHome(t)
	t.Setenv(SilenceNoticeEnv, "")
	store := installStore(t, home, "0.0.1-ancient")

	var out bytes.Buffer
	Notice(&out, releaseVersion)

	said := out.String()
	if said == "" {
		t.Fatal("said nothing about skills from another release")
	}
	for _, expected := range []string{store, "0.0.1-ancient", "skills install", SilenceNoticeEnv} {
		if !strings.Contains(said, expected) {
			t.Errorf("the notice does not mention %q:\n%s", expected, said)
		}
	}
}

// Every uncertainty resolves to silence, because a notice that fires when
// nothing is wrong is one people learn to skip past.
func TestNoticeStaysQuiet(t *testing.T) {
	t.Run("skills match this binary", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		if _, err := skillsuite.Export(filepath.Join(home, ".agents", "skills")); err != nil {
			t.Fatalf("export: %v", err)
		}
		assertSilent(t, releaseVersion)
	})

	t.Run("nothing is installed", func(t *testing.T) {
		sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		assertSilent(t, releaseVersion)
	})

	t.Run("a development build", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		installStore(t, home, "0.0.1-ancient")
		assertSilent(t, "0.0.0-development")
	})

	t.Run("silenced on purpose", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "1")
		installStore(t, home, "0.0.1-ancient")
		assertSilent(t, releaseVersion)
	})

	t.Run("an unreadable copy", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		store := installStore(t, home, "0.0.1-ancient")
		entry := filepath.Join(store, probe, skillsuite.EntryFile)
		if err := os.WriteFile(entry, []byte("no frontmatter here\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", entry, err)
		}
		assertSilent(t, releaseVersion)
	})
}

func assertSilent(t *testing.T, binaryVersion string) {
	t.Helper()
	var out bytes.Buffer
	Notice(&out, binaryVersion)
	if out.Len() != 0 {
		t.Errorf("expected silence, got:\n%s", out.String())
	}
}

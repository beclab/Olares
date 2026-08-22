package skills

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	skillsuite "github.com/beclab/Olares/cli/skills"
)

const releaseVersion = "1.12.7-cli.3"

// installStore writes a store exactly as this binary would, marker included.
func installStore(t *testing.T, home string) string {
	t.Helper()
	store := filepath.Join(home, ".agents", "skills")
	if _, err := skillsuite.Export(store); err != nil {
		t.Fatalf("export: %v", err)
	}
	return store
}

// writtenByAnother rewrites the marker so the store looks like the work of a
// different olares-cli. The digest is provenance — which binary wrote this
// tree — so this, and not editing a skill, is what another release leaves
// behind.
func writtenByAnother(t *testing.T, store, digest, version string) {
	t.Helper()
	source, err := json.Marshal(skillsuite.Identity{Digest: digest, Version: version})
	if err != nil {
		t.Fatalf("encode marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(store, skillsuite.SuiteMarker), source, 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
}

// beforeMarkersExisted removes the marker, which is every machine that ran
// `skills install` from a build older than this one.
func beforeMarkersExisted(t *testing.T, store string) {
	t.Helper()
	if err := os.Remove(filepath.Join(store, skillsuite.SuiteMarker)); err != nil {
		t.Fatalf("remove marker: %v", err)
	}
}

// declareVersion rewrites what the probe skill on disk says about itself,
// which is all the fallback check has to go on.
func declareVersion(t *testing.T, store, version string) {
	t.Helper()
	entry := filepath.Join(store, probe, skillsuite.EntryFile)
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatalf("read %s: %v", entry, err)
	}
	meta, err := skillsuite.ParseMeta(source)
	if err != nil {
		t.Fatalf("parse %s: %v", entry, err)
	}
	rewritten := strings.Replace(string(source), "version: "+meta.Version, "version: "+version, 1)
	if rewritten == string(source) {
		t.Fatalf("could not rewrite the version in %s", entry)
	}
	if err := os.WriteFile(entry, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write %s: %v", entry, err)
	}
}

func TestNoticeNamesTheDriftAndTheFix(t *testing.T) {
	home := sandboxHome(t)
	t.Setenv(SilenceNoticeEnv, "")
	store := installStore(t, home)
	writtenByAnother(t, store, "an older release's digest", "0.0.1-ancient")

	said := notice(t, releaseVersion)
	if said == "" {
		t.Fatal("said nothing about skills from another release")
	}
	mine, err := skillsuite.SuiteVersion()
	if err != nil {
		t.Fatalf("suite version: %v", err)
	}
	for _, expected := range []string{store, "0.0.1-ancient", mine, "skills install", SilenceNoticeEnv} {
		if !strings.Contains(said, expected) {
			t.Errorf("the notice does not mention %q:\n%s", expected, said)
		}
	}
}

// The daily channel ships whatever is on main under a version somebody bumps
// every few weeks, so the two copies genuinely carry one label. Reporting
// that as "they declare 1.12.7-cli.5, this build carries 1.12.7-cli.5" reads
// as a bug in the notice and gets ignored.
func TestNoticeSaysWhenTwoCopiesShareOneVersion(t *testing.T) {
	home := sandboxHome(t)
	t.Setenv(SilenceNoticeEnv, "")
	store := installStore(t, home)
	mine, err := skillsuite.SuiteVersion()
	if err != nil {
		t.Fatalf("suite version: %v", err)
	}
	writtenByAnother(t, store, "the same version, other content", mine)

	said := notice(t, releaseVersion)
	if said == "" {
		t.Fatal("said nothing about a different copy of the same version")
	}
	if strings.Contains(said, "declare") {
		t.Errorf("the notice compares two identical labels:\n%s", said)
	}
	if !strings.Contains(said, mine) {
		t.Errorf("the notice does not name the version both copies carry:\n%s", said)
	}
}

// Upgrading across the release that introduced the marker is the common
// case, and the store it finds was written without one.
func TestNoticeFallsBackToVersionsWithoutAMarker(t *testing.T) {
	home := sandboxHome(t)
	t.Setenv(SilenceNoticeEnv, "")
	store := installStore(t, home)
	beforeMarkersExisted(t, store)
	declareVersion(t, store, "0.0.1-ancient")

	said := notice(t, releaseVersion)
	if !strings.Contains(said, "0.0.1-ancient") {
		t.Errorf("a store with no marker was not compared by version:\n%s", said)
	}
}

// Every uncertainty resolves to silence, because a notice that fires when
// nothing is wrong is one people learn to skip past.
func TestNoticeStaysQuiet(t *testing.T) {
	t.Run("this binary wrote them", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		installStore(t, home)
		assertSilent(t, releaseVersion)
	})

	t.Run("no marker, and the versions agree", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		beforeMarkersExisted(t, installStore(t, home))
		assertSilent(t, releaseVersion)
	})

	// The question is which binary wrote this tree, not whether anybody has
	// touched it since. Somebody adapting an installed skill for their own
	// machine is not drift, and hashing fifty files on every command to find
	// out would be a cost paid on every command.
	t.Run("a skill edited after it was installed", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		store := installStore(t, home)
		declareVersion(t, store, "9.9.9-mine.1")
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
		writtenByAnother(t, installStore(t, home), "another digest", "0.0.1-ancient")
		assertSilent(t, "0.0.0-development")
	})

	t.Run("silenced on purpose", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "1")
		writtenByAnother(t, installStore(t, home), "another digest", "0.0.1-ancient")
		assertSilent(t, releaseVersion)
	})

	t.Run("an unreadable copy and no marker", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		store := installStore(t, home)
		beforeMarkersExisted(t, store)
		entry := filepath.Join(store, probe, skillsuite.EntryFile)
		if err := os.WriteFile(entry, []byte("no frontmatter here\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", entry, err)
		}
		assertSilent(t, releaseVersion)
	})

	t.Run("a marker nothing can read", func(t *testing.T) {
		home := sandboxHome(t)
		t.Setenv(SilenceNoticeEnv, "")
		store := installStore(t, home)
		if err := os.WriteFile(filepath.Join(store, skillsuite.SuiteMarker), []byte("{"), 0o644); err != nil {
			t.Fatalf("write marker: %v", err)
		}
		// It falls back to the versions, which agree.
		assertSilent(t, releaseVersion)
	})
}

func notice(t *testing.T, binaryVersion string) string {
	t.Helper()
	var out bytes.Buffer
	Notice(&out, binaryVersion)
	return out.String()
}

func assertSilent(t *testing.T, binaryVersion string) {
	t.Helper()
	if said := notice(t, binaryVersion); said != "" {
		t.Errorf("expected silence, got:\n%s", said)
	}
}

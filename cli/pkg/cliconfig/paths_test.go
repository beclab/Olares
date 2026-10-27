package cliconfig

import (
	"path/filepath"
	"testing"
)

// Inside an application container the platform mounts a writable emptyDir and
// exports $OLARES_CLI_CACHE_DIR; config.json has to land there because HOME is
// often read-only or owned by another uid.
func TestHomeUsesCacheDir(t *testing.T) {
	cache := t.TempDir()
	t.Setenv(homeEnv, "")
	t.Setenv(cacheDirEnv, cache)

	got, err := Home()
	if err != nil {
		t.Fatalf("Home(): %v", err)
	}
	if want := filepath.Join(cache, "config"); got != want {
		t.Fatalf("Home() = %q, want %q", got, want)
	}
}

func TestHomeOverrideBeatsCacheDir(t *testing.T) {
	override := t.TempDir()
	t.Setenv(homeEnv, override)
	t.Setenv(cacheDirEnv, t.TempDir())

	got, err := Home()
	if err != nil {
		t.Fatalf("Home(): %v", err)
	}
	if got != override {
		t.Fatalf("Home() = %q, want %q", got, override)
	}
}

// A host install has neither variable set and must keep using ~/.olares-cli.
func TestHomeFallsBackToUserHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv(homeEnv, "")
	t.Setenv(cacheDirEnv, "")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := Home()
	if err != nil {
		t.Fatalf("Home(): %v", err)
	}
	if want := filepath.Join(home, defaultDir); got != want {
		t.Fatalf("Home() = %q, want %q", got, want)
	}
}

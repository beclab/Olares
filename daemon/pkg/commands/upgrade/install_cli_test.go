package upgrade

import (
	"context"
	"github.com/Masterminds/semver/v3"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallCLIRejectsWrongVersionWithoutReplacingWorkingBinary(t *testing.T) {
	dir := t.TempDir()
	source, destination := filepath.Join(dir, "downloaded"), filepath.Join(dir, "installed")
	old := []byte("working old CLI")
	if err := os.WriteFile(destination, old, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("#!/bin/sh\necho 'olares-cli version 1.12.7'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := installCLIFile(context.Background(), source, destination, semver.MustParse("1.12.8")); err == nil {
		t.Fatal("accepted wrong CLI version")
	}
	got, err := os.ReadFile(destination)
	if err != nil || string(got) != string(old) {
		t.Fatalf("old CLI was changed: %q, %v", got, err)
	}
	if err := os.WriteFile(source, []byte("#!/bin/sh\necho 'olares-cli version 1.12.8'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := installCLIFile(context.Background(), source, destination, semver.MustParse("1.12.8")); err != nil {
		t.Fatal(err)
	}
	v, err := cliVersionAt(context.Background(), destination)
	if err != nil || v.String() != "1.12.8" {
		t.Fatalf("installed version = %v, %v", v, err)
	}
}

package oac

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
)

func TestCheckResourceLimits_ModernV1_UsesInlineManifestLimits(t *testing.T) {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	dir := filepath.Join("testdata", "resourcelimits_v1_inline")
	m, err := c.LoadManifestFile(dir)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	sc := ownerScenario{owner: "alice", admin: "admin"}
	if err := c.checkResourceLimits(dir, m, sc, nil); err != nil {
		t.Fatalf("checkResourceLimits: %v", err)
	}
}

func TestCheckResourceLimits_ModernV1_InlineMismatchFails(t *testing.T) {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	dir := filepath.Join("testdata", "resourcelimits_v1_inline_toobig")
	m, err := c.LoadManifestFile(dir)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	sc := ownerScenario{owner: "alice", admin: "admin"}
	limErr := c.checkResourceLimits(dir, m, sc, nil)
	if limErr == nil {
		t.Fatal("expected error: container requests exceed inline spec.requiredCpu")
	}
	if !strings.Contains(limErr.Error(), "spec.requiredCpu") {
		t.Fatalf("expected spec.requiredCpu in error, got: %v", limErr)
	}
}

func TestCheckResourceLimits_ModernV3_SameAsV1Inline(t *testing.T) {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	dir := filepath.Join("testdata", "resourcelimits_v1_inline")
	m, err := c.LoadManifestFile(dir)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		t.Fatal("expected *AppConfiguration")
	}
	cfg.APIVersion = manifest.APIVersionV3
	sc := ownerScenario{owner: "alice", admin: "admin"}
	if err := c.checkResourceLimits(dir, m, sc, nil); err != nil {
		t.Fatalf("v3 should use same limit path as v1: %v", err)
	}
}

func TestCheckResourceLimits_UnsupportedAPIVersion(t *testing.T) {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	dir := filepath.Join("testdata", "resourcelimits_v1_inline")
	m, err := c.LoadManifestFile(dir)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		t.Fatal("expected *AppConfiguration")
	}
	cfg.APIVersion = "v0"
	sc := ownerScenario{owner: "alice", admin: "admin"}
	limErr := c.checkResourceLimits(dir, m, sc, nil)
	if limErr == nil {
		t.Fatal("expected unsupported apiVersion error")
	}
	if !strings.Contains(limErr.Error(), "不支持该版本") {
		t.Fatalf("expected 不支持该版本 in error, got: %v", limErr)
	}
}

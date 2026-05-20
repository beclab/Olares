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

// TestCheckResourceLimits_V2_SkipsRegardlessOfManifestVersion pins down the
// contract that apiVersion=v2 always short-circuits the limit check before
// either the legacy or the modern branch can fire. It reuses the modern
// "toobig" fixture (whose containers exceed the inline spec.requiredCpu /
// spec.requiredMemory and would normally fail on v1) and:
//   - first verifies v2 + modern returns nil,
//   - then flips ConfigVersion to legacy ("0.11.0") and clears
//     spec.resources[], so absent the v2 short-circuit the legacy branch
//     would run resources.CheckResourceLimits against zero limits and fail
//     for every container that declares any CPU/memory request. The
//     check must still return nil.
func TestCheckResourceLimits_V2_SkipsRegardlessOfManifestVersion(t *testing.T) {
	c := New(WithOwner("alice"), WithAdmin("admin"))
	dir := filepath.Join("testdata", "resourcelimits_v1_inline_toobig")
	m, err := c.LoadManifestFile(dir)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	cfg, ok := m.Raw().(*manifest.AppConfiguration)
	if !ok {
		t.Fatal("expected *AppConfiguration")
	}
	sc := ownerScenario{owner: "alice", admin: "admin"}

	cfg.APIVersion = manifest.APIVersionV2
	if err := c.checkResourceLimits(dir, m, sc, nil); err != nil {
		t.Fatalf("v2 + modern must skip the limit check, got: %v", err)
	}

	cfg.ConfigVersion = "0.11.0"
	cfg.Spec.Resources = nil
	if err := c.checkResourceLimits(dir, m, sc, nil); err != nil {
		t.Fatalf("v2 + legacy must skip the limit check, got: %v", err)
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
	if !strings.Contains(limErr.Error(), "not supported version") {
		t.Fatalf("expected not supported version in error, got: %v", limErr)
	}
}

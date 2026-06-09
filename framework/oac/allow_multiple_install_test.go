package oac

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
)

func TestLint_AllowMultipleInstall_ClusterScopedFixedName_Bad(t *testing.T) {
	err := Lint("testdata/multiclusterbad",
		WithOwnerAdmin("alice"),
		SkipResourceCheck(),
		SkipHostPathCheck(),
	)
	if err == nil {
		t.Fatal("expected Lint to fail: fixed cluster-scoped name with allowMultipleInstall")
	}
	if !strings.Contains(err.Error(), "fixed-multicluster-role") {
		t.Fatalf("error should mention the fixed ClusterRole name, got: %v", err)
	}
}

func TestLint_AllowMultipleInstall_ClusterScopedDynamicName_OK(t *testing.T) {
	err := Lint("testdata/multiclusterok",
		WithOwnerAdmin("alice"),
		SkipResourceCheck(),
		SkipHostPathCheck(),
	)
	if err != nil {
		t.Fatalf("Lint(multiclusterok): %v", err)
	}
}

// TestLint_AllowMultipleInstall_FixedWorkloadName_Bad covers the
// release-name workload check: when options.allowMultipleInstall=true the
// chart must declare at least one Deployment/StatefulSet whose
// metadata.name is templated on `{{ .Release.Name }}`. The fixture's
// Deployment uses a fixed name so the new check must fire; the existing
// cluster-scoped fixed-name check still passes because the ClusterRole
// is properly release-scoped.
func TestLint_AllowMultipleInstall_FixedWorkloadName_Bad(t *testing.T) {
	err := Lint("testdata/multiclusterbadworkload",
		WithOwnerAdmin("alice"),
		SkipResourceCheck(),
		SkipHostPathCheck(),
	)
	if err == nil {
		t.Fatal("expected Lint to fail: allowMultipleInstall without a {{ .Release.Name }} workload")
	}
	if !strings.Contains(err.Error(), "options.allowMultipleInstall=true requires at least one Deployment or StatefulSet") {
		t.Fatalf("error should flag the missing {{ .Release.Name }} workload, got: %v", err)
	}
}

// TestReleaseNameWorkloadCheckApplies pins the gate to its current
// definition: the release-name workload rule fires whenever
// options.allowMultipleInstall is true, regardless of apiVersion. The
// flag (not the schema version) is what enables multiple coexisting
// installs and therefore demands a release-scoped primary workload.
func TestReleaseNameWorkloadCheckApplies(t *testing.T) {
	cases := []struct {
		api   string
		allow bool
		want  bool
	}{
		{"v1", true, true},
		{"v3", true, true},
		{"v2", true, true},
		{"", true, true},
		{"v1", false, false},
		{"v2", false, false},
		{"v3", false, false},
	}
	for _, tc := range cases {
		cfg := &manifest.AppConfiguration{
			APIVersion: tc.api,
			Options:    manifest.Options{AllowMultipleInstall: tc.allow},
		}
		if got := releaseNameWorkloadCheckApplies(cfg); got != tc.want {
			t.Errorf("api=%q allow=%v: got %v want %v", tc.api, tc.allow, got, tc.want)
		}
	}
}

func TestAllowMultipleInstallClusterScopedCheckApplies(t *testing.T) {
	cases := []struct {
		api    string
		allow  bool
		want   bool
		reason string
	}{
		{"v1", true, true, "v1 + allowMultipleInstall"},
		{"v3", true, true, "v3 + allowMultipleInstall"},
		{"v2", true, false, "v2 skipped"},
		{"v1", true, true, "flag off"},
		{"", true, true, "empty apiVersion defaults to v1"},
	}
	for _, tc := range cases {
		cfg := &manifest.AppConfiguration{
			APIVersion: tc.api,
			Options:    manifest.Options{AllowMultipleInstall: tc.allow},
		}
		got := allowClusterScopedCheckApplies(cfg)
		if got != tc.want {
			t.Errorf("api=%q allow=%v: got %v want %v (%s)", tc.api, tc.allow, got, tc.want, tc.reason)
		}
	}
}

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

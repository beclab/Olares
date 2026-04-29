package oac

import "testing"

func TestIsNewOlaresManifestVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		want    bool
	}{
		{"empty is legacy", "", false},
		{"malformed is legacy", "not-a-semver", false},
		{"0.7.2 is legacy", "0.7.2", false},
		{"0.11.99 is legacy", "0.11.99", false},
		{"0.12.0 boundary is new", "0.12.0", true},
		{"0.12.1 is new", "0.12.1", true},
		{"0.13.0 is new", "0.13.0", true},
		{"1.0.0 is new", "1.0.0", true},
		{"v-prefixed 0.12.0 is new", "v0.12.0", true},
		{"v-prefixed 0.11.0 is legacy", "v0.11.0", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNewOlaresManifestVersion(tc.version); got != tc.want {
				t.Errorf("IsNewOlaresManifestVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}

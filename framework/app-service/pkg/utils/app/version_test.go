package app

import (
	"strings"
	"testing"
)

func TestIsDowngrade(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		deployed    string
		expected    bool
		errContains string
	}{
		{
			name:     "newer target upgrades",
			target:   "0.1.3",
			deployed: "0.1.2",
			expected: false,
		},
		{
			name:     "reinstalling the deployed version is not a downgrade",
			target:   "0.1.2",
			deployed: "0.1.2",
			expected: false,
		},
		{
			name:     "older target is a downgrade",
			target:   "0.1.2",
			deployed: "0.1.3",
			expected: true,
		},
		{
			name:     "older major is a downgrade",
			target:   "1.9.9",
			deployed: "2.0.0",
			expected: true,
		},
		{
			name:     "an empty target asks for the newest available",
			target:   "",
			deployed: "0.1.2",
			expected: false,
		},
		{
			name:        "an empty deployed version returns an error",
			target:      "0.1.2",
			deployed:    "",
			errContains: "parse deployed version",
		},
		{
			name:        "an invalid target version returns an error",
			target:      "latest",
			deployed:    "0.1.2",
			errContains: "parse target version",
		},
		{
			name:        "an invalid deployed version returns an error",
			target:      "0.1.2",
			deployed:    "nightly",
			errContains: "parse deployed version",
		},
		{
			name:     "a two-part version is semver enough to compare",
			target:   "1.0",
			deployed: "1.1",
			expected: true,
		},
		{
			name:     "a prerelease ahead of the deployed version upgrades",
			target:   "0.2.0-rc.1",
			deployed: "0.1.9",
			expected: false,
		},
		{
			name:     "a prerelease of the deployed version is a downgrade",
			target:   "0.2.0-rc.1",
			deployed: "0.2.0",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := IsDowngrade(test.target, test.deployed)
			if test.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), test.errContains) {
					t.Fatalf("IsDowngrade(%q, %q) error = %v, want %q", test.target, test.deployed, err, test.errContains)
				}
			} else if err != nil {
				t.Fatalf("IsDowngrade(%q, %q) unexpected error: %v", test.target, test.deployed, err)
			}
			if got != test.expected {
				t.Errorf("IsDowngrade(%q, %q) = %v, want %v", test.target, test.deployed, got, test.expected)
			}
		})
	}
}

package app

import "testing"

func TestIsDowngrade(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		deployed string
		expected bool
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
			name:     "an unknown deployed version cannot be compared against",
			target:   "0.1.2",
			deployed: "",
			expected: false,
		},
		{
			name:     "a target that is not semver is left alone",
			target:   "latest",
			deployed: "0.1.2",
			expected: false,
		},
		{
			name:     "a deployed version that is not semver is left alone",
			target:   "0.1.2",
			deployed: "nightly",
			expected: false,
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
			if got := IsDowngrade(test.target, test.deployed); got != test.expected {
				t.Errorf("IsDowngrade(%q, %q) = %v, want %v", test.target, test.deployed, got, test.expected)
			}
		})
	}
}

package utils

import "testing"

// CompareOlaresVersion mirrors the JS implementation 1:1 (including its
// peculiar dash-prerelease handling). These cases are deliberately chosen to
// lock in the same behavior so a future "fix" doesn't silently drift away
// from what the SPA / backend expect.
func TestCompareOlaresVersion(t *testing.T) {
	cases := []struct {
		v0, v1 string
		want   int
	}{
		// Dotted numerics — straightforward.
		{"1.12.0", "1.12.0", 0},
		{"1.12.1", "1.12.0", 1},
		{"1.12.0", "1.12.1", -1},
		{"1.13.0", "1.12.99", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.12.6", "1.12.5", 1},
		{"1.12.5", "1.12.6", -1},

		// Missing parts treated as zero.
		{"1.12", "1.12.0", 0},
		{"1.12", "1.12.0.1", -1},

		// "longer split-on-dash wins -1" — exactly one side has a
		// dash-qualifier. The plain side is reported as newer.
		{"1.12.0-0", "1.12.0", -1},
		{"1.12.0", "1.12.0-0", 1},

		// Two non-rc prereleases compare numerically on the dash tail.
		{"1.12.0-2", "1.12.0-1", 1},
		{"1.12.0-1", "1.12.0-2", -1},
		{"1.12.0-1", "1.12.0-1", 0},

		// Two rc prereleases compare numerically on rc.<n>.
		{"1.12.0-rc.2", "1.12.0-rc.1", 1},
		{"1.12.0-rc.1", "1.12.0-rc.2", -1},
		{"1.12.0-rc.2", "1.12.0-rc.2", 0},

		// Mixed: when only one side is rc, the rc side is reported as
		// newer (return value mirrors the JS "compare" field — yes, it is
		// counter-intuitive vs. real-world semver, but it is what the
		// upstream returns and we must match it bit-for-bit).
		{"1.12.0-1", "1.12.0-rc.5", -1},
		{"1.12.0-rc.5", "1.12.0-1", 1},
	}
	for _, c := range cases {
		got := CompareOlaresVersion(c.v0, c.v1)
		if got != c.want {
			t.Errorf("CompareOlaresVersion(%q, %q) = %d, want %d", c.v0, c.v1, got, c.want)
		}
	}
}

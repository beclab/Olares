package utils

import (
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func ParseOlaresVersionString(versionString string) (*semver.Version, error) {
	// todo: maybe some other custom processing only for olares
	return semver.NewVersion(versionString)
}

// CompareOlaresVersion mirrors @bytetrade/core's compareOlaresVersion (the
// same comparison the SPA and user-service use). Returns -1, 0, or 1 with
// the same edge cases as the JS implementation:
//   - if exactly one side carries a "-<prerelease>" qualifier, the side
//     without one is *newer* (compare = -1 for v0)
//   - if both sides carry one, "rc.<n>" prereleases are compared numerically
//     and a non-rc prerelease beats an rc one
//   - dotted numeric parts are compared left-to-right; missing parts are
//     treated as 0
//
// The peculiar "longer split-on-dash wins -1" rule is non-obvious but matches
// the JS source 1:1 — keep it that way so cross-stack comparisons agree. This
// is the single source of truth for Olares OS version comparison in the CLI:
// command-side version branching (e.g. "is the backend >= 1.12.6?") and the
// password-salt threshold check both call it.
//
// Reference (ground truth): user-service node_modules/@bytetrade/core's
// compareOlaresVersion, and the SPA's apps/packages/app/src/utils/account.ts.
func CompareOlaresVersion(v0, v1 string) int {
	v0 = strings.TrimSpace(v0)
	v1 = strings.TrimSpace(v1)
	v0Splits := strings.SplitN(v0, "-", 2)
	v1Splits := strings.SplitN(v1, "-", 2)

	if len(v0Splits) != len(v1Splits) {
		if len(v0Splits) > len(v1Splits) {
			return -1
		}
		return 1
	}

	if len(v0Splits) > 1 {
		v0pre := v0Splits[1]
		v1pre := v1Splits[1]
		v0Rc := strings.HasPrefix(v0pre, "rc")
		v1Rc := strings.HasPrefix(v1pre, "rc")
		switch {
		case v0Rc && v1Rc:
			if v0pre != v1pre {
				v0n, _ := strconv.Atoi(strings.TrimPrefix(v0pre, "rc."))
				v1n, _ := strconv.Atoi(strings.TrimPrefix(v1pre, "rc."))
				if v0n > v1n {
					return 1
				}
				return -1
			}
		case !v0Rc && !v1Rc:
			v0n, _ := strconv.Atoi(v0pre)
			v1n, _ := strconv.Atoi(v1pre)
			if v0n != v1n {
				if v0n > v1n {
					return 1
				}
				return -1
			}
		default:
			// exactly one of them is rc — non-rc wins.
			if v0Rc {
				return 1
			}
			return -1
		}
	}

	v0Parts := splitDotInts(v0Splits[0])
	v1Parts := splitDotInts(v1Splits[0])
	maxLen := len(v0Parts)
	if len(v1Parts) > maxLen {
		maxLen = len(v1Parts)
	}
	for i := 0; i < maxLen; i++ {
		a := getOrZero(v0Parts, i)
		b := getOrZero(v1Parts, i)
		if a > b {
			return 1
		}
		if a < b {
			return -1
		}
	}
	return 0
}

func splitDotInts(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, _ := strconv.Atoi(p)
		out[i] = n
	}
	return out
}

func getOrZero(xs []int, i int) int {
	if i >= len(xs) {
		return 0
	}
	return xs[i]
}

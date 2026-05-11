package me

import (
	"crypto/md5" // #nosec G401, G501 -- mandated by the SPA's hashing scheme; matches user-service.
	"encoding/hex"
	"strconv"
	"strings"
)

// Olares' SPA salts user passwords with a fixed suffix and hashes the
// result with MD5 before sending them over the wire — but only when the
// running OS version is at least 1.12.0-0 (older versions accept the raw
// password). The CLI must mirror that scheme exactly when calling the
// password-change endpoint, otherwise authentication will fail silently
// at the upstream check.
//
// Reference (SPA, ground truth):
//   apps/packages/app/src/utils/salted-md5.ts:8     — saltedMD5
//   apps/packages/app/src/utils/account.ts:209      — passwordAddSort
//   user-service/node_modules/@bytetrade/core/dist/index.js:59 — compareOlaresVersion
//
// We re-implement compareOlaresVersion + saltedMD5 here in Go rather
// than depending on a shared CGo wrapper (there is none) or shelling
// out to node, both of which would add weight just to MD5 a string.

const (
	passwordSaltSuffix    = "@Olares2025"
	passwordSaltApplyFrom = "1.12.0-0"
)

// compareOlaresVersion mirrors @bytetrade/core's compareOlaresVersion.
// Returns -1, 0, or 1 with the same edge cases as the JS implementation:
//   - if exactly one side carries a "-<prerelease>" qualifier, the side
//     without one is *newer* (compare = -1 for v0)
//   - if both sides carry one, "rc.<n>" prereleases are compared numerically
//     and a non-rc prerelease beats an rc one
//   - dotted numeric parts are compared left-to-right; missing parts are
//     treated as 0
//
// The peculiar "longer split-on-dash wins -1" rule is non-obvious but
// matches the JS source 1:1 — keep it that way so cross-stack comparisons
// agree.
func compareOlaresVersion(v0, v1 string) int {
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

// saltedPassword applies the SPA's hashing scheme to a raw password.
// When osVersion is empty (we couldn't read /api/olares-info) or older
// than passwordSaltApplyFrom, the password is sent as-is to match the
// SPA's saltedMD5 fallback. Otherwise it returns md5(password+suffix)
// in lowercase hex.
//
// We deliberately swallow the version-comparison error path (any
// malformed version string falls into the "older than 1.12.0-0" branch
// since splitDotInts treats non-digits as zero) — the upstream BFL
// rejects unsalted-against-salted equally well, so the worst case is a
// "incorrect current password" error message rather than a corrupted
// account.
func saltedPassword(password, osVersion string) string {
	if osVersion == "" {
		return password
	}
	if compareOlaresVersion(osVersion, passwordSaltApplyFrom) < 0 {
		return password
	}
	sum := md5.Sum([]byte(password + passwordSaltSuffix)) // #nosec G401 -- protocol mandate
	return hex.EncodeToString(sum[:])
}

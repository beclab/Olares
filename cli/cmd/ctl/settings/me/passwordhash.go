package me

import (
	"crypto/md5" // #nosec G401, G501 -- mandated by the SPA's hashing scheme; matches user-service.
	"encoding/hex"

	"github.com/beclab/Olares/cli/pkg/utils"
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
// The version comparison itself lives in pkg/utils.CompareOlaresVersion —
// the single source of truth shared with command-side version branching —
// so we only re-implement the saltedMD5 half here.

const (
	passwordSaltSuffix    = "@Olares2025"
	passwordSaltApplyFrom = "1.12.0-0"
)

// saltedPassword applies the SPA's hashing scheme to a raw password.
// When osVersion is empty (we couldn't read /api/olares-info) or older
// than passwordSaltApplyFrom, the password is sent as-is to match the
// SPA's saltedMD5 fallback. Otherwise it returns md5(password+suffix)
// in lowercase hex.
//
// We deliberately swallow the version-comparison error path (any
// malformed version string falls into the "older than 1.12.0-0" branch
// since the comparison treats non-digits as zero) — the upstream BFL
// rejects unsalted-against-salted equally well, so the worst case is a
// "incorrect current password" error message rather than a corrupted
// account.
func saltedPassword(password, osVersion string) string {
	if osVersion == "" {
		return password
	}
	if utils.CompareOlaresVersion(osVersion, passwordSaltApplyFrom) < 0 {
		return password
	}
	sum := md5.Sum([]byte(password + passwordSaltSuffix)) // #nosec G401 -- protocol mandate
	return hex.EncodeToString(sum[:])
}

// SaltedPassword exports saltedPassword for sibling settings packages that
// call the same /api/users wire contract as the SPA (e.g. settings users).
func SaltedPassword(password, osVersion string) string {
	return saltedPassword(password, osVersion)
}

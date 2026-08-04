package manifest

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// reservedSystemAppIDs enumerates the metadata.appid values that the Olares
// platform reserves for its built-in system apps. The list mirrors the
// SYS_APPS environment value baked into the app-service deployment
// (framework/app-service/.olares/config/cluster/deploy/appservice_deploy.yaml)
// so a third-party app cannot claim an identity that would collide with a
// system app and silently shadow it at install / routing time.
//
// Membership is exact (case-sensitive after trimming); the platform compares
// the raw appid string when deciding whether an install is a system upgrade,
// so we mirror that contract here rather than fuzzy-matching.
var reservedSystemAppIDs = map[string]struct{}{
	"market":         {},
	"auth":           {},
	"citus":          {},
	"desktop":        {},
	"did":            {},
	"docs":           {},
	"files":          {},
	"fsnotify":       {},
	"headscale":      {},
	"infisical":      {},
	"intentprovider": {},
	"ksserver":       {},
	"message":        {},
	"mongo":          {},
	"monitoring":     {},
	"notifications":  {},
	"profile":        {},
	"redis":          {},
	"recommend":      {},
	"seafile":        {},
	"search":         {},
	"search-admin":   {},
	"settings":       {},
	"systemserver":   {},
	"tapr":           {},
	"vault":          {},
	"video":          {},
	"zinc":           {},
	"accounts":       {},
	"control-hub":    {},
	"dashboard":      {},
	"nitro":          {},
	"olares-app":     {},
}

// IsReservedSystemAppID reports whether s collides with a reserved system
// appid. Surrounding whitespace is trimmed before lookup so a manifest that
// accidentally writes "  market  " is treated the same as "market", matching
// what the YAML parser would observe after the value goes through
// downstream string normalization.
func IsReservedSystemAppID(s string) bool {
	_, ok := reservedSystemAppIDs[strings.TrimSpace(s)]
	return ok
}

// AppIDFromName returns the deterministic appid the Olares platform derives
// from an app name: the first 8 hex characters of md5(name). It mirrors
// framework/app-service/pkg/appcfg.(AppName).GetAppID's user-app branch
// (and framework/app-service/pkg/utils/app.GetAppID) so the loader's
// normalization and the runtime's runtime-derived id agree byte-for-byte.
//
// An empty name yields an empty string -- the loader's normalization is a
// no-op when metadata.name is missing, which keeps validation errors keyed
// off the absent name rather than producing a meaningless hash.
func AppIDFromName(name string) string {
	if name == "" {
		return ""
	}
	sum := md5.Sum([]byte(name))
	return hex.EncodeToString(sum[:])[:8]
}

package manifest

import (
	"os"
	"strings"
)

// loginOlaresCLIAllowListEnv is the ConfigMap-backed override for which
// metadata.name values may declare permission.loginOlaresCLI. Comma-separated,
// whitespace around each name is ignored. An unset or empty value falls back
// to loginOlaresCLIAllowlist.
const loginOlaresCLIAllowListEnv = "LOGIN_OLARES_CLI_ALLOW_LIST"

// loginOlaresCLIAllowlist enumerates the metadata.name values permitted to
// declare permission.loginOlaresCLI when LOGIN_OLARES_CLI_ALLOW_LIST is not
// set.
//
// The permission makes app-service mint a ten-year, non-rotating Olares
// refresh token for the installing user and mount it into every container of
// the app. Anything running in that pod — including an AI agent acting on the
// user's behalf — can read the credential, and it stays valid until the app is
// uninstalled or the grant is revoked. That blast radius is why the field is
// allowlisted here rather than being self-service: membership is a review
// decision about whether the app is trusted with a standing user identity, not
// a manifest-authoring choice.
//
// Cluster operators override the list by editing the os-framework ConfigMap
// that feeds LOGIN_OLARES_CLI_ALLOW_LIST; a non-empty env value replaces this
// map entirely rather than merging with it.
var loginOlaresCLIAllowlist = map[string]struct{}{
	"lares": {},
}

// IsLoginOlaresCLIAllowed reports whether name may declare
// permission.loginOlaresCLI. Surrounding whitespace is trimmed before lookup
// so a manifest that writes "  lares  " is treated like "lares", matching
// IsReservedSystemAppID's normalization.
func IsLoginOlaresCLIAllowed(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	_, ok := loginOlaresCLIAllowlistForLookup()[name]
	return ok
}

func loginOlaresCLIAllowlistForLookup() map[string]struct{} {
	if parsed, ok := parseLoginOlaresCLIAllowList(os.Getenv(loginOlaresCLIAllowListEnv)); ok {
		return parsed
	}
	return loginOlaresCLIAllowlist
}

// parseLoginOlaresCLIAllowList returns the env-provided set and true when the
// raw value contains at least one name. Whitespace-only or empty input is a
// miss so the compiled default still applies.
func parseLoginOlaresCLIAllowList(raw string) (map[string]struct{}, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out[part] = struct{}{}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

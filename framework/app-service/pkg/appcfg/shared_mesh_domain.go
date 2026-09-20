package appcfg

import "strings"

// SharedMeshDomain is the fixed DNS suffix for gateway sharedEntrances
// (SRR hostPatterns and in-cluster chart URLs). Application entrances keep
// using cluster.GetPlatformDomain(). UI GenSharedEntranceURL is owned by a
// separate change.
const SharedMeshDomain = "olares.com"

// SharedZone returns "shared." + SharedMeshDomain.
func SharedZone() string {
	return "shared." + SharedMeshDomain
}

// IsSharedMeshExactHost reports whether host is an exact shared mesh FQDN
// "{label}.shared.olares.com" (no wildcards). Used by mesh-in allow-lists so
// contract hosts survive when platformDomain is not olares.com.
func IsSharedMeshExactHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || strings.Contains(host, "*") {
		return false
	}
	suffix := "." + SharedZone()
	return host != SharedZone() && strings.HasSuffix(host, suffix)
}

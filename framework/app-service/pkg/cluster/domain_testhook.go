package cluster

// PrimePlatformDomainForTest seeds the platform domain for unit tests.
func PrimePlatformDomainForTest(domain string) {
	platformDomainTestOverride = domain
	platformDomainTestOverrideSet = true
	defaultCache.set(domain)
}

// ResetPlatformDomainForTest clears the platform domain test override and cache.
func ResetPlatformDomainForTest() {
	platformDomainTestOverrideSet = false
	platformDomainTestOverride = ""
	defaultCache.mu.Lock()
	defer defaultCache.mu.Unlock()
	defaultCache.loaded = false
}

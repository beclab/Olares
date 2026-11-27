export function isValidDomain(domain: string) {
	if (domain.length === 0 || domain.length > 253) return false;

	const domainRegex =
		/^(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?))*$/;

	return domainRegex.test(domain);
}

// A custom route ID becomes a single DNS label (e.g. `<id>.<olaresId>`), so it
// must follow DNS label rules: lowercase letters/digits/hyphens, start with a
// letter, not end with a hyphen, 1-63 chars. Chinese/special chars are rejected
// because they cannot be resolved as a subdomain.
export function isValidThirdLevelDomain(routeId: string) {
	if (routeId.length === 0 || routeId.length > 63) return false;

	const labelRegex = /^[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

	return labelRegex.test(routeId);
}

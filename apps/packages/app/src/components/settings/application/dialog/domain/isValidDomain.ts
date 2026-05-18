export function isValidDomain(domain: string) {
	if (domain.length === 0 || domain.length > 253) return false;

	const domainRegex =
		/^(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?))*$/;

	return domainRegex.test(domain);
}

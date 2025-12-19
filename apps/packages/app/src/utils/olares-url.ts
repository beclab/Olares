export function replaceLastSubdomain(newSubdomain: string): string {
	const host = document.location.host;

	const parts = host.split('.');
	if (parts.length > 1) {
		parts[0] = newSubdomain;
		return parts.join('.');
	}
	return host;
}

/**
 * Canonicalize a files route/full-path so that two paths pointing to the same
 * location compare equal regardless of trailing slashes, URL-encoding or
 * query-parameter ordering.
 *
 * Shared by the route<->store sync guard (`isRouteSyncedToFilesStore`) and the
 * browser back/forward back-stack reconciliation (`syncBrowserHistoryUrl`) so
 * both compare URLs the exact same way.
 */
export function canonicalizeFilesPath(fullPath: string): string {
	const queryIndex = fullPath.indexOf('?');
	let base = queryIndex === -1 ? fullPath : fullPath.slice(0, queryIndex);
	const query = queryIndex === -1 ? '' : fullPath.slice(queryIndex + 1);

	try {
		base = decodeURIComponent(base);
	} catch (error) {
		/* keep raw base when it is not a valid encoded string */
	}
	base = base.replace(/\/+$/, '');

	const params = new URLSearchParams(query);
	const normalizedQuery = [...params.entries()]
		.sort((a, b) => a[0].localeCompare(b[0]))
		.map(([key, value]) => `${key}=${value}`)
		.join('&');

	return normalizedQuery ? `${base}?${normalizedQuery}` : base;
}

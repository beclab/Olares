export function encodeUrl(url: string): string {
	if (!url) {
		return '';
	}

	const encodedPath = url
		.split('/')
		.map((part) => encodeURIComponent(part))
		.join('/');

	return encodedPath;
}

export function decodeUrl(url: string): string {
	if (!url) {
		return '';
	}

	const encodedPath = url
		.split('/')
		.map((part) => {
			try {
				return decodeURIComponent(part);
			} catch (error) {
				return part;
			}
		})
		.join('/');

	return encodedPath;
}

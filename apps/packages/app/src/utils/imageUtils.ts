export function getRequireImage(path: string): string {
	if (!path) {
		return '';
	}
	if (path.startsWith('http')) {
		return path;
	}
	return require(`../assets/${path}`);
}

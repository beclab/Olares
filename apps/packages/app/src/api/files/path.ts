export const appendPath = (path: string, ...appends: string[]) => {
	if (!appends || appends.length == 0) return path;

	if (appends.filter((e) => e.length != 0).length == 0) {
		return path;
	}

	if (!path.endsWith('/')) {
		path = path + '/';
	}

	const appendSplit = () => {
		if (!path.endsWith('/')) {
			path = path + '/';
		}
	};

	appendSplit();

	for (let index = 0; index < appends.length; index++) {
		const p = appends[index];
		if (!p || p.length == 0) {
			continue;
		}
		appendSplit();
		if (p == '/') {
			continue;
		}
		path = path + (p.startsWith('/') ? p.substring(1) : p);
	}

	return path;
};

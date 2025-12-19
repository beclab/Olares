export const isValidImageUrl = (url: string): boolean => {
	if (!url || typeof url !== 'string') {
		return false;
	}

	try {
		new URL(url);
	} catch {
		return false;
	}

	const imageExtensions = [
		'.jpg',
		'.jpeg',
		'.png',
		'.gif',
		'.bmp',
		'.webp',
		'.svg'
	];
	return imageExtensions.some((ext) => url.toLowerCase().endsWith(ext));
};

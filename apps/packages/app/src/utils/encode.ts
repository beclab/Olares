import { Base64 } from 'js-base64';

export function stringToBase64(text: string, urlSafe = true): string {
	return Base64.encode(text, urlSafe);
}

export function base64ToString(text: string): string {
	return Base64.decode(text);
}

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

export const decodeURIComponentSafe = (path: string): string => {
	try {
		return decodeURIComponent(path);
	} catch (error) {
		return path;
	}
};

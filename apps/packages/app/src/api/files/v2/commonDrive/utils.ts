import { CommonFetch } from '../../fetch';
import { encodeUrl } from 'src/utils/encode';
import { CommonUrlApiType, commonUrlPrefix } from '../common/utils';
import { appendPath } from '../path';

export function formatResourcesUrl(url: string) {
	const newUrl = commonRemovePrefix(url);
	return commonCommonUrl('resources', newUrl);
}

export async function createDir(url) {
	await CommonFetch.post(formatResourcesUrl(url));
}

export async function remove(url) {
	await CommonFetch.delete(formatResourcesUrl(url));
}

export async function put(url, content = '') {
	return await CommonFetch.put(formatResourcesUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

export function commonRemovePrefix(url: string) {
	if (!url.startsWith('/Common') && !url.startsWith('/drive/Common')) {
		return url;
	}
	if (url.startsWith('/Common')) return url.slice(7);
	return url.slice(13);
}

export const commonCommonUrl = (type: CommonUrlApiType, path: string) => {
	return (
		commonUrlPrefix(type) +
		'drive/Common' +
		(path.startsWith('/') ? path : '/' + path)
	);
};

export const formatPathtoUrl = (path: string) => {
	path = commonRemovePrefix(path);
	return appendPath('/drive/Common', path);
};

export const displayPath = (file: {
	isDir: boolean;
	fileExtend?: string;
	path: string;
	fileType?: string;
}) => {
	return appendPath(
		'/',
		file.fileExtend || '',
		encodeUrl(file.path),
		file.isDir ? '/' : ''
	);
};

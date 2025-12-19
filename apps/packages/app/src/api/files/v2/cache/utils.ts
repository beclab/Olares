import { getAppDataPath } from 'src/utils/file';
import { CommonFetch } from '../../fetch';
import { appendPath } from '../path';
import { CommonUrlApiType, commonUrlPrefix } from '../common/utils';
import { encodeUrl } from 'src/utils/encode';

export function formatResourcesUrl(url: string) {
	const { path, node } = getAppDataPath(url);
	return cacheCommonUrl('resources', path, node);
}

export async function createDir(url) {
	await CommonFetch.post(formatResourcesUrl(url));
}

export async function remove(url) {
	return CommonFetch.delete(formatResourcesUrl(url));
}

export async function put(url, content = '') {
	return CommonFetch.put(formatResourcesUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

export function cacheRemovePrefix(url: string) {
	url = cacheRemoveCachePrefix(url);
	return url;
}

export function cacheRemoveCachePrefix(url: string) {
	if (!url.startsWith('/Cache') && !url.startsWith('/cache')) {
		return url;
	}
	return url.slice(6);
}

export const cacheCommonUrl = (
	type: CommonUrlApiType,
	path: string,
	node?: string
) => {
	return appendPath(commonUrlPrefix(type), 'cache', node ? node : '', path);
};

export const formatPathtoUrl = (path: string) => {
	path = cacheRemovePrefix(path);
	return appendPath('/cache', path);
};

export const displayPath = (file: {
	isDir: boolean;
	fileExtend?: string;
	path: string;
	fileType?: string;
}) => {
	return appendPath(
		'/Cache',
		file.fileExtend || '',
		encodeUrl(file.path),
		file.isDir ? '/' : ''
	);
};

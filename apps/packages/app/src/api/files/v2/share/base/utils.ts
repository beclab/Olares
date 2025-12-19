import { CommonFetch } from '../../../fetch';
import { encodeUrl } from 'src/utils/encode';
import { CommonUrlApiType, commonUrlPrefix } from '../../common/utils';
import { appendPath } from '../../path';

export function formatResourcesUrl(url: string) {
	const { path, path_id } = getShareDataPath(url);
	return shareCommonUrl('resources', path_id, path);
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

export function shareRemovePrefix(url: string) {
	if (!url.startsWith('/Share')) return url;
	return url.slice(6);
}

export const shareCommonUrl = (
	type: CommonUrlApiType,
	path_id: string,
	path: string
) => {
	return appendPath(commonUrlPrefix(type), 'share', path_id, path);
};

export const formatPathtoUrl = (path: string) => {
	path = shareRemovePrefix(path);
	return appendPath('/share', path);
};

export const displayPath = (file: {
	isDir: boolean;
	fileExtend?: string;
	path: string;
	fileType?: string;
}) => {
	return appendPath(
		'/Share',
		file.fileExtend || '',
		encodeUrl(file.path),
		file.isDir ? '/' : ''
	);
};

export const displaySharePath = (file: { id: string }) => {
	return appendPath('/Share', file.id, '/');
};

export function getShareDataPath(url: string) {
	const res = url.split('/');
	if (res[1] != 'Share' && res[1] != 'share') {
		throw Error('Invalid AppData path');
	}
	const path_id = res[2];
	let path = '';
	for (let i = 3; i < res.length; i++) {
		path = path + '/';
		path = path + res[i];
	}

	return { path_id, path };
}

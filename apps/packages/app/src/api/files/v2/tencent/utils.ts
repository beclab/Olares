import { useDataStore } from 'src/stores/data';
import { DriveType } from 'src/utils/interface/files';
import { CommonFetch } from '../../fetch';
import { CommonUrlApiType, commonUrlPrefix } from '../common/utils';
import { appendPath } from '../path';

export function formatResourcesUrl(url: string) {
	const newUrl = tencentRemovePrefix(url);
	return tencentCommonUrl('resources', newUrl);
}

export async function createDir(url) {
	await CommonFetch.post(formatResourcesUrl(url));
}

export async function remove(url: string) {
	return CommonFetch.delete(formatResourcesUrl(url));
}

export async function put(url: string, content = '') {
	return CommonFetch.put(formatResourcesUrl(url), content);
}

export function download(_format, files) {
	const name = files.path.split('/')[3];
	const path = '/' + files.path.split('/').slice(4).join('/');
	return generateDownloadUrl(files.driveType, path, name);
}

export function generateDownloadUrl(
	driveType: DriveType,
	path: string,
	name: string
) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	return `${baseURL}/drive/download_sync_stream?drive=${driveType}&cloud_file_path=${path}&name=${name}`;
}

export function tencentRemovePrefix(url: string) {
	url = tencentRemoveHomePrefix(url);
	return url;
}

export function tencentRemoveHomePrefix(url: string) {
	if (!url.startsWith('/Drive/tencent')) {
		return url;
	}
	return url.slice(14);
}

export const tencentCommonUrl = (type: CommonUrlApiType, path: string) => {
	return (
		commonUrlPrefix(type) +
		'tencent' +
		(path.startsWith('/') ? path : '/' + path)
	);
};

export const formatPathtoUrl = (path: string) => {
	path = tencentRemovePrefix(path);
	return appendPath('/tencent', path);
};

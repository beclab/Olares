import { CommonFetch } from '../../fetch';
import { removePrefix } from '../utils';
import { CommonUrlApiType, commonUrlPrefix } from '../common/utils';
import { useDataStore } from 'src/stores/data';
import { appendPath } from '../path';
import { encodeUrl } from 'src/utils/encode';

export function formatResourcesUrl(url: string) {
	const newUrl = externalRemovePrefix(url);
	return externalCommonUrl('resources', newUrl);
}

export async function remove(url) {
	await CommonFetch.delete(formatResourcesUrl(url));
}

export async function createDir(url) {
	await CommonFetch.post(formatResourcesUrl(url));
}

export async function put(url, content = '') {
	CommonFetch.put(formatResourcesUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

export function getDownloadUrl(format, ...files) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	if (files.length <= 0) {
		return '';
	} else if (files.length == 1) {
		return appendPath(
			baseURL,
			externalCommonUrl('raw', externalRemovePrefix(files[0]))
		);
	}

	let url = baseURL + externalCommonUrl('raw', '/');

	let arg = '';
	for (const file of files) {
		arg += encodeUrl(externalRemovePrefix(file)) + ',';
	}
	arg = arg.substring(0, arg.length - 1);
	arg = encodeURIComponent(arg);
	url += `/?files=${arg}&`;

	if (format) {
		url += `algo=${format}&`;
	}
	if (store.jwt) {
		url += `auth=${store.jwt}&`;
	}
	return url;
}

export function externalRemovePrefix(url: string) {
	url = removePrefix(url);
	url = externalRemoveHomePrefix(url);
	return url;
}

export function externalRemoveHomePrefix(url: string) {
	if (!url.startsWith('/External')) {
		return url;
	}
	return url.slice(9);
}

export const externalCommonUrl = (type: CommonUrlApiType, path: string) => {
	return (
		commonUrlPrefix(type) +
		'external' +
		(path.startsWith('/') ? path : '/' + path)
	);
};

export const formatPathtoUrl = (path: string) => {
	path = externalRemovePrefix(path);
	return appendPath('/external', path);
};

export const displayPath = (file: {
	isDir: boolean;
	fileExtend?: string;
	path: string;
	fileType?: string;
}) => {
	return appendPath(
		'/Files/External',
		file.fileExtend || '',
		encodeUrl(file.path),
		file.isDir ? '/' : ''
	);
};

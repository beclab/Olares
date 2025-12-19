import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { formatUrltoDriveType } from '../common/common';
import { getAppDataPath } from 'src/utils/file';
import { CommonFetch } from '../../fetch';
import { useDataStore } from 'src/stores/data';
import { removePrefix } from '../utils';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';

export function formatUrl(url: string) {
	return `/api/resources${url}`;
}

export async function createDir(url) {
	return CommonFetch.post(formatUrl(url));
}

export async function pasteAction(fromUrl): Promise<any> {
	const opts: any = {};

	let res: any;
	if (formatUrltoDriveType(fromUrl) === DriveType.Cache) {
		const { path, node } = getAppDataPath(fromUrl);

		if (node) {
			const headers = {
				auth: true,
				'X-Terminus-Node': node
			};

			const options = { headers: headers };

			res = await CommonFetch.patch(`/api/paste/AppData${path}`, {}, options);
		}
	} else {
		res = await CommonFetch.patch(`/api/paste${fromUrl}`, {}, opts);
	}

	if (res?.data?.split('\n')[1] === '413 Request Entity Too Large') {
		return BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: res.data.split('\n')[0]
		});
	}

	return res;
}

export async function remove(url) {
	CommonFetch.delete(formatUrl(url));
}

export async function put(url, content = '') {
	return CommonFetch.put(formatUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

export async function rename(from, to) {
	const url = `${from}?action=rename&destination=${to}&override=${false}&rename=${false}`;

	const res = await CommonFetch.patch(formatUrl(url));

	return res;
}

function moveCopy(items, copy = false, overwrite = false, rename = false) {
	const promises: any[] = [];

	for (const item of items) {
		const from = item.from;
		const to = item.to;

		const url = `${from}?action=${
			copy ? 'copy' : 'rename'
		}&destination=${to}&override=${overwrite}&rename=${rename}&src_type=${
			item.src_drive_type
		}&dst_type=${item.dst_drive_type}`;

		promises.push(pasteAction(url));
	}

	return Promise.all(promises);
}

export function move(items, overwrite = false, rename = false) {
	return moveCopy(items, false, overwrite, rename);
}

export function copy(items, overwrite = false, rename = false) {
	return moveCopy(items, true, overwrite, rename);
}

export function download(format, ...files) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	let url = `${baseURL}/api/raw`;

	if (files.length === 1) {
		url += removePrefix(files[0]) + '?';
	} else {
		let arg = '';

		for (const file of files) {
			arg += encodeUrl(removePrefix(file)) + ',';
		}

		arg = arg.substring(0, arg.length - 1);
		arg = encodeURIComponent(arg);
		url += `/?files=${arg}&`;
	}

	if (format) {
		url += `algo=${format}&`;
	}

	if (store.jwt) {
		url += `auth=${store.jwt}&`;
	}

	return url;
}

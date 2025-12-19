import { removePrefix } from '../utils';
import { getAppDataPath } from 'src/utils/file';
import { formatUrltoDriveType } from '../common/common';
import { useDataStore } from 'src/stores/data';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { CommonFetch } from '../../fetch';
import { DriveType } from 'src/utils/interface/files';

export function formatUrl(url: string) {
	const { path, node } = getAppDataPath(url);

	return `/api/cache/${node}/resources/AppData${path}`;
}

export async function createDir(url) {
	await CommonFetch.post(formatUrl(url));
}

export async function pasteAction(fromUrl, terminusNode): Promise<any> {
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
		if (terminusNode) {
			opts.headers = {
				'X-Terminus-Node': terminusNode
			};
		}

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
	return CommonFetch.delete(formatUrl(url));
}

export async function put(url, content = '') {
	return CommonFetch.put(formatUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

function moveCopy(items, copy = false, overwrite = false, rename = false) {
	const promises: any[] = [];

	for (const item of items) {
		const from = item.from;

		let to = item.to;
		let terminusNode = '';

		if (formatUrltoDriveType(item.to) === DriveType.Cache) {
			const { path, node } = getAppDataPath(item.to);
			to = `/AppData${path}`;
			terminusNode = node;
		}

		const url = `${from}?action=${
			copy ? 'copy' : 'rename'
		}&destination=${to}&override=${overwrite}&rename=${rename}&src_type=${
			item.src_drive_type
		}&dst_type=${item.dst_drive_type}`;

		console.log('urlurlurlurlurl', url);

		promises.push(pasteAction(url, terminusNode));
	}

	return Promise.all(promises);
}

export async function rename(from, to) {
	const enc_to = removePrefix(to);
	const url = `${from}?action=rename&destination=${enc_to}&override=${false}&rename=${false}`;
	const res = await CommonFetch.patch(formatUrl(url));

	return res;
}

export function move(items, overwrite = false, rename = false) {
	return moveCopy(items, false, overwrite, rename);
}

export function copy(items, overwrite = false, rename = false) {
	return moveCopy(items, true, overwrite, rename);
}

export function downloadZip(format, ...files) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	const node = files[0].split('/')[3];
	const replaceFiles = files[0].replace(`/${node}`, '');
	let url = `${baseURL}/api/cache/${node}/raw`;
	url += removePrefix(replaceFiles) + '?';

	if (format) {
		url += `algo=${format}&`;
	}

	if (store.jwt) {
		url += `auth=${store.jwt}&`;
	}

	return url;
}

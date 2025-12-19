import { fetchURL, removePrefix } from '../utils';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDataStore } from 'src/stores/data';
import { formatUrltoDriveType } from '../common/common';
import { getAppDataPath } from 'src/utils/file';
import { DriveType } from 'src/utils/interface/files';
import { CommonFetch } from '../../fetch';

export async function resourceAction(
	url: string,
	method: string,
	content?: any
) {
	// const newUrl = removePrefix(url);

	const opts: any = { method };

	if (content) {
		opts.headers = {
			'Content-Type': 'text/plain'
		};
		opts.data = content;
	}

	const res = await fetchURL(`/api/resources${url}`, opts);
	return res;
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

			res = await await CommonFetch.patch(
				`/api/paste/AppData${path}`,
				{},
				options
			);
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
	return resourceAction(url, 'DELETE');
}

export async function put(url, content = '') {
	return resourceAction(url, 'PUT', content);
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

		console.log('urlurlurlurlurl', url);

		promises.push(pasteAction(url));
	}

	return Promise.all(promises);
}

export async function rename(from, to) {
	console.log('fromfrom', from);
	console.log('toto', to);
	const enc_to = removePrefix(to);
	const url = `${from}?action=rename&destination=${enc_to}&override=${false}&rename=${false}&src=${
		DriveType.Dropbox
	}`;
	const res = await resourceAction(url, 'PATCH');
	return res;
}

export function move(items, overwrite = false, rename = false) {
	return moveCopy(items, false, overwrite, rename);
}

export function copy(items, overwrite = false, rename = false) {
	return moveCopy(items, true, overwrite, rename);
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

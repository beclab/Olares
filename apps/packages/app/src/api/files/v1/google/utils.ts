import { fetchURL } from '../utils';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDataStore } from 'src/stores/data';
import { useFilesStore } from 'src/stores/files';
import { formatUrltoDriveType } from '../common/common';
import { getAppDataPath } from 'src/utils/file';
import { useIntegrationStore } from 'src/stores/integration';
import { DriveType } from 'src/utils/interface/files';
import { CommonFetch } from '../../fetch';

export async function fetchRepo(): Promise<any[]> {
	const res: any = await CommonFetch.post('/drive/accounts', {});

	const integrationStore = useIntegrationStore();
	const supports = integrationStore.clientFilesCloudSupportList();
	const repos: any[] = res.data.data.filter((el) => {
		return el.available && el.type != 'space' && supports.includes(el.type);
	});
	return repos;
}

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
	const filesStore = useFilesStore();
	const activeMenu = filesStore.activeMenu();
	const promises: any[] = [];

	for (const item of items) {
		const from = item.from;

		let to = item.to.endsWith('/') ? item.to : item.to + '/';

		if (to.split('/')[4] === '') {
			const emailPosition = to.indexOf(activeMenu.label);
			if (emailPosition !== -1) {
				to =
					to.slice(0, emailPosition + activeMenu.label.length) +
					'/root' +
					to.slice(emailPosition + activeMenu.label.length);
			}
		}

		const url = `${from}?action=${
			copy ? 'copy' : 'rename'
		}&destination=${to}&override=${overwrite}&rename=${rename}&src_type=${
			item.src_drive_type
		}&dst_type=${item.dst_drive_type}`;

		promises.push(pasteAction(url));
	}

	return Promise.all(promises);
}

export async function rename(from, to) {
	const enc_to = to;
	const url = `${from}?action=rename&destination=${enc_to}&override=${false}&rename=${false}&src=${
		DriveType.GoogleDrive
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
	const lastIndex = files.path.lastIndexOf('/');
	const secondLastIndex = files.path.lastIndexOf('/', lastIndex - 1);
	const id = files.path.substring(secondLastIndex + 1, lastIndex);
	const name = files.path.split('/')[3];
	return generateDownloadUrl(files.driveType, id, name);
}

export function generateDownloadUrl(
	driveType: DriveType,
	id: string,
	name: string
) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	return `${baseURL}/drive/download_sync_stream?drive=${driveType}&cloud_file_path=${id}&name=${name}`;
}

export function saveGoogleDirInfo(items, origin_id) {
	const filesStore = useFilesStore();
	for (let i = 0; i < items.length; i++) {
		const item = items[i];
		if (item.isDir) {
			filesStore.googleDirMap[origin_id][item.id] = item.name;
		}
	}
}

export function extensionByMimeType(type: string) {
	const mimeTypes = {
		'application/msword': 'doc',
		'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
			'docx',
		'application/vnd.oasis.opendocument.text': 'odt',
		'application/vnd.apple.pages': 'pages',
		'application/pdf': 'pdf',
		'application/vnd.openxmlformats-officedocument.presentationml.presentation':
			'pptx',
		'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'xlsx',
		'application/rtf': 'rtf',
		'text/xml': 'xml',
		'text/html': 'html',
		'image/jpeg': 'jpeg',
		'image/png': 'png',
		'image/gif': 'gif',
		'image/bmp': 'bmp',
		'image/svg+xml': 'svg',
		'image/webp': 'webp',
		'image/tiff': 'tiff',
		'text/plain': 'txt',
		'text/css': 'css',
		'application/javascript': 'js',
		'application/json': 'json',
		'application/zip': 'zip',
		'application/x-rar-compressed': 'rar',
		'application/x-7z-compressed': '7z',
		'application/x-tar': 'tar',
		'application/gzip': 'gz',
		'application/x-bzip2': 'bz2',
		'audio/mpeg': 'mp3',
		'audio/wav': 'wav',
		'audio/aac': 'aac',
		'video/mp4': 'mp4',
		'video/x-msvideo': 'avi',
		'video/quicktime': 'mov',
		'video/webm': 'webm',
		'application/vnd.google-apps.folder': 'folder'
	};

	return mimeTypes[type] || 'unknown';
}

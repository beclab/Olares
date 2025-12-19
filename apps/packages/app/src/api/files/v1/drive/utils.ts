import { createURL, removePrefix } from '../utils';
import { useDataStore } from 'src/stores/data';
import { checkAppData, getAppDataPath } from 'src/utils/file';
// import { seahubGetRepos } from './syncMenu';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { formatUrltoDriveType } from '../common/common';
import { CommonFetch } from '../../fetch';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import DriveDataAPI from './data';

export function formatUrl(url: string) {
	let newUrl = url;
	if (url.startsWith('/Files')) {
		newUrl = removePrefix(url);
	}
	return `/api/resources${newUrl}`;
}

export async function createDir(url) {
	await CommonFetch.post(formatUrl(url));
}

export async function pasteAction(fromUrl): Promise<any> {
	const opts: any = {};

	console.log('pasteAction fromUrl', fromUrl);

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

	if (res.data.code === -1) {
		return BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: i18n.global.t('files.backslash_upload')
		});
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
	await CommonFetch.delete(formatUrl(url));
}

export async function put(url, content = '') {
	CommonFetch.put(formatUrl(url), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
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

export async function post(url, content = '', overwrite = false, onupload) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	url = removePrefix(url);
	let bufferContent;
	if (
		new Blob([content], { type: 'text/plain' }) instanceof Blob &&
		!['http:', 'https:'].includes(window.location.protocol)
	) {
		bufferContent = await new Response(content).arrayBuffer();
	}

	return new Promise((resolve, reject) => {
		const request = new XMLHttpRequest();

		if (checkAppData(url)) {
			const { path, node } = getAppDataPath(url);
			if (node) {
				request.open(
					'POST',
					`${baseURL}/api/resources/AppData${path}?override=${overwrite}`,
					true
				);
				request.setRequestHeader('X-Terminus-Node', node);
			}
		} else {
			request.open(
				'POST',
				`${baseURL}/api/resources${url}?override=${overwrite}`,
				true
			);
		}

		if (typeof onupload === 'function') {
			request.upload.onprogress = onupload;
		}

		request.onload = () => {
			if (request.status === 200) {
				resolve(request.responseText);
			} else if (request.status === 409) {
				reject(request.status);
			} else {
				reject(request.responseText);
			}
		};

		request.onerror = () => {
			reject(new Error('001 Connection aborted'));
		};

		request.send(bufferContent || content);
	});
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

export function getSubtitlesURL(file) {
	const params = {
		inline: 'true'
	};

	const subtitles: string[] = [];
	for (const sub of file.subtitles) {
		subtitles.push(createURL('api/raw' + sub, params));
	}

	return subtitles;
}

export async function getContentUrlByPath(filePath) {
	if (!filePath) {
		return;
	}
	return '';
	// const store = useDataStore();
	// return await store.api.getFileDownloadLink(store.repo.repo_id, filePath);
}

// export async function errorRetry(
// 	url,
// 	content: File,
// 	overwrite = false,
// 	onupload,
// 	timer
// ) {
// 	timer = timer - 1;

// 	const dataAPI = dataAPIs(formatUrltoDriveType(url));

// 	await dataAPI.fetchUploader(url, content, overwrite, onupload, timer);
// }

export async function uploadChunks(
	fileInfo,
	chunkFile,
	i,
	exportProgress,
	node
) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	const formData = new FormData();

	const offset = fileInfo.offset + DriveDataAPI.SIZE * i;
	formData.append('file', chunkFile.file);
	formData.append('upload_offset', offset);

	const headers = {};
	if (node) {
		headers['X-Terminus-Node'] = node;
	}

	const response = await CommonFetch.patch(
		`${baseURL}/upload/${fileInfo.id}`,
		formData,
		{
			...headers,
			onUploadProgress: (progressEvent) => {
				const event = {
					loaded: progressEvent.loaded,
					total: progressEvent.total,
					lengthComputable: progressEvent.lengthComputable
				};
				if (progressEvent.lengthComputable) {
					event.loaded += offset;
					event.total = fileInfo.file_size;
					exportProgress(event);
				}
			}
		}
	);

	if (!response.data) {
		throw 'error0';
	}
}

export async function createFileChunk(fileInfo: { offset: any }, file: any) {
	const size = DriveDataAPI.SIZE;
	const fileChunkList: { file: string }[] = [];
	let cur = fileInfo.offset;
	while (cur < file.size) {
		fileChunkList.push({
			file: file.slice(cur, cur + size >= file.size ? file.size : cur + size)
		});
		cur += size;
	}

	return fileChunkList;
}

export async function getUploadInfo(url: string, prefix: string, content: any) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	const appendUrl = splitUrl(url, content);

	const params = new FormData();
	params.append('storage_path', `${prefix}${appendUrl.storage_path}`);
	params.append('file_relative_path', appendUrl.file_relative_path);
	params.append('file_type', content.type);
	params.append('file_size', String(content.size));

	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 10000,
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
	});

	const res = await instance.post(baseURL + '/upload', params);
	return res.data.data;
}

const splitUrl = (url, content) => {
	let storage_path = '';
	let file_relative_path = '';
	let slicePathIndex = 0;
	if (content.fullPath) {
		slicePathIndex = url.indexOf(content.fullPath);
		file_relative_path = content.fullPath;
	} else {
		slicePathIndex = url.indexOf(content.name);
		file_relative_path = content.name;
	}

	storage_path = url.slice(0, slicePathIndex);
	return {
		storage_path,
		file_relative_path
	};
};

// import axios from 'axios';
import { logout } from 'src/utils/auth';
import { encodePath } from 'src/utils/url';
import { useDataStore } from 'src/stores/data';
import { useFilesStore } from 'src/stores/files';
import { busEmit, NetworkErrorMode } from 'src/utils/bus';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { InOfflineText } from 'src/utils/checkTerminusState';
import { CopyStoragesType } from 'src/stores/operation';
import { dataAPIs } from '.';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';

export async function fetchURL(
	url: string,
	opts?: any,
	auth = true,
	node = ''
) {
	const store = useDataStore();
	const baseURL: string = store.baseURL();
	opts = opts || {};
	opts.headers = opts.headers || {};

	const { headers: originalHeaders, ...rest } = opts;

	let headers = {
		...originalHeaders
	};

	try {
		if (node) {
			headers = {
				'X-Terminus-Node': node,
				...originalHeaders
			};
		}
	} catch (e) {
		console.error(e);
	}
	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: opts.timeout || 600000
	});
	headers = {
		'Access-Control-Allow-Origin': '*',
		'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
		'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS,PATCH',
		'Content-Type': 'application/json',
		'X-Unauth-Error': 'Non-Redirect',
		...headers
	};

	instance.interceptors.request.use(
		(config) => {
			return config;
		},
		(error) => {
			return Promise.reject(error);
		}
	);

	instance.interceptors.response.use(
		(response) => {
			return response;
		},
		async (error) => {
			if (error.message == InOfflineText()) {
				throw error;
			}
			busEmit('network_error', {
				type: NetworkErrorMode.file,
				error: error.message
			});
		}
	);

	let res: any;
	try {
		res = await instance({
			url: url,
			method: opts.method || 'get',
			baseURL: baseURL,
			headers: {
				...headers
			},
			...rest
		});
	} catch (e) {
		if (e.message == InOfflineText()) {
			e.status;
			throw e;
		}
		const error = new Error('000 No connection');
		throw error;
	}
	if (res.redirect) {
		const selfUrl =
			'/api/' + res.redirect.slice(res.redirect.indexOf('resources'));
		return fetchURL(selfUrl, {});
	}
	if (res.status === 459) {
		return window.history.go(-1);
	}
	if (res.status < 200 || res.status > 299) {
		const error: any = new Error(await res.text());
		error.status = res.status;
		if (auth && res.status == 401) {
			logout();
		}
		throw error;
	}
	return res;
}

export async function fetchJSON(
	url: string,
	opts?: { method?: string; body?: string } | undefined
) {
	const res = await fetchURL(url, opts);

	if (res.status === 200) {
		return res.json();
	} else {
		throw new Error(res.status);
	}
}

export function removePrefix(url) {
	url = url.split('/').splice(2).join('/');
	if (url === '') url = '/';
	if (url[0] !== '/') url = '/' + url;
	return url;
}

export function createURL(endpoint, params = {}, auth = true) {
	const store = useDataStore();
	const baseURL: string = store.baseURL();

	let prefix = baseURL;
	if (!prefix.endsWith('/')) {
		prefix = prefix + '/';
	}
	const url = new URL(prefix + encodePath(endpoint), origin);

	const searchParams = {
		...(auth && { auth: store.jwt }),
		...params
	};

	for (const key in searchParams) {
		url.searchParams.set(key, searchParams[key]);
	}

	return url.toString();
}

export function filterHiddenDir(res, id) {
	const filesStore = useFilesStore();
	if (filesStore.filterHiddenDir[id]) {
		return res.items.filter((item) => !item.name.startsWith('.'));
	}
}

export function getNotifyMsg(items: CopyStoragesType[]) {
	let notifyMsg = '';
	if (items.length <= 0) {
		return;
	} else if (items.length === 1) {
		let name = items[0].name;
		const maxLength = 60;
		if (name.length > maxLength) {
			name = name.substring(0, maxLength) + '...';
		}
		notifyMsg = `Pasting ${name}`;
	} else {
		notifyMsg = `Pasting ${items.length} items`;
	}

	return notifyMsg;
}

export async function getPreviewDownloadInfo(origin_id: number) {
	const filesStore = useFilesStore();
	const previewItem = filesStore.previewItem[origin_id];

	const url = dataAPIs(previewItem.driveType).getDownloadURL(
		previewItem,
		false
	);

	const result: TransferItem[] = [];
	const fileObj: TransferItem = {
		url,
		path: previewItem.path,
		parentPath: previewItem.parentPath,
		size: previewItem.isFolder ? 0 : previewItem.size,
		name: previewItem.name,
		type: previewItem.type,
		driveType: previewItem.driveType,
		front: TransferFront.download,
		status: TransferStatus.Prepare,
		uniqueIdentifier: previewItem.uniqueIdentifier,
		isPaused: false,
		isFolder: previewItem.isDir ? true : false,
		currentPhase: 1,
		totalPhase: 1
	};

	result.push(fileObj);

	return result;
}

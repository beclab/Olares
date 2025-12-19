// import axios from 'axios';
import { logout } from 'src/utils/auth';
import { encodePath } from 'src/utils/url';
import { useDataStore } from 'src/stores/data';
import { useFilesStore } from 'src/stores/files';
import { busEmit, NetworkErrorMode } from 'src/utils/bus';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { InOfflineText } from 'src/utils/checkTerminusState';
import { CopyStoragesType } from 'src/stores/operation';
import {
	TransferItem,
	TransferFront,
	TransferStatus
} from 'src/utils/interface/transfer';
import { dataAPIs } from './index';

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
	if (!prefix.endsWith('/') && !endpoint.startsWith('/')) {
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

export function getPurePath(path) {
	const new_dest_path = path.endsWith('/') ? path.slice(0, -1) : path;
	const pure_path = new_dest_path.slice(0, new_dest_path.lastIndexOf('/') + 1);
	return pure_path;
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

import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { CommonFetch } from '../../fetch';
import { CopyStoragesType } from 'src/stores/operation';
import { OPERATE_ACTION } from 'src/utils/contact';
import { uuid } from '@didvault/sdk/src/core';
import { getNotifyMsg } from '../utils';
import { notifyHide, notifyWaitingShow } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';
import { TransferFront } from 'src/utils/interface/transfer';
import {
	FileItem,
	FileNode,
	useFilesStore,
	ShareUserList
} from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { appendPath } from '../path';
import { encodeUrl } from 'src/utils/encode';
import { useDataStore } from 'src/stores/data';
import { files } from 'jszip';

export type CommonUrlApiType =
	| 'resources'
	| 'paste'
	| 'raw'
	| 'md5'
	| 'permission'
	| 'preview'
	| 'repos'
	| 'nodes'
	| 'stream'
	| 'tree'
	| 'share'
	| 'users';

export type FileType = 'drive' | 'external' | 'sync' | 'cache';

type TaskActionType = {
	taskId: string;
	type: TransferFront;
};

export const commonUrlPrefix = (apiType: CommonUrlApiType) => {
	return `/api/${apiType}/`;
};

export const commonUrlTypeExtend = (
	apiType: CommonUrlApiType,
	fileType: string,
	fileExtend: string
) => {
	return appendPath('/api', apiType, fileType, fileExtend, '/');
};

export const pasteMutiNodesDriveType: DriveType[] = [
	DriveType.External,
	DriveType.Cache
];

export async function pasteAction(
	item: CopyStoragesType,
	action: 'copy' | 'move'
): Promise<any> {
	const opts: any = {};

	const filesStore = useFilesStore();

	const node =
		item.dst_node ||
		item.src_node ||
		(filesStore.nodes.length > 0 ? filesStore.nodes[0].name : '');

	let destination = item.to;
	let source = item.from;
	try {
		destination = decodeURIComponent(destination);
		source = decodeURIComponent(source);
	} catch (error) {
		/* empty */
	}

	const res = await CommonFetch.patch(
		commonUrlPrefix('paste') + node + '/',
		{
			action,
			destination,
			source: source
		},
		opts
	);

	if (res.data.code != undefined) {
		if (res.data.code === -1) {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: i18n.global.t('files.backslash_upload')
			});
		}
		return undefined;
	}

	return {
		taskId: res.data.task_id,
		type: action == 'copy' ? TransferFront.copy : TransferFront.move,
		node,
		src_drive_type: item.src_drive_type,
		dst_drive_type: item.dst_drive_type
	};
}

export function moveCopy(items: CopyStoragesType[], copy = false) {
	const promises: any[] = [];

	for (const item of items) {
		promises.push(pasteAction(item, copy ? 'copy' : 'move'));
	}

	return Promise.all(promises);
}

export function move(items: CopyStoragesType[]) {
	return moveCopy(items, false);
}

export function copy(items) {
	return moveCopy(items, true);
}

export const action = async (
	// overwrite: boolean | undefined,
	// rename: boolean | undefined,
	items: CopyStoragesType[],
	path: string,
	isMove: boolean | undefined,
	callback: (action: OPERATE_ACTION, data: any) => Promise<void>
): Promise<TaskActionType[]> => {
	const dest = path;
	const notifyId = await uuid();
	const notifyMsg = getNotifyMsg(items);
	let tasks: TaskActionType[] = [];
	notifyWaitingShow(notifyMsg, notifyId);

	if (isMove) {
		await move(items)
			.then((res) => {
				tasks = res.filter((e) => e != undefined);
				callback(OPERATE_ACTION.MOVE, dest);
				notifyHide(notifyId);
			})
			.catch(() => {
				notifyHide(notifyId);
			});
	} else {
		await copy(items)
			.then((res) => {
				tasks = res.filter((e) => e != undefined);
				callback(OPERATE_ACTION.PASTE, dest);
				notifyHide(notifyId);
			})
			.catch(() => {
				notifyHide(notifyId);
			});
	}

	return tasks;
};

export async function rename(from: string, to: string) {
	const url = `${from}?action=rename&destination=${to}&override=${false}&rename=${false}`;
	const res = await CommonFetch.patch(url);
	return res;
}

export async function rename2(path: string, destination: string) {
	const res = await CommonFetch.patch(`${path}?destination=${destination}`);
	return res;
}

export async function renameFileItem(item: FileItem, newName: string) {
	console.log('item ===>', item);

	const url = appendPath(
		commonUrlPrefix('resources'),
		item.fileType || '',
		item.fileExtend,
		encodeUrl(item.oPath || ''),
		item.isDir ? '/' : ''
	);

	const params = {
		destination: encodeURIComponent(newName),
		driveId: item.driveType == DriveType.GoogleDrive ? item.id : undefined
	};

	return CommonFetch.patch(url, undefined, {
		params
	});
}

export async function batchDelete(path: string, dirents: string[]) {
	return await CommonFetch.delete(
		appendPath(commonUrlPrefix('resources'), path, '/'),
		{
			data: {
				dirents
			}
		}
	);
}

export async function postCreateFile(path: string, isDir: boolean, body: any) {
	return await CommonFetch.post(
		appendPath(commonUrlPrefix('resources'), path, isDir ? '/' : ''),
		body
	);
}

export async function batchDeleteFileItems(items: FileItem[]) {
	if (items.length == 0) {
		return;
	}

	const groups: FileItem[][] = [];
	for (let i = 0; i < items.length; i++) {
		const item = items[i];
		const index = groups.findIndex(
			(e) => e.length > 0 && e[0].oParentPath == item.oParentPath
		);
		if (index >= 0) {
			groups[index].push(item);
		} else {
			groups.push([item]);
		}
	}

	for (let index = 0; index < groups.length; index++) {
		const items = groups[index];
		const path = appendPath(
			'/',
			items[0].fileType || '',
			items[0].fileExtend,
			items[0].oParentPath || '/'
		);

		const dirents = items.map((e) => {
			return appendPath('/', e.name, e.isDir ? '/' : '');
		});

		try {
			await batchDelete(path, dirents);
		} catch (error) {
			/* empty */
		}
	}
}

export function formatAppDataNode(
	url: string,
	data: FileNode[],
	driveType: DriveType,
	parentPath: string
) {
	const nodeDir = {
		path: url,
		name: '',
		size: 0,
		extension: '',
		modified: 0,
		mode: 0,
		isDir: true,
		isSymlink: false,
		type: '',
		numDirs: 0,
		numFiles: 0,
		sorting: {
			by: 'modified',
			asc: true
		},
		fileSize: 0,
		numTotalFiles: 0,
		items: <FileItem[]>[],
		driveType,
		fileExtend: '',
		filePath: '',
		fileType: ''
	};

	if (data.length > 0) {
		nodeDir.numDirs = data.length;
		data.forEach((el, index) => {
			const path = appendPath(parentPath, el.name, '/');
			const item: FileItem = {
				path: path,
				name: el.name,
				size: 4096,
				extension: '',
				modified: 0,
				mode: 0,
				isDir: true,
				isSymlink: false,
				type: '',
				sorting: {
					by: 'size',
					asc: false
				},
				driveType,
				param: '',
				url: '',
				index: index,
				fileExtend: el.name,
				isNode: true
			};
			nodeDir.items.push(item);
		});
	}

	return nodeDir;
}

export async function fetchNodeList(): Promise<FileNode[]> {
	try {
		const res: any = await CommonFetch.get(commonUrlPrefix('nodes'), {});
		const filesStore = useFilesStore();
		filesStore.nodes = res.data.nodes;
		return res.data.nodes;
	} catch (error) {
		return [];
	}
}

export async function fetchUserList(): Promise<ShareUserList | undefined> {
	try {
		const res: any = await CommonFetch.get(commonUrlPrefix('users'), {});
		const filesStore = useFilesStore();
		filesStore.users = res.data;
		if (!filesStore.shareFilter.ownerInit) {
			filesStore.shareFilter.ownerInit = true;
			filesStore.shareFilter.owner =
				filesStore.users?.users.map((e) => e.name) || [];
		}
		return res.data;
	} catch (error) {
		return;
	}
}

export function getStreamListUrl(item: FileItem) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	const url = appendPath(
		baseURL,
		commonUrlPrefix('tree'),
		item.fileType || '',
		item.fileExtend,
		encodeUrl(item.oPath || ''),
		item.isDir ? '/' : ''
	);
	return url;
}

export function getDownloadUrl(item: FileItem, params = {}) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	if (item.isDir && process.env.APPLICATION === 'LAREPASS') {
		return getStreamListUrl(item);
	}

	const path = appendPath(
		baseURL,
		commonUrlPrefix('raw'),
		item.fileType || '',
		item.fileExtend,
		encodeUrl(item.oPath || ''),
		item.isDir ? `?algo=zip` : ''
	);

	const url = new URL(path, origin);

	const searchParams = {
		...params
	};

	for (const key in searchParams) {
		url.searchParams.set(key, searchParams[key]);
	}

	return url.toString();
}

import { DriveDataAPI } from './../index';
import { getAppDataPath } from 'src/utils/file';
import { formatAppDataNode, formatAppData } from './filesFormat';
import { MenuItem } from 'src/utils/contact';
import { OPERATE_ACTION } from 'src/utils/contact';
// import { files } from './../index';
import { useFilesStore } from 'src/stores/files';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { FileItem, FileResType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import { useDataStore } from 'src/stores/data';
import { notifyWaitingShow, notifyHide } from 'src/utils/notifyRedefinedUtil';
import {
	rename,
	copy,
	move,
	downloadZip,
	remove,
	createDir,
	put
} from './utils';
import url from 'src/utils/url';
import {
	TransferItem,
	TransferStatus,
	TransferFront
} from 'src/utils/interface/transfer';

import { createURL, getNotifyMsg } from '../utils';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import { getFileIcon } from '@bytetrade/core';
import { filterPcvPath } from '../common/common';
import { getextension } from 'src/utils/utils';

export default class CacheDataAPI extends DriveDataAPI {
	breadcrumbsBase = '';

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchCache(url);

		return res;
	}

	async fetchCache(url: string): Promise<FileResType> {
		const { path, node } = getAppDataPath(url);
		let headers = {
			auth: false,
			'X-Terminus-Node': ''
		};
		if (node) {
			headers = {
				auth: true,
				'X-Terminus-Node': node
			};
		}
		const options = headers.auth ? { headers: headers } : {};

		const res: any = await this.commonAxios.get(
			`/api/resources/AppData${path}`,
			options
		);

		let data: FileResType;
		if (res.data) {
			data = formatAppDataNode(url, JSON.parse(JSON.stringify(res)));
		} else {
			data = formatAppData(
				node,
				JSON.parse(JSON.stringify(res)),
				url,
				this.origin_id
			);
		}

		return data;
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.CACHE}`),
				key: MenuItem.CACHE,
				icon: 'sym_r_analytics',
				driveType: DriveType.Cache
			}
		];
	}

	async formatRepotoPath(item: any): Promise<string> {
		if (item.key === 'Cache') return '/Cache/';
		return '/Cache/' + item.key;
	}

	formatPathtoUrl(path: string): string {
		if (path.endsWith('/Cache/')) {
			return '/AppData';
		}

		const purePath = path.replace('/Cache', '/AppData');

		return purePath;
	}

	async openPreview(item: any): Promise<FileResType> {
		const cur_item = JSON.parse(JSON.stringify(item));
		cur_item.path = cur_item.path.replace('/AppData', '');
		return cur_item;
	}

	getPreviewURL(file: any, thumb: string): string {
		const { path } = getAppDataPath(file.path);
		if (['video'].includes(file.type)) {
			return '/AppData' + path;
		}

		// const path = file.path.split('/').slice(3).join('/');

		const previewPath = `api/cache/${
			file.path.split('/')[2]
		}/preview/${thumb}/AppData/${decodeURIComponent(path)}`;

		return createURL(previewPath, {}, false);
	}

	getDownloadURL(file: any, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const path = file.path.split('/').slice(3).join('/');
		const url = `api/cache/${file.path.split('/')[2]}/raw/AppData/${path}`;

		return createURL(url, params, false);
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		if (
			!['audio', 'video', 'text', 'txt', 'textImmutable', 'pdf'].includes(
				file.type
			)
		) {
			return file;
		}

		if (['text', 'txt', 'textImmutable'].includes(file.type)) {
			try {
				// const url = decodeURIComponent(file.path);
				const path = file.path.split('/').slice(3).join('/');
				const previewPath = `api/cache/${
					file.path.split('/')[2]
				}/resources/AppData/${path}`;

				const res = await this.commonAxios.get(previewPath, {});

				file.content = res.content;
			} catch (error) {
				console.error(error.message);
			}
		}
		return file;
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		const url = item.url.replace('/Files', '');
		put(url, content);
	}

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const copyItem: CopyStoragesType = {
			from: el.path,
			to: '',
			name: el.name,
			src_drive_type: DriveType.Cache
		};

		if (type === 'cut') {
			copyItem.key = 'x';
		}
		return copyItem;
	}

	async paste(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void> {
		const operateinStore = useOperateinStore();
		const items: CopyStoragesType[] = [];

		const modifiedPath = path.endsWith('/') ? path : path + '/';

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element: any = operateinStore.copyFiles[i];

			const to = modifiedPath + encodeUrl(element.name);

			items.push({
				from: element.from,
				to: to,
				name: element.name,
				src_drive_type: element.src_drive_type,
				dst_drive_type: DriveType.Cache
			});
			if (path + element.name === element.from) {
				await this.action(false, true, items, path, false, callback);
				return;
			}
		}

		let overwrite = false;
		const rename = true;
		let isMove = false;

		if (
			operateinStore.copyFiles[0] &&
			operateinStore.copyFiles[0].key === 'x'
		) {
			overwrite = true;
			isMove = true;
		}

		await this.action(overwrite, rename, items, path, isMove, callback);
	}

	async move(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void> {
		const filesStore = useFilesStore();
		const items: CopyStoragesType[] = [];
		for (const i of filesStore.selected[this.origin_id]) {
			const element = filesStore.getTargetFileItem(i, this.origin_id);
			if (!element) {
				continue;
			}
			const from = element.path;
			const to = path + element.name;
			items.push({
				from: from,
				to: to,
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: DriveType.Cache
			});
		}
		const overwrite = true;
		await this.action(overwrite, true, items, path, true, callback);
	}

	async action(
		overwrite: boolean | undefined,
		rename: boolean | undefined,
		items: CopyStoragesType[],
		path: string,
		isMove: boolean | undefined,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void> {
		const dest = path;

		const notifyMsg = getNotifyMsg(items);
		const timestamp = Date.now();
		const notify_id = `${items[0].name}_${timestamp}`;

		notifyWaitingShow(notifyMsg, notify_id);

		if (isMove) {
			try {
				await move(items, overwrite, rename)
					.then(() => {
						callback(OPERATE_ACTION.MOVE, dest);
						notifyHide(notify_id);
					})
					.catch(() => {
						notifyHide(notify_id);
					});
			} catch (error) {
				notifyHide(notify_id);
			}
		} else {
			try {
				await copy(items, overwrite, rename)
					.then(() => {
						callback(OPERATE_ACTION.PASTE, dest);
						notifyHide(notify_id);
					})
					.catch(() => {
						notifyHide(notify_id);
					});
			} catch (error) {
				notifyHide(notify_id);
			}
		}
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const oldLink = item.path;
		const newLink = url.removeLastDir(oldLink) + '/' + encodeUrl(newName);

		await rename(oldLink, newLink);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		let url = path + '/' + encodeURIComponent(dirName) + '/';
		url = url.replace('//', '/');

		await createDir(url);
	}

	async getDownloadInfo(path: string): Promise<TransferItem[]> {
		const filesStore = useFilesStore();
		const selected = filesStore.selected[this.origin_id];

		const store = useDataStore();
		const baseURL = store.baseURL();

		const nodes = path.split('/')[2];
		const p = path.slice(path.indexOf(nodes) + nodes.length);

		const result: TransferItem[] = [];

		for (let i = 0; i < selected.length; i++) {
			const index = selected[i];
			const selectFile = filesStore.getTargetFileItem(index, this.origin_id);

			if (!selectFile) {
				continue;
			}

			let url = '';
			if (selectFile.isDir) {
				if (process.env.APPLICATION === 'FILES') {
					url = downloadZip('zip', selectFile.path);
				} else {
					const store = useDataStore();
					const baseURL = store.baseURL();
					url =
						baseURL +
						'/api/cache/' +
						nodes +
						'/resources/AppData' +
						p +
						encodeUrl(selectFile.name) +
						'?stream=1';
				}
			} else {
				url = `${baseURL}/api/cache/${nodes}/raw/AppData${p}/${encodeUrl(
					selectFile.name
				)}`;
			}

			const fileObj: TransferItem = {
				url,
				path: selectFile.path,
				parentPath: path,
				size: selectFile.size,
				name: selectFile.name,
				type: selectFile.type,
				driveType: selectFile.driveType,
				front: TransferFront.download,
				status: TransferStatus.Prepare,
				isPaused: false,
				isFolder: selectFile.isDir ? true : false,
				currentPhase: 1,
				totalPhase: 1
			};

			result.push(fileObj);
		}

		return result;
	}

	async formatUploaderPath(path: string): Promise<string> {
		if (!path.endsWith('/')) {
			path = path + '/';
		}
		const splitPath = path.split('/');
		splitPath.splice(2, 1);
		const newpath = splitPath.join('/');
		return newpath.replace('/Cache', '/AppData');
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		const promises: any = [];

		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			promises.push(remove(item.path));
		}

		await Promise.all(promises);
	}

	getDiskPath(selectFiles: any, type: string) {
		const syncPath = selectFiles.path.split('/').slice(3).join('/');
		const path = `/api/cache/${
			selectFiles.path.split('/')[2]
		}/${type}/AppData/${syncPath}`;
		return path;
	}

	formatTransferToFileItem(item: TransferItem): FileItem {
		const extension = getextension(item.name);
		if (item.path.endsWith(item.name)) {
			item.path =
				item.path.substring(0, item.path.length - item.name.length) +
				encodeUrl(item.name);
			if (item.isFolder && !item.path.endsWith('/')) {
				item.path = item.path + '/';
			}
		}

		const res: FileItem = {
			extension,
			isDir: item.isFolder,
			isSymlink: false,
			mode: 0,
			modified: item.updateTime || 0,
			name: item.name,
			path:
				item.front == TransferFront.download ? encodeUrl(item.path) : item.path,
			size: item.size,
			type: item.type,
			parentPath: item.parentPath,
			index: 0,
			url: item.url || '',
			driveType: item.driveType!,
			param: '',
			fileExtend: ''
		};

		return res;
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const path = file.path.split('/').slice(3).join('/');

		const url = `api/cache/${file.path.split('/')[2]}/raw/AppData/${path}`;

		return createURL(url, params, false);
	}

	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;

		file.driveType = DriveType.Cache;

		const { node } = getAppDataPath(parentPath || '');
		const splitPath = filterPcvPath(file.path).split('/');
		splitPath.splice(splitPath.indexOf('AppData') + 1, 0, node);
		const joinPath = splitPath.join('/').replace('AppData', 'Cache');

		file.path = joinPath;
		file.driveType = DriveType.Cache;
		file.parentPath = '/' + parentPath?.split('/').splice(1).join('/');
		file.url = this.getFormatSteamDownloadFileURL(file);
	}
}

import { DriveDataAPI } from './../index';
import { MenuItem } from 'src/utils/contact';
import { formatData } from './filesFormat';
import { FileResType, FileItem } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import url from 'src/utils/url';
import { createURL, getNotifyMsg } from '../utils';
import { OPERATE_ACTION } from 'src/utils/contact';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { rename, copy, move, remove, createDir } from './utils';
import { useFilesStore } from 'src/stores/files';
import { notifyWaitingShow, notifyHide } from 'src/utils/notifyRedefinedUtil';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import {
	TransferItem,
	TransferStatus,
	TransferFront
} from 'src/utils/interface/transfer';
import * as files from './utils';
import { getFileIcon } from '@bytetrade/core';
import { filterPcvPath } from '../common/common';
import md5 from 'js-md5';
import { useDataStore } from 'src/stores/data';

export default class DataDataAPI extends DriveDataAPI {
	breadcrumbsBase = '';

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = this.formatPathtoUrl(decodeURIComponent(url));

		const res: FileResType = await this.fetchData(pureUrl);

		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(
			`/api/resources${encodeUrl(url)}`,
			{}
		);

		const data: FileResType = await formatData(
			JSON.parse(JSON.stringify(res)),
			url,
			this.origin_id
		);

		return data;
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.DATA}`),
				key: MenuItem.DATA,
				icon: 'sym_r_database',
				driveType: DriveType.Data
			}
		];
	}

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const from = el.url.slice(6);
		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: DriveType.Drive
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

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element: any = operateinStore.copyFiles[i];
			let lastPathIndex = this.formatPathtoUrl(path);

			lastPathIndex = lastPathIndex.endsWith('/')
				? lastPathIndex
				: lastPathIndex + '/';

			const to = lastPathIndex + encodeUrl(element.name);
			items.push({
				from: element.from,
				to: to,
				name: element.name,
				src_drive_type: element.src_drive_type,
				dst_drive_type: DriveType.Drive
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

			const from = this.formatPathtoUrl(decodeURIComponent(element.path));
			const to = this.formatPathtoUrl(decodeURIComponent(path + element.name));

			items.push({
				from: from,
				to: to,
				name: element.name,
				src_drive_type: DriveType.Drive,
				dst_drive_type: DriveType.Drive
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

	async deleteItem(items: FileItem[]): Promise<void> {
		const promises: any = [];

		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			item.path = this.formatPathtoUrl(item.path);
			promises.push(remove(item.path));
		}

		await Promise.all(promises);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const oldLink = this.formatPathtoUrl(item.path);
		const newLink = this.formatPathtoUrl(
			url.removeLastDir(oldLink) + '/' + encodeUrl(newName)
		);

		await rename(oldLink, newLink);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		let url =
			this.formatPathtoUrl(path) + '/' + encodeURIComponent(dirName) + '/';
		url = url.replace('//', '/');

		await createDir(url);
	}

	async formatRepotoPath(item: any): Promise<string> {
		if (item.key === 'Data') return '/Data/';
		return '/Data/' + item.key;
	}

	formatPathtoUrl(path: string): string {
		let purePath = '';

		if (path.startsWith('/Data')) {
			purePath = '/Application' + decodeURIComponent(path).slice(5);
		} else {
			purePath = path;
		}

		return purePath;
	}

	async formatUploaderPath(path: string): Promise<string> {
		const purePath =
			'/data' +
			(this.formatPathtoUrl(path).endsWith('/')
				? this.formatPathtoUrl(path)
				: this.formatPathtoUrl(path) + '/');

		return purePath;
	}

	getDiskPath(selectFiles: any, type: string) {
		const path = this.formatPathtoUrl(selectFiles.path);
		return `/api/${type}${path}`;
	}

	getPreviewURL(file: any, thumb: string): string {
		if (['video'].includes(file.type)) {
			return file.path;
		}
		const params = {
			inline: 'true',
			key: file.modified
		};

		const curPath = this.formatPathtoUrl(file.path);

		const res = createURL(
			'api/preview/' + decodeURIComponent(thumb + curPath),
			params
		);

		return res;
	}

	getDownloadURL(file: any, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const curPath = this.formatPathtoUrl(file.path);

		const url = createURL('api/raw' + decodeURIComponent(curPath), params);

		return url;
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
				const curPath = this.formatPathtoUrl(file.path);

				const res = await this.commonAxios.get(`/api/resources${curPath}`, {});

				file.content = res.content;
			} catch (error) {
				console.error(error.message);
			}
		}
		return file;
	}

	async getDownloadInfo(path: string): Promise<TransferItem[]> {
		const filesStore = useFilesStore();
		const selected = filesStore.selected[this.origin_id];

		const result: TransferItem[] = [];
		const parentPath = '/' + path.split('/').splice(2).join('/');

		for (let i = 0; i < selected.length; i++) {
			const index = selected[i];
			const selectFile = filesStore.getTargetFileItem(index, this.origin_id);
			if (!selectFile) {
				continue;
			}

			const selectFilePath =
				'/' + selectFile.path.split('/').splice(2).join('/');

			let url = '';

			if (selectFile.isDir) {
				if (process.env.APPLICATION === 'FILES') {
					url = files.download('zip', selectFile.path);
				} else {
					const store = useDataStore();
					const baseURL = store.baseURL();
					url =
						baseURL +
						'/api/resources/Application' +
						selectFilePath +
						'?stream=1';
				}
			} else {
				url = files.download(null, selectFile.url);
			}

			const fileObj: TransferItem = {
				url,
				path: '/Files' + selectFilePath,
				parentPath,
				size: selectFile.isDir ? 0 : selectFile.size,
				name: selectFile.name,
				type: selectFile.type,
				driveType: selectFile.driveType,
				front: TransferFront.download,
				status: TransferStatus.Prepare,
				uniqueIdentifier: selectFile.uniqueIdentifier,
				isPaused: false,
				isFolder: selectFile.isDir ? true : false,
				currentPhase: 1,
				totalPhase: 1
			};

			result.push(fileObj);
		}

		return result;
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};
		const curPath = this.formatPathtoUrl(file.path);
		const url = createURL('api/raw' + curPath, params);
		return url;
	}

	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;
		file.path = filterPcvPath(file.path);
		file.driveType = DriveType.Data;

		file.parentPath = '/' + parentPath?.split('/').splice(1).join('/');
		file.uniqueIdentifier =
			md5(file.relativePath + new Date()) + file.relativePath;

		file.url = this.getFormatSteamDownloadFileURL(file);
	}
}

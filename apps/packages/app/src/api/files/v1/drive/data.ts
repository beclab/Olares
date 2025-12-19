/* eslint-disable @typescript-eslint/no-unused-vars */
import Origin from './../origin';
import { removePrefix, createURL, getNotifyMsg } from '../utils';
import { formatDrive } from './filesFormat';
import { MenuItem } from 'src/utils/contact';
import { OPERATE_ACTION } from 'src/utils/contact';
import * as files from './utils';
import { notifyWaitingShow, notifyHide } from 'src/utils/notifyRedefinedUtil';

import { checkSameName } from 'src/utils/file';
import url from 'src/utils/url';
import { CommonFetch } from '../../fetch';
import { useFilesStore, FilesIdType } from 'src/stores/files';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';

import { FileItem, FileResType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import { useDataStore } from 'src/stores/data';
import { useTransfer2Store } from 'src/stores/transfer2';
import {
	TransferItem,
	TransferStatus,
	TransferFront
} from 'src/utils/interface/transfer';
import { filterPcvPath, filterPcvPath2 } from '../common/common';
import { getextension } from 'src/utils/utils';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import { getFileIcon } from '@bytetrade/core';
import { IUploadCloudParams } from 'src/platform/interface/electron/interface';
import { useUserStore } from 'src/stores/user';
import { Router } from 'vue-router';
import md5 from 'js-md5';
import { SyncRepoItemType } from '../sync/type';
import { appendPath } from '../../path';

export default class DriveDataAPI extends Origin {
	public commonAxios: any;

	public origin_id: number;
	static SIZE: number;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
		this.commonAxios = CommonFetch;
	}

	breadcrumbsBase = '/Files';

	SIZE = 8 * 1024 * 1024;

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = removePrefix(url);

		const res: FileResType = await this.fetchDrive(pureUrl);

		return res;
	}

	async fetchDrive(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(`/api/resources${url}`, {});

		const data: FileResType = await formatDrive(
			JSON.parse(JSON.stringify(res)),
			url,
			this.origin_id
		);

		return data;
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.HOME}`),
				key: MenuItem.HOME,
				icon: 'sym_r_other_houses',
				driveType: DriveType.Drive
			},
			{
				label: i18n.global.t(`files_menu.${MenuItem.DOCUMENTS}`),
				key: MenuItem.DOCUMENTS,
				icon: 'sym_r_news',
				driveType: DriveType.Drive
			},
			{
				label: i18n.global.t(`files_menu.${MenuItem.PICTURES}`),
				key: MenuItem.PICTURES,
				icon: 'sym_r_art_track',
				driveType: DriveType.Drive
			},
			{
				label: i18n.global.t(`files_menu.${MenuItem.MOVIES}`),
				key: MenuItem.MOVIES,
				icon: 'sym_r_smart_display',
				driveType: DriveType.Drive
			},
			{
				label: i18n.global.t(`files_menu.${MenuItem.DOWNLOADS}`),
				key: MenuItem.DOWNLOADS,
				icon: 'sym_r_browser_updated',
				driveType: DriveType.Drive
			}
		];
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
					url = baseURL + '/api/resources' + selectFilePath + '?stream=1';
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

	async downloadFile(fileUrl: any, filename = ''): Promise<void> {
		const a = document.createElement('a');
		a.style.display = 'none';
		a.href = fileUrl.url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	}

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const from = el.path.slice(6);
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
		if (!operateinStore.copyFiles || operateinStore.copyFiles.length == 0) {
			return;
		}

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element: any = operateinStore.copyFiles[i];

			let lastPathIndex =
				path.indexOf('?') > -1
					? path.slice(6, path.indexOf('?'))
					: path.slice(6);

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
				this.action(false, true, items, path, false, callback);
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

			const from = element.path.slice(6);
			const to = (path + element.name).slice(6);
			items.push({
				from: from,
				to: to,
				name: element.name,
				src_drive_type: element.driveType,
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
				await files
					.move(items, overwrite, rename)
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
				await files
					.copy(items, overwrite, rename)
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

	uploadFiles(): void {
		let element: any = null;
		element = document.getElementById('uploader-input');
		element.value = '';
		element.removeAttribute('webkitdirectory');
		element.click();
	}

	uploadFolder(): void {
		let element: any = null;
		element = document.getElementById('uploader-input');
		element.value = '';
		element.setAttribute('webkitdirectory', 'webkitdirectory');
		element.click();
	}

	openLocalFolder(): string | undefined {
		return undefined;
	}

	async openPreview(item: any): Promise<FileResType> {
		const cur_item = JSON.parse(JSON.stringify(item));
		cur_item.path = cur_item.path.replace('/Files', '');
		return cur_item;
	}

	getPreviewURL(file: any, thumb: string): string {
		if (['video'].includes(file.type)) {
			return file.path;
		}
		const params = {
			inline: 'true',
			key: file.modified
		};

		return createURL(
			'api/preview/' + thumb + decodeURIComponent(file.path),
			params
		);
	}

	getDownloadURL(
		file: any,
		inline: boolean,
		_download?: boolean,
		decodeUrl = true
	): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		let file_path = file.path;

		if (decodeUrl) {
			try {
				file_path = decodeURIComponent(file.path);
			} catch (error) {
				console.log(file_path);
			}
		}

		const url = createURL('api/raw' + file_path, params);
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
				const url = file.path;
				console.log('url ===>', url);

				const res = await this.commonAxios.get(`/api/resources${url}`, {});

				file.content = res.content;
			} catch (error) {
				console.error(error.message);
			}
		}
		return file;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return (
			'/Files/Home/' +
			(item.key && item.key != MenuItem.HOME ? item.key + '/' : '')
		);
	}

	formatPathtoUrl(path: string): string {
		return path;
	}

	async formatUploaderPath(path: string): Promise<string> {
		if (!path.endsWith('/')) {
			path = path + '/';
		}
		return path.replace('/Files', '/data');
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		// const filesStore = useFilesStore();

		const promises: any = [];

		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			promises.push(files.remove(item.path));
		}

		await Promise.all(promises);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const oldLink = item.path;
		const newLink = url.removeLastDir(oldLink) + '/' + encodeUrl(newName);

		await files.rename(oldLink, newLink);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const filesStore = useFilesStore();
		const newName = await checkSameName(
			dirName,
			filesStore.currentFileList[this.origin_id]?.items
		);

		let url = path + '/' + encodeURIComponent(newName) + '/';
		url = url.replace('//', '/');

		await files.createDir(url);
	}

	getAttrPath(item: FileItem): string {
		return item.path.slice(0, item.path.indexOf(item.name));
	}

	async getFileServerUploadLink(
		folderPath: string,
		_repo_id?: string,
		_dirName?: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const path = folderPath;
		const url =
			baseURL + '/upload/upload-link/?p=' + encodeUrl(path) + '&from=web';
		const res = await this.commonAxios.get(url, {
			responseType: 'text'
		});
		return res + '?ret-json=1';
	}

	async getFileUploadedBytes(
		filePath: any,
		fileName: any,
		_repo_id: string,
		_task_id: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const url = baseURL + '/upload/file-uploaded-bytes/';
		const params = {
			parent_dir: filePath,
			file_name: fileName
		};
		const res = this.commonAxios.get(url, { params: params });
		return res;
	}

	async getCurrentRepoInfo(path: string): Promise<any> {
		const res = await this.fetch(path);
		return res;
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		files.put(item.url, content);
	}

	getResumePath(path: string, relativePath: string) {
		return path + relativePath;
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

	formatTransferPath(item: TransferItem) {
		const path = item.path;
		return decodeURIComponent(path.slice(0, path.lastIndexOf('/') + 1));
	}

	async formatUploadTransferPath(item: TransferItem) {
		const uploadPathSplit = item.path;
		const relatePath = item.relatePath
			? item.relatePath + item.name
			: item.name;
		let uploadPath = uploadPathSplit.split(relatePath)[0];
		if (!uploadPath.startsWith('/')) {
			uploadPath = '/' + uploadPath;
		}
		const pathname = (await this.formatUploaderPath(uploadPath)) || '/';
		return pathname;
	}

	getPurePath(path: string) {
		let curPath = filterPcvPath(path);
		curPath = curPath.endsWith('/') ? curPath : `${curPath}/`;
		return curPath;
	}

	getDiskPath(selectFiles: any, type: string) {
		let currentPath = selectFiles.path;
		if (currentPath.indexOf('/Files') > -1) {
			currentPath = currentPath.slice(6);
		}

		return `/api/${type}${currentPath}`;
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};
		const url = createURL('api/raw' + file.path, params);
		return url;
	}
	formatFolderSubItemDownloadPath(
		item: TransferItem,
		parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	) {
		let parentPath = item.parentPath || '';
		if (parentPath.startsWith('/Files')) {
			parentPath = parentPath.substring('/Files'.length);
		}
		parentPath = parentPath + parentItem.name + '/';
		const path = item.path;
		const releatePath = path.substring(
			parentPath.length,
			path.length - item.name.length
		);
		const parentSavePath =
			defaultDownloadPath + appendPath + folderSavePath + appendPath;

		const itemSavePath =
			parentSavePath +
			(releatePath && releatePath.length > 0 ? releatePath + appendPath : '');
		return {
			parentSavePath,
			itemSavePath
		};
	}
	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;
		file.path = filterPcvPath(file.path);
		file.driveType = DriveType.Drive;

		file.parentPath = '/' + parentPath?.split('/').splice(2).join('/');
		file.uniqueIdentifier =
			md5(file.relativePath + new Date()) + file.relativePath;

		file.url = this.getFormatSteamDownloadFileURL(file);
	}

	getUploadTransferItemMoreInfo(
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_item: TransferItem
	): IUploadCloudParams | undefined {
		return undefined;
	}

	fileEditEnable = true;

	videoPlayEnable = true;

	audioPlayEnable = true;

	fileEditLimitSize = 1024 * 1024;

	async transferItemUploadSuccessResponse(
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_tranfeItemId: number,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_response: any
	) {
		return true;
	}

	async uploadSuccessRefreshData(tranfeItemId: number) {
		try {
			const transferStore = useTransfer2Store();
			const userStore = useUserStore();
			const transferItem =
				transferStore.transferMap[tranfeItemId] ||
				transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

			if (!transferItem) {
				return;
			}

			if (transferItem.userId && transferItem.userId != userStore.current_id) {
				return;
			}

			const filesStore = useFilesStore();

			const fullPath = url.getWindowFullpath();

			const cur_file = this.formatTransferToFileItem(transferItem);
			let decodeCurlFilePath = cur_file.path;
			try {
				decodeCurlFilePath = decodeURIComponent(cur_file.path);
			} catch (error) {
				console.log('error', error);
			}

			let decoodeFullPath = fullPath;
			try {
				decoodeFullPath = decodeURIComponent(fullPath);
			} catch (error) {
				console.log('error', error);
			}

			if (decodeCurlFilePath.indexOf(decoodeFullPath) >= 0) {
				filesStore.setBrowserUrl(fullPath, cur_file.driveType, false);
			} else {
				let parentPath = decodeCurlFilePath;

				let decodeCurlFileName = cur_file.name;
				try {
					decodeCurlFileName = decodeURIComponent(decodeCurlFileName);
				} catch (error) {
					/* empty */
				}
				if (
					!parentPath.endsWith('/') &&
					parentPath.endsWith(decodeCurlFileName)
				) {
					parentPath = parentPath.substring(
						0,
						parentPath.length - decodeCurlFileName.length
					);
				}

				parentPath = encodeUrl(parentPath);

				filesStore.requestPathItems(parentPath, transferItem.driveType);
			}
		} catch (error) {
			console.log('error ===>', error);
		}
	}

	async transferItemBackToFiles(item: TransferItem, router: Router) {
		await router.push({
			path: item.path
		});
	}

	async renameRepo(_item: SyncRepoItemType, _newName: string): Promise<void> {}

	getUploadNode(): string {
		return '';
	}
	getPanelJumpPath(file: any): string {
		return '';
	}

	formatSearchPath(search: string): string {
		search = search.substring(search.startsWith('/') ? 5 : 4);
		const path = filterPcvPath2(search);
		const formatSearchPath = appendPath('/Files', path);
		return formatSearchPath;
	}
	getOriginalPath(file: FileItem) {
		return this.getAttrPath(file);
	}
}

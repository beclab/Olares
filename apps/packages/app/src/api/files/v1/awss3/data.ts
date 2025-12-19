import { DriveDataAPI } from './../index';
import { format } from './filesFormat';
import {
	FileItem,
	FileResType,
	FilesIdType,
	useFilesStore
} from 'src/stores/files';

import { DriveType } from 'src/utils/interface/files';

import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { OPERATE_ACTION } from 'src/utils/contact';
import { notifyWaitingShow, notifyHide } from 'src/utils/notifyRedefinedUtil';
import {
	move,
	copy,
	remove,
	rename,
	resourceAction,
	download,
	generateDownloadUrl
} from './utils';
import url from 'src/utils/url';
import { createURL } from '../utils';
import { useDataStore } from 'src/stores/data';
import { useTransfer2Store } from 'src/stores/transfer2';
import {
	TransferItem,
	TransferStatus,
	TransferFront
} from 'src/utils/interface/transfer';
import { getFileIcon } from '@bytetrade/core';
import { getextension } from 'src/utils/utils';
import md5 from 'js-md5';
import { i18n } from 'src/boot/i18n';
import { uuid } from '@didvault/sdk/src/core';

export default class Awss3DataAPI extends DriveDataAPI {
	public origin_id: number;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
	}

	breadcrumbsBase = '/Drive/awss3';

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchData(url);
		res.url = url;

		console.log('fetch res', res);

		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(`/api/resources${url}`, {});

		const data: FileResType = format(
			JSON.parse(JSON.stringify(res.data)),
			url,
			this.origin_id
		);

		return data;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return `/Drive/${DriveType.Awss3}/${item.key}/`;
	}

	formatPathtoUrl(path: string): string {
		return `${path}?src=${DriveType.Awss3}`;
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
					url = download('zip', selectFile);
				} else {
					const store = useDataStore();
					const baseURL = store.baseURL();
					url = baseURL + '/api/resources' + selectFile.url + '&stream=1';
				}
			} else {
				url = download(null, selectFile);
			}

			const fileObj: TransferItem = {
				url,
				path: selectFilePath,
				parentPath,
				size: selectFile.size,
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

	async getFileUploadedBytes(
		_filePath: any,
		fileName: any,
		_repo_id: string,
		task_id: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const url = baseURL + '/drive/get_direct_upload_bytes';
		const params = {
			task_id: task_id,
			resumable_relative_path: fileName
		};
		const res = await this.commonAxios.post(url, params);

		return { uploadedBytes: res.data.data };
	}

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const from = this.breadcrumbsBase + decodeURIComponent(el.path).slice(12);
		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: DriveType.Awss3,
			isDir: el.isDir
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
		console.log('paste path', path);

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element: any = operateinStore.copyFiles[i];
			console.log('paste element', element);
			let lastPathIndex =
				path.indexOf('?') > -1
					? decodeURIComponent(path).slice(6, path.indexOf('?'))
					: decodeURIComponent(path).slice(6);

			lastPathIndex = lastPathIndex.endsWith('/')
				? lastPathIndex
				: lastPathIndex + '/';
			let name = element.name;
			try {
				name = decodeURIComponent(element.name);
			} catch (error) {
				/* empty */
			}
			// const to = path + name;
			const to = path + name + (element.isDir ? '/' : '');
			items.push({
				from: element.from,
				to: to,
				name: element.name,
				src_drive_type: element.src_drive_type,
				dst_drive_type: DriveType.Awss3
			});

			console.log('paste item', items);
			if (path + decodeURIComponent(element.name) === element.from) {
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

		console.log('operateinStoreoperateinStore', items);

		await this.action(overwrite, rename, items, path, isMove, callback);
	}

	async move(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void> {
		console.log('move path', path);
		const filesStore = useFilesStore();
		const items: CopyStoragesType[] = [];
		for (const i of filesStore.selected[this.origin_id]) {
			const element = filesStore.getTargetFileItem(i, this.origin_id);
			if (!element) {
				continue;
			}
			console.log('move element', element);
			let name = element.name;
			try {
				name = decodeURIComponent(element.name);
			} catch (error) {
				/* empty */
			}
			const to = path + name + (element.isDir ? '/' : '');
			items.push({
				from: element.path,
				to: to,
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: DriveType.Awss3
			});

			console.log('move items', items);
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
		const notifyId = await uuid();
		notifyWaitingShow(i18n.global.t('Pasting, Please wait...'), notifyId);

		if (isMove) {
			await move(items, overwrite, rename)
				.then(() => {
					callback(OPERATE_ACTION.MOVE, dest);
					notifyHide(notifyId);
				})
				.catch(() => {
					notifyHide(notifyId);
				});
		} else {
			await copy(items, overwrite, rename)
				.then(() => {
					callback(OPERATE_ACTION.PASTE, dest);
					notifyHide(notifyId);
				})
				.catch(() => {
					notifyHide(notifyId);
				});
		}
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		// const filesStore = useFilesStore();

		const promises: any = [];

		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			promises.push(remove(item.path + '?src=' + DriveType.Awss3));
		}

		await Promise.all(promises);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const oldLink = decodeURIComponent(item.path);
		const newLink =
			url.removeLastDir(oldLink) + '/' + encodeURIComponent(newName);

		await rename(oldLink, newLink);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		let url =
			path + '/' + encodeURIComponent(dirName) + '?src=' + DriveType.Awss3;
		url = url.replace('//', '/');
		await resourceAction(url, 'post');
	}

	async openPreview(item: any): Promise<FileResType> {
		return item;
	}

	getPreviewURL(file: any, thumb: string): string {
		const pathSplit = file.path.split('?')[0];

		const params = {
			inline: 'true',
			key: file.modified,
			src: DriveType.Awss3
		};

		return createURL('api/preview/' + thumb + pathSplit, params);
	}

	getDownloadURL(file: any, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' }),
			src: DriveType.Awss3
		};
		const url = createURL('api/raw' + file.path, params);
		return url;
	}

	async getFileServerUploadLink(
		folderPath: string,
		_repo_id?: string,
		dirName?: string
	): Promise<any> {
		let parent_directory = folderPath.split('/').splice(4).join('/');
		if (!parent_directory) {
			parent_directory = '/';
		}
		if (parent_directory != '/' && !parent_directory.startsWith('/')) {
			parent_directory = '/' + parent_directory;
		}
		if (parent_directory != '/' && !parent_directory.endsWith('/')) {
			parent_directory = parent_directory + '/';
		}

		const name = folderPath.split('/')[3];

		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		// const path = folderPath;
		const url = baseURL + '/drive/create_direct_upload_task';
		const res = await this.commonAxios.post(url, {
			parent_directory,
			upload_type: dirName ? 'folder' : 'file',
			new_folder_name: dirName,
			name,
			drive: DriveType.Awss3
		});

		return `/drive/direct_upload_file/${res.data.data.id}?ret-json=1`;
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const path = '/' + file.path.split('/').slice(3).join('/');
		const name = file.path.split('/')[2];
		if (getFileIcon(file.name) === 'image') {
			const url = generateDownloadUrl(file.driveType, path, name);
			return url;
		}
		const params = {
			...(inline && { inline: 'true' }),
			src: DriveType.Awss3
		};
		const url = createURL('api/raw/Drive' + file.path, params);
		return url;
	}

	formatTransferToFileItem(item: TransferItem): FileItem {
		const extension = getextension(item.name);
		const res: FileItem = {
			extension,
			isDir: item.isFolder,
			isSymlink: false,
			mode: 0,
			modified: item.updateTime || 0,
			name: item.name,
			path:
				item.front == TransferFront.download ? `/Drive${item.path}` : item.path,
			size: item.size,
			type: item.type,
			parentPath: item.parentPath,
			index: 0,
			url: item.url || '',
			driveType: item.driveType!,
			fileExtend: '',
			param: ''
		};
		return res;
	}

	formatFolderSubItemDownloadPath(
		item: TransferItem,
		parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	) {
		let parentPath = parentItem.parentPath || '';
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
		if (!file.size && file.fileSize) {
			file.size = file.fileSize;
		}
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;

		const path = parentPath?.split('/').slice(2, 4).join('/');
		file.path = (path?.startsWith('/') ? path : `/${path}`) + file.path;

		file.driveType = DriveType.Awss3;
		file.parentPath = '/' + parentPath?.split('/').splice(2).join('/');
		file.uniqueIdentifier =
			md5(file.relativePath + new Date()) + file.relativePath;

		file.url = this.getFormatSteamDownloadFileURL(file);
	}

	getUploadTransferItemMoreInfo(item: TransferItem) {
		let taskId = item.uniqueIdentifier;
		let folderName = '';
		let relativePath = '';

		let leftPath = '';

		let isFolder = false;

		if (item.task && item.task > 0) {
			const transferStore = useTransfer2Store();

			const parentItem = transferStore.transferMap[item.task];

			taskId = parentItem.uniqueIdentifier;
			folderName = parentItem.name;
			isFolder = true;

			relativePath =
				(item.relatePath?.endsWith('/')
					? item.relatePath.substring(0, item.relatePath.length - 1)
					: item.relatePath) || '';

			leftPath = parentItem.path.substring(
				0,
				parentItem.path.length - parentItem.name.length
			);
		} else {
			leftPath = item.path.substring(0, item.path.length - item.name.length);
		}
		const splitArray = leftPath.split('/');

		const account = splitArray[3];

		let path = leftPath.split('/').slice(4).join('/');
		if (!path || path.length == 0) {
			path = '/';
		} else {
			if (!path.startsWith('/')) {
				path = '/' + path;
			}
			if (path.length > 1 && path.endsWith('/')) {
				path = path.substring(0, path.length - 1);
			}
		}

		return {
			taskId: taskId,
			account: account,
			isFolder,
			cloudFilePath: path,
			folderName,
			relativePath
		};
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
				const res = await this.commonAxios.get(
					`/api/raw${file.path}?src=${DriveType.Awss3}`,
					{}
				);
				file.content = res;
			} catch (error) {
				console.error(error.message);
			}
		}
		return file;
	}

	fileEditEnable = false;

	videoPlayEnable = false;

	audioPlayEnable = false;
}

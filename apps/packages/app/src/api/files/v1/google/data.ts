import { DriveDataAPI } from './../index';
import { getFileIcon } from '@bytetrade/core';
import { format } from './filesFormat';
import { DriveMenuType } from './type';
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
	saveGoogleDirInfo,
	generateDownloadUrl,
	fetchRepo
} from './utils';
import {
	useFilesStore,
	FilesIdType,
	FileItem,
	FileResType
} from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { useDataStore } from 'src/stores/data';
// import url from '../../utils/url';
import { createURL } from '../utils';
import { useTransfer2Store } from 'src/stores/transfer2';
import { encodeUrl } from 'src/utils/encode';
import { getextension } from 'src/utils/utils';
import md5 from 'js-md5';
import { useUserStore } from 'src/stores/user';
import url from 'src/utils/url';
import { Router } from 'vue-router';
import { i18n } from 'src/boot/i18n';
import { uuid } from '@didvault/sdk/src/core';
// import { getParams } from 'src/utils/utils';
// import { checkSameName } from './../../utils/file';

// import { getParams } from '../../utils/utils';

export default class GoogleDataAPI extends DriveDataAPI {
	breadcrumbsBase = '/Drive/google';

	public origin_id: number;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
	}

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchData(url);
		res.url = url;

		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const splitUrl_0 = url.split('?')[0];
		const splitUrl_1 = url.split('?')[1] ? `?${url.split('?')[1]}` : '';
		const splitUrl_2 = splitUrl_0.split('/');
		let cur_url = url;
		if (splitUrl_2[4] === '') {
			cur_url = splitUrl_0 + 'root/' + splitUrl_1;
		}

		const res = await this.commonAxios.get(`/api/resources${cur_url}`, {});

		const data: FileResType = format(
			JSON.parse(JSON.stringify(res.data)),
			cur_url,
			this.origin_id
		);

		saveGoogleDirInfo(data.items, this.origin_id);
		return data;
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		const res1: any = await fetchRepo();
		const imgObj = {
			dropbox: './img/dropbox.svg',
			google: './img/google.svg',
			awss3: './img/awss3.svg',
			tencent: './img/tencent.svg'
		};

		const mineChildren: any[] = [];
		for (let i = 0; i < res1.length; i++) {
			const el = res1[i];
			mineChildren.push({
				label: el.name,
				key: el.name,
				icon: '',
				img: imgObj[el.type],
				defaultHide: true,
				driveType: el.type,
				...el
			});
		}

		return mineChildren;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return `/Drive/${DriveType.GoogleDrive}/${item.key}/`;
	}

	formatPathtoUrl(path: string): string {
		return `${path}?src=${DriveType.GoogleDrive}`;
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

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const from = this.breadcrumbsBase + el.path.slice(13);
		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: DriveType.GoogleDrive
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
			let lastPathIndex =
				path.indexOf('?') > -1
					? path.slice(6, path.indexOf('?'))
					: path.slice(6);

			lastPathIndex = lastPathIndex.endsWith('/')
				? lastPathIndex
				: lastPathIndex + '/';

			let to = path + encodeUrl(element.name);

			if (path.split('/')[4] === '') {
				to = path + 'root/' + encodeUrl(element.name);
			}

			items.push({
				from: element.from,
				to: to,
				name: element.name,
				src_drive_type: element.src_drive_type,
				dst_drive_type: DriveType.GoogleDrive
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

			items.push({
				from: element.path,
				to: path,
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: DriveType.GoogleDrive
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
			promises.push(remove(item.path + '?src=' + DriveType.GoogleDrive));
		}

		await Promise.all(promises);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const oldLink = item.path;
		const newLink = oldLink + encodeUrl(newName);

		await rename(oldLink, newLink);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const splitUrl_2 = path.split('/');
		let cur_url = path;
		if (splitUrl_2[4] === '') {
			cur_url =
				path +
				'/root/' +
				encodeURIComponent(dirName) +
				'?src=' +
				DriveType.GoogleDrive;
		} else {
			cur_url =
				path + encodeURIComponent(dirName) + '?src=' + DriveType.GoogleDrive;
		}

		cur_url = cur_url.replace('//', '/');

		await resourceAction(cur_url, 'post');
	}

	async openPreview(item: any): Promise<FileResType> {
		if (getFileIcon(item.name) === 'video') {
			let url = download(null, item);
			const store = useDataStore();
			const baseURL = store.baseURL();
			url = url.replace(baseURL, '/Drive/google');
			item.path = url;
		}

		return item;
	}

	getPreviewURL(file: any, thumb: string): string {
		const pathSplit = file.path.split('?')[0];

		const params = {
			inline: 'true',
			key: file.modified,
			src: DriveType.GoogleDrive
		};

		return createURL('api/preview/' + thumb + pathSplit, params);
	}

	getDownloadURL(file: any, inline: boolean): string {
		if (getFileIcon(file.name) === 'image') {
			const url = download(null, file);
			return url;
		}

		const params = {
			...(inline && { inline: 'true' }),
			src: DriveType.GoogleDrive
		};
		const url = createURL('api/raw' + file.path, params);
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
				const res = await this.commonAxios.get(
					`/api/raw${file.path}?src=${DriveType.GoogleDrive}`,
					{}
				);
				file.content = res;
			} catch (error) {
				console.error(error.message);
			}
		}
		return file;
	}

	async getFileServerUploadLink(
		folderPath: string,
		_repo_id?: string,
		dirName?: string
	): Promise<any> {
		const parent_directory = folderPath.split('/')[4]
			? folderPath.split('/')[4]
			: '/';
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
			drive: DriveType.GoogleDrive
		});

		return `/drive/direct_upload_file/${res.data.data.id}?ret-json=1`;
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

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const lastIndex = file.path.lastIndexOf('/');
		const secondLastIndex = file.path.lastIndexOf('/', lastIndex - 1);
		const google_id = file.path.substring(secondLastIndex + 1, lastIndex);
		const name = file.path.split('/')[2];
		if (getFileIcon(file.name) === 'image') {
			const url = generateDownloadUrl(file.driveType, google_id, name);
			return url;
		}

		const params = {
			...(inline && { inline: 'true' }),
			src: DriveType.GoogleDrive
		};
		const url = createURL('api/raw/Drive' + file.path, params);
		return url;
	}

	formatFolderSubItemDownloadPath(
		item: TransferItem,
		_parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	) {
		// const parentPath = item.parentPath || '';
		// const path = item.path;

		// const releatePath = path.substring(parentPath.length, path.length);
		let releatePath: string | undefined = undefined;
		if (item.params) {
			const params = JSON.parse(item.params);
			if (params.relatePath) {
				releatePath = params.relatePath;
				if (appendPath != '/') {
					releatePath = releatePath?.replace('/', appendPath);
				}
			}
		}
		const parentSavePath =
			defaultDownloadPath + appendPath + folderSavePath + appendPath;

		const itemSavePath =
			parentSavePath +
			(releatePath && releatePath.length > 0
				? releatePath.endsWith(appendPath)
					? releatePath
					: appendPath + appendPath
				: '');

		return {
			parentSavePath,
			itemSavePath
		};
	}

	getResumePath(path: string) {
		// return path + relativePath;
		// const splitePath = path.split('/');
		// if (path.length <= 5) {
		// 	return path;
		// }
		// return splitePath.splice(0, 4).join('/') + '/';
		return path;
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
			param: '',
			fileExtend: ''
		};

		return res;
	}
	formatSteamDownloadItem(file: any, infoPath?: string, parentPath?: string) {
		if (!file.size && file.fileSize) {
			file.size = file.fileSize;
		}
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;
		file.driveType = DriveType.GoogleDrive;
		file.parentPath = '/' + parentPath?.split('/').splice(2).join('/');
		file.uniqueIdentifier =
			md5(file.relativePath + new Date()) + file.relativePath;

		if (file.id_path && infoPath) {
			if (!file.id_path.startsWith('/')) {
				file.id_path = '/' + file.id_path;
			}

			const lastIndex = infoPath.lastIndexOf('/');
			const secondLastIndex = infoPath.lastIndexOf('/', lastIndex - 1);
			const parent_google_id = infoPath.substring(
				secondLastIndex + 1,
				lastIndex
			);

			const ids = (file.id_path as string).split('/');

			// if (ids.includes(parentPath))
			const index = ids.findIndex((e) => e == parent_google_id);
			if (index >= 0) {
				const paths = (file.path as string).split('/');
				if (paths.length == ids.length) {
					let relatePath = paths.splice(index + 1).join('/');
					if (!relatePath.startsWith('/')) {
						relatePath = '/' + relatePath;
					}
					if (relatePath.endsWith(file.name)) {
						relatePath = relatePath.substring(
							0,
							relatePath.length - file.name.length
						);
					}

					if (relatePath.length > 0 && relatePath != '/') {
						file.params = JSON.stringify({
							relatePath
						});
					}
				}
			}
		}
		const path = parentPath?.split('/').slice(2, 4).join('/');
		file.path =
			(path?.startsWith('/') ? path : `/${path}`) + `/${file.meta.id}/`;
		file.url = this.getFormatSteamDownloadFileURL(file);
	}

	getUploadTransferItemMoreInfo(item: TransferItem) {
		let taskId = item.uniqueIdentifier;
		let folderName = '';
		let isFolder = false;
		let relativePath = '';

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
		}

		const splitArray = item.path.split('/');
		let google_id = '/';
		if (splitArray.length >= 6) {
			google_id = splitArray[4];
		}
		const account = splitArray[3];

		return {
			taskId: taskId,
			account,
			isFolder,
			cloudFilePath: google_id,
			folderName: folderName,
			relativePath
		};
	}

	fileEditEnable = false;

	videoPlayEnable = false;

	audioPlayEnable = false;

	async transferItemUploadSuccessResponse(
		tranfeItemId: number,
		response: string
	) {
		if (!tranfeItemId) {
			return false;
		}
		const res = JSON.parse(response);

		const uploadPathId = res.data.meta.id;

		const transferStore = useTransfer2Store();

		const item =
			transferStore.transferMap[tranfeItemId] ||
			transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

		if (!item || item.id != tranfeItemId) {
			return false;
		}
		const splitArray = item.path.split('/');
		const id_pathArray = (res.data.id_path as string).split('/');

		const parentPathId =
			id_pathArray.length > 3 ? id_pathArray[id_pathArray.length - 2] : '';
		let google_id = '';
		if (splitArray.length >= 6) {
			google_id = splitArray[4];
		}
		if (google_id.length > 0) {
			item.parentPath = item.path.replace(google_id, parentPathId);
			if (item.parentPath.endsWith('//')) {
				item.parentPath = item.parentPath.substring(
					0,
					item.parentPath.length - 1
				);
			}
			item.path = item.path.replace(google_id, uploadPathId);
		} else {
			item.parentPath =
				item.path + (parentPathId.length > 0 ? parentPathId + '/' : '');

			item.path = item.path + uploadPathId + '/';
		}
		item.to = res.data.id_path;

		transferStore.update(tranfeItemId, {
			path: item.path,
			to: item.path,
			parentPath: item.parentPath
		});
		return true;
	}

	async uploadSuccessRefreshData(tranfeItemId: number) {
		const transferStore = useTransfer2Store();
		const userStore = useUserStore();
		const transferItem =
			transferStore.transferMap[tranfeItemId] ||
			transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

		console.log('transferItem ===>', transferItem);

		if (!transferItem) {
			return;
		}

		if (transferItem.userId && transferItem.userId != userStore.current_id) {
			return;
		}

		const filesStore = useFilesStore();
		const cur_file = this.formatTransferToFileItem(transferItem);
		const fullPath = url.getWindowFullpath();
		const splitArray = fullPath.split('/');
		const id_pathArray = transferItem.to?.split('/');
		console.log('fullPath ===>', fullPath);

		if (fullPath.startsWith('/Drive/google')) {
			let google_id = '';
			if (splitArray.length >= 6) {
				google_id = splitArray[4];
			}
			console.log(google_id);
			console.log(id_pathArray);

			if (google_id.length == 0 || id_pathArray?.length == 0) {
				filesStore.setBrowserUrl(fullPath, cur_file.driveType, false);
				return;
			}

			if (
				id_pathArray &&
				id_pathArray.length > 0 &&
				id_pathArray.includes(google_id)
			) {
				filesStore.setBrowserUrl(fullPath, cur_file.driveType, false);
				return;
			}
		}

		const parentPath = cur_file.parentPath;
		if (parentPath) {
			filesStore.requestPathItems(parentPath, transferItem.driveType);
		}
	}

	async transferItemBackToFiles(item: TransferItem, router: Router) {
		if (item.isFolder) {
			await router.push({
				path: item.path
			});
		} else {
			await router.push({
				path: item.parentPath
			});
		}
	}
}

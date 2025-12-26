import { getFileIcon } from '@bytetrade/core';
import Origin from './../origin';
import { MenuItem } from 'src/utils/contact';
import { OPERATE_ACTION } from 'src/utils/contact';
import { useDataStore } from 'src/stores/data';
import { i18n } from 'src/boot/i18n';
import { createURL } from '../utils';

import {
	notifyWaitingShow,
	notifyHide,
	notifySuccess,
	notifyFailed
} from 'src/utils/notifyRedefinedUtil';

import {
	ShareInfoResType,
	SyncRepoSharedType,
	SyncRepoMineType,
	SyncRepoItemType
} from './type';

import { formatSeahub } from './filesFormat';
import { getParams } from 'src/utils/utils';

import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';

import {
	useFilesStore,
	FileItem,
	FileResType,
	FilesIdType
} from 'src/stores/files';
import { fetchRepo } from './utils';

import { CommonFetch } from '../../fetch';
import { isPad } from 'src/utils/platform';
import { useTransfer2Store } from 'src/stores/transfer2';
import { getextension } from 'src/utils/utils';
import { DriveType } from 'src/utils/interface/files';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { decodeUrl, encodeUrl } from 'src/utils/encode';
import md5 from 'js-md5';
import { IUploadCloudParams } from 'src/platform/interface/electron/interface';
import url from 'src/utils/url';
import { useUserStore } from 'src/stores/user';
import { Router } from 'vue-router';

import * as files from './utils';
import { appendPath } from '../path';

import * as filesUtil from '../common/utils';

export default class SyncDataAPI extends Origin {
	public commonAxios: any;

	public origin_id: number;

	breadcrumbsBase = '/Seahub';

	fileEditEnable = true;

	videoPlayEnable = true;

	audioPlayEnable = true;

	fileEditLimitSize = 1024 * 1024;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
		this.commonAxios = CommonFetch;
	}

	async fetch(url: string): Promise<FileResType> {
		let res = await this.commonAxios.get(url, {});
		if (typeof res === 'object') {
			res = JSON.stringify(res);
		}
		const data: FileResType = formatSeahub(
			url,
			JSON.parse(res),
			this.origin_id
		);

		return data;
	}

	async fetchMenuRepo(): Promise<SyncRepoMineType[]> {
		// const [res2, res3]: any = await fetchRepo(MenuItem.SHAREDWITH);
		// const shareChildren: SyncRepoSharedType[] = [];
		const defaultHide = !isPad();
		// for (let i = 0; i < res2.length; i++) {
		// 	const el = res2[i];
		// 	const hsaShareRepo = shareChildren.find((item) => item.id === el.repo_id);
		// 	if (hsaShareRepo) {
		// 		continue;
		// 	}

		// 	shareChildren.push({
		// 		label: el.repo_name,
		// 		key: el.repo_id,
		// 		icon: 'sym_r_folder_shared',
		// 		id: el.repo_id,
		// 		defaultHide: defaultHide,
		// 		driveType: DriveType.Sync,
		// 		...el
		// 	});
		// }

		// const sharedme: SyncRepoSharedType[] = [];
		// for (let i = 0; i < res3.length; i++) {
		// 	const el = res3[i];
		// 	sharedme.push({
		// 		label: el.repo_name,
		// 		key: el.repo_id,
		// 		name: el.repo_name,
		// 		icon: 'sym_r_folder_supervised',
		// 		id: el.repo_id,
		// 		defaultHide: defaultHide,
		// 		driveType: DriveType.Sync,
		// 		...el
		// 	});
		// }

		const res1: any = await fetchRepo(MenuItem.MYLIBRARIES);
		const mineChildren: SyncRepoMineType[] = [];
		for (let i = 0; i < res1.length; i++) {
			const el = res1[i];

			// const hasShareWith = shareChildren.find(
			// 	(item) => item.repo_id === el.repo_id
			// );
			// const hasShareMe = sharedme.find((item) => item.repo_id === el.repo_id);
			// const hasShare = hasShareWith;

			mineChildren.push({
				label: el.repo_name,
				key: el.repo_id,
				icon: 'sym_r_folder',
				id: el.repo_id,
				name: el.repo_name,
				shard_user_hide_flag: false,
				// share_type: hasShare ? hasShare.share_type : undefined,
				// user_email: hasShare ? hasShare.user_email : undefined,
				defaultHide: defaultHide,
				driveType: DriveType.Sync,
				fileType: 'sync',
				fileExtend: el.repo_id,
				path: `/Seahub/${el.repo_name}`,
				modified: new Date(el.last_modified).getTime(),
				oPath: '/',
				...el
			});
		}

		// const myLibraries = {
		// 	label: i18n.global.t(`files_menu.${MenuItem.MYLIBRARIES}`),
		// 	key: 'MyLibraries',
		// 	icon: '',
		// 	expationFlag: true,
		// 	muted: true,
		// 	disableClickable: true,
		// 	driveType: DriveType.Sync
		// };
		// const shardWith = {
		// 	label: i18n.global.t(`files_menu.${MenuItem.SHAREDWITH}`),
		// 	key: 'SharedLibraries',
		// 	icon: '',
		// 	expationFlag: false,
		// 	muted: true,
		// 	disableClickable: true,
		// 	driveType: DriveType.Sync
		// };

		// let shardArr: any = [];
		// if (shareChildren.length > 0 || sharedme.length > 0) {
		// 	shardArr = [shardWith, ...shareChildren, ...sharedme];
		// }

		// myLibraries,
		const syncMenu = [...mineChildren];
		console.log('syncMenu ===>', syncMenu);

		return syncMenu as any;
	}

	async fetchShareInfo(repo_id: string): Promise<ShareInfoResType> {
		// const res: ShareInfoResType = await this.commonAxios.get(
		// 	`/seahub/api/v2.1/repos/${repo_id}/share-info/`,
		// 	{}
		// );
		return {
			shared_group_ids: [],
			shared_user_emails: []
		};
	}

	formatPathtoUrl(path: string, param?: string): string {
		const repo_id = getParams(param ? param : path, 'id');
		const cur_path = path.slice(0, path.indexOf('?'));

		const pathList = cur_path.split('/');
		let paths = '';
		for (let i = 3; i < pathList.length; i++) {
			const p = pathList[i];
			paths += `/${p}`;
		}

		return files.syncCommonUrl(
			'resources',
			paths.endsWith('/') ? paths : paths + '/',
			repo_id
		);
	}

	async getDownloadInfo(path: string): Promise<TransferItem[]> {
		const filesStore = useFilesStore();
		const selected = filesStore.selected[this.origin_id];
		const result: TransferItem[] = [];

		for (let i = 0; i < selected.length; i++) {
			const index = selected[i];
			const selectFile = filesStore.getTargetFileItem(index, this.origin_id);

			if (!selectFile) {
				continue;
			}

			const url = filesUtil.getDownloadUrl(selectFile);
			const fileObj: TransferItem = {
				url,
				path: decodeUrl(selectFile.path),
				parentPath: selectFile.oParentPath,
				size: selectFile.isDir ? 0 : selectFile.size,
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

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const repo_id = getParams(el.path, 'id');
		const fromPath = files.formatPathtoUrl(el.parentPath + el.name, repo_id);
		const from = el.isDir ? fromPath + '/' : fromPath;

		const copyItem: CopyStoragesType = {
			from: encodeUrl(from),
			to: '',
			name: el.name,
			src_drive_type: DriveType.Sync,
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
		const filesStore = useFilesStore();

		const items: CopyStoragesType[] = [];
		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const el: any = operateinStore.copyFiles[i];

			const repo_id = getParams(path, 'id');
			const pathFromStart =
				decodeURIComponent(path).indexOf(
					filesStore.activeMenu(this.origin_id).label
				) + filesStore.activeMenu(this.origin_id).label.length;
			const pathFromEnd = decodeURIComponent(path).lastIndexOf('?');

			const to = files.formatPathtoUrl(
				decodeURIComponent(path).slice(pathFromStart, pathFromEnd) +
					el.name +
					(el.from.endsWith('/') ? '/' : ''),
				repo_id
			);
			items.push({
				...el,
				to: encodeUrl(to),
				dst_drive_type: DriveType.Sync
			});
			if (path + el.name === el.from) {
				return await this.action(false, true, items, path, false, callback);
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

		return await this.action(overwrite, rename, items, path, isMove, callback);
	}

	async move(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void> {
		const filesStore = useFilesStore();
		const items: CopyStoragesType[] = [];
		for (const i of filesStore.selected[this.origin_id]) {
			const el = filesStore.getTargetFileItem(i, this.origin_id);
			if (!el) {
				continue;
			}

			const pathFromStart =
				decodeURIComponent(el.path).indexOf(
					filesStore.activeMenu(this.origin_id).label
				) + filesStore.activeMenu(this.origin_id).label.length;

			const pathFromEnd = decodeURIComponent(el.path).lastIndexOf('?');
			const repo_id = getParams(el.path, 'id');

			const from =
				'/' +
				repo_id +
				decodeURIComponent(el.path).slice(pathFromStart, pathFromEnd) +
				(el.isDir ? '' : el.name);

			const toStart =
				decodeURIComponent(path).indexOf(
					filesStore.activeMenu(this.origin_id).label
				) + filesStore.activeMenu(this.origin_id).label.length;
			const toEnd = decodeURIComponent(path).lastIndexOf('?');
			const to =
				'/' +
				repo_id +
				decodeURIComponent(path).slice(toStart, toEnd) +
				el.name;

			items.push({
				from: encodeUrl(from),
				to: encodeUrl(to),
				name: el.name,
				src_drive_type: el.driveType,
				dst_drive_type: DriveType.Sync
			});
		}
		const overwrite = true;
		return await this.action(overwrite, true, items, path, true, callback);
	}

	async action(
		overwrite: boolean | undefined,
		rename: boolean | undefined,
		items: CopyStoragesType[],
		path: string,
		isMove: boolean | undefined,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<any> {
		return await filesUtil.action(items, path, isMove, callback);
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

	async openPreview(item: any): Promise<FileResType> {
		const selfItem = JSON.parse(JSON.stringify(item));
		selfItem.url = selfItem.path;
		await this.formatFileContent(selfItem);
		selfItem.driveType = DriveType.Sync;
		return selfItem;
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		let repo_id =
			getParams(file.path, 'id') || getParams((files as any).to, 'id');
		if (!repo_id) {
			repo_id = file.path.split('/')[5];
		}

		if (['video'].includes(file.type)) {
			const url = new URL(file.url);
			return files.formatPathtoUrl(url.pathname, repo_id);
		}

		let parentPath = '';
		if (file.parentPath) {
			parentPath = file.parentPath;
		} else {
			parentPath = this.getParentPath(file.path);
		}

		const path = appendPath(decodeUrl(parentPath), file.name);

		const params = {
			inline: 'true',
			key: file.modified,
			thumb: thumb
		};

		const url = createURL(
			files.syncCommonUrl('preview', path, repo_id),
			params
		);

		return url;
	}

	findNthOccurrence(str, char, n) {
		let count = 0;
		let index = -1;
		while (count < n) {
			index = str.indexOf(char, index + 1);
			if (index === -1) break;
			count++;
		}
		return index;
	}

	getDownloadURL(file: any, inline?: boolean, download?: boolean): string {
		file.parentPath = file.parent_dir || file.parentPath;
		const repo_id = getParams(file.path, 'id');
		if (['audio', 'video', 'pdf'].includes(file.type) && !download) {
			return file.url;
		} else {
			const params = {
				...(inline && { inline: 'true' })
			};

			const file_path = files.syncCommonUrl(
				'raw',
				this.getFileItemPath(file),
				repo_id
			);

			const url = createURL(decodeUrl(file_path), params);

			return url;
		}
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const fileType = getFileIcon(file.name);

		if (
			!['text', 'txt', 'textImmutable', 'audio', 'video', 'pdf'].includes(
				fileType
			)
		) {
			return file;
		}

		try {
			const repo_id = getParams(file.path, 'id');
			let parentPath = encodeUrl(file.parentPath || '');
			if (!parentPath) {
				parentPath = this.getParentPath(file.path);
			}
			const url = files.syncCommonUrl(
				'raw',
				appendPath(parentPath, encodeUrl(file.name)),
				repo_id
			);
			const res = await this.commonAxios.get(url, {
				params: {
					inline: true,
					dict: 1,
					meta: ['audio', 'video', 'pdf'].includes(fileType) ? 'true' : 'false'
				}
			});

			if (['audio', 'video', 'pdf'].includes(fileType)) {
				const store = useDataStore();
				file.url = store.baseURL() + res.raw_path;
			} else if (['text', 'txt', 'textImmutable'].includes(fileType)) {
				file.content = res;
			}
		} catch (error) {
			console.error(error.message);
		}
		return file;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return `/Seahub/${encodeUrl(item.label)}/?id=${item.id}&type=${
			item.type || 'mine'
		}&p=${item.permission ? item.permission.trim() : 'rw'}`;
	}

	async formatUploaderPath(path: string): Promise<string> {
		const splitPath = path.split('/');
		let slotPath = '/';
		for (let i = 3; i < splitPath.length; i++) {
			const el = splitPath[i];
			if (el) slotPath = slotPath + el + '/';
		}

		if (!slotPath.endsWith('/')) {
			slotPath = slotPath + '/';
		}

		return slotPath;
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		await filesUtil.batchDeleteFileItems(items);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		await filesUtil.renameFileItem(item, newName);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const filesStore = useFilesStore();
		const repoId = filesStore.activeMenu(this.origin_id).id;
		const pathlen =
			decodeURIComponent(path).indexOf(
				filesStore.activeMenu(this.origin_id).label
			) + filesStore.activeMenu(this.origin_id).label.length;
		const p = `${decodeURIComponent(path).slice(pathlen)}${dirName}/`;
		await files.createDir(p, repoId);
	}

	async deleteRepo(item: SyncRepoItemType): Promise<void> {
		notifyWaitingShow(i18n.global.t('Deleting, Please wait...'));
		try {
			await files.deleteRepo({
				repoId: item.repo_id
			});
			notifySuccess('Successful!');
		} catch (error) {
			notifyFailed('Failed!');
		}
		notifyHide();
	}

	async renameRepo(item: SyncRepoItemType, newName: string): Promise<void> {
		await files.renameRepo({
			repoId: item.repo_id,
			destination: encodeURIComponent(newName)
		});
	}

	getAttrPath(item: FileItem): string {
		const decodePath = decodeURIComponent(item.path);
		const path = decodePath.slice(
			0,
			decodePath.indexOf('?') > 0 ? decodePath.indexOf('?') : decodePath.length
		);
		return path.slice(0, path.indexOf(item.name));
	}

	async getFileServerUploadLink(
		folderPath: string,
		repoID?: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const path = appendPath('/sync', repoID || '', folderPath);
		const node = this.getUploadNode();
		const url =
			baseURL +
			`/upload/upload-link/${node}/?file_path=` +
			encodeUrl(path) +
			'&from=web';

		const res = await this.commonAxios.get(url);

		return res + '?ret-json=1';
	}

	async getFileUploadedBytes(
		filePath: any,
		fileName: any,
		repoID: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const node = this.getUploadNode();
		const url = baseURL + '/upload/file-uploaded-bytes/' + node + '/';
		const params = {
			parent_dir: appendPath('/sync', repoID, decodeUrl(filePath)),
			file_name: fileName
		};
		return this.commonAxios.get(url, { params: params });
	}

	async getCurrentRepoInfo(path: string): Promise<any> {
		const res = await this.fetch(this.formatPathtoUrl(path));
		return res;
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		const { repo_id } = this.getRepoIdByFileItem(item);
		files.updateFile(
			appendPath(item.parentPath || '', item.name, item.isDir ? '/' : ''),
			repo_id,
			content
		);
	}

	private getParentPath(path: string) {
		const thirdSlashIndex = this.findNthOccurrence(path, '/', 3);
		const questionMarkIndex = path.indexOf('?');
		if (
			thirdSlashIndex === -1 ||
			questionMarkIndex === -1 ||
			thirdSlashIndex >= questionMarkIndex
		) {
			return '/';
		}

		return path.substring(thirdSlashIndex, questionMarkIndex);
	}

	getResumePath(path: string, relativePath: string) {
		const relativePathNoName = relativePath.slice(
			0,
			relativePath.lastIndexOf('/') + 1
		);
		const pathArr = path.split('?');
		return decodeUrl(pathArr[0]) + relativePathNoName + '?' + pathArr[1];
	}

	formatTransferToFileItem(item: TransferItem): FileItem {
		const extension = getextension(item.name);
		const fullPathSplit = item.path.split('?');
		const path = `${
			item.front == TransferFront.download
				? encodeUrl(fullPathSplit[0])
				: fullPathSplit[0]
		}?${fullPathSplit[1]}`;
		const parentPath = item.parentPath
			? this.getParentPath(item.parentPath)
			: '';
		const res: FileItem = {
			extension,
			isDir: item.isFolder,
			isSymlink: false,
			mode: 0,
			modified: item.updateTime || 0,
			name: item.name,
			path,
			size: item.size,
			type: item.type,
			parentPath: parentPath == '/' ? '' : parentPath,
			index: 0,
			url: item.url || '',
			driveType: item.driveType!,
			param: '',
			fileExtend: 'sync'
		};

		return res;
	}

	formatTransferPath(item: TransferItem) {
		const path = item.path.slice(0, item.path.indexOf('?'));
		const pathArr = path.split('/');
		const arr: string[] = [];
		for (let i = 2; i < pathArr.length; i++) {
			const p: string = pathArr[i];
			arr.push(p);
		}
		const res = decodeURIComponent(arr.join('/'));
		if (res.startsWith('/')) {
			return res;
		} else {
			return '/' + res;
		}
	}

	async formatUploadTransferPath(item: TransferItem) {
		const uploadPathSplit = item.path.split('?')[0];
		let uploadPath =
			item.relatePath && item.relatePath.length > 1
				? uploadPathSplit.split(item.relatePath)[0]
				: uploadPathSplit;
		if (!uploadPath.startsWith('/')) {
			uploadPath = '/' + uploadPath;
		}
		const pathname = (await this.formatUploaderPath(uploadPath)) || '/';
		return pathname;
	}

	getPurePath(path: string) {
		return path;
	}

	getDiskPath(selectFiles: any) {
		const repo_id = getParams(selectFiles.path, 'id');
		const resPath = appendPath(
			'/',
			repo_id,
			selectFiles.parentPath,
			selectFiles.name
		);
		return resPath;
	}

	getFormatSteamDownloadFileURL(file: any): string {
		file.parentPath = file.parent_dir || file.parentPath;
		const repo_id = getParams(file.path, 'id');
		const url = files.downloaFile(file, repo_id);
		return url;
	}
	formatFolderSubItemDownloadPath(
		item: TransferItem,
		parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	) {
		let parentPath = decodeUrl(parentItem.parentPath || '');
		parentPath =
			(parentPath.endsWith('/') ? parentPath : `${parentPath}/`) +
			parentItem.name +
			'/';
		let path = decodeUrl(item.path || '');
		const realPath = path.split('?');
		if (realPath.length > 1) {
			path = realPath[0];
		}

		const releatePath = path.substring(parentPath.length);

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
	formatSteamDownloadItem(file: any, infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		const query = infoPath?.slice(infoPath.indexOf('?'));
		const parentSep = parentPath?.split('/').slice(0, 3).join('/');
		const realPath = parentSep + file.parent_dir + query;
		file.path = realPath;
		file.type = fileType;
		file.relativePath = file.parent_dir + file.name;
		file.driveType = DriveType.Sync;

		const addParentPath =
			parentPath + (file.parent_dir ? file.parent_dir.substring(1) : '');

		file.parentPath = addParentPath;
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

	async transferItemUploadSuccessResponse(tranfeItemId: number, response: any) {
		return true;
	}

	async uploadEmptyFile(tranfeItemId: number) {
		const transferStore = useTransfer2Store();
		const transferItem =
			transferStore.transferMap[tranfeItemId] ||
			transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

		if (!transferItem || transferItem.size > 0) {
			return undefined;
		}

		const path = appendPath(
			this.utilsFormatPathtoUrl(transferItem.path),
			transferItem.name
		);
		return await filesUtil.postCreateFile(path, false, undefined);
	}

	utilsFormatPathtoUrl(path: string) {
		const filesStore = useFilesStore();
		const repo_id = getParams(path, 'id');
		const pathFromStart =
			decodeURIComponent(path).indexOf(
				filesStore.activeMenu(this.origin_id).label
			) + filesStore.activeMenu(this.origin_id).label.length;
		const pathFromEnd = decodeURIComponent(path).lastIndexOf('?');

		const to = files.formatPathtoUrl(
			decodeURIComponent(path).slice(pathFromStart, pathFromEnd),
			repo_id
		);
		return to;
	}

	async uploadSuccessRefreshData(tranfeItemId: number) {
		try {
			const transferStore = useTransfer2Store();
			const transferItem =
				transferStore.transferMap[tranfeItemId] ||
				transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

			const userStore = useUserStore();

			if (!transferItem) {
				return;
			}

			if (transferItem.userId && transferItem.userId != userStore.current_id) {
				return;
			}

			const path = url.getWindowFullpath();
			const filesStore = useFilesStore();

			let pathFullpath = path.split('?')[0];

			if (this.origin_id) {
				pathFullpath = filesStore.currentPath[this.origin_id].path;
			}

			const cur_file = this.formatTransferToFileItem(transferItem);

			let decodeCurlFilePath = cur_file.path;
			try {
				decodeCurlFilePath = decodeURIComponent(cur_file.path);
			} catch (error) {
				console.log('error', error);
			}

			let decoodeFullPath = pathFullpath;
			try {
				decoodeFullPath = decodeURIComponent(pathFullpath);
			} catch (error) {
				console.log('error', error);
			}

			if (decodeCurlFilePath.indexOf(decoodeFullPath) >= 0) {
				filesStore.setBrowserUrl(
					path,
					cur_file.driveType,
					false,
					this.origin_id
				);
			} else {
				const fullPathSplit = cur_file.path.split('?');

				const parentPath = fullPathSplit[0];

				let decodeCurlFilePath = parentPath;
				try {
					decodeCurlFilePath = decodeURIComponent(decodeCurlFilePath);
				} catch (error) {
					console.log('error', error);
				}

				const url = encodeUrl(decodeCurlFilePath) + '?' + fullPathSplit[1];

				filesStore.requestPathItems(url, transferItem.driveType);
			}
		} catch (error) {
			console.log('sync refresh data error', error);
		}
	}

	async transferItemBackToFiles(item: TransferItem, router: Router) {
		const path = this.resolveUrl(item.path);
		await router.push(path);
	}

	private resolveUrl(url: string) {
		const urlObj = new URL(url, 'https://desktop.myterminus.com');
		const path = urlObj.pathname;
		const query = Object.fromEntries(urlObj.searchParams);

		return {
			path,
			query
		};
	}

	private getRepoIdByFileItem(item: FileItem) {
		const repo_id = getParams(item.path, 'id');
		return {
			repo_id
		};
	}

	private getFileItemPath(item: FileItem) {
		return appendPath(
			encodeUrl(item.parentPath || '/'),
			encodeUrl(item.name),
			item.isDir ? '/' : ''
		);
	}

	formatCopyPath(path: string, destname: string, isDir: boolean): string {
		const filesStore = useFilesStore();
		const syncMenu = filesStore.menu[this.origin_id].find(
			(item) => item.key === 'Sync'
		)?.children;

		const repo_id = path.split('/')[2];

		const repo_name = syncMenu?.find((item) => item.key === repo_id)?.label;

		const new_path = path.split('/').slice(3).join('/');

		const relativePathNoName = new_path.slice(0, new_path.lastIndexOf('/') + 1);

		const res_path = `/Seahub/${repo_name}/${relativePathNoName}?id=${repo_id}`;

		return res_path;
	}

	getPanelJumpPath(file: any): string {
		return file.path;
	}

	formatSearchPath(search: string): string {
		const formatSearchPath = search.startsWith('/') ? search : '/' + search;
		return formatSearchPath;
	}

	getUploadNode(): string {
		const filesStore = useFilesStore();
		return filesStore.masterNode;
	}

	displayPath(file: {
		isDir: boolean;
		fileExtend: string;
		path: string;
		fileType: string;
	}): string {
		return '';
	}

	getOriginalPath(file: FileItem) {
		return this.getAttrPath(file);
	}

	pathToFrontendFile(path: string) {
		return {
			isDir: path.endsWith('/'),
			fileExtend: '',
			path: ''
		};
	}
}

import { getFileIcon } from '@bytetrade/core';
import Origin from './../origin';
import { MenuItem } from 'src/utils/contact';
import { OPERATE_ACTION } from 'src/utils/contact';
import { useDataStore } from 'src/stores/data';
import { files } from './../index';
import * as seahub from './utils';
import { i18n } from 'src/boot/i18n';
import { getNotifyMsg } from '../utils';

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

export default class SyncDataAPI extends Origin {
	public commonAxios: any;

	public origin_id: number;

	fileEditEnable = true;

	videoPlayEnable = true;

	audioPlayEnable = true;

	fileEditLimitSize = 1024 * 1024;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
		this.commonAxios = CommonFetch;
	}

	breadcrumbsBase = '/Seahub';

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
		const [res2, res3]: any = await fetchRepo(MenuItem.SHAREDWITH);
		const shareChildren: SyncRepoSharedType[] = [];
		const defaultHide = !isPad();
		for (let i = 0; i < res2.length; i++) {
			const el = res2[i];
			const hsaShareRepo = shareChildren.find((item) => item.id === el.repo_id);
			if (hsaShareRepo) {
				continue;
			}

			shareChildren.push({
				label: el.repo_name,
				key: el.repo_id,
				icon: 'sym_r_folder_shared',
				id: el.repo_id,
				defaultHide: defaultHide,
				driveType: DriveType.Sync,
				...el
			});
		}

		const sharedme: SyncRepoSharedType[] = [];
		for (let i = 0; i < res3.length; i++) {
			const el = res3[i];
			sharedme.push({
				label: el.repo_name,
				key: el.repo_id,
				name: el.repo_name,
				icon: 'sym_r_folder_supervised',
				id: el.repo_id,
				defaultHide: defaultHide,
				driveType: DriveType.Sync,
				...el
			});
		}

		const res1: any = await fetchRepo(MenuItem.MYLIBRARIES);
		const mineChildren: SyncRepoMineType[] = [];
		for (let i = 0; i < res1.length; i++) {
			const el = res1[i];

			const hasShareWith = shareChildren.find(
				(item) => item.repo_id === el.repo_id
			);
			const hasShareMe = sharedme.find((item) => item.repo_id === el.repo_id);
			const hasShare = hasShareWith || hasShareMe;

			mineChildren.push({
				label: el.repo_name,
				key: el.repo_id,
				icon: 'sym_r_folder',
				id: el.repo_id,
				name: el.repo_name,
				shard_user_hide_flag: hasShare ? false : true,
				share_type: hasShare ? hasShare.share_type : undefined,
				user_email: hasShare ? hasShare.user_email : undefined,
				defaultHide: defaultHide,
				driveType: DriveType.Sync,
				...el
			});
		}

		const myLibraries = {
			label: i18n.global.t(`files_menu.${MenuItem.MYLIBRARIES}`),
			key: 'MyLibraries',
			icon: '',
			expationFlag: true,
			muted: true,
			disableClickable: true,
			driveType: DriveType.Sync
		};
		const shardWith = {
			label: i18n.global.t(`files_menu.${MenuItem.SHAREDWITH}`),
			key: 'SharedLibraries',
			icon: '',
			expationFlag: false,
			muted: true,
			disableClickable: true,
			driveType: DriveType.Sync
		};

		let shardArr: any = [];
		if (shareChildren.length > 0 || sharedme.length > 0) {
			shardArr = [shardWith, ...shareChildren, ...sharedme];
		}

		const syncMenu = [myLibraries, ...mineChildren, ...shardArr];

		return syncMenu;
	}

	async fetchShareInfo(repo_id: string): Promise<ShareInfoResType> {
		const res: ShareInfoResType = await this.commonAxios.get(
			`/seahub/api/v2.1/repos/${repo_id}/share-info/`,
			{}
		);
		return res;
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

			let url = '';

			if (selectFile.isDir) {
				if (process.env.APPLICATION === 'FILES') {
					url = await seahub.downloaFileZip(selectFile, this.origin_id);
				} else {
					const repo_id = getParams(selectFile.path, 'id');
					const store = useDataStore();
					const baseURL = store.baseURL();
					url =
						baseURL +
						'/api/resources/' +
						repo_id +
						encodeUrl(selectFile.parentPath || '/') +
						encodeUrl(selectFile.name) +
						'?stream=1&src=sync';
				}
			} else {
				url = await seahub.downloaFile(selectFile, this.origin_id);
			}

			console.log('url ===>', url);

			const fileObj: TransferItem = {
				url,
				path: decodeUrl(selectFile.path),
				parentPath: decodeUrl(path),
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
		const repo_id = getParams(el.path, 'id');
		const from = '/' + repo_id + el.parentPath + el.name;
		const copyItem: CopyStoragesType = {
			from: encodeUrl(from),
			to: '',
			name: el.name,
			src_drive_type: DriveType.Sync
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

			const to =
				'/' +
				repo_id +
				decodeURIComponent(path).slice(pathFromStart, pathFromEnd) +
				el.name;
			items.push({
				from: el.from,
				to: encodeUrl(to),
				name: el.name,
				src_drive_type: el.src_drive_type,
				dst_drive_type: DriveType.Sync
			});
			if (path + el.name === el.from) {
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
		const filesStore = useFilesStore();
		const item =
			filesStore.currentFileList[this.origin_id]?.items[
				filesStore.selected[this.origin_id][0]
			];
		if (!item || !item.isDir) {
			return undefined;
		}
		const itemUrl = decodeURIComponent(item.path);
		const pathFromStart =
			itemUrl.indexOf(filesStore.activeMenu(this.origin_id).label) +
			filesStore.activeMenu(this.origin_id).label.length;
		const path = itemUrl.slice(pathFromStart, itemUrl.length - 1);
		return path;
	}

	async openPreview(item: any): Promise<FileResType> {
		const selfItem = JSON.parse(JSON.stringify(item));
		selfItem.url = selfItem.path;
		await this.formatFileContent(selfItem);
		selfItem.driveType = DriveType.Sync;
		return selfItem;
	}

	getPreviewURL(file: any, thumb: string): string {
		if (['video'].includes(file.type)) {
			const url = new URL(file.url);
			return '/Seahub' + url.pathname;
		}
		const dataStore = useDataStore();
		let repo_id = getParams(file.path, 'id');
		if (!repo_id) {
			repo_id = file.path.split('/')[5];
		}
		let seflSize = '1080';
		if (thumb === 'thumb') {
			seflSize = '128';
		}

		let parentPath = '';
		if (file.parentPath) {
			parentPath = encodeUrl(file.parentPath);
		} else {
			parentPath = this.getParentPath(file.path);
		}

		const path = `${parentPath}/${encodeUrl(file.name)}`;

		return `${dataStore.baseURL()}/seahub/thumbnail/${repo_id}/${seflSize}${path}`;
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
		const store = useDataStore();
		const repo_id = getParams(file.path, 'id');

		if (['audio', 'video', 'pdf'].includes(file.type) && !download) {
			return file.url;
		} else {
			return `${store.baseURL()}/seahub/lib/${repo_id}/file${encodeUrl(
				file.parentPath || '/'
			)}${encodeUrl(file.name)}?dl=1`;
		}
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const store = useDataStore();
		const fileType = getFileIcon(file.name);

		if (
			!['audio', 'video', 'text', 'txt', 'textImmutable', 'pdf'].includes(
				fileType
			)
		) {
			return file;
		}

		const repo_id = getParams(file.path, 'id');

		let parentPath = encodeUrl(file.parentPath || '');
		if (!parentPath) {
			parentPath = this.getParentPath(file.path);
		}

		const res = await this.commonAxios.get(
			`/seahub/lib/${repo_id}/file${parentPath}/${encodeUrl(file.name)}?dict=1`,
			{}
		);

		if (['audio', 'video', 'pdf'].includes(fileType)) {
			// file.path = store.baseURL() + res.raw_path;
			const store = useDataStore();
			file.url = store.baseURL() + res.raw_path;
		} else if (['text', 'txt', 'textImmutable'].includes(fileType)) {
			file.content = res.file_content;
		}

		return file;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return `/Seahub/${encodeUrl(item.label)}/?id=${item.id}&type=${
			item.type || 'mine'
		}&p=${item.permission ? item.permission.trim() : 'rw'}`;
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

		return `/api/resources/${repo_id}/${paths}?src=sync`;

		// return `/seahub/api/v2.1/repos/${repo_id}/dir/?p=${paths}&with_thumbnail=true`;
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
		const filesStore = useFilesStore();
		const dirents: string[] = [];

		for (let i = 0; i < items.length; i++) {
			const item = items[i];
			dirents.push(item.name);
		}

		const parmas = {
			dirents: dirents,
			parent_dir: items[0].parentPath,
			repo_id: filesStore.activeMenu(this.origin_id).id
		};

		await seahub.batchDeleteItem(parmas);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		const p = item.parentPath + item.name;
		const parmas = {
			operation: 'rename',
			newname: newName
		};
		const url = 'api/v2.1/repos';
		await seahub.fileOperate(
			p,
			url,
			parmas,
			item.isDir ? 'dir' : 'file',
			this.origin_id
		);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const filesStore = useFilesStore();
		const pathlen =
			decodeURIComponent(path).indexOf(
				filesStore.activeMenu(this.origin_id).label
			) + filesStore.activeMenu(this.origin_id).label.length;
		const p = `${decodeURIComponent(path).slice(pathlen)}${dirName}`;

		const parmas = {
			operation: 'mkdir'
		};
		const url = 'api2/repos';
		await seahub.fileOperate(p, url, parmas, 'dir', this.origin_id);
	}

	async deleteRepo(item: SyncRepoItemType): Promise<void> {
		notifyWaitingShow(i18n.global.t('Deleting, Please wait...'));
		const path = `seahub/api/v2.1/repos/${item.repo_id}/`;
		try {
			await seahub.deleteRepo(path);
			notifySuccess('Successful!');
		} catch (error) {
			notifyFailed('Failed!');
		}
		notifyHide();
	}

	async renameRepo(item: SyncRepoItemType, newName: string): Promise<void> {
		notifyWaitingShow('Renaming, Please wait...');
		const url = `seahub/api2/repos/${item.repo_id}/?op=rename`;
		const data = {
			repo_name: newName
		};
		try {
			await seahub.reRepoName(url, data);
			notifySuccess('Successful!');
		} catch (error) {
			notifyFailed('Failed!');
		}
		notifyHide();
	}

	getAttrPath(item: FileItem): string {
		const path = decodeURIComponent(item.path).slice(0, item.path.indexOf('?'));
		const lastIndex = path.lastIndexOf('/');
		if (lastIndex === -1) return path;
		return path.substring(0, lastIndex);
	}

	async getFileServerUploadLink(
		folderPath: string | number | boolean,
		repoID?: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const path = encodeURIComponent(folderPath);
		const url =
			baseURL +
			'/seahub/api2/repos/' +
			repoID +
			'/upload-link/?p=' +
			path +
			'&from=web';
		const res = this.commonAxios.get(url);

		return res + '?ret-json=1';
	}

	async getFileUploadedBytes(
		filePath: any,
		fileName: any,
		repoID: string
	): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const url =
			baseURL + '/seahub/api/v2.1/repos/' + repoID + '/file-uploaded-bytes/';
		const params = {
			parent_dir: filePath,
			file_name: fileName
		};
		return this.commonAxios.get(url, { params: params });
	}

	async getCurrentRepoInfo(path: string): Promise<any> {
		const filesStore = useFilesStore();
		const [beforeQuestionMark, afterQuestionMark] = path.split(/(\?)/);

		let currentTab = '';
		const splitPath = beforeQuestionMark.split('/');

		let newPath = '';

		if (splitPath.length > 4) {
			currentTab = splitPath[splitPath.length - 2];
			for (let i = 0; i < splitPath.length - 2; i++) {
				const el = splitPath[i];
				newPath += el + '/';
			}
		} else {
			newPath = beforeQuestionMark;
			currentTab = splitPath[splitPath.length - 2];
		}

		const afterIncludingQuestionMark = afterQuestionMark
			? afterQuestionMark + path.split('?')[1]
			: '';

		const key = filesStore.registerUniqueKey(
			newPath,
			DriveType.Sync,
			afterIncludingQuestionMark
		);

		return filesStore.cached[key]?.items.find(
			(item) => item.name === currentTab
		);
	}

	async onSaveFile(item: any, content: any, isNative?: boolean): Promise<void> {
		seahub.updateFile(item, content, isNative, this.origin_id);
	}

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	// getRelativePath(file: any, _parentPath: string) {
	// 	const repo_name = decodeURIComponent(
	// 		urlFormat.getWindowPathname().split('/')[2]
	// 	);
	// 	const path = file.parent_dir + file.name;
	// 	const repo_id = getParams(window.location.href, 'id');
	// 	const type = getParams(window.location.href, 'type');
	// 	const p = getParams(window.location.href, 'p');

	// 	const itemPath = `/Seahub/${repo_name}${path}?id=${repo_id}&type=${type}&p=${p}`;

	// 	return {
	// 		...file,
	// 		path: itemPath,
	// 		relativePath: file.parent_dir + file.name
	// 	};
	// }

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
		// const params = item.params ? item.params : '';
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
			fileExtend: ''
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
		const resPath = `/${repo_id}${selectFiles.parentPath}${selectFiles.name}`;
		return resPath;
	}

	getFormatSteamDownloadFileURL(file: any): string {
		file.parentPath = file.parent_dir || file.parentPath;
		const store = useDataStore();
		const repo_id = getParams(file.path, 'id');
		return `${store.baseURL()}/seahub/lib/${repo_id}/file${encodeUrl(
			file.parentPath || '/'
		)}${encodeUrl(file.name)}?dl=1`;
	}
	formatFolderSubItemDownloadPath(
		item: TransferItem,
		parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	) {
		let parentPath = parentItem.parentPath || '';
		parentPath =
			(parentPath.endsWith('/') ? parentPath : `${parentPath}/`) +
			parentItem.name +
			'/';
		let path = item.path;
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

			const pathFullpath = path.split('?')[0];

			const cur_file = this.formatTransferToFileItem(transferItem);

			const filesStore = useFilesStore();
			if (
				decodeURIComponent(cur_file.path).indexOf(
					decodeURIComponent(pathFullpath)
				) >= 0
			) {
				filesStore.setBrowserUrl(path, cur_file.driveType, false);
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
		console.log('query ===>', query);

		return {
			path,
			query
		};
	}

	getUploadNode(): string {
		return '';
	}

	getPanelJumpPath(file: any): string {
		return '';
	}

	formatSearchPath(search: string): string {
		const formatSearchPath = search.startsWith('/') ? search : '/' + search;
		return formatSearchPath;
	}

	getOriginalPath(file: FileItem) {
		return this.getAttrPath(file);
	}
}

/* eslint-disable @typescript-eslint/no-unused-vars */
import Origin from './../origin';
import { createURL, getPurePath } from '../utils';
import { formatDrive } from './filesFormat';
import { MenuItem } from 'src/utils/contact';
import { OPERATE_ACTION } from 'src/utils/contact';
import * as files from './utils';

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
import { filterPcvPath } from '../common/common';
import { getextension } from 'src/utils/utils';
import { encodeUrl, decodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import { getFileIcon } from '@bytetrade/core';
import { IUploadCloudParams } from 'src/platform/interface/electron/interface';
import { useUserStore } from 'src/stores/user';
import { Router } from 'vue-router';
import * as filesUtil from '../common/utils';
import { CommonUrlApiType, pasteMutiNodesDriveType } from '../common/utils';
import { appendPath } from '../path';
import { SyncRepoItemType } from '../sync/type';

export default class DriveDataAPI extends Origin {
	public commonAxios: any;

	public origin_id: number;

	public driveType: DriveType = DriveType.Drive;

	static SIZE: number;

	constructor(origin_id: number = FilesIdType.PAGEID) {
		super();
		this.origin_id = origin_id;
		this.commonAxios = CommonFetch;
	}

	breadcrumbsBase = '/Files';

	SIZE = 8 * 1024 * 1024;

	fileEditEnable = true;

	videoPlayEnable = true;

	audioPlayEnable = true;

	fileEditLimitSize = 1024 * 1024;

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = files.driveRemovePrefix(url);
		const res: FileResType = await this.fetchDrive(pureUrl);
		return res;
	}

	async fetchDrive(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(
			files.driveCommonUrl('resources', url),
			{}
		);

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

		for (let i = 0; i < selected.length; i++) {
			const index = selected[i];
			const selectFile = filesStore.getTargetFileItem(index, this.origin_id);

			if (!selectFile) {
				continue;
			}
			const url = filesUtil.getDownloadUrl(selectFile);

			const fileObj: TransferItem = {
				url,
				path: selectFile.path,
				parentPath: selectFile.oParentPath,
				size: selectFile.isDir ? 0 : selectFile.size,
				name: selectFile.name,
				type: selectFile.type,
				driveType: selectFile.driveType,
				front: TransferFront.download,
				status: TransferStatus.Prepare,
				uniqueIdentifier: selectFile.uniqueIdentifier,
				isPaused: false,
				isFolder: selectFile.isDir ? true : false,
				node: '',
				currentPhase: 1,
				totalPhase: 1
			};

			result.push(fileObj);
		}

		return result;
	}

	async copy(el: FileItem, type: string): Promise<CopyStoragesType> {
		const from = appendPath(
			'/',
			el.fileType || '',
			el.fileExtend,
			el.oPath || '',
			el.isDir ? '/' : ''
		);

		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: el.driveType,
			src_node: pasteMutiNodesDriveType.includes(el.driveType)
				? el.fileExtend
				: el.driveType == DriveType.Share && el.node
				? el.node
				: undefined,
			isDir: el.isDir
		};

		if (type === 'cut') {
			copyItem.key = 'x';
		}

		return copyItem;
	}

	utilsFormatPathtoUrl(path: string) {
		return files.formatPathtoUrl(path);
	}

	async paste(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<any> {
		const operateinStore = useOperateinStore();
		const items: CopyStoragesType[] = [];
		if (!operateinStore.copyFiles || operateinStore.copyFiles.length == 0) {
			return;
		}

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element = operateinStore.copyFiles[i];
			const parentPath = this.utilsFormatPathtoUrl(path);
			const to = appendPath(
				'/',
				parentPath,
				encodeUrl(element.name),
				element.isDir ? '/' : ''
			);
			const filesStore = useFilesStore();

			let dst_node: string | undefined = undefined;
			if (
				pasteMutiNodesDriveType.includes(this.driveType) &&
				filesStore.currentFileList[this.origin_id]
			) {
				dst_node = filesStore.currentFileList[this.origin_id]?.fileExtend;
			} else if (
				this.driveType == DriveType.Share &&
				filesStore.currentFileList[this.origin_id]?.node
			) {
				dst_node = filesStore.currentFileList[this.origin_id]?.node;
			}

			items.push({
				...element,
				dst_node: dst_node,
				to: to,
				dst_drive_type: this.driveType
			});
		}

		let isMove = false;
		if (operateinStore.copyFiles[0].key === 'x') {
			isMove = true;
		}
		return await filesUtil.action(items, path, isMove, callback);
	}

	async move(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<any> {
		const filesStore = useFilesStore();
		const items: CopyStoragesType[] = [];
		for (const i of filesStore.selected[this.origin_id]) {
			const element = filesStore.getTargetFileItem(i, this.origin_id);
			if (!element) {
				continue;
			}
			const from = appendPath(
				'/',
				element.fileType || '',
				element.fileExtend,
				element.oPath || '',
				element.isDir ? '/' : ''
			);

			const parentPath = this.utilsFormatPathtoUrl(path);

			const to = appendPath(
				'/',
				parentPath,
				element.name,
				element.isDir ? '/' : ''
			);

			items.push({
				from: from,
				to: to,
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: element.driveType
			});
		}
		return await filesUtil.action(items, path, true, callback);
	}

	async action(
		_overwrite: boolean | undefined,
		_rename: boolean | undefined,
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
		return item;
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return this.utilsFormatPathtoUrl(file.path);
		}

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};
		const file_path = files.driveRemovePrefix(decodeUrl(file.path));

		const url = createURL(files.driveCommonUrl(`preview`, file_path), params);
		return url;
	}

	getDownloadURL(
		file: FileItem,
		inline: boolean,
		_download?: boolean,
		decodeUrl = true
	): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		let file_path = files.driveRemovePrefix(file.path);

		if (decodeUrl) {
			try {
				file_path = decodeURIComponent(file_path);
			} catch (error) {
				console.log(file_path);
			}
		}

		const url = createURL(files.driveCommonUrl('raw', file_path), params);
		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		if (!['text', 'txt', 'textImmutable'].includes(file.type)) {
			return file;
		}
		try {
			const newPath = file.path;

			const url = files.driveRemovePrefix(newPath);
			const res = await this.commonAxios.get(files.driveCommonUrl('raw', url), {
				params: {
					inline: true
				}
			});
			file.content = res;
		} catch (error) {
			console.error(error.message);
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
		return this.utilsFormatPathtoUrl(path);
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		await filesUtil.batchDeleteFileItems(items);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		await filesUtil.renameFileItem(item, newName);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const filesStore = useFilesStore();
		const newName = checkSameName(
			dirName,
			filesStore.currentFileList[this.origin_id]?.items
		);
		let url = path + '/' + encodeURIComponent(newName) + '/';
		url = url.replace('//', '/');
		await files.createDir(url);
	}

	getAttrPath(item: FileItem): string {
		const path = decodeURIComponent(item.path);
		if (item.name) return path.slice(0, path.indexOf(item.name));

		return path;
	}

	async getFileServerUploadLink(folderPath: string): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const node = this.getUploadNode();
		const path = folderPath;

		const url =
			baseURL +
			`/upload/upload-link/${node}/?file_path=` +
			encodeUrl(path) +
			'&from=web';
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
		const node = this.getUploadNode();
		const url = baseURL + '/upload/file-uploaded-bytes/' + node + '/';
		const params = {
			parent_dir: filePath,
			file_name: fileName
		};
		const res = this.commonAxios.get(url, { params: params });
		return res;
	}

	async getCurrentRepoInfo(path: string): Promise<any> {
		const res = await this.fetch(this.formatPathtoUrl(path));
		return res;
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		files.saveFile(item.path, content);
	}

	getResumePath(path: string, relativePath: string) {
		return appendPath(path, relativePath);
	}

	formatTransferToFileItem(item: TransferItem): FileItem {
		const extension = getextension(item.name);

		if (item.isFolder && !item.path.endsWith('/')) {
			item.path = item.path + '/';
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
		const curPath = appendPath(path, '/');
		return curPath;
	}

	getDiskPath(selectFiles: any, type: CommonUrlApiType) {
		let currentPath = selectFiles.path;
		if (currentPath.indexOf('/Files') > -1) {
			currentPath = files.driveRemovePrefix(currentPath);
		} else {
			currentPath = files.driveRemoveHomePrefix(currentPath);
		}

		return files.driveCommonUrl(type, currentPath);
	}

	getFormatSteamDownloadFileURL(
		file: { fileType: string; fileExtend: string; path: string },
		inline?: boolean
	): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const path = appendPath(
			filesUtil.commonUrlTypeExtend('raw', file.fileType, file.fileExtend),
			encodeUrl(file.path)
		);

		const url = createURL(path, params);
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
		let path = item.path;
		if (path.startsWith('/Files')) {
			path = path.substring('/Files'.length);
		}
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
		file.driveType = DriveType.Drive;
		file.parentPath = parentPath;
		file.url = this.getFormatSteamDownloadFileURL(file);
		file.path = appendPath('/Files', file.fileExtend, encodeUrl(file.path));
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
		const path = this.utilsFormatPathtoUrl(transferItem.path);
		return await filesUtil.postCreateFile(path, false, undefined);
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

			let fullPath = url.getWindowFullpath();
			if (this.origin_id) {
				fullPath = filesStore.currentPath[this.origin_id].path;
			}

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
				filesStore.setBrowserUrl(
					fullPath,
					cur_file.driveType,
					false,
					this.origin_id
				);
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

	formatCopyPath(
		path: string,
		destname: string,
		isDir: boolean,
		node: string
	): string {
		if (path.startsWith(appendPath('/', node))) {
			path = path.substring(appendPath('/', node).length);
		}
		return appendPath('/Files', getPurePath(path), destname, isDir ? '/' : '');
	}

	getPanelJumpPath(file: any): string {
		return file.path;
	}

	formatSearchPath(search: string): string {
		search = search.substring(search.startsWith('/') ? 6 : 5);
		const path = filterPcvPath(search);
		const formatSearchPath = appendPath('/Files', path);
		return formatSearchPath;
	}

	async renameRepo(_item: SyncRepoItemType, _newName: string): Promise<void> {}

	getUploadNode(): string {
		const filesStore = useFilesStore();
		return filesStore.nodes.length > 0 ? filesStore.nodes[0].name : '';
	}

	utilsDisplayPath(file: {
		isDir: boolean;
		fileExtend: string;
		path: string;
		fileType: string;
	}) {
		return files.displayPath(file);
	}

	displayPath(file: {
		isDir: boolean;
		fileExtend: string;
		path: string;
		fileType: string;
	}) {
		return this.utilsDisplayPath(file);
	}

	getOriginalPath(file: FileItem) {
		return this.getAttrPath(file);
	}
	pathToFrontendFile(path: string) {
		return {
			isDir: path.endsWith('/'),
			fileExtend: 'Home',
			path: files.driveRemovePrefix(path)
		};
	}
}

import { DriveDataAPI } from './../index';
import { getFileIcon } from '@bytetrade/core';
import { formatGd } from './filesFormat';
import { createURL, getPurePath } from '../utils';
import url from 'src/utils/url';
import {
	FileItem,
	FileResType,
	FilesIdType,
	useFilesStore
} from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { OPERATE_ACTION } from 'src/utils/contact';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { useTransfer2Store } from 'src/stores/transfer2';
import { encodeUrl } from 'src/utils/encode';
import { useDataStore } from 'src/stores/data';
import { getextension } from 'src/utils/utils';
import md5 from 'js-md5';
import * as files from './utils';
import * as filesUtil from '../common/utils';
import { appendPath } from '../path';

export default class DropboxDataAPI extends DriveDataAPI {
	breadcrumbsBase = '/Drive/dropbox';

	public origin_id: number;

	public driveType: DriveType = DriveType.Dropbox;

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
		const requestUrl = files.dropboxRemovePrefix(url);
		const res = await this.commonAxios.get(
			files.dropboxCommonUrl('resources', requestUrl),
			{}
		);

		const data: FileResType = formatGd(
			JSON.parse(JSON.stringify(res)),
			this.origin_id
		);

		return data;
	}

	async formatRepotoPath(item: any): Promise<string> {
		return `/Drive/${DriveType.Dropbox}/${item.key}/`;
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
					url = files.download('zip', selectFile);
				} else {
					url = filesUtil.getStreamListUrl(selectFile);
				}
			} else {
				url = filesUtil.getDownloadUrl(selectFile);
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
		// let path = el.path;
		// try {
		// 	path = decodeURIComponent(el.path);
		// } catch (error) {
		// 	/* empty */
		// }
		// const from = files.formatPathtoUrl(path);
		const from = appendPath(
			'/',
			el.fileType || '',
			el.fileExtend,
			el.oPath || '',
			el.isDir ? '/' : ''
		);
		console.log('from ===>', from);

		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: DriveType.Dropbox,
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

		for (let i = 0; i < operateinStore.copyFiles.length; i++) {
			const element: any = operateinStore.copyFiles[i];
			let lastPathIndex =
				path.indexOf('?') > -1
					? path.slice(6, path.indexOf('?'))
					: path.slice(6);

			lastPathIndex = lastPathIndex.endsWith('/')
				? lastPathIndex
				: lastPathIndex + '/';

			let name = element.name;
			try {
				name = decodeURIComponent(element.name);
			} catch (error) {
				/* empty */
			}
			const to = files.formatPathtoUrl(
				appendPath(path, name, element.isDir ? '/' : '')
			);

			items.push({
				...element,
				from: element.from,
				to: to,
				name: element.name,
				src_drive_type: element.src_drive_type,
				dst_drive_type: DriveType.Dropbox
			});

			if (path + element.name === element.from) {
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
			const element = filesStore.getTargetFileItem(i, this.origin_id);
			if (!element) {
				continue;
			}
			let name = element.name;
			try {
				name = decodeURIComponent(element.name);
			} catch (error) {
				/* empty */
			}
			const to = appendPath(path, name);
			items.push({
				from: files.formatPathtoUrl(element.path),
				to: files.formatPathtoUrl(to),
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: DriveType.Dropbox
			});
		}
		const overwrite = true;
		return await this.action(overwrite, true, items, path, true, callback);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const url = appendPath(path, encodeURIComponent(dirName), '/');
		await files.createDir(url);
	}

	async openPreview(item: any): Promise<FileResType> {
		return item;
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return this.utilsFormatPathtoUrl(file.path);
		}

		let pathSplit = file.path.split('?')[0];
		try {
			pathSplit = decodeURIComponent(pathSplit);
		} catch (error) {
			/* empty */
		}

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};
		const path = files.dropboxRemovePrefix(pathSplit);
		return createURL(files.dropboxCommonUrl('preview', path), params);
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};
		let path = files.dropboxRemovePrefix(file.path);
		try {
			path = decodeURIComponent(path);
		} catch (error) {
			/* empty */
		}
		const url = createURL(files.dropboxCommonUrl('raw', path), params);
		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		if (!['text', 'txt', 'textImmutable', 'pdf'].includes(file.type)) {
			return file;
		}
		try {
			const newPath = file.path;
			const url = files.dropboxRemovePrefix(newPath);
			const res = await this.commonAxios.get(
				files.dropboxCommonUrl('raw', url),
				{
					params: {
						inline: true
					}
				}
			);
			file.content = res;

			// newFile.content = res;
		} catch (error) {
			console.error(error.message);
		}

		return file;
	}

	utilsFormatPathtoUrl(path: string) {
		return files.formatPathtoUrl(path);
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const path = '/' + file.path.split('/').slice(3).join('/');
		const name = file.path.split('/')[2];
		if (getFileIcon(file.name) === 'image') {
			const url = files.generateDownloadUrl(file.driveType, path, name);
			return url;
		}
		const params = {
			...(inline && { inline: 'true' }),
			src: DriveType.Dropbox
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
			param: '',
			fileExtend: 'dropbox'
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

		file.driveType = DriveType.Dropbox;
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

	fileEditEnable = false;

	videoPlayEnable = true;

	audioPlayEnable = true;

	formatCopyPath(path: string, destname: string, isDir: boolean): string {
		return appendPath('/Drive', getPurePath(path), destname, isDir ? '/' : '');
	}
}

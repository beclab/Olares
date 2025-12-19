import { DriveDataAPI } from './../index';
import { getFileIcon } from '@bytetrade/core';
import { format } from './filesFormat';
import { DriveMenuType, GoogleDriveFileItem } from './type';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { OPERATE_ACTION } from 'src/utils/contact';
import * as files from './utils';
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
import { createURL, getPurePath } from '../utils';
import { useTransfer2Store } from 'src/stores/transfer2';
import { encodeUrl, decodeUrl } from 'src/utils/encode';
import { getextension } from 'src/utils/utils';
import md5 from 'js-md5';

import * as filesUtil from '../common/utils';
import { appendPath } from '../path';

export default class GoogleDataAPI extends DriveDataAPI {
	breadcrumbsBase = '/Drive/google';

	public driveType: DriveType = DriveType.GoogleDrive;

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
		// const splitUrl_0 = url.split('?')[0];
		let cur_url = url;
		// if (this.isRoot(splitUrl_0)) {
		// 	cur_url = splitUrl_0 + 'root/';
		// }

		cur_url = files.googleRemovePrefix(cur_url);

		const res = await this.commonAxios.get(
			files.googleCommonUrl('resources', cur_url),
			{}
		);

		const data: FileResType = format(
			JSON.parse(JSON.stringify(res)),
			this.origin_id
		);

		files.saveGoogleDirInfo(data.items, this.origin_id);
		return data;
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		const res1: any = await files.fetchRepo();
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
					// url = files.download('zip', selectFile);
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
		const from = files.formatPathtoUrl(el.path);
		const copyItem: CopyStoragesType = {
			from: from,
			to: '',
			name: el.name,
			src_drive_type: DriveType.GoogleDrive,
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

			let to = appendPath(
				path,
				encodeUrl(element.name),
				element.isDir ? '/' : ''
			);

			console.log('tp ====>', to);

			to = files.formatPathtoUrl(to);

			items.push({
				...element,
				to: to,
				dst_drive_type: DriveType.GoogleDrive
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

			items.push({
				from: files.formatPathtoUrl(element.path),
				to: files.formatPathtoUrl(path),
				name: element.name,
				src_drive_type: element.driveType,
				dst_drive_type: DriveType.GoogleDrive
			});
		}
		const overwrite = true;
		return await this.action(overwrite, true, items, path, true, callback);
	}

	async action(
		_overwrite: boolean | undefined,
		_rename: boolean | undefined,
		items: CopyStoragesType[],
		path: string,
		isMove: boolean | undefined,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<any> {
		return await filesUtil.action(
			// overwrite,
			// rename,
			items,
			path,
			isMove,
			callback
		);
	}

	async createDir(dirName: string, path: string): Promise<void> {
		let cur_url = path;
		// if (this.isRoot(path)) {
		// 	cur_url = path + 'root/';
		// }
		cur_url = appendPath(cur_url, encodeURIComponent(dirName), '/');
		await files.createDir(cur_url);
	}

	async openPreview(item: any): Promise<FileResType> {
		return item;
	}

	getPreviewURL(file: GoogleDriveFileItem, thumb: 'big' | 'thumb'): string {
		console.log('file ===>', file);

		if (['video'].includes(file.type)) {
			// return this.utilsFormatPathtoUrl(file.path);
			return appendPath('/google', file.fileExtend!, file.google_file_id);
		}

		const pathSplit = file.path.split('?')[0];

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		// const path = files.googleRemovePrefix(pathSplit);
		let path = files.googleRemovePrefix(pathSplit);
		try {
			path = decodeURIComponent(path);
		} catch (error) {
			console.log(path);
		}
		return createURL(files.googleCommonUrl('preview', path), params);
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};
		let path = files.googleRemovePrefix(file.path);
		try {
			path = decodeURIComponent(path);
		} catch (error) {
			console.log(path);
		}
		const url = createURL(files.googleCommonUrl('raw', path), params);
		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		if (!['text', 'txt', 'textImmutable', 'pdf'].includes(file.type)) {
			return file;
		}
		try {
			const newPath = file.path;
			const url = files.googleRemovePrefix(newPath);

			const res = await this.commonAxios.get(
				files.googleCommonUrl('raw', url),
				{
					params: {
						inline: true
					}
				}
			);
			file.content = res;
		} catch (error) {
			console.error(error.message);
		}
		return file;
	}

	utilsFormatPathtoUrl(path: string) {
		return files.formatPathtoUrl(path);
	}

	getFormatSteamDownloadFileURL(file: any, inline?: boolean): string {
		const lastIndex = file.path.lastIndexOf('/');
		const secondLastIndex = file.path.lastIndexOf('/', lastIndex - 1);
		// const google_id = file.path.substring(secondLastIndex + 1, lastIndex);
		// const name = file.path.split('/')[2];
		// if (getFileIcon(file.name) === 'image') {
		// 	const url = files.generateDownloadUrl(file.driveType, google_id, name);
		// 	return url;
		// }

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
			fileExtend: 'google'
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

	videoPlayEnable = true;

	audioPlayEnable = true;

	formatCopyPath(path: string, destname: string, isDir: boolean): string {
		return appendPath('/Drive', getPurePath(path), destname, isDir ? '/' : '');
	}
}

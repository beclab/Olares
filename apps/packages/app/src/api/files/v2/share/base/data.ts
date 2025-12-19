import { dataAPIs, DriveDataAPI } from '../../index';
import { MenuItem } from 'src/utils/contact';
import { formatShareListData, formatDrive } from './filesFormat';
import {
	FileResType,
	FileItem,
	useFilesStore,
	ExoirationTime
} from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import { createURL, getPurePath } from '../../utils';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import * as files from './utils';
import { getFileIcon } from '@bytetrade/core';
import {
	driveTypeByFileTypeAndFileExtend,
	filterPcvPath
} from '../../common/common';
import share from '../../common/share';
import { appendPath } from '../../path';
import * as filesUtil from '../../common/utils';
import { checkSameName } from 'src/utils/file';
import { SharePermission, ShareType } from 'src/utils/interface/share';
import { useDataStore } from 'src/stores/data';

export default class ShareAPI extends DriveDataAPI {
	breadcrumbsBase = '';
	driveType: DriveType = DriveType.Share;

	fileEditEnable = false;

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchData(url);
		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		if (url == '/Share' || url == '/Share/') {
			const filesStore = useFilesStore();
			if (!filesStore.users) {
				await filesUtil.fetchUserList();
			}

			const res = await share.getShareList(this.getQueryParams());

			const data: FileResType = await formatShareListData(
				res,
				this.driveType,
				this.origin_id
			);

			return data;
		} else {
			const pureUrl = files.formatResourcesUrl(url);

			const res = await this.commonAxios.get(pureUrl, {});

			const data: FileResType = await formatDrive(res, this.origin_id);

			return data;
		}
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.SHARE}`),
				key: MenuItem.SHARE,
				icon: 'sym_r_folder_supervised',
				driveType: DriveType.Share
			}
			// {
			// 	label: i18n.global.t(`files_menu.${MenuItem.SHAREBYME}`),
			// 	key: MenuItem.SHAREBYME,
			// 	icon: '',
			// 	driveType: DriveType.ShareByMe,
			// 	img: './img/share_by_me.svg'
			// }
		];
	}

	async formatRepotoPath(): Promise<string> {
		return appendPath('/Share/');
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		files.put(item.path, content);
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return files.formatPathtoUrl(file.path);
		}

		const { path, path_id } = files.getShareDataPath(file.path);
		let curPath = path;
		try {
			curPath = decodeURIComponent(curPath);
		} catch (error) {
			console.log(path);
		}

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		const res = createURL(
			files.shareCommonUrl('preview', path_id, curPath),
			params
		);

		return res;
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		// let curPath = files.shareRemovePrefix(file.path);
		const { path, path_id } = files.getShareDataPath(file.path);
		let curPath = path;

		try {
			curPath = decodeURIComponent(curPath);
		} catch (error) {
			console.log(curPath);
		}
		const url = createURL(
			files.shareCommonUrl('raw', path_id, curPath),
			params
		);

		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const newFile = JSON.parse(JSON.stringify(file));
		if (!['text', 'txt', 'textImmutable'].includes(newFile.type)) {
			return newFile;
		}
		try {
			const { path, path_id } = files.getShareDataPath(file.path);
			const res = await this.commonAxios.get(
				files.shareCommonUrl('raw', path_id, path),
				{
					params: {
						inline: true
					}
				}
			);

			file.content = res;
			newFile.content = res;
		} catch (error) {
			console.error(error.message);
		}
		return newFile;
	}

	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.driveType = DriveType.Data;
		file.parentPath = parentPath;
		file.url = this.getFormatSteamDownloadFileURL(file);
		file.path = appendPath('/', file.fileExtend, encodeUrl(file.path));
	}

	formatCopyPath(path: string, destname: string, isDir: boolean): string {
		return appendPath(getPurePath(path), destname, isDir ? '/' : '');
	}

	formatSearchPath(search: string): string {
		search = search.substring(search.startsWith('/') ? 5 : 4);
		const path = filterPcvPath(search);
		const formatSearchPath = appendPath('/', path);
		return formatSearchPath;
	}

	utilsFormatPathtoUrl(path: string) {
		return files.formatPathtoUrl(path);
	}

	utilsDisplayPath(file: {
		isDir: boolean;
		fileExtend: string;
		path: string;
		fileType: string;
	}) {
		return files.displayPath(file);
	}

	async renameItem(item: FileItem, newName: string): Promise<void> {
		await filesUtil.renameFileItem(item, newName);
	}

	async deleteItem(items: FileItem[]): Promise<void> {
		if (items.length == 0) {
			return;
		}
		if (items[0].isShareItem) {
			await share.remove(items.map((it) => it.fileExtend).join(','));
		} else {
			await filesUtil.batchDeleteFileItems(items);
		}
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

	private getQueryParams() {
		const filesStore = useFilesStore();
		let expire_in = -1;
		let expire_over = -1;
		if (filesStore.shareFilter.expire == ExoirationTime.within1days) {
			expire_in = 24 * 3600 * 1000;
		} else if (filesStore.shareFilter.expire == ExoirationTime.within7days) {
			expire_in = 7 * 24 * 3600 * 1000;
		} else if (filesStore.shareFilter.expire == ExoirationTime.within30days) {
			expire_in = 30 * 24 * 3600 * 1000;
		} else if (filesStore.shareFilter.expire == ExoirationTime.within1year) {
			expire_in = 365 * 24 * 3600 * 1000;
		} else if (filesStore.shareFilter.expire == ExoirationTime.over1year) {
			expire_over = 365 * 24 * 3600 * 1000;
		}

		const scopes: ShareType[] = [];
		if (
			!filesStore.shareFilter.scope.public ||
			!filesStore.shareFilter.scope.smb ||
			!filesStore.shareFilter.scope.internal
		) {
			if (filesStore.shareFilter.scope.internal) {
				scopes.push(ShareType.INTERNAL);
			}
			if (filesStore.shareFilter.scope.smb) {
				scopes.push(ShareType.SMB);
			}
			if (filesStore.shareFilter.scope.public) {
				scopes.push(ShareType.PUBLIC);
			}
		}

		const share_type =
			scopes.length > 0
				? scopes.join(',')
				: filesStore.shareFilter.scope.public
				? undefined
				: "''";

		const permission: SharePermission[] = [];
		if (
			!filesStore.shareFilter.permission.manage ||
			!filesStore.shareFilter.permission.edit ||
			!filesStore.shareFilter.permission.view
		) {
			if (filesStore.shareFilter.permission.manage) {
				permission.push(SharePermission.ADMIN);
			}
			if (filesStore.shareFilter.permission.edit) {
				permission.push(SharePermission.Edit);
			}
			if (filesStore.shareFilter.permission.view) {
				permission.push(SharePermission.View);
			}
		}

		const permisson_scopes =
			permission.length > 0
				? permission.join(',')
				: filesStore.shareFilter.permission.manage
				? ''
				: `${SharePermission.EMPTY}`;

		return {
			shared_to_me: filesStore.shareFilter.shared.withMe,
			shared_by_me: filesStore.shareFilter.shared.byMe,
			expire_in: expire_in > 0 ? expire_in : undefined,
			expire_over: expire_over > 0 ? expire_over : undefined,
			share_type: share_type,
			owner:
				filesStore.shareFilter.owner.length == 0
					? "''"
					: filesStore.shareFilter.owner.length !=
					  filesStore.users?.users.length
					? filesStore.shareFilter.owner.join(',')
					: '',
			permission: permisson_scopes
		};
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

		return res + '?ret-json=1&share=1';
	}

	async formatUploaderPath(path: string): Promise<string> {
		return this.utilsFormatPathtoUrl(path);
	}

	getOriginalPath(file: FileItem) {
		if (!file.file_type || !file.extend) {
			return '';
		}
		const driveType = driveTypeByFileTypeAndFileExtend(
			file.file_type,
			file.extend
		);

		if (driveType == DriveType.Sync) {
			return (
				appendPath('/Seahub', file.sync_repo_name || '', file.oPath || '') +
				`?id=${file.extend}&type=mine&p=rw`
			);
		}

		return dataAPIs(driveType).displayPath({
			isDir: true,
			fileExtend: file.extend,
			path: file.oPath || '',
			fileType: file.file_type
		});
	}

	getUploadNode(): string {
		const filesStore = useFilesStore();
		if (filesStore.currentNode[this.origin_id]) {
			return filesStore.currentNode[this.origin_id].name;
		}
		return super.getUploadNode();
	}

	pathToFrontendFile(path: string) {
		const { path: newPath, path_id } = files.getShareDataPath(path);
		return {
			isDir: path.endsWith('/'),
			fileExtend: path_id,
			path: newPath
		};
	}
}

import { DriveDataAPI } from './../index';
import { MenuItem } from 'src/utils/contact';
import { formatData } from './filesFormat';
import { FileResType, FileItem } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import { createURL, decodeURIComponentSafe, getPurePath } from '../utils';
import { createDir } from './utils';
import { encodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import * as files from './utils';
import { getFileIcon, getFilePreviewType, PreviewType } from '@bytetrade/core';
import { filterPcvPath } from '../common/common';
import { CommonUrlApiType } from '../common/utils';
import { appendPath } from '../path';
import { TransferItem } from 'src/utils/interface/transfer';
import { Router } from 'vue-router';
import { isCommonEnable } from '../..';

export default class DataDataAPI extends DriveDataAPI {
	breadcrumbsBase = '';
	public driveType: DriveType = DriveType.Data;

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = files.dataRemovePrefix(url);
		const res: FileResType = await this.fetchData(pureUrl);
		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(
			files.dataCommonUrl('resources', url),
			{}
		);

		const data: FileResType = await formatData(
			JSON.parse(JSON.stringify(res)),
			url,
			this.origin_id
		);

		return data;
	}

	async fetchMenuRepo(origins?: DriveType[]): Promise<DriveMenuType[]> {
		let items = [
			{
				label: i18n.global.t(`files_menu.${MenuItem.DATA}`),
				key: MenuItem.DATA,
				icon: 'sym_r_database',
				driveType: DriveType.Data
			}
		];

		if (origins && !origins.includes(DriveType.Data)) {
			items = [];
		}

		if (!origins || origins.includes(DriveType.Cache)) {
			items.push({
				label: i18n.global.t(`files_menu.${MenuItem.CACHE}`),
				key: MenuItem.CACHE,
				icon: 'sym_r_analytics',
				driveType: DriveType.Cache
			});
		}

		if (isCommonEnable() && (!origins || origins.includes(DriveType.Common))) {
			items.push({
				label: i18n.global.t(`files_menu.${MenuItem.COMMON}`),
				key: MenuItem.COMMON,
				icon: 'sym_r_interests',
				driveType: DriveType.Common
			});
		}

		return items;
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const url =
			(path.endsWith('/') ? path : path + '/') +
			encodeURIComponent(dirName) +
			'/';
		await createDir(url);
	}

	async formatRepotoPath(item: any): Promise<string> {
		if (item.key === 'Data') return '/Data/';
		return '/Data/' + item.key;
	}

	getDiskPath(selectFiles: any, type: CommonUrlApiType) {
		const path = files.dataRemovePrefix(selectFiles.path);
		return files.dataCommonUrl(type, path);
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		await files.put(item.path, content);
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return files.formatPathtoUrl(file.path);
		}

		const curPath = decodeURIComponentSafe(files.dataRemovePrefix(file.path));

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		const res = createURL(files.dataCommonUrl('preview', curPath), params);

		return res;
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const curPath = decodeURIComponentSafe(files.dataRemovePrefix(file.path));

		const url = createURL(files.dataCommonUrl('raw', curPath), params);

		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const newFile = JSON.parse(JSON.stringify(file));
		if (getFilePreviewType(file.name) !== PreviewType.TXT) {
			return newFile;
		}
		try {
			const newPath = file.path;
			const url = files.dataRemovePrefix(newPath);
			const res = await this.commonAxios.get(files.dataCommonUrl('raw', url), {
				params: {
					inline: true
				}
			});

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

	pathToFrontendFile(path: string) {
		return {
			isDir: path.endsWith('/'),
			fileExtend: 'Data',
			path: files.dataRemovePrefix(path)
		};
	}

	async transferItemBackToFiles(item: TransferItem, router: Router) {
		const path = files.dataRemovePrefix(item.path);
		await router.push({
			path: appendPath('/Data', path)
		});
	}
}

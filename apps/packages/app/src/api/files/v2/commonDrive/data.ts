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

export default class CommonDataAPI extends DriveDataAPI {
	breadcrumbsBase = '';
	public driveType: DriveType = DriveType.Common;

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = files.commonRemovePrefix(url);
		const res: FileResType = await this.fetchData(pureUrl);
		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const res = await this.commonAxios.get(
			files.commonCommonUrl('resources', url),
			{}
		);

		const data: FileResType = await formatData(
			JSON.parse(JSON.stringify(res)),
			url,
			this.origin_id
		);

		return data;
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const url =
			(path.endsWith('/') ? path : path + '/') +
			encodeURIComponent(dirName) +
			'/';
		await createDir(url);
	}

	async formatRepotoPath(item: any): Promise<string> {
		if (item.key === 'Common') return '/Common/';
		return '/Common/' + item.key;
	}

	getDiskPath(selectFiles: any, type: CommonUrlApiType) {
		const path = files.commonRemovePrefix(selectFiles.path);
		return files.commonCommonUrl(type, path);
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		await files.put(item.path, content);
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return files.formatPathtoUrl(file.path);
		}

		const curPath = decodeURIComponentSafe(files.commonRemovePrefix(file.path));

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		const res = createURL(files.commonCommonUrl('preview', curPath), params);

		return res;
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};

		const curPath = decodeURIComponentSafe(files.commonRemovePrefix(file.path));

		const url = createURL(files.commonCommonUrl('raw', curPath), params);

		return url;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const newFile = JSON.parse(JSON.stringify(file));
		if (getFilePreviewType(file.name) !== PreviewType.TXT) {
			return newFile;
		}
		try {
			const newPath = file.path;
			const url = files.commonRemovePrefix(newPath);
			const res = await this.commonAxios.get(
				files.commonCommonUrl('raw', url),
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
		file.driveType = DriveType.Common;
		file.parentPath = parentPath;
		file.url = this.getFormatSteamDownloadFileURL(file);
		file.path = appendPath('/', file.fileExtend, encodeUrl(file.path));
	}

	formatCopyPath(path: string, destname: string, isDir: boolean): string {
		return appendPath(getPurePath(path), destname, isDir ? '/' : '');
	}

	formatSearchPath(search: string): string {
		search = search.substring(search.startsWith('/') ? 7 : 6);
		const path = filterPcvPath(search);
		const formatSearchPath = appendPath('/Common', path);
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
			fileExtend: 'Common',
			path: files.commonRemovePrefix(path)
		};
	}

	async transferItemBackToFiles(item: TransferItem, router: Router) {
		const path = files.commonRemovePrefix(item.path);
		await router.push({
			path: appendPath('/Common', path)
		});
	}
}

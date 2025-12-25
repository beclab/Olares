import { DriveDataAPI } from './../index';
import { checkSameName, getAppDataPath } from 'src/utils/file';
import { formatAppData } from './filesFormat';
import { MenuItem, OPERATE_ACTION } from 'src/utils/contact';
import { useFilesStore, FileItem, FileResType } from 'src/stores/files';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { DriveType } from 'src/utils/interface/files';
import { DriveMenuType } from './type';
import * as files from './utils';
import { TransferItem, TransferFront } from 'src/utils/interface/transfer';

import { createURL, getPurePath } from '../utils';
import { encodeUrl, decodeUrl } from 'src/utils/encode';
import { i18n } from 'src/boot/i18n';
import { getFileIcon } from '@bytetrade/core';
import { filterPcvPath } from '../common/common';
import { getextension } from 'src/utils/utils';
import { appendPath } from '../path';
import * as filesUtil from '../common/utils';
import { CommonUrlApiType, formatAppDataNode } from '../common/utils';

export default class CacheDataAPI extends DriveDataAPI {
	breadcrumbsBase = '';

	public driveType: DriveType = DriveType.Cache;

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchCache(url);

		return res;
	}

	async fetchCache(url: string): Promise<FileResType> {
		if (url == '/Cache/') {
			const filesStore = useFilesStore();
			if (filesStore.nodes.length == 0) {
				await filesUtil.fetchNodeList();
			}
			return formatAppDataNode(
				url,
				filesStore.onlyMasterNodes[this.origin_id]
					? filesStore.nodes.filter((e) => e.master == true)
					: filesStore.nodes,
				DriveType.Cache,
				'/Cache'
			);
		}

		const { path, node } = getAppDataPath(url);

		const res: any = await this.commonAxios.get(
			files.cacheCommonUrl('resources', path, node),
			{}
		);
		return formatAppData(
			node,
			JSON.parse(JSON.stringify(res)),
			url,
			this.origin_id
		);
	}

	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		const filesStore = useFilesStore();
		if (filesStore.nodes.length == 0) {
			await filesUtil.fetchNodeList();
		}
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.CACHE}`),
				key: MenuItem.CACHE,
				icon: 'sym_r_analytics',
				driveType: DriveType.Cache
			}
		];
	}

	async formatRepotoPath(item: any): Promise<string> {
		const filesStore = useFilesStore();
		if (filesStore.nodes.length == 0) {
			await filesUtil.fetchNodeList();
		}

		if (filesStore.onlyMasterNodes[this.origin_id]) {
			return appendPath('/Cache/', filesStore.masterNode, '/');
		}

		if (filesStore.nodes.length > 1) {
			return '/Cache/';
		}
		if (filesStore.nodes.length > 0) {
			filesStore.currentNode[this.origin_id] = filesStore.nodes[0];
		}

		return appendPath(
			'/Cache/',
			filesStore.nodes.length == 0 ? '' : filesStore.nodes[0].name,
			'/'
		);
	}

	async openPreview(item: any): Promise<FileResType> {
		return item;
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return files.formatPathtoUrl(file.path);
		}

		const { path, node } = getAppDataPath(file.path);

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		const previewPath = files.cacheCommonUrl(
			'preview',
			decodeURIComponent(path),
			node
		);
		return createURL(previewPath, params, false);
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const params = {
			...(inline && { inline: 'true' })
		};
		const { path, node } = getAppDataPath(file.path);

		let newPath = path;
		try {
			newPath = decodeURIComponent(path);
		} catch (error) {
			console.log(error);
		}

		const res = createURL(
			files.cacheCommonUrl('raw', newPath, node),
			params,
			false
		);

		return res;
	}

	async formatFileContent(file: FileItem): Promise<FileItem> {
		if (!['text', 'txt', 'textImmutable'].includes(file.type)) {
			return file;
		}
		try {
			let newPath = file.path;
			if ((file as any).front) {
				newPath = file.path + encodeURIComponent(file.name);
			}
			const url = files.cacheRemoveCachePrefix(newPath);

			const res = await this.commonAxios.get(files.cacheCommonUrl('raw', url), {
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

	async onSaveFile(item: any, content: any): Promise<void> {
		files.put(item.path, content);
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

	getDiskPath(selectFiles: any, type: CommonUrlApiType) {
		const { path, node } = getAppDataPath(selectFiles.path);
		return files.cacheCommonUrl(type, path, node);
	}

	formatTransferToFileItem(item: TransferItem): FileItem {
		const extension = getextension(item.name);
		if (item.path.endsWith(item.name)) {
			item.path =
				item.path.substring(0, item.path.length - item.name.length) +
				encodeUrl(item.name);
			if (item.isFolder && !item.path.endsWith('/')) {
				item.path = item.path + '/';
			}
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
			fileExtend: 'cache'
		};

		return res;
	}

	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.driveType = DriveType.Cache;
		file.parentPath = parentPath;
		file.url = this.getFormatSteamDownloadFileURL(file);
		file.path = appendPath('/Cache', file.fileExtend, encodeUrl(file.path));
	}

	formatCopyPath(
		path: string,
		destname: string,
		isDir: boolean,
		node: string
	): string {
		let newPath = appendPath(getPurePath(path), destname, isDir ? '/' : '');
		if (newPath.startsWith('/cache')) {
			newPath = appendPath('/Cache', newPath.substring(6));
		}
		return newPath;
	}

	formatSearchPath(search: string): string {
		search = search.substring(search.startsWith('/') ? 6 : 5);
		const path = filterPcvPath(search, 2);
		const formatSearchPath = appendPath('/Cache', path);
		return formatSearchPath;
	}

	getUploadNode(): string {
		const filesStore = useFilesStore();
		return filesStore.currentNode[this.origin_id].name;
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
		const { path: newPath, node } = getAppDataPath(path);
		return {
			isDir: path.endsWith('/'),
			fileExtend: node,
			path: newPath
		};
	}
}

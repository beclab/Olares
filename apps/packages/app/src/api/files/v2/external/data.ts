import { DriveDataAPI } from './../index';
import { MenuItem } from 'src/utils/contact';
import { DriveMenuType } from './type';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import * as files from './utils';
import { formatDrive } from './filesFormat';
import { createURL, getPurePath } from '../utils';
import { encodeUrl } from 'src/utils/encode';
import { checkSameName, getAppDataPath } from 'src/utils/file';
import * as filesUtil from '../common/utils';
import { CommonUrlApiType, formatAppDataNode } from '../common/utils';
import { TransferFront, TransferItem } from 'src/utils/interface/transfer';
import { appendPath } from '../path';
import { getFileIcon } from '@bytetrade/core';
import { getextension } from 'src/utils/utils';

export default class ExternalDataAPI extends DriveDataAPI {
	public driveType: DriveType = DriveType.External;

	async fetch(url: string): Promise<FileResType> {
		const pureUrl = files.externalRemovePrefix(url);
		const res: FileResType = await this.fetchDrive(pureUrl);
		return res;
	}

	async fetchDrive(url: string): Promise<FileResType> {
		if (url == '/') {
			const filesStore = useFilesStore();
			if (filesStore.nodes.length == 0) {
				await filesUtil.fetchNodeList();
			}
			return formatAppDataNode(
				url,
				filesStore.onlyMasterNodes[this.origin_id]
					? filesStore.nodes.filter((e) => e.master == true)
					: filesStore.nodes,
				DriveType.External,
				'/Files/External'
			);
		}

		const res = await this.commonAxios.get(
			files.externalCommonUrl('resources', url),
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
				label: i18n.global.t(`files_menu.${MenuItem.EXTERNAL}`),
				key: MenuItem.EXTERNAL,
				icon: 'sym_r_hard_drive',
				driveType: DriveType.External
			}
		];
	}

	async formatRepotoPath(item: any): Promise<string> {
		const filesStore = useFilesStore();

		if (filesStore.nodes.length == 0) {
			await filesUtil.fetchNodeList();
		}

		if (filesStore.onlyMasterNodes[this.origin_id]) {
			return appendPath('/Files/External/', filesStore.masterNode, '/');
		}

		if (filesStore.nodes.length > 1) {
			return '/Files/External/';
		}
		if (filesStore.nodes.length > 0) {
			filesStore.currentNode[this.origin_id] = filesStore.nodes[0];
		}

		return appendPath(
			'/Files/External/',
			filesStore.nodes.length == 0 ? '' : filesStore.nodes[0].name,
			'/'
		);
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

		let file_path = files.externalRemovePrefix(file.path);

		if (decodeUrl) {
			try {
				file_path = decodeURIComponent(file_path);
			} catch (error) {
				console.log(file_path);
			}
		}

		const url = createURL(files.externalCommonUrl('raw', file_path), params);
		return url;
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
			const url = files.externalRemovePrefix(newPath);
			const res = await this.commonAxios.get(
				files.externalCommonUrl('raw', url),
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

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		if (['video'].includes(file.type)) {
			return this.utilsFormatPathtoUrl(file.path);
		}

		const params = {
			inline: 'true',
			key: file.modified,
			size: thumb
		};

		let file_path = files.externalRemovePrefix(file.path);

		try {
			file_path = decodeURIComponent(file_path);
		} catch (error) {
			console.log(file_path);
		}
		const url = createURL(
			files.externalCommonUrl('preview', file_path),
			params
		);
		return url;
	}

	async createDir(dirName: string, path: string): Promise<void> {
		const filesStore = useFilesStore();
		const newName = await checkSameName(
			dirName,
			filesStore.currentFileList[this.origin_id]?.items
		);

		let url = path + '/' + encodeURIComponent(newName) + '/';
		url = url.replace('//', '/');

		await files.createDir(url);
	}

	getDiskPath(selectFiles: any, type: CommonUrlApiType) {
		let currentPath = selectFiles.path;
		if (currentPath.indexOf('/Files') > -1) {
			currentPath = files.externalRemovePrefix(currentPath);
		} else {
			currentPath = files.externalRemoveHomePrefix(currentPath);
		}
		return files.externalCommonUrl(type, currentPath);
	}

	async onSaveFile(item: any, content: any): Promise<void> {
		files.put(item.path, content);
	}

	formatSearchPath(search: string): string {
		let path = search;
		const searchSplt = search.split('/');
		if (search.startsWith('hdd') || search.startsWith('/hdd')) {
			if (searchSplt.length > 3) {
				const subStr = searchSplt.slice(0, 3).join('/');
				path = search.substring(subStr.length);
			}
		}
		const formatSearchPath = appendPath('/Files/External', path);
		return formatSearchPath;
	}

	formatSteamDownloadItem(file: any, _infoPath?: string, parentPath?: string) {
		const fileType = getFileIcon(file.name);
		file.type = fileType;
		file.driveType = DriveType.External;
		file.parentPath = parentPath;
		file.url = this.getFormatSteamDownloadFileURL(file);
		file.path = appendPath(
			'/Files/External',
			file.fileExtend,
			encodeUrl(file.path)
		);
	}

	getUploadNode(): string {
		const filesStore = useFilesStore();
		return filesStore.currentNode[this.origin_id].name;
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
		if (path.startsWith('/external')) {
			path = path.substring(9);
		}
		return appendPath(
			'/Files/External',
			getPurePath(path),
			destname,
			isDir ? '/' : ''
		);
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
			fileExtend: item.node ? item.node : ''
		};

		return res;
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

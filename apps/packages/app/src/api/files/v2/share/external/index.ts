import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import ShareAPI from '../base/data';
import * as files from '../base/utils';
import { formatDrive } from '../base/filesFormat';

import { DriveType } from 'src/utils/interface/files';
import { useDataStore } from 'src/stores/data';
import { decodeUrl, encodeUrl } from 'src/utils/encode';
import { useShareStore } from 'src/stores/share/share';
import { createURL } from '../../utils';
import { useTransfer2Store } from 'src/stores/transfer2';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import url from 'src/utils/url';
import { SharePermission } from 'src/utils/interface/share';
import * as filesUtil from '../../common/utils';
import { appendPath } from '../../path';

export default class PublicShareAPI extends ShareAPI {
	driveType: DriveType = DriveType.PublicShare;

	async fetch(url: string): Promise<FileResType> {
		const res: FileResType = await this.fetchData(url);
		return res;
	}

	async fetchData(url: string): Promise<FileResType> {
		const pureUrl = files.formatResourcesUrl(url);

		const res = await this.commonAxios.get(pureUrl, {});

		const data: FileResType = await formatDrive(
			res,
			this.origin_id,
			this.driveType
		);

		return data;
	}

	async getFileServerUploadLink(folderPath: string): Promise<any> {
		const dataStore = useDataStore();
		const baseURL = dataStore.baseURL();
		const node = this.getUploadNode();
		const path = folderPath;

		const url =
			appendPath(baseURL, '/upload/upload-link/', !!node ? `${node}/` : '') +
			`?file_path=${encodeUrl(path)}&from=web`;

		const res = await this.commonAxios.get(url, {
			responseType: 'text'
		});

		const shareStore = useShareStore();

		return res + `?ret-json=1&token=${shareStore.token}`;
	}

	getPreviewURL(file: FileItem, thumb: 'big' | 'thumb'): string {
		const shareStore = useShareStore();
		if (['video'].includes(file.type)) {
			return files.formatPathtoUrl(file.path) + '&token=' + shareStore.token;
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
			size: thumb,
			token: shareStore.token
		};

		const res = createURL(
			files.shareCommonUrl('preview', path_id, curPath),
			params
		);

		return res;
	}

	getDownloadURL(file: FileItem, inline: boolean): string {
		const shareStore = useShareStore();

		const params = {
			...(inline && { inline: 'true' }),
			token: shareStore.token
		};

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
			const shareStore = useShareStore();

			const url = filesUtil.getDownloadUrl(selectFile, {
				token: shareStore.token
			});

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

	async formatFileContent(file: FileItem): Promise<FileItem> {
		const newFile = JSON.parse(JSON.stringify(file));
		if (!['text', 'txt', 'textImmutable'].includes(newFile.type)) {
			return newFile;
		}
		try {
			const shareStore = useShareStore();
			const { path, path_id } = files.getShareDataPath(file.path);

			const res = await this.commonAxios.get(
				files.shareCommonUrl('raw', path_id, path),
				{
					params: {
						inline: true,
						token: shareStore.token
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

	async uploadSuccessRefreshData(tranfeItemId: number) {
		try {
			const shareStore = useShareStore();
			if (
				!shareStore.share ||
				shareStore.share.permission == SharePermission.UploadOnly
			) {
				return;
			}
			const transferStore = useTransfer2Store();
			const transferItem =
				transferStore.transferMap[tranfeItemId] ||
				transferStore.getSubTransferItem(TransferFront.upload, tranfeItemId);

			if (!transferItem) {
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

	getUploadNode(): string {
		const shareStore = useShareStore();
		console.log('shareStore.share ===>', shareStore.share);

		if (!shareStore.share || !shareStore.share.node) {
			return '';
		}
		return shareStore.share.node;
	}
}

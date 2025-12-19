import { FileItem, FileResType } from './../../../stores/files';
import { CopyStoragesType } from 'src/stores/operation';

import { OPERATE_ACTION } from '../../../utils/contact';

import { DriveMenuType } from './drive/type';
import { SyncRepoItemType, SyncRepoMineType } from './sync/type';
import { TransferItem } from 'src/utils/interface/transfer';
import { IUploadCloudParams } from 'src/platform/interface/electron/interface';
import { Router } from 'vue-router';

export default abstract class Origin {
	public origin_id: number;

	/**
	 * Retry timer when uplaod
	 */
	// public RETRY_TIMER = 3;

	/**
	 * Chunk size when uplaod
	 */
	// public SIZE = 8 * 1024 * 1024;

	/**
	 * Breadcrumbs base
	 */
	abstract breadcrumbsBase: string;

	/**
	 * This function retrieves the data from all files in the specified directory.
	 */
	abstract fetch(url: string): Promise<FileResType>;

	/**
	 * Retrieves this menu from the Sync
	 */
	abstract fetchMenuRepo(
		menu?: string
	): Promise<DriveMenuType[] | SyncRepoMineType[]>;

	/**
	 * get download files url info
	 */
	abstract getDownloadInfo(path: string): Promise<TransferItem[]>;

	/**
	 * download file
	 */
	abstract downloadFile(fileUrl: any, filename: string): Promise<void>;

	/**
	 * This function handles the copying of files or directories in an event
	 */
	abstract copy(el: FileItem, type?: string): Promise<CopyStoragesType>;

	/**
	 * Paste
	 */
	abstract paste(
		route: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void>;

	/**
	 * Move
	 */
	abstract move(
		path: string,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void>;

	/**
	 * Event middleware
	 */
	abstract action(
		overwrite: boolean | undefined,
		rename: boolean | undefined,
		items: CopyStoragesType[],
		path: string,
		isMove: boolean | undefined,
		callback: (action: OPERATE_ACTION, data: any) => Promise<void>
	): Promise<void>;

	/**
	 * Upload files
	 */
	abstract uploadFiles(): void;

	/**
	 * Upload folder
	 */
	abstract uploadFolder(): void;

	/**
	 * Open local folder
	 */
	abstract openLocalFolder(): string | undefined;

	/**
	 * Open preview when upload modal
	 */
	abstract openPreview(item: any): Promise<FileResType>;

	/**
	 * get url for preview page
	 */
	abstract getPreviewURL(res: any, thumb: string): string;

	/**
	 * get DownloadURL url for preview page
	 */
	abstract getDownloadURL(
		file: any,
		inline?: boolean,
		download?: boolean,
		decodeUrl?: boolean
	): string;

	/**
	 * Format File Content for 'audio', 'video', 'text', 'txt', 'textImmutable', 'pdf'
	 */
	abstract formatFileContent(file: FileItem): Promise<FileItem>;

	/**
	 * format Repo to Path
	 */
	abstract formatRepotoPath(item?: any): Promise<string>;

	/**
	 * format path to url
	 */
	abstract formatPathtoUrl(path: string, param?: string): string;

	/**
	 * Handle Delete Item
	 */
	abstract deleteItem(item: FileItem[]): Promise<void>;

	/**
	 * Handle Rename Item
	 */
	abstract renameItem(item: FileItem, newName: string): Promise<void>;

	/**
	 * Handle Create New Dir
	 */
	abstract createDir(dirName: string, path: string): Promise<void>;

	/**
	 * Get Attr Path
	 */
	abstract getAttrPath(item: FileItem): string;

	/**
	 * get file server upload link
	 */

	abstract getFileServerUploadLink(
		folderPath: string,
		repoID?: string,
		dirName?: string
	): Promise<any>;

	/**
	 * get file uploaded bytes
	 */
	abstract getFileUploadedBytes(
		filePath: string,
		fileName: string,
		repoID?: string,
		taskId?: string
	): Promise<any>;

	/**
	 * get current repo Info
	 */
	abstract getCurrentRepoInfo(path: string): Promise<any>;

	/**
	 * Save File
	 */
	abstract onSaveFile(
		url: string,
		content: any,
		isNative?: boolean
	): Promise<void>;

	/**
	 * Transfer download path
	 */
	// abstract getRelativePath(file: any, parentPath?: string);

	/**
	 * getResumePath
	 */
	abstract getResumePath(path: string, relativePath: string);

	/**
	 * formatTransferToFileItem
	 */
	abstract formatTransferToFileItem(item: TransferItem): FileItem;

	/**
	 * formatTransferPath
	 */
	abstract formatTransferPath(item: TransferItem): string;

	/**
	 * formatUploadTransferPath
	 */
	abstract formatUploadTransferPath(item: TransferItem): Promise<string>;

	/**
	 * getPurePath
	 */
	abstract getPurePath(path: string): string;

	/**
	 * getDiskPath
	 */
	abstract getDiskPath(selectFiles: any, type?: string): string;

	/**
	 *	get format steam file item url
	 * @param file
	 * @param inline
	 * @param download
	 * @param decodeUrl
	 */
	abstract getFormatSteamDownloadFileURL(file: any, inline?: boolean): string;

	/**
	 * format download item local save path
	 * @param item
	 * @param parentItem
	 * @param folderSavePath
	 * @param defaultDownloadPath
	 * @param appendPath
	 */
	abstract formatFolderSubItemDownloadPath(
		item: TransferItem,
		parentItem: TransferItem,
		folderSavePath: string,
		defaultDownloadPath: string,
		appendPath: string
	): {
		parentSavePath: string;
		itemSavePath: string;
	};

	abstract formatSteamDownloadItem(
		file: any,
		infoPath?: string,
		parentPath?: string
	): void;

	abstract getUploadTransferItemMoreInfo(
		item: TransferItem
	): IUploadCloudParams | undefined;

	abstract fileEditEnable: boolean;

	abstract fileEditLimitSize: number;

	abstract videoPlayEnable: boolean;

	abstract audioPlayEnable: boolean;

	abstract transferItemUploadSuccessResponse(
		tranfeItemId: number,
		response: any
	): Promise<boolean>;

	abstract uploadSuccessRefreshData(tranfeItemId: number): Promise<void>;

	abstract transferItemBackToFiles(
		item: TransferItem,
		router: Router
	): Promise<void>;

	abstract formatUploaderPath(path: string): Promise<string>;

	abstract renameRepo(item: SyncRepoItemType, newName: string): Promise<void>;

	abstract getUploadNode(): string;

	abstract getPanelJumpPath(file: any): string;

	abstract formatSearchPath(search: string): string;

	abstract getOriginalPath(file: FileItem): string;
}

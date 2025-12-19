import { FileInfo } from 'src/utils/rss-types';
import { ClientUploadFileType } from './upload';
import { DriveType } from 'src/utils/interface/files';
import { TransferItem } from 'src/utils/interface/transfer';
import { DownloadRecord } from 'src/utils/interface/rss';

export interface TransferActionInterface {
	start(item: TransferItem): Promise<boolean>;
	cancel(item: TransferItem): Promise<boolean>;
	pause(item: TransferItem): Promise<boolean>;
	resume(item: TransferItem): Promise<boolean>;
	complete(item: TransferItem): Promise<boolean>;
	getTransferInfo(
		item: TransferItem
	): Promise<{ id: number; bytes: number } | undefined>;

	// prepare?(item: TransferItem): Promise<{
	// 	child?: TransferItem[];
	// 	finished: boolean;
	// 	callback?: (addStatus: boolean, ids: number[]) => void;
	// }>;
	addSubtasksSuccess?(
		offset: number,
		taskId: number,
		subtaskIds: number[],
		identifys?: string[]
	): void;

	restartEnable(item: TransferItem): Promise<boolean>;
}

export interface TransferDownloadInterface extends TransferActionInterface {
	restartAutoResume: boolean;
}

export interface TransferUploadInterface extends TransferActionInterface {
	restartAutoResume: boolean;
}

export type TransferCopyInterface = TransferActionInterface;

export interface TransferCloudInterface extends TransferActionInterface {
	init(): void;
	queryUrl(url: string): Promise<CloudFileInfo | undefined>;
	downloadFile(
		name: string,
		download_url: string,
		larepassId: string,
		path: string,
		file_type: string
	): Promise<DownloadRecord | undefined>;
	removeTask(item: TransferItem, deleteFile: boolean): Promise<boolean>;
	cancelTask(item: TransferItem): Promise<boolean>;
	removeAllTask(): Promise<boolean>;
	cancelAllTask(): Promise<boolean>;
	getQueryId(): string;
	taskIdentify(taskId: string): string;
	getTaskIdByIdentify(uniqueIdentifier?: string): string;
	taskBaseIdentify(): string;
}

export interface TransferClientService {
	downloader: TransferDownloadInterface;
	uploader: TransferUploadInterface;
	clouder?: TransferCloudInterface;
	copier?: TransferCopyInterface;

	restartAutoResume: boolean;
	errorRetryNumber: number;
}

export interface TransferClient {
	client: TransferClientService;
	waitAddSubtasks: Record<
		number,
		{
			finished: boolean;
			subtasks: TransferItem[];
		}
	>;
	doAction(
		item: TransferItem,
		action: 'start' | 'cancel' | 'pause' | 'resume'
	): Promise<boolean>;
}

export enum ClouderTransferStatus {
	DOWNLOADING = 'downloading',
	PAUSE = 'pause',
	CANCEL = 'cancel',
	ERROR = 'error',
	COMPLETE = 'complete',
	WAITING = 'waiting',
	REMOVE = 'remove'
}

export interface ClouderWSModel {
	task_id: number;
	url: string;
	status: ClouderTransferStatus;
	percent: number | null; // 0 - 100 | null
	name: string;
	path: string;
	mimeType: string;
	size: number | null;
	downloaded_bytes: number;
	startTime: number;
}

export type CloudFileInfo = FileInfo;

export const getUploadType = (item: TransferItem): ClientUploadFileType => {
	if (item.driveType == DriveType.Sync) {
		return 'SEAFILE';
	} else if (item.driveType == DriveType.GoogleDrive) {
		return 'GOOGLE_DRIVE';
	} else if (item.driveType == DriveType.Dropbox) {
		return 'DROP_BOX';
	} else if (item.driveType == DriveType.Awss3) {
		return 'AWS_S3';
	} else if (item.driveType == DriveType.Share) {
		return 'SHARE';
	}
	return 'DRIVE';
};

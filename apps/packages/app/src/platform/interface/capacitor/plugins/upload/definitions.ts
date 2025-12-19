import { ClientUploadFileType } from 'src/services/abstractions/transfer/upload';
import {
	FileDownloadProgressStatus,
	ProgressListener
} from '../download/definitions';

import { PluginListenerHandle } from '@capacitor/core';
import { DriveType } from 'src/utils/interface/files';

export interface FileUploadOptions {
	id: number;
	baseUrl: string;
	uploadPath: string;
	filePath: string;
	fileType: ClientUploadFileType;

	uniqueIdentifier?: string;

	// SEAFILE
	repoId?: string;

	// CLOUD
	account?: string;

	node?: string;
}

export interface FileUploadResult {
	code: number; //0 成功 1失败
	message: string;
	path: string;
	id: number;
	taskId?: string;
}

export interface FileUploadAddTasks {
	files: {
		path: string;
		name: string;
		size: number;
		mimeType: string;
		target: {
			driveType: DriveType;
			path: string;
			params: any;
		};
	}[];
}

export interface FileCloudUpdateTaskId {
	id: number;
	taskId: string;
}

export interface FileUploadPluginInterface {
	start(options: FileUploadOptions): Promise<{
		status: boolean;
	}>;

	pause(options: { id: number }): Promise<{
		status: boolean;
	}>;

	resume(options: { id: number; baseUrl?: string }): Promise<{
		status: boolean;
	}>;

	cancel(options: { id: number }): Promise<{
		status: boolean;
	}>;

	addListener(
		eventName: 'progress',
		listenerFunc: ProgressListener
	): Promise<PluginListenerHandle>;

	addListener(
		eventName: 'uploadResult',
		listenerFunc: (result: FileUploadResult) => void
	): Promise<PluginListenerHandle>;

	addListener(
		eventName: 'addUploadTask',
		listenerFunc: (result: FileUploadAddTasks) => void
	): Promise<PluginListenerHandle>;

	addListener(
		eventName: 'cloudUploadTaskId',
		listenerFunc: (result: FileCloudUpdateTaskId) => void
	): Promise<PluginListenerHandle>;

	// isImage only iOS
	selectFiles(options: {
		isImage?: boolean;
		target: {
			driveType: DriveType;
			path: string;
			params: any;
		};
	}): Promise<void>;

	getTransferInfo(options: {
		id: number;
	}): Promise<FileDownloadProgressStatus | undefined>;

	clearData(options: { savePath: string }): Promise<void>;
}

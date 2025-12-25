import { PluginListenerHandle, registerPlugin } from '@capacitor/core';
import { DownloadFileOptions, ProgressStatus } from '@capacitor/filesystem';

export interface FileDownloadOptions extends DownloadFileOptions {
	id: number;
	fileSize?: number;
}
export interface FileDownloadProgressStatus extends ProgressStatus {
	id: number;
}

export interface FileDownloadMoreResult {
	code: number; //0 成功 1失败
	message: string;
	path: string;
	id: number;
}

export type ProgressListener = (progress: FileDownloadProgressStatus) => void;

export interface FileDownloadPluginInterface {
	start(options: FileDownloadOptions): Promise<{
		status: boolean;
	}>;

	pause(options: { id: number }): Promise<{
		status: boolean;
	}>;

	resume(options: { id: number; url?: string }): Promise<{
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
		eventName: 'downloadResult',
		listenerFunc: (result: FileDownloadMoreResult) => void
	): Promise<PluginListenerHandle>;

	openFile(options: { path: string }): Promise<{
		status: boolean;
	}>;

	getTransferInfo(options: {
		id: number;
	}): Promise<FileDownloadProgressStatus | undefined>;
}

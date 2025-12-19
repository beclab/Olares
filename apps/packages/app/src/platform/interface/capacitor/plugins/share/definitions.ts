import { PluginListenerHandle } from '@capacitor/core';
import { DriveType } from 'src/utils/interface/files';
import { FileUploadAddTasks } from '../upload/definitions';

export declare type onFileUpload = (
	data: {
		base64: string;
		fileName: string;
		mimeType: string;
	}[]
) => void;

export declare type onShareUrl = (data: { url: string }) => void;
export interface FileSharePlugin {
	addListener(
		eventName: 'onFileUpload',
		listenerFunc: onFileUpload
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	addListener(
		eventName: 'onShareUrl',
		listenerFunc: onShareUrl
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	addListener(
		eventName: 'addUploadTask',
		listenerFunc: (result: FileUploadAddTasks) => void
	): Promise<PluginListenerHandle>;

	sharedFiles(options: {
		target: {
			driveType: DriveType;
			path: string;
			params: any;
		};
	}): Promise<void>;

	reset(options?: { clear?: boolean }): Promise<{
		status: boolean;
	}>;
}

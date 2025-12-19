import { registerPlugin } from '@capacitor/core';
import { FileDownloadPluginInterface } from './definitions';

const fileDownload =
	registerPlugin<FileDownloadPluginInterface>('FileDownloadPlugin');

export { fileDownload as FileDownloadPlugin };

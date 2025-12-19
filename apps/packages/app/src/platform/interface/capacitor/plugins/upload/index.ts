import { registerPlugin } from '@capacitor/core';
import { FileUploadPluginInterface } from './definitions';

const fileUpload =
	registerPlugin<FileUploadPluginInterface>('FileUploadPlugin');

export { fileUpload as FileUploadPlugin };

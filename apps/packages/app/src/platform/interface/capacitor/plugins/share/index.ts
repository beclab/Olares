import { registerPlugin } from '@capacitor/core';
import { FileSharePlugin } from './definitions';

const FileSharedService = registerPlugin<FileSharePlugin>('FileSharePlugin');

export { FileSharedService };

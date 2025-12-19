import { registerPlugin } from '@capacitor/core';
import { ScanPhotoQRPlugin } from './definitions';

const ScanPhotoQR = registerPlugin<ScanPhotoQRPlugin>('ScanPhotoQR');

export { ScanPhotoQR };

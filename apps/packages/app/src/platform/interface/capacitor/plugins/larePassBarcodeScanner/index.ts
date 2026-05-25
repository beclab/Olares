import { registerPlugin } from '@capacitor/core';

export interface LarePassScanOptions {
	title?: string;
	showAlbum?: boolean;
	albumTitle?: string;
}

export type LarePassScanAction = 'scan' | 'album';

export interface LarePassScanResult {
	action: LarePassScanAction;
	value?: string;
}

export interface LarePassDecodeImageOptions {
	content: string;
}

export interface LarePassDecodeImageResult {
	result: string[];
}

export interface LarePassBarcodeScannerPlugin {
	scan(options?: LarePassScanOptions): Promise<LarePassScanResult>;
	decodeImage(
		options: LarePassDecodeImageOptions
	): Promise<LarePassDecodeImageResult>;
	openSettings(): Promise<void>;
}

export const LarePassBarcodeScanner =
	registerPlugin<LarePassBarcodeScannerPlugin>('LarePassBarcodeScanner');

export type LarePassScanErrorCode =
	| 'USER_CANCELLED'
	| 'CAMERA_PERMISSION_DENIED'
	| 'SCAN_ERROR'
	| 'DECODE_ERROR';

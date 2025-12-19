export interface ScanPhotoQRPlugin {
	scan(options: { content: string }): Promise<{ result: string[] }>;
}

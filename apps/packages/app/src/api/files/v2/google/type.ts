import { FileItem } from 'src/stores/files';

export type { DriveMenuType } from '../drive/type';

export class DriveFileItem extends FileItem {
	iconLink?: string;
	webContentLink?: string;
	webViewLink?: string;
	thumbnailLink?: string;
}

export class GoogleDriveFileItem extends DriveFileItem {
	id_path: string;
	// o_path: string;
	google_file_id: string;
}

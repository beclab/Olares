import { FileItem } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';

export interface DriveMenuType {
	label: string;
	key: string | number;
	icon: string;
	driveType: DriveType;
}

export class DriveFileItem extends FileItem {
	iconLink?: string;
	webContentLink?: string;
	webViewLink?: string;
	thumbnailLink?: string;
}

export class GoogleDriveFileItem extends DriveFileItem {
	id_path: string;
	// o_path: string;
}

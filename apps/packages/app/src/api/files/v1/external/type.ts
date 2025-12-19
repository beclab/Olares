import { DriveType } from 'src/utils/interface/files';
export interface DriveMenuType {
	label: string;
	key: string | number;
	icon: string;
	driveType: DriveType;
}

export interface SmbMountType {
	password?: string;
	timestamp?: number;
	url: string;
	username?: string;
}

import { DriveType } from 'src/utils/interface/files';
export interface DriveMenuType {
	label: string;
	key: string | number;
	icon: string;
	driveType: DriveType;
}

export enum DriveType {
	Drive = 'drive',
	Sync = 'sync',
	Data = 'data',
	Cache = 'cache',
	GoogleDrive = 'google',
	Dropbox = 'dropbox',
	Awss3 = 'awss3',
	Tencent = 'tencent',
	External = 'external',
	Share = 'share',
	PublicShare = 'publicShare'
	// ShareWithMe = 'shareWithMe',
	// ShareByMe = 'shareByMe'
}

export const supportDriveTypes = [
	DriveType.Drive,
	DriveType.Sync,
	DriveType.Data,
	DriveType.Cache,
	DriveType.GoogleDrive,
	DriveType.Dropbox,
	DriveType.Awss3,
	DriveType.Tencent,
	DriveType.Share,
	DriveType.External
];

export interface ActiveMenuType {
	label: string;
	id: string;
	driveType: DriveType;
	params?: string;
}

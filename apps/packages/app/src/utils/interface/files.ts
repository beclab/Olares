import { FilesSortType } from '../contact';

export enum DriveType {
	Drive = 'drive',
	Sync = 'sync',
	Data = 'data',
	Common = 'common',
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
	DriveType.Common,
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

export const filesSortOptions = [
	{
		name: 'name',
		icon: 'sym_r_grid_view',
		action: 'name',
		type: FilesSortType.NAME
	},
	{
		name: 'type',
		icon: 'sym_r_edit_calendar',
		action: 'type',
		type: FilesSortType.TYPE
	},
	{
		name: 'modified',
		icon: 'sym_r_edit_document',
		action: 'modified',
		type: FilesSortType.Modified
	},
	{
		name: 'size',
		icon: 'sym_r_folder_copy',
		action: 'size',
		type: FilesSortType.SIZE
	}
];

export type FileViewMode = 'list' | 'mosaic';

// Reserved POSIX directory names that must not be used as file/folder/library
// names. Slash/backslash characters are handled separately by the slash check
// in useFileNameValidation, so they are intentionally not included here.
export const disabledFolderNames = ['.', '..'];

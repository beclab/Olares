import { DriveType } from './files';
export { getArchiveFileExtension, isArchiveFile } from '@bytetrade/core';
import { getArchiveFileExtension } from '@bytetrade/core';

export enum ArchiveFormat {
	Zip = 'zip',
	SevenZ = '7z',
	Tar = 'tar',
	TarGz = 'tar.gz',
	TGZ = 'tgz',
	TarBz2 = 'tar.bz2',
	TarXz = 'tar.xz',
	GZIP = 'gzip',
	BZIP2 = 'bzip2',
	XZ = 'xz'
}

export enum ArchiveConflict {
	Rename = 'rename',
	Overwrite = 'overwrite',
	Skip = 'skip'
}

export const archiveFormatOptions = [
	{
		label: 'files.archive.zip_file',
		value: ArchiveFormat.Zip,
		icon: 'sym_r_folder_zip'
	},
	{
		label: 'files.archive.seven_z_file',
		value: ArchiveFormat.SevenZ,
		icon: 'sym_r_folder_zip'
	},
	{
		label: 'files.archive.tar_file',
		value: ArchiveFormat.Tar,
		icon: 'sym_r_folder_zip'
	},
	{
		label: 'files.archive.more_options',
		icon: 'sym_r_settings'
	}
];

export const archiveAdvancedFormatOptions = [
	{
		label: 'ZIP',
		value: ArchiveFormat.Zip
	},
	{
		label: '7z',
		value: ArchiveFormat.SevenZ
	},
	{
		label: 'TAR',
		value: ArchiveFormat.Tar
	},
	{
		label: 'tar.gz',
		value: ArchiveFormat.TarGz
	},
	{
		label: 'tgz',
		value: ArchiveFormat.TGZ
	},
	{
		label: 'tar.bz2',
		value: ArchiveFormat.TarBz2
	},
	{
		label: 'tar.xz',
		value: ArchiveFormat.TarXz
	},
	{
		label: 'gzip',
		value: ArchiveFormat.GZIP
	},
	{
		label: 'bzip2',
		value: ArchiveFormat.BZIP2
	},
	{
		label: 'xz',
		value: ArchiveFormat.XZ
	}
];

export const archiveConflictOptions = [
	{
		label: 'files.archive.conflict_rename',
		value: ArchiveConflict.Rename
	},
	{
		label: 'files.archive.conflict_overwrite',
		value: ArchiveConflict.Overwrite
	},
	{
		label: 'files.archive.conflict_skip',
		value: ArchiveConflict.Skip
	}
];

export const archiveExtractOptions = [
	{
		label: 'files.archive.extract_current_folder',
		value: 'current',
		icon: 'sym_r_folder'
	},
	// Hidden for now.
	// {
	// 	label: 'files.archive.extract_new_folder',
	// 	value: 'new_folder',
	// 	icon: 'sym_r_create_new_folder'
	// },
	{
		label: 'files.archive.extract_choose_location',
		value: 'choose_location',
		icon: 'sym_r_drive_folder_upload'
	}
];

export const archiveSupportedDriveTypes = [
	DriveType.Drive,
	DriveType.Data,
	DriveType.Common,
	DriveType.Cache,
	DriveType.External
];

export const isArchiveDriveTypeSupported = (driveType?: DriveType) =>
	!!driveType && archiveSupportedDriveTypes.includes(driveType);

const singleFileCompressionFormats = [
	ArchiveFormat.GZIP,
	ArchiveFormat.BZIP2,
	ArchiveFormat.XZ
];

export const isSingleFileCompressionFormat = (format: ArchiveFormat) =>
	singleFileCompressionFormats.includes(format);

export const canCompressItemsWithFormat = (
	items: ArchiveSourceItem[],
	format: ArchiveFormat
) =>
	!isSingleFileCompressionFormat(format) ||
	(items.length === 1 && !items[0]?.isDir);

const unsupportedArchivePreviewExtensions = ['bz2', 'bzip2', 'xz', 'rar'];

export const isArchivePreviewSupported = (item?: { name?: string }) => {
	const ext = getArchiveFileExtension(item?.name);
	return !!ext && !unsupportedArchivePreviewExtensions.includes(ext);
};

// Archive formats that are recognized by name but are not supported for
// preview/extract at all (e.g. rar). These are intentionally excluded from
// ARCHIVE_FILE_EXTENSIONS so they expose no archive actions, but the file list
// still tags them with an "Unsupported" badge to explain why.
const unsupportedArchiveExtractExtensions = ['rar'];

export const isUnsupportedArchiveExtract = (item?: {
	name?: string;
	isDir?: boolean;
}) => {
	if (!item?.name || item.isDir) {
		return false;
	}
	const ext = item.name.toLowerCase().split('.').pop() || '';
	return unsupportedArchiveExtractExtensions.includes(ext);
};

export const getArchiveErrorCode = (error: any) => {
	const code =
		error?.response?.data?.code || error?.code || error?._error?.code;
	return Number(code);
};

export const getArchiveErrorMessage = (error: any) => {
	const raw =
		error?._error ||
		error?.response?.data?.message ||
		error?.response?.data?.error ||
		error?.response?.data?.msg ||
		error?.message ||
		'';

	if (typeof raw === 'string') {
		return raw;
	}
	return raw?.message || raw?.error || raw?.msg || '';
};

export const isArchivePasswordError = (error: any) =>
	[30001, 30002].includes(getArchiveErrorCode(error));

export interface ArchiveSourceItem {
	name: string;
	path: string;
	isDir: boolean;
	size?: number;
}

export const getArchiveFileTypeI18nKey = (name?: string) => {
	const ext = getArchiveFileExtension(name);
	switch (ext) {
		case 'zip':
			return 'files.file_types.zip';
		case 'rar':
			return 'files.file_types.rar';
		case '7z':
			return 'files.file_types.seven_zip';
		case 'tar':
			return 'files.file_types.tar';
		case 'gz':
		case 'gzip':
		case 'tgz':
		case 'tar.gz':
			return 'files.file_types.gzip';
		case 'bz2':
		case 'bzip2':
		case 'tar.bz2':
			return 'files.file_types.bzip2';
		case 'xz':
		case 'tar.xz':
			return 'files.file_types.tar';
		default:
			return '';
	}
};

export const getArchiveFileIcon = (name?: string) =>
	getArchiveFileExtension(name) ? 'zip' : '';

export const getArchiveDefaultExtractName = (item?: { name: string }) => {
	if (!item?.name) {
		return '';
	}

	const extension = getArchiveFileExtension(item.name);
	if (!extension) {
		return item.name.replace(/\.[^/.]+$/, '') || item.name;
	}

	const escapedExtension = extension.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	return (
		item.name.replace(
			new RegExp(`\\.${escapedExtension}(?:\\.\\d{3})?$`, 'i'),
			''
		) || item.name
	);
};

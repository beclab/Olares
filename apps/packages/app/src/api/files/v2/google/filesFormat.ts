// import { getFileType } from '@bytetrade/core';
import { FileResType, useFilesStore } from 'src/stores/files';
import { GoogleDriveFileItem } from './type';
import { extensionByMimeType } from './utils';
import { DriveType } from 'src/utils/interface/files';
import { encodeUrl } from 'src/utils/encode';
import { appendPath } from '../path';
import { i18n } from 'src/boot/i18n';
import { getFileType } from '@bytetrade/core';

export function format(
	data: {
		data: any[];
		fileExtend: string;
		filePath: string;
		fileType: string;
		name: string;
	},
	origin_id: number
) {
	const filesStore = useFilesStore();

	let dirent_lists = data.data;
	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}

	const hasDirLen = dirent_lists.filter((item) => item.isDir).length;
	const hasFileLen = dirent_lists.filter((item) => !item.isDir).length;

	const googleDir: FileResType = {
		path: appendPath('/Drive/google', data.filePath, data.fileExtend),
		name: data.fileExtend,
		size: 0,
		extension: '',
		modified: 0,
		mode: 0,
		isDir: true,
		isSymlink: false,
		type: '',
		numDirs: hasDirLen,
		numFiles: hasFileLen,
		sorting: {
			by: 'modified',
			asc: true
		},
		numTotalFiles: 0,
		items: [],
		driveType: DriveType.GoogleDrive,
		fileExtend: data.fileExtend,
		filePath: data.filePath,
		fileType: data.fileType
	};

	dirent_lists.forEach((el, index) => {
		const itemPath = appendPath(
			'/Drive/google',
			data.fileExtend,
			encodeUrl(data.filePath),
			encodeUrl(el.name),
			el.isDir ? '/' : ''
		);

		const oPath = el.path;

		const obj: GoogleDriveFileItem = {
			path: itemPath,
			name: el.name.endsWith('/') ? el.name.slice(0, -1) : el.name,
			size: el.fileSize || 0,
			extension: '',
			modified: Date.parse(el.modified) || 0,
			mode: 0,
			isDir: el.isDir,
			isSymlink: false,
			type: el.isDir ? i18n.global.t('files.folders') : getFileType(el.name),
			sorting: {
				by: 'size',
				asc: false
			},
			numDirs: el.numDirs,
			numFiles: el.numFiles,
			numTotalFiles: el.numTotalFiles,
			index,
			url: itemPath + '?src=' + DriveType.GoogleDrive,
			driveType: DriveType.GoogleDrive,
			param: `?canDownload=${!el.canDownload && !el.canExport ? false : true}`,
			id: el.id,
			iconLink: el.meta?.iconLink,
			webContentLink: el.meta?.webContentLink,
			webViewLink: el.meta?.webViewLink,
			thumbnailLink: el.meta?.thumbnailLink,
			id_path: el.id_path,
			fileExtend: el.fileExtend,
			oPath: oPath,
			oParentPath: oPath.substring(
				0,
				oPath.length - el.name.length - (oPath.endsWith('/') ? 1 : 0)
			),
			fileType: el.fileType,
			google_file_id: el.id
		};
		googleDir.items.push(obj);
	});

	return googleDir;
}

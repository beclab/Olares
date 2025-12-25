// import { getFileType } from '@bytetrade/core';
import { FileResType, useFilesStore } from 'src/stores/files';
import { GoogleDriveFileItem } from './type';
import { extensionByMimeType } from './utils';
import { DriveType } from 'src/utils/interface/files';

export function format(data, url, origin_id: number) {
	const filesStore = useFilesStore();

	const name = url.split('/')[2];
	let dirent_lists = data;
	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}

	const hasDirLen = dirent_lists.filter((item) => item.isDir).length;
	const hasFileLen = dirent_lists.filter((item) => !item.isDir).length;

	const seahubDir: FileResType = {
		path: '/',
		name,
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
		fileExtend: '',
		filePath: '',
		fileType: ''
	};

	dirent_lists.forEach((el, index) => {
		const splitUrl_0 = url.split('?')[0];
		const pathSplit = splitUrl_0.split('/').slice(0, 4).join('/');

		let itemPath = `${pathSplit}/${el.meta.id}/`;
		if (el.isDir) {
			itemPath = itemPath.endsWith('/') ? itemPath : `${itemPath}/`;
		}

		const obj: GoogleDriveFileItem = {
			path: itemPath,
			name: el.name.endsWith('/') ? el.name.slice(0, -1) : el.name,
			size: el.fileSize || 0,
			extension: '',
			modified: Date.parse(el.modified) || 0,
			mode: 0,
			isDir: el.isDir,
			isSymlink: false,
			type: extensionByMimeType(el.type),
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
			id: el.meta.id,
			iconLink: el.meta?.iconLink,
			webContentLink: el.meta?.webContentLink,
			webViewLink: el.meta?.webViewLink,
			thumbnailLink: el.meta?.thumbnailLink,
			id_path: el.id_path,
			fileExtend: ''
		};
		seahubDir.items.push(obj);
	});

	return seahubDir;
}

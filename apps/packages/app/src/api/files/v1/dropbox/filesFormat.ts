import { getFileType } from '@bytetrade/core';
import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';

import { i18n } from 'src/boot/i18n';
import { encodeUrl } from 'src/utils/encode';

export function formatGd(data, url, origin_id) {
	const filesStore = useFilesStore();

	const name = url.split('/')[2];
	let dirent_lists = data;
	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}

	const hasDirLen = dirent_lists.filter((item) => item.isDir).length;
	const hasFileLen = dirent_lists.filter((item) => !item.isDir).length;

	const dropboxDir: FileResType = {
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

		const pathname = splitUrl_0.endsWith('/') ? splitUrl_0 : `${splitUrl_0}/`;

		let itemPath = `${pathname}${encodeUrl(el.name)}`;
		if (el.isDir) {
			itemPath = itemPath.endsWith('/') ? itemPath : `${itemPath}/`;
		}

		const obj: FileItem = {
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
			url: itemPath + '?src=' + DriveType.Dropbox,
			driveType: DriveType.Dropbox,
			param: '',
			fileExtend: ''
		};
		dropboxDir.items.push(obj);
	});

	return dropboxDir;
}

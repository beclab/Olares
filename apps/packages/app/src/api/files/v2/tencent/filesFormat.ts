import { getFileType } from '@bytetrade/core';
import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import { appendPath } from '../path';
import { encodeUrl } from 'src/utils/encode';

export function formatGd(
	data: {
		data: any[];
		fileExtend: string;
		filePath: string;
		fileType: string;
		name: string;
	},
	origin_id
) {
	const filesStore = useFilesStore();

	let dirent_lists = data.data;
	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}

	const hasDirLen = dirent_lists.filter((item) => item.isDir).length;
	const hasFileLen = dirent_lists.filter((item) => !item.isDir).length;

	const seahubDir: FileResType = {
		path: appendPath('/Drive/tencent', data.filePath),
		name: data.name,
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
		driveType: DriveType.Tencent,
		fileExtend: data.fileExtend,
		filePath: data.filePath,
		fileType: data.fileType
	};

	dirent_lists.forEach((el, index) => {
		const itemPath = appendPath(
			'/Drive/tencent',
			data.fileExtend,
			encodeUrl(data.filePath),
			encodeUrl(el.name),
			el.isDir ? '/' : ''
		);
		const obj: FileItem = {
			oPath: el.path,
			oParentPath: el.path.substring(
				0,
				el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
			),
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
			url: itemPath + '?src=' + DriveType.Tencent,
			driveType: DriveType.Tencent,
			param: '',
			fileExtend: el.fileExtend,
			fileType: el.fileType
		};
		seahubDir.items.push(obj);
	});
	return seahubDir;
}

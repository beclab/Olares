import { getFileType } from '@bytetrade/core';
import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import { getParams } from 'src/utils/utils';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';
import { displayPath } from './utils';
import { appendPath } from '../path';

export function formatSeahub(
	url: string,
	data: {
		items: any;
		filePath: string;
		fileExtend: string;
		name?: string;
		fileType: string;
		path: string;
		modified?: string;
	},
	origin_id: number
) {
	const filesStore = useFilesStore();
	const activeMenu = filesStore.activeMenu(origin_id);

	let dirent_lists = data.items;

	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}
	const hasDirLen = dirent_lists.filter((item) => item.type === 'dir').length;
	const hasFileLen = dirent_lists.filter((item) => item.type === 'file').length;

	const repo_name = activeMenu.label;
	const type = getParams(activeMenu.params!, 'type');
	const p = getParams(activeMenu.params!, 'p');

	const seahubDir: FileResType = {
		path: appendPath('/Seahub', repo_name, data.path),
		name: data.name && data.name.length > 0 ? data.name : repo_name,
		size: 0,
		extension: '',
		modified: data.modified ? new Date(data.modified).getTime() : 0,
		mode: 0,
		isDir: true,
		isSymlink: false,
		type: i18n.global.t('files.folders'),
		numDirs: hasDirLen,
		numFiles: hasFileLen,
		sorting: {
			by: 'modified',
			asc: true
		},
		numTotalFiles: 0,
		items: [],
		driveType: DriveType.Sync,
		fileExtend: data.fileExtend,
		filePath: data.filePath,
		fileType: data.fileType
	};

	dirent_lists.forEach((el, index) => {
		const extension = getextension(el.name);
		const path =
			displayPath(el, repo_name) + `?id=${el.fileExtend}&type=${type}&p=${p}`;
		el.oPath = el.path;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = path;
		el.index = index;
		el.driveType = DriveType.Sync;
		el.parentPath = el.parent_dir;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = el.isDir ? i18n.global.t('files.folders') : getFileType(el.name);
		seahubDir.items.push(el);
	});

	return seahubDir;
}

export function formatSeahubRepos(name, datas) {
	const dirent_lists = datas;
	const hasDirLen = dirent_lists.length;
	const hasFileLen = 0;

	const seahubDir: FileResType = {
		path: '',
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
		driveType: DriveType.Sync,
		fileExtend: '',
		filePath: '',
		fileType: ''
	};

	dirent_lists.forEach((el, index) => {
		const itemPath = `/Seahub/${encodeUrl(el.repo_name)}${el.path || ''}/?id=${
			el.repo_id
		}&type=${el.type}&p=${el.permission.trim()}`;

		const obj: FileItem = {
			path: itemPath,
			name: el.repo_name,
			size: el.size || 0,
			extension: '',
			modified: Date.parse(el.last_modified) || 0,
			mode: 0,
			isDir: true,
			isSymlink: false,
			type: el.type,
			sorting: {
				by: 'size',
				asc: false
			},
			numDirs: el.numDirs,
			numFiles: el.numFiles,
			numTotalFiles: el.numTotalFiles,
			index,
			url: '',
			driveType: DriveType.Sync,
			param: '',
			fileExtend: el.repo_id,
			fileType: DriveType.Sync,
			oPath: '/'
		};
		seahubDir.items.push(obj);
	});
	return seahubDir;
}

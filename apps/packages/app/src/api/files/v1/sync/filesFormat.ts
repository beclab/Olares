import { getFileType } from '@bytetrade/core';
import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import { getParams } from 'src/utils/utils';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';

export function formatSeahub(
	url: string,
	data: { dirent_list: any },
	origin_id: number
) {
	const filesStore = useFilesStore();
	const activeMenu = filesStore.activeMenu(origin_id);

	const selUrl = url.split('/')[url.split('/').length - 2];
	let dirent_lists = data.dirent_list;
	if (filesStore.filterHiddenDir[origin_id]) {
		dirent_lists = dirent_lists.filter((item) => !item.name.startsWith('.'));
	}

	const hasDirLen = dirent_lists.filter((item) => item.type === 'dir').length;
	const hasFileLen = dirent_lists.filter((item) => item.type === 'file').length;

	const repo_name = activeMenu.label;
	const repo_id = getParams(activeMenu.params!, 'id');
	const type = getParams(activeMenu.params!, 'type');
	const p = getParams(activeMenu.params!, 'p');
	let curPath = `/Files${url}`;
	if (!curPath.endsWith('/')) curPath += '/';

	const seahubDir: FileResType = {
		path: url,
		url: curPath,
		name: selUrl,
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
		const extension = getextension(el.name);
		const fileTypeName =
			el.type === 'dir' ? i18n.global.t('files.folders') : getFileType(el.name);
		const path =
			el.type === 'dir' ? el.parent_dir + el.name + '/' : el.parent_dir;
		const itemPath = `/Seahub/${encodeUrl(repo_name)}${encodeUrl(
			path
		)}?id=${repo_id}&type=${type}&p=${p}`;
		const url =
			curPath.split('p=/')[0] +
			'p=/' +
			encodeUrl(el.name) +
			curPath.split('p=/')[1];

		const obj: FileItem = {
			path: itemPath,
			name: el.name,
			size: el.size || 0,
			extension: extension,
			modified: el.mtime * 1000 || 0,
			mode: 0,
			isDir: el.type === 'dir' ? true : false,
			isSymlink: false,
			type: fileTypeName,

			parentPath: el.parent_dir,
			sorting: {
				by: 'size',
				asc: false
			},
			numDirs: el.numDirs,
			numFiles: el.numFiles,
			numTotalFiles: el.numTotalFiles,
			encoded_thumbnail_src: el.encoded_thumbnail_src || undefined,
			driveType: DriveType.Sync,
			param: '',
			url,
			index,
			fileExtend: el.fileExtend || 'sync'
		};
		seahubDir.items.push(obj);
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
			fileExtend: el.fileExtend || 'sync'
		};
		seahubDir.items.push(obj);
	});
	return seahubDir;
}

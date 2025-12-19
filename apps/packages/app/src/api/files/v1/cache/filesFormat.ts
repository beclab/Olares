import { FileItem, FileResType, useFilesStore } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { filterPcvPath } from './../common/common';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';

export function formatAppData(node, data, url, origin_id) {
	const filesStore = useFilesStore();
	data.origin = DriveType.Cache;
	data.path = filterPcvPath(data.path);
	data.url = `/Files${url}`;
	if (data.isDir) {
		if (!data.url.endsWith('/')) data.url += '/';
	}

	let curItems: FileItem[] = data.items || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = data.items.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		const extension = getextension(el.name);
		const splitPath = filterPcvPath(el.path).split('/');
		splitPath.splice(splitPath.indexOf('AppData') + 1, 0, node);
		const joinPath = encodeUrl(splitPath.join('/').replace('AppData', 'Cache'));
		el.index = index;
		el.url = encodeUrl(`${data.url}${el.name}`);
		if (el.isDir) el.url += '/';
		el.path = el.isDir ? `${joinPath}/` : joinPath;
		el.driveType = DriveType.Cache;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
	});

	data.items = curItems;

	return data;
}

export function formatAppDataNode(url, data) {
	const nodeDir: FileResType = {
		path: url,
		name: 'AppData',
		size: 0,
		extension: '',
		modified: 0,
		mode: 0,
		isDir: true,
		isSymlink: false,
		type: '',
		numDirs: 0,
		numFiles: 0,
		sorting: {
			by: 'modified',
			asc: true
		},
		fileSize: 0,
		numTotalFiles: 0,
		items: <FileItem[]>[],
		driveType: DriveType.Cache,
		fileExtend: '',
		filePath: '',
		fileType: ''
	};

	if (data.code == 200) {
		nodeDir.numDirs = data.data.length;

		data.data.forEach((el, index) => {
			const extension = getextension(el.metadata.name);
			const path = '/Cache/' + el.metadata.name;
			const item: FileItem = {
				path: path.endsWith('/') ? path : `${path}/`,
				name: el.metadata.name,
				size: 4096,
				extension: extension,
				modified: 0,
				mode: 0,
				isDir: true,
				isSymlink: false,
				type: '',
				sorting: {
					by: 'size',
					asc: false
				},
				driveType: DriveType.Cache,
				param: '',
				url: '',
				index: index,
				fileExtend: ''
			};

			nodeDir.items.push(item);
		});
	}

	return nodeDir;
}

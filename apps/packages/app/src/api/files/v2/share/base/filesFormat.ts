import { getFileType } from '@bytetrade/core';
import { FileItem, useFilesStore } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import { displayPath, displaySharePath } from './utils';

export function formatShareListData(data, driveType: DriveType, origin_id) {
	const filesStore = useFilesStore();
	data.origin = driveType;
	data.driveType = driveType;
	data.path = displayPath(data);

	if (data.share_paths) {
		let curItems = data.share_paths || [];
		if (filesStore.filterHiddenDir[origin_id]) {
			curItems = data.share_paths.filter((item) => !item.name.startsWith('.'));
		}

		curItems.map((el, index) => {
			el.index = index;
			el.oPath = el.path;
			el.path = displaySharePath(el);
			el.driveType = driveType;
			el.isDir = true;
			el.type = i18n.global.t('files.folders');
			el.isShareItem = true;
			el.fileExtend = el.id;
			el.node = el.extend;
		});
		data.shareRoot = true;
		data.items = curItems;
	} else {
		data.items = [];
	}

	return data;
}

export function formatDrive(
	data: any,
	origin_id,
	driveType: DriveType = DriveType.Share
) {
	const filesStore = useFilesStore();
	data.origin = driveType;
	data.driveType = driveType;
	data.path = displayPath(data);

	let curItems: FileItem[] = data.items || data.dirent_list || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = data.items.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		// if (data.dirent_list) {
		// 	const secondSlashIndex = el.path.indexOf('/', el.path.indexOf('/') + 1);
		// 	el.path = el.path.slice(secondSlashIndex + 1);
		// }

		const extension = getextension(el.name);
		const p = displayPath(el);
		const isDir =
			el.isDir != undefined ? el.isDir : el.type === 'dir' ? true : false;

		el.oPath = el.path;
		el.isDir = isDir;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = p;
		el.index = index;
		el.driveType = driveType;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = isDir ? i18n.global.t('files.folders') : getFileType(el.name);
	});

	data.items = curItems;

	return data;
}

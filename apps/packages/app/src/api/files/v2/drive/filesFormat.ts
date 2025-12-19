import { getFileType } from '@bytetrade/core';
import { FileItem, useFilesStore } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import { displayPath } from './utils';

export function formatDrive(data, url, origin_id) {
	const filesStore = useFilesStore();
	data.origin = DriveType.Drive;
	data.driveType = DriveType.Drive;
	data.path = displayPath(data);

	let curItems: FileItem[] = data.items || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = data.items.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		const extension = getextension(el.name);
		const p = displayPath(el);
		el.oPath = el.path;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = p;
		el.index = index;
		el.driveType = DriveType.Drive;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = el.isDir ? i18n.global.t('files.folders') : getFileType(el.name);
	});

	data.items = curItems;

	return data;
}

import { getFileType } from '@bytetrade/core';
import { FileItem, useFilesStore } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';
import { appendPath } from '../path';
import { normalExternalTypes } from './type';
import { displayPath } from './utils';

export function formatDrive(data: any, url: string, origin_id: number) {
	const filesStore = useFilesStore();
	data.origin = DriveType.External;
	data.driveType = DriveType.External;
	data.path = displayPath(data);

	let curItems: FileItem[] = data.items || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = data.items.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		const extension = getextension(el.name);
		el.oPath = el.path;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = displayPath(el);
		el.index = index;
		el.driveType = DriveType.External;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = el.isDir ? i18n.global.t('files.folders') : getFileType(el.name);
		el.externalType =
			el.externalType && normalExternalTypes.includes(el.externalType)
				? undefined
				: el.externalType;
	});

	data.items = curItems;

	return data;
}

import { getFileType } from '@bytetrade/core';
import { useFilesStore, FileItem } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { filterPcvPath } from './../common/common';
import { i18n } from 'src/boot/i18n';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';

export function formatDrive(data, url, origin_id) {
	const filesStore = useFilesStore();
	data.origin = DriveType.Drive;
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
		const pvcPath = filterPcvPath(el.path);

		const path = `/Files${encodeUrl(pvcPath)}`;
		el.url = `${data.url}${encodeUrl(el.name)}`;
		if (el.isDir) {
			el.path = path.endsWith('/') ? path : `${path}/`;
			el.url = el.url + '/';
		} else {
			el.path = path;
		}

		el.index = index;
		el.driveType = DriveType.Drive;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = el.isDir ? i18n.global.t('files.folders') : getFileType(el.name);
	});

	data.items = curItems;

	return data;
}
